# MicroShift

MicroShift is a project that is exploring how OpenShift Kubernetes can be optimized for small form factor and edge computing.

Edge devices deployed out in the field pose very different operational, environmental, and business challenges from those of cloud computing. These motivate different engineering trade-offs for Kubernetes at the far edge than for cloud or near-edge scenarios. MicroShift's design goals cater to this:

- make frugal use of system resources (CPU, memory, network, storage, etc.),
- tolerate severe networking constraints,
- update (resp. rollback) securely, safely, speedily, and seamlessly (without disrupting workloads), and
- build on and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge, while
- providing a consistent development and management experience with standard OpenShift.

We believe these properties should also make MicroShift a great tool for other use cases such as Kubernetes applications development on resource-constrained systems, scale testing, and provisioning of lightweight Kubernetes control planes.

## Developing MicroShift

To build MicroShift from source and contribute to its development, refer to the MicroShift [Developer Documentation](./docs/devenv_rhel8.md).

The following productization and development topics are documented in the current repository:
- [MicroShift Design](./docs/design.md)
- [MicroShift Development Environment on RHEL 8.x](./docs/devenv_rhel8.md)

## Community

Community documentation sources are managed at <https://github.com/redhat-et/microshift-documentation> and published on <https://microshift.io>.

To get started with MicroShift, please refer to the [Getting Started](https://microshift.io/docs/getting-started/) section of the MicroShift [User Documentation](https://microshift.io/docs/user-documentation/).

For information about getting in touch with the MicroShift community, check our [community page](https://microshift.io/docs/community/). 
