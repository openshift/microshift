The kickstart template files in this directory can be used for installing
a host running MicroShift as described in the remainder of the document.

## Procedure Overview

* Follow the instructions in [Prerequisites](#prerequisites).
* Depending on the desired installation type, follow the instructions one of the
  [RPM](#RPM), [Image Mode](#image-mode) or [OSTree](#OSTree) sections to create
  a working kickstart file from a template.
* Follow the instructions in [Create Virtual Machine](#create-virtual-machine)
  to bootstrap a host using the kickstart file you created in the previous step.

## Prepare Kickstart File

### Prerequisites

Install the `microshift-release-info` RPM package containing the sample kickstart
files that are copied to the `/usr/share/microshift/kickstart` directory.

```bash
sudo dnf install -y microshift-release-info
```

Install the utilities used for kickstart file creation.

```bash
sudo dnf install -y openssl gettext
```

Set variables pointing to secrets included in `kickstart.ks`:
* `PULL_SECRET` file contents are copied to `/etc/crio/openshift-pull-secret`
  at the post-install stage to authenticate OpenShift registry access.
* `USER_PASSWD` setting is used as an encrypted password for the `redhat` user
  for logging into the host.

```bash
export PULL_SECRET="$(cat ~/.pull-secret.json)"
PASSWD_TEXT=<my_redhat_user_password>
export USER_PASSWD="$(openssl passwd -6 "${PASSWD_TEXT}")"
```

### RPM
<TODO>

Run the following command to create the `kickstart.ks` file to be used during
the virtual machine installation.

```bash
envsubst < \
    /usr/share/microshift/kickstart/kickstart-rpm.ks.template > \
    "${HOME}/kickstart.ks"
```

### Image Mode

The following variables need to be added for creating an Image Mode kickstart file.
* `BOOTC_IMAGE_URL` contains a reference of the image to be installed using the
  [ostreecontainer](https://pykickstart.readthedocs.io/en/latest/kickstart-docs.html#ostreecontainer) kickstart command.
* `AUTH_CONFIG` contents are copied to `/etc/ostree/auth.json` at the pre-install
  stage to authenticate access to the `BOOTC_IMAGE_URL` image. If no registry
  authentication is required, skip this setting.
* `REGISTRY_CONFIG` contents are copied to `/etc/containers/registries.conf.d/999-microshift-registry.conf`
  at the pre-install stage to configure access to the registry containing the
  `BOOTC_IMAGE_URL` image. If no registry configuration is required, skip this
  setting.

```bash
export BOOTC_IMAGE_URL=quay.io/myorg/mypath/microshift-image:tag
export AUTH_CONFIG="$(cat ~/.quay-auth.json)"
export REGISTRY_CONFIG="$(cat ~/.quay-config.conf)"
```

Run the following command to create the `kickstart.ks` file to be used during
the virtual machine installation.

```bash
envsubst < \
    /usr/share/microshift/kickstart/kickstart-bootc.ks.template > \
    "${HOME}/kickstart.ks"
```

### OSTree
<TODO>

Run the following command to create the `kickstart.ks` file to be used during
the virtual machine installation.

```bash
envsubst < \
    /usr/share/microshift/kickstart/kickstart-ostree.ks.template > \
    "${HOME}/kickstart.ks"
```

## Create Virtual Machine

Download a RHEL boot ISO image from https://developers.redhat.com/products/rhel/download.
Copy the downloaded file to the `/var/lib/libvirt/images` directory.

Run the following commands to create a RHEL virtual machine with 2 cores, 2GB of
RAM and 20GB of storage. The command uses the kickstart file prepared in the
previous steps to install the RHEL operating system and MicroShift.

```bash
VMNAME=microshift-host
NETNAME=default

sudo virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 2048 \
    --disk path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --location /var/lib/libvirt/images/rhel-9.4-$(uname -m)-boot.iso \
    --initrd-inject "${HOME}/kickstart.ks" \
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
