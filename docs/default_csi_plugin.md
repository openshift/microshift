# MicroShift Default CSI Plugin

MicroShift enables dynamic storage provisioning out of the box with the ODF-LVM CSI plugin. This plugin is downstream
Red Hat distribution of TopoLVM. This feature is not currently configurable but will be in the near future. For more
information on ODF-LVM, visit the repo's [README](https://github.com/red-hat-storage/topolvm).

## Design

ODF-LVM manifests can be found in [microshift/assets/components/odf-lvm](../assets/components/odf-lvm). For an overview
of ODF-LVM architecture, see the [design doc](https://github.com/red-hat-storage/topolvm/blob/main/docs/design.md).

## Deployment

ODF-LVM is deployed onto the cluster in the `openshift-storage` namespace, after MicroShift boots. In order to reduce
MicroShift's storage footprint and cluster overhead,
[StorageCapacity tracking](https://kubernetes.io/docs/concepts/storage/storage-capacity/) is used in lieu of the
plugin's extended-scheduler.

## System Requirements

### Volume Group Name

LVMD, a component of ODF-LVM, requires that volume groups specified in
the [device-class configuration](../assets/components/odf-lvm/topolvm-lvmd-config_configmap_v1.yaml) exist at the
runtime. If the volume group doesn't exist, the LVMD container will fail to start and enter a CrashLoopBackoff state.
The initial integration of ODF-LVM assumes volume-group named `rhel` exists. Unallocated space must exist on the volume
group if the plugin is to dynamically provision logical volumes on it.

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

ODF-LVM's [StorageClass](../assets/components/odf-lvm/topolvm_default-storage-class.yaml) is deployed by with a default
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
      storage: 1Gi
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