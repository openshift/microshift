# Mirror MicroShift Container Images

When deploying MicroShift in [air gapped networks](https://en.wikipedia.org/wiki/Air_gap_(networking))
it is often necessary to use a custom container registry server because the access
to the Internet is not allowed.

Note that it is possible to embed the container images in the MicroShift ISO and
also in the subsequent `ostree` updates by using the `-embed_containers` option
of the `scripts/image-builder/build.sh` script. Such ISO images and updates can
be transferred to air gapped environments and installed on MicroShift instances.

> The container embedding procedures are described in the
> [Offline Containers](../contributor/rhel4edge_iso.md#offline-containers) and
> [The `ostree` Update Server](../contributor/rhel4edge_iso.md#the-ostree-update-server)
> sections.

However, a custom air gapped container registry may still be necessary due to
the user environment and workload requirements. 

This document describes how to mirror MicroShift container images into an existing
registry in an air gapped environment.

## Mirror Images to Container Registry

Mirroring container images to an air gapped site involves the following steps:
* Obtain the [Container Image List](#container-image-list) to be mirrored
* Configure the [Mirroring Prerequisites](#mirroring-prerequisites)
* [Download Images](#download-images) on a host with the Internet access
* Copy the downloaded image directory to an air gapped site
* [Upload Images](#upload-images) to a mirror registry in an air gapped site

## Container Image List
The list of the container image references used by a specific version of MicroShift
is provided in the `release-<arch>.json` files that are part of the
`microshift-release-info` RPM package.

If the package is installed on a MicroShift host, the files can be accessed at
the following location.
```
$ rpm -ql microshift-release-info
/usr/share/microshift/release/release-aarch64.json
/usr/share/microshift/release/release-x86_64.json
```

Alternatively, download and unpack the RPM package without installing it.
```
$ rpm2cpio microshift-release-info*.noarch.rpm | cpio -idmv
./usr/share/microshift/release/release-aarch64.json
./usr/share/microshift/release/release-x86_64.json
```

The list of container images can be extracted into the `microshift-container-refs.txt`
file using the following command.
```
RELEASE_FILE=/usr/share/microshift/release/release-$(uname -m).json
jq -r '.images | .[]' ${RELEASE_FILE} > ~/microshift-container-refs.txt
```

> After the `microshift-container-refs.txt` file is created with the MicroShift
> container image list, other user-specific image references can be appended to
> the file before the mirroring procedure is run.

## Mirroring Prerequisites

Follow the instructions in the [Configuring credentials that allow images to be mirrored](https://docs.openshift.com/container-platform/latest/installing/disconnected_install/installing-mirroring-disconnected.html#installation-adding-registry-pull-secret_installing-mirroring-disconnected)
document to create a `~/.pull-secret-mirror.json` file containing the user credentials
for accessing the mirror.

As an example, the following section should be added to the pull secret file for
the `microshift-quay:8443` mirror registry using `microshift:microshift` user name
and password.
```
    "microshift-quay:8443": {
      "auth": "bWljcm9zaGlmdDptaWNyb3NoaWZ0",
      "email": "microshift-quay@example.com"
    },
```

## Download Images

> Install the `skopeo` tool used for copying the container images.
 
Run the `./scripts/mirror-images.sh` script with `--reg-to-dir`
option to initiate the image download procedure into a local directory on a
host with the Internet connection.
```
IMAGE_PULL_FILE=~/.pull-secret-mirror.json
IMAGE_LIST_FILE=~/microshift-container-images.txt
IMAGE_LOCAL_DIR=~/microshift-containers

mkdir -p "${IMAGE_LOCAL_DIR}"
./scripts/mirror-images.sh --reg-to-dir "${IMAGE_PULL_FILE}" "${IMAGE_LIST_FILE}" "${IMAGE_LOCAL_DIR}"
```

The contents of the local directory can now be transferred to an air gapped site
and imported into the mirror registry.

## Upload Images

> Install the `skopeo` tool used for copying the container images.

Run the `./scripts/mirror-images.sh` script with `--dir-to-reg` option
in the air gapped environment to initiate the image upload procedure from a local
directory to a mirror registry.
```
IMAGE_PULL_FILE=~/.pull-secret-mirror.json
IMAGE_LOCAL_DIR=~/microshift-containers
TARGET_REGISTRY=microshift-quay:8443

./scripts/mirror-images.sh --dir-to-reg "${IMAGE_PULL_FILE}" "${IMAGE_LOCAL_DIR}" "${TARGET_REGISTRY}"
```
