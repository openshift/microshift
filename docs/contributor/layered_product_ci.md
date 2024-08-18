# Layered Product Testing with MicroShift

A Layered Product is a software package developed independently from the MicroShift
source repository and installed as a dependency atop MicroShift. For example, see
[MicroShift GitOps](https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/4.16/html/installing/microshift-install-rpm#microshift-installing-rpms-for-gitops_microshift-install-rpm).

This document outlines a technique for setting up runtime environments using
bootable containers to facilitate testing of such software packages. These tests
can be conducted manually or integrated into the package CI/CD processes. See
[Image Mode for MicroShift](./image_mode.md) for more information.

> MicroShift includes comprehensive [Automated Functional and Integration Tests](../../test/README.md)
> for features developed within its source repository.

## Build and Publish MicroShift Container Image

Follow the instructions in the [Build Image](./image_mode.md#build-image) section
to implement the MicroShift container image layer build procedure.

> Prebuilt MicroShift bootc container images are not currently available for
> download.

Customize the `Containerfile` file according to the requirements of the layered
product to be tested. A typical customization would be to select a custom version
of MicroShift, which may also include pre-released ones that are published at
[OpenShift Mirror](mirror.openshift.com).

**Example: MicroShift 4.17 Engineering Candidate Packages (fragment)**

```docker
ARG USHIFT_VER=4.17
ARG USHIFT_REPO="microshift-${USHIFT_VER}-mirror"
ARG OCPDEP_REPO="openshift-${USHIFT_VER}-mirror"

RUN cat > "/etc/yum.repos.d/${USHIFT_REPO}.repo" <<EOF
[${USHIFT_REPO}]
name=MicroShift ${USHIFT_VER} Dev Preview Mirror
baseurl="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/microshift/ocp-dev-preview/latest-${USHIFT_VER}/el9/os"
gpgcheck=0
enabled=1
EOF

RUN cat > "/etc/yum.repos.d/${OCPDEP_REPO}.repo" <<EOF
[${OCPDEP_REPO}]
name=OpenShift ${USHIFT_VER} Dependencies Mirror
baseurl="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/${USHIFT_VER}-el9-beta/"
gpgcheck=0
enabled=1
EOF

RUN dnf install -y firewalld microshift && \
    systemctl enable microshift && \
    dnf clean all
```

**Example: MicroShift 4.16 Release Candidate Packages (fragment)**

```docker
ARG USHIFT_VER=4.16
ARG USHIFT_REPO="microshift-${USHIFT_VER}-mirror"
ARG OCPDEP_REPO="openshift-${USHIFT_VER}-mirror"

RUN cat > "/etc/yum.repos.d/${USHIFT_REPO}.repo" <<EOF
[${USHIFT_REPO}]
name=MicroShift ${USHIFT_VER} Release Candidate Mirror
baseurl="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/microshift/ocp/latest-${USHIFT_VER}/el9/os/"
gpgcheck=0
enabled=1
EOF

RUN cat > "/etc/yum.repos.d/${OCPDEP_REPO}.repo" <<EOF
[${OCPDEP_REPO}]
name=OpenShift ${USHIFT_VER} Dependencies Mirror
baseurl="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/${USHIFT_VER}-el9-beta/"
gpgcheck=0
enabled=1
EOF

RUN dnf install -y firewalld microshift && \
    systemctl enable microshift && \
    dnf clean all
```

Finally, follow the instructions in the [Publish Image](./image_mode.md#publish-image)
section to push the MicroShift images to a container registry.

## Build and Publish Layered Product Container Image

Follow the instructions in the [Build Image](./image_mode.md#build-image) section
to implement the Layered Product container image layer build procedure.

Customize the `Containerfile` file according to the requirements of the layered
product to be tested. A typical customization would be to select a custom version
of the Layered Product and install it on top of the base MicroShift container image.

**Example: MicroShift GitOps 1.12 Packages (complete)**

```docker
FROM quay.io/myorg/mypath/microshift-4.16-bootc

ARG GITOPS_VER=1.12
RUN dnf config-manager --set-enabled gitops-${GITOPS_VER}-for-rhel-9-$(uname -m)-rpms
RUN dnf install -y microshift-gitops && \
    dnf clean all
```

> The `FROM` statement should be updated to denote a valid reference to the base
> MicroShift container image.

Finally, follow the instructions in the [Publish Image](./image_mode.md#publish-image)
section to push the Layered Product images to a container registry.

## Run Layered Product Container Image

Follow the instructions in [Run MicroShift Bootc Image](./image_mode.md#run-microshift-bootc-image)
to [Configure CNI](./image_mode.md#configure-cni), [Configure CSI](./image_mode.md#configure-csi)
and [Run Container](./image_mode.md#run-container) using the Layered Product image.

```bash
PULL_SECRET=~/.pull-secret.json
IMAGE_NAME=microshift-gitops-1.12-bootc

sudo modprobe openvswitch
sudo podman run --rm -it --privileged \
    -v "${PULL_SECRET}":/etc/crio/openshift-pull-secret:ro \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    --name "${IMAGE_NAME}" \
    "${IMAGE_NAME}"
```

Log into the running container and verify that the MicroShift and Layered Product
pods are up and running without errors.

## Run Layered Product Tests
