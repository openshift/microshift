# Microshift

Microshift is OpenShift<sup>1</sup> Kubernetes in a small form factor and optimized for edge computing.

Edge devices deployed out in the field pose very different operational, environmental, and business challenges from those of cloud computing. These motivate different engineering trade-offs for Kubernetes at the far edge than for cloud or near-edge scenarios. Microshift's design goals cater to this:

* make frugal use of system resources (CPU, memory, network, storage, etc.),
* tolerate severe networking constraints,
* update (resp. roll back) securely, safely, speedily, and seamlessly (without disrupting workloads), and
* build on and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge, while
* providing a consistent development and management experience with standard OpenShift.

We believe these properties should also make Microshift a great tool for other use cases such as Kubernetes applications development on resource-constrained systems, scale testing, and provisioning of lightweight Kubernetes control planes.

**Note: Microshift is still early days and moving fast. Features are missing. Things break. But you can still help shape it, too.**

<sup>1) more precisely [OKD](https://www.okd.io/), the Kubernetes distribution by the OpenShift community</sup>

## Using Microshift
To give Microshift a try, simply install a recent test version (we don't provide stable releases yet) on a Fedora-derived Linux distro (we've only tested Fedora, RHEL, and CentOS Stream so far) using:

**WARNING: At this time the script will disable SELinux.**

```
curl -sfL https://raw.githubusercontent.com/redhat-et/microshift/main/install.sh | sh -
```

This will install Microshift's dependencies (CRI-O), install it as a systemd service and start it.

For convenience, the script will also add a new "microshift" context to your `$HOME/.kube/config`, so you'll be able to access your cluster using, e.g.:
```
kubectl get all -A --context microshift
```
or
```
kubectl config use-context microshift
kubectl get all -A
```

Notes: When installing Microshift on a system with an older version already installed, it is safest to remove the old data directory and start fresh:
```
rm -rf /var/lib/microshift && rm -r $HOME/.microshift
```

## Developing Microshift

> Note: when building or running **ARM64** container images, Linux host environments must have the `qemu-user-static` package installed.  E.g. on Fedora: `dnf install qemu-user-static`.

### Building

You can locally build Microshift using one of two methods, either using a container build (recommended) on Podman or Docker:
```
sudo yum -y install make golang
make microshift
```

or directly on the host after installing the build-time dependencies. When using RHEL ensure the system is registered and run the following before installing the prerequisites.

```
ARCH=$( /bin/arch )
sudo subscription-manager repos --enable "codeready-builder-for-rhel-8-${ARCH}-rpms"
```

The following packages are required for Fedora and RHEL.
```
sudo yum install -y glibc-static gcc make golang
make
```

### Environment Configuration

Before running Microshift, the host must first be configured.  This can be handled by running  

```
CONFIG_ENV_ONLY=true ./install.sh
```

Microshift keeps all its state in its data-dir, which defaults to `/var/lib/microshift` when running Microshift as privileged user and `$HOME/.microshift` otherwise. Note that running Microshift unprivileged only works without node role at the moment (i.e. using `--roles=controlplane` instead of the default of `--roles=controlplane,node`).

### Kubeconfig
The `install.sh` script should place the kubeconfig file for you. If you need it for another user or to use externally the kubeadmin's kubeconfig is placed `/var/lib/microshift/resources/kubeadmin/kubeconfig` during configuration.

