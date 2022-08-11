// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/bindata_timestamp.txt
// assets/components/odf-lvm/csi-driver.yaml
// assets/components/odf-lvm/topolvm-controller_deployment.yaml
// assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrole.yaml
// assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
// assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_role.yaml
// assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_rolebinding.yaml
// assets/components/odf-lvm/topolvm-controller_v1_serviceaccount.yaml
// assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrole.yaml
// assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
// assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_role.yaml
// assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_rolebinding.yaml
// assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrole.yaml
// assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
// assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_role.yaml
// assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_rolebinding.yaml
// assets/components/odf-lvm/topolvm-lvmd-config_configmap_v1.yaml
// assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrole.yaml
// assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
// assets/components/odf-lvm/topolvm-node_daemonset.yaml
// assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrole.yaml
// assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
// assets/components/odf-lvm/topolvm-node_v1_serviceaccount.yaml
// assets/components/odf-lvm/topolvm-openshift-storage_namespace.yaml
// assets/components/odf-lvm/topolvm.cybozu.com_logicalvolumes.yaml
// assets/components/odf-lvm/topolvm_default-storage-class.yaml
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
// assets/components/ovn/clusterrole.yaml
// assets/components/ovn/clusterrolebinding.yaml
// assets/components/ovn/configmap.yaml
// assets/components/ovn/master/daemonset.yaml
// assets/components/ovn/master/serviceaccount.yaml
// assets/components/ovn/namespace.yaml
// assets/components/ovn/node/daemonset.yaml
// assets/components/ovn/node/serviceaccount.yaml
// assets/components/ovn/role.yaml
// assets/components/ovn/rolebinding.yaml
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
// assets/core/csr_approver_clusterrole.yaml
// assets/core/csr_approver_clusterrolebinding.yaml
// assets/core/namespace-openshift-infra.yaml
// assets/core/namespace-openshift-kube-controller-manager.yaml
// assets/core/namespace-security-allocation-controller-clusterrole.yaml
// assets/core/namespace-security-allocation-controller-clusterrolebinding.yaml
// assets/crd/0000_01_route.crd.yaml
// assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml
// assets/crd/0000_03_security-openshift_01_scc.crd.yaml
// assets/crd/0000_03_securityinternal-openshift_02_rangeallocation.crd.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml
// assets/version/microshift-version.yaml
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

var _assetsBindata_timestampTxt = []byte(`1658914160
`)

func assetsBindata_timestampTxtBytes() ([]byte, error) {
	return _assetsBindata_timestampTxt, nil
}

func assetsBindata_timestampTxt() (*asset, error) {
	bytes, err := assetsBindata_timestampTxtBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/bindata_timestamp.txt", size: 11, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmCsiDriverYaml = []byte(`# Source: topolvm/templates/controller/csidriver.yaml
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: topolvm.cybozu.com
spec:
  attachRequired: false
  podInfoOnMount: true
  volumeLifecycleModes:
    - Persistent
    - Ephemeral
`)

func assetsComponentsOdfLvmCsiDriverYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmCsiDriverYaml, nil
}

func assetsComponentsOdfLvmCsiDriverYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmCsiDriverYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/csi-driver.yaml", size: 247, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmController_deploymentYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: topolvm-controller
  namespace: openshift-storage
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app.kubernetes.io/name: topolvm-controller
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app.kubernetes.io/name: topolvm-controller
      name: topolvm-controller
      namespace: openshift-storage
    spec:
      containers:
      - command:
        - /topolvm-controller
        - --cert-dir=/certs
        image: {{ .ReleaseImage.odf_topolvm }}
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /healthz
            port: healthz
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 60
          successThreshold: 1
          timeoutSeconds: 3
        name: topolvm-controller
        ports:
        - containerPort: 9808
          name: healthz
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /metrics
            port: 8080
            scheme: HTTP
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 250m
            memory: 250Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: socket-dir
        - mountPath: /certs
          name: certs
      - args:
        - --csi-address=/run/topolvm/csi-topolvm.sock
        - --enable-capacity
        - --capacity-ownerref-level=2
        - --capacity-poll-interval=30s
        - --feature-gates=Topology=true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        image: {{ .ReleaseImage.ose_csi_ext_provisioner }}
        imagePullPolicy: IfNotPresent
        name: csi-provisioner
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: socket-dir
      - args:
        - --csi-address=/run/topolvm/csi-topolvm.sock
        image: {{ .ReleaseImage.ose_csi_ext_resizer }}
        imagePullPolicy: IfNotPresent
        name: csi-resizer
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: socket-dir
      - args:
        - --csi-address=/run/topolvm/csi-topolvm.sock
        image: {{ .ReleaseImage.ose_csi_livenessprobe }}
        imagePullPolicy: IfNotPresent
        name: liveness-probe
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: socket-dir
      dnsPolicy: ClusterFirst
      initContainers:
      - command:
        - /usr/bin/bash
        - -c
        - openssl req -nodes -x509 -newkey rsa:4096 -subj '/DC=self_signed_certificate'
          -keyout /certs/tls.key -out /certs/tls.crt -days 3650
        image: {{ .ReleaseImage.odf_lvm_operator }}
        imagePullPolicy: IfNotPresent
        name: self-signed-cert-generator
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /certs
          name: certs
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccountName: topolvm-controller
      terminationGracePeriodSeconds: 30
      volumes:
      - emptyDir: {}
        name: socket-dir
      - emptyDir: {}
        name: certs
`)

func assetsComponentsOdfLvmTopolvmController_deploymentYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmController_deploymentYaml, nil
}

func assetsComponentsOdfLvmTopolvmController_deploymentYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmController_deploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-controller_deployment.yaml", size: 4170, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-controller
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - get
  - list
  - watch
  - update
  - delete
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  - csidrivers
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - topolvm.cybozu.com
  resources:
  - logicalvolumes
  - logicalvolumes/status
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
`)

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYaml, nil
}

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrole.yaml", size: 698, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-controller
subjects:
- kind: ServiceAccount
  name: topolvm-controller
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml", size: 288, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: topolvm-controller
  namespace: openshift-storage
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - get
      - watch
      - list
      - delete
      - update
      - create`)

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYaml, nil
}

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_role.yaml", size: 281, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: topolvm-controller
  namespace: openshift-storage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: topolvm-controller
subjects:
  - kind: ServiceAccount
    name: topolvm-controller
    namespace: openshift-storage`)

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_rolebinding.yaml", size: 310, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: topolvm-controller
  namespace: openshift-storage`)

func assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYaml, nil
}

func assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-controller_v1_serviceaccount.yaml", size: 103, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-csi-provisioner
rules:
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - list
  - watch
  - create
  - update
  - patch
- apiGroups:
  - snapshot.storage.k8s.io
  resources:
  - volumesnapshots
  verbs:
  - get
  - list
- apiGroups:
  - snapshot.storage.k8s.io
  resources:
  - volumesnapshotcontents
  verbs:
  - get
  - list
- apiGroups:
  - storage.k8s.io
  resources:
  - csinodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - storage.k8s.io
  resources:
  - volumeattachments
  verbs:
  - get
  - list
  - watch
`)

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrole.yaml", size: 1015, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-csi-provisioner
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-csi-provisioner
subjects:
- kind: ServiceAccount
  name: topolvm-controller
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml", size: 298, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: topolvm-csi-provisioner
  namespace: openshift-storage
rules:
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - watch
  - list
  - delete
  - update
  - create
- apiGroups:
  - storage.k8s.io
  resources:
  - csistoragecapacities
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
- apiGroups:
  - apps
  resources:
  - replicasets
  verbs:
  - get
`)

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_role.yaml", size: 538, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: topolvm-csi-provisioner
  namespace: openshift-storage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: topolvm-csi-provisioner
subjects:
- kind: ServiceAccount
  name: topolvm-controller
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_rolebinding.yaml", size: 315, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-csi-resizer
rules:
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - watch
  - patch
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - list
  - watch
  - create
  - update
  - patch
`)

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrole.yaml", size: 569, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-csi-resizer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-csi-resizer
subjects:
- kind: ServiceAccount
  name: topolvm-controller
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml", size: 290, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: topolvm-csi-resizer
  namespace: openshift-storage
rules:
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - watch
  - list
  - delete
  - update
  - create
`)

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_role.yaml", size: 258, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: topolvm-csi-resizer
  namespace: openshift-storage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: topolvm-csi-resizer
subjects:
- kind: ServiceAccount
  name: topolvm-controller
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_rolebinding.yaml", size: 307, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1Yaml = []byte(`# Source: topolvm/templates/lvmd/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: lvmd
  namespace: openshift-storage
data:
  lvmd.yaml: |
    socket-name: /run/lvmd/lvmd.sock
    device-classes: 
      - default: true
        name: ssd
        spare-gb: 2
        volume-group: rhel
`)

func assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1YamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1Yaml, nil
}

func assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1Yaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-lvmd-config_configmap_v1.yaml", size: 299, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-node-scc
rules:
- apiGroups:
  - security.openshift.io
  resourceNames:
  - topolvm-node
  resources:
  - securitycontextconstraints
  verbs:
  - use
`)

func assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYaml, nil
}

func assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrole.yaml", size: 235, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-node-scc
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-node-scc
subjects:
- kind: ServiceAccount
  name: topolvm-node
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml", size: 278, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmNode_daemonsetYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: topolvm-node
  name: topolvm-node
  namespace: openshift-storage
spec:
  selector:
    matchLabels:
      app: topolvm-node
  template:
    metadata:
      labels:
        app: topolvm-node
      name: lvmcluster-sample
    spec:
      containers:
      - command:
        - /lvmd
        - --config=/etc/topolvm/lvmd.yaml
        - --container=true
        image: {{ .ReleaseImage.odf_topolvm }}  #registry.redhat.io/odf4/odf-topolvm-rhel8@sha256:bd9fb330fc35f88fae65f1598b802923c8a9716eeec8432bdf05d16bd4eced64
        imagePullPolicy: IfNotPresent
        name: lvmd
        resources:
          requests:
            cpu: 250m
            memory: 250Mi
        securityContext:
          privileged: true
          runAsUser: 0
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/lvmd
          name: lvmd-socket-dir
        - mountPath: /etc/topolvm
          name: lvmd-config-dir
      - command:
        - /topolvm-node
        - --lvmd-socket=/run/lvmd/lvmd.sock
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        image: {{ .ReleaseImage.odf_topolvm }}
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /healthz
            port: healthz
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 60
          successThreshold: 1
          timeoutSeconds: 3
        name: topolvm-node
        ports:
        - containerPort: 9808
          name: healthz
          protocol: TCP
        resources:
          requests:
            cpu: 250m
            memory: 250Mi
        securityContext:
          privileged: true
          runAsUser: 0
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: node-plugin-dir
        - mountPath: /run/lvmd
          name: lvmd-socket-dir
        - mountPath: /var/lib/kubelet/pods
          mountPropagation: Bidirectional
          name: pod-volumes-dir
        - mountPath: /var/lib/kubelet/plugins/kubernetes.io/csi
          mountPropagation: Bidirectional
          name: csi-plugin-dir
      - args:
        - --csi-address=/run/topolvm/csi-topolvm.sock
        - --kubelet-registration-path=/var/lib/kubelet/plugins/topolvm.cybozu.com/node/csi-topolvm.sock
        image: {{ .ReleaseImage.ose_csi_node_registrar }}
        imagePullPolicy: IfNotPresent
        lifecycle:
          preStop:
            exec:
              command:
              - /bin/sh
              - -c
              - rm -rf /registration/topolvm.cybozu.com /registration/topolvm.cybozu.com-reg.sock
        name: csi-registrar
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: node-plugin-dir
        - mountPath: /registration
          name: registration-dir
      - args:
        - --csi-address=/run/topolvm/csi-topolvm.sock
        image: {{ .ReleaseImage.ose_csi_livenessprobe }}
        imagePullPolicy: IfNotPresent
        name: liveness-probe
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /run/topolvm
          name: node-plugin-dir
      dnsPolicy: ClusterFirst
      hostPID: true
      initContainers:
      - command:
        - /usr/bin/bash
        - -c
        - until [ -f /etc/topolvm/lvmd.yaml ]; do echo waiting for lvmd config file;
          sleep 5; done
        image: {{ .ReleaseImage.odf_lvm_operator }}
        imagePullPolicy: IfNotPresent
        name: file-checker
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/topolvm
          name: lvmd-config-dir
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: topolvm-node
      serviceAccountName: topolvm-node
      terminationGracePeriodSeconds: 30
      volumes:
      - hostPath:
          path: /var/lib/kubelet/plugins_registry/
          type: Directory
        name: registration-dir
      - hostPath:
          path: /var/lib/kubelet/plugins/topolvm.cybozu.com/node
          type: DirectoryOrCreate
        name: node-plugin-dir
      - hostPath:
          path: /var/lib/kubelet/plugins/kubernetes.io/csi
          type: DirectoryOrCreate
        name: csi-plugin-dir
      - hostPath:
          path: /var/lib/kubelet/pods/
          type: DirectoryOrCreate
        name: pod-volumes-dir
      - name: lvmd-config-dir
        configMap:
          name: lvmd
          items:
            - key: lvmd.yaml
              path: lvmd.yaml
      - emptyDir:
          medium: Memory
        name: lvmd-socket-dir
  updateStrategy:
    rollingUpdate:
      maxSurge: 0
      maxUnavailable: 1
    type: RollingUpdate
`)

func assetsComponentsOdfLvmTopolvmNode_daemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmNode_daemonsetYaml, nil
}

func assetsComponentsOdfLvmTopolvmNode_daemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmNode_daemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-node_daemonset.yaml", size: 5221, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-node
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - topolvm.cybozu.com
  resources:
  - logicalvolumes
  - logicalvolumes/status
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
  - patch
- apiGroups:
  - storage.k8s.io
  resources:
  - csidrivers
  verbs:
  - get
  - list
  - watch
`)

func assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYaml, nil
}

func assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrole.yaml", size: 466, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-node
subjects:
- kind: ServiceAccount
  name: topolvm-node
  namespace: openshift-storage
`)

func assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, nil
}

func assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml", size: 270, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYaml = []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: topolvm-node
  namespace: openshift-storage

`)

func assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYaml, nil
}

func assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-node_v1_serviceaccount.yaml", size: 99, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: openshift-storage
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  labels:
    # ODF-LVM should not attempt to manage openshift or kube infra namespaces
    topolvm.cybozu.com/webhook: "ignore"`)

func assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYaml, nil
}

func assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm-openshift-storage_namespace.yaml", size: 293, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: logicalvolumes.topolvm.cybozu.com
spec:
  group: topolvm.cybozu.com
  names:
    kind: LogicalVolume
    listKind: LogicalVolumeList
    plural: logicalvolumes
    singular: logicalvolume
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        description: LogicalVolume is the Schema for the logicalvolumes API
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
            description: LogicalVolumeSpec defines the desired state of LogicalVolume
            properties:
              deviceClass:
                type: string
              name:
                type: string
              nodeName:
                type: string
              size:
                anyOf:
                - type: integer
                - type: string
                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                x-kubernetes-int-or-string: true
            required:
            - name
            - nodeName
            - size
            type: object
          status:
            description: LogicalVolumeStatus defines the observed state of LogicalVolume
            properties:
              code:
                description: A Code is an unsigned 32-bit error code as defined in
                  the gRPC spec.
                format: int32
                type: integer
              currentSize:
                anyOf:
                - type: integer
                - type: string
                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                x-kubernetes-int-or-string: true
              message:
                type: string
              volumeID:
                description: 'INSERT ADDITIONAL STATUS FIELD - define observed state
                  of cluster Important: Run "make" to regenerate code after modifying
                  this file'
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)

func assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYaml, nil
}

func assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm.cybozu.com_logicalvolumes.yaml", size: 3096, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOdfLvmTopolvm_defaultStorageClassYaml = []byte(`apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
  labels:
  name: topolvm-provisioner
parameters:
  csi.storage.k8s.io/fstype: xfs
provisioner: topolvm.cybozu.com
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
`)

func assetsComponentsOdfLvmTopolvm_defaultStorageClassYamlBytes() ([]byte, error) {
	return _assetsComponentsOdfLvmTopolvm_defaultStorageClassYaml, nil
}

func assetsComponentsOdfLvmTopolvm_defaultStorageClassYaml() (*asset, error) {
	bytes, err := assetsComponentsOdfLvmTopolvm_defaultStorageClassYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/odf-lvm/topolvm_default-storage-class.yaml", size: 334, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/cluster-role-binding.yaml", size: 223, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/cluster-role.yaml", size: 492, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/configmap.yaml", size: 610, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/daemonset.yaml", size: 3217, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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
    # ODF-LVM should not attempt to manage openshift or kube infra namespaces
    topolvm.cybozu.com/webhook: "ignore"
`)

func assetsComponentsOpenshiftDnsDnsNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftDnsDnsNamespaceYaml, nil
}

func assetsComponentsOpenshiftDnsDnsNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftDnsDnsNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/namespace.yaml", size: 536, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/service-account.yaml", size: 85, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/dns/service.yaml", size: 691, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/node-resolver/daemonset.yaml", size: 4823, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-dns/node-resolver/service-account.yaml", size: 95, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/cluster-role-binding.yaml", size: 329, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/cluster-role.yaml", size: 883, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/configmap.yaml", size: 168, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/deployment.yaml", size: 4746, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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
    # ODF-LVM should not attempt to manage openshift or kube infra namespaces
    topolvm.cybozu.com/webhook: "ignore"`)

func assetsComponentsOpenshiftRouterNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOpenshiftRouterNamespaceYaml, nil
}

func assetsComponentsOpenshiftRouterNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOpenshiftRouterNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/openshift-router/namespace.yaml", size: 617, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/service-account.yaml", size: 213, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/service-cloud.yaml", size: 567, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/openshift-router/service-internal.yaml", size: 727, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnClusterroleYaml = []byte(`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: openshift-ovn-kubernetes-node
rules:
- apiGroups: [""]
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - patch
- apiGroups: [""]
  resources:
  - namespaces
  - endpoints
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
- apiGroups: ["networking.k8s.io"]
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups: ["", "events.k8s.io"]
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups: [""]
  resources:
  - nodes
  verbs:
  - get
  - list
  - watch
  - patch
  - update
- apiGroups: ["k8s.ovn.org"]
  resources:
  - egressips
  verbs:
  - get
  - list
  - watch
- apiGroups: ["apiextensions.k8s.io"]
  resources:
  - customresourcedefinitions
  verbs:
    - get
    - list
    - watch
- apiGroups: ['authentication.k8s.io']
  resources: ['tokenreviews']
  verbs: ['create']
- apiGroups: ['authorization.k8s.io']
  resources: ['subjectaccessreviews']
  verbs: ['create']

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: openshift-ovn-kubernetes-controller
rules:
- apiGroups: [""]
  resources:
  - namespaces
  - nodes
  - pods
  verbs:
  - get
  - list
  - patch
  - watch
  - update
- apiGroups: [""]
  resources:
  - pods
  verbs:
  - get
  - list
  - patch
  - watch
  - delete
- apiGroups: [""]
  resources:
  - configmaps
  verbs:
  - get
  - create
  - update
  - patch
- apiGroups: [""]
  resources:
  - services
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
- apiGroups: ["networking.k8s.io"]
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups: ["", "events.k8s.io"]
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups: ["security.openshift.io"]
  resources:
  - securitycontextconstraints
  verbs:
  - use
  resourceNames:
  - privileged
- apiGroups: [""]
  resources:
  - "nodes/status"
  verbs:
  - patch
  - update
- apiGroups: ["k8s.ovn.org"]
  resources:
  - egressfirewalls
  - egressips
  - egressqoses
  verbs:
  - get
  - list
  - watch
  - update
  - patch
- apiGroups: ["cloud.network.openshift.io"]
  resources:
  - cloudprivateipconfigs
  verbs:
  - create
  - patch
  - update
  - delete
  - get
  - list
  - watch
- apiGroups: ["apiextensions.k8s.io"]
  resources:
  - customresourcedefinitions
  verbs:
    - get
    - list
    - watch
- apiGroups: ['authentication.k8s.io']
  resources: ['tokenreviews']
  verbs: ['create']
- apiGroups: ['authorization.k8s.io']
  resources: ['subjectaccessreviews']
  verbs: ['create']
`)

func assetsComponentsOvnClusterroleYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnClusterroleYaml, nil
}

func assetsComponentsOvnClusterroleYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/clusterrole.yaml", size: 2771, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnClusterrolebindingYaml = []byte(`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: openshift-ovn-kubernetes-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: openshift-ovn-kubernetes-node
subjects:
- kind: ServiceAccount
  name: ovn-kubernetes-node
  namespace: openshift-ovn-kubernetes

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: openshift-ovn-kubernetes-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: openshift-ovn-kubernetes-controller
subjects:
- kind: ServiceAccount
  name: ovn-kubernetes-controller
  namespace: openshift-ovn-kubernetes
`)

func assetsComponentsOvnClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnClusterrolebindingYaml, nil
}

func assetsComponentsOvnClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/clusterrolebinding.yaml", size: 663, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnConfigmapYaml = []byte(`---
# The ovnconfig config file. Used by both node and master processes.
kind: ConfigMap
apiVersion: v1
metadata:
  name: ovnkube-config
  namespace: openshift-ovn-kubernetes
data:
  ovnkube.conf:   |-
    [default]
    mtu="1400"
    cluster-subnets={{.ClusterCIDR}}
    encap-port="6081"
    enable-lflow-cache=false
    lflow-cache-limit-kb=870

    [kubernetes]
    service-cidrs={{.ServiceCIDR}}
    ovn-config-namespace="openshift-ovn-kubernetes"
    kubeconfig={{.KubeconfigPath}}
    host-network-namespace="openshift-host-network"
    platform-type="BareMetal"

    [ovnkubernetesfeature]
    enable-egress-ip=false
    enable-egress-firewall=false
    enable-egress-qos=false

    [gateway]
    mode=shared
    nodeport=true

    [masterha]
    election-lease-duration=137
    election-renew-deadline=107
    election-retry-period=26
`)

func assetsComponentsOvnConfigmapYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnConfigmapYaml, nil
}

func assetsComponentsOvnConfigmapYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/configmap.yaml", size: 844, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnMasterDaemonsetYaml = []byte(`---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: ovnkube-master
  namespace: openshift-ovn-kubernetes
  annotations:
    kubernetes.io/description: |
      This daemonset launches the ovn-kubernetes controller (master) networking components.
spec:
  selector:
    matchLabels:
      app: ovnkube-master
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      # by default, Deployments spin up the new pod before terminating the old one
      # but we don't want that - because ovsdb holds the lock.
      maxSurge: 0
      maxUnavailable: 1
  template:
    metadata:
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        app: ovnkube-master
        ovn-db-pod: "true"
        component: network
        type: infra
        openshift.io/component: network
        kubernetes.io/os: "linux"
    spec:
      serviceAccountName: ovn-kubernetes-controller
      hostNetwork: true
      hostPID: true
      dnsPolicy: Default
      priorityClassName: "system-cluster-critical"
      # volumes in all containers:
      # (container) -> (host)
      # /etc/openvswitch -> /var/lib/ovn/etc - ovsdb data
      # /var/lib/openvswitch -> /var/lib/ovn/data - ovsdb pki state
      # /run/openvswitch -> tmpfs - sockets
      # /env -> configmap env-overrides - debug overrides
      containers:
      # ovn-northd: convert network objects in nbdb to flows in sbdb
      - name: northd
        image: {{ .ReleaseImage.ovn_kubernetes }}
        command:
        - /bin/bash
        - -c
        - |
          set -xem
          if [[ -f /env/_master ]]; then
            set -o allexport
            source /env/_master
            set +o allexport
          fi

          quit() {
            echo "$(date -Iseconds) - stopping ovn-northd"
            OVN_MANAGE_OVSDB=no /usr/share/ovn/scripts/ovn-ctl stop_northd
            echo "$(date -Iseconds) - ovn-northd stopped"
            rm -f /var/run/ovn/ovn-northd.pid
            exit 0
          }
          # end of quit
          trap quit TERM INT

          echo "$(date -Iseconds) - starting ovn-northd"
          exec ovn-northd \
            --no-chdir "-vconsole:${OVN_LOG_LEVEL}" -vfile:off "-vPATTERN:console:%D{%Y-%m-%dT%H:%M:%S.###Z}|%05N|%c%T|%p|%m" \
            --pidfile /var/run/ovn/ovn-northd.pid &

          wait $!
        lifecycle:
          preStop:
            exec:
              command:
                - /bin/bash
                - -c
                - OVN_MANAGE_OVSDB=no /usr/share/ovn/scripts/ovn-ctl stop_northd
        env:
        - name: OVN_LOG_LEVEL
          value: info
        volumeMounts:
        - mountPath: /run/openvswitch/
          name: run-openvswitch
        - mountPath: /run/ovn/
          name: run-ovn
        - mountPath: /env
          name: env-overrides
        resources:
          requests:
            cpu: 10m
            memory: 10Mi
        terminationMessagePolicy: FallbackToLogsOnError

      # nbdb: the northbound, or logical network object DB. In raft mode
      - name: nbdb
        image: {{ .ReleaseImage.ovn_kubernetes }}
        command:
        - /bin/bash
        - -c
        - |
          set -xem
          if [[ -f /env/_master ]]; then
            set -o allexport
            source /env/_master
            set +o allexport
          fi

          quit() {
            echo "$(date -Iseconds) - stopping nbdb"
            /usr/share/ovn/scripts/ovn-ctl stop_nb_ovsdb
            echo "$(date -Iseconds) - nbdb stopped"
            rm -f /var/run/ovn/ovnnb_db.pid
            exit 0
          }
          # end of quit
          trap quit TERM INT

          bracketify() { case "$1" in *:*) echo "[$1]" ;; *) echo "$1" ;; esac }
          # initialize variables
          db="nb"
          ovn_db_file="/etc/ovn/ovn${db}_db.db"

          OVN_ARGS="--db-nb-cluster-local-port=9643 --no-monitor"

          echo "$(date -Iseconds) - starting nbdb"
          initialize="false"

          if [[ ! -e ${ovn_db_file} ]]; then
            initialize="true"
          fi

          if [[ "${initialize}" == "true" ]]; then
                exec /usr/share/ovn/scripts/ovn-ctl ${OVN_ARGS} \
                --ovn-nb-log="-vconsole:${OVN_LOG_LEVEL} -vfile:off -vPATTERN:console:%D{%Y-%m-%dT%H:%M:%S.###Z}|%05N|%c%T|%p|%m" \
                run_nb_ovsdb &

                wait $!
          else
            exec /usr/share/ovn/scripts/ovn-ctl ${OVN_ARGS} \
              --ovn-nb-log="-vconsole:${OVN_LOG_LEVEL} -vfile:off -vPATTERN:console:%D{%Y-%m-%dT%H:%M:%S.###Z}|%05N|%c%T|%p|%m" \
              run_nb_ovsdb &

              wait $!
          fi

        lifecycle:
          postStart:
            exec:
              command:
              - /bin/bash
              - -c
              - |
                set -x
                rm -f /var/run/ovn/ovnnb_db.pid

                #configure northd_probe_interval
                northd_probe_interval=${OVN_NORTHD_PROBE_INTERVAL:-10000}
                echo "Setting northd probe interval to ${northd_probe_interval} ms"
                retries=0
                current_probe_interval=0
                while [[ "${retries}" -lt 10 ]]; do
                  current_probe_interval=$(ovn-nbctl --if-exists get NB_GLOBAL . options:northd_probe_interval)
                  if [[ $? == 0 ]]; then
                    current_probe_interval=$(echo ${current_probe_interval} | tr -d '\"')
                    break
                  else
                    sleep 2
                    (( retries += 1 ))
                  fi
                done

                if [[ "${current_probe_interval}" != "${northd_probe_interval}" ]]; then
                  retries=0
                  while [[ "${retries}" -lt 10 ]]; do
                    ovn-nbctl set NB_GLOBAL . options:northd_probe_interval=${northd_probe_interval}
                    if [[ $? != 0 ]]; then
                      echo "Failed to set northd probe interval to ${northd_probe_interval}. retrying....."
                      sleep 2
                      (( retries += 1 ))
                    else
                      echo "Successfully set northd probe interval to ${northd_probe_interval} ms"
                      break
                    fi
                  done
                fi

          preStop:
            exec:
              command:
              - /bin/bash
              - -c
              - |
                echo "$(date -Iseconds) - stopping nbdb"
                /usr/share/ovn/scripts/ovn-ctl stop_nb_ovsdb
                echo "$(date -Iseconds) - nbdb stopped"
                rm -f /var/run/ovn/ovnnb_db.pid
        readinessProbe:
          timeoutSeconds: 5
          exec:
            command:
            - /bin/bash
            - -c
            - |
              set -xeo pipefail
              /usr/bin/ovn-appctl -t /var/run/ovn/ovnnb_db.ctl --timeout=5 ovsdb-server/memory-trim-on-compaction on 2>/dev/null

        env:
        - name: OVN_LOG_LEVEL
          value: info
        - name: OVN_NORTHD_PROBE_INTERVAL
          value: "5000"
        - name: K8S_NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        volumeMounts:
        - mountPath: /run/openvswitch/
          name: run-openvswitch
        - mountPath: /run/ovn/
          name: run-ovn
        - mountPath: /env
          name: env-overrides
        resources:
          requests:
            cpu: 10m
            memory: 10Mi
        terminationMessagePolicy: FallbackToLogsOnError

      # sbdb: The southbound, or flow DB. In raft mode
      - name: sbdb
        image: {{ .ReleaseImage.ovn_kubernetes }}
        command:
        - /bin/bash
        - -c
        - |
          set -xm
          if [[ -f /env/_master ]]; then
            set -o allexport
            source /env/_master
            set +o allexport
          fi

          quit() {
            echo "$(date -Iseconds) - stopping sbdb"
            /usr/share/ovn/scripts/ovn-ctl stop_sb_ovsdb
            echo "$(date -Iseconds) - sbdb stopped"
            rm -f /var/run/ovn/ovnsb_db.pid
            exit 0
          }
          # end of quit
          trap quit TERM INT

          bracketify() { case "$1" in *:*) echo "[$1]" ;; *) echo "$1" ;; esac }

          # initialize variables
          db="sb"
          ovn_db_file="/etc/ovn/ovn${db}_db.db"

          OVN_ARGS="--db-sb-cluster-local-port=9644 --no-monitor"

          echo "$(date -Iseconds) - starting sbdb "
          initialize="false"

          if [[ ! -e ${ovn_db_file} ]]; then
            initialize="true"
          fi

          if [[ "${initialize}" == "true" ]]; then
                exec /usr/share/ovn/scripts/ovn-ctl ${OVN_ARGS} \
                --ovn-sb-log="-vconsole:${OVN_LOG_LEVEL} -vfile:off -vPATTERN:console:%D{%Y-%m-%dT%H:%M:%S.###Z}|%05N|%c%T|%p|%m" \
                run_sb_ovsdb &

                wait $!
          else
            exec /usr/share/ovn/scripts/ovn-ctl ${OVN_ARGS} \
            --ovn-sb-log="-vconsole:${OVN_LOG_LEVEL} -vfile:off -vPATTERN:console:%D{%Y-%m-%dT%H:%M:%S.###Z}|%05N|%c%T|%p|%m" \
            run_sb_ovsdb &

            wait $!
          fi
        lifecycle:
          postStart:
            exec:
              command:
              - /bin/bash
              - -c
              - |
                set -x
                rm -f /var/run/ovn/ovnsb_db.pid

          preStop:
            exec:
              command:
              - /bin/bash
              - -c
              - |
                echo "$(date -Iseconds) - stopping sbdb"
                /usr/share/ovn/scripts/ovn-ctl stop_sb_ovsdb
                echo "$(date -Iseconds) - sbdb stopped"
                rm -f /var/run/ovn/ovnsb_db.pid
        readinessProbe:
          timeoutSeconds: 5
          exec:
            command:
            - /bin/bash
            - -c
            - |
              set -xeo pipefail
              /usr/bin/ovn-appctl -t /var/run/ovn/ovnsb_db.ctl --timeout=5 ovsdb-server/memory-trim-on-compaction on 2>/dev/null
        env:
        - name: OVN_LOG_LEVEL
          value: info
        - name: K8S_NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        volumeMounts:
        - mountPath: /run/openvswitch/
          name: run-openvswitch
        - mountPath: /run/ovn/
          name: run-ovn
        - mountPath: /env
          name: env-overrides
        resources:
          requests:
            cpu: 10m
            memory: 10Mi
        terminationMessagePolicy: FallbackToLogsOnError

      # ovnkube master: convert kubernetes objects in to nbdb logical network components
      - name: ovnkube-master
        image: {{ .ReleaseImage.ovn_kubernetes }}
        command:
        - /bin/bash
        - -c
        - |
          set -xe
          if [[ -f "/env/_master" ]]; then
            set -o allexport
            source "/env/_master"
            set +o allexport
          fi

          echo "I$(date "+%m%d %H:%M:%S.%N") - copy ovn-k8s-cni-overlay"
          cp -f /usr/libexec/cni/ovn-k8s-cni-overlay /cni-bin-dir/

          echo "I$(date "+%m%d %H:%M:%S.%N") - disable conntrack on geneve port"
          iptables -t raw -A PREROUTING -p udp --dport 6081 -j NOTRACK
          iptables -t raw -A OUTPUT -p udp --dport 6081 -j NOTRACK
          ip6tables -t raw -A PREROUTING -p udp --dport 6081 -j NOTRACK
          ip6tables -t raw -A OUTPUT -p udp --dport 6081 -j NOTRACK
          echo "I$(date "+%m%d %H:%M:%S.%N") - starting ovnkube-node"

          gateway_mode_flags="--gateway-mode shared --gateway-interface br-ex"

          echo "I$(date "+%m%d %H:%M:%S.%N") - ovnkube-master - start ovnkube --init-master ${K8S_NODE} --init-node ${K8S_NODE}"
          exec /usr/bin/ovnkube \
            --init-master "${K8S_NODE}" \
            --init-node "${K8S_NODE}" \
            --config-file=/run/ovnkube-config/ovnkube.conf \
            --loglevel "${OVN_KUBE_LOG_LEVEL}" \
            ${gateway_mode_flags} \
            --inactivity-probe="180000" \
            --nb-address "" \
            --sb-address "" \
            --enable-multicast \
            --disable-snat-multiple-gws \
            --acl-logging-rate-limit "20"
        lifecycle:
          preStop:
            exec:
              command: ["rm","-f","/etc/cni/net.d/10-ovn-kubernetes.conf"]
        readinessProbe:
          exec:
            command: ["test", "-f", "/etc/cni/net.d/10-ovn-kubernetes.conf"]
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        # for checking ovs-configuration service
        - mountPath: /etc/systemd/system
          name: systemd-units
          readOnly: true
        - mountPath: /run/openvswitch/
          name: run-openvswitch
        - mountPath: /run/ovn/
          name: run-ovn
        - mountPath: /run/ovnkube-config/
          name: ovnkube-config
        - mountPath: {{.KubeconfigDir}}
          name: kubeconfig
        - mountPath: /env
          name: env-overrides
        - mountPath: /etc/cni/net.d
          name: host-cni-netd
        - mountPath: /cni-bin-dir
          name: host-cni-bin
        - mountPath: /run/ovn-kubernetes/
          name: host-run-ovn-kubernetes
        - mountPath: /dev/log
          name: log-socket
        - mountPath: /var/log/ovn
          name: node-log
        - mountPath: /host
          name: host-slash
          readOnly: true
        - mountPath: /run/netns
          name: host-run-netns
          readOnly: true
          mountPropagation: HostToContainer
        - mountPath: /etc/openvswitch
          name: etc-openvswitch-node
        - mountPath: /etc/ovn/
          name: etc-openvswitch-node
        resources:
          requests:
            cpu: 10m
            memory: 60Mi
        env:
        - name: OVN_KUBE_LOG_LEVEL
          value: "4"
        - name: K8S_NODE
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        securityContext:
          privileged: true
        terminationMessagePolicy: FallbackToLogsOnError
      nodeSelector:
        kubernetes.io/os: "linux"
      volumes:
      # for checking ovs-configuration service
      - name: systemd-units
        hostPath:
          path: /etc/systemd/system
      - name: run-openvswitch
        hostPath:
          path: /var/run/openvswitch
      - name: run-ovn
        hostPath:
          path: /var/run/ovn

      # used for iptables wrapper scripts
      - name: host-slash
        hostPath:
          path: /
      - name: host-run-netns
        hostPath:
          path: /run/netns
      - name: etc-openvswitch-node
        hostPath:
          path: /etc/openvswitch
      # Used for placement of ACL audit logs
      - name: node-log
        hostPath:
          path: /var/log/ovn
      - name: log-socket
        hostPath:
          path: /dev/log
      # For CNI server
      - name: host-run-ovn-kubernetes
        hostPath:
          path: /run/ovn-kubernetes
      - name: host-cni-netd
        hostPath:
          path: "/etc/cni/net.d"
      - name: host-cni-bin
        hostPath:
          path: "/opt/cni/bin"

      - name: kubeconfig
        hostPath:
          path: {{.KubeconfigDir}}
      - name: ovnkube-config
        configMap:
          name: ovnkube-config
      - name: env-overrides
        configMap:
          name: env-overrides
          optional: true
      tolerations:
      - operator: "Exists"
`)

func assetsComponentsOvnMasterDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnMasterDaemonsetYaml, nil
}

func assetsComponentsOvnMasterDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnMasterDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/master/daemonset.yaml", size: 15481, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnMasterServiceaccountYaml = []byte(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ovn-kubernetes-controller
  namespace: openshift-ovn-kubernetes
`)

func assetsComponentsOvnMasterServiceaccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnMasterServiceaccountYaml, nil
}

func assetsComponentsOvnMasterServiceaccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnMasterServiceaccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/master/serviceaccount.yaml", size: 122, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnNamespaceYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  # NOTE: ovnkube.sh in the OVN image currently hardcodes this namespace name
  name: openshift-ovn-kubernetes
  labels:
    openshift.io/run-level: "0"
    openshift.io/cluster-monitoring: "true"
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged
  annotations:
    openshift.io/node-selector: ""
    openshift.io/description: "OVN Kubernetes components"
    workload.openshift.io/allowed: "management"
`)

func assetsComponentsOvnNamespaceYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnNamespaceYaml, nil
}

func assetsComponentsOvnNamespaceYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/namespace.yaml", size: 542, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnNodeDaemonsetYaml = []byte(`---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: ovnkube-node
  namespace: openshift-ovn-kubernetes
  annotations:
    kubernetes.io/description: |
      This daemonset launches the ovn-kubernetes per node networking components.
spec:
  selector:
    matchLabels:
      app: ovnkube-node
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 10%
  template:
    metadata:
      annotations:
        target.workload.openshift.io/management: '{"effect": "PreferredDuringScheduling"}'
      labels:
        app: ovnkube-node
        component: network
        type: infra
        openshift.io/component: network
        kubernetes.io/os: "linux"
    spec:
      serviceAccountName: ovn-kubernetes-node
      hostNetwork: true
      dnsPolicy: Default
      hostPID: true
      priorityClassName: "system-node-critical"
      # volumes in all containers:
      # (container) -> (host)
      # /etc/openvswitch -> /etc/openvswitch - ovsdb system id
      # /var/lib/openvswitch -> /var/lib/openvswitch/data - ovsdb data
      # /run/openvswitch -> tmpfs - ovsdb sockets
      # /env -> configmap env-overrides - debug overrides
      containers:
      # ovn-controller: programs the vswitch with flows from the sbdb
      - name: ovn-controller
        image: {{ .ReleaseImage.ovn_kubernetes }}
        command:
        - /bin/bash
        - -c
        - |
          set -e
          if [[ -f "/env/${K8S_NODE}" ]]; then
            set -o allexport
            source "/env/${K8S_NODE}"
            set +o allexport
          fi

          echo "$(date -Iseconds) - starting ovn-controller"
          exec ovn-controller unix:/var/run/openvswitch/db.sock -vfile:off \
            --no-chdir --pidfile=/var/run/ovn/ovn-controller.pid \
            --syslog-method="null" \
            --log-file=/var/log/ovn/acl-audit-log.log \
            -vFACILITY:"local0" \
            -vconsole:"${OVN_LOG_LEVEL}" -vconsole:"acl_log:off" \
            -vPATTERN:console:"%D{%Y-%m-%dT%H:%M:%S.###Z}|%05N|%c%T|%p|%m" \
            -vsyslog:"acl_log:info" \
            -vfile:"acl_log:info"
        securityContext:
          privileged: true
        env:
        - name: OVN_LOG_LEVEL
          value: info
        - name: K8S_NODE
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        volumeMounts:
        - mountPath: /run/openvswitch
          name: run-openvswitch
        - mountPath: /run/ovn/
          name: run-ovn
        - mountPath: /etc/openvswitch
          name: etc-openvswitch
        - mountPath: /etc/ovn/
          name: etc-openvswitch
        - mountPath: /var/lib/openvswitch
          name: var-lib-openvswitch
        - mountPath: /env
          name: env-overrides
        - mountPath: /var/log/ovn
          name: node-log
        - mountPath: /dev/log
          name: log-socket
        terminationMessagePolicy: FallbackToLogsOnError
        resources:
          requests:
            cpu: 10m
            memory: 10Mi
      nodeSelector:
        kubernetes.io/os: "linux"
      volumes:
      - name: var-lib-openvswitch
        hostPath:
          path: /var/lib/openvswitch/data
      - name: etc-openvswitch
        hostPath:
          path: /etc/openvswitch
      - name: run-openvswitch
        hostPath:
          path: /var/run/openvswitch
      - name: run-ovn
        hostPath:
          path: /var/run/ovn
      # Used for placement of ACL audit logs
      - name: node-log
        hostPath:
          path: /var/log/ovn
      - name: log-socket
        hostPath:
          path: /dev/log
      - name: env-overrides
        configMap:
          name: env-overrides
          optional: true
      tolerations:
      - operator: "Exists"
`)

func assetsComponentsOvnNodeDaemonsetYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnNodeDaemonsetYaml, nil
}

func assetsComponentsOvnNodeDaemonsetYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnNodeDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/node/daemonset.yaml", size: 3736, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnNodeServiceaccountYaml = []byte(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ovn-kubernetes-node
  namespace: openshift-ovn-kubernetes
`)

func assetsComponentsOvnNodeServiceaccountYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnNodeServiceaccountYaml, nil
}

func assetsComponentsOvnNodeServiceaccountYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnNodeServiceaccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/node/serviceaccount.yaml", size: 116, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnRoleYaml = []byte(`---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: openshift-ovn-kubernetes-node
  namespace: openshift-ovn-kubernetes
rules:
- apiGroups: [""]
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups: [certificates.k8s.io]
  resources: ['certificatesigningrequests']
  verbs:
    - create
    - get
    - delete
    - update
    - list

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: openshift-ovn-kubernetes-sbdb
  namespace: openshift-ovn-kubernetes
rules:
- apiGroups: [""]
  resources:
  - endpoints
  verbs:
  - create
  - update
  - patch
`)

func assetsComponentsOvnRoleYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnRoleYaml, nil
}

func assetsComponentsOvnRoleYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/role.yaml", size: 615, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsComponentsOvnRolebindingYaml = []byte(`---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: openshift-ovn-kubernetes-node
  namespace: openshift-ovn-kubernetes
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: openshift-ovn-kubernetes-node
subjects:
- kind: ServiceAccount
  name: ovn-kubernetes-node
  namespace: openshift-ovn-kubernetes

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: openshift-ovn-kubernetes-sbdb
  namespace: openshift-ovn-kubernetes
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: openshift-ovn-kubernetes-sbdb
subjects:
- kind: ServiceAccount
  name: ovn-kubernetes-controller
  namespace: openshift-ovn-kubernetes
`)

func assetsComponentsOvnRolebindingYamlBytes() ([]byte, error) {
	return _assetsComponentsOvnRolebindingYaml, nil
}

func assetsComponentsOvnRolebindingYaml() (*asset, error) {
	bytes, err := assetsComponentsOvnRolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/ovn/rolebinding.yaml", size: 699, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/clusterrole.yaml", size: 864, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/clusterrolebinding.yaml", size: 298, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/deployment.yaml", size: 1866, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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
  labels:
    # ODF-LVM should not attempt to manage openshift or kube infra namespaces
    topolvm.cybozu.com/webhook: "ignore"
`)

func assetsComponentsServiceCaNsYamlBytes() ([]byte, error) {
	return _assetsComponentsServiceCaNsYaml, nil
}

func assetsComponentsServiceCaNsYaml() (*asset, error) {
	bytes, err := assetsComponentsServiceCaNsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/components/service-ca/ns.yaml", size: 297, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/role.yaml", size: 634, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/rolebinding.yaml", size: 343, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/sa.yaml", size: 99, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/signing-cabundle.yaml", size: 123, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/components/service-ca/signing-secret.yaml", size: 144, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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
    # ODF-LVM should not attempt to manage openshift or kube infra namespaces
    topolvm.cybozu.com/webhook: "ignore"
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

	info := bindataFileInfo{name: "assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml", size: 373, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCoreCsr_approver_clusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: system:openshift:controller:cluster-csr-approver-controller
rules:
  - apiGroups:
    - certificates.k8s.io
    resources:
    - certificatesigningrequests
    verbs:
    - get
    - list
    - watch
  - apiGroups:
    - certificates.k8s.io
    resources:
    - certificatesigningrequests/approval
    verbs:
    - update
  - apiGroups:
    - certificates.k8s.io
    resources:
    - signers
    resourceNames:
    - kubernetes.io/kube-apiserver-client
    verbs:
    - approve
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
      - patch
      - update
`)

func assetsCoreCsr_approver_clusterroleYamlBytes() ([]byte, error) {
	return _assetsCoreCsr_approver_clusterroleYaml, nil
}

func assetsCoreCsr_approver_clusterroleYaml() (*asset, error) {
	bytes, err := assetsCoreCsr_approver_clusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/csr_approver_clusterrole.yaml", size: 737, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCoreCsr_approver_clusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: system:openshift:controller:cluster-csr-approver-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:controller:cluster-csr-approver-controller
subjects:
  - kind: ServiceAccount
    name: cluster-csr-approver-controller
    namespace: openshift-infra`)

func assetsCoreCsr_approver_clusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsCoreCsr_approver_clusterrolebindingYaml, nil
}

func assetsCoreCsr_approver_clusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsCoreCsr_approver_clusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/csr_approver_clusterrolebinding.yaml", size: 457, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCoreNamespaceOpenshiftInfraYaml = []byte(`kind: Namespace
apiVersion: v1
metadata:
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  name: openshift-infra
  labels:
    # set value to avoid depending on kube admission that depends on openshift apis
    openshift.io/run-level: "0"
    # allow openshift-monitoring to look for ServiceMonitor objects in this namespace
    openshift.io/cluster-monitoring: "true"
    # ODF-LVM should not attempt to manage openshift or kube infra namespaces
    topolvm.cybozu.com/webhook: "ignore"`)

func assetsCoreNamespaceOpenshiftInfraYamlBytes() ([]byte, error) {
	return _assetsCoreNamespaceOpenshiftInfraYaml, nil
}

func assetsCoreNamespaceOpenshiftInfraYaml() (*asset, error) {
	bytes, err := assetsCoreNamespaceOpenshiftInfraYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/namespace-openshift-infra.yaml", size: 537, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCoreNamespaceOpenshiftKubeControllerManagerYaml = []byte(`apiVersion: v1
kind: Namespace
metadata:
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
  name: openshift-kube-controller-manager
  labels:
    openshift.io/run-level: "0"
    openshift.io/cluster-monitoring: "true"
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged
`)

func assetsCoreNamespaceOpenshiftKubeControllerManagerYamlBytes() ([]byte, error) {
	return _assetsCoreNamespaceOpenshiftKubeControllerManagerYaml, nil
}

func assetsCoreNamespaceOpenshiftKubeControllerManagerYaml() (*asset, error) {
	bytes, err := assetsCoreNamespaceOpenshiftKubeControllerManagerYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/namespace-openshift-kube-controller-manager.yaml", size: 415, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCoreNamespaceSecurityAllocationControllerClusterroleYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  creationTimestamp: null
  name: system:openshift:controller:namespace-security-allocation-controller
rules:
- apiGroups:
  - security.openshift.io
  - security.internal.openshift.io
  resources:
  - rangeallocations
  verbs:
  - create
  - get
  - update
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - update
  - watch
  - patch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
`)

func assetsCoreNamespaceSecurityAllocationControllerClusterroleYamlBytes() ([]byte, error) {
	return _assetsCoreNamespaceSecurityAllocationControllerClusterroleYaml, nil
}

func assetsCoreNamespaceSecurityAllocationControllerClusterroleYaml() (*asset, error) {
	bytes, err := assetsCoreNamespaceSecurityAllocationControllerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/namespace-security-allocation-controller-clusterrole.yaml", size: 587, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  creationTimestamp: null
  name: system:openshift:controller:namespace-security-allocation-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:controller:namespace-security-allocation-controller
subjects:
- kind: ServiceAccount
  name: namespace-security-allocation-controller
  namespace: openshift-infra`)

func assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYaml, nil
}

func assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/core/namespace-security-allocation-controller-clusterrolebinding.yaml", size: 504, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_01_routeCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: routes.route.openshift.io
spec:
  group: route.openshift.io
  names:
    kind: Route
    plural: routes
    singular: route
  scope: Namespaced
  versions:
    - additionalPrinterColumns:
        - jsonPath: .status.ingress[0].host
          name: Host
          type: string
        - jsonPath: .status.ingress[0].conditions[?(@.type=="Admitted")].status
          name: Admitted
          type: string
        - jsonPath: .spec.to.name
          name: Service
          type: string
        - jsonPath: .spec.tls.type
          name: TLS
          type: string
      name: v1
      schema:
        openAPIV3Schema:
          description: "A route allows developers to expose services through an HTTP(S)
          aware load balancing and proxy layer via a public DNS entry. The route may
          further specify TLS options and a certificate, or specify a public CNAME
          that the router should also accept for HTTP and HTTPS traffic. An administrator
          typically configures their router to be visible outside the cluster firewall,
          and may also add additional security, caching, or traffic controls on the
          service content. Routers usually talk directly to the service endpoints.
          \n Once a route is created, the ` + "`" + `host` + "`" + ` field may not be changed. Generally,
          routers use the oldest route with a given host when resolving conflicts.
          \n Routers are subject to additional customization and may support additional
          controls via the annotations field. \n Because administrators may configure
          multiple routers, the route status field is used to return information to
          clients about the names and states of the route under each router. If a
          client chooses a duplicate name, for instance, the route status conditions
          are used to indicate the route cannot be chosen. \n To enable HTTP/2 ALPN
          on a route it requires a custom (non-wildcard) certificate. This prevents
          connection coalescing by clients, notably web browsers. We do not support
          HTTP/2 ALPN on routes that use the default certificate because of the risk
          of connection re-use/coalescing. Routes that do not have their own custom
          certificate will not be HTTP/2 ALPN-enabled on either the frontend or the
          backend. \n Compatibility level 1: Stable within a major release for a minimum
          of 12 months or 3 minor releases (whichever is longer)."
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
              allOf:
                - anyOf:
                    - properties:
                        path:
                          maxLength: 0
                    - not:
                        properties:
                          tls:
                            properties:
                              termination:
                                enum:
                                  - passthrough
                - anyOf:
                    - not:
                        properties:
                          host:
                            maxLength: 0
                    - not:
                        properties:
                          wildcardPolicy:
                            enum:
                              - Subdomain
              description: spec is the desired state of the route
              properties:
                alternateBackends:
                  description: alternateBackends allows up to 3 additional backends
                    to be assigned to the route. Only the Service kind is allowed, and
                    it will be defaulted to Service. Use the weight field in RouteTargetReference
                    object to specify relative preference.
                  items:
                    description: RouteTargetReference specifies the target that resolve
                      into endpoints. Only the 'Service' kind is allowed. Use 'weight'
                      field to emphasize one over others.
                    properties:
                      kind:
                        default: Service
                        description: The kind of target that the route is referring
                          to. Currently, only 'Service' is allowed
                        enum:
                          - Service
                          - ""
                        type: string
                      name:
                        description: name of the service/target that is being referred
                          to. e.g. name of the service
                        minLength: 1
                        type: string
                      weight:
                        description: weight as an integer between 0 and 256, default
                          100, that specifies the target's relative weight against other
                          target reference objects. 0 suppresses requests to this backend.
                        format: int32
                        maximum: 256
                        minimum: 0
                        type: integer
                    required:
                      - kind
                      - name
                    type: object
                  maxItems: 3
                  type: array
                host:
                  description: host is an alias/DNS that points to the service. Optional.
                    If not specified a route name will typically be automatically chosen.
                    Must follow DNS952 subdomain conventions.
                  maxLength: 253
                  pattern: ^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$
                  type: string
                path:
                  description: path that the router watches for, to route traffic for
                    to the service. Optional
                  pattern: ^/
                  type: string
                port:
                  description: If specified, the port to be used by the router. Most
                    routers will use all endpoints exposed by the service by default
                    - set this value to instruct routers which port to use.
                  properties:
                    targetPort:
                      allOf:
                        - not:
                            enum:
                              - 0
                        - not:
                            enum:
                              - ""
                      x-kubernetes-int-or-string: true
                  required:
                    - targetPort
                  type: object
                subdomain:
                  description: "subdomain is a DNS subdomain that is requested within
                  the ingress controller's domain (as a subdomain). If host is set
                  this field is ignored. An ingress controller may choose to ignore
                  this suggested name, in which case the controller will report the
                  assigned name in the status.ingress array or refuse to admit the
                  route. If this value is set and the server does not support this
                  field host will be populated automatically. Otherwise host is left
                  empty. The field may have multiple parts separated by a dot, but
                  not all ingress controllers may honor the request. This field may
                  not be changed after creation except by a user with the update routes/custom-host
                  permission. \n Example: subdomain ` + "`" + `frontend` + "`" + ` automatically receives
                  the router subdomain ` + "`" + `apps.mycluster.com` + "`" + ` to have a full hostname
                  ` + "`" + `frontend.apps.mycluster.com` + "`" + `."
                  maxLength: 253
                  pattern: ^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$
                  type: string
                tls:
                  allOf:
                    - anyOf:
                        - properties:
                            caCertificate:
                              maxLength: 0
                            certificate:
                              maxLength: 0
                            destinationCACertificate:
                              maxLength: 0
                            key:
                              maxLength: 0
                        - not:
                            properties:
                              termination:
                                enum:
                                  - passthrough
                    - anyOf:
                        - properties:
                            destinationCACertificate:
                              maxLength: 0
                        - not:
                            properties:
                              termination:
                                enum:
                                  - edge
                    - anyOf:
                        - properties:
                            insecureEdgeTerminationPolicy:
                              enum:
                                - ""
                                - None
                                - Allow
                                - Redirect
                        - not:
                            properties:
                              termination:
                                enum:
                                  - edge
                                  - reencrypt
                    - anyOf:
                        - properties:
                            insecureEdgeTerminationPolicy:
                              enum:
                                - ""
                                - None
                                - Redirect
                        - not:
                            properties:
                              termination:
                                enum:
                                  - passthrough
                  description: The tls field provides the ability to configure certificates
                    and termination for the route.
                  properties:
                    caCertificate:
                      description: caCertificate provides the cert authority certificate
                        contents
                      type: string
                    certificate:
                      description: certificate provides certificate contents. This should
                        be a single serving certificate, not a certificate chain. Do
                        not include a CA certificate.
                      type: string
                    destinationCACertificate:
                      description: destinationCACertificate provides the contents of
                        the ca certificate of the final destination.  When using reencrypt
                        termination this file should be provided in order to have routers
                        use it for health checks on the secure connection. If this field
                        is not specified, the router may provide its own destination
                        CA and perform hostname validation using the short service name
                        (service.namespace.svc), which allows infrastructure generated
                        certificates to automatically verify.
                      type: string
                    insecureEdgeTerminationPolicy:
                      description: "insecureEdgeTerminationPolicy indicates the desired
                      behavior for insecure connections to a route. While each router
                      may make its own decisions on which ports to expose, this is
                      normally port 80. \n * Allow - traffic is sent to the server
                      on the insecure port (default) * Disable - no traffic is allowed
                      on the insecure port. * Redirect - clients are redirected to
                      the secure port."
                      type: string
                    key:
                      description: key provides key file contents
                      type: string
                    termination:
                      description: "termination indicates termination type. \n * edge
                      - TLS termination is done by the router and http is used to
                      communicate with the backend (default) * passthrough - Traffic
                      is sent straight to the destination without the router providing
                      TLS termination * reencrypt - TLS termination is done by the
                      router and https is used to communicate with the backend"
                      enum:
                        - edge
                        - reencrypt
                        - passthrough
                      type: string
                  required:
                    - termination
                  type: object
                to:
                  description: to is an object the route should use as the primary backend.
                    Only the Service kind is allowed, and it will be defaulted to Service.
                    If the weight field (0-256 default 100) is set to zero, no traffic
                    will be sent to this backend.
                  properties:
                    kind:
                      default: Service
                      description: The kind of target that the route is referring to.
                        Currently, only 'Service' is allowed
                      enum:
                        - Service
                        - ""
                      type: string
                    name:
                      description: name of the service/target that is being referred
                        to. e.g. name of the service
                      minLength: 1
                      type: string
                    weight:
                      description: weight as an integer between 0 and 256, default 100,
                        that specifies the target's relative weight against other target
                        reference objects. 0 suppresses requests to this backend.
                      format: int32
                      maximum: 256
                      minimum: 0
                      type: integer
                  required:
                    - kind
                    - name
                  type: object
                wildcardPolicy:
                  default: None
                  description: Wildcard policy if any for the route. Currently only
                    'Subdomain' or 'None' is allowed.
                  enum:
                    - None
                    - Subdomain
                    - ""
                  type: string
              required:
                - to
              type: object
            status:
              description: status is the current state of the route
              properties:
                ingress:
                  description: ingress describes the places where the route may be exposed.
                    The list of ingress points may contain duplicate Host or RouterName
                    values. Routes are considered live once they are ` + "`" + `Ready` + "`" + `
                  items:
                    description: RouteIngress holds information about the places where
                      a route is exposed.
                    properties:
                      conditions:
                        description: Conditions is the state of the route, may be empty.
                        items:
                          description: RouteIngressCondition contains details for the
                            current condition of this route on a particular router.
                          properties:
                            lastTransitionTime:
                              description: RFC 3339 date and time when this condition
                                last transitioned
                              format: date-time
                              type: string
                            message:
                              description: Human readable message indicating details
                                about last transition.
                              type: string
                            reason:
                              description: (brief) reason for the condition's last transition,
                                and is usually a machine and human readable constant
                              type: string
                            status:
                              description: Status is the status of the condition. Can
                                be True, False, Unknown.
                              type: string
                            type:
                              description: Type is the type of the condition. Currently
                                only Admitted.
                              type: string
                          required:
                            - status
                            - type
                          type: object
                        type: array
                      host:
                        description: Host is the host string under which the route is
                          exposed; this value is required
                        type: string
                      routerCanonicalHostname:
                        description: CanonicalHostname is the external host name for
                          the router that can be used as a CNAME for the host requested
                          for this route. This value is optional and may not be set
                          in all cases.
                        type: string
                      routerName:
                        description: Name is a name chosen by the router to identify
                          itself; this value is required
                        type: string
                      wildcardPolicy:
                        description: Wildcard policy is the wildcard policy that was
                          allowed where this route is exposed.
                        type: string
                    type: object
                  type: array
              type: object
          required:
            - spec
          type: object
      served: true
      storage: true
      subresources:
        status: {}`)

func assetsCrd0000_01_routeCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_01_routeCrdYaml, nil
}

func assetsCrd0000_01_routeCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_01_routeCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_01_route.crd.yaml", size: 19360, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml", size: 9898, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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
    shortNames:
      - scc
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

	info := bindataFileInfo{name: "assets/crd/0000_03_security-openshift_01_scc.crd.yaml", size: 16038, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYaml = []byte(`# 0000_03_securityinternal-openshift_02_rangeallocation.crd.yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    api-approved.openshift.io: https://github.com/openshift/api/pull/751
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: rangeallocations.security.internal.openshift.io
spec:
  group: security.internal.openshift.io
  names:
    kind: RangeAllocation
    listKind: RangeAllocationList
    plural: rangeallocations
    singular: rangeallocation
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        description: "RangeAllocation is used so we can easily expose a RangeAllocation
          typed for security group This is an internal API, not intended for external
          consumption. \n Compatibility level 1: Stable within a major release for
          a minimum of 12 months or 3 minor releases (whichever is longer)."
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          data:
            description: data is a byte array representing the serialized state of
              a range allocation.  It is a bitmap with each bit set to one to represent
              a range is taken.
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          range:
            description: range is a string representing a unique label for a range
              of uids, "1000000000-2000000000/10000".
            type: string
        type: object
    served: true
    storage: true
`)

func assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYaml, nil
}

func assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_securityinternal-openshift_02_rangeallocation.crd.yaml", size: 2388, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml", size: 1048, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml", size: 1267, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml", size: 1298, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml", size: 1123, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml", size: 1166, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml", size: 1291, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml", size: 1213, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsVersionMicroshiftVersionYaml = []byte(`# Values are filled in at runtime by the VersionController
apiVersion: v1
kind: ConfigMap
metadata:
  name: microshift-version
  namespace: kube-public
data:
  major: ""
  minor: ""
  version: ""
`)

func assetsVersionMicroshiftVersionYamlBytes() ([]byte, error) {
	return _assetsVersionMicroshiftVersionYaml, nil
}

func assetsVersionMicroshiftVersionYaml() (*asset, error) {
	bytes, err := assetsVersionMicroshiftVersionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/version/microshift-version.yaml", size: 196, mode: os.FileMode(420), modTime: time.Unix(1658914160, 0)}
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
	"assets/bindata_timestamp.txt":                                                                           assetsBindata_timestampTxt,
	"assets/components/odf-lvm/csi-driver.yaml":                                                              assetsComponentsOdfLvmCsiDriverYaml,
	"assets/components/odf-lvm/topolvm-controller_deployment.yaml":                                           assetsComponentsOdfLvmTopolvmController_deploymentYaml,
	"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrole.yaml":             assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYaml,
	"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":      assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml,
	"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_role.yaml":                    assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYaml,
	"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_rolebinding.yaml":             assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYaml,
	"assets/components/odf-lvm/topolvm-controller_v1_serviceaccount.yaml":                                    assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYaml,
	"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrole.yaml":        assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYaml,
	"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml": assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml,
	"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_role.yaml":               assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYaml,
	"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_rolebinding.yaml":        assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYaml,
	"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrole.yaml":            assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYaml,
	"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":     assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml,
	"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_role.yaml":                   assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYaml,
	"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_rolebinding.yaml":            assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYaml,
	"assets/components/odf-lvm/topolvm-lvmd-config_configmap_v1.yaml":                                        assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1Yaml,
	"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrole.yaml":               assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYaml,
	"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":        assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml,
	"assets/components/odf-lvm/topolvm-node_daemonset.yaml":                                                  assetsComponentsOdfLvmTopolvmNode_daemonsetYaml,
	"assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrole.yaml":                   assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYaml,
	"assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":            assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml,
	"assets/components/odf-lvm/topolvm-node_v1_serviceaccount.yaml":                                          assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYaml,
	"assets/components/odf-lvm/topolvm-openshift-storage_namespace.yaml":                                     assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYaml,
	"assets/components/odf-lvm/topolvm.cybozu.com_logicalvolumes.yaml":                                       assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYaml,
	"assets/components/odf-lvm/topolvm_default-storage-class.yaml":                                           assetsComponentsOdfLvmTopolvm_defaultStorageClassYaml,
	"assets/components/openshift-dns/dns/cluster-role-binding.yaml":                                          assetsComponentsOpenshiftDnsDnsClusterRoleBindingYaml,
	"assets/components/openshift-dns/dns/cluster-role.yaml":                                                  assetsComponentsOpenshiftDnsDnsClusterRoleYaml,
	"assets/components/openshift-dns/dns/configmap.yaml":                                                     assetsComponentsOpenshiftDnsDnsConfigmapYaml,
	"assets/components/openshift-dns/dns/daemonset.yaml":                                                     assetsComponentsOpenshiftDnsDnsDaemonsetYaml,
	"assets/components/openshift-dns/dns/namespace.yaml":                                                     assetsComponentsOpenshiftDnsDnsNamespaceYaml,
	"assets/components/openshift-dns/dns/service-account.yaml":                                               assetsComponentsOpenshiftDnsDnsServiceAccountYaml,
	"assets/components/openshift-dns/dns/service.yaml":                                                       assetsComponentsOpenshiftDnsDnsServiceYaml,
	"assets/components/openshift-dns/node-resolver/daemonset.yaml":                                           assetsComponentsOpenshiftDnsNodeResolverDaemonsetYaml,
	"assets/components/openshift-dns/node-resolver/service-account.yaml":                                     assetsComponentsOpenshiftDnsNodeResolverServiceAccountYaml,
	"assets/components/openshift-router/cluster-role-binding.yaml":                                           assetsComponentsOpenshiftRouterClusterRoleBindingYaml,
	"assets/components/openshift-router/cluster-role.yaml":                                                   assetsComponentsOpenshiftRouterClusterRoleYaml,
	"assets/components/openshift-router/configmap.yaml":                                                      assetsComponentsOpenshiftRouterConfigmapYaml,
	"assets/components/openshift-router/deployment.yaml":                                                     assetsComponentsOpenshiftRouterDeploymentYaml,
	"assets/components/openshift-router/namespace.yaml":                                                      assetsComponentsOpenshiftRouterNamespaceYaml,
	"assets/components/openshift-router/service-account.yaml":                                                assetsComponentsOpenshiftRouterServiceAccountYaml,
	"assets/components/openshift-router/service-cloud.yaml":                                                  assetsComponentsOpenshiftRouterServiceCloudYaml,
	"assets/components/openshift-router/service-internal.yaml":                                               assetsComponentsOpenshiftRouterServiceInternalYaml,
	"assets/components/ovn/clusterrole.yaml":                                                                 assetsComponentsOvnClusterroleYaml,
	"assets/components/ovn/clusterrolebinding.yaml":                                                          assetsComponentsOvnClusterrolebindingYaml,
	"assets/components/ovn/configmap.yaml":                                                                   assetsComponentsOvnConfigmapYaml,
	"assets/components/ovn/master/daemonset.yaml":                                                            assetsComponentsOvnMasterDaemonsetYaml,
	"assets/components/ovn/master/serviceaccount.yaml":                                                       assetsComponentsOvnMasterServiceaccountYaml,
	"assets/components/ovn/namespace.yaml":                                                                   assetsComponentsOvnNamespaceYaml,
	"assets/components/ovn/node/daemonset.yaml":                                                              assetsComponentsOvnNodeDaemonsetYaml,
	"assets/components/ovn/node/serviceaccount.yaml":                                                         assetsComponentsOvnNodeServiceaccountYaml,
	"assets/components/ovn/role.yaml":                                                                        assetsComponentsOvnRoleYaml,
	"assets/components/ovn/rolebinding.yaml":                                                                 assetsComponentsOvnRolebindingYaml,
	"assets/components/service-ca/clusterrole.yaml":                                                          assetsComponentsServiceCaClusterroleYaml,
	"assets/components/service-ca/clusterrolebinding.yaml":                                                   assetsComponentsServiceCaClusterrolebindingYaml,
	"assets/components/service-ca/deployment.yaml":                                                           assetsComponentsServiceCaDeploymentYaml,
	"assets/components/service-ca/ns.yaml":                                                                   assetsComponentsServiceCaNsYaml,
	"assets/components/service-ca/role.yaml":                                                                 assetsComponentsServiceCaRoleYaml,
	"assets/components/service-ca/rolebinding.yaml":                                                          assetsComponentsServiceCaRolebindingYaml,
	"assets/components/service-ca/sa.yaml":                                                                   assetsComponentsServiceCaSaYaml,
	"assets/components/service-ca/signing-cabundle.yaml":                                                     assetsComponentsServiceCaSigningCabundleYaml,
	"assets/components/service-ca/signing-secret.yaml":                                                       assetsComponentsServiceCaSigningSecretYaml,
	"assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml":                             assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml,
	"assets/core/csr_approver_clusterrole.yaml":                                                              assetsCoreCsr_approver_clusterroleYaml,
	"assets/core/csr_approver_clusterrolebinding.yaml":                                                       assetsCoreCsr_approver_clusterrolebindingYaml,
	"assets/core/namespace-openshift-infra.yaml":                                                             assetsCoreNamespaceOpenshiftInfraYaml,
	"assets/core/namespace-openshift-kube-controller-manager.yaml":                                           assetsCoreNamespaceOpenshiftKubeControllerManagerYaml,
	"assets/core/namespace-security-allocation-controller-clusterrole.yaml":                                  assetsCoreNamespaceSecurityAllocationControllerClusterroleYaml,
	"assets/core/namespace-security-allocation-controller-clusterrolebinding.yaml":                           assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYaml,
	"assets/crd/0000_01_route.crd.yaml":                                                                      assetsCrd0000_01_routeCrdYaml,
	"assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml":                          assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml,
	"assets/crd/0000_03_security-openshift_01_scc.crd.yaml":                                                  assetsCrd0000_03_securityOpenshift_01_sccCrdYaml,
	"assets/crd/0000_03_securityinternal-openshift_02_rangeallocation.crd.yaml":                              assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml":                                          assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml":                                      assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml":                                assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml":                                     assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml":                                         assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml":                                      assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml":                                      assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml,
	"assets/version/microshift-version.yaml":                                                                 assetsVersionMicroshiftVersionYaml,
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
		"bindata_timestamp.txt": {assetsBindata_timestampTxt, map[string]*bintree{}},
		"components": {nil, map[string]*bintree{
			"odf-lvm": {nil, map[string]*bintree{
				"csi-driver.yaml":                    {assetsComponentsOdfLvmCsiDriverYaml, map[string]*bintree{}},
				"topolvm-controller_deployment.yaml": {assetsComponentsOdfLvmTopolvmController_deploymentYaml, map[string]*bintree{}},
				"topolvm-controller_rbac.authorization.k8s.io_v1_clusterrole.yaml":             {assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterroleYaml, map[string]*bintree{}},
				"topolvm-controller_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":      {assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, map[string]*bintree{}},
				"topolvm-controller_rbac.authorization.k8s.io_v1_role.yaml":                    {assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_roleYaml, map[string]*bintree{}},
				"topolvm-controller_rbac.authorization.k8s.io_v1_rolebinding.yaml":             {assetsComponentsOdfLvmTopolvmController_rbacAuthorizationK8sIo_v1_rolebindingYaml, map[string]*bintree{}},
				"topolvm-controller_v1_serviceaccount.yaml":                                    {assetsComponentsOdfLvmTopolvmController_v1_serviceaccountYaml, map[string]*bintree{}},
				"topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrole.yaml":        {assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterroleYaml, map[string]*bintree{}},
				"topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml": {assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, map[string]*bintree{}},
				"topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_role.yaml":               {assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_roleYaml, map[string]*bintree{}},
				"topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_rolebinding.yaml":        {assetsComponentsOdfLvmTopolvmCsiProvisioner_rbacAuthorizationK8sIo_v1_rolebindingYaml, map[string]*bintree{}},
				"topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrole.yaml":            {assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterroleYaml, map[string]*bintree{}},
				"topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":     {assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, map[string]*bintree{}},
				"topolvm-csi-resizer_rbac.authorization.k8s.io_v1_role.yaml":                   {assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_roleYaml, map[string]*bintree{}},
				"topolvm-csi-resizer_rbac.authorization.k8s.io_v1_rolebinding.yaml":            {assetsComponentsOdfLvmTopolvmCsiResizer_rbacAuthorizationK8sIo_v1_rolebindingYaml, map[string]*bintree{}},
				"topolvm-lvmd-config_configmap_v1.yaml":                                        {assetsComponentsOdfLvmTopolvmLvmdConfig_configmap_v1Yaml, map[string]*bintree{}},
				"topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrole.yaml":               {assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterroleYaml, map[string]*bintree{}},
				"topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":        {assetsComponentsOdfLvmTopolvmNodeScc_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, map[string]*bintree{}},
				"topolvm-node_daemonset.yaml":                                                  {assetsComponentsOdfLvmTopolvmNode_daemonsetYaml, map[string]*bintree{}},
				"topolvm-node_rbac.authorization.k8s.io_v1_clusterrole.yaml":                   {assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterroleYaml, map[string]*bintree{}},
				"topolvm-node_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml":            {assetsComponentsOdfLvmTopolvmNode_rbacAuthorizationK8sIo_v1_clusterrolebindingYaml, map[string]*bintree{}},
				"topolvm-node_v1_serviceaccount.yaml":                                          {assetsComponentsOdfLvmTopolvmNode_v1_serviceaccountYaml, map[string]*bintree{}},
				"topolvm-openshift-storage_namespace.yaml":                                     {assetsComponentsOdfLvmTopolvmOpenshiftStorage_namespaceYaml, map[string]*bintree{}},
				"topolvm.cybozu.com_logicalvolumes.yaml":                                       {assetsComponentsOdfLvmTopolvmCybozuCom_logicalvolumesYaml, map[string]*bintree{}},
				"topolvm_default-storage-class.yaml":                                           {assetsComponentsOdfLvmTopolvm_defaultStorageClassYaml, map[string]*bintree{}},
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
			"ovn": {nil, map[string]*bintree{
				"clusterrole.yaml":        {assetsComponentsOvnClusterroleYaml, map[string]*bintree{}},
				"clusterrolebinding.yaml": {assetsComponentsOvnClusterrolebindingYaml, map[string]*bintree{}},
				"configmap.yaml":          {assetsComponentsOvnConfigmapYaml, map[string]*bintree{}},
				"master": {nil, map[string]*bintree{
					"daemonset.yaml":      {assetsComponentsOvnMasterDaemonsetYaml, map[string]*bintree{}},
					"serviceaccount.yaml": {assetsComponentsOvnMasterServiceaccountYaml, map[string]*bintree{}},
				}},
				"namespace.yaml": {assetsComponentsOvnNamespaceYaml, map[string]*bintree{}},
				"node": {nil, map[string]*bintree{
					"daemonset.yaml":      {assetsComponentsOvnNodeDaemonsetYaml, map[string]*bintree{}},
					"serviceaccount.yaml": {assetsComponentsOvnNodeServiceaccountYaml, map[string]*bintree{}},
				}},
				"role.yaml":        {assetsComponentsOvnRoleYaml, map[string]*bintree{}},
				"rolebinding.yaml": {assetsComponentsOvnRolebindingYaml, map[string]*bintree{}},
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
			"0000_50_cluster-openshift-controller-manager_00_namespace.yaml":   {assetsCore0000_50_clusterOpenshiftControllerManager_00_namespaceYaml, map[string]*bintree{}},
			"csr_approver_clusterrole.yaml":                                    {assetsCoreCsr_approver_clusterroleYaml, map[string]*bintree{}},
			"csr_approver_clusterrolebinding.yaml":                             {assetsCoreCsr_approver_clusterrolebindingYaml, map[string]*bintree{}},
			"namespace-openshift-infra.yaml":                                   {assetsCoreNamespaceOpenshiftInfraYaml, map[string]*bintree{}},
			"namespace-openshift-kube-controller-manager.yaml":                 {assetsCoreNamespaceOpenshiftKubeControllerManagerYaml, map[string]*bintree{}},
			"namespace-security-allocation-controller-clusterrole.yaml":        {assetsCoreNamespaceSecurityAllocationControllerClusterroleYaml, map[string]*bintree{}},
			"namespace-security-allocation-controller-clusterrolebinding.yaml": {assetsCoreNamespaceSecurityAllocationControllerClusterrolebindingYaml, map[string]*bintree{}},
		}},
		"crd": {nil, map[string]*bintree{
			"0000_01_route.crd.yaml": {assetsCrd0000_01_routeCrdYaml, map[string]*bintree{}},
			"0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": {assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml, map[string]*bintree{}},
			"0000_03_security-openshift_01_scc.crd.yaml":                         {assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, map[string]*bintree{}},
			"0000_03_securityinternal-openshift_02_rangeallocation.crd.yaml":     {assetsCrd0000_03_securityinternalOpenshift_02_rangeallocationCrdYaml, map[string]*bintree{}},
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
		"version": {nil, map[string]*bintree{
			"microshift-version.yaml": {assetsVersionMicroshiftVersionYaml, map[string]*bintree{}},
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
