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

If there is only one volume group on the system, LVMS uses it by
default. If there are multiple volume groups, and no configuration
file, LVMS looks for a volume group named `microshift`. If there is no
volume group named `microshift`, LVMS is disabled.

LVMS expects all volume groups to exist prior to launching the
service. If LVMS is configured to use a volume group that does not
exist, the node-controller Pod will fail and enter a CrashLoopBackoff
state.

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
[Expanding CSI Volumes](https://docs.openshift.com/container-platform/4.13/storage/expanding-persistent-volumes.html#expanding-csi-volumes_expanding-persistent-volumes)
for resizing instructions.

### PVC to PVC Cloning

LVMS supports PVC cloning for LVM thin-volumes (thick volumes are not supported).  Cloning is
only allowed within the same namespace; you cannot clone a PVC from one namespace to another.
For more details on PVC to PVC cloning, see the [OpenShift documentation](https://docs.openshift.com/container-platform/4.13/storage/container_storage_interface/persistent-storage-csi-cloning.html).