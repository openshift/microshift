# Image Mode for MicroShift Users

Image mode is a new approach to operating system deployment that lets users build,
deploy, and manage Red Hat Enterprise Linux as a bootable container (`bootc`)
image. Such an image uses standard OCI/Docker containers as a transport and delivery
format for base operating system updates. A `bootc` image includes a Linux kernel,
which is used to boot.

MicroShift build and deployment procedures can utilize bootable containers to
benefit from this technology.

This document demonstrates how to create a bootable container image, store this
image in a remote registry and use it for installing a new RHEL operating system.

> See [Image mode for Red Hat Enterprise Linux](https://developers.redhat.com/products/rhel-image-mode/overview)
for more information.

The procedures described below require the following setup:
* A `RHEL 9.4 host` with an active Red Hat subscription for building MicroShift `bootc`
images. For development purposes, you can use the [Red Hat Developer subscription](https://developers.redhat.com/products/rhel/download), which is free of charge.
* A `hypervisor host` with a virtualization technology that supports RHEL. In
this documentation, [libvirt](https://libvirt.org/) virtualization is used as
an example.
* A `remote registry` (e.g. `quay.io`) for storing and accessing `bootc` images

## Build MicroShift Bootc Image

Log into the `RHEL 9.4 host` using the user credentials that have SUDO
permissions configured.

### Build Image

Download the [Containerfile](../config/Containerfile.bootc-rhel9) using the following
command and use it for subsequent image builds.

```bash
URL=https://raw.githubusercontent.com/openshift/microshift/refs/heads/main/docs/config/Containerfile.bootc-rhel9

curl -s -o Containerfile "${URL}"
```

> **Important:**<br>
> When building a container image, podman uses host subscription information and
> repositories inside the container. With only `BaseOS` and `AppStream` system
> repositories enabled by default, we need to enable the `rhocp` and `fast-datapath`
> repositories for installing MicroShift and its dependencies. These repositories
> must be accessible in the host subscription, but not necessarily enabled on the host.

Examine the `rhocp` and `fast-datapath` repositories that are available in the host
subscription.

```bash
sudo subscription-manager repos --list | \
    grep '^Repo ID:' | \
    grep -E 'fast-datapath|rhocp-4.*-for-rhel' | sort
```

Run the following image build command to create a local `bootc` image.

Note how secrets are used during the image build:
* The podman `--authfile` argument is required to pull the base image from the
`registry.redhat.io` registry
* The build `USER_PASSWD` argument is used to set a password for the `redhat` user

```bash
PULL_SECRET=~/.pull-secret.json
USER_PASSWD="<your_redhat_user_password>"
IMAGE_NAME=microshift-4.18-bootc

sudo podman build --authfile "${PULL_SECRET}" -t "${IMAGE_NAME}" \
    --build-arg USER_PASSWD="${USER_PASSWD}" \
    -f Containerfile
```

> **Important:**<br>
> If `dnf upgrade` command is used in the container image build procedure, it
> may cause unintended operating system version upgrade (e.g. from `9.4` to
> `9.6`). To prevent this from happening, use the following command instead.
> ```
> RUN . /etc/os-release && dnf upgrade -y --releasever="${VERSION_ID}"
> ```

Verify that the local MicroShift 4.18 `bootc` image was created.

```bash
$ sudo podman images "${IMAGE_NAME}"
REPOSITORY                       TAG         IMAGE ID      CREATED        SIZE
localhost/microshift-4.18-bootc  latest      193425283c00  2 minutes ago  2.31 GB
```

### Push Image to Registry

Run the following commands to log into a remote registry and push the image to
a remote registry.

> The image from the remote registry can be used for running the container on
> another host, or when installing a new operating system with the `bootc`
> image layer.

```bash
REGISTRY_URL=<myreg>
REGISTRY_IMG=<myorg>/<mypath>/"${IMAGE_NAME}"

sudo podman login "${REGISTRY_URL}"
sudo podman push localhost/"${IMAGE_NAME}" "${REGISTRY_URL}/${REGISTRY_IMG}"
```

> Replace `<myreg>` with the URL to your remote registry, and `<myorg>/<mypath>`
> with your organization name and path inside your remote registry.

## Run MicroShift Bootc Virtual Machine

Log into the `physical hypervisor host` using the user credentials that have
SUDO permissions configured.

### Prepare Kickstart File

Set variables pointing to secret files that are included in `kickstart.ks` for
gaining access to private container registries:
* `AUTH_CONFIG` file contents are copied to `/etc/ostree/auth.json` at the
pre-install stage to authenticate `<myreg>/<myorg>` registry access
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
at the post-install stage to authenticate OpenShift registry access
* `IMAGE_REF` variable contains the MicroShift bootc container image reference
to be installed

```bash
AUTH_CONFIG=~/.registry-auth.json
PULL_SECRET=~/.pull-secret.json
IMAGE_REF="<myreg>/<myorg>/<mypath>/microshift-4.18-bootc"
```

> Replace `<myreg>` with the URL to your remote registry, and `<myorg>/<mypath>`
> with your organization name and path inside your remote registry.

> See the `containers-auth.json(5)` manual pages for more information on the
> syntax of the `AUTH_CONFIG` registry authentication file.

Run the following commands to create the `kickstart.ks` file to be used during
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

%pre-install --log=/dev/console --erroronfail

# Create a 'bootc' image registry authentication file
mkdir -p /etc/ostree
cat > /etc/ostree/auth.json <<'EOF'
$(cat "${AUTH_CONFIG}")
EOF

%end

# Pull a 'bootc' image from a remote registry
ostreecontainer --url "${IMAGE_REF}"

%post --log=/dev/console --erroronfail

# Create an OpenShift pull secret file
cat > /etc/crio/openshift-pull-secret <<'EOF'
$(cat "${PULL_SECRET}")
EOF
chmod 600 /etc/crio/openshift-pull-secret

%end
EOFKS
```

The kickstart file uses a special [ostreecontainer](https://pykickstart.readthedocs.io/en/latest/kickstart-docs.html#ostreecontainer)
directive to pull a `bootc` image from the remote registry and use it to install
the RHEL operating system.

### Create Virtual Machine

Download a RHEL boot ISO image from https://developers.redhat.com/products/rhel/download.
Copy the downloaded file to the `/var/lib/libvirt/images` directory.

Run the following commands to create a RHEL virtual machine with 2 cores, 2GB of
RAM and 20GB of storage. The command uses the kickstart file prepared in the
previous step to pull a `bootc` image from the remote registry and use it to install
the RHEL operating system.

```bash
VMNAME=microshift-4.18-bootc
NETNAME=default

sudo virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 2048 \
    --disk path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --location /var/lib/libvirt/images/rhel-9.4-$(uname -m)-boot.iso \
    --initrd-inject kickstart.ks \
    --extra-args "inst.ks=file://kickstart.ks" \
    --wait
```

Log into the virtual machine using the `redhat:<password>` credentials.
Run the following command to verify that all the MicroShift pods are up and running
without errors.

```bash
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```

## Using Bootc Image Builder (BIB)

The [bootc-image-builder](https://github.com/osbuild/bootc-image-builder), is a
containerized tool to create disk images from bootc images. You can use the tool
to generate various image artifacts and deploy them in different environments,
such as the edge, server, and clouds.

### Create ISO image using BIB

```bash
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-4.18-bootc

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
to iso using BIB please refer to [upstream docs](https://osbuild.org/docs/bootc/#anaconda-iso-installer-options-installer-mapping)

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

### Create Virtual Machine

Run the following commands to copy the `./output/install.iso` file to the
`/var/lib/libvirt/images` directory and create a virtual machine.

```bash
VMNAME=microshift-4.18-bootc
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
    --extra-args "inst.ks=file://kickstart.ks" \
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
in the [Build Image](#build-image) section.

Note how secrets are used during the image build:
* The podman `--authfile` argument is required to pull the base image from the
`registry.redhat.io` registry
* The podman `--secret` argument is required to pull image dependencies from the
OpenShift container registries.

```bash
PULL_SECRET=~/.pull-secret.json
BASE_IMAGE_NAME=microshift-4.18-bootc
BASE_IMAGE_TAG=latest
IMAGE_NAME=microshift-4.18-bootc-embedded

sudo podman build --authfile "${PULL_SECRET}" -t "${IMAGE_NAME}" \
    --secret "id=pullsecret,src=${PULL_SECRET}" \
    --build-arg USHIFT_BASE_IMAGE_NAME="${BASE_IMAGE_NAME}" \
    --build-arg USHIFT_BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
    -f Containerfile.embedded
```

Verify that the local MicroShift 4.18 `bootc` image was created.

```bash
$ sudo podman images "${IMAGE_NAME}"
REPOSITORY                                TAG         IMAGE ID      CREATED             SIZE
localhost/microshift-4.18-bootc-embedded  latest      6490d8f5752a  About a minute ago  3.75 GB
```

### Build Installation Image

Follow the instructions in [Create ISO Image Using BIB](#create-iso-image-using-bib)
to build an ISO from the container image with embedded container dependencies.

> Note: Make sure to set the `IMAGE_NAME` variable to `microshift-4.18-bootc-embedded`

### Prepare Kickstart File

Set variables pointing to secret files that are included in `kickstart.ks` for
gaining access to private container registries:
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
  at the post-install stage to authenticate OpenShift registry access

```bash
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-4.18-bootc-embedded
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

Follow the instructions in [Create Virtual Machine](#create-virtual-machine-1)
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
