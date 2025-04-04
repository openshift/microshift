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

## System Requirements
To run MicroShift, the minimum system requirements are:

- x86_64 or aarch64 CPU architecture
- Red Hat Enterprise Linux 9 with Extended Update Support (9.2 or later)
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
[user](./docs/user) folder.

## Contributor Documentation
To build MicroShift from source and contribute to its development, refer to the
documentation in [contributor](./docs/contributor) folder.

## Community
Community documentation sources are managed at <https://github.com/redhat-et/microshift-documentation> and published on <https://microshift.io>.

To get started with MicroShift, please refer to the [Getting Started](https://microshift.io/docs/getting-started/) section of the MicroShift [User Documentation](https://microshift.io/docs/user-documentation/).

For information about getting in touch with the MicroShift community, check our [community page](https://microshift.io/docs/community/).
