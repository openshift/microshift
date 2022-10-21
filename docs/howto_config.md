# Configuration

MicroShift can be configured in the following ways, in order of precedence:
* Command line arguments
* Environment variables
* Configuration file

The MicroShift configuration file may be located at `~/.microshift/config.yaml` (user-specific) and `/etc/microshift/config.yaml` (system-wide), while the former takes precedence if it exists. Use the `--config` command line argument or `MICROSHIFT_CONFIGFILE` environment variable to specify a custom location of the configuration file.

The format of the `config.yaml` configuration file is as follows.

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

The configuration settings alongside with the supported command line arguments and environment variables are presented below.

| Field Name          | CLI Argument              | Environment Variable                    | Description |
|---------------------|---------------------------|-----------------------------------------|-------------|
| clusterCIDR         | --cluster-cidr            | MICROSHIFT_CLUSTER_CLUSTERCIDR          | A block of IP addresses from which Pod IP addresses are allocated
| serviceCIDR         | --service-cidr            | MICROSHIFT_CLUSTER_SERVICECIDR          | A block of virtual IP addresses for Kubernetes services
| serviceNodePortRange| --service-node-port-range | MICROSHIFT_CLUSTER_SERVICENODEPORTRANGE | The port range allowed for Kubernetes services of type NodePort
| dns                 | --cluster-dns             | MICROSHIFT_CLUSTER_DNS                  | The Kubernetes service IP address where pods query for name resolution
| domain              | --cluster-domain          | MICROSHIFT_CLUSTER_DOMAIN               | Base DNS domain used to construct fully qualified pod and service domain names
| url                 | --url                     | MICROSHIFT_CLUSTER_URL                  | URL of the API server for the cluster.
| mtu                 | --cluster-mtu             | MICROSHIFT_CLUSTER_MTU                  | The maximum transmission unit for the Generic Network Virtualization Encapsulation overlay network
| nodeIP              | --node-ip                 | MICROSHIFT_NODEIP                       | The IP address of the node, defaults to IP of the default route
| nodeName            | --node-name               | MICROSHIFT_NODENAME                     | The name of the node, defaults to hostname
| auditLogDir         | --audit-log-dir           | MICROSHIFT_AUDITLOGDIR                  | Location for storing audit logs
| dataDir             | --data-dir                | MICROSHIFT_DATADIR                      | Location for data created by MicroShift
| logVLevel           | --v                       | MICROSHIFT_LOGVLEVEL                    | Log verbosity (0-5)
| manifests           | n/a                       | n/a                                     | Locations to scan for manifests to be loaded on startup

## Default Settings

In case `config.yaml` is not provided, the following default settings will be used.

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
```

# Auto-applying Manifests

MicroShift leverages `kustomize` for Kubernetes-native templating and declarative management of resource objects. Upon start-up, it searches `/etc/microshift/manifests` and `/usr/lib/microshift/manifests` directories for a `kustomization.yaml` file. If it finds one, it automatically runs `kubectl apply -k` command to apply that manifest.

The reason for providing multiple directories is to allow a flexible method to manage MicroShift workloads.

| Location                      | Intent |
|-------------------------------|--------|
| /etc/microshift/manifests     | Read-write location for configuration management systems or development
| /usr/lib/microshift/manifests | Read-only location for embedding configuration manifests on ostree based systems

The list of manifest locations can be customized via configuration using the above-mentioned `manifests` section of the `config.yaml` file or via the `MICROSHIFT_MANIFESTS` environment variable as comma separated directories.

## Manifest Example

The example demonstrates automatic deployment of a `busybox` container using `kustomize` manifests in the `/etc/microshift/manifests` directory.

Run the following command to create the manifest files.

```bash
MANIFEST_DIR=/etc/microshift/manifests
sudo mkdir -p ${MANIFEST_DIR}

sudo tee ${MANIFEST_DIR}/busybox.yaml &>/dev/null <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-deployment
spec:
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: busybox
        image: BUSYBOX_IMAGE
        command:
          - sleep
          - "3600"
EOF

sudo tee ${MANIFEST_DIR}/kustomization.yaml &>/dev/null <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: busybox
resources:
  - busybox.yaml
images:
  - name: BUSYBOX_IMAGE
    newName: registry.k8s.io/busybox
EOF
```

Restart the MicroShift service to apply the manifests and verify that the `busybox` pod is running.

```bash
sudo systemctl restart microshift
oc get pods -n busybox
```
