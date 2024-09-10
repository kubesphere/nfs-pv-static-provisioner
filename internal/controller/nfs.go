package controller

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog/v2"
)

const (
	AnnotationNFSStaticProvision = "storage.kubesphere.io/nfs-static-provision"
	AnnotationNFSServer          = "storage.kubesphere.io/nfs-server"
	AnnotationNFSPath            = "storage.kubesphere.io/nfs-path"
	AnnotationNFSReadonly        = "storage.kubesphere.io/nfs-readonly"
	AnnotationMountOptions       = "storage.kubesphere.io/mount-options"
	AnnotationReclaimPolicy      = "storage.kubesphere.io/reclaim-policy"
)

type NFSPV corev1.PersistentVolume

func (p NFSPV) IsStaticProvision() bool {
	sps, ok := p.Annotations[AnnotationNFSStaticProvision]
	sp, err := strconv.ParseBool(sps)
	return ok && err == nil && sp
}

func (p NFSPV) NeedDelete() bool {
	return p.IsStaticProvision() &&
		p.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete &&
		(p.Status.Phase == corev1.VolumeFailed || p.Status.Phase == corev1.VolumeReleased)
}

type NFSPVC corev1.PersistentVolumeClaim

func (p NFSPVC) IsStaticProvision() bool {
	sps, ok := p.Annotations[AnnotationNFSStaticProvision]
	sp, err := strconv.ParseBool(sps)
	return ok && err == nil && sp
}

var newAnnotationInvalidError = func(annotation string) error {
	return fmt.Errorf("annotation %s not found or has invalid value", annotation)
}

func (p NFSPVC) ParsePV() (*corev1.PersistentVolume, error) {
	var err error

	server, ok := p.Annotations[AnnotationNFSServer]
	if !ok || server == "" {
		return nil, newAnnotationInvalidError(AnnotationNFSServer)
	}

	path, ok := p.Annotations[AnnotationNFSPath]
	if !ok || path == "" {
		return nil, newAnnotationInvalidError(AnnotationNFSPath)
	}

	mountOptions := make([]string, 0)
	mountOptionsStr, ok := p.Annotations[AnnotationMountOptions]
	if ok && len(mountOptionsStr) > 0 {
		err = json.Unmarshal([]byte(mountOptionsStr), &mountOptions)
		if err != nil {
			klog.ErrorS(err, "failed to unmarshal mountOptions")
			return nil, newAnnotationInvalidError(AnnotationMountOptions)
		}
	}

	reclaimPolicy := corev1.PersistentVolumeReclaimDelete
	validReclaimPolicies := []corev1.PersistentVolumeReclaimPolicy{
		corev1.PersistentVolumeReclaimRetain,
		corev1.PersistentVolumeReclaimRecycle,
		corev1.PersistentVolumeReclaimDelete,
	}
	reclaimPolicyStr, ok := p.Annotations[AnnotationReclaimPolicy]
	if ok {
		if slices.Contains(validReclaimPolicies, corev1.PersistentVolumeReclaimPolicy(reclaimPolicyStr)) {
			reclaimPolicy = corev1.PersistentVolumeReclaimPolicy(reclaimPolicyStr)
		} else {
			return nil, newAnnotationInvalidError(AnnotationReclaimPolicy)
		}
	}

	var readonly bool
	readonlyStr, ok := p.Annotations[AnnotationNFSReadonly]
	if ok && len(readonlyStr) > 0 {
		readonly, err = strconv.ParseBool(readonlyStr)
		if err != nil {
			return nil, newAnnotationInvalidError(AnnotationNFSReadonly)
		}
	}

	name := p.Spec.VolumeName
	if name == "" {
		name = fmt.Sprintf("pvc-%s", uuid.NewUUID())
	}

	var storageClassName string
	if p.Spec.StorageClassName != nil {
		storageClassName = *p.Spec.StorageClassName
	}

	pv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				AnnotationNFSStaticProvision: "true",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity:                      p.Spec.Resources.Requests,
			AccessModes:                   p.Spec.AccessModes,
			PersistentVolumeReclaimPolicy: reclaimPolicy,
			MountOptions:                  mountOptions,
			ClaimRef: &corev1.ObjectReference{
				Kind:       "PersistentVolumeClaim",
				Namespace:  p.Namespace,
				Name:       p.Name,
				UID:        p.UID,
				APIVersion: "v1",
			},
			StorageClassName: storageClassName,
			VolumeMode:       p.Spec.VolumeMode,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   server,
					Path:     path,
					ReadOnly: readonly,
				},
			},
		},
	}

	return pv, nil
}
