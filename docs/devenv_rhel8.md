## Create Development Virtual Machine
Start by downloading the RHEL 8.6 or above ISO image from the https://developers.redhat.com/products/rhel/download location. 

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

Run the following commands to upgrade the system, install basic dependencies, enable remote Cockpit console and configure SUDO.
```bash
sudo dnf update -y
sudo dnf install -y git cockpit make golang
sudo systemctl enable --now cockpit.socket
echo -e 'microshift\tALL=(ALL)\tNOPASSWD: ALL' | sudo tee -a /etc/sudoers
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
oc status
oc get pods -A
```
### Stopping MicroShift
Run the following command to stop the MicroShift and make sure it is shut down by examining its log file.
```bash
sudo kill microshift && sleep 3
tail -3 /tmp/microshift.log 
```

Note tha this command only stops the MicroShift executable. To perform full cleanup including CRI-O images, run the following script.
```bash
./hack/cleanup.sh
```

## Build RHEL for Edge Installer ISO
Log into the development virtual machine with the `microshift` user credentials.

Follow the instructions in the RPM Packages section to build the binary and create MicroShift RPM packages.

The scripts for building the installer are located in the `scripts/image-builder` subdirectory.

### Prerequisites
Execute the `configure.sh` script to install the tools necessary for building the installer image.

Make sure there is more than 20GB of free disk space necessary for the build artifacts. Run the following command to free the space if necessary.
```bash
./scripts/image-builder/cleanup.sh -full
```

Note that the command deletes various user and system data, including:
- The `scripts/image-builder/_builds` directory containing image build artifacts
- MicroShift `ostree` server container and all the unused container images
- All the Image Builder jobs are canceled and deleted
- Project-specific Image Builder sources are deleted
- The user `~/.cache` directory is deleted to clean Golang cache
- The `/var/cache/osbuild-worker` directory contents are deleted to clean Image Builder cache files

### Building Installer
It is recommended to execute the partial cleanup procedure (without `-full` argument) before each build to reclaim cached disk space from previous builds. 
```bash
./scripts/image-builder/cleanup.sh
```
Note that the command deletes various user and system data, including:
- MicroShift `ostree` server container
- All the Image Builder jobs are canceled
- Project-specific Image Builder sources are deleted
- The `/var/cache/osbuild-worker` directory contents are deleted to clean Image Builder cache files

Continue by running the build script and wait until it is finished. It may take over 30 minutes to complete a full build cycle.
```bash
./scripts/image-builder/build.sh
```
The script performs the following tasks:
- Check for minimum 10GB of available disk space
- Set up local MicroShift RPM repository using `packaging/rpm/_rpmbuild` directory contents as a source
- Set up local OpenShift RPM repository using public `rhocp-${OCP_VERSION}-for-rhel-8-${ARCH}-rpms` repository necessary for CRI-O and OpenShift client package installation
- Configure Image Builder to use the local MicroShift and OpenShift RPM repositories for image builds
- Run Image Builder to create an edge container image containing all the MicroShift binaries and dependencies
- Start a local `ostree` server with the above image
- Run Image Builder to create an edge installer image comprised of RHEL for Edge OS and MicroShift binaries
- Rebuild the installer image with the `kickstart.ks` file for performing various OS setup when the host is first booted
- Perform partial cleanup procedure to reclaim cached disk space

The artifact of the build is the `scripts/image-builder/_builds/microshift-installer.${ARCH}.iso` bootable RHEL for Edge OS image.

## Install MicroShift for Edge
Log into the host machine using your user credentials. The remainder of this section describes how to install a virtual machine running RHEL for Edge OS containing MicroShift binaries.

Start by copying the installer image from the development virtual machine to the host file system.
```bash
sudo scp microshift@microshift-dev:/home/microshift/microshift/scripts/image-builder/_builds/microshift-installer.*.iso /var/lib/libvirt/images/
```

Run the following commands to create a virtual machine using the installer image.
```bash
VMNAME="microshift-edge"
sudo -b bash -c " \
cd /var/lib/libvirt/images/ && \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 3096 \
    --disk path=./${VMNAME}.qcow2,size=20 \
    --network network=default,model=virtio \
    --os-type generic \
    --events on_reboot=restart \
    --cdrom ./microshift-installer.$(uname -i).iso \
"
```

Watch the OS console to see the progress of the installation, waiting until the machine is rebooted and the login prompt appears. 

Note that it may be more convenient to access the machine using SSH. Run the following command to get its IP address and use it to remotely connect to the system.
```bash
sudo virsh domifaddr microshift-edge
```

Log into the system using `redhat:redhat` credentials and run the following commands to configure the MicroShift access.
```bash
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
```

Finally, check if the MicroShift is up and running by executing `oc` commands.
```bash
oc status
oc get pods -A
```

## Quick Development and Edge Testing Cycle
In the development environment, it is practical to build and run MicroShift executable as described in the [Build MicroShift](#build-microshift) and [Run MicroShift Executable](#run-microshift-executable) sections above. However, it is also necessary to have a convenient technique to occasionally test the system in a RHEL for Edge environment.

The MicroShift Installer ISO build procedure takes over 30 minutes to complete. Once a RHEL for Edge virtual machine is created, it is running a version of MicroShift with the latest changes. 

The remainder of this section describes a simple procedure for replacing the MicroShift executable file on an existing RHEL for Edge OS installation.

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

Rebuild the MicroShift executable as described in the [Build MicroShift](#build-microshift) section and run the following commands to copy, cleanup and restart the new service on the RHEL for Edge system.
```bash
scp ~/microshift/microshift redhat@microshift-edge:
ssh redhat@microshift-edge ' \
    sudo systemctl stop microshift && \
    sleep 3 && \
    sudo cp ~redhat/microshift /usr/bin/microshift && \
    echo 1 | /usr/bin/cleanup-all-microshift-data && \
    sudo systemctl enable microshift --now && \
    echo Done '
```
