# Image Mode for MicroShift

Image mode is a new approach to operating system deployment that lets users build,
deploy, and manage Red Hat Enterprise Linux as a bootable container (`bootc`)
image. Such an image uses standard OCI/Docker containers as a transport and delivery
format for base operating system updates. A `bootc` image includes a Linux kernel,
which is used to boot.

MicroShift build and deployment procedures can utilize bootable containers to
benefit from this technology.

This document demonstrates how to:
* Create a bootable container image and boot RHEL from this image using `podman`
* Store a `bootc` image in a remote registry and use it for installing a new RHEL
operating system

> See [Image mode for Red Hat Enterprise Linux](https://developers.redhat.com/products/rhel-image-mode/overview)
for more information.

The procedures described below require the following setup:
* A `RHEL 9.4 host` with an active Red Hat subscription for building MicroShift `bootc`
images and running containers
* A `physical hypervisor host` with the [libvirt](https://libvirt.org/) virtualization
platform for starting virtual machines that run RHEL OS containing MicroShift binaries
* A `remote registry` (e.g. `quay.io`) for storing and accessing `bootc` images

## Build MicroShift Bootc Image

Log into the `RHEL 9.4 host` using the user credentials that have SUDO
permissions configured.

### Build Image

Create the `Containerfile` file with the following contents.

```
FROM registry.redhat.io/rhel9/rhel-bootc:9.4

ARG USHIFT_VER=4.16
RUN dnf config-manager \
        --set-enabled rhocp-${USHIFT_VER}-for-rhel-9-$(uname -m)-rpms \
        --set-enabled fast-datapath-for-rhel-9-$(uname -m)-rpms
RUN dnf install -y firewalld microshift && \
    systemctl enable microshift && \
    dnf clean all

# Create a default 'redhat' user with the specified password.
# Add it to the 'wheel' group to allow for running sudo commands.
ARG USER_PASSWD
RUN if [ -z "${USER_PASSWD}" ] ; then \
        echo USER_PASSWD is a mandatory build argument && exit 1 ; \
    fi
RUN useradd -m -d /var/home/redhat -G wheel redhat && \
    echo "redhat:${USER_PASSWD}" | chpasswd

# Mandatory firewall configuration
RUN firewall-offline-cmd --zone=public --add-port=22/tcp && \
    firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16 && \
    firewall-offline-cmd --zone=trusted --add-source=169.254.169.1
# Application-specific firewall configuration
RUN firewall-offline-cmd --zone=public --add-port=80/tcp && \
    firewall-offline-cmd --zone=public --add-port=443/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/udp

# Create a systemd unit to recursively make the root filesystem subtree
# shared as required by OVN images
RUN cat > /etc/systemd/system/microshift-make-rshared.service <<'EOF'
[Unit]
Description=Make root filesystem shared
Before=microshift.service
ConditionVirtualization=container
[Service]
Type=oneshot
ExecStart=/usr/bin/mount --make-rshared /
[Install]
WantedBy=multi-user.target
EOF
RUN systemctl enable microshift-make-rshared.service
```

> **Important:**<br>
> When building a container image, podman uses host subscription information and
> repositories inside the container. With only `BaseOS` and `AppStream` system
> repositories enabled by default, we need to enable the `rhocp` and `fast-datapath`
> repositories for installing MicroShift and its dependencies. These repositories
> must be accessible in the host subscription, but not necessarily enabled on the host.

Run the following image build command to create a local `bootc` image.

Note how secrets are used during the image build:
* The podman `--authfile` argument is required to pull the base `rhel-bootc:9.4`
image from the `registry.redhat.io` registry
* The build `USER_PASSWD` argument is used to set a password for the `redhat` user

```
PULL_SECRET=~/.pull-secret.json
USER_PASSWD=<your_redhat_user_password>
IMAGE_NAME=microshift-4.16-bootc

sudo podman build --authfile "${PULL_SECRET}" -t "${IMAGE_NAME}" \
    --build-arg USER_PASSWD="${USER_PASSWD}" \
    -f Containerfile
```

Verify that the local MicroShift 4.16 `bootc` image was created.

```
$ sudo podman images "${IMAGE_NAME}"
REPOSITORY                       TAG         IMAGE ID      CREATED        SIZE
localhost/microshift-4.16-bootc  latest      193425283c00  2 minutes ago  2.31 GB
```

### Publish Image

Run the following commands to log into a remote registry and publish the image.

> The image from the remote registry can be used for running the container on
> another host, or when installing a new operating system with the `bootc`
> image layer.

```
REGISTRY_URL=quay.io
REGISTRY_IMG=myorg/mypath/"${IMAGE_NAME}"

sudo podman login "${REGISTRY_URL}"
sudo podman push localhost/"${IMAGE_NAME}" "${REGISTRY_URL}/${REGISTRY_IMG}"
```

> Replace `myorg/mypath` with your remote registry organization name and path.

## Run MicroShift Bootc Image

Log into the `RHEL 9.4 host` using the user credentials that have SUDO
permissions configured.

### Configure CNI

The MicroShift CNI driver (OVN) requires the Open vSwitch service to function
properly. The service depends on the `openvswitch` kernel module to be available
in the `bootc` image.

Run the following commands on the host to check the `openvswitch` module presence
and the version of `kernel-core` package used in the `bootc` image. Note that the
kernel versions are different.

```
$ find /lib/modules/$(uname -r) -name "openvswitch*"
/lib/modules/6.9.9-200.fc40.x86_64/kernel/net/openvswitch
/lib/modules/6.9.9-200.fc40.x86_64/kernel/net/openvswitch/openvswitch.ko.xz

$ IMAGE_NAME=microshift-4.16-bootc
$ sudo podman inspect "${IMAGE_NAME}" | grep kernel-core
        "created_by": "kernel-core-5.14.0-427.26.1.el9_4.x86_64"
```

When a `bootc` image is started as a container, it uses the host kernel, which is
not necessarily the same one used for building the image. This means that the
`openvswitch` module cannot be loaded in the container due to the kernel version
mismatch with the modules present in the `/lib/modules` directory.

One way to work around this problem is to pre-load the `openvswitch` module before
starting the container as described in the [Run Container](#run-container) section.

### Configure CSI

If the host is already configured to have a `rhel` volume group with free space,
this configuration is inherited by the container so that it can be used by the
MicroShift CSI driver to allocate storage.

Run the following command to determine if the volume group exists and it has the
necessary free space.
```
$ sudo vgs
  VG   #PV #LV #SN Attr   VSize   VFree
  rhel   1   1   0 wz--n- <91.02g <2.02g
```

Otherwise, a new volume group should be set up for MicroShift CSI driver to allocate
storage in `bootc` MicroShift containers.

Run the following commands to create a file to be used for LVM partitioning and
configure it as a loop device.

```
VGFILE=/var/lib/microshift-lvm-storage.img
VGSIZE=1G

sudo truncate --size="${VGSIZE}" "${VGFILE}"
sudo losetup -f "${VGFILE}"
```

Query the loop device name and create a free volume group on the device according
to the MicroShift CSI driver requirements described in [Storage Configuration](./storage/configuration.md).

```
VGLOOP=$(losetup -j ${VGFILE} | cut -d: -f1)
sudo vgcreate -f -y rhel "${VGLOOP}"
```

The device will now be shared with the newly created containers as described in
the next section.

> The following commands can be run to detach the loop device and delete the LVM
> volume group file.
>
> ```
> sudo losetup -d "${VGLOOP}"
> sudo rm -f "${VGFILE}"
> ```

### Run Container

Run the following commands to start the MicroShift `bootc` image in an interactive
terminal session.

The host shares the following configuration with the container:
* The `openvswitch` kernel module to be used by the Open vSwitch service
* A pull secret file for downloading the required OpenShift container images
* Host container storage for reusing available container images

```
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-4.16-bootc

sudo modprobe openvswitch
sudo podman run --rm -it --privileged \
    -v "${PULL_SECRET}":/etc/crio/openshift-pull-secret:ro \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    --name "${IMAGE_NAME}" \
    "${IMAGE_NAME}"
```

> The `systemd-modules-load` service will fail to start in the container if the
> host kernel version is different from the `bootc` image kernel version. This
> failure can be safely ignored as all the necessary kernel modules have already
> been loaded by the host.

> If additional LVM volume group device was allocated as described in the
> [Configure CSI](#configure-csi) section, the loop device should automatically
> be shared with the container and used by the MicroShift CSI driver.

After the MicroShift `bootc` image has been successfully started, a login prompt
will be presented in the terminal. Log into the running container using the
`redhat:<password>` credentials.

Run the following command to verify that all the MicroShift pods are up and running
without errors.

```
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```

> Run the `sudo shutdown now` command to stop the container.

## Run MicroShift Bootc Virtual Machine

Log into the `physical hypervisor host` using the user credentials that have
SUDO permissions configured.

### Prepare Kickstart File

Set variables pointing to secret files that are included in `kickstart.ks` for
gaining access to private container registries:
* `AUTH_CONFIG` file contents are copied to `/etc/ostree/auth.json` at the
pre-install stage to authenticate `quay.io/myorg` registry access
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
at the post-install stage to authenticate OpenShift registry access

```
AUTH_CONFIG=~/.quay-auth.json
PULL_SECRET=~/.pull-secret.json
```

> See the `containers-auth.json(5)` manual pages for more information on the
> syntax of the `AUTH_CONFIG` registry authentication file.

Run the following commands to create the `kickstart.ks` file to be used during
the virtual machine installation.

```
cat > kickstart.ks <<EOFKS
lang en_US.UTF-8
keyboard us
timezone UTC
text
reboot

# Partition disk with a 1GB boot XFS partition and an LVM volume containing a 10GB+
# system root. The remainder of the volume is for the CSI driver for storing data.
zerombr
clearpart --all --initlabel
part biosboot --fstype=biosboot --size=1 --asprimary
part /boot/efi --fstype=efi --size=200
part /boot --fstype=xfs --asprimary --size=800
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
ostreecontainer --url quay.io/myorg/mypath/microshift-4.16-bootc

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

> Replace `myorg/mypath` with your remote registry organization name and path.

### Create Virtual Machine

Download a RHEL boot ISO image from https://developers.redhat.com/products/rhel/download.
Copy the downloaded file to the `/var/lib/libvirt/images` directory.

Run the following commands to create a RHEL virtual machine with 2 cores, 2GB of
RAM and 20GB of storage. The command uses the kickstart file prepared in the
previous step to pull a `bootc` image from the remote registry and use it to install
the RHEL operating system.

```
VMNAME=microshift-4.16-bootc
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

```
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```