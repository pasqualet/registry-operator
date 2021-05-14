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
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DockerioCredentialsReconciler reconciles a DockerioCredentials object
type DockerioCredentialsReconciler struct {
	CredentialsReconciler
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=registry.astrokube.com,resources=dockeriocredentials,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=dockeriocredentials/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=dockeriocredentials/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DockerioCredentials object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *DockerioCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dockeriocredentials", req.NamespacedName)
	dockerioCredentials := &registryv1alpha1.DockerioCredentials{}

	// Skip if dockerioCredentials doesn't exists
	if err := r.Get(ctx, req.NamespacedName, dockerioCredentials); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get DockerioCredentials")
		return ctrl.Result{}, err
	}

	// dockerioCredentials is not going to be deleted
	if dockerioCredentials.ObjectMeta.DeletionTimestamp.IsZero() {

		// If Provisioning status if is not set
		if dockerioCredentials.Status.Phase == "" {
			if err := r.setStatus(log, dockerioCredentials, registryv1alpha1.DockerioCredentialsProvisioning); err != nil {
				return ctrl.Result{}, err
			}
		}

		credentials, err := r.getToken(log, dockerioCredentials)
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
		if err := r.setStatus(log, dockerioCredentials, registryv1alpha1.DockerioCredentialsActive); err != nil {
			return ctrl.Result{}, err
		}

	} else {
		// Set Terminating status
		if err := r.setStatus(log, dockerioCredentials, registryv1alpha1.DockerioCredentialsTerminating); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	log.Info("Reconciled successfully")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DockerioCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.DockerioCredentials{}).
		Complete(r)
}

func (r *DockerioCredentialsReconciler) setStatus(log logr.Logger, dockerioCredentials *registryv1alpha1.DockerioCredentials, phase registryv1alpha1.DockerioCredentialsPhase) error {
	ctx := context.Background()

	dockerioCredentials.Status.Phase = phase
	log.Info(fmt.Sprintf("Setting status to %v", dockerioCredentials.Status.Phase))
	if err := r.Status().Update(ctx, dockerioCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}

func (r *DockerioCredentialsReconciler) getToken(log logr.Logger, dockerioCredentials *registryv1alpha1.DockerioCredentials) (*RegistryCredentials, error) {
	authStr := dockerioCredentials.Spec.User + ":" + dockerioCredentials.Spec.Password
	msg := []byte(authStr)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	token := string(encoded)

	return &RegistryCredentials{
		Name:               dockerioCredentials.ObjectMeta.Name,
		Namespace:          dockerioCredentials.ObjectMeta.Namespace,
		Host:               "https://index.docker.io/v1/",
		AuthorizationToken: token,
		ExpiresAt:          nil,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(dockerioCredentials, registryv1alpha1.GroupVersion.WithKind("DockerioCredentials")),
		},
	}, nil
}
