# Image Mode Container Build Tools

Image mode container build tools are implemented in the `microshift/packaging/imagemode`
directory using `make` rules. Run the following command to see the available options.

```bash
$ cd packaging/imagemode
$ make
make [rhocp | repourl | repobase | <build_ver> | run | clean]
   rhocp:       build a MicroShift bootc image using 'rhocp' repository packages
                with versions specified as 'USHIFT_VER=value'
   repourl:     build a MicroShift bootc image using custom repository URLs
                specified as 'USHIFT_URL=value' and 'OCPDEP_URL=value'
   repobase:    build a MicroShift bootc image using preconfigured repositories
                from the base image specified as 'BASE_IMAGE_URL=value' and
                'BASE_IMAGE_TAG=value'. The produced image version should also
                be provided as 'IMAGE_VER=value' in this case.
   <build_ver>: build a MicroShift bootc image of a specific version from the
                available predefined configurations listed below
   run:         run the 'localhost/microshift-${IMAGE_VER}' bootc image version
                specified as 'IMAGE_VER=value'
   stop:        stop all running 'microshift-*' containers
   clean:       delete all 'localhost/microshift-*' container images

Available build versions:
   4.16-el94
   4.17-rc-el94
   4.18-ec-cos9
   4.18-ec-el94
```

## Build Image Mode Containers

Log into the `RHEL 9.4 host` using the user credentials that have SUDO permissions
configured.

The `rhocp`, `repourl` and `repobase` targets can be used for building `bootc`
container images.

### Build from `rhocp` Repository

The `rhocp` target allows for building `bootc` container images that include MicroShift
packages from the `rhocp-<version>-for-rhel-9-$(uname -m)-rpms` repository.

The target requires the `USHIFT_VER=value` argument, which defines the version
of the `rhocp` repository to be used when building the image.

For example, run the following command to build an image including the latest
released MicroShift 4.16 version.

```bash
make rhocp USHIFT_VER=4.16
```

The resulting image will be named `microshift-4.16.z` where `z` is the latest
available MicroShift package version in the repository.

```bash
$ sudo podman images --format "{{.Repository}}" | grep ^localhost/microshift-4.16
localhost/microshift-4.16.8
```

### Build from Custom URL Repository

The `repourl` target allows for building `bootc` container images that include
MicroShift packages from custom repositories defined by URLs specified in the
command line.

The target requires the `USHIFT_URL=value` and `OCPDEP_URL=value` arguments
which define the URL of repositories containing MicroShift RPM packages and
OpenShift dependency RPM packages.

For example, run the following command to build an image including the MicroShift
4.17 Release Candidate version from `mirror.openshift.com` site.

```bash
BASE_URL="https://mirror.openshift.com/pub/openshift-v4"
make repourl \
    USHIFT_URL="${BASE_URL}/$(uname -m)/microshift/ocp/latest-4.17/el9/os/" \
    OCPDEP_URL="${BASE_URL}/$(uname -m)/dependencies/rpms/4.17-el9-beta/"
```

The resulting image will be named `microshift-4.17.z` where `z` is the latest
available MicroShift package version in the repository.

```bash
$ sudo podman images --format "{{.Repository}}" | grep ^localhost/microshift-4.17
localhost/microshift-4.17.0-rc.0
```

### Build from Custom Base Image Repository

The `repobase` target allows for building `bootc` container images that include
MicroShift packages from custom repositories defined in the base image specified
in the command line.

The target requires the `BASE_IMAGE_URL=value`, `BASE_IMAGE_TAG=value` and `IMAGE_VER=value`
arguments, which define the base image URL, tag and the version of the produced
local MicroShift `bootc` container image (i.e. `microshift-${IMAGE_VER}`). All
the required RPM repository configuration is assumed to be part of the base image.

For example, run the following command to build an image using the local MicroShift
4.16 image with `rhocp` repositories built in the previous step.

> This example is superficial for the sake of simplicity. The typical use of the
> `repobase` target would be to decouple the repository configuration and MicroShift
> image build steps.

```bash
BASE_IMAGE_URL="localhost/microshift-4.16.9"
BASE_IMAGE_TAG="latest"
IMAGE_VER="4.16.9-update"

make repobase \
    BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
    BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
    IMAGE_VER="${IMAGE_VER}"
```

The resulting image will be named `microshift-4.16.9-update` as defined by the
`IMAGE_VER` argument.

```bash
$ sudo podman images --format "{{.Repository}}" | grep ^localhost/microshift-"${IMAGE_VER}"
localhost/microshift-4.16.9-update
```

### Predefined Build Targets

Run the following command to see the list of predefined build targets for selected
MicroShift versions and configurations.

```bash
$ make | grep -A10 'Available build versions'
Available build versions:
   4.16-el94
   4.17-rc-el94
   4.18-ec-el94
```

These builds use `rhocp` and `repourl` targets with hardcoded parameters to simplify
`make` command invocation.

For example, run the following command to build an image including the MicroShift
4.17 Release Candidate version from `mirror.openshift.com` site.

```bash
make 4.17-rc-el94
```

### Override Build Variables

The following container build parameters can be used to override some of the
default values used in `Containerfile` for `rhocp` and `repourl` targets.

| Parameter Name | Default Value | Comment |
|----------------|---------------|---------|
| PULL_SECRET    | `~/.pull-secret.json` | Used for accessing base `bootc` images |
| BASE_IMAGE_URL | `registry.redhat.io/rhel9-eus/rhel-9.4-bootc` | Base `bootc` image URL |
| BASE_IMAGE_TAG | `9.4` | Base `bootc` image tag |
| DNF_OPTIONS    | none | Additional options to be passed to the `dnf` command |

For example, run the following command to override the base `bootc` image default
tag when building the container image.

```bash
make rhocp USHIFT_VER=4.16 \
    BASE_IMAGE_TAG=latest
```

### Clean Local MicroShift Container Images

Run the following command to delete all the `localhost/microshift-*` container images.

```bash
make clean
```

## Appendix A: Run Image Mode Containers

> The purpose of this section is to demonstrate how to test generated MicroShift
> image mode containers.

Log into the `RHEL 9.4 host` using the user credentials that have SUDO permissions
configured.

The `run` target allows for running the specified `localhost/microshift-*` container
image. The target requires the `IMAGE_VER=value` argument which defines the version
of the image to be started.

For example, run the following commands to see the available images and start
one of them.

```bash
$ sudo podman images --format "{{.Repository}}" | grep ^localhost/microshift-
localhost/microshift-4.17.0-rc.0
localhost/microshift-4.16.8

$ make run IMAGE_VER=4.16.8
...
...
sudo podman exec -it 65339346b957c7b02353bf859b07d75a2127398266d6d3f3b2708b692745609f bash
```

> If the container is started successfully, the last line of the output shows a
> command to be used for logging into the running container.

The `stop` target stops all running `microshift-*` containers.

For example, run the following command to stop all the running MicroShift containers.

```bash
$ make stop
...
...
microshift-4.16.8
```

## Appendix B: Boot RHEL Using Image Mode Containers

> The purpose of this section is to demonstrate how to test generated MicroShift
> image mode containers.

Follow the procedure below to create a virtual machine using pre-built MicroShift
bootc images. A similar procedure can be used for booting physical devices.

Log into the physical hypervisor host using the user credentials that have SUDO
permissions configured.

### Prepare Kickstart File

Set variables pointing to secrets that are included in `kickstart.ks`.
* `USER_PASSWD` is used to set a password for the `redhat` user
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
to authenticate OpenShift registry access

```
USER_PASSWD=<your_redhat_user_password>
PULL_SECRET=~/.pull-secret.json
```

Run the following commands to create the `kickstart.ks` file to be used during
the virtual machine installation.

```
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

# Pull a 'bootc' image from the remote registry
ostreecontainer --url registry.redhat.io/microshift-4.18-bootc:latest

%post --log=/dev/console --erroronfail

# Create the 'redhat' user
useradd -G wheel redhat
echo "redhat:${USER_PASSWD}" | chpasswd

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

> Replace `registry.redhat.io/microshift-4.18-bootc:latest` with the image reference
> you would like to install.

### Create Virtual Machine

Download a RHEL boot ISO image from https://developers.redhat.com/products/rhel/download.
Copy the downloaded file to the `/var/lib/libvirt/images` directory.

Run the following commands to create a RHEL virtual machine with 2 cores, 2GB of
RAM and 20GB of storage. The command uses the kickstart file prepared in the
previous step to install the RHEL operating system.

```
VMNAME=microshift-4.18-el94
NETNAME=default

sudo virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 2048 \
    --disk path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --location "/var/lib/libvirt/images/${VMNAME}.iso" \
    --osinfo detect=on \
    --initrd-inject kickstart.ks \
    --extra-args "inst.ks=file://kickstart.ks" \
    --wait
```

Log into the virtual machine using the `redhat:<password>` credentials.
Run the following command to verify that all the MicroShift pods are up and running
without errors.

```
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```

## Appendix C: Create CentOS 9 Stream Image Mode Containers

Creating MicroShift image mode containers with CentOS Stream 9 image base can be
accomplished by using the `repourl` target with build variables override.

Log into a Linux host (e.g. `Fedora`, `CentOS` or `RHEL`) using the user credentials
that have SUDO permissions configured.

For example, run the following command to build a CentOS Stream 9 image using the
MicroShift 4.17 Release Candidate version from `mirror.openshift.com` site.

```bash
BASE_IMAGE_URL=quay.io/centos-bootc/centos-bootc
BASE_IMAGE_TAG=stream9
BASE_URL="https://mirror.openshift.com/pub/openshift-v4"

make repourl \
    BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
    BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
    USHIFT_URL="${BASE_URL}/$(uname -m)/microshift/ocp/latest-4.17/el9/os/" \
    OCPDEP_URL="${BASE_URL}/$(uname -m)/dependencies/rpms/4.17-el9-beta/"
```

Note that RPM packages referenced by the `USHIFT_URL` and `OCPDEP_URL` may conflict
with those delivered in the CentOS repositories. To work around this problem, add
the `DNF_OPTIONS="--allowerasing --nobest"` argument to the image build command.

Run a container using the generated image as described in [Appendix A: Run Image Mode Containers](#appendix-a-run-image-mode-containers),
log into the running container and execute the following command to verify the
base operating system version.

```bash
$ cat /etc/redhat-release
CentOS Stream release 9
```
