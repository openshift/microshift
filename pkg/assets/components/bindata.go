// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/components/flannel/0000_00_flannel-clusterrole.yaml
// assets/components/flannel/0000_00_flannel-clusterrolebinding.yaml
// assets/components/flannel/0000_00_flannel-configmap.yaml
// assets/components/flannel/0000_00_flannel-daemonset.yaml
// assets/components/flannel/0000_00_flannel-service-account.yaml
// assets/components/flannel/0000_00_podsecuritypolicy-flannel.yaml
// assets/components/flannel/kustomization.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-clusterrole.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-clusterrolebinding.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-daemonset.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-namespace.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-securitycontextconstraints.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-serviceaccount.yaml
// assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-storageclass.yaml
// assets/components/hostpath-provisioner/kustomization.yaml
// assets/components/openshift-dns/0000_70_dns_00-namespace.yaml
// assets/components/openshift-dns/0000_70_dns_01-cluster-role-binding.yaml
// assets/components/openshift-dns/0000_70_dns_01-cluster-role.yaml
// assets/components/openshift-dns/0000_70_dns_01-configmap.yaml
// assets/components/openshift-dns/0000_70_dns_01-daemonset.yaml
// assets/components/openshift-dns/0000_70_dns_01-service-account.yaml
// assets/components/openshift-dns/0000_70_dns_01-service.yaml
// assets/components/openshift-dns/kustomization.yaml
// assets/components/openshift-router/0000_80_openshift-router-cluster-role-binding.yaml
// assets/components/openshift-router/0000_80_openshift-router-cluster-role.yaml
// assets/components/openshift-router/0000_80_openshift-router-cm.yaml
// assets/components/openshift-router/0000_80_openshift-router-deployment.yaml
// assets/components/openshift-router/0000_80_openshift-router-namespace.yaml
// assets/components/openshift-router/0000_80_openshift-router-service-account.yaml
// assets/components/openshift-router/0000_80_openshift-router-service.yaml
// assets/components/openshift-router/kustomization.yaml
// assets/components/service-ca/0000_60_service-ca_00_roles.yaml
// assets/components/service-ca/0000_60_service-ca_01_namespace.yaml
// assets/components/service-ca/0000_60_service-ca_04_sa.yaml
// assets/components/service-ca/0000_60_service-ca_05_deploy.yaml
// assets/components/service-ca/kustomization.yaml
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

var _assetsComponentsFlannel0000_00_flannelClusterroleYaml = []byte(`kind: ClusterRole
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

func assetsComponentsFlannel0000_00_flannelClusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannel0000_00_flannelClusterroleYaml, nil
}

func assetsComponentsFlannel0000_00_flannelClusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannel0000_00_flannelClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/0000_00_flannel-clusterrole.yaml", size: 418, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannel0000_00_flannelClusterrolebindingYaml = []byte(`kind: ClusterRoleBinding
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

func assetsComponentsFlannel0000_00_flannelClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannel0000_00_flannelClusterrolebindingYaml, nil
}

func assetsComponentsFlannel0000_00_flannelClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannel0000_00_flannelClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/0000_00_flannel-clusterrolebinding.yaml", size: 248, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannel0000_00_flannelConfigmapYaml = []byte(`kind: ConfigMap
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

func assetsComponentsFlannel0000_00_flannelConfigmapYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannel0000_00_flannelConfigmapYaml, nil
}

func assetsComponentsFlannel0000_00_flannelConfigmapYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannel0000_00_flannelConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/0000_00_flannel-configmap.yaml", size: 640, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannel0000_00_flannelDaemonsetYaml = []byte(`apiVersion: apps/v1
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
      - name: install-cni
        image: kube_flannel
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
        image: kube_flannel
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
          name: kube-flannel-cfg`)

func assetsComponentsFlannel0000_00_flannelDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannel0000_00_flannelDaemonsetYaml, nil
}

func assetsComponentsFlannel0000_00_flannelDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannel0000_00_flannelDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/0000_00_flannel-daemonset.yaml", size: 2159, mode: os.FileMode(420), modTime: time.Unix(1631530548, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannel0000_00_flannelServiceAccountYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: flannel
  namespace: kube-system`)

func assetsComponentsFlannel0000_00_flannelServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannel0000_00_flannelServiceAccountYaml, nil
}

func assetsComponentsFlannel0000_00_flannelServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannel0000_00_flannelServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/0000_00_flannel-service-account.yaml", size: 86, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYaml = []byte(`apiVersion: policy/v1beta1
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

func assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYaml, nil
}

func assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/0000_00_podsecuritypolicy-flannel.yaml", size: 1195, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsFlannelKustomizationYaml = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 0000_00_flannel-clusterrolebinding.yaml
  - 0000_00_flannel-clusterrole.yaml
  - 0000_00_flannel-configmap.yaml
  - 0000_00_flannel-daemonset.yaml
  - 0000_00_flannel-service-account.yaml
  - 0000_00_podsecuritypolicy-flannel.yaml
`)

func assetsComponentsFlannelKustomizationYamlBytes() ([]byte, error) {
	return _assetsComponentsFlannelKustomizationYaml, nil
}

func assetsComponentsFlannelKustomizationYaml() (*asset, error) {
	bytes, err := assetsComponentsFlannelKustomizationYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/flannel/kustomization.yaml", size: 311, mode: os.FileMode(420), modTime: time.Unix(1631715744, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYaml = []byte(`kind: ClusterRole
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

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-clusterrole.yaml", size: 609, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
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

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-clusterrolebinding.yaml", size: 338, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYaml = []byte(`apiVersion: apps/v1
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
          image: kubevirt_hostpath_provisioner
          imagePullPolicy: Always
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

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-daemonset.yaml", size: 1205, mode: os.FileMode(420), modTime: time.Unix(1631530604, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: kubevirt-hostpath-provisioner`)

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-namespace.yaml", size: 78, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYaml = []byte(`kind: SecurityContextConstraints
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
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowedCapabilities: []
defaultAddCapabilities: []
priority: 0
readOnlyRootFilesystem: false
`)

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-securitycontextconstraints.yaml", size: 659, mode: os.FileMode(420), modTime: time.Unix(1631722129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubevirt-hostpath-provisioner-admin
  namespace: kubevirt-hostpath-provisioner`)

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-serviceaccount.yaml", size: 132, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYaml = []byte(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: kubevirt-hostpath-provisioner
provisioner: kubevirt.io/hostpath-provisioner
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer`)

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYaml, nil
}

func assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-storageclass.yaml", size: 204, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsHostpathProvisionerKustomizationYaml = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 0000_80_hostpath-provisioner-clusterrolebinding.yaml
  - 0000_80_hostpath-provisioner-clusterrole.yaml
  - 0000_80_hostpath-provisioner-daemonset.yaml
  - 0000_80_hostpath-provisioner-namespace.yaml
  - 0000_80_hostpath-provisioner-securitycontextconstraints.yaml
  - 0000_80_hostpath-provisioner-serviceaccount.yaml
  - 0000_80_hostpath-provisioner-storageclass.yaml
`)

func assetsComponentsHostpathProvisionerKustomizationYamlBytes() ([]byte, error) {
	return _assetsComponentsHostpathProvisionerKustomizationYaml, nil
}

func assetsComponentsHostpathProvisionerKustomizationYaml() (*asset, error) {
	bytes, err := assetsComponentsHostpathProvisionerKustomizationYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/hostpath-provisioner/kustomization.yaml", size: 448, mode: os.FileMode(420), modTime: time.Unix(1631715752, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYaml = []byte(`kind: Namespace
apiVersion: v1
metadata:
  annotations:
    openshift.io/node-selector: ""
  name: openshift-dns
  labels:
    # set value to avoid depending on kube admission that depends on openshift apis
    openshift.io/run-level: "0"
    # allow openshift-monitoring to look for ServiceMonitor objects in this namespace
    openshift.io/cluster-monitoring: "true"
`)

func assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_00-namespace.yaml", size: 369, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYaml = []byte(`kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
    name: openshift-dns
subjects:
- kind: ServiceAccount
  name: dns
  namespace: openshift-dns
roleRef:
  kind: ClusterRole
  apiGroup: rbac.authorization.k8s.io
  name: openshift-dns
`)

func assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_01-cluster-role-binding.yaml", size: 261, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYaml = []byte(`kind: ClusterRole
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
  - create`)

func assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_01-cluster-role.yaml", size: 488, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYaml = []byte(`apiVersion: v1
data:
  Corefile: |
    .:5353 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            upstream
            fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf {
            policy sequential
        }
        cache 30
        reload
    }
kind: ConfigMap
metadata:
  labels:
    dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns
`)

func assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_01-configmap.yaml", size: 511, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYaml = []byte(`kind: DaemonSet
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
    spec:
      serviceAccountName: dns
      priorityClassName: system-node-critical
      containers:
      - name: dns
        image: coredns
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
          failureThreshold: 3
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 10
        livenessProbe:
          failureThreshold: 5
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
        resources:
          requests:
            cpu: 50m
            memory: 70Mi
      - name: kube-rbac-proxy
        image: kube_rbac_proxy
        args:
        - --logtostderr
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
      - name: dns-node-resolver
        image: cli
        imagePullPolicy: IfNotPresent
        terminationMessagePolicy: FallbackToLogsOnError
        securityContext:
          privileged: true
        volumeMounts:
        - name: hosts-file
          mountPath: /etc/hosts
        env:
        - name: SERVICES
          value: "image-registry.openshift-image-registry.svc"
        - name: CLUSTER_DOMAIN
          value: cluster.local        
        command:
        - /bin/bash
        - -c
        - |
          #!/bin/bash
          set -uo pipefail
          NAMESERVER=${DNS_DEFAULT_SERVICE_HOST}

          trap 'jobs -p | xargs kill || true; wait; exit 0' TERM

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
        resources:
          requests:
            cpu: 5m
            memory: 21Mi
      dnsPolicy: Default
      nodeSelector:
        kubernetes.io/os: linux      
      volumes:
      - name: config-volume
        configMap:
          defaultMode: 420
          items:
          - key: Corefile
            path: Corefile
          name: dns-default
      - name: hosts-file
        hostPath:
          path: /etc/hosts
          type: File
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
      # Note: The daemon controller rounds the percentage up
      # (unlike the deployment controller, which rounds down).
      maxUnavailable: 10%
`)

func assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_01-daemonset.yaml", size: 6518, mode: os.FileMode(420), modTime: time.Unix(1631530679, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYaml = []byte(`kind: ServiceAccount
apiVersion: v1
metadata:
  name: dns
  namespace: openshift-dns
`)

func assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_01-service-account.yaml", size: 85, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDns0000_70_dns_01ServiceYaml = []byte(`kind: Service
apiVersion: v1
metadata:
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: dns-default-metrics-tls
  labels:
      dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns          
# name, namespace,labels and annotations are set at runtime
spec:
  clusterIP: 10.43.0.10
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
  selector:
    dns.operator.openshift.io/daemonset-dns: default    
`)

func assetsComponentsOpenshiftDns0000_70_dns_01ServiceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDns0000_70_dns_01ServiceYaml, nil
}

func assetsComponentsOpenshiftDns0000_70_dns_01ServiceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDns0000_70_dns_01ServiceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/0000_70_dns_01-service.yaml", size: 634, mode: os.FileMode(420), modTime: time.Unix(1631531694, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftDnsKustomizationYaml = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 0000_70_dns_00-namespace.yaml
  - 0000_70_dns_01-cluster-role-binding.yaml
  - 0000_70_dns_01-cluster-role.yaml
  - 0000_70_dns_01-configmap.yaml
  - 0000_70_dns_01-daemonset.yaml
  - 0000_70_dns_01-service-account.yaml
  - 0000_70_dns_01-service.yaml
`)

func assetsComponentsOpenshiftDnsKustomizationYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsKustomizationYaml, nil
}

func assetsComponentsOpenshiftDnsKustomizationYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsKustomizationYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/kustomization.yaml", size: 332, mode: os.FileMode(420), modTime: time.Unix(1631715756, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYaml = []byte(`# Binds the router role to its Service Account.
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
  apiGroup: rbac.authorization.k8s.io
`)

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-cluster-role-binding.yaml", size: 336, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYaml = []byte(`# Cluster scoped role for routers. This should be as restrictive as possible.
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

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-cluster-role.yaml", size: 883, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYaml = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: openshift-ingress
  name: service-ca-bundle 
  annotations:
    service.beta.openshift.io/inject-cabundle: "true"
`)

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-cm.yaml", size: 168, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYaml = []byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: router-default
  namespace: openshift-ingress
  labels:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
spec:
  progressDeadlineSeconds: 600
  selector:
    matchLabels:
      ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  template:
    metadata:
      labels:
        ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
    spec:
      serviceAccountName: router
      # nodeSelector is set at runtime.
      priorityClassName: system-cluster-critical
      containers:
        - name: router
          image: haproxy_router
          imagePullPolicy: IfNotPresent
          terminationMessagePolicy: FallbackToLogsOnError
          ports:
          - name: http
            containerPort: 80
            protocol: TCP
          - name: https
            containerPort: 443
            protocol: TCP
          - name: metrics
            containerPort: 1936
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
               path: /healthz/ready
               port: 1936
               scheme: HTTP
             initialDelaySeconds: 10
             periodSeconds: 10
             successThreshold: 1
             timeoutSeconds: 1
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
          securityContext:
            privileged: true              
          volumeMounts:
          - mountPath: /etc/pki/tls/private
            name: default-certificate
            readOnly: true
          - mountPath: /var/run/configmaps/service-ca
            name: service-ca-bundle
            readOnly: true
      volumes:
      - name: default-certificate
        secret:
          secretName: router-certs-default
      - name: service-ca-bundle
        configMap:
          items:
          - key: service-ca.crt
            path: service-ca.crt
          name: service-ca-bundle
          optional: false
`)

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-deployment.yaml", size: 3837, mode: os.FileMode(420), modTime: time.Unix(1631530699, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYaml = []byte(`kind: Namespace
apiVersion: v1
metadata:
  name: openshift-ingress
  annotations:
    openshift.io/node-selector: ""
  labels:
    # allow openshift-monitoring to look for ServiceMonitor objects in this namespace
    openshift.io/cluster-monitoring: "true"
    name: openshift-ingress
    network.openshift.io/policy-group: ingress
`)

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-namespace.yaml", size: 332, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYaml = []byte(`# Account for routers created by the operator. It will require cluster scoped
# permissions related to Route processing.
kind: ServiceAccount
apiVersion: v1
metadata:
  name: router
  namespace: openshift-ingress
`)

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-service-account.yaml", size: 213, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYaml = []byte(`kind: Service
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

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYaml, nil
}

func assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/0000_80_openshift-router-service.yaml", size: 628, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOpenshiftRouterKustomizationYaml = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 0000_80_openshift-router-cluster-role-binding.yaml
  - 0000_80_openshift-router-cluster-role.yaml
  - 0000_80_openshift-router-cm.yaml
  - 0000_80_openshift-router-deployment.yaml
  - 0000_80_openshift-router-namespace.yaml
  - 0000_80_openshift-router-service-account.yaml
  - 0000_80_openshift-router-service.yaml
`)

func assetsComponentsOpenshiftRouterKustomizationYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterKustomizationYaml, nil
}

func assetsComponentsOpenshiftRouterKustomizationYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterKustomizationYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/kustomization.yaml", size: 396, mode: os.FileMode(420), modTime: time.Unix(1631715764, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCa0000_60_serviceCa_00_rolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:operator:service-ca
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  namespace: openshift-service-ca
  name: service-ca
`)

func assetsComponentsServiceCa0000_60_serviceCa_00_rolesYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCa0000_60_serviceCa_00_rolesYaml, nil
}

func assetsComponentsServiceCa0000_60_serviceCa_00_rolesYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCa0000_60_serviceCa_00_rolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/0000_60_service-ca_00_roles.yaml", size: 296, mode: os.FileMode(420), modTime: time.Unix(1631716637, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  labels:
    openshift.io/run-level: "1"
    openshift.io/cluster-monitoring: "true"
  name: openshift-service-ca
`)

func assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYaml, nil
}

func assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/0000_60_service-ca_01_namespace.yaml", size: 156, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCa0000_60_serviceCa_04_saYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: openshift-service-ca
  name: service-ca
  labels:
    app: service-ca
`)

func assetsComponentsServiceCa0000_60_serviceCa_04_saYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCa0000_60_serviceCa_04_saYaml, nil
}

func assetsComponentsServiceCa0000_60_serviceCa_04_saYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCa0000_60_serviceCa_04_saYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/0000_60_service-ca_04_sa.yaml", size: 129, mode: os.FileMode(420), modTime: time.Unix(1631293129, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCa0000_60_serviceCa_05_deployYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: openshift-service-ca
  name: service-ca
  labels:
    app: service-ca
spec:
  replicas: 1
  selector:
    matchLabels:
      app: service-ca
  template:
    metadata:
      name: service-ca
      labels:
        app: service-ca
    spec:
      serviceAccountName: service-ca
      containers:
      - name: service-ca-controller
        image: service_ca_operator
        imagePullPolicy: IfNotPresent
        command: ["service-ca-operator", "controller"]
        args:
        - "-v=4"
        ports:
          - containerPort: 8443
            protocol: TCP
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
          hostPath:
            path: /var/lib/microshift/resources/service-ca/secrets/service-ca
        - name: signing-cabundle
          hostPath:
            path: /var/lib/microshift/certs/ca-bundle
      #nodeSelector:
      #  node-role.kubernetes.io/master: ""
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

func assetsComponentsServiceCa0000_60_serviceCa_05_deployYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCa0000_60_serviceCa_05_deployYaml, nil
}

func assetsComponentsServiceCa0000_60_serviceCa_05_deployYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCa0000_60_serviceCa_05_deployYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/0000_60_service-ca_05_deploy.yaml", size: 1649, mode: os.FileMode(420), modTime: time.Unix(1631531913, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsServiceCaKustomizationYaml = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - 0000_60_service-ca_00_roles.yaml
  - 0000_60_service-ca_01_namespace.yaml
  - 0000_60_service-ca_04_sa.yaml
  - 0000_60_service-ca_05_deploy.yaml
`)

func assetsComponentsServiceCaKustomizationYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaKustomizationYaml, nil
}

func assetsComponentsServiceCaKustomizationYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaKustomizationYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/kustomization.yaml", size: 226, mode: os.FileMode(420), modTime: time.Unix(1631715770, 0)}
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
	"assets/components/flannel/0000_00_flannel-clusterrole.yaml":                                          assetsComponentsFlannel0000_00_flannelClusterroleYaml,
	"assets/components/flannel/0000_00_flannel-clusterrolebinding.yaml":                                   assetsComponentsFlannel0000_00_flannelClusterrolebindingYaml,
	"assets/components/flannel/0000_00_flannel-configmap.yaml":                                            assetsComponentsFlannel0000_00_flannelConfigmapYaml,
	"assets/components/flannel/0000_00_flannel-daemonset.yaml":                                            assetsComponentsFlannel0000_00_flannelDaemonsetYaml,
	"assets/components/flannel/0000_00_flannel-service-account.yaml":                                      assetsComponentsFlannel0000_00_flannelServiceAccountYaml,
	"assets/components/flannel/0000_00_podsecuritypolicy-flannel.yaml":                                    assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYaml,
	"assets/components/flannel/kustomization.yaml":                                                        assetsComponentsFlannelKustomizationYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-clusterrole.yaml":                assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-clusterrolebinding.yaml":         assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-daemonset.yaml":                  assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-namespace.yaml":                  assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-securitycontextconstraints.yaml": assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-serviceaccount.yaml":             assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYaml,
	"assets/components/hostpath-provisioner/0000_80_hostpath-provisioner-storageclass.yaml":               assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYaml,
	"assets/components/hostpath-provisioner/kustomization.yaml":                                           assetsComponentsHostpathProvisionerKustomizationYaml,
	"assets/components/openshift-dns/0000_70_dns_00-namespace.yaml":                                       assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYaml,
	"assets/components/openshift-dns/0000_70_dns_01-cluster-role-binding.yaml":                            assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYaml,
	"assets/components/openshift-dns/0000_70_dns_01-cluster-role.yaml":                                    assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYaml,
	"assets/components/openshift-dns/0000_70_dns_01-configmap.yaml":                                       assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYaml,
	"assets/components/openshift-dns/0000_70_dns_01-daemonset.yaml":                                       assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYaml,
	"assets/components/openshift-dns/0000_70_dns_01-service-account.yaml":                                 assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYaml,
	"assets/components/openshift-dns/0000_70_dns_01-service.yaml":                                         assetsComponentsOpenshiftDns0000_70_dns_01ServiceYaml,
	"assets/components/openshift-dns/kustomization.yaml":                                                  assetsComponentsOpenshiftDnsKustomizationYaml,
	"assets/components/openshift-router/0000_80_openshift-router-cluster-role-binding.yaml":               assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYaml,
	"assets/components/openshift-router/0000_80_openshift-router-cluster-role.yaml":                       assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYaml,
	"assets/components/openshift-router/0000_80_openshift-router-cm.yaml":                                 assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYaml,
	"assets/components/openshift-router/0000_80_openshift-router-deployment.yaml":                         assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYaml,
	"assets/components/openshift-router/0000_80_openshift-router-namespace.yaml":                          assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYaml,
	"assets/components/openshift-router/0000_80_openshift-router-service-account.yaml":                    assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYaml,
	"assets/components/openshift-router/0000_80_openshift-router-service.yaml":                            assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYaml,
	"assets/components/openshift-router/kustomization.yaml":                                               assetsComponentsOpenshiftRouterKustomizationYaml,
	"assets/components/service-ca/0000_60_service-ca_00_roles.yaml":                                       assetsComponentsServiceCa0000_60_serviceCa_00_rolesYaml,
	"assets/components/service-ca/0000_60_service-ca_01_namespace.yaml":                                   assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYaml,
	"assets/components/service-ca/0000_60_service-ca_04_sa.yaml":                                          assetsComponentsServiceCa0000_60_serviceCa_04_saYaml,
	"assets/components/service-ca/0000_60_service-ca_05_deploy.yaml":                                      assetsComponentsServiceCa0000_60_serviceCa_05_deployYaml,
	"assets/components/service-ca/kustomization.yaml":                                                     assetsComponentsServiceCaKustomizationYaml,
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
				"0000_00_flannel-clusterrole.yaml":        {assetsComponentsFlannel0000_00_flannelClusterroleYaml, map[string]*bintree{}},
				"0000_00_flannel-clusterrolebinding.yaml": {assetsComponentsFlannel0000_00_flannelClusterrolebindingYaml, map[string]*bintree{}},
				"0000_00_flannel-configmap.yaml":          {assetsComponentsFlannel0000_00_flannelConfigmapYaml, map[string]*bintree{}},
				"0000_00_flannel-daemonset.yaml":          {assetsComponentsFlannel0000_00_flannelDaemonsetYaml, map[string]*bintree{}},
				"0000_00_flannel-service-account.yaml":    {assetsComponentsFlannel0000_00_flannelServiceAccountYaml, map[string]*bintree{}},
				"0000_00_podsecuritypolicy-flannel.yaml":  {assetsComponentsFlannel0000_00_podsecuritypolicyFlannelYaml, map[string]*bintree{}},
				"kustomization.yaml":                      {assetsComponentsFlannelKustomizationYaml, map[string]*bintree{}},
			}},
			"hostpath-provisioner": {nil, map[string]*bintree{
				"0000_80_hostpath-provisioner-clusterrole.yaml":                {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterroleYaml, map[string]*bintree{}},
				"0000_80_hostpath-provisioner-clusterrolebinding.yaml":         {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerClusterrolebindingYaml, map[string]*bintree{}},
				"0000_80_hostpath-provisioner-daemonset.yaml":                  {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerDaemonsetYaml, map[string]*bintree{}},
				"0000_80_hostpath-provisioner-namespace.yaml":                  {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerNamespaceYaml, map[string]*bintree{}},
				"0000_80_hostpath-provisioner-securitycontextconstraints.yaml": {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerSecuritycontextconstraintsYaml, map[string]*bintree{}},
				"0000_80_hostpath-provisioner-serviceaccount.yaml":             {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerServiceaccountYaml, map[string]*bintree{}},
				"0000_80_hostpath-provisioner-storageclass.yaml":               {assetsComponentsHostpathProvisioner0000_80_hostpathProvisionerStorageclassYaml, map[string]*bintree{}},
				"kustomization.yaml":                                           {assetsComponentsHostpathProvisionerKustomizationYaml, map[string]*bintree{}},
			}},
			"openshift-dns": {nil, map[string]*bintree{
				"0000_70_dns_00-namespace.yaml":            {assetsComponentsOpenshiftDns0000_70_dns_00NamespaceYaml, map[string]*bintree{}},
				"0000_70_dns_01-cluster-role-binding.yaml": {assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleBindingYaml, map[string]*bintree{}},
				"0000_70_dns_01-cluster-role.yaml":         {assetsComponentsOpenshiftDns0000_70_dns_01ClusterRoleYaml, map[string]*bintree{}},
				"0000_70_dns_01-configmap.yaml":            {assetsComponentsOpenshiftDns0000_70_dns_01ConfigmapYaml, map[string]*bintree{}},
				"0000_70_dns_01-daemonset.yaml":            {assetsComponentsOpenshiftDns0000_70_dns_01DaemonsetYaml, map[string]*bintree{}},
				"0000_70_dns_01-service-account.yaml":      {assetsComponentsOpenshiftDns0000_70_dns_01ServiceAccountYaml, map[string]*bintree{}},
				"0000_70_dns_01-service.yaml":              {assetsComponentsOpenshiftDns0000_70_dns_01ServiceYaml, map[string]*bintree{}},
				"kustomization.yaml":                       {assetsComponentsOpenshiftDnsKustomizationYaml, map[string]*bintree{}},
			}},
			"openshift-router": {nil, map[string]*bintree{
				"0000_80_openshift-router-cluster-role-binding.yaml": {assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleBindingYaml, map[string]*bintree{}},
				"0000_80_openshift-router-cluster-role.yaml":         {assetsComponentsOpenshiftRouter0000_80_openshiftRouterClusterRoleYaml, map[string]*bintree{}},
				"0000_80_openshift-router-cm.yaml":                   {assetsComponentsOpenshiftRouter0000_80_openshiftRouterCmYaml, map[string]*bintree{}},
				"0000_80_openshift-router-deployment.yaml":           {assetsComponentsOpenshiftRouter0000_80_openshiftRouterDeploymentYaml, map[string]*bintree{}},
				"0000_80_openshift-router-namespace.yaml":            {assetsComponentsOpenshiftRouter0000_80_openshiftRouterNamespaceYaml, map[string]*bintree{}},
				"0000_80_openshift-router-service-account.yaml":      {assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceAccountYaml, map[string]*bintree{}},
				"0000_80_openshift-router-service.yaml":              {assetsComponentsOpenshiftRouter0000_80_openshiftRouterServiceYaml, map[string]*bintree{}},
				"kustomization.yaml":                                 {assetsComponentsOpenshiftRouterKustomizationYaml, map[string]*bintree{}},
			}},
			"service-ca": {nil, map[string]*bintree{
				"0000_60_service-ca_00_roles.yaml":     {assetsComponentsServiceCa0000_60_serviceCa_00_rolesYaml, map[string]*bintree{}},
				"0000_60_service-ca_01_namespace.yaml": {assetsComponentsServiceCa0000_60_serviceCa_01_namespaceYaml, map[string]*bintree{}},
				"0000_60_service-ca_04_sa.yaml":        {assetsComponentsServiceCa0000_60_serviceCa_04_saYaml, map[string]*bintree{}},
				"0000_60_service-ca_05_deploy.yaml":    {assetsComponentsServiceCa0000_60_serviceCa_05_deployYaml, map[string]*bintree{}},
				"kustomization.yaml":                   {assetsComponentsServiceCaKustomizationYaml, map[string]*bintree{}},
			}},
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
