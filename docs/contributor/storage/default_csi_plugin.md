# MicroShift Storage Plugin Overview

> **IMPORTANT!** The default LVMS configuration is intended to match the developer environment described in [MicroShift Development Environment](../devenv_setup.md). See section **[Configuring LVMS](#Configuring-LVMS)** for guidance on configuring LVMS for your environment.

MicroShift enables dynamic storage provisioning out of the box with the LVMS CSI plugin. This plugin is a downstream
Red Hat build of TopoLVM. This provisioner will create a new LVM logical volume in the `rhel` volume group for each
PersistenVolumeClaim(PVC), and make these volumes available to pods. For more information on LVMS, visit the repo's
[README](https://github.com/red-hat-storage/topolvm).

## Design

LVMS manifests can be found in [microshift/assets/components/lvms](../../../assets/components/lvms). For an overview
of LVMS architecture, see the [design doc](https://github.com/red-hat-storage/topolvm/blob/main/docs/design.md).

## Deployment

LVMS is deployed onto the cluster in the `openshift-storage` namespace, after MicroShift
boots. [StorageCapacity tracking](https://kubernetes.io/docs/concepts/storage/storage-capacity/) is used to ensure that
Pods with an LVMS PVC are not scheduled if the requested storage is greater than the volume group's remaining free
storage.

### Configuring LVMS

MicroShift supports pass-through of users' LVMS configuration and allows users to specify custom volume-groups, thin volume provisioning parameters, reserved unallocated volume-group space, among others.  This file can be specified anytime.  If MicroShift is already running, simply restart it to pick up the latest config.

This configuration is specific to the LVMD runtime and informs how it interacts with the host. LVMD is a subcomponent of LVMS, and in the case of MicroShift, deployed automatically as a DaemonSet.

#### Specification

Full documentation of the config spec can be found at [github.com/red-hat-storage/topolvm/blob/v4.11.0/docs/lvmd.md](https://github.com/red-hat-storage/topolvm/blob/v4.11.0/docs/lvmd.md).

#### Path

The user provided lvmd config should be written to the same directory as the MicroShift config.  If an lvmd configuration file does not exist in `/etc/microshift/lvmd.yaml`, MicroShift will use default values.

## System Requirements

### Default Volume Group

When an LVMS configuration is not found on the MicroShift host, the service will proceed through a list of cases to
determine a default volume group. MicroShift uses `vgs` to volume groups. The result of this command will trigger one 
of the following cases:

* **No VolumeGroups Exist:** If no volume groups are discovered, LVMS will not be deployed. If a volume group is created
after MicroShift has started, LVMS will not be deployed until the next service restart.

* **One VolumeGroup Exists:** If there is only one volume group on the system, LVMS uses it by default. MicroShift does
not verify the unallocated space on the volume group. If this volume group is fully allocated, PVCs will hang in the 
`Pending` state.

* **Multiple VolumeGroups Exist:** When MicroShift discovers more than one volume group, it checks for a volume group
named `microshift`, and if found, will deploy LVMS with this volume group. If the `microshift` volume group does not 
exist, LVMS will not be deployed. This prevents the nondeterministic adoption of a volume group by LVMS.  

### Volume Size Increments

LVMS provisions storage in increments of 1Gb. Storage requests are rounded up to the near gigabyte. When a volume
group's capacity is less than 1Gb, the PersistentVolumeClaim will register a `ProvisioningFailed` event:

```shell
  Warning  ProvisioningFailed    3s (x2 over 5s)  topolvm.cybozu.
  com_topolvm-controller-858c78d96c-xttzp_0fa83aef-2070-4ae2-bcb9-163f818dcd9f failed to provision volume with 
  StorageClass "topolvm-provisioner": rpc error: code = ResourceExhausted desc = no enough space left on VG: 
  free=(BYTES_INT), requested=(BYTES_INT)
```

## Usage

### Create Workload Storage

LVMS's [StorageClass](../../../assets/components/lvms/topolvm_default-storage-class.yaml) is deployed with a default
StorageClass. Any PersistentVolumeClaim without a `.spec.storageClassName` defined will automatically have a
PersistentVolume provisioned from the default StorageClass.

Here is a simple example workflow to provision and mount a logical volume to a pod:

```shell
cat <<'EOF' | oc apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: my-lv-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1G
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: nginx
    image: nginx
    command: ["/usr/bin/sh", "-c"]
    args: ["sleep", "1h"]
    volumeMounts:
    - mountPath: /mnt
      name: my-volume
  volumes:
    - name: my-volume
      persistentVolumeClaim:
        claimName: my-lv-pvc
EOF
```

### Resize Workload Storage

Resizing is the process of expanding the backend storage volume's capacity via the OpenShift GUI or CLI client. LVMS 
supports volume expansion.  It does not support volume shrinking. Refer to OpenShift documentation on 
[Expanding CSI Volumes](https://docs.okd.io/latest/storage/expanding-persistent-volumes.html#expanding-csi-volumes_expanding-persistent-volumes)
for resizing instructions.

### PVC to PVC Cloning

LVMS supports PVC cloning for LVM thin-volumes (thick volumes are not supported).  Cloning is
only allowed within the same namespace; you cannot clone a PVC from one namespace to another.
For more details on PVC to PVC cloning, see the [OpenShift documentation](https://docs.okd.io/latest/storage/container_storage_interface/persistent-storage-csi-cloning.html).

### Volume Snapshotting

> NOTE: Only supported for LVM thin volumes.  LVM does not support cloning or snapshotting of thick LVs. See 
[configuration.md](./configuration.md#lvm-thin-volumes) for information on setting up thin volume provisioning and snapshotting.

> For details on VolumeSnapshot APIs, see [OKD documentation](https://docs.okd.io/latest/storage/container_storage_interface/persistent-storage-csi-snapshots.html).

To avoid data corruption, it is HIGHLY recommended that writes to the volume are halted while the snapshot is being created.  This
In this example, we delete the pod that the source volume is mounted to.  If the pod is managed via a replication controller
(deployment, statefulset, etc), scale the replica count to zero instead of deleting the pod directly.  After snapshotting
is complete, the source PVC can be re-mounted to a new pod.

To create a snapshot, you need a source PVC or VolumeSnapshotContent object backed by a thin LV.  This sample workload
deploys a single Pod and PVC for this use-case.  LVMS only supports `WaitForFirstConsumer` volumeBindingMode, 
which means the storage volume will not be provisioned until a pod is ready to mount it.


```shell
cat <<EOF | oc apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-claim-thin
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: topolvm-provisioner-thin
---
apiVersion: v1
kind: Pod
metadata:
  name: base
spec:
  containers:
  - command:
    - sh
    - -c
    - sleep 1d
    image: nginxinc/nginx-unprivileged:latest
    name: test-container
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
    volumeMounts:
    - mountPath: /vol
      name: test-vol
  securityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
  volumes:
  - name: test-vol
    persistentVolumeClaim:
      claimName: test-claim-thin
EOF
```

Wait for the PVC to bind and the pod to enter a Running state, then write some simple data to the persistent volume.
Execute the following:

`oc exec my-pod -- bash -c 'echo FOOBAR > /data/demo.txt'`

Next, delete the pod to ensure no data is written to the volume during snapshotting.

`oc delete my-pod`

Once the pod has been deleted, it is safe to generate a snapshot.  This is done by creating an instance of a 
VolumeSnapshot API object.

_Snapshot of a PVC_
```shell
cat <<'EOF' | oc apply -f -
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-snap
spec:
  volumeSnapshotClassName: topolvm-snapclass
  source:
    persistentVolumeClaimName: test-claim-thin
```

> It is also possible to create a snapshot from an existing snapshot by specifying the 
VolumeSnapshotContent as the source. See [OKD documentation](https://docs.okd.io/latest/storage/container_storage_interface/persistent-storage-csi-snapshots.html#persistent-storage-csi-snapshots-create_persistent-storage-csi-snapshots) for more information

Wait for the storage driver to finish creating the snapshot with the following command:

`oc wait volumesnapshot/my-snap --for=jsonpath\='{.status.readyToUse}=true'`

Once the volumeSnapshot is `ReadyToUse`, it can be restored as a volume for future PVCs.

### Restoring a Snapshot

Restoring a snapshot is done by specifying the VolumeSnapshot object as the `dataSource` in a PersistentVolumeClaim. The
following workflow demonstrates snapshot restoration.  For demonstration purposes, we also verify the data we wrote to 
source PVC is preserved and restored on the new PVC.

```shell
cat <<'EOF' | oc apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: snapshot-restore
spec:
  accessModes:
  - ReadWriteOnce
  dataSource:
    apiGroup: snapshot.storage.k8s.io
    kind: VolumeSnapshot
    name: my-snap
  resources:
    requests:
      storage: 1Gi
  storageClassName: topolvm-provisioner-thin
---
apiVersion: v1
kind: Pod
metadata:
  name: base
spec:
  containers:
  - command:
    - sh
    - -c
    - sleep 1d
    image: nginxinc/nginx-unprivileged:latest
    name: test-container
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
    volumeMounts:
    - mountPath: /vol
      name: test-vol
  securityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
  volumes:
  - name: test-vol
    persistentVolumeClaim:
      claimName: snapshot-restore
```

Once the new pod enters the Running state, verify that the data we wrote early was cloned to the new volume:

```shell
oc exec base -- cat /data/demo.txt
FOOBAR
```

# LVMS Versioning

LVMS is released at a different cadence that MicroShift and is not couple to the MicroShift version.  This is primarily
because LVMS is a subcomponent of MicroShift and deployed as a workload on the cluster.  The version of LVMS is tracked
by image tag, with only the major version correlating the major MicroShift version.

The LVMS version is not exposed by LVMS itself. For troubleshooting purposes, MicroShift exposes the LVMS version 
via a configmap in the `kube-public` namespace. To get the LVMS version, run the following command:

```shell
$ oc get configmap -n kube-public lvms-version -o jsonpath='{.data.version}'
v4.14.0-10
```