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
To install, configure and run MicroShift, refer to the following documentation:

- [Getting Started with MicroShift](./docs/user/getting_started.md)
- [MicroShift Configuration](./docs/user/howto_config.md)
- [MicroShift kubeconfig Handling](./docs/user/howto_kubeconfig.md)
- [Embedding MicroShift Container Images for Offline Deployments](./docs/user/howto_offline_containers.md)
- [MicroShift Behind Proxy](./docs/user/howto_http_proxy.md)
- [Load Balancer for User Workloads](./docs/user/howto_load_balancer.md)
- [AMQ Broker on MicroShift](./docs/user/howto_amq_broker.md)
- [MicroShift Mitigation of System Configuration Changes](./docs/user/howto_sysconf_watch.md)
- [Firewall Configuration](./docs/user/howto_firewall.md)
- [Integrating MicroShift with Greenboot](./docs/user/greenboot.md)
- [Mirror MicroShift Container Images](./docs/user/howto_mirror_images.md)
- [Debugging Tips](./docs/user/debugging_tips.md)
- [Known Limitations](./docs/user/known_limitations.md)

## Contributor Documentation
To build MicroShift from source and contribute to its development, refer to the following documentation:

- [MicroShift Design](./docs/contributor/design.md)
- [Enabled OpenShift APIs](./docs/contributor/enabled_apis.md)
- [MicroShift Development Environment](./docs/contributor/devenv_setup.md)
- [MicroShift Development Environment in Cloud](./docs/contributor/devenv_cloud.md)
- [Rebasing MicroShift](./docs/contributor/rebase.md)
- [Install MicroShift on RHEL for Edge](./docs/contributor/rhel4edge_iso.md)
- [OpenShift CI for MicroShift](./docs/contributor/openshift_ci.md)
- [RPM Packages for Development and Testing](./docs/contributor/rpm_packages.md)
- [MicroShift Storage Plugin Overview](./docs/contributor/storage/default_csi_plugin.md)
- [MicroShift Network Plugin Overview](./docs/contributor/network/default_cni_plugin.md)
- [MicroShift Host Networking Overview](./docs/contributor/network/host_networking.md)
- [MicroShift Traffic Flows Overview](./docs/contributor/network/ovn_kubernetes_traffic_flows.md)
- [Testing MicroShift Integration with Greenboot](./docs/contributor/greenboot.md)
- [Quay Mirror Registry Setup for Testing](./docs/contributor/howto_quay_mirror.md)
- [Multinode Testing Environment for MicroShift](./docs/contributor/multinode/setup.md)
- [Image Mode for MicroShift](./docs/contributor/image_mode.md)
- [Layered Product Testing with MicroShift](./docs/contributor/layered_product_ci.md)

## Community
Community build and CI sources managed at <https://github.com/microshift-io/microshift> and documentation sources are managed at <https://github.com/redhat-et/microshift-documentation> and published on <https://microshift.io>.

To get started with MicroShift, please refer to the [Getting Started](https://microshift.io/docs/getting-started/) section of the MicroShift [User Documentation](https://microshift.io/docs/user-documentation/).

For information about getting in touch with the MicroShift community, check our [community page](https://microshift.io/docs/community/).
