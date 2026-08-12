# Install MicroShift on RHEL for Edge
To test MicroShift in a setup similar to the production environment, it is necessary to create a RHEL for Edge ISO installer with all the necessary components preloaded on the image.

The official [Embedding in a RHEL for Edge image](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/embedding_in_a_rhel_for_edge_image/microshift-embed-in-rpm-ostree) documentation covers the full procedure for building an installer ISO from released MicroShift RPMs. This document describes a modified workflow for building from **locally compiled** RPMs, which is necessary when testing changes that have not been released.

The procedures described in this document require the following setup:
* A `physical hypervisor host` running RHEL with the [libvirt](https://libvirt.org/) virtualization platform and at least 50GB of free disk space
  * Packages: `libvirt`, `virt-install`, `virt-viewer`, `qemu-kvm`
* A `development virtual machine` set up according to the [MicroShift Development Environment](./devenv_setup.md) instructions, to be used for building a RHEL for Edge ISO installer
  * An active RHEL subscription is required for building images

## Build RHEL for Edge Installer ISO

Log into the `development virtual machine` with the `microshift` user credentials.

### Prerequisites

Execute the `scripts/devenv-builder/configure-composer.sh` script to install `osbuild-composer` and its dependencies.
```bash
~/microshift/scripts/devenv-builder/configure-composer.sh
```

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and save it into the `~/.pull-secret.json` file.

Make sure there is more than 20GB of free disk space necessary for the build artifacts. Run the following command to free the space if necessary.
```bash
~/microshift/scripts/devenv-builder/cleanup-composer.sh -full
```

### Build MicroShift RPMs

Follow the instructions in the [RPM Packages](./devenv_setup.md#rpm-packages) section or run:
```bash
cd ~/microshift
make rpm
```

The RPMs are placed under `_output/rpmbuild/RPMS/`.

### Create a Local RPM Repository

Create a local repository from the built RPMs so that `osbuild-composer` can resolve them as a package source. This replaces the released MicroShift RPMs that the [official procedure](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/embedding_in_a_rhel_for_edge_image/microshift-embed-in-rpm-ostree#adding-microshift-repos-image-builder_microshift-embed-in-rpm-ostree) obtains from CDN.

```bash
BUILDDIR=~/microshift/_output/image-builder
mkdir -p "${BUILDDIR}/microshift-local"
cp ~/microshift/_output/rpmbuild/RPMS/*/*.rpm "${BUILDDIR}/microshift-local/"
createrepo "${BUILDDIR}/microshift-local"
chmod -R a+rX "${BUILDDIR}/microshift-local"
```

Register it with `osbuild-composer`:
```bash
cat <<EOF | sudo tee /tmp/microshift-local.toml
id = "microshift-local"
name = "MicroShift Local RPM Repo"
type = "yum-baseurl"
url = "file://${BUILDDIR}/microshift-local/"
check_gpg = false
check_ssl = false
system = false
EOF

sudo composer-cli sources add /tmp/microshift-local.toml
```

### Build the Image

With the local RPM source registered, follow the official documentation starting from [Adding MicroShift repositories to image builder](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/embedding_in_a_rhel_for_edge_image/microshift-embed-in-rpm-ostree#adding-microshift-repos-image-builder_microshift-embed-in-rpm-ostree) through [Download the ISO and prepare it for use](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/embedding_in_a_rhel_for_edge_image/microshift-embed-in-rpm-ostree#microshift-download-iso-prep-for-use_microshift-embed-in-rpm-ostree). The procedure is identical — `osbuild-composer` will resolve `microshift` packages from the local repository instead of CDN.

Use `rhel/9/x86_64/edge` as the ostree ref in all `composer-cli compose start-ostree --ref` commands. This must match the ref in the [`microshift-edge.ks`](../config/microshift-edge.ks) kickstart.

### Disk Partitioning
The [`microshift-edge.ks`](../config/microshift-edge.ks) file is configured to partition the main disk using `Logical Volume Manager` (LVM). Such partitioning is required for the data volume to be utilized by the MicroShift CSI driver and it allows for flexible file system customization if the disk space runs out.

By default, the following partition layout is created. The `/boot` and root partitions use the `XFS` file system:
* EFI System Partition with FAT file system (600MB)
* Boot partition is allocated on a 1GB volume
* The rest of the disk is managed by the `LVM` in a single volume group named `rhel`
  * System root partition is allocated on a 10GB volume (minimal recommended size for a root partition)
  * The remainder of the volume group will be used by the CSI driver for storing data (no need to format and mount it)

> The swap partition is not created as it is not required by MicroShift.
> The system root partition size should be specified in megabytes.

As an example, a 20GB disk is partitioned in the following manner by default.
```
$ lsblk /dev/vda
NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
vda           252:0    0   20G  0 disk
├─vda1        252:1    0  600M  0 part /boot/efi
├─vda2        252:2    0    1G  0 part /boot
└─vda3        252:3    0 18.4G  0 part
  └─rhel-root 253:0    0   10G  0 lvm  /sysroot

$ sudo vgdisplay -s
  "rhel" 18.41 GiB [10.00 GiB  used / 8.41 GiB free]
```

> Unallocated disk space of 8GB size remains in the `rhel` volume group to be used by the CSI driver.

## Install MicroShift for Edge

Log into the `physical hypervisor host` using your user credentials. The remainder of this section describes how to install a virtual machine running RHEL for Edge OS containing MicroShift binaries.

Start by copying the installer image and kickstart from the `development virtual machine` to the host file system. Replace `<dev-vm-ip>` with the IP address of your development VM (run `sudo virsh domifaddr <vm-name>` on the hypervisor to find it).
```bash
scp microshift@<dev-vm-ip>:/home/microshift/microshift/_output/image-builder/${BUILDID}-installer.iso ~/
scp microshift@<dev-vm-ip>:/home/microshift/microshift/docs/config/microshift-edge.ks ~/
```

Run the following commands to create a virtual machine using the installer image. The `--boot uefi` flag is required because the ostree image uses `bootupd` for bootloader management, which only supports UEFI. The `--location` flag extracts the installer kernel from the ISO for direct boot, and `--initrd-inject` embeds the kickstart into the installer initrd.
```bash
VMNAME="microshift-edge"
NETNAME="default"
ISOFILE="${HOME}/${BUILDID}-installer.iso"

sudo virt-install \
    --name "${VMNAME}" \
    --vcpus 2 \
    --memory 4096 \
    --boot uefi \
    --disk path="${HOME}/${VMNAME}.qcow2,size=50" \
    --network network="${NETNAME}",model=virtio \
    --events on_reboot=restart \
    --location "${ISOFILE}" \
    --initrd-inject "${HOME}/microshift-edge.ks" \
    --extra-args "inst.ks=file:/microshift-edge.ks" \
    --noautoconsole \
    --wait
```

Watch the OS console to see the progress of the installation, waiting until the machine is rebooted and the login prompt appears.

Note that it may be more convenient to access the machine using SSH. Run the following command to get its IP address and use it to remotely connect to the system.
```bash
sudo virsh domifaddr "${VMNAME}"
```

Log into the system using `redhat:redhat` credentials (as configured in [`microshift-edge.ks`](../config/microshift-edge.ks)) and run the following commands to configure MicroShift access.
```bash
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
chmod go-r ~/.kube/config
```

Verify that MicroShift is up and running.
```bash
oc get pods -A
```
