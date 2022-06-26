# Install MicroShift on RHEL for Edge
To install MicroShift in a production environment, it is necessary to create a RHEL for Edge ISO installer with all the necessary components preloaded on the image.

## Build RHEL for Edge Installer ISO
Log into the development virtual machine with the `microshift` user credentials.
> Use insructions in [MicroShift Development Environment on RHEL 8.x](./devenv_rhel8.md) document for creating and configuring the machine.

Follow the instructions in the [RPM Packages](./devenv_rhel8.md#rpm-packages) section to create MicroShift RPM packages.

The scripts for building the installer are located in the `scripts/image-builder` subdirectory.

### Prerequisites
Execute the `scripts/image-builder/configure.sh` script to install the tools necessary for building the installer image.

Download the OpenShift pull secret from the https://console.redhat.com/openshift/downloads#tool-pull-secret page and save it into the `~microshift/pull-secret.txt` file. 

Make sure there is more than 20GB of free disk space necessary for the build artifacts. Run the following command to free the space if necessary.
```bash
./scripts/image-builder/cleanup.sh -full
```

Note that the command deletes various user and system data, including:
- The `scripts/image-builder/_builds` directory containing image build artifacts
- MicroShift `ostree` server container and all the unused container images
- All the Image Builder jobs are canceled and deleted
- Project-specific Image Builder sources are deleted
- The user `~/.cache`, `/tmp` and `/var/tmp` directory contents are deleted to clean various cache files
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

Run the build script without arguments to see its usage.
```bash
./scripts/image-builder/build.sh
Usage: build.sh <-pull_secret_file path_to_file> [-ostree_server_name name_or_ip] [-offline_containers /path/to/file1.rpm,...,/path/to/fileN.rpm]
   -pull_secret_file   Path to a file containing the OpenShift pull secret
   -ostree_server_name Name or IP address of the OS tree server (default: )
   -offline_containers Path to one or more RPM packages with offline CRI-O container images

Note: The OpenShift pull secret can be downloaded from https://console.redhat.com/openshift/downloads#tool-pull-secret.
```

Continue by running the build script with the pull secret file argument and wait until build process is finished. It may take over 30 minutes to complete a full build cycle.
```bash
./scripts/image-builder/build.sh -pull_secret_file ~/pull-secret.txt
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
oc get cs
oc get pods -A
```
