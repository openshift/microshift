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
  * MicroShift supports deployments with 1 or 3 control plane and 0..N worker instances.
  * MicroShift can be deployed containerized on Podman or Docker or non-containerized via RPM and managed via systemd; it is compatible with `rpm-ostree`-based systems.
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
  * MicroShift runs all workloads that OpenShift runs, except those which depend on OpenShift's cluster operators.
  * MicroShift clusters can be managed like OpenShift clusters through [Open Cluster Management](https://github.com/open-cluster-management), except where functions depend on OpenShift's cluster operators.


## Design Principles
When deciding between different design options, we follow the following principles:

* **Minimal core**: We keep MicroShift to a minimal set of functionality, but provide mechanisms for extension.
  * Discriminator: If a functionality can be added post-cluster-up with reasonable effort, then it should not be part of the MicroShift core/binary.
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
* MicroShift is an application deployed onto a running OS, preferably as container on `podman`, managed through `systemd`. As such, it cannot assume any responsiblity or control over the device or OS it runs on, including OS software or configuration updates or typical device management tasks such as configuring host CA certs or host telemetry. 
* MicroShift runs as a single binary embedding as goroutines only those services strictly necessary to bring up a *minimal Kubernetes/OpenShift control and data plane*. Motivation:
  * Maximizes reproducibility; cluster will come up fully or not at all.
  * Does not require external orchestration, for example through operators, and allows for very fast start-up/update times.
  * Makes it simple to grok as workload for a a Linux admin persona, works well / easier to implement with systemd.
  * Smaller resource footprint has _not_ been a motivation, it may be a welcome side-effect.
* MicroShift provides a small, optional set of infrastructure services to support common use cases and reuses OpenShift's container images for these:
  * openshift-dns, openshift-router, service-ca, local storage provider
* MicroShift instances (processes) run directly on the host or containerized on Podman. They can take on the roles of Control Plane, Node, or both:
  * Instances with Control Plane role run etcd and the Kubernetes and OpenShift control plane services. As these services don't require a kubelet, pure Control Plane instances are not nodes in the Kubernetes sense and require fewer system privileges.
  * Instances with Node role run a kubelet (and thus register as node) and kube-proxy and interface with CRI-O for running workloads. They may thus require higher system privileges.
* While it's possible to run a single MicroShift instance with both Control Plane and Node roles, there may be reasons to run two instances - one Control Plane and one Node - on the same host, e.g. to run the Control Plane with fewer privileges for security reasons. Implementation decisions should consider this.
* MicroShift does not bundle any OS user space! Bundling makes maintenance and security hard, breaks compliance. Instead, user space is provided by the host OS, the container image base layer or a sidecar container.

### CLI
* The `microshift` binary runs the Control Plane / Node process, it is not a tool to manage or be clients to those processes (like `oc` or `kubeadmin`). This is reflected in the sub-commands and paraemters offered by it, e.g. using the `run` verb (which implies run-to-cancel/run-to-completion) instead of `start`/`stop` verb-pairs (which imply asynch commands that return immediately).
* For consistency and to play nicely with systemd, we avoid command line parameters that would need to be different between invokations (e.g. first-run vs subsequent runs) or instantiations (e.g. 1st Control Plane instances vs. 2nd or 3rd Control Plane instance).

### Configuration
* MicroShift uses a strictly declarative style of configuration.
* MicroShift uses as few configuration options as possible. Where it provides configuration options, they are intuitive and have sensible defauls respectively are auto-configured.
* MicroShift is preferably configured through config files, but allows overriding of parameters via environment variables (for use in containers, systemd) and command line flags (for ad-hoc use).
* MicroShift can use both user-local and system-wide configuration.

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
* No Multus.
* Open issues / questions:
  * Lightweight CNI?
  * API Load Balancing?
  * Service Load Balancing?
  * Ingress?

### Storage
* MicroShift defaults to local ephemeral storage (enough for basic use cases).
* Open issues / questions:
  * Provide escape hatch to add own CSI (which?).

### Production / Supply Chain / Release Management
* MicroShift vendors OCP source code without modification. Where it deploys container images for additional services, it deploys OCP's published container images, not the OpenShift downstream's.
* MicroShift's versioning scheme follows OCP's. This scheme signals the base OpenShift version (4.x) and order/age of builds, but intentionally avoids signaling patch level, backward compatibility (as SemVer, for example), or stability.
* We ensure the tip of our development branch is deployable and while MicroShift is still early days and experimental we expect developers (and users who want the "latest") to build & deploy from source.
* Releases are mainly provided for convenience to users that just want to give MicroShift a quick try without friction. They are cut irregularly, e.g. to make a new feature available.
* When rebasing onto a new OCP version, we vendor that version's packages and update the container image digests of the infrastructure services MicroShift deploys, i.e. the "release metadata" is baked into the MicroShift binary.
* Eventually, we expect there to be a "MicroShift Release Image" that is based on / derived from the OpenShift Release Image: It references the MicroShift container image plus the subset of container images shared with and published by OpenShift. Defining a release image should allow to reuse the proven OpenShift CI and release tooling later.
