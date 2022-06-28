# MicroShift Development Environment on RHEL 8.x

## Create Development Virtual Machine
Start by downloading the RHEL 8.6 or above ISO image from the https://developers.redhat.com/products/rhel/download location. 
> RHEL 9.x operating system is not currently supported.

### Creating VM
Create a RHEL 8.x virtual machine with 2 cores, 4096MB of RAM and 40GB of storage. Move the ISO image to `/var/lib/libvirt/images` directory and run the following command to create a virtual machine.
```bash
VMNAME="microshift-dev"
sudo -b bash -c " \
cd /var/lib/libvirt/images/ && \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 4096 \
    --disk path=./${VMNAME}.qcow2,size=40 \
    --network network=default,model=virtio \
    --os-type generic \
    --events on_reboot=restart \
    --cdrom ./rhel-8.6-$(uname -i)-boot.iso \
"
```

In the OS installation wizard, set the following options:
- Root password and `microshift` administrator user
- In the Installation Destination, select automatic partitioning on the disk without encryption
- Connect network card and set the hostname (i.e. `microshift-dev.localdomain`)
- Register the system with Red Hat using your credentials (toggle off Red Hat Insights connection)
- In the Software Selection, select Minimal Install base environment and toggle on Headless Management to enable Cockpit

Click on Begin Installation and wait until the OS installation is complete.

### Configuring VM
Log into the virtual machine using SSH with the `microshift` user credentials.

Run the following commands to configure SUDO, upgrade the system, install basic dependencies and enable remote Cockpit console.
```bash
sudo echo -e 'microshift\tALL=(ALL)\tNOPASSWD: ALL' > /etc/sudoers.d/microshift
sudo dnf update -y
sudo dnf install -y git cockpit make golang
sudo systemctl enable --now cockpit.socket
```
You should now be able to access the VM Cockpit console using `https://<vm_ip>:9090` URL.

## Build MicroShift 
Log into the development virtual machine with the `microshift` user credentials.
Clone the repository to be used for building various artifacts.
```bash
git clone https://github.com/redhat-et/microshift.git
cd microshift
```

### Executable
Run `make` command in the top-level directory. If necessary, add `DEBUG=true` argument to the `make` command for building a binary with debug symbols.
```bash
make
```
The artifact of the build is the `microshift` executable file located in the top level directory of the source tree.

### RPM Packages
Run make command with the `rpm` or `srpm` argument in the top-level directory. 
```bash
make rpm
make srpm
```

The artifacts of the build are located in the `packaging` directory.
```bash
$ find packaging -name \*.rpm
packaging/rpm/_rpmbuild/RPMS/x86_64/microshift-4.10.0-nightly_1654189204_34_gc871db21.el8.x86_64.rpm
packaging/rpm/_rpmbuild/RPMS/noarch/microshift-selinux-4.10.0-nightly_1654189204_34_gc871db21.el8.noarch.rpm
packaging/rpm/_rpmbuild/SRPMS/microshift-4.10.0-nightly_1654189204_34_gc871db21.el8.src.rpm
```

## Run MicroShift Executable
Log into the development virtual machine with the `microshift` user credentials.
### Installing Clients
Run the following commands to install `oc` and `kubectl` utilities.
```bash
curl -O https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/clients/ocp/stable/openshift-client-linux.tar.gz
sudo tar -xf openshift-client-linux.tar.gz -C /usr/local/bin oc kubectl
rm -f openshift-client-linux.tar.gz
```
### Runtime Prerequisites
Run the following commands to install CRI-O.
```bash
sudo subscription-manager repos --enable rhocp-4.10-for-rhel-8-$(uname -i)-rpms
sudo dnf install -y cri-o cri-tools
sudo systemctl enable crio --now
```

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and copy it into the `/etc/crio/openshift-pull-secret` file. 

Run the following commands to configure CRI-O for using the pull secret when fetching container images.
```bash
sudo chmod 600 /etc/crio/openshift-pull-secret
sudo mkdir -p /etc/crio/crio.conf.d/
sudo cp ~microshift/microshift/packaging/crio.conf.d/microshift.conf /etc/crio/crio.conf.d/
sudo systemctl restart crio && sleep 3
echo 1 | ~microshift/microshift/hack/cleanup.sh
```

### Running MicroShift
Run the MicroShift in the background using the following command.
```bash
nohup sudo ./microshift run >> /tmp/microshift.log &
```
Examine the `/tmp/microshift.log` log file to ensure successful startup.

Copy `kubeconfig` to the default location that can be accessed without the administrator privilege.
```bash
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
```

Verify that the MicroShift is running.
```bash
oc get cs
oc get pods -A
```
### Stopping MicroShift
Run the following command to stop the MicroShift and make sure it is shut down by examining its log file.
```bash
sudo kill microshift && sleep 3
tail -3 /tmp/microshift.log 
```

Note that this command only stops the MicroShift executable. To perform full cleanup including CRI-O images, run the following script.
```bash
./hack/cleanup.sh
```

## Quick Development and Edge Testing Cycle
During the development cycle, it is practical to build and run MicroShift executable as demonstrated in the [Build MicroShift](#build-microshift) and [Run MicroShift Executable](#run-microshift-executable) sections above. However, it is also necessary to have a convenient technique for testing the system in a setup resembling the production environment. Such an environment can be created in a virtual machine as described in the [Install MicroShift on RHEL for Edge](./rhel4edge_iso.md) document. 

Once a RHEL for Edge virtual machine is created, it is running a version of MicroShift with the latest changes. When MicroShift code is updated and the executable file is rebuilt with the new changes, the updates need to be installed on RHEL for Edge OS.

Since it takes a long time to create a new RHEL for Edge installer ISO and deploy it on a virtual machine, the remainder of this section describes a simple technique for replacing the MicroShift executable file on an existing RHEL for Edge OS installation.

### Configuring ostree
Log into the RHEL for Edge machine using `redhat:redhat` credentials. Run the following command for configuring the ostree to allow transient overlays on top of the /usr directory.
```bash
sudo rpm-ostree usroverlay
```

This would enable a development mode where users can overwrite `/usr` directory contents. Note that all changes will be discarded on reboot.

### Updating MicroShift Executable
Log into the development virtual machine with the `microshift` user credentials. 

It is recommended to update the local `/etc/hosts` to resolve the `microshift-edge` host name. Also, generate local SSH keys and allow the `microshift` user to run SSH commands without the password on the RHEL for Edge machine.
```bash
ssh-keygen
ssh-copy-id redhat@microshift-edge
```

Rebuild the MicroShift executable as described in the [Build MicroShift](#build-microshift) section and run the following commands to copy, cleanup, replace and restart the new service on the RHEL for Edge system.
```bash
scp ~/microshift/microshift redhat@microshift-edge:
ssh redhat@microshift-edge ' \
    echo 1 | /usr/bin/cleanup-all-microshift-data && \
    sudo cp ~redhat/microshift /usr/bin/microshift && \
    sudo systemctl enable microshift --now && \
    echo Done '
```
