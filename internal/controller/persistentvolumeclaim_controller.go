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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PersistentVolumeClaimReconciler reconciles a PersistentVolumeClaim object
type PersistentVolumeClaimReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PersistentVolumeClaim object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *PersistentVolumeClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("pvc", req.NamespacedName.String())

	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, req.NamespacedName, pvc)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !pvc.DeletionTimestamp.IsZero() {
		logger.V(4).Info("pvc is being deleted")
		return ctrl.Result{}, nil
	}

	if pvc.Status.Phase == corev1.ClaimBound {
		logger.V(4).Info("pvc is already bound")
		return ctrl.Result{}, nil
	}

	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "" {
		logger.V(4).Info("pvc's StorageClassName is not empty string which means pv is not static provision")
		return ctrl.Result{}, nil
	}

	nfsPVC := NFSPVC(*pvc)
	staticProvision := nfsPVC.IsStaticProvision()
	if !staticProvision {
		logger.V(4).Info("pvc's annotation shows pv is not static provision")
		return ctrl.Result{}, nil
	}

	var pv *corev1.PersistentVolume
	pv, err = nfsPVC.ParsePV()
	if err == nil {
		r.Recorder.Eventf(pvc, corev1.EventTypeNormal, "PVParsed", "parsed pv from pvc successfully")
	} else {
		r.Recorder.Eventf(pvc, corev1.EventTypeWarning, "ParsePVFailed", "failed to parse pv from pvc, error: %s", err.Error())
		return ctrl.Result{}, err
	}

	if pvc.Spec.VolumeName == "" {
		pvc.Spec.VolumeName = pv.Name
		err = r.Client.Update(ctx, pvc)
		if err == nil {
			r.Recorder.Eventf(pvc, corev1.EventTypeNormal, "VolumeNameUpdated", "volumeName updated successfully")
		} else {
			r.Recorder.Eventf(pvc, corev1.EventTypeWarning, "UpdateVolumeNameFailed", "failed to update volumeName, error: %s", err.Error())
		}
		return ctrl.Result{}, err
	} else {
		pv2 := &corev1.PersistentVolume{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: pvc.Spec.VolumeName}, pv2)
		if err != nil {
			if apierrors.IsNotFound(err) {
				err = r.Client.Create(ctx, pv)
				if err == nil {
					r.Recorder.Eventf(pvc, corev1.EventTypeNormal, "PVCreated", "pv %s created successfully", pv.Name)
				} else {
					r.Recorder.Eventf(pvc, corev1.EventTypeWarning, "CreatePVFailed", "failed to create pv %s, error: %s", pv.Name, err.Error())
				}
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PersistentVolumeClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}).
		WithEventFilter(predicate.Funcs{
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				return false
			},
			CreateFunc: func(event event.CreateEvent) bool {
				return true
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
