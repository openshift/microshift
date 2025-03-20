# Configuration
The MicroShift configuration file must be located at `/etc/microshift/config.yaml`. A sample `/etc/microshift/config.yaml.default` configuration file is installed by the MicroShift RPM and it can be used as a template when customizing MicroShift.

The format of the `config.yaml` configuration file is as follows.
<!---
{{- template "docsReplaceBasic" . }}
{{- with deleteCurrent -}}
--->
```yaml
apiServer:
    advertiseAddress: ""
    auditLog:
        maxFileAge: 0
        maxFileSize: 0
        maxFiles: 0
        profile: ""
    namedCertificates:
        - certPath: ""
          keyPath: ""
          names: []
    subjectAltNames: []
    tls:
        cipherSuites: []
        minVersion: ""
debugging:
    logLevel: ""
dns:
    baseDomain: ""
etcd:
    memoryLimitMB: 0
ingress:
    certificateSecret: ""
    clientTLS:
        allowedSubjectPatterns: []
        clientCA:
            name: ""
        clientCertificatePolicy: ""
    defaultHTTPVersion: 0
    forwardedHeaderPolicy: ""
    httpCompression:
        mimeTypes: []
    httpEmptyRequestsPolicy: ""
    listenAddress: []
    logEmptyRequests: ""
    ports:
        http: 0
        https: 0
    routeAdmissionPolicy:
        namespaceOwnership: ""
        wildcardPolicy: ""
    status: ""
    tlsSecurityProfile:
        custom:
            ciphers: []
            minTLSVersion: ""
        intermediate: {}
        modern: {}
        old: {}
        type: ""
    tuningOptions:
        clientFinTimeout: ""
        clientTimeout: ""
        headerBufferBytes: 0
        headerBufferMaxRewriteBytes: 0
        healthCheckInterval: ""
        maxConnections: 0
        serverFinTimeout: ""
        serverTimeout: ""
        threadCount: 0
        tlsInspectDelay: ""
        tunnelTimeout: ""
kubelet:
manifests:
    kustomizePaths: []
network:
    clusterNetwork: []
    cniPlugin: ""
    multusStatus: ""
    serviceNetwork: []
    serviceNodePortRange: ""
node:
    hostnameOverride: ""
    nodeIP: ""
    nodeIPv6: ""
storage:
    driver: ""
    optionalCsiComponents: []
telemetry:
    endpoint: ""
    status: ""

```
<!---
{{- end }}
--->

## Default Settings

In case `config.yaml` is not provided, the following default settings will be used.
<!---
{{- template "docsReplaceFull" . }}
{{- with deleteCurrent -}}
--->
```yaml
apiServer:
    advertiseAddress: ""
    auditLog:
        maxFileAge: 0
        maxFileSize: 200
        maxFiles: 10
        profile: Default
    namedCertificates:
        - certPath: ""
          keyPath: ""
          names: []
    subjectAltNames: []
    tls:
        cipherSuites: []
        minVersion: VersionTLS12
debugging:
    logLevel: Normal
dns:
    baseDomain: example.com
etcd:
    memoryLimitMB: 0
ingress:
    certificateSecret: router-certs-default
    clientTLS:
        allowedSubjectPatterns: []
        clientCA:
            name: ""
        clientCertificatePolicy: ""
    defaultHTTPVersion: 1
    forwardedHeaderPolicy: ""
    httpCompression:
        mimeTypes: []
    httpEmptyRequestsPolicy: Respond
    listenAddress: []
    logEmptyRequests: Log
    ports:
        http: 80
        https: 443
    routeAdmissionPolicy:
        namespaceOwnership: InterNamespaceAllowed
        wildcardPolicy: ""
    status: Managed
    tlsSecurityProfile:
        custom:
            ciphers: []
            minTLSVersion: ""
        intermediate: {}
        modern: {}
        old: {}
        type: ""
    tuningOptions:
        clientFinTimeout: 1s
        clientTimeout: 30s
        headerBufferBytes: 0
        headerBufferMaxRewriteBytes: 0
        healthCheckInterval: 5s
        maxConnections: 0
        serverFinTimeout: 1s
        serverTimeout: 30s
        threadCount: 0
        tlsInspectDelay: 5s
        tunnelTimeout: 1h
kubelet:
manifests:
    kustomizePaths:
        - /usr/lib/microshift/manifests
        - /usr/lib/microshift/manifests.d/*
        - /etc/microshift/manifests
        - /etc/microshift/manifests.d/*
network:
    clusterNetwork:
        - 10.42.0.0/16
    cniPlugin: ""
    multusStatus: ""
    serviceNetwork:
        - 10.43.0.0/16
    serviceNodePortRange: 30000-32767
node:
    hostnameOverride: ""
    nodeIP: ""
    nodeIPv6: ""
storage:
    driver: ""
    optionalCsiComponents: []
telemetry:
    endpoint: https://infogw.api.openshift.com
    status: Enabled

```
<!---
{{- end }}
--->

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

## Etcd Memory Limit

By default, etcd will be allowed to use as much memory as it needs to handle the load on the system; however, in memory constrained systems, it may be preferred or necessary to limit the amount of memory etcd is allowed to use at a given time.

Setting the `memoryLimitMB` to a value greater than 0 will result in a soft memory limit being applied to etcd; etcd will be allowed to go over this value during operation, but memory will be more aggresively reclaimed from it if it does. A value of `128` megabytes is the  configuration floor - attempting to set the limit below 128 megabytes will result in the configuration being 128 megabytes.

Please note that values close to the floor may be more likely to impact etcd performance - the memory limit is a trade-off of memory footprint and etcd performance. The lower the limit, the more time etcd will spend on paging memory to disk and will take longer to respond to queries or even timing requests out if the limit is low and the etcd usage is high.

## Auto-applying Manifests

MicroShift leverages `kustomize` for Kubernetes-native templating and declarative management of resource objects. Upon start-up, it searches `/etc/microshift/manifests`, `/etc/microshift/manifests.d/*`, `/usr/lib/microshift/manifests`, and `/usr/lib/microshift/manifests.d/*` directories for a `kustomization.yaml`, `kustomization.yml`, or `Kustomization` file. If it finds one, it automatically runs `kubectl apply -k` command to apply that manifest.

Loading from multiple directories allows you to manage MicroShift workloads more flexibly. Different workloads can be independent of each other, instead of having to be loaded from a merged set of inputs.

| Location                          | Intent |
|-----------------------------------|--------|
| /etc/microshift/manifests         | Read-write location for configuration management systems or development
| /etc/microshift/manifests.d/*     | Read-write location for configuration management systems or development
| /usr/lib/microshift/manifests     | Read-only location for embedding configuration manifests on ostree based systems
| /usr/lib/microshift/manifestsd./* | Read-only location for embedding configuration manifests on ostree based systems

To override the list of paths, set `manifests.kustomizePaths` in the configuration file.

```yaml
manifests:
    kustomizePaths:
        - "/opt/alternate/path"
```

The values of `kustomizePaths` may be glob patterns.

```yaml
manifests:
    kustomizePaths:
        - "/opt/alternative/path.d/*"
```

To disable loading manifests, set the configuration option to an empty
list.

```yaml
manifests:
    kustomizePaths: []
```


### Manifest Example

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
  namespace: busybox
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
    newName: busybox:1.35
EOF
```

Restart the MicroShift service to apply the manifests and verify that the `busybox` pod is running.

```bash
sudo systemctl restart microshift
oc get pods -n busybox
```

### Deleting Manifests

MicroShift supports resource manifest deletion for data removal or upgrade scenarios.
Upgrade scenarios include situations where some objects should be removed, but not all of them to keep the data.

MicroShift scans `delete` subdirectories of configured manifests directory.
Given the default configuration, MicroShift will run `kubectl delete -k --ignore-not-found=true .` for any kustomization file found in following paths:
- `/etc/microshift/manifests/delete`
- `/etc/microshift/manifests.d/delete/*`
- `/usr/lib/microshift/manifests/delete`
- `/usr/lib/microshift/manifests.d/delete/*`

For delete scenarios, just move the existing manifest to one of the `delete` directories.

For resource removal in upgrade scenarios, is not necessary to include `spec`. Specify `group/version`, `kind`, `name`, and `namespace` of an object.
```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - all_resources.yaml

# all_resources.yaml
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: multus
  namespace: openshift-multus
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: dhcp-daemon
  namespace: openshift-multus
```

## Storage Configuration

MicroShift's included CSI plugin manages LVM LogicalVolumes to provide persistent workload storage. For LVMS
configuration, refer to the dedicated [documentation](../contributor/storage/configuration.md).

### Opt-Out of LVMS / Snapshotting Components

Users may prevent either the LVMS CSI driver or the CSI Snapshotter, or both, from being deployed. This is done by
specifying supported values under `.storage` node of the MicroShift config in the following ways:

```yaml
  storage
    driver: **ENUM**
  ```
  - Accepted values: `"none"`, `"lvms"`
  - Empty value or null field defaults to deploying LVMS.

```yaml
storage
  optionalCsiComponents: **ARRAY**.
```
  - Expected values are: `['csi-snapshot-controller', 'none']`. `'none'` is mutually exclusive
  with all other values.
  - Empty array defaults to deploying `snapshot-controller`.

### Automated Uninstallation is Not Supported

Automated uninstallation is not supported because it poses a risk of orphaning provisioned volumes. Without the LVMS CSI
driver, the cluster has no knowledge of the underlying storage interface and thus cannot perform
provisioning/deprovisioning or mount/unmount operations. Workloads with attached volumes must be manually stopped, and
those volumes must then be manually deleted by the user. Once the MicroShift config `storage` section is specified with
supported values, the user may restart MicroShift. They should see that MicroShift does not redeploy the disabled
components after restart.

## Drop-in configuration directory

In addition to the existing `/etc/microshift/config.yaml` configuration file there is a `/etc/microshift/config.d` configuration directory where you can place fragments of configuration.

At runtime, `/etc/microshift/config.yaml` and `.yaml` files inside `/etc/microshift/config.d` are merged together to create one configuration file which overrides the defaults.

Files in `/etc/microshift/config.d` are sorted lexicographilly. It is recommended to use numerical prefix for easy reasoning about the priority of the fragments.

For example, given following files:
- `/etc/microshift/config.yaml`
- `/etc/microshift/config.d/10-subjectAltNames.yaml`
- `/etc/microshift/config.d/20-kubelet.yaml`

The final user config will be created by using `config.yaml` as a base, and then overwriting it with `10-subjectAltNames.yaml`, and then overwriting it with `20-kubelet.yaml`:
```
20-kubelet.yaml
     ||
     \/
10-subjectAltNames.yaml
     ||
     \/
config.yaml
```

Some additional rules:
- Lists are not merged together, they are overwritten. For example:
  ```yaml
  # 10-san.yaml
  apiServer:
    subjectAltNames:
      - host1
      - host2

  # 20-san.yaml
  apiServer:
    subjectAltNames:
      - hostZ

  # end result
  apiServer:
    subjectAltNames:
      - hostZ
  ```
- Contents of `kubelet:` configuration are merged together (unless specific field is a list). For example:
  ```yaml
  # 10-kubelet.yaml
  kubelet:
    some_setting: True

  # 20-kubelet.yaml
  kubelet:
    another_setting: True

  # end result
  kubelet:
    some_setting: True
    another_setting: True
  ```
