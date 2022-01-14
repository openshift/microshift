# MicroShift

MicroShift is a research project that is exploring how OpenShift<sup>1</sup> Kubernetes can be optimized for small form factor and edge computing.

Edge devices deployed out in the field pose very different operational, environmental, and business challenges from those of cloud computing. These motivate different engineering trade-offs for Kubernetes at the far edge than for cloud or near-edge scenarios. MicroShift's design goals cater to this:

- make frugal use of system resources (CPU, memory, network, storage, etc.),
- tolerate severe networking constraints,
- update (resp. rollback) securely, safely, speedily, and seamlessly (without disrupting workloads), and
- build on and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge, while
- providing a consistent development and management experience with standard OpenShift.

We believe these properties should also make MicroShift a great tool for other use cases such as Kubernetes applications development on resource-constrained systems, scale testing, and provisioning of lightweight Kubernetes control planes.

Watch this [end-to-end MicroShift provisioning demo video](https://youtu.be/QOiB8NExtA4) to get a first impression of MicroShift deployed onto a [RHEL for edge computing](https://www.redhat.com/en/technologies/linux-platforms/enterprise-linux/edge-computing) device and managed through [Open Cluster Management](https://github.com/open-cluster-management-io).

**Note: MicroShift is still early days and moving fast. Features are missing. Things break. But you can still help shape it, too.**

<sup>1) more precisely [OKD](https://www.okd.io/), the Kubernetes distribution by the OpenShift community</sup>

## Using MicroShift

To get started with MicroShift, please refer to the [Getting Started](https://microshift.io/docs/getting-started/) section of the MicroShift [User Documentation](https://microshift.io/docs/user-documentation/).

## Developing MicroShift

To build MicroShift from source and contribute to its development, please refer to the MicroShift [Developer Documentation](https://microshift.io/docs/developer-documentation/).

## Community

For information about getting in touch with the MicroShift community, please check our [community page](https://microshift.io/docs/community/).