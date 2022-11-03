# MicroShift Storage Plugin Overview

> **IMPORTANT!** The default ODF-LVM configuration is intended to match the developer environment described in [MicroShift Development Environment on RHEL 8](./devenv_rhel8.md). See section **[Configuring ODF-LVM](#Configuring-ODF-LVM)** for guidance on configuring ODF-LVM for your environment.

MicroShift enables dynamic storage provisioning out of the box with the ODF-LVM CSI plugin. This plugin is a downstream
Red Hat fork of TopoLVM. This provisioner will create a new LVM logical volume in the `rhel` volume group for each
PersistenVolumeClaim(PVC), and make these volumes available to pods. For more information on ODF-LVM, visit the repo's
[README](https://github.com/red-hat-storage/topolvm).

## Design

ODF-LVM manifests can be found in [microshift/assets/components/odf-lvm](../assets/components/odf-lvm). For an overview
of ODF-LVM architecture, see the [design doc](https://github.com/red-hat-storage/topolvm/blob/main/docs/design.md).

## Deployment

ODF-LVM is deployed onto the cluster in the `openshift-storage` namespace, after MicroShift
boots. [StorageCapacity tracking](https://kubernetes.io/docs/concepts/storage/storage-capacity/) is used to ensure that
Pods with an ODF-LVM PVC are not scheduled if the requested storage is greater than the volume group's remaining free
storage.

### Configuring ODF-LVM

MicroShift supports pass-through of users' ODF-LVM configuration and allows users to specify custom volume-groups, thin volume provisioning parameters, reserved unallocated volume-group space, among others.  This file can be specified anytime.  If MicroShift is already running, simply restart it to pick up the latest config.

This configuration is specific to the LVMD runtime and informs how it interacts with the host. LVMD is a subcomponent of ODF-LVM, and in the case of MicroShift, deployed automatically as a DaemonSet.

#### Specification

Full documentation of the config spec can be found at [github.com/red-hat-storage/topolvm/blob/v4.11.0/docs/lvmd.md](https://github.com/red-hat-storage/topolvm/blob/v4.11.0/docs/lvmd.md).

#### Path

The user provided lvmd config should be written to the same directory as the MicroShift config.  If a MicroShift config
doesn't exist, MicroShift will assume default lvmd values. These paths will be checked for the config, depending on the user MicroShift
is run as.

1. User config dir: `~/.microshift/lvmd.yaml`
2. Global config dir: `/etc/microshift/lvmd.yaml`

## System Requirements

### Volume Group Name

The default integration of ODF-LVM assumes a volume-group named `rhel`. ODF-LVM's node-controller expects that volume
group to exist prior to launching the service. If the volume group does not exist, the node-controller will fail to
start and enter a CrashLoopBackoff state.

### Volume Size Increments

ODF-LVM provisions storage in increments of 1Gb. Storage requests are rounded up to the near gigabyte. When a volume
group's capacity is less than 1Gb, the PersistentVolumeClaim will register a `ProvisioningFailed` event:

```shell
  Warning  ProvisioningFailed    3s (x2 over 5s)  topolvm.cybozu.
  com_topolvm-controller-858c78d96c-xttzp_0fa83aef-2070-4ae2-bcb9-163f818dcd9f failed to provision volume with 
  StorageClass "topolvm-provisioner": rpc error: code = ResourceExhausted desc = no enough space left on VG: 
  free=(BYTES_INT), requested=(BYTES_INT)
```

## Usage

ODF-LVM's [StorageClass](../assets/components/odf-lvm/topolvm_default-storage-class.yaml) is deployed with a default
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
