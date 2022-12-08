# Configuration

MicroShift can be configured in the following ways, in order of precedence:
* Command line arguments
* Environment variables
* Configuration file

The MicroShift configuration file must be located at `~/.microshift/config.yaml` (user-specific) and `/etc/microshift/config.yaml` (system-wide), while the former takes precedence if it exists.

The format of the `config.yaml` configuration file is as follows.

```yaml
subjectAltNames:
  - ""
nodeName: ""
nodeIP: ""
url: ""
network:
  clusterNetwork:
    - cidr: ""
  serviceNetwork:
    - ""
  serviceNodePortRange: ""
dns:
  baseDomain: ""
debugging:
  logLevel: ""
```

The configuration settings alongside with the supported command line arguments and environment variables are presented below.

| Field Name            | CLI Argument              | Environment Variable                    | Description |
|-----------------------|---------------------------|-----------------------------------------|-------------|
| cidr (clusterNetwork) | --cluster-cidr            | MICROSHIFT_CLUSTER_CLUSTERCIDR          | A block of IP addresses from which Pod IP addresses are allocated
| serviceNetwork        | --service-cidr            | MICROSHIFT_CLUSTER_SERVICECIDR          | A block of virtual IP addresses for Kubernetes services
| serviceNodePortRange  | --service-node-port-range | MICROSHIFT_CLUSTER_SERVICENODEPORTRANGE | The port range allowed for Kubernetes services of type NodePort
| baseDomain            | --base-domain             | MICROSHIFT_BASEDOMAIN                   | Base DNS domain used to construct fully qualified pod and service domain names
| url                   | --url                     | MICROSHIFT_CLUSTER_URL                  | URL of the API server for the cluster.
| nodeIP                | --node-ip                 | MICROSHIFT_NODEIP                       | The IP address of the node, defaults to IP of the default route
| nodeName              | --node-name               | MICROSHIFT_NODENAME                     | The name of the node, defaults to hostname
| logLevel              | --v                       | MICROSHIFT_LOGVLEVEL                    | Log verbosity (Normal, Debug, Trace, TraceAll)
| subjectAltNames     | --subject-alt-names       | MICROSHIFT_SUBJECTALTNAMES              | Subject Alternative Names for apiserver certificates

## Default Settings

In case `config.yaml` is not provided, the following default settings will be used.

```yaml
nodeName: ""
nodeIP: ""
url: https://127.0.0.1:6443
network:
  clusterNetwork:
    - cidr: 10.42.0.0/16
  serviceNetwork:
    - 10.43.0.0/16
  serviceNodePortRange: 30000-32767
dns:
  baseDomain: microshift.example.com
debugging:
  logLevel: "Normal"
```

## Service NodePort range

The `serviceNodePortRange` setting allows the extension of the port range available
to NodePort Services. This option is useful when specific standard ports under the
`30000-32767` need to be exposed. i.e., your device needs to expose the `1883/tcp`
MQTT port on the network because some client devices cannot use a different port.

If you use this option, you must be careful; NodePorts can overlap with system ports,
causing malfunction of the system or MicroShift. Take the following considerations
into account:

* Do not create any NodePort service without an explicit `nodePort` selection, in this
  case the port is assigned randomly by the kube-apiserver.
* Do not create any NodePort service for any system service port, MicroShift port,
  or other services you expose on your device HostNetwork.

List of ports that you must avoid:

| Port          | Description                                                     |
|---------------|-----------------------------------------------------------------|
| 22/tcp        | SSH port
| 80/tcp        | OpenShift Router HTTP endpoint
| 443/tcp       | OpenShift Router HTTPS endpoint
| 1936/tcp      | Metrics service for the openshift-router, not exposed today
| 2379/tcp      | etcd port
| 2380/tcp      | etcd port
| 6443          | kubernetes API
| 8445/tcp      | openshift-route-controller-manager
| 9537/tcp      | cri-o metrics
| 10250/tcp     | kubelet
| 10248/tcp     | kubelet healthz port
| 10259/tcp     | kube scheduler
|---------------|-----------------------------------------------------------------|

# Auto-applying Manifests

MicroShift leverages `kustomize` for Kubernetes-native templating and declarative management of resource objects. Upon start-up, it searches `/etc/microshift/manifests` and `/usr/lib/microshift/manifests` directories for a `kustomization.yaml` file. If it finds one, it automatically runs `kubectl apply -k` command to apply that manifest.

The reason for providing multiple directories is to allow a flexible method to manage MicroShift workloads.

| Location                      | Intent |
|-------------------------------|--------|
| /etc/microshift/manifests     | Read-write location for configuration management systems or development
| /usr/lib/microshift/manifests | Read-only location for embedding configuration manifests on ostree based systems

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
