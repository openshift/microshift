# Image Mode for MicroShift Contributors

Follow the instructions in [Image Mode for MicroShift Users](../user/image_mode.md)
to create a bootable container image, store this image in a remote registry and
use it for installing a new RHEL operating system.

This document demonstrates how to run a `bootc` image using `podman`.

> **NOTE**:<br>
> Use the `podman` approach only for development purposes to benefit from
> the fast turnaround times it allows. Do not use it for production use cases.

The procedures described below require the following setup:
* A `RHEL 9.4 host` with an active Red Hat subscription for building MicroShift `bootc`
images and running containers
* A `remote registry` (e.g. `quay.io`) for storing and accessing `bootc` images

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

```bash
$ find /lib/modules/$(uname -r) -name "openvswitch*"
/lib/modules/6.9.9-200.fc40.x86_64/kernel/net/openvswitch
/lib/modules/6.9.9-200.fc40.x86_64/kernel/net/openvswitch/openvswitch.ko.xz

$ IMAGE_NAME=microshift-4.18-bootc
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

```bash
$ sudo vgs
  VG   #PV #LV #SN Attr   VSize   VFree
  rhel   1   1   0 wz--n- <91.02g <2.02g
```

Otherwise, a new volume group should be set up for MicroShift CSI driver to allocate
storage in `bootc` MicroShift containers.

Run the following commands to create a file to be used for LVM partitioning and
configure it as a loop device.

```bash
VGFILE=/var/lib/microshift-lvm-storage.img
VGSIZE=1G

sudo truncate --size="${VGSIZE}" "${VGFILE}"
sudo losetup -f "${VGFILE}"
```

Query the loop device name and create a free volume group on the device according
to the MicroShift CSI driver requirements described in [Storage Configuration](./storage/configuration.md).

```bash
VGLOOP=$(losetup -j ${VGFILE} | cut -d: -f1)
sudo vgcreate -f -y rhel "${VGLOOP}"
```

The device will now be shared with the newly created containers as described in
the next section.

> The following commands can be run to detach the loop device and delete the LVM
> volume group file.
>
> ```bash
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

```bash
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-4.18-bootc

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

```bash
watch sudo oc get pods -A \
    --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig
```

> Run the `sudo shutdown now` command to stop the container.

## Appendix A: Multi-Architecture Image Build

It is often convenient to build multi-architecture container images and store
them under the same registry URL using manifest lists.

> See [podman-manifest](https://docs.podman.io/en/latest/markdown/podman-manifest.1.html) for more information.

The [Build Image](#build-image) procedure needs to be adjusted in the following
manner to create multi-architecture images.

```bash
PULL_SECRET=~/.pull-secret.json
USER_PASSWD="<your_redhat_user_password>"
IMAGE_ARCH=amd64 # Use amd64 or arm64 depending on the current platform
IMAGE_PLATFORM="linux/${IMAGE_ARCH}"
IMAGE_NAME="microshift-4.18-bootc:linux-${IMAGE_ARCH}"

sudo podman build --authfile "${PULL_SECRET}" -t "${IMAGE_NAME}" \
    --platform "${IMAGE_PLATFORM}" \
    --build-arg USER_PASSWD="${USER_PASSWD}" \
    -f Containerfile
```

Verify that the local MicroShift 4.18 `bootc` image was created for the specified
platform.

```bash
$ sudo podman images "${IMAGE_NAME}"
REPOSITORY                       TAG          IMAGE ID      CREATED         SIZE
localhost/microshift-4.18-bootc  linux-amd64  3f7e136fccb5  13 minutes ago  2.19 GB
```

Repeat the procedure on the other platform (i.e. `arm64`) and proceed by publishing
the platform-specific `amd64` and `arm64` images to the remote registry as described
in the [Publish Image](#publish-image) section.

> Cross-platform `podman` builds are not in the scope of this document. Log into
> the RHEL 9.4 host running on the appropriate architecture to perform the container
> image builds and publish the platform-specific image to the remote registry.

Finally, create a manifest containing the platform-specific image references
and publish it to the remote registry.

> Images for **both** `amd64` and `arm64` architectures should have been pushed
> to the remote registry before creating the manifest.

```bash
REGISTRY_URL=quay.io
REGISTRY_ORG=myorg/mypath
BASE_NAME=microshift-4.18-bootc
MANIFEST_NAME="${BASE_NAME}:latest"

sudo podman manifest create -a "localhost/${MANIFEST_NAME}" \
    "${REGISTRY_URL}/${REGISTRY_ORG}/${BASE_NAME}:linux-amd64" \
    "${REGISTRY_URL}/${REGISTRY_ORG}/${BASE_NAME}:linux-arm64"

sudo podman manifest push \
    "localhost/${MANIFEST_NAME}" \
    "${REGISTRY_URL}/${REGISTRY_ORG}/${MANIFEST_NAME}"
```

> Replace `myorg/mypath` with your remote registry organization name and path.

Inspect the remote manifest to make sure it contains image digests from multiple
architectures.

```bash
$ sudo podman manifest inspect \
    "${REGISTRY_URL}/${REGISTRY_ORG}/${MANIFEST_NAME}" | \
    jq .manifests[].platform.architecture
"amd64"
"arm64"
```

It is now possible to access images using the manifest name with the `latest` tag
(e.g. `quay.io/myorg/mypath/microshift-4.18-bootc:latest`). The image for the
current platform will automatically be pulled from the registry if it is part of
the manifest list.

## Appendix B: The rpm-ostree to Image Mode Upgrade Procedure

Refer to RHEL documentation for generic instructions on upgrading `rpm-ostree`
systems to Image Mode. The upgrade process should be planned carefully considering
the following guidelines:
* Follow instructions in RHEL documentation for converting `rpm-ostree` blueprints to
  Image Mode container files
* Consider using [rpm-ostree compose container-encapsulate](https://coreos.github.io/rpm-ostree/container/#converting-ostree-commits-to-new-base-images)
  to experiment with Image Mode based on the existing `ostree` commits
* Invest in defining a proper container build pipeline for fully adopting Image Mode

If reinstalling MicroShift devices from scratch is not an option, read the remainder
of this section that outlines the upgrade details specific to MicroShift.

Upgrading existing systems during the transition from `rpm-ostree` to Image Mode
may pose the challenge of [UID / GID Drift](https://github.com/bootc-dev/bootc/issues/673)
because the existing `rpm-ostree` and the new Image Mode images are not derived
from the same parent image.

One way of working around this issue is to add `systemd` units that run before the
affected system services and apply the necessary fixes.

> Note: The workaround is only necessary for `rpm-ostree` to Image Mode upgrade
> and it can be removed once all the devices are running the upgraded image.

Add the following command to the MicroShift image build procedure to create a
`systemd` unit file solving a potential UID / GID drift for `ovsdb-server.service`.

```
# Install systemd configuration drop-ins to fix potential permission problems
# when upgrading from older rpm-ostree commits to Image Mode container layers
RUN mkdir -p /etc/systemd/system/ovsdb-server.service.d && \
    cat > /etc/systemd/system/ovsdb-server.service.d/microshift-ovsdb-ownership.conf <<'EOF'
# The openvswitch database files must be owned by the appropriate user and its
# primary group. Note that the user and its group may be overwritten too, so
# they need to be recreated in this case.
[Service]
ExecStartPre=/bin/sh -c '/bin/getent passwd openvswitch >/dev/null || useradd -r openvswitch'
ExecStartPre=/bin/sh -c '/bin/getent group hugetlbfs >/dev/null || groupadd -r hugetlbfs'
ExecStartPre=/sbin/usermod -a -G hugetlbfs openvswitch
ExecStartPre=/bin/chown -Rhv openvswitch. /etc/openvswitch
EOF
