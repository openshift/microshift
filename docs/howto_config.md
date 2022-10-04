# Configuration

MicroShift components are configured by modifying `config.yaml` instead of using Kubernetes APIs. MicroShift binary looks for both user-local `~/.microshift/config.yaml` and system-wide `/etc/microshift/config.yaml` configuration files in order, user-local file takes precedence if exists.

The following fields are supported in `config.yaml`.

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

| Field Name          | Description |
|---------------------|-------------|
| clusterCIDR         | A block of IP addresses from which Pod IP addresses are allocated
| serviceCIDR         | A block of virtual IP addresses for Kubernetes services
| serviceNodePortRange| The port range allowed for Kubernetes services of type NodePort
| dns                 | The Kubernetes service IP address where pods query for name resolution
| domain              | Base DNS domain used to construct fully qualified pod and service domain names
| url                 | URL of the API server for the cluster.
| mtu                 | The maximum transmission unit for the Geneve (Generic Network Virtualization Encapulation) overlay network
| nodeIP              | The IP address of the node, defaults to IP of the default route
| nodeName            | The name of the node, defaults to hostname
| auditLogDir         | Location for storing audit logs
| dataDir             | Location for data created by MicroShift
| logVLevel           | Log verbosity (0-5)
| manifests           | Locations to scan for manifests to be loaded on startup

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

The example demonstrates automatic deployment of an `nginx` container using `kustomize` manifests in the `/etc/microshift/manifests` directory.

Run the following command to create the manifest files.

```bash
MANIFEST_DIR=/etc/microshift/manifests
sudo mkdir -p ${MANIFEST_DIR}

sudo tee ${MANIFEST_DIR}/nginx.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: NGINX_IMAGE
        ports:
        - containerPort: 8080
EOF

sudo tee ${MANIFEST_DIR}/kustomization.yaml <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - nginx.yaml
images:
  - name: NGINX_IMAGE
    newName: nginx:1.21
EOF
```

Restart the MicroShift service to apply the manifests and verify that the `nginx` container was created.

```bash
sudo systemctl restart microshift
oc get pods -l app=nginx
```
