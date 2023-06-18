# Quay Mirror Registry Setup for MicroShift

When deploying MicroShift in [air gapped networks](https://en.wikipedia.org/wiki/Air_gap_(networking))
it is often necessary to set up a custom container registry server because the
access to the Internet is not allowed.

Note that it is possible to embed the container images in the MicroShift ISO and
also in the subsequent `ostree` updates by using the `-embed_containers` option
of the `scripts/image-builder/build.sh` script. Such ISO images and updates can
be transferred to air gapped environments and installed on MicroShift instances.

> The container embedding procedures are described in the
> [Offline Containers](./rhel4edge_iso.md#offline-containers) and
> [The `ostree` Update Server](./rhel4edge_iso.md#the-ostree-update-server)
> sections.

However, a custom air gapped container registry may still be necessary due to
the user environment and workload requirements. The remainder of this document
describes an opinionated, non-production setup to facilitate the configuration
of a custom container registry virtual server for MicroShift, which can also be
used in air gapped environments.

## Prerequisites

The following main components are used for the mirror container registry setup.
* A hypervisor host running the [libvirt](https://libvirt.org/) virtualization platform
* A virtual machine for running the `mirror registry for Red Hat OpenShift`
* A virtual machine for running MicroShift that uses the mirror registry

> Note that the `mirror registry for Red Hat OpenShift` is only supported on the
> `x86_64` architecture.

Log into the hypervisor host and download the latest MicroShift repository from
https://github.com/openshift/microshift. The scripts used in this document are
part of that repository and they are run relative to its root directory.
```
git clone https://github.com/openshift/microshift ~/microshift
cd ~/microshift
```

Create an isolated network as described in the [Offline Mode](./rhel4edge_iso.md#offline-mode)
document. It will be used by the virtual machines to make sure they cannot access
the Internet.

## Create Mirror Registry Host

Log into the hypervisor host and download the RHEL 9.2 DVD image for the `x86_64`
architecture from the https://developers.redhat.com/products/rhel/download site.

Run the following commands to create the `microshift-quay` virtual machine with
2 cores, 6GB RAM and 30GB of disk.
```bash
export NCPUS=2
export RAMSIZE=6
export DISKSIZE=30
export SWAPSIZE=0
export NETWORK=isolated

VMNAME=microshift-quay
ISO=/var/lib/libvirt/images/rhel-9.2-$(uname -m)-dvd.iso

./scripts/devenv-builder/manage-vm.sh create -n ${VMNAME} -i ${ISO}
```

After the virtual machine installation is finished, the `manage-vm.sh` script
prompts for a user name and password to register the operating system with a
Red Hat subscription.

## Install Mirror Registry Service

Download the [mirror registry for Red Hat OpenShift](https://console.redhat.com/openshift/downloads#tool-mirror-registry)
and copy the archive to the `microshift-quay` host.

```
QUAY_IP=192.168.111.128
scp mirror-registry.tar.gz microshift@${QUAY_IP}:
```

Log into the `microshift-quay` host using `microshift:microshift` credentials and
run the following commands to unpack and install the Quay mirror registry.

```
MIRROR_HOST=microshift-quay
MIRROR_USER=microshift
MIRROR_PASS=microshift
MIRROR_ROOT="/var/lib/quay-root"

mkdir -p ~/mirror-registry
cd ~/mirror-registry
tar zxf ~/mirror-registry.tar.gz

sudo dnf install -y podman
sudo ./mirror-registry install \
    --quayHostname "${MIRROR_HOST}" \
    --initUser     "${MIRROR_USER}" \
    --initPassword "${MIRROR_PASS}" \
    --quayRoot     "${MIRROR_ROOT}"
```

> See the [Creating a mirror registry for Red Hat OpenShift](https://docs.openshift.com/container-platform/latest/installing/disconnected_install/installing-mirroring-creating-registry.html)
> documentation for more information on how to install and configure the mirror registry.

## Configure Certificates

The mirror registry installer automatically generates an SSH key and an SSL
certificate unless existing certificate files are specified from the command
line using the `--ssh-key` and `--sslCert` arguments.

The default keys and certificates are found in the `~/.ssh` and `quay-rootCA`
subdirectory under the Quay root (i.e. `/var/lib/quay-root/quay-rootCA`).

It is necessary to enable the SSL certificate trust on any host accessing the
mirror registry. Copy the `rootCA.pem` file from the mirror registry to the
target host at the `/etc/pki/ca-trust/source/anchors` directory and run the
`update-ca-trust` command to enable the certificate in the system-wide trust
store configuration.

As an example, the following commands can be used to enable the certificate
trust locally on the `microshift-quay` host.
```
MIRROR_ROOT="/var/lib/quay-root"

sudo cp ${MIRROR_ROOT}/quay-rootCA/rootCA.pem /etc/pki/ca-trust/source/anchors/microshift-quay.pem
sudo update-ca-trust
```

Test the connection to the mirror registry running on the `microshift-quay` host.
The `curl` command should return no errors if the mirror registry service is up
and running and the certificates are trusted.
```
MIRROR_HOST=microshift-quay:8443
MIRROR_USER=microshift
MIRROR_PASS=microshift

curl -I -u ${MIRROR_USER}:${MIRROR_PASS} https://${MIRROR_HOST}
```

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

> Optionally use the `scripts/image-builder/download-rpms.sh` script for
> downloading the released version of MicroShift RPM packages.

The list of container images can be extracted into the `microshift-container-refs.txt`
file using the following command.
```
RELEASE_FILE=/usr/share/microshift/release/release-$(uname -m).json
jq -r '.images | .[]' ${RELEASE_FILE} > ~/microshift-container-refs.txt
```

> After the `microshift-container-refs.txt` file is created with the MicroShift
> container image list, other user-specific image references can be appended to
> the file before the mirroring procedure is run.

## Mirror Container Images

Log into the hypervisor host and install the `skopeo` tool to be used for copying
the container images.
```
sudo dnf install -y skopeo
```

Follow the instructions in the [Configuring credentials that allow images to be mirrored](https://docs.openshift.com/container-platform/latest/installing/disconnected_install/installing-mirroring-disconnected.html#installation-adding-registry-pull-secret_installing-mirroring-disconnected)
document to create a `~/.pull-secret-mirror.json` file containing the user credentials
for accessing the `microshift-quay` mirror. Use `microshift:microshift` user name
and password as an input for the `base64` command.

As an example, the following section should be added to the pull secret file.
```
    "microshift-quay:8443": {
      "auth": "bWljcm9zaGlmdDptaWNyb3NoaWZ0",
      "email": "microshift-quay@example.com"
    },
```

> Make sure to resolve the `microshift-quay` host name on the hypervisor host and
> enable the mirror registry certificate trust as described in the
> [Configure Certificates](#configure-certificates) section.

### Scenario 1: Transfer Virtual Machine

If the `microshift-quay` virtual machine image can be installed at an air gapped
site, it is recommended to mirror images directly into the target registry.

Run the `./scripts/image-builder/mirror-images.sh` script with `--mirror` option
to initiate the image copy procedure into the mirror registry without saving them
locally on the hypervisor host.
```
IMAGE_PULL_FILE=~/.pull-secret-mirror.json
IMAGE_LIST_FILE=~/microshift-container-images.txt
TARGET_REGISTRY=microshift-quay:8443

./scripts/image-builder/mirror-images.sh --mirror "${IMAGE_PULL_FILE}" "${IMAGE_LIST_FILE}" "${TARGET_REGISTRY}"
```

The `microshift-quay` virtual machine image can now be transferred to an air
gapped site, installed and used as a MicroShift mirror registry.

### Scenario 2: Transfer Image Directory

If the `microshift-quay` virtual machine image cannot be installed at an air gapped
site, it is necessary to perform the image mirroring in two steps.

Run the `./scripts/image-builder/mirror-images.sh` script with `--reg-to-dir` option
to initiate the image download procedure into a local directory on the hypervisor host.
```
IMAGE_PULL_FILE=~/.pull-secret-mirror.json
IMAGE_LIST_FILE=~/microshift-container-images.txt
IMAGE_LOCAL_DIR=~/microshift-containers

mkdir -p "${IMAGE_LOCAL_DIR}"
./scripts/image-builder/mirror-images.sh --reg-to-dir "${IMAGE_PULL_FILE}" "${IMAGE_LIST_FILE}" "${IMAGE_LOCAL_DIR}"
```

The contents of the local directory can now be transferred to an air gapped site
and imported into the mirror registry.

Run the `./scripts/image-builder/mirror-images.sh` script with `--dir-to-reg` option
to initiate the image upload procedure from a local directory to a mirror registry.
```
IMAGE_PULL_FILE=~/.pull-secret-mirror.json
IMAGE_LOCAL_DIR=~/microshift-containers
TARGET_REGISTRY=microshift-quay:8443

./scripts/image-builder/mirror-images.sh --dir-to-reg "${IMAGE_PULL_FILE}" "${IMAGE_LOCAL_DIR}" "${TARGET_REGISTRY}"
```

## Build MicroShift on RHEL for Edge

Follow the instructions in the [Build RHEL for Edge Installer ISO](./rhel4edge_iso.md#build-rhel-for-edge-installer-iso)
document for creating the MicroShift installer ISO.

Use the following command line arguments for the `scripts/image-builder/build.sh`
script when building the installer ISO.
* `-pull_secret_file` with a pull secret containing the mirror registry credentials
* `-microshift_rpms` pointing to the version of RPMs with mirrored container images
* `-mirror_registry_host` pointing to the mirror registry host
* `ca_trust_files` pointing to the `rootCA.pem` file from the mirror registry host

```
PULL_SECRET_FILE=~/.pull-secret-mirror.json
MICROSHIFT_RPMS_DIR=~/microshift-rpms
MIRROR_REGISTRY_HOST=microshift-quay:8443
CA_TRUST_FILES=~/microshift-mirror-rootCA.pem

./scripts/image-builder/build.sh \
    -pull_secret_file     "${PULL_SECRET_FILE}"    \
    -microshift_rpms      "${MICROSHIFT_RPMS_DIR}" \
    -mirror_registry_host "${MIRROR_REGISTRY_HOST} \
    -ca_trust_files       "${CA_CA_TRUST_FILES}"
```

## Install MicroShift on RHEL for Edge

Log into the hypervisor host and run the following command to create the `microshift-edge`
virtual machine. Make sure to use an isolated network as described in the
[Offline Mode](./rhel4edge_iso.md#offline-mode) section to prevent it from
accessing the Internet.
```
VMNAME=microshift-edge
NETNAME=isolated
ISONAME=microshift-installer-4.13.1.$(uname -m).iso

./scripts/image-builder/create-vm.sh "${VMNAME}" "${NETNAME}" "${ISONAME}"
```

After the virtual machine is created, log into the system using `redhat:redhat`
credentials and verify that the Internet is not accessible.
```bash
$ curl -I redhat.com
curl: (6) Could not resolve host: redhat.com
```

Finally, wait until all the MicroShift pods are up and running.
```bash
watch sudo $(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A
```
