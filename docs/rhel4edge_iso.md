# Install MicroShift on RHEL for Edge
To test MicroShift in a setup similar to the production environment, it is necessary to create a RHEL for Edge ISO installer with all the necessary components preloaded on the image.

## Build RHEL for Edge Installer ISO
Log into the development virtual machine with the `microshift` user credentials.
> The development machine configuration guidelines can be found in the [MicroShift Development Environment on RHEL 8.x](./devenv_rhel8.md) document.

Follow the instructions in the [RPM Packages](./devenv_rhel8.md#rpm-packages) section to create MicroShift RPM packages.

The scripts for building the installer are located in the `scripts/image-builder` subdirectory.

### Prerequisites
Execute the `scripts/image-builder/configure.sh` script to install the tools necessary for building the installer image.

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and save it into the `~microshift/.pull-secret.json` file.

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
~/microshift/scripts/image-builder/build.sh
Usage: build.sh <-pull_secret_file path_to_file> [OPTION]...

  -pull_secret_file path_to_file
          Path to a file containing the OpenShift pull secret, which
          can be downloaded from https://console.redhat.com/openshift/downloads#tool-pull-secret

Optional arguments:
  -microshift_rpms path_or_URL
          Path or URL to the MicroShift RPM packages to be included
          in the image (default: _output/rpmbuild/RPMS)
  -custom_rpms /path/to/file1.rpm,...,/path/to/fileN.rpm
          Path to one or more comma-separated RPM packages to be
          included in the image (default: none)
  -ostree_server_name name_or_ip
          Name or IP address and optionally port of the ostree
          server (default: 127.0.0.1:8080)
  -lvm_sysroot_size num_in_MB
          Size of the system root LVM partition. The remaining
          disk space will be allocated for data (default: 8192)
  -authorized_keys_file
          Path to an SSH authorized_keys file to allow SSH access
          into the default 'redhat' account
```

Continue by running the build script with the pull secret file argument and wait until build process is finished. It may take over 30 minutes to complete a full build cycle.
```bash
~/microshift/scripts/image-builder/build.sh -pull_secret_file ~/.pull-secret.json
```
The script performs the following tasks:
- Check for minimum 10GB of available disk space
- Set up a local MicroShift RPM repository using a local RPM build or a remote URL (if specified in the command line) as a source
- Set up a local OpenShift RPM repository using public OpenShift repositories necessary for CRI-O and OpenShift client package installation
- Configure Image Builder to use the local MicroShift and OpenShift RPM repositories for image builds
- Run Image Builder to create an edge container image containing all the MicroShift binaries and dependencies
- Start a local `ostree` server with the above image
- Run Image Builder to create an edge installer image comprised of RHEL for Edge OS and MicroShift binaries
- Rebuild the installer image with the `kickstart.ks` file for performing various OS setup when the host is first booted
- Perform partial cleanup procedure to reclaim cached disk space

The artifact of the build is the `scripts/image-builder/_builds/microshift-installer-${VERSION}.${ARCH}.iso` bootable RHEL for Edge OS image.

### Disk Partitioning
The `kickstart.ks` file is configured to partition the main disk using `Logical Volume Manager` (LVM). Such parititioning is required for the data volume to be utilized by the MicroShift CSI driver and it allows for flexible file system customization if the disk space runs out.

By default, the following partition layout is created and formatted with the `XFS` file system:
* Boot partition is allocated on a 1GB volume
* The rest of the disk is managed by the `LVM` in a single volume group named `rhel`
  * System root partition is allocated on a 8GB volume (minimal recommended size for a root partition)
  * The remainder of the volume group will be used by the CSI driver for storing data (no need to format and mount it)

> The swap partition is not created as it is not required by MicroShift.

The `scripts/image-builder/build.sh` script provides for the optional `-lvm_sysroot_size` command line parameter allowing to increase the system root partition size from the default 8GB.
> The system root partition size should be specified in megabytes.

As an example, a 20GB disk is partitioned in the following manner by default.
```
$ lsblk /dev/sda
NAME          MAJ:MIN RM SIZE RO TYPE MOUNTPOINT
sda             8:0    0  20G  0 disk
├─sda1          8:1    0   1G  0 part /boot
└─sda2          8:2    0  19G  0 part
  └─rhel-root 253:0    0   8G  0 lvm  /sysroot

$ sudo vgdisplay -s
  "rhel" <19.00 GiB [8.00 GiB  used / <11.00 GiB free]
```

> Unallocated disk space of 14GB size remains in the `rhel` volume group to be used by the CSI driver.

### Offline Containers
The `scripts/image-builder/build.sh` script supports a special mode for including user-specific RPM files into the generated ISO. The remainder of this section demonstrates how to generate container image RPMs for MicroShift and include them in the installation ISO.

When the container images required by MicroShift are preinstalled, `CRI-O` does not attempt to pull them when MicroShift service is first started, saving network bandwidth and avoiding external network connections.
>If user workloads depend on additional container images, the respective RPMs need to be created by the user separately.

Start by running the `packaging/rpm/make-microshift-images-rpm.sh` script without arguments to see its usage.
```bash
$ ~/microshift/packaging/rpm/make-microshift-images-rpm.sh
Usage:
   make-microshift-images-rpm.sh rpm  <pull_secret> <architectures> <rpm_mock_target>
   make-microshift-images-rpm.sh srpm <pull_secret> <architectures>
   make-microshift-images-rpm.sh copr <pull_secret> <architectures> <copr_repo>

pull_secret:     Path to a file containing the OpenShift pull secret
architectures:   One or more RPM architectures
rpm_mock_target: Target for building RPMs inside a chroot (e.g. 'rhel-8-x86_64')
copr_repo:       Target Fedora Copr repository name (e.g. '@redhat-et/microshift-containers')

Notes:
 - The OpenShift pull secret can be downloaded from https://console.redhat.com/openshift/downloads#tool-pull-secret
 - Use 'x86_64:amd64' or 'aarch64:arm64' as an architecture value
 - See /etc/mock/*.cfg for possible RPM mock target values
```

Run the script in the `rpm` mode to pull the images required by MicroShift and generate the RPMs including those image data.
```bash
~/microshift/packaging/rpm/make-microshift-images-rpm.sh rpm ~/.pull-secret.json x86_64:amd64 rhel-8-x86_64
```

If the procedure runs successfully, the RPM artifacts can be found in the `paack-result` directory.
```bash
$ find ~/microshift/paack-result -name microshift-containers-\*.rpm
/home/microshift/microshift/paack-result/microshift-containers-4.10.18-1.x86_64.rpm
/home/microshift/microshift/paack-result/microshift-containers-4.10.18-1.src.rpm
```

Finally, run the build script with the `-custom_rpms` argument to include the specified container image RPMs into the generated ISO.
```bash
CONTAINERS_RPM=$(find ~/microshift/paack-result -name microshift-containers-\*.$(uname -i).rpm)
~/microshift/scripts/image-builder/build.sh -pull_secret_file ~/.pull-secret.json -custom_rpms $CONTAINERS_RPM
```
> If user-specific container images need to be included into the ISO, multiple comma-separated RPM files can be specified as the `-custom_rpms` argument value.

When executed in this mode, the `scripts/image-builder/build.sh` script performs an extra step to set up a custom RPM repository with one or more specified RPM files as a source. These RPMs are then appended to the blueprint so that they are installed when the operating system boots for the first time.

## Install MicroShift for Edge
Log into the host machine using your user credentials. The remainder of this section describes how to install a virtual machine running RHEL for Edge OS containing MicroShift binaries.

Start by copying the installer image from the development virtual machine to the host file system.
```bash
sudo scp microshift@microshift-dev:/home/microshift/microshift/scripts/image-builder/_builds/microshift-installer-*.$(uname -i).iso /var/lib/libvirt/images/
```

Run the following commands to create a virtual machine using the installer image.
```bash
VMNAME="microshift-edge"
VERSION=4.10.18
sudo -b bash -c " \
cd /var/lib/libvirt/images/ && \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 3096 \
    --disk path=./${VMNAME}.qcow2,size=20 \
    --network network=default,model=virtio \
    --os-type generic \
    --events on_reboot=restart \
    --cdrom ./microshift-installer-${VERSION}.$(uname -i).iso \
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

### Offline Mode
It may sometimes be necessary to install a virtual machine that does not have access to the Internet. For instance, such a configuration can be used for testing the [Offline Containers](#offline-containers) or any other setup that needs to work without the Internet access.

Create a new isolated `libvirt` network configuration by running the following commands.
```bash
NETCONFIG_FILE=/tmp/isolated-network.xml
cat > $NETCONFIG_FILE <<EOF
<network>
  <name>isolated</name>
  <dns>
    <host ip='192.168.100.1'>
      <hostname>gateway</hostname>
    </host>
  </dns>
  <ip address='192.168.100.1' netmask='255.255.255.0' localPtr='yes'>
    <dhcp>
      <range start='192.168.100.10' end='192.168.100.20'/>
    </dhcp>
  </ip>
</network>
EOF
sudo virsh net-define    $NETCONFIG_FILE
sudo virsh net-start     isolated
sudo virsh net-autostart isolated
```

Follow the instruction in the [Install MicroShift for Edge](#install-microshift-for-edge) section to install a new virtual machine using the `isolated` network configuration.
> When running the `virt-install` command, specify the `--network network=isolated,model=virtio` option to select the `isolated` network configuration.

After the virtual machine is created, log into the system and verify that the Internet is not accessible.
```bash
$ curl -I redhat.com
curl: (6) Could not resolve host: redhat.com
```

Make sure that `CRI-O` has access to all the container images required by MicroShift.
```bash
$ sudo crictl images
IMAGE                                            TAG                 IMAGE ID            SIZE
k8s.gcr.io/pause                                 3.6                 6270bb605e12e       690kB
quay.io/coreos/flannel                           v0.14.0             8522d622299ca       68.9MB
quay.io/microshift/flannel-cni                   v0.14.0             4324dc7a1ffa5       8.12MB
quay.io/openshift-release-dev/ocp-v4.0-art-dev   <none>              334363f37666a       401MB
quay.io/openshift-release-dev/ocp-v4.0-art-dev   <none>              eb9d5c9681cd5       376MB
quay.io/openshift-release-dev/ocp-v4.0-art-dev   <none>              60f52af9fc4ba       413MB
quay.io/openshift-release-dev/ocp-v4.0-art-dev   <none>              ef1c6b04ebe2a       415MB
quay.io/openshift-release-dev/ocp-v4.0-art-dev   <none>              a538d5965f4fc       458MB
```

Finally, check if MicroShift is up and running by executing `oc` commands.
```bash
oc get cs
oc get pods -A
```
