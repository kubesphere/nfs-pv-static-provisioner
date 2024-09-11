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
Create a PVC and check if it can be automatically bound. Take [this](./config/samples/PVC.yaml) for example.

### Undeploy
To uninstall the controller:
```sh
make undeploy
```

## Events
PV creation events will be issued targeting the PVC object.

e.g.
```sh
$ kubectl describe pvc pvc-nfs
Name:          pvc-nfs
Namespace:     default
...
Events:
  Type     Reason             Age                 From                         Message
  ----     ------             ----                ----                         -------
  Warning  ParsePVFailed      30s (x14 over 71s)  nfs-pv-static-provisioner    failed to parse pv from pvc, error: annotation storage.kubesphere.io/nfs-path not found or has invalid value
  Normal   FailedBinding      6s (x6 over 71s)    persistentvolume-controller  no persistent volumes available for this claim and no storage class is set
  Normal   VolumeNameUpdated  6s                  nfs-pv-static-provisioner    volumeName updated successfully
  Warning  CreatePVFailed     3s (x10 over 6s)    nfs-pv-static-provisioner    failed to create pv pvc-091ecc93-698e-4421-b758-d4883a786aea, error: PersistentVolume "pvc-091ecc93-698e-4421-b758-d4883a786aea" is invalid: spec.nfs.path: Invalid value: "aaa": must be an absolute path
```
