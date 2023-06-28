# MicroShift on RHEL for Edge
Review the [Install MicroShift on RHEL for Edge](../rhel4edge_iso.md)
document to understand the process of building MicroShift ISO images in the
MicroShift development environment.

## Prerequisites
A `development host` (either virtual or physical) configured according to the
instructions in the [MicroShift Development Environment](../devenv_setup.md)
document.

> While it is possible to configure the MicroShift development environment on
> your hypervisor host, it is recommended to use a dedicated virtual machine
> to simplify the installation and configuration process.

## Create ISO Image
Log into the `development host` and run the following commands to download
the released MicroShift RPMs to be included in the ISO image.
```
MICROSHIFT_VERSION=4.14
MICROSHIFT_RPMS_DIR=$(mktemp -d /tmp/microshift-rpms.XXXXXXXXXX)
OCP_REPO_NAME="rhocp-${MICROSHIFT_VERSION}-for-rhel-9-$(uname -m)-rpms"

sudo subscription-manager repos --enable ${OCP_REPO_NAME}
sudo dnf download --downloaddir "${MICROSHIFT_RPMS_DIR}" microshift\*
chmod go+rx "${MICROSHIFT_RPMS_DIR}"
```

> Until the MicroShift 4.14 software is released, it is necessary to compile the
> MicroShift RPMs from the latest sources on the `development host` and copy them
> into the `${MICROSHIFT_RPMS_DIR}` directory.
>
> See the [RPM Packages](../devenv_setup.md#rpm-packages) documentation
> for more information on building MicroShift RPMs.

Proceed by building the ISO image with your pull secret, the downloaded MicroShift
RPMS, your SSH authorized keys and the firewall configuration necessary for the
multiple nodes to access each other.
```
cd ~/microshift/
./scripts/image-builder/build.sh \
    -pull_secret_file ~/.pull-secret.json \
    -microshift_rpms "${MICROSHIFT_RPMS_DIR}" \
    -authorized_keys_file ~/.ssh/authorized_keys \
    -open_firewall_ports 6443:tcp,9642:tcp
rm -rf "${MICROSHIFT_RPMS_DIR}"
```

> The multinode configuration procedure uses SSH for copying files among the
> nodes and remotely running scripts on the primary and secondary hosts.
> Having your SSH keys authorized in the image would make the node configuration
> procedure more convenient and streamlined.

## Install Virtual Machines
Log into the `hypervisor host` and follow the instructions described in the
[Install MicroShift for Edge](../rhel4edge_iso.md#install-microshift-for-edge)
section to create the `microshift-pri` and `microshift-sec` virtual machines
for running primary and secondary instances.

After the virtual machines are up and running, use the `virsh` command to determine their IP addresses.
```
sudo virsh domifaddr microshift-pri
sudo virsh domifaddr microshift-sec
```
