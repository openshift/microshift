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
// assets/crd/0000_03_security-openshift_01_scc.crd.yaml
// assets/crd/0000_10_config-operator_01_featuregate.crd.yaml
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

	info := bindataFileInfo{name: "assets/components/flannel/clusterrole.yaml", size: 418, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/flannel/clusterrolebinding.yaml", size: 248, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/flannel/configmap.yaml", size: 674, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/flannel/daemonset.yaml", size: 2657, mode: os.FileMode(420), modTime: time.Unix(1653503271, 0)}
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

	info := bindataFileInfo{name: "assets/components/flannel/podsecuritypolicy.yaml", size: 1195, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/flannel/service-account.yaml", size: 86, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/clusterrole.yaml", size: 609, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/clusterrolebinding.yaml", size: 338, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/daemonset.yaml", size: 1231, mode: os.FileMode(420), modTime: time.Unix(1653503271, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/namespace.yaml", size: 78, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/scc.yaml", size: 480, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/service-account.yaml", size: 132, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/storageclass.yaml", size: 276, mode: os.FileMode(420), modTime: time.Unix(1652799551, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/cluster-role-binding.yaml", size: 223, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/cluster-role.yaml", size: 492, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/configmap.yaml", size: 610, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/daemonset.yaml", size: 3217, mode: os.FileMode(420), modTime: time.Unix(1653503271, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/namespace.yaml", size: 417, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/service-account.yaml", size: 85, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/service.yaml", size: 691, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/node-resolver/daemonset.yaml", size: 4823, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/node-resolver/service-account.yaml", size: 95, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/cluster-role-binding.yaml", size: 329, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/cluster-role.yaml", size: 883, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/configmap.yaml", size: 168, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/deployment.yaml", size: 4746, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/namespace.yaml", size: 499, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/service-account.yaml", size: 213, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/service-cloud.yaml", size: 567, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/service-internal.yaml", size: 727, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/clusterrole.yaml", size: 864, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/clusterrolebinding.yaml", size: 298, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/deployment.yaml", size: 1866, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/ns.yaml", size: 168, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/role.yaml", size: 634, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/rolebinding.yaml", size: 343, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/sa.yaml", size: 99, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/signing-cabundle.yaml", size: 123, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/signing-secret.yaml", size: 144, mode: os.FileMode(420), modTime: time.Unix(1648137870, 0)}
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

	info := bindataFileInfo{name: "assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml", size: 254, mode: os.FileMode(420), modTime: time.Unix(1624916338, 0)}
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

	info := bindataFileInfo{name: "assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml", size: 9898, mode: os.FileMode(420), modTime: time.Unix(1652799551, 0)}
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

	info := bindataFileInfo{name: "assets/crd/0000_03_security-openshift_01_scc.crd.yaml", size: 16010, mode: os.FileMode(420), modTime: time.Unix(1652799551, 0)}
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

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_featuregate.crd.yaml", size: 3438, mode: os.FileMode(420), modTime: time.Unix(1652799551, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml", size: 1048, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml", size: 1267, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml", size: 1298, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml", size: 1123, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml", size: 1166, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml", size: 1291, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml", size: 1213, mode: os.FileMode(420), modTime: time.Unix(1633965440, 0)}
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
	"assets/crd/0000_03_security-openshift_01_scc.crd.yaml":                         assetsCrd0000_03_securityOpenshift_01_sccCrdYaml,
	"assets/crd/0000_10_config-operator_01_featuregate.crd.yaml":                    assetsCrd0000_10_configOperator_01_featuregateCrdYaml,
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
			"0000_03_security-openshift_01_scc.crd.yaml":                         {assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_featuregate.crd.yaml":                    {assetsCrd0000_10_configOperator_01_featuregateCrdYaml, map[string]*bintree{}},
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
