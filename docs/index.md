---
modified: "2021-10-26T13:12:07.131+02:00"
title: The  Project
tags: microshift project, edge, µShift
layout: page
toc: true
---

MicroShift is a research project that is exploring how OpenShift<sup>1</sup> Kubernetes can be optimized for small form factor and edge computing.

Edge devices deployed out in the field pose very different operational, environmental, and business challenges from those of cloud computing. These motivate different engineering trade-offs for Kubernetes at the far edge than for cloud or near-edge scenarios. MicroShift's design goals cater to this:

- make frugal use of system resources (CPU, memory, network, storage, etc.),
- tolerate severe networking constraints,
- update (resp. roll back) securely, safely, speedily, and seamlessly (without disrupting workloads), and
- build on and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge, while
- providing a consistent development and management experience with standard OpenShift.

We believe these properties should also make MicroShift a great tool for other use cases such as Kubernetes applications development on resource-constrained systems, scale testing, and provisioning of lightweight Kubernetes control planes.

Watch this [end-to-end MicroShift provisioning demo video](https://youtu.be/QOiB8NExtA4) to get a first impression of MicroShift deployed onto a [RHEL for edge computing](https://www.redhat.com/en/technologies/linux-platforms/enterprise-linux/edge-computing) device and managed through [Open Cluster Management](https://github.com/open-cluster-management).

> warning ""
> MicroShift is still early days and moving fast. Features are missing. Things break. But you can still help shape it, too.\*\*


<sup>1) more precisely [OKD](https://www.okd.io/), the Kubernetes distribution by the OpenShift community</sup>

## Minimum specs

In order to run MicroShift, you will need at least:

- 2 CPU cores
- 2GB of RAM
- ~124MB of free storage space for the MicroShift binary
- 64-bit CPU (although 32-bit is _technically_ possible, if you're up for the challenge)

For barebones development the minimum requirement is 3GB of RAM, though this can increase
if you are using resource-intensive devtools.

### OS Requirements

The all-in-one containerized MicroShift can run on Windows, MacOS, and Linux.

Currently, the MicroShift binary is known to be supported on the following Operating Systems:

- Fedora 33/34
- CentOS 8 Stream
- RHEL 8
- CentOS 7
- Ubuntu 20.04

It may be possible to run MicroShift on other systems, however they haven't been tested so you may run into issues.

## Using MicroShift

To give MicroShift a try, simply install a recent test version (we don't provide stable releases yet) on a Fedora-derived Linux distro (we've only tested Fedora, RHEL, and CentOS Stream so far) using:

```sh
curl -sfL https://raw.githubusercontent.com/redhat-et/microshift/main/install.sh | bash
```

This will install MicroShift's dependencies (CRI-O), install it as a systemd service and start it.

For convenience, the script will also add a new "microshift" context to your `$HOME/.kube/config`, so you'll be able to access your cluster using, e.g.:

```sh
kubectl get all -A --context microshift
```

or

```sh
kubectl config use-context microshift
kubectl get all -A
```

> info ""
> When installing MicroShift on a system with an older version already installed, it is safest to remove the old data directory and start fresh:

```sh
rm -rf /var/lib/microshift && rm -r $HOME/.microshift
```

## [Known Issues](./known-issues.md)

## Developing MicroShift

> warning ""
> when building or running **ARM64** container images, Linux host environments must have the `qemu-user-static` package installed. E.g. on Fedora: `dnf install qemu-user-static`.

### Building

You can locally build MicroShift using one of two methods, either using a container build (recommended) on Podman or Docker:

```sh
sudo yum -y install make golang
make microshift
```

or directly on the host after installing the build-time dependencies. When using RHEL ensure the system is registered and run the following before installing the prerequisites.

```sh
ARCH=$( /bin/arch )
sudo subscription-manager repos --enable "codeready-builder-for-rhel-8-${ARCH}-rpms"
```

The following packages are required for Fedora and RHEL.

```sh
sudo yum install -y glibc-static gcc make golang
make
```

### Environment Configuration

Before running MicroShift, the host must first be configured. This can be handled by running

```sh
CONFIG_ENV_ONLY=true ./install.sh
```

MicroShift keeps all its state in its data-dir, which defaults to `/var/lib/microshift` when running MicroShift as privileged user and `$HOME/.microshift` otherwise. Note that running MicroShift unprivileged only works without node role at the moment (i.e. using `--roles=controlplane` instead of the default of `--roles=controlplane,node`).

### Kubeconfig

When starting the MicroShift for the first time the Kubeconfig file is created. If you need it for another user or to use externally the kubeadmin's kubeconfig is placed at `/var/lib/microshift/resources/kubeadmin/kubeconfig`.

### Contributing

For more information on working with MicroShift, you can find a contributor's guide in [CONTRIBUTING.md](./CONTRIBUTING.md)

### Community

Join us on [Slack](https://microshift.slack.com)!

Community meetings are held weekly, Wednesdays at 10:30AM - 11:00AM EST. Be sure to join the community [calendar](https://calendar.google.com/calendar/embed?src=nj6l882mfe4d2g9nr1h7avgrcs%40group.calendar.google.com&ctz=America%2FChicago)! Click "Google Calendar" in the lower right hand corner to subscribe.
