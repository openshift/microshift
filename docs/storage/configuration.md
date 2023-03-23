# Storage Configuration

LVM backed storage is indirectly configured through `/etc/microshift/lvmd.yaml`. LVMD is a sub-component of the CSI plugin responsible for managing LVM host operations.  An example with default values is provided by MicroShift at `/etc/microshift/lvmd.yaml.default`.  Users may copy this file and rename it to `/etc/microshift/lvmd.yaml` to customize their configuration.

### Default Config

Below is an example of what users will find at `/etc/microshift/lvmd.yaml.default`.  

```yaml
# Unix domain socket endpoint of gRPC
#socket-name: /run/lvmd/lvmd.socket

device-classes:
  # The name of a device-class
  #- name: default

    # The group where this device-class creates the logical volumes
    #volume-group: microshift

    # Storage capacity in GiB to be spared
    #spare-gb: 0

    # A flag to indicate that this device-class is used by default
    #default: true

    # The number of stripes in the logical volume
    #stripe: ""

    # The amount of data that is written to one device before moving to the next device
    #stripe-size: ""

    # Extra arguments to pass to lvcreate, e.g. ["--type=raid1"]
    #lvcreate-options:
      #- ""
```

### Device Classes

The storage configuration must specify at least one device class.  One and only one device class must be set as default.

Multiple device classes may be defined in the `device-classes` array, and they can be a mix of thick and thin volume configurations. E.g.

```yaml
socket-name: /run/topolvm/lvmd.sock
device-classes:
  - name: ssd
    volume-group: ssd-vg
    spare-gb: 0
    default: true
  - name: hdd
    volume-group: hdd-vg
    spare-gb: 0
  - name: striped
    volume-group: multi-pv-vg
    spare-gb: 0
    stripe: 2
    stripe-size: "64"
  - name: raid
    volume-group: raid-vg
    lvcreate-options:
      - --type=raid1
```
Caveats:
- Specifying lvcreate-options is at your own risk and only provided as an escape hatch by the CSI plugin. 
- Setting spare-gb to anything other than 0 is not recommended because it does not behave predictably and usually results in more space being allocated that expected.

### Storage Class

Storage Classes provide the workload layer interface for selecting a given device class.  The following storage class parameters are supported.

- `csi.storage.k8s.io/fstype` selects filesystem type: `xfs` or `ext4` are supported
- `topolvm.io/device-class` maps the backend device class. If not provided, the default device class is assumed.

Multiple storage classes can refer to the same device class.  This allows admins to provide varying sets of parameters for the same backing device class, e.g. to provide xfs and ext4 variants.

MicroShift defines the following as the default storage class.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  annotations:
    storageclass.kubernetes.io/is-default-class: "true" [1]
  name: topolvm-provisioner
parameters:
  "csi.storage.k8s.io/fstype": "xfs" [2]
  "topolvm.io/device-class": "ssd" [3]
provisioner: topolvm.io [4]
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer [5]
allowVolumeExpansion: false [6]
```

1. Denotes this storage class is the cluster's default.  If a PVC does not specify a storage class, this class is assumed. Exclude the annotation if not a default storage class.
2. Which filesystem to provision on the volume. Options are "xfs" and "ext4".
3. Maps this storage class to a device class.
4. Identifies which provisioner should manage this class.
5. Whether to provision the volume before a client pod is present or immediately.  Options are `WaitForFirstConsumer` and `Immediate`. `WaitForFirstConsumer` is recommended to ensure storage is only provisioned for schedulable pods.
6. Specifies whether the storage provide supports volume expansion.  MicroShift's CSI plugin does not support volume expansion, so this field has no effect.