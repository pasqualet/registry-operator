/*
Copyright 2021 AstroKube.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ECRCredentialsReconciler reconciles a ECRCredentials object
type ECRCredentialsReconciler struct {
	CredentialsReconciler
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=registry.astrokube.com,resources=ecrcredentials,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=ecrcredentials/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=ecrcredentials/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ECRCredentials object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ECRCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("ecrcredentials", req.NamespacedName)

	ecrCredentials := &registryv1alpha1.ECRCredentials{}

	// Skip if ecrCredentials doesn't exists
	if err := r.Get(ctx, req.NamespacedName, ecrCredentials); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get ECRCredentials")
		return ctrl.Result{}, err
	}

	// ecrCredentials is not going to be deleted
	if ecrCredentials.ObjectMeta.DeletionTimestamp.IsZero() {

		// If Provisioning status if is not set
		if ecrCredentials.Status.Phase == "" {
			if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsProvisioning); err != nil {
				return ctrl.Result{}, err
			}
		}

		credentials, err := r.getToken(log, ecrCredentials)
		if err != nil {
			log.Error(err, "Unable to get token")
			return ctrl.Result{}, err
		}

		secret := r.getSecret(*credentials)
		log.Info(string(secret.Data[corev1.DockerConfigKey]))

		err = r.createOrUpdateSecret(log, &secret)
		if err != nil {
			log.Error(err, "Unable to create secret")
			return ctrl.Result{}, err
		}

		// Set Active status
		if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsActive); err != nil {
			return ctrl.Result{}, err
		}

	} else {
		// Set Terminating status
		if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsTerminating); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	log.Info("Reconciled successfully")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ECRCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.ECRCredentials{}).
		Complete(r)
}

func (r *ECRCredentialsReconciler) setStatus(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials, phase registryv1alpha1.ECRCredentialsPhase) error {
	ctx := context.Background()

	ecrCredentials.Status.Phase = phase
	log.Info(fmt.Sprintf("Setting status to %v", ecrCredentials.Status.Phase))
	if err := r.Status().Update(ctx, ecrCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}

func (r *ECRCredentialsReconciler) getToken(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials) (*RegistryCredentials, error) {
	credentials := credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     ecrCredentials.Spec.AccessKeyID,
		SecretAccessKey: ecrCredentials.Spec.SecretAccessKey,
	})
	awsConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String(ecrCredentials.Spec.Region),
	}
	awsSession := session.New(awsConfig)

	svc := ecr.New(awsSession)
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(aerr, "Unable to get authorization token")
			return nil, aerr
		} else {
			log.Error(err, "Unable to get authorization token")
			return nil, err
		}
	}

	stsSvc := sts.New(awsSession)
	identity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Error(err, "Unable to get CallerIdentity")
	}

	host := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", *identity.Account, ecrCredentials.Spec.Region)
	return &RegistryCredentials{
		Name:               ecrCredentials.ObjectMeta.Name,
		Namespace:          ecrCredentials.ObjectMeta.Namespace,
		Host:               host,
		AuthorizationToken: *result.AuthorizationData[0].AuthorizationToken,
		ExpiresAt:          result.AuthorizationData[0].ExpiresAt,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(ecrCredentials, registryv1alpha1.GroupVersion.WithKind("ECRCredentials")),
		},
	}, nil
}
