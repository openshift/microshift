# Configuration

MicroShift components are configured by modifying `config.yaml` instead of using Kubernetes APIs.
MicroShift binary looks for both user-local `~/.microshift/config.yaml` and system-wide `/etc/microshift/config.yaml` confguration files in order, user-local file takes precedence if exists.

The following fields are supported in `config.yaml`:

```yaml
cluster:
  clusterCIDR: ""
  serviceCIDR: ""
  serviceNodePortRange: ""
  dns: ""
  domain: ""
  url: ""
  mtu: ""
nodeIP: ""
nodeName: ""
auditLogDir: ""
dataDir: ""
logVLevel: ""
manifests: []
```

* clusterCIDR: A block of IP addresses from which Pod IP addresses are allocated.
* serviceCIDR: A block of virtual IP addresses for Kubernetes services.
* serviceNodePortRange: The port range allowed for Kubernetes services of type NodePort.
* dns: The Kubernetes service IP address where pods query for name resolution.
* domain: Base DNS domain used to construct fully qualified pod and service domain names.
* url: URL of the API server for the cluster.
* mtu: The maximum transmission unit for the Geneve (Generic Network Virtualization Encapulation) overlay network.
* nodeIP: The IP address of the node, defaults to IP of the default route.
* nodeName: The name of the node, defaults to hostname.
* auditLogDir: Location for storing audit logs.
* dataDir: Location for data created by MicroShift.
* logVLevel: Log verbosity (0-5).
* manifests: Locations to scan for manifests to load on startup.

In case `config.yaml` is not provided, the following default settings will be used:

```yaml
cluster:
  clusterCIDR: 10.42.0.0/16
  serviceCIDR: 10.43.0.0/16
  serviceNodePortRange: 30000-32767
  dns: 10.43.0.10
  domain: cluster.local
  url: https://127.0.0.1:6443
  mtu: "1400"
nodeIP: ""
nodeName: ""
auditLogDir: ""
dataDir: /var/lib/microshift
logVLevel: 0
manifests:
- /usr/lib/microshift/manifests
- /etc/microshift/manifests
- /var/lib/microshift/manifests
```
