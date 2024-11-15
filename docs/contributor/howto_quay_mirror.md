# Quay Mirror Registry Setup for Testing

This section describes an opinionated, non-production setup to facilitate the
configuration of a custom container registry virtual server for MicroShift,
simulating air gapped environments. Such a setup can be used in the development
or testing environments to experiment with container image mirroring.

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

VMNAME=microshift-quay
ISO=/var/lib/libvirt/images/rhel-9.2-$(uname -m)-dvd.iso

./scripts/devenv-builder/manage-vm.sh create -n ${VMNAME} -i ${ISO}
```
After the virtual machine installation is finished, the `manage-vm.sh` script
prompts for a user name and password to register the operating system with a
Red Hat subscription.

Attach a second interface to the  `microshift-quay` virtual machine in the isolated network:
```
NETWORK=isolated
VMNAME=microshift-quay
virsh  attach-interface ${VMNAME} --type network --source ${NETWORK} --model virtio --config --live
```


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

## Mirror Container Images

Log into the hypervisor host and follow the steps below to mirror the container
images to the `microshift-quay` host.
* Obtain the [Container Image List](../user/howto_mirror_images.md#container-image-list) to be mirrored
* Configure the [Mirroring Prerequisites](../user/howto_mirror_images.md#mirroring-prerequisites)
* [Download Images](../user/howto_mirror_images.md#download-images) to a local directory
* [Upload Images](../user/howto_mirror_images.md#upload-images) to the `microshift-quay` host

> Make sure to resolve the `microshift-quay` host name on the hypervisor host and
> enable the mirror registry certificate trust as described in the
> [Configure Certificates](#configure-certificates) section.

## Build MicroShift on RHEL for Edge

Follow the instructions in the [Build RHEL for Edge Installer ISO](./rhel4edge_iso.md#build-rhel-for-edge-installer-iso)
document for creating the MicroShift installer ISO.

> TODO: The `image-builder/build.sh` script has been deprecated.
> This section will be rewritten in the context of [USHIFT-4289](https://issues.redhat.com/browse/USHIFT-4289).

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
