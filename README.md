# MicroShift

MicroShift is a project that is exploring how OpenShift Kubernetes can be optimized for small form factor and edge computing.

Edge devices deployed out in the field pose very different operational, environmental, and business challenges from those of cloud computing. These motivate different engineering trade-offs for Kubernetes at the far edge than for cloud or near-edge scenarios. MicroShift's design goals cater to this:

- make frugal use of system resources (CPU, memory, network, storage, etc.),
- tolerate severe networking constraints,
- update (resp. rollback) securely, safely, speedily, and seamlessly (without disrupting workloads), and
- build on and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge, while
- providing a consistent development and management experience with standard OpenShift.

We believe these properties should also make MicroShift a great tool for other use cases such as Kubernetes applications development on resource-constrained systems, scale testing, and provisioning of lightweight Kubernetes control planes.

## System Requirements
To run MicroShift, you need a machine with at least:

- x86_64/AMD64 or ARM64 CPU architecture
- Red Hat Enterprise Linux 8 with Extended Update Support (8.6 or later)
- 2 CPU cores
- 2GB of RAM
- 1GB of free storage space for MicroShift

> The system requirements also include resources for the operating system unless explicitly mentioned otherwise.

## Deploying MicroShift on Edge Devices
For production deployments, MicroShift can be run on bare metal hardware or hypervisors supported and certified for the Red Hat Enterprise Linux 8 operating system.

- [Edge systems certified for Red Hat Enterprise Linux](https://catalog.redhat.com/hardware/search?c_catalog_channel=Edge%20System&p=1)
- [Hypervisors certified for Red Hat Enterprise Linux](https://access.redhat.com/solutions/certified-hypervisors)

## User Documentation
To install, configure and run MicroShift, refer to the following documentation:

- [Getting Started with MicroShift](./docs/getting_started.md)
- [MicroShift Configuration](./docs/howto_config.md)
- [MicroShift Behind Proxy](./docs/howto_http_proxy.md)
- [Load Balancer for User Workloads](./docs/howto_load_balancer.md)
- [AMQ Broker on MicroShift](./docs/howto_amq_broker.md)
- [MicroShift Mitigation of System Configuration Changes](./docs/howto_sysconf_watch.md)
- [Firewall Configuration](./docs/howto_firewall.md)
- [Debugging Tips](./docs/debugging_tips.md)

## Developer Documentation
To build MicroShift from source and contribute to its development, refer to the following documentation:

- [MicroShift Design](./docs/design.md)
- [Enabled OpenShift APIs](./docs/enabled_apis.md)
- [MicroShift Development Environment on RHEL 8](./docs/devenv_rhel8.md)
- [Install MicroShift on RHEL for Edge](./docs/rhel4edge_iso.md)
- [OpenShift CI for MicroShift](./docs/openshift_ci.md)
- [RPM Packages for Development and Testing](./docs/rpm_packages.md)
- [MicroShift Storage Plugin Overview](./docs/default_csi_plugin.md)
- [MicroShift Network Plugin Overview](./docs/default_cni_plugin.md)

## Community
Community documentation sources are managed at <https://github.com/redhat-et/microshift-documentation> and published on <https://microshift.io>.

To get started with MicroShift, please refer to the [Getting Started](https://microshift.io/docs/getting-started/) section of the MicroShift [User Documentation](https://microshift.io/docs/user-documentation/).

For information about getting in touch with the MicroShift community, check our [community page](https://microshift.io/docs/community/). 
