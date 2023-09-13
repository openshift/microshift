# Install MicroShift on RHEL for Edge
To test MicroShift in a setup similar to the production environment, it is necessary to create a RHEL for Edge ISO installer with all the necessary components preloaded on the image.

The procedures described in this document require the following setup:
* A `physical hypervisor host` with the [libvirt](https://libvirt.org/) virtualization platform, to be used for starting virtual machines that run RHEL for Edge OS containing MicroShift binaries
* A `development virtual machine` set up according to the [MicroShift Development Environment](./devenv_setup.md) instructions, to be used for building a RHEL for Edge ISO installer

## Build RHEL for Edge Installer ISO
Log into the `development virtual machine` with the `microshift` user credentials.

Follow the instructions in the [RPM Packages](./devenv_setup.md#rpm-packages) section to create MicroShift RPM packages.

The scripts for building the installer are located in the `scripts/image-builder` subdirectory.

### Prerequisites
Execute the `scripts/image-builder/configure.sh` script to install the tools necessary for building the installer image.

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and save it into the `~/.pull-secret.json` file.

Make sure there is more than 20GB of free disk space necessary for the build artifacts. Run the following command to free the space if necessary.
```bash
~/microshift/scripts/image-builder/cleanup.sh -full
```

Note that the command deletes various user and system data, including:
- The `_output/image-builder/` directory containing image build artifacts
- MicroShift `ostree` server container and all the unused container images
- All the Image Builder jobs are canceled and deleted
- Project-specific Image Builder sources are deleted
- The user `~/.cache` and `/tmp/containers` directory contents are deleted to clean various cache files
- The `/var/cache/osbuild-worker` directory contents are deleted to clean Image Builder cache files

### Building Installer
It is recommended to execute the partial cleanup procedure (without `-full` argument) before each build to reclaim cached disk space from previous builds.
```bash
~/microshift/scripts/image-builder/cleanup.sh
```
Note that the command deletes various user and system data, including:
- MicroShift `ostree` server container
- All the Image Builder jobs are canceled
- Project-specific Image Builder sources are deleted
- The `/var/cache/osbuild-worker` directory contents are deleted to clean Image Builder cache files

Run the build script without arguments to see its usage.
```bash
$ ~/microshift/scripts/image-builder/build.sh
Usage: build.sh <-pull_secret_file path_to_file> [OPTION]...

  -pull_secret_file path_to_file
          Path to a file containing the OpenShift pull secret, which can be
          obtained from https://console.redhat.com/openshift/downloads#tool-pull-secret

Optional arguments:
  -microshift_rpms path_or_URL
          Path or URL to the MicroShift RPM packages to be included
          in the image (default: _output/rpmbuild/RPMS)
  -custom_rpms /path/to/file1.rpm,...,/path/to/fileN.rpm
          Path to one or more comma-separated RPM packages to be
          included in the image (default: none)
  -embed_containers
          Embed the MicroShift container dependencies in the image
  -ostree_server_url URL
          URL of the ostree server (default: file:///var/lib/ostree-local/repo)
  -build_edge_commit
          Build edge commit archive instead of an ISO image. The
          archive contents can be used for serving ostree updates.
  -lvm_sysroot_size num_in_MB
          Size of the system root LVM partition. The remaining
          disk space will be allocated for data (default: 10240)
  -authorized_keys_file path_to_file
          Path to an SSH authorized_keys file to allow SSH access
          into the default 'redhat' account
  -open_firewall_ports port1[:protocol1],...,portN[:protocolN]
          One or more comma-separated ports (optionally with protocol)
          to be allowed by firewall (default: none)
  -mirror_registry_host host[:port]
          Host and optionally port of the mirror container registry to
          be used by the container runtime when pulling images. The connection
          to the mirror is configured as unsecure unless a CA trust certificate
          is specified using -ca_trust_files parameter
  -ca_trust_files /path/to/file1.pem,...,/path/to/fileN.pem
          Path to one or more comma-separated public certificate files
          to be included in the image at the /etc/pki/ca-trust/source/anchors
          directory and installed using the update-ca-trust utility
  -prometheus
          Add Prometheus process exporter to the image. See
          https://github.com/ncabatoff/process-exporter for more information
```

Continue by running the build script with the pull secret file argument and wait until build process is finished. It may take over 30 minutes to complete a full build cycle.
```bash
~/microshift/scripts/image-builder/build.sh -pull_secret_file ~/.pull-secret.json
```
The script performs the following tasks:
- Check for minimum 20GB of available disk space
- Set up a local MicroShift RPM repository using a local RPM build or a remote URL (if specified in the command line) as a source
- Set up a local OpenShift RPM repository using public OpenShift repositories necessary for CRI-O and OpenShift client package installation
- Configure Image Builder to use the local MicroShift and OpenShift RPM repositories for image builds
- Run Image Builder to create an edge container image containing all the MicroShift binaries and dependencies
- Start a local `ostree` server with the above image
- Run Image Builder to create an edge installer image comprised of RHEL for Edge OS and MicroShift binaries
- Rebuild the installer image with the `kickstart.ks` file for performing various OS setup when the host is first booted
- Perform partial cleanup procedure to reclaim cached disk space

The artifact of the build is the `_output/image-builder/microshift-installer-${VERSION}.${ARCH}.iso` bootable RHEL for Edge OS image.

### Disk Partitioning
The `kickstart.ks` file is configured to partition the main disk using `Logical Volume Manager` (LVM). Such partitioning is required for the data volume to be utilized by the MicroShift CSI driver and it allows for flexible file system customization if the disk space runs out.

By default, the following partition layout is created and formatted with the `XFS` file system:
* BIOS boot partition (1MB)
  * It is required to boot ISO to systems with legacy BIOS while using GPT as the partitioning scheme.
* EFI partition with EFI file system (200MB)
* Boot partition is allocated on a 1GB volume
* The rest of the disk is managed by the `LVM` in a single volume group named `rhel`
  * System root partition is allocated on a 10GB volume (minimal recommended size for a root partition)
  * The remainder of the volume group will be used by the CSI driver for storing data (no need to format and mount it)

> The swap partition is not created as it is not required by MicroShift.

The `scripts/image-builder/build.sh` script provides for the optional `-lvm_sysroot_size` command line parameter allowing to increase the system root partition size from the default 10GB.
> The system root partition size should be specified in megabytes.

As an example, a 20GB disk is partitioned in the following manner by default.
```
$ lsblk /dev/sda
NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
sda             8:0    0   20G  0 disk
├─sda1          8:1    0    1M  0 part
├─sda2          8:2    0  200M  0 part /boot/efi
├─sda3          8:3    0  800M  0 part /boot
└─sda4          8:4    0   19G  0 part
  └─rhel-root 253:0    0   10G  0 lvm  /sysroot

$ sudo vgdisplay -s
  "rhel" <19.02 GiB [10.00 GiB  used / <9.02 GiB free]
```

> Unallocated disk space of 9GB size remains in the `rhel` volume group to be used by the CSI driver.

### Offline Containers
The `scripts/image-builder/build.sh` script supports a special mode for including the MicroShift container image dependencies into the generated ISO. When the container images required by MicroShift are preinstalled, `CRI-O` does not attempt to pull them when MicroShift service is first started, saving network bandwidth and avoiding external network connections.

Image Builder needs to be configured to use the pull secret for downloading container images during the build procedure.
Run the following commands to reuse the `CRI-O` pull secret file.
```bash
sudo mkdir -p /etc/osbuild-worker
sudo ln -s /etc/crio/openshift-pull-secret /etc/osbuild-worker/pull-secret.json
sudo tee /etc/osbuild-worker/osbuild-worker.toml &>/dev/null <<EOF
[containers]
auth_file_path = "/etc/osbuild-worker/pull-secret.json"
EOF
```

Proceed by running the build script with the `-embed_containers` argument to include the dependent container images into the generated ISO.
```bash
~/microshift/scripts/image-builder/build.sh -pull_secret_file ~/.pull-secret.json -embed_containers
```

> If user workloads depend on additional container images, they need to be included by the user separately.

When executed in this mode, the `scripts/image-builder/build.sh` script performs an extra step to append the list of the MicroShift container images to the blueprint so that they are installed when the operating system boots for the first time. The list of these images can be obtained by the following command.
```bash
jq -r '.images | .[]' ~/microshift/assets/release/release-$(uname -m).json
```

> See [Embedding MicroShift Container Images for Offline Deployments](../user/howto_offline_containers.md) for more information.

## Install MicroShift for Edge
Log into the `physical hypervisor host` using your user credentials. The remainder of this section describes how to install a virtual machine running RHEL for Edge OS containing MicroShift binaries.

Start by copying the installer image from the `development virtual machine` to the host file system.
```bash
sudo scp microshift@microshift-dev:/home/microshift/microshift/_output/image-builder/microshift-installer-*.$(uname -m).iso /var/lib/libvirt/images/
```

Run the following commands to create a virtual machine using the installer image.
```bash
VMNAME="microshift-edge"
NETNAME="default"
sudo bash -c " \
cd /var/lib/libvirt/images/ && \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 3072 \
    --disk path=./${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --cdrom ./microshift-installer-*.$(uname -m).iso \
    --noautoconsole \
    --wait \
"
```

Watch the OS console to see the progress of the installation, waiting until the machine is rebooted and the login prompt appears.

Note that it may be more convenient to access the machine using SSH. Run the following command to get its IP address and use it to remotely connect to the system.
```bash
sudo virsh domifaddr microshift-edge
```

Log into the system using `redhat:redhat` credentials and run the following commands to configure MicroShift access.
```bash
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
```

Finally, check if MicroShift is up and running by executing `oc` commands.
```bash
oc get cs
oc get pods -A
```

### The `ostree` Update Server

**Default Configuration**

The default ISO image is configured to use the local `/var/lib/ostree-local`
directory as the source for `ostree` updates. This directory contains an empty
`ostree` update repository.

Log into the MicroShift server using `redhat:redhat` credentials and run the
following commands to check the local `ostree` server configuration.

```bash
$ ostree remote list
edge

$ ostree remote show-url edge
file:///var/lib/ostree-local/repo

$ ostree remote summary edge
Repository Mode (ostree.summary.mode): archive-z2
Last-Modified (ostree.summary.last-modified): 2023-04-28T04:55:59-04
Has Tombstone Commits (ostree.summary.tombstone-commits): No
ostree.summary.indexed-deltas: true
```

> The default `ostree` local repository is empty and it does not contain any
> commit revisions to be installed.

**Custom Server**

This default behavior can be overriden by specifying the `-ostree_server_url`
command line argument when running the `scripts/image-builder/build.sh` script.
The URL parameter of this argument should point to a custom `ostree` server to
be used for installing updates on an existing image.

Log into the MicroShift server using `redhat:redhat` credentials and run the
following commands to check the custom `ostree` server configuration and
install updates.

```bash
$ ostree remote list
edge

$ ostree remote show-url edge
<YOUR_OSTREE_SERVER_URL>

$ ostree remote summary edge
<YOUR_OSTREE_REPO_SUMMARY_AND_COMMIT_REV>

$ sudo rpm-ostree deploy <YOUR_OSTREE_COMMIT_REV>
...
...
# Reboot the system for the update to become active
```

**Local Updates**

When the local `ostree` repository is configured for the `ostree` updates, users
can overwrite the contents of the `/var/lib/ostree-local` directory and install
the updates locally.

One of the techniques for generating `ostree` updates is supported by the
`scripts/image-builder/build.sh` script. When the `-build_edge_commit` command
line argument is specified, the script builds an edge commit archive instead of
an ISO image.

```bash
~/microshift/scripts/image-builder/build.sh -pull_secret_file ~/.pull-secret.json -build_edge_commit
...
...
# Edge commit created
The contents of the archive can be used for serving ostree updates:
/home/microshift/microshift/_output/image-builder/microshift-0.0.1-commit.tar

# Done
```

> Optionally specify additional `-microshift_rpms`, `-custom_rpms`, or
> `-embed_containers` arguments to customize the edge commit archive contents.

Copy the produced `microshift-0.0.1-commit.tar` archive to the
MicroShift server and unpack it at the `/var/lib/ostree-local` directory,
deleting the previous contents.

Run the following commands to generate the update summary, check the new
`ostree` commit revision and install the update.

```bash
$ sudo ostree summary --repo /var/lib/ostree-local/repo --update

$ ostree remote summary edge
* rhel/9/x86_64/edge
    Latest Commit (17.3 kB):
      8982a0afae721e55cd75954f23760c77a27b4cfd7d7c5bf39c3115c9d134cec6
    Version (ostree.commit.version): 9.2
    Timestamp (ostree.commit.timestamp): 2023-04-28T06:41:55-04

Repository Mode (ostree.summary.mode): archive-z2
Last-Modified (ostree.summary.last-modified): 2023-04-28T06:58:16-04
Has Tombstone Commits (ostree.summary.tombstone-commits): No
ostree.summary.indexed-deltas: true

$ sudo rpm-ostree deploy 8982a0afae721e55cd75954f23760c77a27b4cfd7d7c5bf39c3115c9d134cec6
...
...
# Reboot the system for the update to become active
```

### Offline Mode
It may sometimes be necessary to install a virtual machine that does not have access to the Internet. For instance, such a configuration can be used for testing the [Offline Containers](#offline-containers) or any other setup that needs to work without the Internet access.

Create a new isolated `libvirt` network configuration by running the following commands.
```bash
NETCONFIG_FILE=$(mktemp /tmp/isolated-network-XXXXX.xml)
cat > $NETCONFIG_FILE <<EOF
<network>
  <name>isolated</name>
  <forward mode='none'/>
  <ip address='192.168.111.1' netmask='255.255.255.0' localPtr='yes'>
    <dhcp>
      <range start='192.168.111.100' end='192.168.111.254'/>
    </dhcp>
  </ip>
</network>
EOF
sudo virsh net-define    $NETCONFIG_FILE
sudo virsh net-start     isolated
sudo virsh net-autostart isolated
rm -f $NETCONFIG_FILE
```

Follow the instruction in the [Install MicroShift for Edge](#install-microshift-for-edge) section to install a new virtual machine using the `isolated` network configuration.
> When running the `virt-install` command, specify the `--network network=isolated,model=virtio` option to select the `isolated` network configuration.

After the virtual machine is created, log into the system and verify that the Internet is not accessible.
```bash
$ curl -I redhat.com
curl: (6) Could not resolve host: redhat.com
```

Make sure that `CRI-O` has access to the container images required by MicroShift.
```bash
$ sudo crictl images | egrep -c 'openshift|redhat'
13
```

Finally, wait until all the MicroShift pods are up and running.
```bash
watch sudo $(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A
```
