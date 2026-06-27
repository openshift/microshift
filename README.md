# MicroShift

MicroShift is a project that optimizes OpenShift Kubernetes for small
form factor and edge computing.

Edge devices deployed out in the field pose very different operational,
environmental, and business challenges from those of cloud computing.
These motivate different engineering trade-offs for Kubernetes at the
far edge than for cloud or near-edge scenarios.

MicroShift design goals cater to this:
- make frugal use of system resources (CPU, memory, network, storage, etc.)
- tolerate severe networking constraints
- update securely, safely, speedily, and seamlessly (without disrupting workloads)
- build on and integrate cleanly with edge-optimized operating systems like RHEL for Edge
- provide a consistent development and management experience with standard OpenShift

These properties should also make MicroShift a great tool for other use cases
such as Kubernetes applications development on resource-constrained systems,
scale testing, and provisioning of lightweight Kubernetes control planes.

## Architecture

MicroShift runs as a single systemd service that embeds a minimal OpenShift
control plane and worker node. A single binary packages kube-apiserver,
kube-controller-manager, kube-scheduler, kubelet, and OpenShift controllers
as goroutines. Etcd runs as a separate process managed by the main binary.

Infrastructure services (DNS, ingress router, service CA, OVN-Kubernetes CNI,
LVMS storage) are deployed onto the cluster after the control plane starts.
Optional components (OLM, Gateway API, cert-manager, SR-IOV) are delivered
as separate RPM packages and auto-discovered via the manifest path system.

MicroShift vendors OpenShift source code without modification. A given
MicroShift release is built from the same content as the corresponding
OpenShift release, following the same versioning scheme.

For technical details, see [architecture](./docs/contributor/architecture.md).
For design goals and principles, see [design](./docs/contributor/design.md).

## System Requirements

To run MicroShift, the minimum system requirements are:

- x86_64 or aarch64 CPU architecture
- Red Hat Enterprise Linux with Extended Update Support
- 2 CPU cores
- 2GB of RAM
- 2GB of free system root storage for MicroShift and its container images

> The system requirements also include resources for the operating system unless explicitly mentioned otherwise.

Depending on user workload requirements, it may be necessary to add more resources i.e. CPU and RAM for better
performance, disk space in a root partition for container images, an LVM group for container storage, etc.

## Deploying MicroShift on Edge Devices

For production deployments, MicroShift can be run on bare metal hardware or hypervisors supported and certified for the Red Hat Enterprise Linux 9 operating system.

- [Edge systems certified for Red Hat Enterprise Linux](https://catalog.redhat.com/hardware/search?c_catalog_channel=Edge%20System&p=1)
- [Hypervisors certified for Red Hat Enterprise Linux](https://access.redhat.com/solutions/certified-hypervisors)

## User Documentation

To install, configure and run MicroShift, refer to the documentation in the
[user](./docs/user/README.md) folder.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to build, test, and submit
changes to MicroShift.

For detailed contributor documentation including development environment setup,
CI, and rebase procedures, see the [contributor](./docs/contributor/README.md)
folder.

## Community

Community build and CI sources managed at <https://github.com/microshift-io/microshift> and documentation sources are managed at <https://github.com/redhat-et/microshift-documentation> and published on <https://microshift.io>.

To get started with MicroShift, please refer to the [Getting Started](https://microshift.io/docs/getting-started/) section of the MicroShift [User Documentation](https://microshift.io/docs/user-documentation/).

For information about getting in touch with the MicroShift community, check our [community page](https://microshift.io/docs/community/).
