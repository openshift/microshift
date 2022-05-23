// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/components/flannel/clusterrole.yaml
// assets/components/flannel/clusterrolebinding.yaml
// assets/components/flannel/configmap.yaml
// assets/components/flannel/daemonset.yaml
// assets/components/flannel/podsecuritypolicy.yaml
// assets/components/flannel/service-account.yaml
// assets/components/hostpath-provisioner/clusterrole.yaml
// assets/components/hostpath-provisioner/clusterrolebinding.yaml
// assets/components/hostpath-provisioner/daemonset.yaml
// assets/components/hostpath-provisioner/namespace.yaml
// assets/components/hostpath-provisioner/scc.yaml
// assets/components/hostpath-provisioner/service-account.yaml
// assets/components/hostpath-provisioner/storageclass.yaml
// assets/components/openshift-dns/dns/cluster-role-binding.yaml
// assets/components/openshift-dns/dns/cluster-role.yaml
// assets/components/openshift-dns/dns/configmap.yaml
// assets/components/openshift-dns/dns/daemonset.yaml
// assets/components/openshift-dns/dns/namespace.yaml
// assets/components/openshift-dns/dns/service-account.yaml
// assets/components/openshift-dns/dns/service.yaml
// assets/components/openshift-dns/node-resolver/daemonset.yaml
// assets/components/openshift-dns/node-resolver/service-account.yaml
// assets/components/openshift-router/cluster-role-binding.yaml
// assets/components/openshift-router/cluster-role.yaml
// assets/components/openshift-router/configmap.yaml
// assets/components/openshift-router/deployment.yaml
// assets/components/openshift-router/namespace.yaml
// assets/components/openshift-router/service-account.yaml
// assets/components/openshift-router/service-cloud.yaml
// assets/components/openshift-router/service-internal.yaml
// assets/components/service-ca/clusterrole.yaml
// assets/components/service-ca/clusterrolebinding.yaml
// assets/components/service-ca/deployment.yaml
// assets/components/service-ca/ns.yaml
// assets/components/service-ca/role.yaml
// assets/components/service-ca/rolebinding.yaml
// assets/components/service-ca/sa.yaml
// assets/components/service-ca/signing-cabundle.yaml
// assets/components/service-ca/signing-secret.yaml
// assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml
// assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml
// assets/crd/0000_03_config-operator_01_proxy.crd.yaml
// assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml
// assets/crd/0000_03_security-openshift_01_scc.crd.yaml
// assets/crd/0000_10_config-operator_01_build.crd.yaml
// assets/crd/0000_10_config-operator_01_featuregate.crd.yaml
// assets/crd/0000_10_config-operator_01_image.crd.yaml
// assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml
// assets/crd/0000_11_imageregistry-configs.crd.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml
package assets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsComponentsFlannelClusterroleYaml = []byte(`kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
rules:
- apiGroups: ['extensions']
  resources: ['podsecuritypolicies']
  verbs: ['use']
  resourceNames: ['psp.flannel.unprivileged']
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch`)

func assetsComponentsFlannelClusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelClusterroleYaml, nil
}

func assetsComponentsFlannelClusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/clusterrole.yaml", size: 418, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannelClusterrolebindingYaml = []byte(`kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: flannel
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: flannel
subjects:
- kind: ServiceAccount
  name: flannel
  namespace: kube-system`)

func assetsComponentsFlannelClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelClusterrolebindingYaml, nil
}

func assetsComponentsFlannelClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/clusterrolebinding.yaml", size: 248, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannelConfigmapYaml = []byte(`kind: ConfigMap
apiVersion: v1
metadata:
  name: kube-flannel-cfg
  namespace: kube-system
  labels:
    tier: node
    app: flannel
data:
  cni-conf.json: |
    {
      "name": "cbr0",
      "cniVersion": "0.3.1",
      "plugins": [
        {
          "type": "flannel",
          "delegate": {
            "hairpinMode": true,
            "forceAddress": true,
            "isDefaultGateway": true
          }
        },
        {
          "type": "portmap",
          "capabilities": {
            "portMappings": true
          }
        }
      ]
    }
  net-conf.json: |
    {
      "Network": "10.42.0.0/16",
      "Backend": {
        "Type": "vxlan"
      }
    }`)

func assetsComponentsFlannelConfigmapYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelConfigmapYaml, nil
}

func assetsComponentsFlannelConfigmapYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/configmap.yaml", size: 674, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannelDaemonsetYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kube-flannel-ds
  namespace: kube-system
  labels:
    tier: node
    app: flannel
spec:
  selector:
    matchLabels:
      app: flannel
  template:
    metadata:
      labels:
        tier: node
        app: flannel
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: kubernetes.io/os
                operator: In
                values:
                - linux
      hostNetwork: true
      priorityClassName: system-node-critical
      tolerations:
      - operator: Exists
        effect: NoSchedule
      serviceAccountName: flannel
      initContainers:
      - name: install-cni-bin
        image: {{ .ReleaseImage.kube_flannel_cni }}
        imagePullPolicy: IfNotPresent
        command:
        - cp
        args:
        - -f
        - /flannel
        - /opt/cni/bin/flannel
        volumeMounts:
        - name: cni-plugin
          mountPath: /opt/cni/bin
      - name: install-cni
        image: {{ .ReleaseImage.kube_flannel }}
        imagePullPolicy: IfNotPresent
        command:
        - cp
        args:
        - -f
        - /etc/kube-flannel/cni-conf.json
        - /etc/cni/net.d/10-flannel.conflist
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      containers:
      - name: kube-flannel
        image: {{ .ReleaseImage.kube_flannel }}
        imagePullPolicy: IfNotPresent
        command:
        - /opt/bin/flanneld
        args:
        - --ip-masq
        - --kube-subnet-mgr
        resources:
          requests:
            cpu: "100m"
            memory: "50Mi"
          limits:
            cpu: "100m"
            memory: "50Mi"
        securityContext:
          privileged: false
          capabilities:
            add: ["NET_ADMIN", "NET_RAW"]
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - name: run
          mountPath: /run/flannel
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
      - name: run
        hostPath:
          path: /run/flannel
      - name: cni
        hostPath:
          path: /etc/cni/net.d
      - name: flannel-cfg
        configMap:
          name: kube-flannel-cfg
      - name: cni-plugin
        hostPath:
          path: /opt/cni/bin`)

func assetsComponentsFlannelDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelDaemonsetYaml, nil
}

func assetsComponentsFlannelDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/daemonset.yaml", size: 2657, mode: os.FileMode(436), modTime: time.Unix(1653308425, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannelPodsecuritypolicyYaml = []byte(`apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: psp.flannel.unprivileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: docker/default
    seccomp.security.alpha.kubernetes.io/defaultProfileName: docker/default
    apparmor.security.beta.kubernetes.io/allowedProfileNames: runtime/default
    apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
spec:
  privileged: false
  volumes:
  - configMap
  - secret
  - emptyDir
  - hostPath
  allowedHostPaths:
  - pathPrefix: "/etc/cni/net.d"
  - pathPrefix: "/etc/kube-flannel"
  - pathPrefix: "/run/flannel"
  readOnlyRootFilesystem: false
  # Users and groups
  runAsUser:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  # Privilege Escalation
  allowPrivilegeEscalation: false
  defaultAllowPrivilegeEscalation: false
  # Capabilities
  allowedCapabilities: ['NET_ADMIN', 'NET_RAW']
  defaultAddCapabilities: []
  requiredDropCapabilities: []
  # Host namespaces
  hostPID: false
  hostIPC: false
  hostNetwork: true
  hostPorts:
  - min: 0
    max: 65535
  # SELinux
  seLinux:
    # SELinux is unused in CaaSP
    rule: 'RunAsAny'`)

func assetsComponentsFlannelPodsecuritypolicyYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelPodsecuritypolicyYaml, nil
}

func assetsComponentsFlannelPodsecuritypolicyYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelPodsecuritypolicyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/podsecuritypolicy.yaml", size: 1195, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannelServiceAccountYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: flannel
  namespace: kube-system`)

func assetsComponentsFlannelServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelServiceAccountYaml, nil
}

func assetsComponentsFlannelServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/service-account.yaml", size: 86, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerClusterroleYaml = []byte(`kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kubevirt-hostpath-provisioner
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]

  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]

  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
`)

func assetsComponentsHostpathProvisionerClusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerClusterroleYaml, nil
}

func assetsComponentsHostpathProvisionerClusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/clusterrole.yaml", size: 609, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubevirt-hostpath-provisioner
subjects:
- kind: ServiceAccount
  name: kubevirt-hostpath-provisioner-admin
  namespace: kubevirt-hostpath-provisioner
roleRef:
  kind: ClusterRole
  name: kubevirt-hostpath-provisioner
  apiGroup: rbac.authorization.k8s.io`)

func assetsComponentsHostpathProvisionerClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerClusterrolebindingYaml, nil
}

func assetsComponentsHostpathProvisionerClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/clusterrolebinding.yaml", size: 338, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerDaemonsetYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kubevirt-hostpath-provisioner
  labels:
    k8s-app: kubevirt-hostpath-provisioner
  namespace: kubevirt-hostpath-provisioner
spec:
  selector:
    matchLabels:
      k8s-app: kubevirt-hostpath-provisioner
  template:
    metadata:
      labels:
        k8s-app: kubevirt-hostpath-provisioner
    spec:
      serviceAccountName: kubevirt-hostpath-provisioner-admin
      containers:
        - name: kubevirt-hostpath-provisioner
          image: {{ .ReleaseImage.kubevirt_hostpath_provisioner }}
          imagePullPolicy: IfNotPresent
          env:
            - name: USE_NAMING_PREFIX
              value: "false" # change to true, to have the name of the pvc be part of the directory
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: PV_DIR
              value: /var/hpvolumes
          volumeMounts:
            - name: pv-volume # root dir where your bind mounts will be on the node
              mountPath: /var/hpvolumes
              #nodeSelector:
              #- name: xxxxxx
      volumes:
        - name: pv-volume
          hostPath:
            path: /var/hpvolumes
`)

func assetsComponentsHostpathProvisionerDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerDaemonsetYaml, nil
}

func assetsComponentsHostpathProvisionerDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/daemonset.yaml", size: 1231, mode: os.FileMode(436), modTime: time.Unix(1653308443, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerNamespaceYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: kubevirt-hostpath-provisioner`)

func assetsComponentsHostpathProvisionerNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerNamespaceYaml, nil
}

func assetsComponentsHostpathProvisionerNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/namespace.yaml", size: 78, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerSccYaml = []byte(`kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: hostpath-provisioner
allowPrivilegedContainer: true
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
allowHostDirVolumePlugin: true
users:
- system:serviceaccount:kubevirt-hostpath-provisioner:kubevirt-hostpath-provisioner-admin
volumes:
- hostPath
- secret
`)

func assetsComponentsHostpathProvisionerSccYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerSccYaml, nil
}

func assetsComponentsHostpathProvisionerSccYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerSccYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/scc.yaml", size: 480, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerServiceAccountYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubevirt-hostpath-provisioner-admin
  namespace: kubevirt-hostpath-provisioner`)

func assetsComponentsHostpathProvisionerServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerServiceAccountYaml, nil
}

func assetsComponentsHostpathProvisionerServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/service-account.yaml", size: 132, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerStorageclassYaml = []byte(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: kubevirt-hostpath-provisioner
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: kubevirt.io/hostpath-provisioner
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
`)

func assetsComponentsHostpathProvisionerStorageclassYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerStorageclassYaml, nil
}

func assetsComponentsHostpathProvisionerStorageclassYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerStorageclassYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/storageclass.yaml", size: 276, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsClusterRoleBindingYaml = []byte(`kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
    name: openshift-dns
subjects:
- kind: ServiceAccount
  name: dns
  namespace: openshift-dns
roleRef:
  kind: ClusterRole
  name: openshift-dns
`)

func assetsComponentsOpenshiftDnsDnsClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsClusterRoleBindingYaml, nil
}

func assetsComponentsOpenshiftDnsDnsClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/cluster-role-binding.yaml", size: 223, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsClusterRoleYaml = []byte(`kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: openshift-dns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch

- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch

- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create

- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
`)

func assetsComponentsOpenshiftDnsDnsClusterRoleYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsClusterRoleYaml, nil
}

func assetsComponentsOpenshiftDnsDnsClusterRoleYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/cluster-role.yaml", size: 492, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsConfigmapYaml = []byte(`apiVersion: v1
data:
  Corefile: |
    .:5353 {
        bufsize 512
        errors
        health {
            lameduck 20s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            fallthrough in-addr.arpa ip6.arpa
        }
        prometheus 127.0.0.1:9153
        forward . /etc/resolv.conf {
            policy sequential
        }
        cache 900 {
            denial 9984 30
        }
        reload
    }
kind: ConfigMap
metadata:
  labels:
    dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns
`)

func assetsComponentsOpenshiftDnsDnsConfigmapYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsConfigmapYaml, nil
}

func assetsComponentsOpenshiftDnsDnsConfigmapYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/configmap.yaml", size: 610, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsDaemonsetYaml = []byte(`kind: DaemonSet
apiVersion: apps/v1
metadata:
  labels:
    dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns
spec:
  selector:
    matchLabels:
      dns.operator.openshift.io/daemonset-dns: default
  template:
    metadata:
      labels:
        dns.operator.openshift.io/daemonset-dns: default
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
    spec:
      serviceAccountName: dns
      priorityClassName: system-node-critical
      containers:
      - name: dns
        image: {{ .ReleaseImage.coredns }}
        imagePullPolicy: IfNotPresent
        terminationMessagePolicy: FallbackToLogsOnError
        command: [ "coredns" ]
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 5353
          name: dns
          protocol: UDP
        - containerPort: 5353
          name: dns-tcp
          protocol: TCP
        readinessProbe:
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 3
          successThreshold: 1
          failureThreshold: 3
          timeoutSeconds: 3
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        resources:
          requests:
            cpu: 50m
            memory: 70Mi
      - name: kube-rbac-proxy
        image: {{ .ReleaseImage.kube_rbac_proxy }}
        imagePullPolicy: IfNotPresent
        args:
        - --secure-listen-address=:9154
        - --tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
        - --upstream=http://127.0.0.1:9153/
        - --tls-cert-file=/etc/tls/private/tls.crt
        - --tls-private-key-file=/etc/tls/private/tls.key
        ports:
        - containerPort: 9154
          name: metrics
        resources:
          requests:
            cpu: 10m
            memory: 40Mi
        volumeMounts:
        - mountPath: /etc/tls/private
          name: metrics-tls
          readOnly: true
      dnsPolicy: Default
      nodeSelector:
        kubernetes.io/os: linux      
      volumes:
      - name: config-volume
        configMap:
          items:
          - key: Corefile
            path: Corefile
          name: dns-default
      - name: metrics-tls
        secret:
          defaultMode: 420
          secretName: dns-default-metrics-tls
      tolerations:
      # DNS needs to run everywhere. Tolerate all taints
      - operator: Exists
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      # TODO: Consider setting maxSurge to a positive value.
      maxSurge: 0
      # Note: The daemon controller rounds the percentage up
      # (unlike the deployment controller, which rounds down).
      maxUnavailable: 10%
`)

func assetsComponentsOpenshiftDnsDnsDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsDaemonsetYaml, nil
}

func assetsComponentsOpenshiftDnsDnsDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/daemonset.yaml", size: 3217, mode: os.FileMode(436), modTime: time.Unix(1653308465, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsNamespaceYaml = []byte(`kind: Namespace
apiVersion: v1
metadata:
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  name: openshift-dns
  labels:
    # set value to avoid depending on kube admission that depends on openshift apis
    openshift.io/run-level: "0"
    # allow openshift-monitoring to look for ServiceMonitor objects in this namespace
    openshift.io/cluster-monitoring: "true"
`)

func assetsComponentsOpenshiftDnsDnsNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsNamespaceYaml, nil
}

func assetsComponentsOpenshiftDnsDnsNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/namespace.yaml", size: 417, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsServiceAccountYaml = []byte(`kind: ServiceAccount
apiVersion: v1
metadata:
  name: dns
  namespace: openshift-dns
`)

func assetsComponentsOpenshiftDnsDnsServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsServiceAccountYaml, nil
}

func assetsComponentsOpenshiftDnsDnsServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/service-account.yaml", size: 85, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsDnsServiceYaml = []byte(`kind: Service
apiVersion: v1
metadata:
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: dns-default-metrics-tls
  labels:
      dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns
spec:
  clusterIP: {{.ClusterIP}}
  selector:
    dns.operator.openshift.io/daemonset-dns: default
  ports:
  - name: dns
    port: 53
    targetPort: dns
    protocol: UDP
  - name: dns-tcp
    port: 53
    targetPort: dns-tcp
    protocol: TCP
  - name: metrics
    port: 9154
    targetPort: metrics
    protocol: TCP
  # TODO: Uncomment when service topology feature gate is enabled.
  #topologyKeys:
  #  - "kubernetes.io/hostname"
  #  - "*"
`)

func assetsComponentsOpenshiftDnsDnsServiceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsServiceYaml, nil
}

func assetsComponentsOpenshiftDnsDnsServiceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsServiceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/service.yaml", size: 691, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsNodeResolverDaemonsetYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: node-resolver
  namespace: openshift-dns
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      dns.operator.openshift.io/daemonset-node-resolver: ""
  template:
    metadata:
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        dns.operator.openshift.io/daemonset-node-resolver: ""
    spec:
      containers:
      - command:
        - /bin/bash
        - -c
        - |
          #!/bin/bash
          set -uo pipefail

          trap 'jobs -p | xargs kill || true; wait; exit 0' TERM

          NAMESERVER=${DNS_DEFAULT_SERVICE_HOST}
          OPENSHIFT_MARKER="openshift-generated-node-resolver"
          HOSTS_FILE="/etc/hosts"
          TEMP_FILE="/etc/hosts.tmp"

          IFS=', ' read -r -a services <<< "${SERVICES}"

          # Make a temporary file with the old hosts file's attributes.
          cp -f --attributes-only "${HOSTS_FILE}" "${TEMP_FILE}"

          while true; do
            declare -A svc_ips
            for svc in "${services[@]}"; do
              # Fetch service IP from cluster dns if present. We make several tries
              # to do it: IPv4, IPv6, IPv4 over TCP and IPv6 over TCP. The two last ones
              # are for deployments with Kuryr on older OpenStack (OSP13) - those do not
              # support UDP loadbalancers and require reaching DNS through TCP.
              cmds=('dig -t A @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"'
                    'dig -t AAAA @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"'
                    'dig -t A +tcp +retry=0 @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"'
                    'dig -t AAAA +tcp +retry=0 @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"')
              for i in ${!cmds[*]}
              do
                ips=($(eval "${cmds[i]}"))
                if [[ "$?" -eq 0 && "${#ips[@]}" -ne 0 ]]; then
                  svc_ips["${svc}"]="${ips[@]}"
                  break
                fi
              done
            done

            # Update /etc/hosts only if we get valid service IPs
            # We will not update /etc/hosts when there is coredns service outage or api unavailability
            # Stale entries could exist in /etc/hosts if the service is deleted
            if [[ -n "${svc_ips[*]-}" ]]; then
              # Build a new hosts file from /etc/hosts with our custom entries filtered out
              grep -v "# ${OPENSHIFT_MARKER}" "${HOSTS_FILE}" > "${TEMP_FILE}"

              # Append resolver entries for services
              for svc in "${!svc_ips[@]}"; do
                for ip in ${svc_ips[${svc}]}; do
                  echo "${ip} ${svc} ${svc}.${CLUSTER_DOMAIN} # ${OPENSHIFT_MARKER}" >> "${TEMP_FILE}"
                done
              done

              # TODO: Update /etc/hosts atomically to avoid any inconsistent behavior
              # Replace /etc/hosts with our modified version if needed
              cmp "${TEMP_FILE}" "${HOSTS_FILE}" || cp -f "${TEMP_FILE}" "${HOSTS_FILE}"
              # TEMP_FILE is not removed to avoid file create/delete and attributes copy churn
            fi
            sleep 60 & wait
            unset svc_ips
          done
        env:
        - name: SERVICES
          # Comma or space separated list of services
          # NOTE: For now, ensure these are relative names; for each relative name,
          # an alias with the CLUSTER_DOMAIN suffix will also be added.
          value: "image-registry.openshift-image-registry.svc"
        - name: NAMESERVER
          value: 172.30.0.10
        - name: CLUSTER_DOMAIN
          value: cluster.local
        image: {{ .ReleaseImage.cli }}
        imagePullPolicy: IfNotPresent
        name: dns-node-resolver
        resources:
          requests:
            cpu: 5m
            memory: 21Mi
        securityContext:
          privileged: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /etc/hosts
          name: hosts-file
      dnsPolicy: ClusterFirst
      hostNetwork: true
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-node-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: node-resolver
      serviceAccountName: node-resolver
      terminationGracePeriodSeconds: 30
      tolerations:
      - operator: Exists
      volumes:
      - hostPath:
          path: /etc/hosts
          type: File
        name: hosts-file
  updateStrategy:
    rollingUpdate:
      maxSurge: 0
      maxUnavailable: 33%
    type: RollingUpdate
`)

func assetsComponentsOpenshiftDnsNodeResolverDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsNodeResolverDaemonsetYaml, nil
}

func assetsComponentsOpenshiftDnsNodeResolverDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsNodeResolverDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/node-resolver/daemonset.yaml", size: 4823, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsNodeResolverServiceAccountYaml = []byte(`kind: ServiceAccount
apiVersion: v1
metadata:
  name: node-resolver
  namespace: openshift-dns
`)

func assetsComponentsOpenshiftDnsNodeResolverServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsNodeResolverServiceAccountYaml, nil
}

func assetsComponentsOpenshiftDnsNodeResolverServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsNodeResolverServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/node-resolver/service-account.yaml", size: 95, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterClusterRoleBindingYaml = []byte(`# Binds the router role to its Service Account.
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: openshift-ingress-router
subjects:
- kind: ServiceAccount
  name: router
  namespace: openshift-ingress
roleRef:
  kind: ClusterRole
  name: openshift-ingress-router
  namespace: openshift-ingress
`)

func assetsComponentsOpenshiftRouterClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterClusterRoleBindingYaml, nil
}

func assetsComponentsOpenshiftRouterClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/cluster-role-binding.yaml", size: 329, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterClusterRoleYaml = []byte(`# Cluster scoped role for routers. This should be as restrictive as possible.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: openshift-ingress-router
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - namespaces
  - services
  verbs:
  - list
  - watch

- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create

- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create

- apiGroups:
  - route.openshift.io
  resources:
  - routes
  verbs:
  - list
  - watch

- apiGroups:
  - route.openshift.io
  resources:
  - routes/status
  verbs:
  - update

- apiGroups:
  - security.openshift.io
  resources:
  - securitycontextconstraints
  verbs:
  - use
  resourceNames:
  - hostnetwork

- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
`)

func assetsComponentsOpenshiftRouterClusterRoleYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterClusterRoleYaml, nil
}

func assetsComponentsOpenshiftRouterClusterRoleYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/cluster-role.yaml", size: 883, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterConfigmapYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: openshift-ingress
  name: service-ca-bundle 
  annotations:
    service.beta.openshift.io/inject-cabundle: "true"
`)

func assetsComponentsOpenshiftRouterConfigmapYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterConfigmapYaml, nil
}

func assetsComponentsOpenshiftRouterConfigmapYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/configmap.yaml", size: 168, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterDeploymentYaml = []byte(`# Deployment with default values
# Ingress Controller specific values are applied at runtime.
kind: Deployment
apiVersion: apps/v1
metadata:
  name: router-default
  namespace: openshift-ingress
  labels:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  strategy:
    rollingUpdate:
      maxSurge: 0
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations:
        "unsupported.do-not-use.openshift.io/override-liveness-grace-period-seconds": "10"
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
    spec:
      serviceAccountName: router
      # nodeSelector is set at runtime.
      priorityClassName: system-cluster-critical
      containers:
        - name: router
          image: {{ .ReleaseImage.haproxy_router }}
          imagePullPolicy: IfNotPresent
          terminationMessagePolicy: FallbackToLogsOnError
          ports:
          - name: http
            containerPort: 80
            hostPort: 80
            protocol: TCP
          - name: https
            containerPort: 443
            hostPort: 443
            protocol: TCP
          - name: metrics
            containerPort: 1936
            hostPort: 1936
            protocol: TCP
          # Merged at runtime.
          env:
          # stats username and password are generated at runtime
          - name: STATS_PORT
            value: "1936"
          - name: ROUTER_SERVICE_NAMESPACE
            value: openshift-ingress
          - name: DEFAULT_CERTIFICATE_DIR
            value: /etc/pki/tls/private
          - name: DEFAULT_DESTINATION_CA_PATH
            value: /var/run/configmaps/service-ca/service-ca.crt
          - name: ROUTER_CIPHERS
            value: TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384
          - name: ROUTER_DISABLE_HTTP2
            value: "true"
          - name: ROUTER_DISABLE_NAMESPACE_OWNERSHIP_CHECK
            value: "false"
          #FIXME: use metrics tls
          - name: ROUTER_METRICS_TLS_CERT_FILE
            value: /etc/pki/tls/private/tls.crt
          - name: ROUTER_METRICS_TLS_KEY_FILE
            value: /etc/pki/tls/private/tls.key
          - name: ROUTER_METRICS_TYPE
            value: haproxy
          - name: ROUTER_SERVICE_NAME
            value: default
          - name: ROUTER_SET_FORWARDED_HEADERS
            value: append
          - name: ROUTER_THREADS
            value: "4"
          - name: SSL_MIN_VERSION
            value: TLSv1.2
          livenessProbe:
            failureThreshold: 3
            httpGet:
              host: localhost
              path: /healthz
              port: 1936
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          readinessProbe:
            failureThreshold: 3
            httpGet:
              host: localhost
              path: /healthz/ready
              port: 1936
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          startupProbe:
            failureThreshold: 120
            httpGet:
              path: /healthz/ready
              port: 1936
            periodSeconds: 1
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
          volumeMounts:
          - mountPath: /etc/pki/tls/private
            name: default-certificate
            readOnly: true
          - mountPath: /var/run/configmaps/service-ca
            name: service-ca-bundle
            readOnly: true
      dnsPolicy: ClusterFirstWithHostNet
      hostNetwork: true
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: router
      volumes:
      - name: default-certificate
        secret:
          defaultMode: 420
          secretName: router-certs-default
      - name: service-ca-bundle
        configMap:
          items:
          - key: service-ca.crt
            path: service-ca.crt
          name: service-ca-bundle
          optional: false
        defaultMode: 420
`)

func assetsComponentsOpenshiftRouterDeploymentYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterDeploymentYaml, nil
}

func assetsComponentsOpenshiftRouterDeploymentYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/deployment.yaml", size: 4746, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterNamespaceYaml = []byte(`kind: Namespace
apiVersion: v1
metadata:
  name: openshift-ingress
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  labels:
    # allow openshift-monitoring to look for ServiceMonitor objects in this namespace
    openshift.io/cluster-monitoring: "true"
    name: openshift-ingress
    # old and new forms of the label for matching with NetworkPolicy
    network.openshift.io/policy-group: ingress
    policy-group.network.openshift.io/ingress: ""
`)

func assetsComponentsOpenshiftRouterNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterNamespaceYaml, nil
}

func assetsComponentsOpenshiftRouterNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/namespace.yaml", size: 499, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterServiceAccountYaml = []byte(`# Account for routers created by the operator. It will require cluster scoped
# permissions related to Route processing.
kind: ServiceAccount
apiVersion: v1
metadata:
  name: router
  namespace: openshift-ingress
`)

func assetsComponentsOpenshiftRouterServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterServiceAccountYaml, nil
}

func assetsComponentsOpenshiftRouterServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/service-account.yaml", size: 213, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterServiceCloudYaml = []byte(`kind: Service
apiVersion: v1
metadata:
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: router-certs-default
  labels:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  name: router-external-default
  namespace: openshift-ingress
spec:
  selector:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  type: NodePort 
  ports:
    - name: http
      port: 80
      targetPort: 80
      nodePort: 30001
    - name: https
      port: 443
      targetPort: 443
      nodePort: 30002
`)

func assetsComponentsOpenshiftRouterServiceCloudYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterServiceCloudYaml, nil
}

func assetsComponentsOpenshiftRouterServiceCloudYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterServiceCloudYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/service-cloud.yaml", size: 567, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterServiceInternalYaml = []byte(`# Cluster Service with default values
# Ingress Controller specific annotations are applied at runtime.
kind: Service
apiVersion: v1
metadata:
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: router-certs-default
  labels:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  name: router-internal-default
  namespace: openshift-ingress
spec:
  selector:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  type: ClusterIP
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: http
  - name: https
    port: 443
    protocol: TCP
    targetPort: https
  - name: metrics
    port: 1936
    protocol: TCP
    targetPort: 1936
`)

func assetsComponentsOpenshiftRouterServiceInternalYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterServiceInternalYaml, nil
}

func assetsComponentsOpenshiftRouterServiceInternalYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterServiceInternalYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/service-internal.yaml", size: 727, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:openshift:controller:service-ca
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - apiregistration.k8s.io
  resources:
  - apiservices
  verbs:
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - update
`)

func assetsComponentsServiceCaClusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaClusterroleYaml, nil
}

func assetsComponentsServiceCaClusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/clusterrole.yaml", size: 864, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:controller:service-ca
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  namespace: openshift-service-ca
  name: service-ca
`)

func assetsComponentsServiceCaClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaClusterrolebindingYaml, nil
}

func assetsComponentsServiceCaClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/clusterrolebinding.yaml", size: 298, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaDeploymentYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: openshift-service-ca
  name: service-ca
  labels:
    app: service-ca
    service-ca: "true"
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: service-ca
      service-ca: "true"
  template:
    metadata:
      name: service-ca
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        app: service-ca
        service-ca: "true"
    spec:
      securityContext: {}
      serviceAccount: service-ca
      serviceAccountName: service-ca
      containers:
      - name: service-ca-controller
        image: {{ .ReleaseImage.service_ca_operator }}
        imagePullPolicy: IfNotPresent
        command: ["service-ca-operator", "controller"]
        ports:
        - containerPort: 8443
        # securityContext:
        #   runAsNonRoot: true
        resources:
          requests:
            memory: 120Mi
            cpu: 10m
        volumeMounts:
        - mountPath: /var/run/secrets/signing-key
          name: signing-key
        - mountPath: /var/run/configmaps/signing-cabundle
          name: signing-cabundle
      volumes:
      - name: signing-key
        secret:
          secretName: {{.TLSSecret}}
      - name: signing-cabundle
        configMap:
          name: {{.CAConfigMap}}
      # nodeSelector:
      #   node-role.kubernetes.io/master: ""
      priorityClassName: "system-cluster-critical"
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: "NoSchedule"
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
`)

func assetsComponentsServiceCaDeploymentYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaDeploymentYaml, nil
}

func assetsComponentsServiceCaDeploymentYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/deployment.yaml", size: 1866, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaNsYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: openshift-service-ca
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
`)

func assetsComponentsServiceCaNsYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaNsYaml, nil
}

func assetsComponentsServiceCaNsYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaNsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/ns.yaml", size: 168, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaRoleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: system:openshift:controller:service-ca
  namespace: openshift-service-ca
rules:
- apiGroups:
  - security.openshift.io
  resources:
  - securitycontextconstraints
  resourceNames:
  - restricted
  verbs:
  - use
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - update
  - create
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - "apps"
  resources:
  - replicasets
  - deployments
  verbs:
  - get
  - list
  - watch`)

func assetsComponentsServiceCaRoleYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaRoleYaml, nil
}

func assetsComponentsServiceCaRoleYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/role.yaml", size: 634, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaRolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: system:openshift:controller:service-ca
  namespace: openshift-service-ca
roleRef:
  kind: Role
  name: system:openshift:controller:service-ca
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  namespace: openshift-service-ca
  name: service-ca
`)

func assetsComponentsServiceCaRolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaRolebindingYaml, nil
}

func assetsComponentsServiceCaRolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/rolebinding.yaml", size: 343, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaSaYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: openshift-service-ca
  name: service-ca
`)

func assetsComponentsServiceCaSaYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaSaYaml, nil
}

func assetsComponentsServiceCaSaYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaSaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/sa.yaml", size: 99, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaSigningCabundleYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: openshift-service-ca
  name: signing-cabundle
data:
  ca-bundle.crt:
`)

func assetsComponentsServiceCaSigningCabundleYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaSigningCabundleYaml, nil
}

func assetsComponentsServiceCaSigningCabundleYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaSigningCabundleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/signing-cabundle.yaml", size: 123, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaSigningSecretYaml = []byte(`apiVersion: v1
kind: Secret
metadata:
  namespace: openshift-service-ca
  name: signing-key
type: kubernetes.io/tls
data:
  tls.crt:
  tls.key:
`)

func assetsComponentsServiceCaSigningSecretYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaSigningSecretYaml, nil
}

func assetsComponentsServiceCaSigningSecretYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaSigningSecretYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/signing-secret.yaml", size: 144, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  annotations:
    include.release.openshift.io/self-managed-high-availability: "true"
    openshift.io/node-selector: ""
  labels:
    openshift.io/cluster-monitoring: "true"
  name: openshift-controller-manager
`)

func assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYamlBytes() ([]byte, error) {
	return _assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml, nil
}

func assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml() (*asset, error) {
	bytes, err := assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml", size: 254, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: rolebindingrestrictions.authorization.openshift.io
spec:
  group: authorization.openshift.io
  names:
    kind: RoleBindingRestriction
    listKind: RoleBindingRestrictionList
    plural: rolebindingrestrictions
    singular: rolebindingrestriction
  scope: Namespaced
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: "RoleBindingRestriction is an object that can be matched against a subject (user, group, or service account) to determine whether rolebindings on that subject are allowed in the namespace to which the RoleBindingRestriction belongs.  If any one of those RoleBindingRestriction objects matches a subject, rolebindings on that subject in the namespace are allowed. \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec defines the matcher.
              type: object
              properties:
                grouprestriction:
                  description: GroupRestriction matches against group subjects.
                  type: object
                  properties:
                    groups:
                      description: Groups is a list of groups used to match against an individual user's groups. If the user is a member of one of the whitelisted groups, the user is allowed to be bound to a role.
                      type: array
                      items:
                        type: string
                      nullable: true
                    labels:
                      description: Selectors specifies a list of label selectors over group labels.
                      type: array
                      items:
                        description: A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.
                        type: object
                        properties:
                          matchExpressions:
                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                            type: array
                            items:
                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                              type: object
                              required:
                                - key
                                - operator
                              properties:
                                key:
                                  description: key is the label key that the selector applies to.
                                  type: string
                                operator:
                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                  type: string
                                values:
                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                  type: array
                                  items:
                                    type: string
                          matchLabels:
                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                            type: object
                            additionalProperties:
                              type: string
                      nullable: true
                  nullable: true
                serviceaccountrestriction:
                  description: ServiceAccountRestriction matches against service-account subjects.
                  type: object
                  properties:
                    namespaces:
                      description: Namespaces specifies a list of literal namespace names.
                      type: array
                      items:
                        type: string
                    serviceaccounts:
                      description: ServiceAccounts specifies a list of literal service-account names.
                      type: array
                      items:
                        description: ServiceAccountReference specifies a service account and namespace by their names.
                        type: object
                        properties:
                          name:
                            description: Name is the name of the service account.
                            type: string
                          namespace:
                            description: Namespace is the namespace of the service account.  Service accounts from inside the whitelisted namespaces are allowed to be bound to roles.  If Namespace is empty, then the namespace of the RoleBindingRestriction in which the ServiceAccountReference is embedded is used.
                            type: string
                  nullable: true
                userrestriction:
                  description: UserRestriction matches against user subjects.
                  type: object
                  properties:
                    groups:
                      description: Groups specifies a list of literal group names.
                      type: array
                      items:
                        type: string
                      nullable: true
                    labels:
                      description: Selectors specifies a list of label selectors over user labels.
                      type: array
                      items:
                        description: A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.
                        type: object
                        properties:
                          matchExpressions:
                            description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                            type: array
                            items:
                              description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                              type: object
                              required:
                                - key
                                - operator
                              properties:
                                key:
                                  description: key is the label key that the selector applies to.
                                  type: string
                                operator:
                                  description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                  type: string
                                values:
                                  description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                  type: array
                                  items:
                                    type: string
                          matchLabels:
                            description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                            type: object
                            additionalProperties:
                              type: string
                      nullable: true
                    users:
                      description: Users specifies a list of literal user names.
                      type: array
                      items:
                        type: string
                  nullable: true
      served: true
      storage: true
`)

func assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml, nil
}

func assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml", size: 9898, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_configOperator_01_proxyCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: proxies.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Proxy
    listKind: ProxyList
    plural: proxies
    singular: proxy
  scope: Cluster
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: "Proxy holds cluster-wide information on how to configure default proxies for the cluster. The canonical name is ` + "`" + `cluster` + "`" + ` \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          required:
            - spec
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec holds user-settable values for the proxy configuration
              type: object
              properties:
                httpProxy:
                  description: httpProxy is the URL of the proxy for HTTP requests.  Empty means unset and will not result in an env var.
                  type: string
                httpsProxy:
                  description: httpsProxy is the URL of the proxy for HTTPS requests.  Empty means unset and will not result in an env var.
                  type: string
                noProxy:
                  description: noProxy is a comma-separated list of hostnames and/or CIDRs and/or IPs for which the proxy should not be used. Empty means unset and will not result in an env var.
                  type: string
                readinessEndpoints:
                  description: readinessEndpoints is a list of endpoints used to verify readiness of the proxy.
                  type: array
                  items:
                    type: string
                trustedCA:
                  description: "trustedCA is a reference to a ConfigMap containing a CA certificate bundle. The trustedCA field should only be consumed by a proxy validator. The validator is responsible for reading the certificate bundle from the required key \"ca-bundle.crt\", merging it with the system default trust bundle, and writing the merged trust bundle to a ConfigMap named \"trusted-ca-bundle\" in the \"openshift-config-managed\" namespace. Clients that expect to make proxy connections must use the trusted-ca-bundle for all HTTPS requests to the proxy, and may use the trusted-ca-bundle for non-proxy HTTPS requests as well. \n The namespace for the ConfigMap referenced by trustedCA is \"openshift-config\". Here is an example ConfigMap (in yaml): \n apiVersion: v1 kind: ConfigMap metadata:  name: user-ca-bundle  namespace: openshift-config  data:    ca-bundle.crt: |      -----BEGIN CERTIFICATE-----      Custom CA certificate bundle.      -----END CERTIFICATE-----"
                  type: object
                  required:
                    - name
                  properties:
                    name:
                      description: name is the metadata.name of the referenced config map
                      type: string
            status:
              description: status holds observed values from the cluster. They may not be overridden.
              type: object
              properties:
                httpProxy:
                  description: httpProxy is the URL of the proxy for HTTP requests.
                  type: string
                httpsProxy:
                  description: httpsProxy is the URL of the proxy for HTTPS requests.
                  type: string
                noProxy:
                  description: noProxy is a comma-separated list of hostnames and/or CIDRs for which the proxy should not be used.
                  type: string
      served: true
      storage: true
      subresources:
        status: {}
`)

func assetsCrd0000_03_configOperator_01_proxyCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_configOperator_01_proxyCrdYaml, nil
}

func assetsCrd0000_03_configOperator_01_proxyCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_configOperator_01_proxyCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_config-operator_01_proxy.crd.yaml", size: 4790, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: clusterresourcequotas.quota.openshift.io
spec:
  group: quota.openshift.io
  names:
    kind: ClusterResourceQuota
    listKind: ClusterResourceQuotaList
    plural: clusterresourcequotas
    singular: clusterresourcequota
  scope: Cluster
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: "ClusterResourceQuota mirrors ResourceQuota at a cluster scope.  This object is easily convertible to synthetic ResourceQuota object to allow quota evaluation re-use. \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          required:
            - metadata
            - spec
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec defines the desired quota
              type: object
              required:
                - quota
                - selector
              properties:
                quota:
                  description: Quota defines the desired quota
                  type: object
                  properties:
                    hard:
                      description: 'hard is the set of desired hard limits for each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                      type: object
                      additionalProperties:
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        anyOf:
                          - type: integer
                          - type: string
                        x-kubernetes-int-or-string: true
                    scopeSelector:
                      description: scopeSelector is also a collection of filters like scopes that must match each object tracked by a quota but expressed using ScopeSelectorOperator in combination with possible values. For a resource to match, both scopes AND scopeSelector (if specified in spec), must be matched.
                      type: object
                      properties:
                        matchExpressions:
                          description: A list of scope selector requirements by scope of the resources.
                          type: array
                          items:
                            description: A scoped-resource selector requirement is a selector that contains values, a scope name, and an operator that relates the scope name and values.
                            type: object
                            required:
                              - operator
                              - scopeName
                            properties:
                              operator:
                                description: Represents a scope's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist.
                                type: string
                              scopeName:
                                description: The name of the scope that the selector applies to.
                                type: string
                              values:
                                description: An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                type: array
                                items:
                                  type: string
                    scopes:
                      description: A collection of filters that must match each object tracked by a quota. If not specified, the quota matches all objects.
                      type: array
                      items:
                        description: A ResourceQuotaScope defines a filter that must match each object tracked by a quota
                        type: string
                selector:
                  description: Selector is the selector used to match projects. It should only select active projects on the scale of dozens (though it can select many more less active projects).  These projects will contend on object creation through this resource.
                  type: object
                  properties:
                    annotations:
                      description: AnnotationSelector is used to select projects by annotation.
                      type: object
                      additionalProperties:
                        type: string
                      nullable: true
                    labels:
                      description: LabelSelector is used to select projects by label.
                      type: object
                      properties:
                        matchExpressions:
                          description: matchExpressions is a list of label selector requirements. The requirements are ANDed.
                          type: array
                          items:
                            description: A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
                            type: object
                            required:
                              - key
                              - operator
                            properties:
                              key:
                                description: key is the label key that the selector applies to.
                                type: string
                              operator:
                                description: operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.
                                type: string
                              values:
                                description: values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
                                type: array
                                items:
                                  type: string
                        matchLabels:
                          description: matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.
                          type: object
                          additionalProperties:
                            type: string
                      nullable: true
            status:
              description: Status defines the actual enforced quota and its current usage
              type: object
              required:
                - total
              properties:
                namespaces:
                  description: Namespaces slices the usage by project.  This division allows for quick resolution of deletion reconciliation inside of a single project without requiring a recalculation across all projects.  This can be used to pull the deltas for a given project.
                  type: array
                  items:
                    description: ResourceQuotaStatusByNamespace gives status for a particular project
                    type: object
                    required:
                      - namespace
                      - status
                    properties:
                      namespace:
                        description: Namespace the project this status applies to
                        type: string
                      status:
                        description: Status indicates how many resources have been consumed by this project
                        type: object
                        properties:
                          hard:
                            description: 'Hard is the set of enforced hard limits for each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                            type: object
                            additionalProperties:
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              anyOf:
                                - type: integer
                                - type: string
                              x-kubernetes-int-or-string: true
                          used:
                            description: Used is the current observed total usage of the resource in the namespace.
                            type: object
                            additionalProperties:
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              anyOf:
                                - type: integer
                                - type: string
                              x-kubernetes-int-or-string: true
                  nullable: true
                total:
                  description: Total defines the actual enforced quota and its current usage across all projects
                  type: object
                  properties:
                    hard:
                      description: 'Hard is the set of enforced hard limits for each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                      type: object
                      additionalProperties:
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        anyOf:
                          - type: integer
                          - type: string
                        x-kubernetes-int-or-string: true
                    used:
                      description: Used is the current observed total usage of the resource in the namespace.
                      type: object
                      additionalProperties:
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        anyOf:
                          - type: integer
                          - type: string
                        x-kubernetes-int-or-string: true
      served: true
      storage: true
      subresources:
        status: {}
`)

func assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml, nil
}

func assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml", size: 11773, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_securityOpenshift_01_sccCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: securitycontextconstraints.security.openshift.io
spec:
  group: security.openshift.io
  names:
    kind: SecurityContextConstraints
    listKind: SecurityContextConstraintsList
    plural: securitycontextconstraints
    singular: securitycontextconstraints
  scope: Cluster
  versions:
    - additionalPrinterColumns:
        - description: Determines if a container can request to be run as privileged
          jsonPath: .allowPrivilegedContainer
          name: Priv
          type: string
        - description: A list of capabilities that can be requested to add to the container
          jsonPath: .allowedCapabilities
          name: Caps
          type: string
        - description: Strategy that will dictate what labels will be set in the SecurityContext
          jsonPath: .seLinuxContext.type
          name: SELinux
          type: string
        - description: Strategy that will dictate what RunAsUser is used in the SecurityContext
          jsonPath: .runAsUser.type
          name: RunAsUser
          type: string
        - description: Strategy that will dictate what fs group is used by the SecurityContext
          jsonPath: .fsGroup.type
          name: FSGroup
          type: string
        - description: Strategy that will dictate what supplemental groups are used by the SecurityContext
          jsonPath: .supplementalGroups.type
          name: SupGroup
          type: string
        - description: Sort order of SCCs
          jsonPath: .priority
          name: Priority
          type: string
        - description: Force containers to run with a read only root file system
          jsonPath: .readOnlyRootFilesystem
          name: ReadOnlyRootFS
          type: string
        - description: White list of allowed volume plugins
          jsonPath: .volumes
          name: Volumes
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: "SecurityContextConstraints governs the ability to make requests that affect the SecurityContext that will be applied to a container. For historical reasons SCC was exposed under the core Kubernetes API group. That exposure is deprecated and will be removed in a future release - users should instead use the security.openshift.io group to manage SecurityContextConstraints. \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          required:
            - allowHostDirVolumePlugin
            - allowHostIPC
            - allowHostNetwork
            - allowHostPID
            - allowHostPorts
            - allowPrivilegedContainer
            - allowedCapabilities
            - defaultAddCapabilities
            - priority
            - readOnlyRootFilesystem
            - requiredDropCapabilities
            - volumes
          properties:
            allowHostDirVolumePlugin:
              description: AllowHostDirVolumePlugin determines if the policy allow containers to use the HostDir volume plugin
              type: boolean
            allowHostIPC:
              description: AllowHostIPC determines if the policy allows host ipc in the containers.
              type: boolean
            allowHostNetwork:
              description: AllowHostNetwork determines if the policy allows the use of HostNetwork in the pod spec.
              type: boolean
            allowHostPID:
              description: AllowHostPID determines if the policy allows host pid in the containers.
              type: boolean
            allowHostPorts:
              description: AllowHostPorts determines if the policy allows host ports in the containers.
              type: boolean
            allowPrivilegeEscalation:
              description: AllowPrivilegeEscalation determines if a pod can request to allow privilege escalation. If unspecified, defaults to true.
              type: boolean
              nullable: true
            allowPrivilegedContainer:
              description: AllowPrivilegedContainer determines if a container can request to be run as privileged.
              type: boolean
            allowedCapabilities:
              description: AllowedCapabilities is a list of capabilities that can be requested to add to the container. Capabilities in this field maybe added at the pod author's discretion. You must not list a capability in both AllowedCapabilities and RequiredDropCapabilities. To allow all capabilities you may use '*'.
              type: array
              items:
                description: Capability represent POSIX capabilities type
                type: string
              nullable: true
            allowedFlexVolumes:
              description: AllowedFlexVolumes is a whitelist of allowed Flexvolumes.  Empty or nil indicates that all Flexvolumes may be used.  This parameter is effective only when the usage of the Flexvolumes is allowed in the "Volumes" field.
              type: array
              items:
                description: AllowedFlexVolume represents a single Flexvolume that is allowed to be used.
                type: object
                required:
                  - driver
                properties:
                  driver:
                    description: Driver is the name of the Flexvolume driver.
                    type: string
              nullable: true
            allowedUnsafeSysctls:
              description: "AllowedUnsafeSysctls is a list of explicitly allowed unsafe sysctls, defaults to none. Each entry is either a plain sysctl name or ends in \"*\" in which case it is considered as a prefix of allowed sysctls. Single * means all unsafe sysctls are allowed. Kubelet has to whitelist all allowed unsafe sysctls explicitly to avoid rejection. \n Examples: e.g. \"foo/*\" allows \"foo/bar\", \"foo/baz\", etc. e.g. \"foo.*\" allows \"foo.bar\", \"foo.baz\", etc."
              type: array
              items:
                type: string
              nullable: true
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            defaultAddCapabilities:
              description: DefaultAddCapabilities is the default set of capabilities that will be added to the container unless the pod spec specifically drops the capability.  You may not list a capabiility in both DefaultAddCapabilities and RequiredDropCapabilities.
              type: array
              items:
                description: Capability represent POSIX capabilities type
                type: string
              nullable: true
            defaultAllowPrivilegeEscalation:
              description: DefaultAllowPrivilegeEscalation controls the default setting for whether a process can gain more privileges than its parent process.
              type: boolean
              nullable: true
            forbiddenSysctls:
              description: "ForbiddenSysctls is a list of explicitly forbidden sysctls, defaults to none. Each entry is either a plain sysctl name or ends in \"*\" in which case it is considered as a prefix of forbidden sysctls. Single * means all sysctls are forbidden. \n Examples: e.g. \"foo/*\" forbids \"foo/bar\", \"foo/baz\", etc. e.g. \"foo.*\" forbids \"foo.bar\", \"foo.baz\", etc."
              type: array
              items:
                type: string
              nullable: true
            fsGroup:
              description: FSGroup is the strategy that will dictate what fs group is used by the SecurityContext.
              type: object
              properties:
                ranges:
                  description: Ranges are the allowed ranges of fs groups.  If you would like to force a single fs group then supply a single range with the same start and end.
                  type: array
                  items:
                    description: 'IDRange provides a min/max of an allowed range of IDs. TODO: this could be reused for UIDs.'
                    type: object
                    properties:
                      max:
                        description: Max is the end of the range, inclusive.
                        type: integer
                        format: int64
                      min:
                        description: Min is the start of the range, inclusive.
                        type: integer
                        format: int64
                type:
                  description: Type is the strategy that will dictate what FSGroup is used in the SecurityContext.
                  type: string
              nullable: true
            groups:
              description: The groups that have permission to use this security context constraints
              type: array
              items:
                type: string
              nullable: true
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            priority:
              description: Priority influences the sort order of SCCs when evaluating which SCCs to try first for a given pod request based on access in the Users and Groups fields.  The higher the int, the higher priority. An unset value is considered a 0 priority. If scores for multiple SCCs are equal they will be sorted from most restrictive to least restrictive. If both priorities and restrictions are equal the SCCs will be sorted by name.
              type: integer
              format: int32
              nullable: true
            readOnlyRootFilesystem:
              description: ReadOnlyRootFilesystem when set to true will force containers to run with a read only root file system.  If the container specifically requests to run with a non-read only root file system the SCC should deny the pod. If set to false the container may run with a read only root file system if it wishes but it will not be forced to.
              type: boolean
            requiredDropCapabilities:
              description: RequiredDropCapabilities are the capabilities that will be dropped from the container.  These are required to be dropped and cannot be added.
              type: array
              items:
                description: Capability represent POSIX capabilities type
                type: string
              nullable: true
            runAsUser:
              description: RunAsUser is the strategy that will dictate what RunAsUser is used in the SecurityContext.
              type: object
              properties:
                type:
                  description: Type is the strategy that will dictate what RunAsUser is used in the SecurityContext.
                  type: string
                uid:
                  description: UID is the user id that containers must run as.  Required for the MustRunAs strategy if not using namespace/service account allocated uids.
                  type: integer
                  format: int64
                uidRangeMax:
                  description: UIDRangeMax defines the max value for a strategy that allocates by range.
                  type: integer
                  format: int64
                uidRangeMin:
                  description: UIDRangeMin defines the min value for a strategy that allocates by range.
                  type: integer
                  format: int64
              nullable: true
            seLinuxContext:
              description: SELinuxContext is the strategy that will dictate what labels will be set in the SecurityContext.
              type: object
              properties:
                seLinuxOptions:
                  description: seLinuxOptions required to run as; required for MustRunAs
                  type: object
                  properties:
                    level:
                      description: Level is SELinux level label that applies to the container.
                      type: string
                    role:
                      description: Role is a SELinux role label that applies to the container.
                      type: string
                    type:
                      description: Type is a SELinux type label that applies to the container.
                      type: string
                    user:
                      description: User is a SELinux user label that applies to the container.
                      type: string
                type:
                  description: Type is the strategy that will dictate what SELinux context is used in the SecurityContext.
                  type: string
              nullable: true
            seccompProfiles:
              description: "SeccompProfiles lists the allowed profiles that may be set for the pod or container's seccomp annotations.  An unset (nil) or empty value means that no profiles may be specifid by the pod or container.\tThe wildcard '*' may be used to allow all profiles.  When used to generate a value for a pod the first non-wildcard profile will be used as the default."
              type: array
              items:
                type: string
              nullable: true
            supplementalGroups:
              description: SupplementalGroups is the strategy that will dictate what supplemental groups are used by the SecurityContext.
              type: object
              properties:
                ranges:
                  description: Ranges are the allowed ranges of supplemental groups.  If you would like to force a single supplemental group then supply a single range with the same start and end.
                  type: array
                  items:
                    description: 'IDRange provides a min/max of an allowed range of IDs. TODO: this could be reused for UIDs.'
                    type: object
                    properties:
                      max:
                        description: Max is the end of the range, inclusive.
                        type: integer
                        format: int64
                      min:
                        description: Min is the start of the range, inclusive.
                        type: integer
                        format: int64
                type:
                  description: Type is the strategy that will dictate what supplemental groups is used in the SecurityContext.
                  type: string
              nullable: true
            users:
              description: The users who have permissions to use this security context constraints
              type: array
              items:
                type: string
              nullable: true
            volumes:
              description: Volumes is a white list of allowed volume plugins.  FSType corresponds directly with the field names of a VolumeSource (azureFile, configMap, emptyDir).  To allow all volumes you may use "*". To allow no volumes, set to ["none"].
              type: array
              items:
                description: FS Type gives strong typing to different file systems that are used by volumes.
                type: string
              nullable: true
      served: true
      storage: true
`)

func assetsCrd0000_03_securityOpenshift_01_sccCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, nil
}

func assetsCrd0000_03_securityOpenshift_01_sccCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_securityOpenshift_01_sccCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_security-openshift_01_scc.crd.yaml", size: 16010, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_buildCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: builds.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Build
    listKind: BuildList
    plural: builds
    singular: build
  preserveUnknownFields: false
  scope: Cluster
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: "Build configures the behavior of OpenShift builds for the entire cluster. This includes default settings that can be overridden in BuildConfig objects, and overrides which are applied to all builds. \n The canonical name is \"cluster\" \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          required:
            - spec
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: Spec holds user-settable values for the build controller configuration
              type: object
              properties:
                additionalTrustedCA:
                  description: "AdditionalTrustedCA is a reference to a ConfigMap containing additional CAs that should be trusted for image pushes and pulls during builds. The namespace for this config map is openshift-config. \n DEPRECATED: Additional CAs for image pull and push should be set on image.config.openshift.io/cluster instead."
                  type: object
                  required:
                    - name
                  properties:
                    name:
                      description: name is the metadata.name of the referenced config map
                      type: string
                buildDefaults:
                  description: BuildDefaults controls the default information for Builds
                  type: object
                  properties:
                    defaultProxy:
                      description: "DefaultProxy contains the default proxy settings for all build operations, including image pull/push and source download. \n Values can be overrode by setting the ` + "`" + `HTTP_PROXY` + "`" + `, ` + "`" + `HTTPS_PROXY` + "`" + `, and ` + "`" + `NO_PROXY` + "`" + ` environment variables in the build config's strategy."
                      type: object
                      properties:
                        httpProxy:
                          description: httpProxy is the URL of the proxy for HTTP requests.  Empty means unset and will not result in an env var.
                          type: string
                        httpsProxy:
                          description: httpsProxy is the URL of the proxy for HTTPS requests.  Empty means unset and will not result in an env var.
                          type: string
                        noProxy:
                          description: noProxy is a comma-separated list of hostnames and/or CIDRs and/or IPs for which the proxy should not be used. Empty means unset and will not result in an env var.
                          type: string
                        readinessEndpoints:
                          description: readinessEndpoints is a list of endpoints used to verify readiness of the proxy.
                          type: array
                          items:
                            type: string
                        trustedCA:
                          description: "trustedCA is a reference to a ConfigMap containing a CA certificate bundle. The trustedCA field should only be consumed by a proxy validator. The validator is responsible for reading the certificate bundle from the required key \"ca-bundle.crt\", merging it with the system default trust bundle, and writing the merged trust bundle to a ConfigMap named \"trusted-ca-bundle\" in the \"openshift-config-managed\" namespace. Clients that expect to make proxy connections must use the trusted-ca-bundle for all HTTPS requests to the proxy, and may use the trusted-ca-bundle for non-proxy HTTPS requests as well. \n The namespace for the ConfigMap referenced by trustedCA is \"openshift-config\". Here is an example ConfigMap (in yaml): \n apiVersion: v1 kind: ConfigMap metadata:  name: user-ca-bundle  namespace: openshift-config  data:    ca-bundle.crt: |      -----BEGIN CERTIFICATE-----      Custom CA certificate bundle.      -----END CERTIFICATE-----"
                          type: object
                          required:
                            - name
                          properties:
                            name:
                              description: name is the metadata.name of the referenced config map
                              type: string
                    env:
                      description: Env is a set of default environment variables that will be applied to the build if the specified variables do not exist on the build
                      type: array
                      items:
                        description: EnvVar represents an environment variable present in a Container.
                        type: object
                        required:
                          - name
                        properties:
                          name:
                            description: Name of the environment variable. Must be a C_IDENTIFIER.
                            type: string
                          value:
                            description: 'Variable references $(VAR_NAME) are expanded using the previously defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "".'
                            type: string
                          valueFrom:
                            description: Source for the environment variable's value. Cannot be used if value is not empty.
                            type: object
                            properties:
                              configMapKeyRef:
                                description: Selects a key of a ConfigMap.
                                type: object
                                required:
                                  - key
                                properties:
                                  key:
                                    description: The key to select.
                                    type: string
                                  name:
                                    description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                    type: string
                                  optional:
                                    description: Specify whether the ConfigMap or its key must be defined
                                    type: boolean
                              fieldRef:
                                description: 'Selects a field of the pod: supports metadata.name, metadata.namespace, ` + "`" + `metadata.labels[''<KEY>'']` + "`" + `, ` + "`" + `metadata.annotations[''<KEY>'']` + "`" + `, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.'
                                type: object
                                required:
                                  - fieldPath
                                properties:
                                  apiVersion:
                                    description: Version of the schema the FieldPath is written in terms of, defaults to "v1".
                                    type: string
                                  fieldPath:
                                    description: Path of the field to select in the specified API version.
                                    type: string
                              resourceFieldRef:
                                description: 'Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.'
                                type: object
                                required:
                                  - resource
                                properties:
                                  containerName:
                                    description: 'Container name: required for volumes, optional for env vars'
                                    type: string
                                  divisor:
                                    description: Specifies the output format of the exposed resources, defaults to "1"
                                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                    anyOf:
                                      - type: integer
                                      - type: string
                                    x-kubernetes-int-or-string: true
                                  resource:
                                    description: 'Required: resource to select'
                                    type: string
                              secretKeyRef:
                                description: Selects a key of a secret in the pod's namespace
                                type: object
                                required:
                                  - key
                                properties:
                                  key:
                                    description: The key of the secret to select from.  Must be a valid secret key.
                                    type: string
                                  name:
                                    description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?'
                                    type: string
                                  optional:
                                    description: Specify whether the Secret or its key must be defined
                                    type: boolean
                    gitProxy:
                      description: "GitProxy contains the proxy settings for git operations only. If set, this will override any Proxy settings for all git commands, such as git clone. \n Values that are not set here will be inherited from DefaultProxy."
                      type: object
                      properties:
                        httpProxy:
                          description: httpProxy is the URL of the proxy for HTTP requests.  Empty means unset and will not result in an env var.
                          type: string
                        httpsProxy:
                          description: httpsProxy is the URL of the proxy for HTTPS requests.  Empty means unset and will not result in an env var.
                          type: string
                        noProxy:
                          description: noProxy is a comma-separated list of hostnames and/or CIDRs and/or IPs for which the proxy should not be used. Empty means unset and will not result in an env var.
                          type: string
                        readinessEndpoints:
                          description: readinessEndpoints is a list of endpoints used to verify readiness of the proxy.
                          type: array
                          items:
                            type: string
                        trustedCA:
                          description: "trustedCA is a reference to a ConfigMap containing a CA certificate bundle. The trustedCA field should only be consumed by a proxy validator. The validator is responsible for reading the certificate bundle from the required key \"ca-bundle.crt\", merging it with the system default trust bundle, and writing the merged trust bundle to a ConfigMap named \"trusted-ca-bundle\" in the \"openshift-config-managed\" namespace. Clients that expect to make proxy connections must use the trusted-ca-bundle for all HTTPS requests to the proxy, and may use the trusted-ca-bundle for non-proxy HTTPS requests as well. \n The namespace for the ConfigMap referenced by trustedCA is \"openshift-config\". Here is an example ConfigMap (in yaml): \n apiVersion: v1 kind: ConfigMap metadata:  name: user-ca-bundle  namespace: openshift-config  data:    ca-bundle.crt: |      -----BEGIN CERTIFICATE-----      Custom CA certificate bundle.      -----END CERTIFICATE-----"
                          type: object
                          required:
                            - name
                          properties:
                            name:
                              description: name is the metadata.name of the referenced config map
                              type: string
                    imageLabels:
                      description: ImageLabels is a list of docker labels that are applied to the resulting image. User can override a default label by providing a label with the same name in their Build/BuildConfig.
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            description: Name defines the name of the label. It must have non-zero length.
                            type: string
                          value:
                            description: Value defines the literal value of the label.
                            type: string
                    resources:
                      description: Resources defines resource requirements to execute the build.
                      type: object
                      properties:
                        limits:
                          description: 'Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                          additionalProperties:
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            anyOf:
                              - type: integer
                              - type: string
                            x-kubernetes-int-or-string: true
                        requests:
                          description: 'Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                          additionalProperties:
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            anyOf:
                              - type: integer
                              - type: string
                            x-kubernetes-int-or-string: true
                buildOverrides:
                  description: BuildOverrides controls override settings for builds
                  type: object
                  properties:
                    forcePull:
                      description: ForcePull overrides, if set, the equivalent value in the builds, i.e. false disables force pull for all builds, true enables force pull for all builds, independently of what each build specifies itself
                      type: boolean
                    imageLabels:
                      description: ImageLabels is a list of docker labels that are applied to the resulting image. If user provided a label in their Build/BuildConfig with the same name as one in this list, the user's label will be overwritten.
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            description: Name defines the name of the label. It must have non-zero length.
                            type: string
                          value:
                            description: Value defines the literal value of the label.
                            type: string
                    nodeSelector:
                      description: NodeSelector is a selector which must be true for the build pod to fit on a node
                      type: object
                      additionalProperties:
                        type: string
                    tolerations:
                      description: Tolerations is a list of Tolerations that will override any existing tolerations set on a build pod.
                      type: array
                      items:
                        description: The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.
                        type: object
                        properties:
                          effect:
                            description: Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
                            type: string
                          key:
                            description: Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
                            type: string
                          operator:
                            description: Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
                            type: string
                          tolerationSeconds:
                            description: TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
                            type: integer
                            format: int64
                          value:
                            description: Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
                            type: string
      served: true
      storage: true
      subresources:
        status: {}
`)

func assetsCrd0000_10_configOperator_01_buildCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_buildCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_buildCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_buildCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_build.crd.yaml", size: 20246, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_featuregateCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: featuregates.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: FeatureGate
    listKind: FeatureGateList
    plural: featuregates
    singular: featuregate
  scope: Cluster
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: "Feature holds cluster-wide information about feature gates.  The canonical name is ` + "`" + `cluster` + "`" + ` \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          required:
            - spec
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: spec holds user settable values for configuration
              type: object
              properties:
                customNoUpgrade:
                  description: customNoUpgrade allows the enabling or disabling of any feature. Turning this feature set on IS NOT SUPPORTED, CANNOT BE UNDONE, and PREVENTS UPGRADES. Because of its nature, this setting cannot be validated.  If you have any typos or accidentally apply invalid combinations your cluster may fail in an unrecoverable way.  featureSet must equal "CustomNoUpgrade" must be set to use this field.
                  type: object
                  properties:
                    disabled:
                      description: disabled is a list of all feature gates that you want to force off
                      type: array
                      items:
                        type: string
                    enabled:
                      description: enabled is a list of all feature gates that you want to force on
                      type: array
                      items:
                        type: string
                  nullable: true
                featureSet:
                  description: featureSet changes the list of features in the cluster.  The default is empty.  Be very careful adjusting this setting. Turning on or off features may cause irreversible changes in your cluster which cannot be undone.
                  type: string
            status:
              description: status holds observed values from the cluster. They may not be overridden.
              type: object
      served: true
      storage: true
      subresources:
        status: {}
`)

func assetsCrd0000_10_configOperator_01_featuregateCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_featuregateCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_featuregateCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_featuregateCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_featuregate.crd.yaml", size: 3438, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_imageCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: images.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Image
    listKind: ImageList
    plural: images
    singular: image
  scope: Cluster
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          description: "Image governs policies related to imagestream imports and runtime configuration for external registries. It allows cluster admins to configure which registries OpenShift is allowed to import images from, extra CA trust bundles for external registries, and policies to block or allow registry hostnames. When exposing OpenShift's image registry to the public, this also lets cluster admins specify the external hostname. \n Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer)."
          type: object
          required:
            - spec
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: spec holds user settable values for configuration
              type: object
              properties:
                additionalTrustedCA:
                  description: additionalTrustedCA is a reference to a ConfigMap containing additional CAs that should be trusted during imagestream import, pod image pull, build image pull, and imageregistry pullthrough. The namespace for this config map is openshift-config.
                  type: object
                  required:
                    - name
                  properties:
                    name:
                      description: name is the metadata.name of the referenced config map
                      type: string
                allowedRegistriesForImport:
                  description: allowedRegistriesForImport limits the container image registries that normal users may import images from. Set this list to the registries that you trust to contain valid Docker images and that you want applications to be able to import from. Users with permission to create Images or ImageStreamMappings via the API are not affected by this policy - typically only administrators or system integrations will have those permissions.
                  type: array
                  items:
                    description: RegistryLocation contains a location of the registry specified by the registry domain name. The domain name might include wildcards, like '*' or '??'.
                    type: object
                    properties:
                      domainName:
                        description: domainName specifies a domain name for the registry In case the registry use non-standard (80 or 443) port, the port should be included in the domain name as well.
                        type: string
                      insecure:
                        description: insecure indicates whether the registry is secure (https) or insecure (http) By default (if not specified) the registry is assumed as secure.
                        type: boolean
                externalRegistryHostnames:
                  description: externalRegistryHostnames provides the hostnames for the default external image registry. The external hostname should be set only when the image registry is exposed externally. The first value is used in 'publicDockerImageRepository' field in ImageStreams. The value must be in "hostname[:port]" format.
                  type: array
                  items:
                    type: string
                registrySources:
                  description: registrySources contains configuration that determines how the container runtime should treat individual registries when accessing images for builds+pods. (e.g. whether or not to allow insecure access).  It does not contain configuration for the internal cluster registry.
                  type: object
                  properties:
                    allowedRegistries:
                      description: "allowedRegistries are the only registries permitted for image pull and push actions. All other registries are denied. \n Only one of BlockedRegistries or AllowedRegistries may be set."
                      type: array
                      items:
                        type: string
                    blockedRegistries:
                      description: "blockedRegistries cannot be used for image pull and push actions. All other registries are permitted. \n Only one of BlockedRegistries or AllowedRegistries may be set."
                      type: array
                      items:
                        type: string
                    containerRuntimeSearchRegistries:
                      description: 'containerRuntimeSearchRegistries are registries that will be searched when pulling images that do not have fully qualified domains in their pull specs. Registries will be searched in the order provided in the list. Note: this search list only works with the container runtime, i.e CRI-O. Will NOT work with builds or imagestream imports.'
                      type: array
                      format: hostname
                      minItems: 1
                      items:
                        type: string
                      x-kubernetes-list-type: set
                    insecureRegistries:
                      description: insecureRegistries are registries which do not have a valid TLS certificates or only support HTTP connections.
                      type: array
                      items:
                        type: string
            status:
              description: status holds observed values from the cluster. They may not be overridden.
              type: object
              properties:
                externalRegistryHostnames:
                  description: externalRegistryHostnames provides the hostnames for the default external image registry. The external hostname should be set only when the image registry is exposed externally. The first value is used in 'publicDockerImageRepository' field in ImageStreams. The value must be in "hostname[:port]" format.
                  type: array
                  items:
                    type: string
                internalRegistryHostname:
                  description: internalRegistryHostname sets the hostname for the default internal image registry. The value must be in "hostname[:port]" format. This value is set by the image registry operator which controls the internal registry hostname. For backward compatibility, users can still use OPENSHIFT_DEFAULT_REGISTRY environment variable but this setting overrides the environment variable.
                  type: string
      served: true
      storage: true
      subresources:
        status: {}
`)

func assetsCrd0000_10_configOperator_01_imageCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_imageCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_imageCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_imageCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_image.crd.yaml", size: 7808, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/470
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: imagecontentsourcepolicies.operator.openshift.io
spec:
  group: operator.openshift.io
  names:
    kind: ImageContentSourcePolicy
    listKind: ImageContentSourcePolicyList
    plural: imagecontentsourcepolicies
    singular: imagecontentsourcepolicy
  scope: Cluster
  versions:
    - name: v1alpha1
      schema:
        openAPIV3Schema:
          description: "ImageContentSourcePolicy holds cluster-wide information about how to handle registry mirror rules. When multiple policies are defined, the outcome of the behavior is defined on each field. \n Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support."
          type: object
          required:
            - spec
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: spec holds user settable values for configuration
              type: object
              properties:
                repositoryDigestMirrors:
                  description: "repositoryDigestMirrors allows images referenced by image digests in pods to be pulled from alternative mirrored repository locations. The image pull specification provided to the pod will be compared to the source locations described in RepositoryDigestMirrors and the image may be pulled down from any of the mirrors in the list instead of the specified repository allowing administrators to choose a potentially faster mirror. Only image pull specifications that have an image digest will have this behavior applied to them - tags will continue to be pulled from the specified repository in the pull spec. \n Each “source” repository is treated independently; configurations for different “source” repositories don’t interact. \n When multiple policies are defined for the same “source” repository, the sets of defined mirrors will be merged together, preserving the relative order of the mirrors, if possible. For example, if policy A has mirrors ` + "`" + `a, b, c` + "`" + ` and policy B has mirrors ` + "`" + `c, d, e` + "`" + `, the mirrors will be used in the order ` + "`" + `a, b, c, d, e` + "`" + `.  If the orders of mirror entries conflict (e.g. ` + "`" + `a, b` + "`" + ` vs. ` + "`" + `b, a` + "`" + `) the configuration is not rejected but the resulting order is unspecified."
                  type: array
                  items:
                    description: 'RepositoryDigestMirrors holds cluster-wide information about how to handle mirros in the registries config. Note: the mirrors only work when pulling the images that are referenced by their digests.'
                    type: object
                    required:
                      - source
                    properties:
                      mirrors:
                        description: mirrors is one or more repositories that may also contain the same images. The order of mirrors in this list is treated as the user's desired priority, while source is by default considered lower priority than all mirrors. Other cluster configuration, including (but not limited to) other repositoryDigestMirrors objects, may impact the exact order mirrors are contacted in, or some mirrors may be contacted in parallel, so this should be considered a preference rather than a guarantee of ordering.
                        type: array
                        items:
                          type: string
                      source:
                        description: source is the repository that users refer to, e.g. in image pull specifications.
                        type: string
      served: true
      storage: true
      subresources:
        status: {}
`)

func assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml", size: 4754, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_11_imageregistryConfigsCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/519
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: configs.imageregistry.operator.openshift.io
spec:
  group: imageregistry.operator.openshift.io
  names:
    kind: Config
    listKind: ConfigList
    plural: configs
    singular: config
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        description: Config is the configuration object for a registry instance managed
          by the registry operator
        type: object
        required:
        - metadata
        - spec
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ImageRegistrySpec defines the specs for the running registry.
            type: object
            required:
            - managementState
            - replicas
            properties:
              affinity:
                description: affinity is a group of node affinity scheduling rules
                  for the image registry pod(s).
                type: object
                properties:
                  nodeAffinity:
                    description: Describes node affinity scheduling rules for the
                      pod.
                    type: object
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: The scheduler will prefer to schedule pods to
                          nodes that satisfy the affinity expressions specified by
                          this field, but it may choose a node that violates one or
                          more of the expressions. The node that is most preferred
                          is the one with the greatest sum of weights, i.e. for each
                          node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling affinity expressions,
                          etc.), compute a sum by iterating through the elements of
                          this field and adding "weight" to the sum if the node matches
                          the corresponding matchExpressions; the node(s) with the
                          highest sum are the most preferred.
                        type: array
                        items:
                          description: An empty preferred scheduling term matches
                            all objects with implicit weight 0 (i.e. it's a no-op).
                            A null preferred scheduling term matches no objects (i.e.
                            is also a no-op).
                          type: object
                          required:
                          - preference
                          - weight
                          properties:
                            preference:
                              description: A node selector term, associated with the
                                corresponding weight.
                              type: object
                              properties:
                                matchExpressions:
                                  description: A list of node selector requirements
                                    by node's labels.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchFields:
                                  description: A list of node selector requirements
                                    by node's fields.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                            weight:
                              description: Weight associated with matching the corresponding
                                nodeSelectorTerm, in the range 1-100.
                              type: integer
                              format: int32
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: If the affinity requirements specified by this
                          field are not met at scheduling time, the pod will not be
                          scheduled onto the node. If the affinity requirements specified
                          by this field cease to be met at some point during pod execution
                          (e.g. due to an update), the system may or may not try to
                          eventually evict the pod from its node.
                        type: object
                        required:
                        - nodeSelectorTerms
                        properties:
                          nodeSelectorTerms:
                            description: Required. A list of node selector terms.
                              The terms are ORed.
                            type: array
                            items:
                              description: A null or empty node selector term matches
                                no objects. The requirements of them are ANDed. The
                                TopologySelectorTerm type implements a subset of the
                                NodeSelectorTerm.
                              type: object
                              properties:
                                matchExpressions:
                                  description: A list of node selector requirements
                                    by node's labels.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchFields:
                                  description: A list of node selector requirements
                                    by node's fields.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                  podAffinity:
                    description: Describes pod affinity scheduling rules (e.g. co-locate
                      this pod in the same node, zone, etc. as some other pod(s)).
                    type: object
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: The scheduler will prefer to schedule pods to
                          nodes that satisfy the affinity expressions specified by
                          this field, but it may choose a node that violates one or
                          more of the expressions. The node that is most preferred
                          is the one with the greatest sum of weights, i.e. for each
                          node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling affinity expressions,
                          etc.), compute a sum by iterating through the elements of
                          this field and adding "weight" to the sum if the node has
                          pods which matches the corresponding podAffinityTerm; the
                          node(s) with the highest sum are the most preferred.
                        type: array
                        items:
                          description: The weights of all of the matched WeightedPodAffinityTerm
                            fields are added per-node to find the most preferred node(s)
                          type: object
                          required:
                          - podAffinityTerm
                          - weight
                          properties:
                            podAffinityTerm:
                              description: Required. A pod affinity term, associated
                                with the corresponding weight.
                              type: object
                              required:
                              - topologyKey
                              properties:
                                labelSelector:
                                  description: A label query over a set of resources,
                                    in this case pods.
                                  type: object
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      type: array
                                      items:
                                        description: A label selector requirement
                                          is a selector that contains values, a key,
                                          and an operator that relates the key and
                                          values.
                                        type: object
                                        required:
                                        - key
                                        - operator
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's
                                              relationship to a set of values. Valid
                                              operators are In, NotIn, Exists and
                                              DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string
                                              values. If the operator is In or NotIn,
                                              the values array must be non-empty.
                                              If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This
                                              array is replaced during a strategic
                                              merge patch.
                                            type: array
                                            items:
                                              type: string
                                    matchLabels:
                                      description: matchLabels is a map of {key,value}
                                        pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions,
                                        whose key field is "key", the operator is
                                        "In", and the values array contains only "value".
                                        The requirements are ANDed.
                                      type: object
                                      additionalProperties:
                                        type: string
                                namespaceSelector:
                                  description: A label query over the set of namespaces
                                    that the term applies to. The term is applied
                                    to the union of the namespaces selected by this
                                    field and the ones listed in the namespaces field.
                                    null selector and null or empty namespaces list
                                    means "this pod's namespace". An empty selector
                                    ({}) matches all namespaces. This field is alpha-level
                                    and is only honored when PodAffinityNamespaceSelector
                                    feature is enabled.
                                  type: object
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      type: array
                                      items:
                                        description: A label selector requirement
                                          is a selector that contains values, a key,
                                          and an operator that relates the key and
                                          values.
                                        type: object
                                        required:
                                        - key
                                        - operator
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's
                                              relationship to a set of values. Valid
                                              operators are In, NotIn, Exists and
                                              DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string
                                              values. If the operator is In or NotIn,
                                              the values array must be non-empty.
                                              If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This
                                              array is replaced during a strategic
                                              merge patch.
                                            type: array
                                            items:
                                              type: string
                                    matchLabels:
                                      description: matchLabels is a map of {key,value}
                                        pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions,
                                        whose key field is "key", the operator is
                                        "In", and the values array contains only "value".
                                        The requirements are ANDed.
                                      type: object
                                      additionalProperties:
                                        type: string
                                namespaces:
                                  description: namespaces specifies a static list
                                    of namespace names that the term applies to. The
                                    term is applied to the union of the namespaces
                                    listed in this field and the ones selected by
                                    namespaceSelector. null or empty namespaces list
                                    and null namespaceSelector means "this pod's namespace"
                                  type: array
                                  items:
                                    type: string
                                topologyKey:
                                  description: This pod should be co-located (affinity)
                                    or not co-located (anti-affinity) with the pods
                                    matching the labelSelector in the specified namespaces,
                                    where co-located is defined as running on a node
                                    whose value of the label with key topologyKey
                                    matches that of any node on which any of the selected
                                    pods is running. Empty topologyKey is not allowed.
                                  type: string
                            weight:
                              description: weight associated with matching the corresponding
                                podAffinityTerm, in the range 1-100.
                              type: integer
                              format: int32
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: If the affinity requirements specified by this
                          field are not met at scheduling time, the pod will not be
                          scheduled onto the node. If the affinity requirements specified
                          by this field cease to be met at some point during pod execution
                          (e.g. due to a pod label update), the system may or may
                          not try to eventually evict the pod from its node. When
                          there are multiple elements, the lists of nodes corresponding
                          to each podAffinityTerm are intersected, i.e. all terms
                          must be satisfied.
                        type: array
                        items:
                          description: Defines a set of pods (namely those matching
                            the labelSelector relative to the given namespace(s))
                            that this pod should be co-located (affinity) or not co-located
                            (anti-affinity) with, where co-located is defined as running
                            on a node whose value of the label with key <topologyKey>
                            matches that of any node on which a pod of the set of
                            pods is running
                          type: object
                          required:
                          - topologyKey
                          properties:
                            labelSelector:
                              description: A label query over a set of resources,
                                in this case pods.
                              type: object
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  type: array
                                  items:
                                    description: A label selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string
                                          values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the
                                          operator is Exists or DoesNotExist, the
                                          values array must be empty. This array is
                                          replaced during a strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchLabels:
                                  description: matchLabels is a map of {key,value}
                                    pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions,
                                    whose key field is "key", the operator is "In",
                                    and the values array contains only "value". The
                                    requirements are ANDed.
                                  type: object
                                  additionalProperties:
                                    type: string
                            namespaceSelector:
                              description: A label query over the set of namespaces
                                that the term applies to. The term is applied to the
                                union of the namespaces selected by this field and
                                the ones listed in the namespaces field. null selector
                                and null or empty namespaces list means "this pod's
                                namespace". An empty selector ({}) matches all namespaces.
                                This field is alpha-level and is only honored when
                                PodAffinityNamespaceSelector feature is enabled.
                              type: object
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  type: array
                                  items:
                                    description: A label selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string
                                          values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the
                                          operator is Exists or DoesNotExist, the
                                          values array must be empty. This array is
                                          replaced during a strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchLabels:
                                  description: matchLabels is a map of {key,value}
                                    pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions,
                                    whose key field is "key", the operator is "In",
                                    and the values array contains only "value". The
                                    requirements are ANDed.
                                  type: object
                                  additionalProperties:
                                    type: string
                            namespaces:
                              description: namespaces specifies a static list of namespace
                                names that the term applies to. The term is applied
                                to the union of the namespaces listed in this field
                                and the ones selected by namespaceSelector. null or
                                empty namespaces list and null namespaceSelector means
                                "this pod's namespace"
                              type: array
                              items:
                                type: string
                            topologyKey:
                              description: This pod should be co-located (affinity)
                                or not co-located (anti-affinity) with the pods matching
                                the labelSelector in the specified namespaces, where
                                co-located is defined as running on a node whose value
                                of the label with key topologyKey matches that of
                                any node on which any of the selected pods is running.
                                Empty topologyKey is not allowed.
                              type: string
                  podAntiAffinity:
                    description: Describes pod anti-affinity scheduling rules (e.g.
                      avoid putting this pod in the same node, zone, etc. as some
                      other pod(s)).
                    type: object
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: The scheduler will prefer to schedule pods to
                          nodes that satisfy the anti-affinity expressions specified
                          by this field, but it may choose a node that violates one
                          or more of the expressions. The node that is most preferred
                          is the one with the greatest sum of weights, i.e. for each
                          node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling anti-affinity expressions,
                          etc.), compute a sum by iterating through the elements of
                          this field and adding "weight" to the sum if the node has
                          pods which matches the corresponding podAffinityTerm; the
                          node(s) with the highest sum are the most preferred.
                        type: array
                        items:
                          description: The weights of all of the matched WeightedPodAffinityTerm
                            fields are added per-node to find the most preferred node(s)
                          type: object
                          required:
                          - podAffinityTerm
                          - weight
                          properties:
                            podAffinityTerm:
                              description: Required. A pod affinity term, associated
                                with the corresponding weight.
                              type: object
                              required:
                              - topologyKey
                              properties:
                                labelSelector:
                                  description: A label query over a set of resources,
                                    in this case pods.
                                  type: object
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      type: array
                                      items:
                                        description: A label selector requirement
                                          is a selector that contains values, a key,
                                          and an operator that relates the key and
                                          values.
                                        type: object
                                        required:
                                        - key
                                        - operator
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's
                                              relationship to a set of values. Valid
                                              operators are In, NotIn, Exists and
                                              DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string
                                              values. If the operator is In or NotIn,
                                              the values array must be non-empty.
                                              If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This
                                              array is replaced during a strategic
                                              merge patch.
                                            type: array
                                            items:
                                              type: string
                                    matchLabels:
                                      description: matchLabels is a map of {key,value}
                                        pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions,
                                        whose key field is "key", the operator is
                                        "In", and the values array contains only "value".
                                        The requirements are ANDed.
                                      type: object
                                      additionalProperties:
                                        type: string
                                namespaceSelector:
                                  description: A label query over the set of namespaces
                                    that the term applies to. The term is applied
                                    to the union of the namespaces selected by this
                                    field and the ones listed in the namespaces field.
                                    null selector and null or empty namespaces list
                                    means "this pod's namespace". An empty selector
                                    ({}) matches all namespaces. This field is alpha-level
                                    and is only honored when PodAffinityNamespaceSelector
                                    feature is enabled.
                                  type: object
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      type: array
                                      items:
                                        description: A label selector requirement
                                          is a selector that contains values, a key,
                                          and an operator that relates the key and
                                          values.
                                        type: object
                                        required:
                                        - key
                                        - operator
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's
                                              relationship to a set of values. Valid
                                              operators are In, NotIn, Exists and
                                              DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string
                                              values. If the operator is In or NotIn,
                                              the values array must be non-empty.
                                              If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This
                                              array is replaced during a strategic
                                              merge patch.
                                            type: array
                                            items:
                                              type: string
                                    matchLabels:
                                      description: matchLabels is a map of {key,value}
                                        pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions,
                                        whose key field is "key", the operator is
                                        "In", and the values array contains only "value".
                                        The requirements are ANDed.
                                      type: object
                                      additionalProperties:
                                        type: string
                                namespaces:
                                  description: namespaces specifies a static list
                                    of namespace names that the term applies to. The
                                    term is applied to the union of the namespaces
                                    listed in this field and the ones selected by
                                    namespaceSelector. null or empty namespaces list
                                    and null namespaceSelector means "this pod's namespace"
                                  type: array
                                  items:
                                    type: string
                                topologyKey:
                                  description: This pod should be co-located (affinity)
                                    or not co-located (anti-affinity) with the pods
                                    matching the labelSelector in the specified namespaces,
                                    where co-located is defined as running on a node
                                    whose value of the label with key topologyKey
                                    matches that of any node on which any of the selected
                                    pods is running. Empty topologyKey is not allowed.
                                  type: string
                            weight:
                              description: weight associated with matching the corresponding
                                podAffinityTerm, in the range 1-100.
                              type: integer
                              format: int32
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: If the anti-affinity requirements specified by
                          this field are not met at scheduling time, the pod will
                          not be scheduled onto the node. If the anti-affinity requirements
                          specified by this field cease to be met at some point during
                          pod execution (e.g. due to a pod label update), the system
                          may or may not try to eventually evict the pod from its
                          node. When there are multiple elements, the lists of nodes
                          corresponding to each podAffinityTerm are intersected, i.e.
                          all terms must be satisfied.
                        type: array
                        items:
                          description: Defines a set of pods (namely those matching
                            the labelSelector relative to the given namespace(s))
                            that this pod should be co-located (affinity) or not co-located
                            (anti-affinity) with, where co-located is defined as running
                            on a node whose value of the label with key <topologyKey>
                            matches that of any node on which a pod of the set of
                            pods is running
                          type: object
                          required:
                          - topologyKey
                          properties:
                            labelSelector:
                              description: A label query over a set of resources,
                                in this case pods.
                              type: object
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  type: array
                                  items:
                                    description: A label selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string
                                          values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the
                                          operator is Exists or DoesNotExist, the
                                          values array must be empty. This array is
                                          replaced during a strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchLabels:
                                  description: matchLabels is a map of {key,value}
                                    pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions,
                                    whose key field is "key", the operator is "In",
                                    and the values array contains only "value". The
                                    requirements are ANDed.
                                  type: object
                                  additionalProperties:
                                    type: string
                            namespaceSelector:
                              description: A label query over the set of namespaces
                                that the term applies to. The term is applied to the
                                union of the namespaces selected by this field and
                                the ones listed in the namespaces field. null selector
                                and null or empty namespaces list means "this pod's
                                namespace". An empty selector ({}) matches all namespaces.
                                This field is alpha-level and is only honored when
                                PodAffinityNamespaceSelector feature is enabled.
                              type: object
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  type: array
                                  items:
                                    description: A label selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string
                                          values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the
                                          operator is Exists or DoesNotExist, the
                                          values array must be empty. This array is
                                          replaced during a strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchLabels:
                                  description: matchLabels is a map of {key,value}
                                    pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions,
                                    whose key field is "key", the operator is "In",
                                    and the values array contains only "value". The
                                    requirements are ANDed.
                                  type: object
                                  additionalProperties:
                                    type: string
                            namespaces:
                              description: namespaces specifies a static list of namespace
                                names that the term applies to. The term is applied
                                to the union of the namespaces listed in this field
                                and the ones selected by namespaceSelector. null or
                                empty namespaces list and null namespaceSelector means
                                "this pod's namespace"
                              type: array
                              items:
                                type: string
                            topologyKey:
                              description: This pod should be co-located (affinity)
                                or not co-located (anti-affinity) with the pods matching
                                the labelSelector in the specified namespaces, where
                                co-located is defined as running on a node whose value
                                of the label with key topologyKey matches that of
                                any node on which any of the selected pods is running.
                                Empty topologyKey is not allowed.
                              type: string
              defaultRoute:
                description: defaultRoute indicates whether an external facing route
                  for the registry should be created using the default generated hostname.
                type: boolean
              disableRedirect:
                description: disableRedirect controls whether to route all data through
                  the Registry, rather than redirecting to the backend.
                type: boolean
              httpSecret:
                description: httpSecret is the value needed by the registry to secure
                  uploads, generated by default.
                type: string
              logLevel:
                description: "logLevel is an intent based logging for an overall component.
                  \ It does not give fine grained control, but it is a simple way
                  to manage coarse grained logging choices that operators have to
                  interpret for their operands. \n Valid values are: \"Normal\", \"Debug\",
                  \"Trace\", \"TraceAll\". Defaults to \"Normal\"."
                type: string
                default: Normal
                enum:
                - ""
                - Normal
                - Debug
                - Trace
                - TraceAll
              logging:
                description: logging is deprecated, use logLevel instead.
                type: integer
                format: int64
              managementState:
                description: managementState indicates whether and how the operator
                  should manage the component
                type: string
                pattern: ^(Managed|Unmanaged|Force|Removed)$
              nodeSelector:
                description: nodeSelector defines the node selection constraints for
                  the registry pod.
                type: object
                additionalProperties:
                  type: string
              observedConfig:
                description: observedConfig holds a sparse config that controller
                  has observed from the cluster state.  It exists in spec because
                  it is an input to the level for the operator
                type: object
                nullable: true
                x-kubernetes-preserve-unknown-fields: true
              operatorLogLevel:
                description: "operatorLogLevel is an intent based logging for the
                  operator itself.  It does not give fine grained control, but it
                  is a simple way to manage coarse grained logging choices that operators
                  have to interpret for themselves. \n Valid values are: \"Normal\",
                  \"Debug\", \"Trace\", \"TraceAll\". Defaults to \"Normal\"."
                type: string
                default: Normal
                enum:
                - ""
                - Normal
                - Debug
                - Trace
                - TraceAll
              proxy:
                description: proxy defines the proxy to be used when calling master
                  api, upstream registries, etc.
                type: object
                properties:
                  http:
                    description: http defines the proxy to be used by the image registry
                      when accessing HTTP endpoints.
                    type: string
                  https:
                    description: https defines the proxy to be used by the image registry
                      when accessing HTTPS endpoints.
                    type: string
                  noProxy:
                    description: noProxy defines a comma-separated list of host names
                      that shouldn't go through any proxy.
                    type: string
              readOnly:
                description: readOnly indicates whether the registry instance should
                  reject attempts to push new images or delete existing ones.
                type: boolean
              replicas:
                description: replicas determines the number of registry instances
                  to run.
                type: integer
                format: int32
              requests:
                description: requests controls how many parallel requests a given
                  registry instance will handle before queuing additional requests.
                type: object
                properties:
                  read:
                    description: read defines limits for image registry's reads.
                    type: object
                    properties:
                      maxInQueue:
                        description: maxInQueue sets the maximum queued api requests
                          to the registry.
                        type: integer
                      maxRunning:
                        description: maxRunning sets the maximum in flight api requests
                          to the registry.
                        type: integer
                      maxWaitInQueue:
                        description: maxWaitInQueue sets the maximum time a request
                          can wait in the queue before being rejected.
                        type: string
                        format: duration
                  write:
                    description: write defines limits for image registry's writes.
                    type: object
                    properties:
                      maxInQueue:
                        description: maxInQueue sets the maximum queued api requests
                          to the registry.
                        type: integer
                      maxRunning:
                        description: maxRunning sets the maximum in flight api requests
                          to the registry.
                        type: integer
                      maxWaitInQueue:
                        description: maxWaitInQueue sets the maximum time a request
                          can wait in the queue before being rejected.
                        type: string
                        format: duration
              resources:
                description: resources defines the resource requests+limits for the
                  registry pod.
                type: object
                properties:
                  limits:
                    description: 'Limits describes the maximum amount of compute resources
                      allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                    type: object
                    additionalProperties:
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      anyOf:
                      - type: integer
                      - type: string
                      x-kubernetes-int-or-string: true
                  requests:
                    description: 'Requests describes the minimum amount of compute
                      resources required. If Requests is omitted for a container,
                      it defaults to Limits if that is explicitly specified, otherwise
                      to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                    type: object
                    additionalProperties:
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      anyOf:
                      - type: integer
                      - type: string
                      x-kubernetes-int-or-string: true
              rolloutStrategy:
                description: rolloutStrategy defines rollout strategy for the image
                  registry deployment.
                type: string
                pattern: ^(RollingUpdate|Recreate)$
              routes:
                description: routes defines additional external facing routes which
                  should be created for the registry.
                type: array
                items:
                  description: ImageRegistryConfigRoute holds information on external
                    route access to image registry.
                  type: object
                  required:
                  - name
                  properties:
                    hostname:
                      description: hostname for the route.
                      type: string
                    name:
                      description: name of the route to be created.
                      type: string
                    secretName:
                      description: secretName points to secret containing the certificates
                        to be used by the route.
                      type: string
              storage:
                description: storage details for configuring registry storage, e.g.
                  S3 bucket coordinates.
                type: object
                properties:
                  azure:
                    description: azure represents configuration that uses Azure Blob
                      Storage.
                    type: object
                    properties:
                      accountName:
                        description: accountName defines the account to be used by
                          the registry.
                        type: string
                      cloudName:
                        description: cloudName is the name of the Azure cloud environment
                          to be used by the registry. If empty, the operator will
                          set it based on the infrastructure object.
                        type: string
                      container:
                        description: container defines Azure's container to be used
                          by registry.
                        type: string
                        maxLength: 63
                        minLength: 3
                        pattern: ^[0-9a-z]+(-[0-9a-z]+)*$
                  emptyDir:
                    description: 'emptyDir represents ephemeral storage on the pod''s
                      host node. WARNING: this storage cannot be used with more than
                      1 replica and is not suitable for production use. When the pod
                      is removed from a node for any reason, the data in the emptyDir
                      is deleted forever.'
                    type: object
                  gcs:
                    description: gcs represents configuration that uses Google Cloud
                      Storage.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, buckets are encrypted by default on GCP. This
                          allows for the use of a custom encryption key.
                        type: string
                      projectID:
                        description: projectID is the Project ID of the GCP project
                          that this bucket should be associated with.
                        type: string
                      region:
                        description: region is the GCS location in which your bucket
                          exists. Optional, will be set based on the installed GCS
                          Region.
                        type: string
                  managementState:
                    description: managementState indicates if the operator manages
                      the underlying storage unit. If Managed the operator will remove
                      the storage when this operator gets Removed.
                    type: string
                    pattern: ^(Managed|Unmanaged)$
                  pvc:
                    description: pvc represents configuration that uses a PersistentVolumeClaim.
                    type: object
                    properties:
                      claim:
                        description: claim defines the Persisent Volume Claim's name
                          to be used.
                        type: string
                  s3:
                    description: s3 represents configuration that uses Amazon Simple
                      Storage Service.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      cloudFront:
                        description: cloudFront configures Amazon Cloudfront as the
                          storage middleware in a registry.
                        type: object
                        required:
                        - baseURL
                        - keypairID
                        - privateKey
                        properties:
                          baseURL:
                            description: baseURL contains the SCHEME://HOST[/PATH]
                              at which Cloudfront is served.
                            type: string
                          duration:
                            description: duration is the duration of the Cloudfront
                              session.
                            type: string
                            format: duration
                          keypairID:
                            description: keypairID is key pair ID provided by AWS.
                            type: string
                          privateKey:
                            description: privateKey points to secret containing the
                              private key, provided by AWS.
                            type: object
                            required:
                            - key
                            properties:
                              key:
                                description: The key of the secret to select from.  Must
                                  be a valid secret key.
                                type: string
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                  TODO: Add other useful fields. apiVersion, kind,
                                  uid?'
                                type: string
                              optional:
                                description: Specify whether the Secret or its key
                                  must be defined
                                type: boolean
                      encrypt:
                        description: encrypt specifies whether the registry stores
                          the image in encrypted format or not. Optional, defaults
                          to false.
                        type: boolean
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, Encrypt must be true, or this parameter is ignored.
                        type: string
                      region:
                        description: region is the AWS region in which your bucket
                          exists. Optional, will be set based on the installed AWS
                          Region.
                        type: string
                      regionEndpoint:
                        description: regionEndpoint is the endpoint for S3 compatible
                          storage services. Optional, defaults based on the Region
                          that is provided.
                        type: string
                      virtualHostedStyle:
                        description: virtualHostedStyle enables using S3 virtual hosted
                          style bucket paths with a custom RegionEndpoint Optional,
                          defaults to false.
                        type: boolean
                  swift:
                    description: swift represents configuration that uses OpenStack
                      Object Storage.
                    type: object
                    properties:
                      authURL:
                        description: authURL defines the URL for obtaining an authentication
                          token.
                        type: string
                      authVersion:
                        description: authVersion specifies the OpenStack Auth's version.
                        type: string
                      container:
                        description: container defines the name of Swift container
                          where to store the registry's data.
                        type: string
                      domain:
                        description: domain specifies Openstack's domain name for
                          Identity v3 API.
                        type: string
                      domainID:
                        description: domainID specifies Openstack's domain id for
                          Identity v3 API.
                        type: string
                      regionName:
                        description: regionName defines Openstack's region in which
                          container exists.
                        type: string
                      tenant:
                        description: tenant defines Openstack tenant name to be used
                          by registry.
                        type: string
                      tenantID:
                        description: tenant defines Openstack tenant id to be used
                          by registry.
                        type: string
              tolerations:
                description: tolerations defines the tolerations for the registry
                  pod.
                type: array
                items:
                  description: The pod this Toleration is attached to tolerates any
                    taint that matches the triple <key,value,effect> using the matching
                    operator <operator>.
                  type: object
                  properties:
                    effect:
                      description: Effect indicates the taint effect to match. Empty
                        means match all taint effects. When specified, allowed values
                        are NoSchedule, PreferNoSchedule and NoExecute.
                      type: string
                    key:
                      description: Key is the taint key that the toleration applies
                        to. Empty means match all taint keys. If the key is empty,
                        operator must be Exists; this combination means to match all
                        values and all keys.
                      type: string
                    operator:
                      description: Operator represents a key's relationship to the
                        value. Valid operators are Exists and Equal. Defaults to Equal.
                        Exists is equivalent to wildcard for value, so that a pod
                        can tolerate all taints of a particular category.
                      type: string
                    tolerationSeconds:
                      description: TolerationSeconds represents the period of time
                        the toleration (which must be of effect NoExecute, otherwise
                        this field is ignored) tolerates the taint. By default, it
                        is not set, which means tolerate the taint forever (do not
                        evict). Zero and negative values will be treated as 0 (evict
                        immediately) by the system.
                      type: integer
                      format: int64
                    value:
                      description: Value is the taint value the toleration matches
                        to. If the operator is Exists, the value should be empty,
                        otherwise just a regular string.
                      type: string
              unsupportedConfigOverrides:
                description: 'unsupportedConfigOverrides holds a sparse config that
                  will override any previously set options.  It only needs to be the
                  fields to override it will end up overlaying in the following order:
                  1. hardcoded defaults 2. observedConfig 3. unsupportedConfigOverrides'
                type: object
                nullable: true
                x-kubernetes-preserve-unknown-fields: true
          status:
            description: ImageRegistryStatus reports image registry operational status.
            type: object
            required:
            - storage
            - storageManaged
            properties:
              conditions:
                description: conditions is a list of conditions and their status
                type: array
                items:
                  description: OperatorCondition is just the standard condition fields.
                  type: object
                  properties:
                    lastTransitionTime:
                      type: string
                      format: date-time
                    message:
                      type: string
                    reason:
                      type: string
                    status:
                      type: string
                    type:
                      type: string
              generations:
                description: generations are used to determine when an item needs
                  to be reconciled or has changed in a way that needs a reaction.
                type: array
                items:
                  description: GenerationStatus keeps track of the generation for
                    a given resource so that decisions about forced updates can be
                    made.
                  type: object
                  properties:
                    group:
                      description: group is the group of the thing you're tracking
                      type: string
                    hash:
                      description: hash is an optional field set for resources without
                        generation that are content sensitive like secrets and configmaps
                      type: string
                    lastGeneration:
                      description: lastGeneration is the last generation of the workload
                        controller involved
                      type: integer
                      format: int64
                    name:
                      description: name is the name of the thing you're tracking
                      type: string
                    namespace:
                      description: namespace is where the thing you're tracking is
                      type: string
                    resource:
                      description: resource is the resource type of the thing you're
                        tracking
                      type: string
              observedGeneration:
                description: observedGeneration is the last generation change you've
                  dealt with
                type: integer
                format: int64
              readyReplicas:
                description: readyReplicas indicates how many replicas are ready and
                  at the desired state
                type: integer
                format: int32
              storage:
                description: storage indicates the current applied storage configuration
                  of the registry.
                type: object
                properties:
                  azure:
                    description: azure represents configuration that uses Azure Blob
                      Storage.
                    type: object
                    properties:
                      accountName:
                        description: accountName defines the account to be used by
                          the registry.
                        type: string
                      cloudName:
                        description: cloudName is the name of the Azure cloud environment
                          to be used by the registry. If empty, the operator will
                          set it based on the infrastructure object.
                        type: string
                      container:
                        description: container defines Azure's container to be used
                          by registry.
                        type: string
                        maxLength: 63
                        minLength: 3
                        pattern: ^[0-9a-z]+(-[0-9a-z]+)*$
                  emptyDir:
                    description: 'emptyDir represents ephemeral storage on the pod''s
                      host node. WARNING: this storage cannot be used with more than
                      1 replica and is not suitable for production use. When the pod
                      is removed from a node for any reason, the data in the emptyDir
                      is deleted forever.'
                    type: object
                  gcs:
                    description: gcs represents configuration that uses Google Cloud
                      Storage.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, buckets are encrypted by default on GCP. This
                          allows for the use of a custom encryption key.
                        type: string
                      projectID:
                        description: projectID is the Project ID of the GCP project
                          that this bucket should be associated with.
                        type: string
                      region:
                        description: region is the GCS location in which your bucket
                          exists. Optional, will be set based on the installed GCS
                          Region.
                        type: string
                  managementState:
                    description: managementState indicates if the operator manages
                      the underlying storage unit. If Managed the operator will remove
                      the storage when this operator gets Removed.
                    type: string
                    pattern: ^(Managed|Unmanaged)$
                  pvc:
                    description: pvc represents configuration that uses a PersistentVolumeClaim.
                    type: object
                    properties:
                      claim:
                        description: claim defines the Persisent Volume Claim's name
                          to be used.
                        type: string
                  s3:
                    description: s3 represents configuration that uses Amazon Simple
                      Storage Service.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      cloudFront:
                        description: cloudFront configures Amazon Cloudfront as the
                          storage middleware in a registry.
                        type: object
                        required:
                        - baseURL
                        - keypairID
                        - privateKey
                        properties:
                          baseURL:
                            description: baseURL contains the SCHEME://HOST[/PATH]
                              at which Cloudfront is served.
                            type: string
                          duration:
                            description: duration is the duration of the Cloudfront
                              session.
                            type: string
                            format: duration
                          keypairID:
                            description: keypairID is key pair ID provided by AWS.
                            type: string
                          privateKey:
                            description: privateKey points to secret containing the
                              private key, provided by AWS.
                            type: object
                            required:
                            - key
                            properties:
                              key:
                                description: The key of the secret to select from.  Must
                                  be a valid secret key.
                                type: string
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                  TODO: Add other useful fields. apiVersion, kind,
                                  uid?'
                                type: string
                              optional:
                                description: Specify whether the Secret or its key
                                  must be defined
                                type: boolean
                      encrypt:
                        description: encrypt specifies whether the registry stores
                          the image in encrypted format or not. Optional, defaults
                          to false.
                        type: boolean
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, Encrypt must be true, or this parameter is ignored.
                        type: string
                      region:
                        description: region is the AWS region in which your bucket
                          exists. Optional, will be set based on the installed AWS
                          Region.
                        type: string
                      regionEndpoint:
                        description: regionEndpoint is the endpoint for S3 compatible
                          storage services. Optional, defaults based on the Region
                          that is provided.
                        type: string
                      virtualHostedStyle:
                        description: virtualHostedStyle enables using S3 virtual hosted
                          style bucket paths with a custom RegionEndpoint Optional,
                          defaults to false.
                        type: boolean
                  swift:
                    description: swift represents configuration that uses OpenStack
                      Object Storage.
                    type: object
                    properties:
                      authURL:
                        description: authURL defines the URL for obtaining an authentication
                          token.
                        type: string
                      authVersion:
                        description: authVersion specifies the OpenStack Auth's version.
                        type: string
                      container:
                        description: container defines the name of Swift container
                          where to store the registry's data.
                        type: string
                      domain:
                        description: domain specifies Openstack's domain name for
                          Identity v3 API.
                        type: string
                      domainID:
                        description: domainID specifies Openstack's domain id for
                          Identity v3 API.
                        type: string
                      regionName:
                        description: regionName defines Openstack's region in which
                          container exists.
                        type: string
                      tenant:
                        description: tenant defines Openstack tenant name to be used
                          by registry.
                        type: string
                      tenantID:
                        description: tenant defines Openstack tenant id to be used
                          by registry.
                        type: string
              storageManaged:
                description: storageManaged is deprecated, please refer to Storage.managementState
                type: boolean
              version:
                description: version is the level this availability applies to
                type: string
    served: true
    storage: true
    subresources:
      status: {}
`)

func assetsCrd0000_11_imageregistryConfigsCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_11_imageregistryConfigsCrdYaml, nil
}

func assetsCrd0000_11_imageregistryConfigsCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_11_imageregistryConfigsCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_11_imageregistry-configs.crd.yaml", size: 90225, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml = []byte(`allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: RunAsAny
groups:
- system:cluster-admins
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: anyuid provides all features of the restricted SCC
      but allows users to run with any UID and any GID.
  name: anyuid
priority: 10
readOnlyRootFilesystem: false
requiredDropCapabilities:
- MKNOD
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml", size: 1048, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml = []byte(`allowHostDirVolumePlugin: true
allowHostIPC: true
allowHostNetwork: true
allowHostPID: true
allowHostPorts: true
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: MustRunAs
groups: []
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: 'hostaccess allows access to all host namespaces but
      still requires pods to be run with a UID and SELinux context that are allocated
      to the namespace. WARNING: this SCC allows host access to namespaces, file systems,
      and PIDS.  It should only be used by trusted pods.  Grant with caution.'
  name: hostaccess
priority:
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- persistentVolumeClaim
- projected
- secret
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml", size: 1267, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml = []byte(`allowHostDirVolumePlugin: true
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: RunAsAny
groups: []
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: |-
      hostmount-anyuid provides all the features of the
      restricted SCC but allows host mounts and any UID by a pod.  This is primarily
      used by the persistent volume recycler. WARNING: this SCC allows host file
      system access as any UID, including UID 0.  Grant with caution.
  name: hostmount-anyuid
priority:
readOnlyRootFilesystem: false
requiredDropCapabilities:
- MKNOD
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:openshift-infra:pv-recycler-controller
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- nfs
- persistentVolumeClaim
- projected
- secret
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml", size: 1298, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml = []byte(`allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: true
allowHostPID: false
allowHostPorts: true
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: MustRunAs
groups: []
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: hostnetwork allows using host networking and host ports
      but still requires pods to be run with a UID and SELinux context that are allocated
      to the namespace.
  name: hostnetwork
priority:
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: MustRunAs
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml", size: 1123, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml = []byte(`allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: RunAsAny
groups: []
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: nonroot provides all features of the restricted SCC
      but allows users to run with any non-root UID.  The user must specify the UID
      or it must be specified on the by the manifest of the container runtime.
  name: nonroot
priority:
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsNonRoot
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml", size: 1166, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml = []byte(`allowHostDirVolumePlugin: true
allowHostIPC: true
allowHostNetwork: true
allowHostPID: true
allowHostPorts: true
allowPrivilegeEscalation: true
allowPrivilegedContainer: true
allowedCapabilities:
- "*"
allowedUnsafeSysctls:
- "*"
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: RunAsAny
groups:
- system:cluster-admins
- system:nodes
- system:masters
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: 'privileged allows access to all privileged and host
      features and the ability to run as any user, any group, any fsGroup, and with
      any SELinux context.  WARNING: this is the most relaxed SCC and should be used
      only for cluster administration. Grant with caution.'
  name: privileged
priority:
readOnlyRootFilesystem: false
requiredDropCapabilities:
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
seccompProfiles:
- "*"
supplementalGroups:
  type: RunAsAny
users:
- system:admin
- system:serviceaccount:openshift-infra:build-controller
volumes:
- "*"
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml", size: 1291, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml = []byte(`allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities:
apiVersion: security.openshift.io/v1
defaultAddCapabilities:
fsGroup:
  type: MustRunAs
groups:
- system:authenticated
kind: SecurityContextConstraints
metadata:
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
    release.openshift.io/create-only: "true"
    kubernetes.io/description: restricted denies access to all host features and requires
      pods to be run with a UID, and SELinux context that are allocated to the namespace.  This
      is the most restrictive SCC and it is used by default for authenticated users.
  name: restricted
priority:
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
`)

func assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYamlBytes() ([]byte, error) {
	return _assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml, nil
}

func assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml() (*asset, error) {
	bytes, err := assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml", size: 1213, mode: os.FileMode(436), modTime: time.Unix(1653308373, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/components/flannel/clusterrole.yaml":                                    assetsComponentsFlannelClusterroleYaml,
	"assets/components/flannel/clusterrolebinding.yaml":                             assetsComponentsFlannelClusterrolebindingYaml,
	"assets/components/flannel/configmap.yaml":                                      assetsComponentsFlannelConfigmapYaml,
	"assets/components/flannel/daemonset.yaml":                                      assetsComponentsFlannelDaemonsetYaml,
	"assets/components/flannel/podsecuritypolicy.yaml":                              assetsComponentsFlannelPodsecuritypolicyYaml,
	"assets/components/flannel/service-account.yaml":                                assetsComponentsFlannelServiceAccountYaml,
	"assets/components/hostpath-provisioner/clusterrole.yaml":                       assetsComponentsHostpathProvisionerClusterroleYaml,
	"assets/components/hostpath-provisioner/clusterrolebinding.yaml":                assetsComponentsHostpathProvisionerClusterrolebindingYaml,
	"assets/components/hostpath-provisioner/daemonset.yaml":                         assetsComponentsHostpathProvisionerDaemonsetYaml,
	"assets/components/hostpath-provisioner/namespace.yaml":                         assetsComponentsHostpathProvisionerNamespaceYaml,
	"assets/components/hostpath-provisioner/scc.yaml":                               assetsComponentsHostpathProvisionerSccYaml,
	"assets/components/hostpath-provisioner/service-account.yaml":                   assetsComponentsHostpathProvisionerServiceAccountYaml,
	"assets/components/hostpath-provisioner/storageclass.yaml":                      assetsComponentsHostpathProvisionerStorageclassYaml,
	"assets/components/openshift-dns/dns/cluster-role-binding.yaml":                 assetsComponentsOpenshiftDnsDnsClusterRoleBindingYaml,
	"assets/components/openshift-dns/dns/cluster-role.yaml":                         assetsComponentsOpenshiftDnsDnsClusterRoleYaml,
	"assets/components/openshift-dns/dns/configmap.yaml":                            assetsComponentsOpenshiftDnsDnsConfigmapYaml,
	"assets/components/openshift-dns/dns/daemonset.yaml":                            assetsComponentsOpenshiftDnsDnsDaemonsetYaml,
	"assets/components/openshift-dns/dns/namespace.yaml":                            assetsComponentsOpenshiftDnsDnsNamespaceYaml,
	"assets/components/openshift-dns/dns/service-account.yaml":                      assetsComponentsOpenshiftDnsDnsServiceAccountYaml,
	"assets/components/openshift-dns/dns/service.yaml":                              assetsComponentsOpenshiftDnsDnsServiceYaml,
	"assets/components/openshift-dns/node-resolver/daemonset.yaml":                  assetsComponentsOpenshiftDnsNodeResolverDaemonsetYaml,
	"assets/components/openshift-dns/node-resolver/service-account.yaml":            assetsComponentsOpenshiftDnsNodeResolverServiceAccountYaml,
	"assets/components/openshift-router/cluster-role-binding.yaml":                  assetsComponentsOpenshiftRouterClusterRoleBindingYaml,
	"assets/components/openshift-router/cluster-role.yaml":                          assetsComponentsOpenshiftRouterClusterRoleYaml,
	"assets/components/openshift-router/configmap.yaml":                             assetsComponentsOpenshiftRouterConfigmapYaml,
	"assets/components/openshift-router/deployment.yaml":                            assetsComponentsOpenshiftRouterDeploymentYaml,
	"assets/components/openshift-router/namespace.yaml":                             assetsComponentsOpenshiftRouterNamespaceYaml,
	"assets/components/openshift-router/service-account.yaml":                       assetsComponentsOpenshiftRouterServiceAccountYaml,
	"assets/components/openshift-router/service-cloud.yaml":                         assetsComponentsOpenshiftRouterServiceCloudYaml,
	"assets/components/openshift-router/service-internal.yaml":                      assetsComponentsOpenshiftRouterServiceInternalYaml,
	"assets/components/service-ca/clusterrole.yaml":                                 assetsComponentsServiceCaClusterroleYaml,
	"assets/components/service-ca/clusterrolebinding.yaml":                          assetsComponentsServiceCaClusterrolebindingYaml,
	"assets/components/service-ca/deployment.yaml":                                  assetsComponentsServiceCaDeploymentYaml,
	"assets/components/service-ca/ns.yaml":                                          assetsComponentsServiceCaNsYaml,
	"assets/components/service-ca/role.yaml":                                        assetsComponentsServiceCaRoleYaml,
	"assets/components/service-ca/rolebinding.yaml":                                 assetsComponentsServiceCaRolebindingYaml,
	"assets/components/service-ca/sa.yaml":                                          assetsComponentsServiceCaSaYaml,
	"assets/components/service-ca/signing-cabundle.yaml":                            assetsComponentsServiceCaSigningCabundleYaml,
	"assets/components/service-ca/signing-secret.yaml":                              assetsComponentsServiceCaSigningSecretYaml,
	"assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml":    assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml,
	"assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml,
	"assets/crd/0000_03_config-operator_01_proxy.crd.yaml":                          assetsCrd0000_03_configOperator_01_proxyCrdYaml,
	"assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml":           assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml,
	"assets/crd/0000_03_security-openshift_01_scc.crd.yaml":                         assetsCrd0000_03_securityOpenshift_01_sccCrdYaml,
	"assets/crd/0000_10_config-operator_01_build.crd.yaml":                          assetsCrd0000_10_configOperator_01_buildCrdYaml,
	"assets/crd/0000_10_config-operator_01_featuregate.crd.yaml":                    assetsCrd0000_10_configOperator_01_featuregateCrdYaml,
	"assets/crd/0000_10_config-operator_01_image.crd.yaml":                          assetsCrd0000_10_configOperator_01_imageCrdYaml,
	"assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml":       assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml,
	"assets/crd/0000_11_imageregistry-configs.crd.yaml":                             assetsCrd0000_11_imageregistryConfigsCrdYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml":                 assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml":             assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml":       assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml":            assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml":                assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml":             assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml":             assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": {nil, map[string]*bintree{
		"components": {nil, map[string]*bintree{
			"flannel": {nil, map[string]*bintree{
				"clusterrole.yaml":        {assetsComponentsFlannelClusterroleYaml, map[string]*bintree{}},
				"clusterrolebinding.yaml": {assetsComponentsFlannelClusterrolebindingYaml, map[string]*bintree{}},
				"configmap.yaml":          {assetsComponentsFlannelConfigmapYaml, map[string]*bintree{}},
				"daemonset.yaml":          {assetsComponentsFlannelDaemonsetYaml, map[string]*bintree{}},
				"podsecuritypolicy.yaml":  {assetsComponentsFlannelPodsecuritypolicyYaml, map[string]*bintree{}},
				"service-account.yaml":    {assetsComponentsFlannelServiceAccountYaml, map[string]*bintree{}},
			}},
			"hostpath-provisioner": {nil, map[string]*bintree{
				"clusterrole.yaml":        {assetsComponentsHostpathProvisionerClusterroleYaml, map[string]*bintree{}},
				"clusterrolebinding.yaml": {assetsComponentsHostpathProvisionerClusterrolebindingYaml, map[string]*bintree{}},
				"daemonset.yaml":          {assetsComponentsHostpathProvisionerDaemonsetYaml, map[string]*bintree{}},
				"namespace.yaml":          {assetsComponentsHostpathProvisionerNamespaceYaml, map[string]*bintree{}},
				"scc.yaml":                {assetsComponentsHostpathProvisionerSccYaml, map[string]*bintree{}},
				"service-account.yaml":    {assetsComponentsHostpathProvisionerServiceAccountYaml, map[string]*bintree{}},
				"storageclass.yaml":       {assetsComponentsHostpathProvisionerStorageclassYaml, map[string]*bintree{}},
			}},
			"openshift-dns": {nil, map[string]*bintree{
				"dns": {nil, map[string]*bintree{
					"cluster-role-binding.yaml": {assetsComponentsOpenshiftDnsDnsClusterRoleBindingYaml, map[string]*bintree{}},
					"cluster-role.yaml":         {assetsComponentsOpenshiftDnsDnsClusterRoleYaml, map[string]*bintree{}},
					"configmap.yaml":            {assetsComponentsOpenshiftDnsDnsConfigmapYaml, map[string]*bintree{}},
					"daemonset.yaml":            {assetsComponentsOpenshiftDnsDnsDaemonsetYaml, map[string]*bintree{}},
					"namespace.yaml":            {assetsComponentsOpenshiftDnsDnsNamespaceYaml, map[string]*bintree{}},
					"service-account.yaml":      {assetsComponentsOpenshiftDnsDnsServiceAccountYaml, map[string]*bintree{}},
					"service.yaml":              {assetsComponentsOpenshiftDnsDnsServiceYaml, map[string]*bintree{}},
				}},
				"node-resolver": {nil, map[string]*bintree{
					"daemonset.yaml":       {assetsComponentsOpenshiftDnsNodeResolverDaemonsetYaml, map[string]*bintree{}},
					"service-account.yaml": {assetsComponentsOpenshiftDnsNodeResolverServiceAccountYaml, map[string]*bintree{}},
				}},
			}},
			"openshift-router": {nil, map[string]*bintree{
				"cluster-role-binding.yaml": {assetsComponentsOpenshiftRouterClusterRoleBindingYaml, map[string]*bintree{}},
				"cluster-role.yaml":         {assetsComponentsOpenshiftRouterClusterRoleYaml, map[string]*bintree{}},
				"configmap.yaml":            {assetsComponentsOpenshiftRouterConfigmapYaml, map[string]*bintree{}},
				"deployment.yaml":           {assetsComponentsOpenshiftRouterDeploymentYaml, map[string]*bintree{}},
				"namespace.yaml":            {assetsComponentsOpenshiftRouterNamespaceYaml, map[string]*bintree{}},
				"service-account.yaml":      {assetsComponentsOpenshiftRouterServiceAccountYaml, map[string]*bintree{}},
				"service-cloud.yaml":        {assetsComponentsOpenshiftRouterServiceCloudYaml, map[string]*bintree{}},
				"service-internal.yaml":     {assetsComponentsOpenshiftRouterServiceInternalYaml, map[string]*bintree{}},
			}},
			"service-ca": {nil, map[string]*bintree{
				"clusterrole.yaml":        {assetsComponentsServiceCaClusterroleYaml, map[string]*bintree{}},
				"clusterrolebinding.yaml": {assetsComponentsServiceCaClusterrolebindingYaml, map[string]*bintree{}},
				"deployment.yaml":         {assetsComponentsServiceCaDeploymentYaml, map[string]*bintree{}},
				"ns.yaml":                 {assetsComponentsServiceCaNsYaml, map[string]*bintree{}},
				"role.yaml":               {assetsComponentsServiceCaRoleYaml, map[string]*bintree{}},
				"rolebinding.yaml":        {assetsComponentsServiceCaRolebindingYaml, map[string]*bintree{}},
				"sa.yaml":                 {assetsComponentsServiceCaSaYaml, map[string]*bintree{}},
				"signing-cabundle.yaml":   {assetsComponentsServiceCaSigningCabundleYaml, map[string]*bintree{}},
				"signing-secret.yaml":     {assetsComponentsServiceCaSigningSecretYaml, map[string]*bintree{}},
			}},
		}},
		"core": {nil, map[string]*bintree{
			"0000_50_cluster-openshift-controller-manager_00_namespace.yaml": {assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml, map[string]*bintree{}},
		}},
		"crd": {nil, map[string]*bintree{
			"0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": {assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml, map[string]*bintree{}},
			"0000_03_config-operator_01_proxy.crd.yaml":                          {assetsCrd0000_03_configOperator_01_proxyCrdYaml, map[string]*bintree{}},
			"0000_03_quota-openshift_01_clusterresourcequota.crd.yaml":           {assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml, map[string]*bintree{}},
			"0000_03_security-openshift_01_scc.crd.yaml":                         {assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_build.crd.yaml":                          {assetsCrd0000_10_configOperator_01_buildCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_featuregate.crd.yaml":                    {assetsCrd0000_10_configOperator_01_featuregateCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_image.crd.yaml":                          {assetsCrd0000_10_configOperator_01_imageCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml":       {assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml, map[string]*bintree{}},
			"0000_11_imageregistry-configs.crd.yaml":                             {assetsCrd0000_11_imageregistryConfigsCrdYaml, map[string]*bintree{}},
		}},
		"scc": {nil, map[string]*bintree{
			"0000_20_kube-apiserver-operator_00_scc-anyuid.yaml":           {assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml":       {assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml": {assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml":      {assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-nonroot.yaml":          {assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-privileged.yaml":       {assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-restricted.yaml":       {assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
