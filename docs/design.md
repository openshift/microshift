# Design Documentation
The is a documentation of MicroShift's design goals, design principles, and fundamental design decisions. For details on specific feature enhancements, please refer to the corresponding low-level design documents.


## Design Goals
MicroShift aims at meeting all of the following design goals:

* **Optimized for field-deployment:**
  * Provisioning and replacing devices running MicroShift is "plug&play"<sup>1</sup>; MicroShift does not add friction to this.
    * e.g. auto-configuring, auto-clustering
  * MicroShift works seamlessly under adverse network conditions.
    * e.g. disconnected or rarely connected, NAT'ed or firewalled, changing IP addresses, IPv4 or v6, high latency / low bandwidth, no control over local network (DNS, DHCP, LBN, GW), connectivity via LTE dongle (i.e. no LAN)
  * MicroShift operates autonomously; it does not require external orchestration.
  * MicroShift is safe to change<sup>1</sup>; it has means to automatically recover from faulty software or configuration updates that would render it unmanageable or non-operational.
  * MicroShift is secure<sup>1</sup> even in environments without physical accesss security.

  <sup>1) when used in combination with an edge-optimized OS like RHEL 4 Edge or Fedora IoT</sup>

* **Production-grade:**
  * MicroShift supports deployments with 1 node acting as control plane and worker. Orchestration of multi-node control plane configurations introduces unwanted complexity into deployment and upgrade processes. A single-node control plane becomes a single point of failure when worker nodes are attached, which eliminates the high availability benefits of running workloads on kubernetes. For application availability, we recommend running two single-node instances that deploy a common application in active/active or active/passive mode and then using existing tools to support failover between those states when either host is unable to provide availability.
  * MicroShift can be deployed via RPM and managed via systemd. It is compatible with `rpm-ostree`-based systems.
  * MicroShift's lifecyle is decoupled from the underlying OS's lifecycle.
  * MicroShift can be deployed such that updates or changes to it do not disrupt running workloads.
  * MicroShift meets DISA STIG and FedRAMP security requirements; it runs as non-privileged workload and supports common CVE and auditing workflows.
  * MicroShift allows segregation between the "edge device administrator" and the "edge service development and operations" personas.
    * the former being responsible for device+OS lifecycle and installing MicroShift as a workload, the latter being responsible for services on and resources of the MicroShift cluster
  * MicroShift provides application-level events and metrics for observability.

* **Usability:**
  * MicroShift does not require Kubernetes-expertise to configure, deploy, or lifecycle-manage it; Linux admins will find it behaves like any other Linux workload.
  * MicroShift does not require Kubernetes-experts to learn new techniques; Kubernetes developers and operations teams will find it supports their tools and processes.
  * MicroShift is highly opinionated, working with minimal/no configuration out of the box, but offers escape hatches for advanced users.

* **Small resource footprint:**
  * MicroShift makes frugal use of system resources. It runs on <1GB RAM and <1 CPU core (Intel Atom- or ARM Cortex-class). It consumes <500MB on the wire (per install/update) and <1GB at rest (excl. etcd state).

* **Consistency with OpenShift:**
  * MicroShift is binary compatible with OpenShift and Kubernetes conforming.
  * MicroShift runs all OpenShift workloads without modification, except for if these rely on OpenShift APIs that are
    * provided by OpenShift's Operators for cluster infrastructure or lifecycle management (not applicable in MicroShift and thus removed) or
    * not relevant to a pure non-interactive, runtime-only cluster (as opposed to a build-cluster with multiple interactive users).
  * MicroShift workloads can be managed like OpenShift workloads through [Red Hat Advanced Cluster Management for Kubernetes](https://www.redhat.com/en/technologies/management/advanced-cluster-management), except where functions depend on unsupported OpenShift APIs (see above).


## Design Principles
When deciding between different design options, we follow the following principles:

* **Minimal core**: We keep MicroShift to a minimal set of functionality, but provide mechanisms for extension.
  * Discriminator 1: If the functionality implements changes to the OS or its configuration, it should probably be done as pre-requisite for running MicroShift binary.
  * Discriminator 2: If the functionality is essential to MicroShift's atomic start/stop/update behavior (i.e. MicroShift running/stopped means the OpenShift control plane is running/stopped) it must be part of the binary, otherwise it should be hosted on the cluster.
  * Discriminator 3: If can be installed "post-cluster-up" and isn't used by 90% of MicroShift users, it should probably not be pre-integrated with MicroShift at all.
* **Least-privileged**: We minimize privileges needed by MicroShift and its workloads.
* **Offline-first**: We prefer mechanisms that seamlessly work with poor or no network connectivity.
* **Minimal configuration**: We minimize the number of configuration parameters exposed to users. Where parameters cannot be avoided, we provide robust defaults or try to auto-configure them.
  * Discriminator: If a parameter can be infered from another parameter, auto-detected, or only covers rare use cases, then likely it should not be exposed to users.
* **Robustness to failure modes**: We expect and gracefully handle failure modes stemming from field-deployment and that MicroShift is just an app on somebody else's OS it cannot control.
* **Production over Development**: We engineer for production-deployments, not Kubernetes application development environments.
* **Frugal use of resources**: We are mindful of MicroShift's resources consumption, considering all resources (memory, CPU, storage I/O, network I/O).
* **No premature resouce optimization**: We do not attempt to squeeze out the last 20% of resouces, e.g. by patching or compressing code, inventing lighter-weight components, etc.
* **Ease-of-Use**: We make MicroShift intutive and simple to use, even if that means providing fewer features and options.
* **Prefer well-established mechanisms**: We meet users where they are by letting them use well-established tools and patterns rather than requiring them to learn new ones.
* **Cheap Control Plane Restarts**: We keep MicroShift restarts cheap (= fast, low resource consumption, low workload impact, etc.) as it is our default model for software and configuration changes.
* **Alignment with OpenShift**: We reuse OpenShift's code, operational logic, and production chain tools and processes where possible with reasonable effort, unless this would conflict with MicroShift's goals.


## Design Decisions
### Overall Architecture
* MicroShift is an application deployed onto a running OS. As such, it cannot assume any responsiblity or control over the device or OS it runs on, including OS software or configuration updates or typical device management tasks such as configuring host CA certs or host telemetry.
* MicroShift runs as a single binary embedding as goroutines only those services strictly necessary to bring up a *minimal Kubernetes/OpenShift control and data plane*. Motivation:
  * Maximizes reproducibility; cluster will come up fully or not at all.
  * Does not require external orchestration, for example through operators, and allows for very fast start-up/update times.
  * Makes it simple to grok as workload for a a Linux admin persona, works well / easier to implement with systemd.
  * Reduces resource footprint by downloading and running "less stuff".
* MicroShift provides a small, optional set of infrastructure services to support common use cases and reuses OpenShift's container images for these:
  * openshift-dns, openshift-router, service-ca, local storage provider
* MicroShift does not bundle any OS user space! Bundling makes maintenance and security hard, breaks compliance. Instead, user space is provided by the host OS, the container image base layer or a sidecar container.

### Supported APIs
* See [enabled apis](./enable_apis.md) for the subset of OpenShift APIs enabled in MicroShift.

### CLI
* The `microshift` binary runs the Control Plane / Node process, it is not a tool to manage or be clients to those processes (like `oc` or `kubeadmin`). This is reflected in the sub-commands and paraemters offered by it, e.g. using the `run` verb (which implies run-to-cancel/run-to-completion) instead of `start`/`stop` verb-pairs (which imply asynch commands that return immediately).
* For consistency and to play nicely with systemd, we avoid command line parameters that would need to be different between invocations (e.g. first-run vs subsequent runs).

### Configuration
* MicroShift uses a strictly declarative style of configuration.
* MicroShift uses as few configuration options as possible. Where it provides configuration options, they are intuitive and have sensible defaults respectively are auto-configured.
* MicroShift is preferably configured through config files, but allows overriding of parameters via environment variables (for use in containers, systemd) and command line flags (for ad-hoc use).
* MicroShift can use both user-local and system-wide configuration.
* Components built into the `microshift` binary are configured by modifying `config.yaml` and restarting the service, instead of through Kubernetes APIs as in standalone OpenShift. This ensures that:
  * MicroShift can act as an "application running on RHEL", which avoids requiring Kubernetes expertise to deploy it.
  * Configuration changes can be packaged as part of an ostree update with the OS.
  * The configurable options can be limited and abstracted for the MicroShift use case.
  * The complexity of applying configuration changes is minimized by not attempting to reconfigure a running component on the fly.
* Components bundled with MicroShift but not built into the `microshift` binary (CSI and CNI drivers, for example) may be configured with configuration files native to the bundled application (`lvmd` or OVN-K). This approach means
    * We can separate the OS-level configuration ("bottom half") of these components from the cluster-level configuration ("top half").
    * The components may be treated as optional or replaceable for deployments with different needs.
    * We avoid importing the configuration API of those bundled components into MicroShift.
* MicroShift does not have knowledge of configuring add-on components such as operators, alternative CSI implementations, etc.
* Applications running on MicroShift are configured with Kubernetes resources, which can be loaded automatically from manifests or managed through the API.

### Security
* MicroShift instances should run with least privileges. In particular Control Plane-only instances should run completely non-privileged.
* MicroShift should minimize the number of open network ports. In particular Node-only instances should not open any listening ports.
* Assume there is no way to access the device or MicroShift instance from remote, i.e. no SSH access nor Kubernetes API access (for kubectl). Instead, communication is *always* initiated from the device / cluster towards the management system.
* Open issues / questions:
  * Model for joining nodes to MicroShift clusters.
  * Model for certificate rotation.
  * Requirements for FedRAMP and DISA STIG compliance.
  * How to leverage / support the host OS's features for FIDO Device Onboard, remote attestation, integrity measurement architecture.
  * How to support secure supply chain.

### Networking
* Host networking is configured by device management. MicroShift has to work with what it's been given by the host OS.
* MicroShift uses the Red Hat OpenShift Networking CNI driver, based on OVN-Kubernetes.
* No Multus.
* Single-node, so no API load balancer.
* Open issues / questions:
  * Ingress?

### Storage
* MicroShift uses ODF-LVM for local storage volume allocation.

### Production / Supply Chain / Release Management
* MicroShift vendors OCP source code without modification. Where it deploys container images for additional services, it deploys OCP's published container images.
* MicroShift's versioning scheme follows OCP's.
* We ensure the tip of our development branch is deployable.
