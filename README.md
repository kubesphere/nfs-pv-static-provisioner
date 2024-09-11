# nfs-pv-static-provisioner
A NFS persistent volume static provisioner, which allows you to quickly bind an existing NFS volume to PVC. 

**This provisioner will NOT provision volumes on NFS server when you create a PVC.** If you are looking for a dynamic provisioner of NFS, please consider other projects, like [this](https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner).

## Description
The provisioner listens PVC CREATE and UPDATE events, when the PVC has demanded annotations, the provisioner will create/update a NFS PV and bind it with the PVC automatically.

| Annotation                                 | Required | Default Value | Explanation                                                  | Example Values                          |
|--------------------------------------------|----------|---------------|--------------------------------------------------------------|-----------------------------------------|
| storage.kubesphere.io/nfs-static-provision | Y        | "false"       | must set to "true" in order to use this provisioner          | "true"                                  |
| storage.kubesphere.io/nfs-server           | Y        | ""            | nfs server hostname or IP address (PV.spec.nfs.server)       | "example.nfs.server.com", "192.168.0.5" |
| storage.kubesphere.io/nfs-path             | Y        | ""            | nfs volume absolute path (PV.spec.nfs.path)                  | "/exports/volume1"                      |
| storage.kubesphere.io/nfs-readonly         | N        | "false"       | whether the volume is read-only (PV.spec.nfs.readOnly)       | "false", "true"                         |
| storage.kubesphere.io/reclaim-policy*      | N        | "Delete"      | reclaim policy of PV (PV.spec.persistentVolumeReclaimPolicy) | "Delete", "Retain"                      |
| storage.kubesphere.io/mount-options        | N        | ""            | mount options of PV (PV.spec.mountOptions)                   | `"[\"nfsvers=3\",\"nolock\",\"hard\"]"` |

- *When reclaim policy is "Delete", the PV will be deleted when the PVC is deleted. However, this only affects the PV resource in k8s cluster, the real backend volume on NFS server still exists.

## Usecase
- As a tenant(e.g. admin of a namespace) on kubernetes cluster, you don't have permissions to create PV resources (as PV is cluster-level resource), but you own an external NFS server and want to use the existing volumes via PVC.

## Deploy
### Deploy
To deploy the controller:
```sh
make deploy
```

### Test
Create a PVC and check if it can be automatically bound. Take [this](./config/samples/pvc.yaml) for example.

### Undeploy
To uninstall the controller:
```sh
make undeploy
```

## Events
PV create/update events will be issued targeting the PVC object.

e.g.
```sh
$ kubectl describe pvc pvc-nfs
Name:          pvc-nfs
Namespace:     default
StorageClass:
Status:        Pending
Volume:        pvc-f6c5a1c9-cbfc-4994-8079-956024db8e52
Labels:        <none>
Annotations:   storage.kubesphere.io/nfs-path: aaa
               storage.kubesphere.io/nfs-server: test.nfs.server.com
               storage.kubesphere.io/nfs-static-provision: true
Finalizers:    [kubernetes.io/pvc-protection]
Capacity:      0
Access Modes:
VolumeMode:    Filesystem
Used By:       <none>
Events:
  Type     Reason             Age               From                         Message
  ----     ------             ----              ----                         -------
  Normal   FailedBinding      9s                persistentvolume-controller  no persistent volumes available for this claim and no storage class is set
  Normal   VolumeNameUpdated  9s                nfs-pv-static-provisioner    volumeName updated successfully
  Warning  CreatePVFailed     4s (x11 over 9s)  nfs-pv-static-provisioner    failed to create pv pvc-f6c5a1c9-cbfc-4994-8079-956024db8e52, error: PersistentVolume "pvc-f6c5a1c9-cbfc-4994-8079-956024db8e52" is invalid: spec.nfs.path: Invalid value: "aaa": must be an absolute path
```

```sh
$ kubectl describe pvc pvc-nfs
Name:          pvc-nfs
Namespace:     default
StorageClass:
Status:        Bound
Volume:        pvc-f6f9747e-3479-4664-a32f-268b2c62f0bd
Labels:        <none>
Annotations:   pv.kubernetes.io/bind-completed: yes
               storage.kubesphere.io/nfs-path: /a/b
               storage.kubesphere.io/nfs-server: test.nfs.server.com
               storage.kubesphere.io/nfs-static-provision: true
Finalizers:    [kubernetes.io/pvc-protection]
Capacity:      1Gi
Access Modes:  RWX
VolumeMode:    Filesystem
Used By:       <none>
Events:
  Type    Reason             Age   From                         Message
  ----    ------             ----  ----                         -------
  Normal  FailedBinding      12s   persistentvolume-controller  no persistent volumes available for this claim and no storage class is set
  Normal  VolumeNameUpdated  12s   nfs-pv-static-provisioner    volumeName updated successfully
  Normal  PVCreated          12s   nfs-pv-static-provisioner    pv pvc-f6f9747e-3479-4664-a32f-268b2c62f0bd created successfully
```
