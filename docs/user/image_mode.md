# Image Mode for MicroShift (from source)

Image mode is an approach to operating system deployment that lets you build,
deploy, and manage Red Hat Enterprise Linux as a bootable container (`bootc`)
image. Such an image uses standard OCI/Docker containers as a transport and
delivery format for base operating system updates. A `bootc` image includes a
Linux kernel, which is used to boot.

MicroShift build and deployment procedures can utilize bootable containers to
benefit from this technology.

> See [Image mode for Red Hat Enterprise Linux](https://developers.redhat.com/products/rhel-image-mode/overview)
> for more information.

> **Source of truth:**<br>
> The [Installing with RHEL image mode](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/installing_with_image_mode_for_rhel/microshift-about-rhel-image-mode)
> chapter of the Red Hat build of MicroShift documentation is the authoritative
> reference for building, publishing, and installing image mode systems using
> **released** MicroShift RPMs.
>
> This document does **not** duplicate that content. It supplements it for users
> and contributors who need to deploy MicroShift built **from source** (i.e. from
> this repository) rather than from the released `rhocp` repositories. Only the
> image build differs; once you have a source-built image, the publish, Kickstart,
> and virtual machine steps are identical and are linked below.

> **Note:**<br>
> The openshift-docs links in this document are pinned to version `4.19`. Use the
> version selector on those pages to match the MicroShift release you are working
> with.

The procedures described below require the following setup:
* A `RHEL 9.8 host` with an active Red Hat subscription for building MicroShift
`bootc` images. For development purposes, you can use the [Red Hat Developer subscription](https://developers.redhat.com/products/rhel/download),
which is free of charge.
* A `hypervisor host` with a virtualization technology that supports RHEL. In
this documentation, [libvirt](https://libvirt.org/) virtualization is used as
an example.
* A `remote registry` (e.g. `quay.io`) for storing and accessing `bootc` images

## Build a MicroShift bootc image from source

Unlike the released flow documented in openshift-docs — which installs MicroShift
from the `rhocp` and `fast-datapath` repositories — a source build installs the
MicroShift RPMs that you compile from this repository into the `bootc` image.

Log into the `RHEL 9.8 host` using the user credentials that have `sudo`
permissions configured, and clone this repository.

### Build the MicroShift RPMs

Build the MicroShift RPMs from the current source tree by running the following
command at the repository root:

```bash
make rpm
```

> Use `make rpm-podman` to build the RPMs inside a container if the host is not
> set up with the Go toolchain and build dependencies.

The RPMs are written under `_output/rpmbuild/RPMS`:

```bash
$ find _output/rpmbuild/RPMS -name '*.rpm' | head
_output/rpmbuild/RPMS/x86_64/microshift-5.0.0_0.nightly...el9.x86_64.rpm
_output/rpmbuild/RPMS/x86_64/microshift-networking-5.0.0_0.nightly...el9.x86_64.rpm
_output/rpmbuild/RPMS/noarch/microshift-release-info-5.0.0_0.nightly...el9.noarch.rpm
...
```

> The version string encodes the source commit (for example
> `5.0.0_0.nightly_..._<git-sha>`), so it differs from any released MicroShift
> version. This is how you confirm the resulting image contains your source build
> rather than a released RPM.

### Create a local RPM repository

Turn the built RPMs into a `dnf` repository so they can be consumed during the
image build:

```bash
createrepo_c _output/rpmbuild/RPMS
```

This creates `_output/rpmbuild/RPMS/repodata`, indexing the RPMs in the `noarch`
and `x86_64` subdirectories.

### Build the bootc image

Use the [Containerfile.bootc-source-rhel9](../config/Containerfile.bootc-source-rhel9)
source build Containerfile from a clone of this repository.

Unlike the released Containerfile, it:
* copies the local RPM repository into the image and installs `microshift` from it
* pulls the MicroShift runtime dependencies that are not part of base RHEL (cri-o,
  cri-tools, openshift-clients and openvswitch) from the public OpenShift
  dependencies mirror rather than from `rhocp`/`fast-datapath`, because the
  matching versions for a pre-release source build are not yet available in those
  subscription repositories

> The dependencies mirror URL tracks the current development stream and is
> architecture specific. See the comments in the Containerfile and the
> `RHOCP_MINOR_Y_BETA` variable in
> [test/bin/common_versions.sh](../../test/bin/common_versions.sh) for the
> authoritative value, and override `DEPS_REPO_URL` with `--build-arg` when
> building for another architecture.

Note how secrets are used during the image build:
* The podman `--authfile` argument is required to pull the base image from the
`registry.redhat.io` registry
* The build `USER_PASSWD` argument is used to set a password for the `redhat` user

Run the following command, using `_output/rpmbuild/RPMS` as the build context so
the local repository is available to the build:

```bash
PULL_SECRET=~/.pull-secret.json
USER_PASSWD="<your_redhat_user_password>"
IMAGE_NAME=microshift-source-bootc

sudo podman build --authfile "${PULL_SECRET}" -t "${IMAGE_NAME}" \
    --build-arg USER_PASSWD="${USER_PASSWD}" \
    -f docs/config/Containerfile.bootc-source-rhel9 \
    _output/rpmbuild/RPMS
```

> **Important:**<br>
> The `Containerfile` runs `dnf upgrade` pinned to the base image release version
> (`dnf upgrade -y --releasever="${VERSION_ID}"`). Pinning `--releasever` prevents
> an unintended minor version upgrade of the operating system (for example, from
> `9.8` to a later RHEL `9.y`) that a plain `dnf upgrade` could otherwise perform,
> keeping the image on the same RHEL version as its base.

Verify that the local MicroShift `bootc` image was created:

```bash
$ sudo podman images "${IMAGE_NAME}"
REPOSITORY                          TAG     IMAGE ID      CREATED        SIZE
localhost/microshift-source-bootc   latest  193425283c00  2 minutes ago  2.89 GB
```

> To run this image directly as a `podman` container for fast development
> turnaround — the quickest way to exercise a source build without installing it
> on a host — see [Image Mode for MicroShift Contributors](../contributor/image_mode.md).

## Deploy the source-built image

Publishing the image to a registry, preparing a Kickstart file, and installing the
image into a virtual machine are identical for a source-built image and a released
image. Rather than duplicate those procedures, follow the openshift-docs
instructions and substitute your source-built image reference wherever a MicroShift
image is required:

* [Installing and publishing a bootc image to a registry](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/installing_with_image_mode_for_rhel/microshift-install-bootc-image) —
  skip the "get a published image" step (you built the image locally above) and
  push `localhost/microshift-source-bootc` to your remote registry.
* [Running the bootc image in a virtual machine](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/latest/html/installing_with_image_mode_for_rhel/microshift-install-running-bootc-image-vm) —
  set the image reference used by the `ostreecontainer` Kickstart directive to your
  source-built image (for example
  `<myreg>/<myorg>/<mypath>/microshift-source-bootc`).

The remaining sections cover workflows that are **not** part of the openshift-docs
image mode documentation: building a self-contained installation ISO with
`bootc-image-builder`, and embedding container images for offline installation.

## Using Bootc Image Builder (BIB)

The [bootc-image-builder](https://github.com/osbuild/bootc-image-builder), is a
containerized tool to create disk images from bootc images. You can use the tool
to generate various image artifacts and deploy them in different environments,
such as the edge, server, and clouds.

Log into the `RHEL 9.8 host` using the user credentials that have SUDO
permissions configured.

### Prepare Kickstart File

Set variables pointing to secret files that are included in `kickstart.ks` for
gaining access to private container registries:
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
  at the post-install stage to authenticate OpenShift registry access

```bash
PULL_SECRET=~/.pull-secret.json
```

Run the following command to create the `kickstart.ks` file to be used during
the virtual machine installation. If you want to embed the kickstart file directly
to ISO using BIB refer to [upstream docs](https://osbuild.org/docs/bootc/#anaconda-iso-installer-options-installer-mapping).

```bash
cat > kickstart.ks <<EOFKS
lang en_US.UTF-8
keyboard us
timezone UTC
text
reboot

# Partition the disk with hardware-specific boot and swap partitions, adding an
# LVM volume that contains a 10GB+ system root. The remainder of the volume will
# be used by the CSI driver for storing data.
zerombr
clearpart --all --initlabel
# Create boot and swap partitions as required by the current hardware platform
reqpart --add-boot
# Add an LVM volume group and allocate a system root logical volume
part pv.01 --grow
volgroup rhel pv.01
logvol / --vgname=rhel --fstype=xfs --size=10240 --name=root

# Lock root user account
rootpw --lock

# Configure network to use DHCP and activate on boot
network --bootproto=dhcp --device=link --activate --onboot=on

%post --log=/dev/console --erroronfail

# Create an OpenShift pull secret file
cat > /etc/crio/openshift-pull-secret <<'EOF'
$(cat "${PULL_SECRET}")
EOF
chmod 600 /etc/crio/openshift-pull-secret

%end
EOFKS
```

### Create ISO image using BIB

```bash
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-source-bootc

mkdir ./output
sudo podman run --authfile ${PULL_SECRET} --rm -it \
    --privileged \
    --security-opt label=type:unconfined_t \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    -v ./output:/output \
    registry.redhat.io/rhel9/bootc-image-builder:latest \
    --local \
    --type iso \
    localhost/${IMAGE_NAME}:latest
```

> **NOTE**:<br>
> Use `--config` argument to optionally specify additional BIB build-time customizations
> as described in [Build config](https://osbuild.org/docs/bootc/#-build-config).

### Create Virtual Machine

Run the following commands to copy the `./output/install.iso` file to the
`/var/lib/libvirt/images` directory and create a virtual machine.

```bash
VMNAME=microshift-source-bootc
NETNAME=default

sudo cp -Z ./output/bootiso/install.iso /var/lib/libvirt/images/${VMNAME}.iso

sudo virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 2048 \
    --disk path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --location /var/lib/libvirt/images/${VMNAME}.iso \
    --initrd-inject kickstart.ks \
    --extra-args "inst.ks=file:/kickstart.ks" \
    --wait
```

Log into the virtual machine using the `redhat:<password>` credentials.
Run the following command to verify that all the MicroShift pods are up and running
without errors.

```bash
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```

## Appendix A: Embedding Container Images in Bootc Builds

Adding MicroShift container image dependencies to bootc images may be necessary
for isolated (no Internet access) setup or for improving MicroShift first startup
performance. The container image references are specific to platform and to each
MicroShift version.

Use this approach to create a fully self contained image that does not have any
external dependencies on startup.

### Build Container Image

Download the [Containerfile.embedded](../config/Containerfile.bootc-embedded-rhel9) using
the following command and use it for subsequent image builds.

```bash
URL=https://raw.githubusercontent.com/openshift/microshift/refs/heads/main/docs/config/Containerfile.bootc-embedded-rhel9

curl -s -o Containerfile.embedded "${URL}"
```

> Review comments in the `Containerfile.embedded` file to understand how container
> dependencies are embedded during the `bootc` image build.

Run the following image build command to create a local `bootc` image with embedded
container dependencies. It is using a base image built according to the instructions
in the [Build a MicroShift bootc image from source](#build-a-microshift-bootc-image-from-source)
section.

Note how secrets are used during the image build:
* The podman `--authfile` argument is required to pull the base image from the
`registry.redhat.io` registry
* The podman `--secret` argument is required to pull image dependencies from the
OpenShift container registries.

```bash
PULL_SECRET=~/.pull-secret.json
BASE_IMAGE_NAME=microshift-source-bootc
BASE_IMAGE_TAG=latest
IMAGE_NAME=microshift-source-bootc-embedded

sudo podman build --authfile "${PULL_SECRET}" -t "${IMAGE_NAME}" \
    --secret "id=pullsecret,src=${PULL_SECRET}" \
    --build-arg USHIFT_BASE_IMAGE_NAME="${BASE_IMAGE_NAME}" \
    --build-arg USHIFT_BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
    -f Containerfile.embedded
```

Verify that the local MicroShift `bootc` image was created.

```bash
$ sudo podman images "${IMAGE_NAME}"
REPOSITORY                                  TAG     IMAGE ID      CREATED             SIZE
localhost/microshift-source-bootc-embedded  latest  6490d8f5752a  About a minute ago  3.75 GB
```

### Build Installation Image

Follow the instructions in [Create ISO Image Using BIB](#create-iso-image-using-bib)
to build an ISO from the container image with embedded container dependencies.

> Note: Make sure to set the `IMAGE_NAME` variable to `microshift-source-bootc-embedded`

### Prepare Kickstart File

Set variables pointing to secret files that are included in `kickstart.ks` for
gaining access to private container registries:
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
  at the post-install stage to authenticate OpenShift registry access

```bash
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-source-bootc-embedded
```

Run the following command to create the `kickstart.ks` file to be used during
the virtual machine installation.

```bash
cat > kickstart.ks <<EOFKS
lang en_US.UTF-8
keyboard us
timezone UTC
text
reboot

# Partition the disk with hardware-specific boot and swap partitions, adding an
# LVM volume that contains a 10GB+ system root. The remainder of the volume will
# be used by the CSI driver for storing data.
zerombr
clearpart --all --initlabel
# Create boot and swap partitions as required by the current hardware platform
reqpart --add-boot
# Add an LVM volume group and allocate a system root logical volume
part pv.01 --grow
volgroup rhel pv.01
logvol / --vgname=rhel --fstype=xfs --size=10240 --name=root

# Lock root user account
rootpw --lock

# Configure network to use DHCP and activate on boot
network --bootproto=dhcp --device=link --activate --onboot=on

# Configure bootc to install from the local embedded container repository.
# See /osbuild-base.ks on ISO images generated by bootc-image-builder.
ostreecontainer --transport oci --url /run/install/repo/container

%post --log=/dev/console --erroronfail

# Update the image reference for updates to work correctly.
# See /osbuild.ks on ISO images generated by bootc-image-builder.
bootc switch --mutate-in-place --transport registry localhost/${IMAGE_NAME}

# Create an OpenShift pull secret file
cat > /etc/crio/openshift-pull-secret <<'EOF'
$(cat "${PULL_SECRET}")
EOF
chmod 600 /etc/crio/openshift-pull-secret

%end
EOFKS
```

### Configure Isolated Network

Before creating a virtual machine, it is necessary to configure a `libvirt`
network without Internet access. Run the following commands to create such
a network.

```bash
VM_ISOLATED_NETWORK=microshift-isolated-network

cat > isolated-network.xml <<EOF
<network>
  <name>${VM_ISOLATED_NETWORK}</name>
  <forward mode='none'/>
  <ip address='192.168.111.1' netmask='255.255.255.0' localPtr='yes'>
    <dhcp>
      <range start='192.168.111.100' end='192.168.111.254'/>
    </dhcp>
  </ip>
</network>
EOF

sudo virsh net-define isolated-network.xml
sudo virsh net-start     "${VM_ISOLATED_NETWORK}"
sudo virsh net-autostart "${VM_ISOLATED_NETWORK}"
```

### Create Virtual Machine

Follow the instructions in [Create Virtual Machine](#create-virtual-machine)
to bootstrap a virtual machine from the ISO with embedded container dependencies.

> Note: Make sure to set the `NETNAME` variable to the `VM_ISOLATED_NETWORK`
> isolated network name.

Log into the virtual machine **console** using the `redhat:<password>` credentials.

Run the following command to verify that there is no Internet access, thus
no container image dependencies could have been pulled over the network.

```bash
$ curl -I redhat.com
curl: (6) Could not resolve host: redhat.com
```

Run the following command to verify that all the MicroShift pods are up and running
without errors.

```bash
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```
