/*
Copyright 2024.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PersistentVolumeReconciler reconciles a PersistentVolume object
type PersistentVolumeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=persistentvolumes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PersistentVolume object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *PersistentVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pv := &corev1.PersistentVolume{}
	err := r.Client.Get(ctx, req.NamespacedName, pv)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !pv.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	nfspv := NFSPV(*pv)
	if nfspv.NeedDelete() {
		err = r.Client.Delete(ctx, pv)
		logger.Error(err, "pv deleted")
		if err != nil {
			r.Recorder.Eventf(pv, corev1.EventTypeWarning, "DeletePVFailed", "failed to delete pv, error: %s", err.Error())
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PersistentVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolume{}).
		WithEventFilter(predicate.Funcs{
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				return false
			},
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				return true
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}