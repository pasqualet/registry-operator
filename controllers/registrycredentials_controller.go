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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	"github.com/go-logr/logr"
)

// RegistryCredentialsReconciler reconciles a RegistryCredentials object
type RegistryCredentialsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=registry.astrokube.com,resources=registrycredentials,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=registrycredentials/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=registrycredentials/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RegistryCredentials object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *RegistryCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	registryCredentials := &registryv1alpha1.RegistryCredentials{}

	// Skip if registryCredentials doesn't exists
	if err := r.Get(ctx, req.NamespacedName, registryCredentials); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		l.Error(err, "Unable to get ECRCredentials")
		return ctrl.Result{}, err
	}

	// registryCredentials is not going to be deleted
	if registryCredentials.ObjectMeta.DeletionTimestamp.IsZero() {

		switch registryCredentials.Status.State {
		case "":
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsAuthenticated:
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsErrored:
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsAuthenticating:
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsUnauthorized:
			return ctrl.Result{}, nil
		}

	} else {
		// Set Terminating status
		if err := r.setStatus(l, registryCredentials, registryv1alpha1.RegistryCredentialsTerminating); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil

	}

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.RegistryCredentials{}).
		Complete(r)
}

func (r *RegistryCredentialsReconciler) setStatus(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials, state registryv1alpha1.RegistryCredentialsState) error {
	ctx := context.Background()

	registryCredentials.Status.State = state
	if err := r.Status().Update(ctx, registryCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}

func (r *RegistryCredentialsReconciler) authenticate(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials) error {
	return nil
}
