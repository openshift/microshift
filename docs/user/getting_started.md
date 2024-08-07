# Getting Started with MicroShift

Refer to the [MicroShift product documentation](https://access.redhat.com/documentation/en-us/red_hat_build_of_microshift) for how to install MicroShift on a machine running RHEL and how to build a RHEL for Edge image embedding MicroShift. If you do not yet have a RHEL subscription, you can get a [no-cost Red Hat Developer subscription](https://developers.redhat.com/blog/2021/02/10/how-to-activate-your-no-cost-red-hat-enterprise-linux-subscription).

The remainder of this document describes an opinionated, non-production setup to facilitate experimentation with MicroShift in a virtual machine running the RHEL 9.4 operating system.

## Prerequisites

Run the following command to install the necessary components for the [libvirt](https://libvirt.org/) virtualization platform and its [QEMU KVM](https://libvirt.org/drvqemu.html) hypervisor driver.
> **Note for Other Virtualization Platform Users** <br>
> Implement the virtual machine creation guidelines from [Bootstrap MicroShift](#bootstrap-microshift) using your virtualization platform and apply one of the following configuration steps:
> * When creating a virtual machine, pass the `inst.ks=...` boot option pointing to the [microshift-starter.ks](https://raw.githubusercontent.com/openshift/microshift/main/docs/config/microshift-starter.ks) kickstart file
> * After creating a virtual machine, manually execute the configuration steps from the [microshift-starter.ks](https://raw.githubusercontent.com/openshift/microshift/main/docs/config/microshift-starter.ks) kickstart file

```bash
sudo dnf install -y libvirt virt-manager virt-install virt-viewer libvirt-client qemu-kvm qemu-img
```

Download the Red Hat Enterprise Linux 9.4 DVD ISO image for the `x86_64` or `aarch64` architectures from [Red Hat Developer](https://developers.redhat.com/products/rhel/download) site, scroll down to see the 9.4 option.

Copy the ISO file to the `/var/lib/libvirt/images` directory.
> Other architectures, versions or flavors of operating systems are not supported in this opinionated environment.
> For this setup, only use the RHEL 9.4 DVD image for the `x86_64` or `aarch64` architectures.

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and save it into the `~/.pull-secret.json` file.

## Bootstrap MicroShift

Run the following commands to initiate the creation process of the `microshift-starter` virtual machine with 2 CPU cores, 3GB RAM and 20GB storage.

```bash
VMNAME=microshift-starter
DVDISO="/var/lib/libvirt/images/rhel-9.4-$(uname -m)-dvd.iso"
KICKSTART=https://raw.githubusercontent.com/openshift/microshift/main/docs/config/microshift-starter.ks

sudo virt-install \
    --name "${VMNAME}" \
    --vcpus 2 \
    --memory 3072 \
    --disk "path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20" \
    --network network=default,model=virtio \
    --events on_reboot=restart \
    --location "${DVDISO}" \
    --extra-args "inst.ks=${KICKSTART}" \
    --noautoconsole \
    --wait
```

The OS console is accessible from the `virt-manager` application, which can be run using the `sudo virt-manager` command.
Watch the OS console of the virtual machine to see the progress of the installation, waiting until the machine is rebooted and the login prompt appears.

## Access MicroShift

From the OS console of the virtual machine, it is possible to log into the machine using `redhat:redhat` credentials.

However, it is most convenient to access the MicroShift virtual machine using SSH.
First, get the machine IP address with the following command.

```bash
sudo virsh domifaddr microshift-starter
```

Copy your pull secret file to the MicroShift virtual machine using `redhat:redhat` credentials.

```bash
USHIFT_IP=192.168.122.2
scp ~/.pull-secret.json redhat@${USHIFT_IP}:
```

Log into the MicroShift virtual machine using `redhat:redhat` credentials.

```bash
ssh redhat@${USHIFT_IP}
```

The remaining commands are to be executed from within the virtual machine as the `redhat` user.

Register your RHEL machine and attach your subscriptions.

```bash
sudo subscription-manager register --auto-attach
```

Enable the MicroShift RPM repos and install MicroShift and the `oc` and `kubectl` clients.

```bash
sudo subscription-manager repos \
    --enable rhocp-4.16-for-rhel-9-$(uname -m)-rpms \
    --enable fast-datapath-for-rhel-9-$(uname -m)-rpms
sudo dnf install -y microshift openshift-clients
```

Confgure the minimum required firewall rules.
```bash
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
sudo firewall-cmd --reload
```

Configure `CRI-O` to use the pull secret.

```bash
sudo cp ~redhat/.pull-secret.json /etc/crio/openshift-pull-secret
```

Start the MicroShift service.

```bash
sudo systemctl enable --now microshift.service
```

Proceed by configuring MicroShift access for the `redhat` user account.

```bash
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
```

Finally, check if all the MicroShift pods are up and running by executing the `oc` command.
> When started for the first time, it may take a few minutes to download and initialize the container images used by MicroShift.
> On subsequent restarts, all the MicroShift services should take a few seconds to become available.

```bash
oc get pods -A
```
