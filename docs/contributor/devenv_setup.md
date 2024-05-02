# MicroShift Development Environment
This document describes how to setup the MicroShift development environment running
in a virtual machine.

## Create Development Virtual Machine
Start by downloading one of the DVD ISO images for the `x86_64` or `aarch64` architecture:
* RHEL 9.4 from https://developers.redhat.com/products/rhel/download
* CentOS 9 Stream from https://www.centos.org/download

### Creating VM
Log into the hypervisor host and run the following commands to create a RHEL virtual
machine with 4 cores, 8GB of RAM and 100GB of storage.

> See [Increase Virtual Machine Disk Size](#increase-virtual-machine-disk-size) section
> for increasing the storage size if necessary.

Move the DVD ISO image to `/var/lib/libvirt/images` directory and run the following
commands to install the `libvirt` packages and create a virtual machine.
```
VMNAME="microshift-dev"

git clone https://github.com/openshift/microshift.git ~/microshift
cd ~/microshift
./scripts/devenv-builder/manage-vm.sh config
./scripts/devenv-builder/manage-vm.sh create -n ${VMNAME}
```

Note that the `manage-vm.sh` script can also be used on the hypervisor host to
manipulate the virtual machines where MicroShift will run.
* Use the `create` subcommand to create a new virtual machine
* Use the `delete` subcommand to remove a virtual machine
* Use the `ip` subcommand to find the IP address of a virtual machine

If your host is configured with virtual machine disk storage somewhere other than the
default, set `MICROSHIFT_VMDISKDIR` to point to the location (e.g. "${HOME}/VMs").

Set `MICROSHIFT_SSH_KEY_FILE` to the path to an SSH public key to have that key
added to the virtual machine for easier login.

### Configuring VM
After the virtual machine installation is finished, the `manage-vm.sh` script
prompts for a user name and password to register the operating system with a
Red Hat subscription.

> A valid Red Hat subscription is required before running the virtual machine
> configuration procedure.

Proceed by downloading the OpenShift pull secret from the
https://console.redhat.com/openshift/downloads#tool-pull-secret page and store
it in the `~/.pull-secret.json` file.

The pull secret will also be used in the `CRI-O` configuration for pulling the
MicroShift container images.

Run the following command to configure SSH, SUDO, upgrade the system, firewall,
build and install MicroShift including its runtime dependencies, Kubernetes client
utilities, and enable remote Cockpit console.

```bash
VMIPADDR=$(./scripts/devenv-builder/manage-vm.sh ip -n ${VMNAME})
# Enter 'microshift' without quotes as a password if prompted
ssh-copy-id -f microshift@${VMIPADDR}

scp ./scripts/devenv-builder/configure-vm.sh microshift@${VMIPADDR}:
scp ~/.pull-secret.json microshift@${VMIPADDR}:
ssh -t microshift@${VMIPADDR} "~/configure-vm.sh ~/.pull-secret.json"
```

> The script prompts for Red Hat subscription credentials if the system has not
> been registered yet.

The configuration script should run unattended and finish with the following message.
```
The configuration phase completed. Run the following commands to:
 - Wait until all MicroShift pods are running
      watch sudo $(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A

 - Get MicroShift logs
      sudo journalctl -u microshift

 - Get microshift-etcd logs
      sudo journalctl -u microshift-etcd.scope

 - Clean up MicroShift service configuration
      echo 1 | sudo /usr/bin/microshift-cleanup-data --all

Done
```

> You should now be able to access the virtual machine Cockpit console using
> the `https://${VMIPADDR}:9090` URL.

## Build and Run MicroShift
Log into the development virtual machine with the `microshift:microshift` user
credentials.

The `https://github.com/openshift/microshift.git` MicroShift repository is already
available at `~/microshift` directory as an artifact of the `configure-vm.sh`
script executed in the previous steps.

### Building MicroShift Executable
Run `make` command in the top-level directory. If necessary, add `DEBUG=true`
argument to the `make` command for building a binary with debug symbols.
```bash
cd ~/microshift
make clean
make
```

The artifact of the build is the `microshift` executable file located in the
`_output/bin` directory.

### Building RPM Packages
Run make command with the `rpm` or `srpm` argument in the top-level directory.
```bash
cd ~/microshift
make rpm
make srpm
```

The artifacts of the build are located in the `_output/rpmbuild` directory.

### Running MicroShift
Stop the MicroShift service and clean up its configuration data before running
a standalone MicroShift executable.
```
echo 1 | sudo ~/microshift/scripts/microshift-cleanup-data.sh --all --keep-images
```

Run the MicroShift executable file in the background using the following command.
```bash
nohup sudo ~/microshift/_output/bin/microshift run >> ~/microshift.log &
```

Examine the `~/microshift.log` log file to ensure a successful startup.

> An alternative way of running MicroShift is to update `/usr/bin/microshift` file
> and restart the service. The logs would then be accessible by running the
> `journalctl -xu microshift` command.
> ```bash
> sudo cp -f ~/microshift/_output/bin/microshift /usr/bin/microshift
> sudo systemctl restart microshift
> ```

Copy `kubeconfig` to the default location that can be accessed without the
administrator privilege.
```bash
mkdir -p ~/.kube/
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
```

Verify that the MicroShift is up and running.
```bash
oc get cs
watch oc get pods -A
```

### Stopping MicroShift
Run the following commands to stop the MicroShift process and make sure it is
shut down by examining its log file.
```bash
sudo kill microshift && sleep 3
tail -3 ~/microshift.log
```

> If MicroShift is running as a service, it is necessary to execute the
> `sudo systemctl stop microshift` command to shut it down and review the output
> of the `journalctl -xu microshift` command to verify the service termination.

This command only stops the MicroShift process. To perform the full cleanup including
`CRI-O`, MicroShift and OVN caches, run the following script.
```bash
echo 1 | sudo ~/microshift/scripts/microshift-cleanup-data.sh --all
```

> The full cleanup does not remove OVS configuration applied by the MicroShift service
> initialization sequence. Run the `sudo /usr/bin/configure-ovs.sh OpenShiftSDN` command
> to revert to the original network settings.

## Quick Development and Edge Testing Cycle
During the development cycle, it is practical to build and run MicroShift executable
as demonstrated in the [Building MicroShift](#building-microshift-executable) and
[Running MicroShift](#running-microshift) sections above. However, it is also necessary
to have a convenient technique for testing the system in a setup resembling the production
environment. Such an environment can be created in a virtual machine as described in the
[Install MicroShift on RHEL for Edge](./rhel4edge_iso.md) document.

Once a RHEL for Edge virtual machine is created, it is running a version of MicroShift
with the latest changes. When MicroShift code is updated and the executable file is
rebuilt with the new changes, the updates need to be installed on RHEL for Edge OS.

Since it takes a long time to create a new RHEL for Edge installer ISO and deploy it
on a virtual machine, the remainder of this section describes a simple technique for
replacing the MicroShift executable file on an existing RHEL for Edge OS installation.

### Configuring ostree
Log into the RHEL for Edge machine using `redhat:redhat` credentials. Run the following
command for configuring the ostree to allow transient overlays on top of the /usr directory.
```bash
sudo rpm-ostree usroverlay
```

This would enable a development mode where users can overwrite `/usr` directory contents.
Note that all changes will be discarded on reboot.

### Updating MicroShift Executable
Log into the development virtual machine with the `microshift:microshift` user credentials.

It is recommended to update the local `/etc/hosts` to resolve the `microshift-edge`
host name or use its IP address as presented below. Also, copy the local SSH keys
to allow the `microshift` user to run SSH commands without a password on the RHEL
for Edge machine.
```bash
VMIPADDR=192.168.122.101
ssh-copy-id redhat@${VMIPADDR}
```

Rebuild the MicroShift executable as described in the [Building MicroShift Executable](#building-microshift-executable)
section and run the following commands to copy, cleanup, replace and restart the
new service on the RHEL for Edge system.
```bash
scp ~/microshift/_output/bin/{microshift,microshift-etcd} redhat@${VMIPADDR}:
ssh redhat@${VMIPADDR} ' \
    echo 1 | sudo /usr/bin/microshift-cleanup-data --all --keep-images && \
    sudo cp -f ~redhat/{microshift,microshift-etcd} /usr/bin/ && \
    sudo systemctl enable microshift --now && \
    echo Done '
```

## Profile MicroShift
Golang [pprof](https://pkg.go.dev/net/http/pprof) is a useful tool for serving
runtime profiling data via an HTTP server in the format expected by the `pprof`
visualization tool.

Runtime profiling data can be accessed from the command line as described in the
pprof documentation. As an example, the following command can be used to look at
the heap profile.

```
oc get --raw /debug/pprof/heap > heap.pprof
go tool pprof heap.pprof
```

To view all the available profiles, run `oc get --raw /debug/pprof`.

## Troubleshooting

### No Valid Repository ID for OpenShift RPM Channels
The following error message may be encountered when enabling the OpenShift RPM repositories.
```
Error: 'fast-datapath-for-rhel-9-x86_64-rpms' does not match a valid repository ID.
Use "subscription-manager repos --list" to see valid repositories.
```

To mitigate this problem, make sure that your system is registered and attached
to the `Red Hat Openshift Container Platform` or equivalent subscription. Once the
proper subscription is configured, run the `subscription-manager` command to verify
the enabled repositories.
```
$ sudo subscription-manager repos --list-enabled
+----------------------------------------------------------+
    Available Repositories in /etc/yum.repos.d/redhat.repo
+----------------------------------------------------------+
Repo ID:   rhel-9-for-x86_64-baseos-rpms
Repo Name: Red Hat Enterprise Linux 9 for x86_64 - BaseOS (RPMs)
Repo URL:  https://cdn.redhat.com/content/dist/rhel9/$releasever/x86_64/baseos/os
Enabled:   1

Repo ID:   fast-datapath-for-rhel-9-x86_64-rpms
Repo Name: Fast Datapath for RHEL 9 x86_64 (RPMs)
Repo URL:  https://cdn.redhat.com/content/dist/layered/rhel9/x86_64/fast-datapath/os
Enabled:   1

Repo ID:   rhel-9-for-x86_64-appstream-rpms
Repo Name: Red Hat Enterprise Linux 9 for x86_64 - AppStream (RPMs)
Repo URL:  https://cdn.redhat.com/content/dist/rhel9/$releasever/x86_64/appstream/os
Enabled:   1
```

### Increase Virtual Machine Disk Size

Log into the hypervisor host, and run the following commands to resize its disk.
```
VM_NAME=microshift-dev
VM_DISK=/var/lib/libvirt/images/${VM_NAME}.qcow2
INCREASE_BY=20

sudo virsh shutdown ${VM_NAME}
# Wait until the host is shut off

sudo qemu-img resize ${VM_DISK} +${INCREASE_BY}G
sudo virsh start ${VM_NAME}
```

Log into the virtual machine and run the following commands to extend its
root partition.
```
# Use 'lsblk' command output to see your device and partition to be resized
DEVICE=/dev/vda
PARTNUM=3

# Resize the device
sudo parted ${DEVICE} ---pretend-input-tty resizepart ${PARTNUM} 100%
sudo pvresize ${DEVICE}${PARTNUM}
sudo lvextend -l +95%FREE /dev/mapper/rhel-root

# Resize the file system
sudo xfs_growfs /dev/mapper/rhel-root
```
