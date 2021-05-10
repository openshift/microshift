// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/rbac/0000_00_cluster-version-operator_02_roles.yaml
// assets/rbac/0000_50_service-ca-operator_00_roles.yaml
// assets/rbac/0000_60_service-ca_00_roles.yaml
// assets/rbac/0000_70_dns-operator_00-cluster-role.yaml
// assets/rbac/0000_70_dns-operator_01-cluster-role-binding.yaml
// assets/rbac/0000_70_dns-operator_01-role-binding.yaml
// assets/rbac/0000_70_dns-operator_01-role.yaml
// assets/rbac/0000_70_dns_01-cluster-role-binding.yaml
// assets/rbac/0000_70_dns_01-cluster-role.yaml
// assets/rbac/0000_80_hostpath-provisioner-clusterrole.yaml
// assets/rbac/0000_80_hostpath-provisioner-clusterrolebinding.yaml
// assets/rbac/0000_80_openshift-router-cluster-role-binding.yaml
// assets/rbac/0000_80_openshift-router-cluster-role.yaml
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

var _assetsRbac0000_00_clusterVersionOperator_02_rolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-version-operator
roleRef:
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  namespace: openshift-cluster-version
  name: default
`)

func assetsRbac0000_00_clusterVersionOperator_02_rolesYamlBytes() ([]byte, error) {
	return _assetsRbac0000_00_clusterVersionOperator_02_rolesYaml, nil
}

func assetsRbac0000_00_clusterVersionOperator_02_rolesYaml() (*asset, error) {
	bytes, err := assetsRbac0000_00_clusterVersionOperator_02_rolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_00_cluster-version-operator_02_roles.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_50_serviceCaOperator_00_rolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:operator:service-ca-operator
roleRef:
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  namespace: openshift-service-ca-operator
  name: service-ca-operator
`)

func assetsRbac0000_50_serviceCaOperator_00_rolesYamlBytes() ([]byte, error) {
	return _assetsRbac0000_50_serviceCaOperator_00_rolesYaml, nil
}

func assetsRbac0000_50_serviceCaOperator_00_rolesYaml() (*asset, error) {
	bytes, err := assetsRbac0000_50_serviceCaOperator_00_rolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_50_service-ca-operator_00_roles.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_60_serviceCa_00_rolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:openshift:operator:service-ca
roleRef:
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  namespace: openshift-service-ca
  name: service-ca
`)

func assetsRbac0000_60_serviceCa_00_rolesYamlBytes() ([]byte, error) {
	return _assetsRbac0000_60_serviceCa_00_rolesYaml, nil
}

func assetsRbac0000_60_serviceCa_00_rolesYaml() (*asset, error) {
	bytes, err := assetsRbac0000_60_serviceCa_00_rolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_60_service-ca_00_roles.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_70_dnsOperator_00ClusterRoleYaml = []byte(`# Cluster role for the operator itself.
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: openshift-dns-operator
rules:
- apiGroups:
  - operator.openshift.io
  resources:
  - dnses
  verbs:
  - "*"

- apiGroups:
  - operator.openshift.io
  resources:
  - dnses/status
  verbs:
  - update

- apiGroups:
  - apps
  - extensions
  resources:
  - daemonsets
  verbs:
  - "*"

- apiGroups:
  - ""
  resources:
  - namespaces
  - services
  - serviceaccounts
  - configmaps
  - endpoints
  - pods
  verbs:
  - "*"

- apiGroups:
  - monitoring.coreos.com
  resources:
  - servicemonitors
  verbs:
  - create
  - update
  - get

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
  - rbac.authorization.k8s.io
  resources:
  - clusterroles
  - clusterrolebindings
  - roles
  - rolebindings
  verbs:
  - create
  - get
  - list
  - watch

- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterroles
  verbs:
  - update

- apiGroups:
  - config.openshift.io
  resources:
  - clusteroperators
  - networks
  verbs:
  - create
  - get

- apiGroups:
  - config.openshift.io
  resources:
  - clusteroperators/status
  verbs:
  - update
`)

func assetsRbac0000_70_dnsOperator_00ClusterRoleYamlBytes() ([]byte, error) {
	return _assetsRbac0000_70_dnsOperator_00ClusterRoleYaml, nil
}

func assetsRbac0000_70_dnsOperator_00ClusterRoleYaml() (*asset, error) {
	bytes, err := assetsRbac0000_70_dnsOperator_00ClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_70_dns-operator_00-cluster-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYaml = []byte(`# Binds the operator cluster role to its Service Account.
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: openshift-dns-operator
subjects:
- kind: ServiceAccount
  name: dns-operator
  namespace: openshift-dns-operator
roleRef:
  kind: ClusterRole
  apiGroup: rbac.authorization.k8s.io
  name: openshift-dns-operator
`)

func assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYaml, nil
}

func assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_70_dns-operator_01-cluster-role-binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_70_dnsOperator_01RoleBindingYaml = []byte(`# Binds the operator role to its Service Account.
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: dns-operator
  namespace: openshift-dns-operator
subjects:
- kind: ServiceAccount
  name: dns-operator
  namespace: openshift-dns-operator
roleRef:
  kind: Role
  apiGroup: rbac.authorization.k8s.io
  name: dns-operator
`)

func assetsRbac0000_70_dnsOperator_01RoleBindingYamlBytes() ([]byte, error) {
	return _assetsRbac0000_70_dnsOperator_01RoleBindingYaml, nil
}

func assetsRbac0000_70_dnsOperator_01RoleBindingYaml() (*asset, error) {
	bytes, err := assetsRbac0000_70_dnsOperator_01RoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_70_dns-operator_01-role-binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_70_dnsOperator_01RoleYaml = []byte(`# Role for the operator itself.
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: dns-operator
  namespace: openshift-dns-operator
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - endpoints
  - events
  - configmaps
  verbs:
  - "*"

- apiGroups:
  - apps
  resources:
  - daemonsets
  - services
  verbs:
  - "*"
`)

func assetsRbac0000_70_dnsOperator_01RoleYamlBytes() ([]byte, error) {
	return _assetsRbac0000_70_dnsOperator_01RoleYaml, nil
}

func assetsRbac0000_70_dnsOperator_01RoleYaml() (*asset, error) {
	bytes, err := assetsRbac0000_70_dnsOperator_01RoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_70_dns-operator_01-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_70_dns_01ClusterRoleBindingYaml = []byte(`kind: ClusterRoleBinding
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

func assetsRbac0000_70_dns_01ClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsRbac0000_70_dns_01ClusterRoleBindingYaml, nil
}

func assetsRbac0000_70_dns_01ClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsRbac0000_70_dns_01ClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_70_dns_01-cluster-role-binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_70_dns_01ClusterRoleYaml = []byte(`kind: ClusterRole
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

func assetsRbac0000_70_dns_01ClusterRoleYamlBytes() ([]byte, error) {
	return _assetsRbac0000_70_dns_01ClusterRoleYaml, nil
}

func assetsRbac0000_70_dns_01ClusterRoleYaml() (*asset, error) {
	bytes, err := assetsRbac0000_70_dns_01ClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_70_dns_01-cluster-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_80_hostpathProvisionerClusterroleYaml = []byte(`kind: ClusterRole
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

func assetsRbac0000_80_hostpathProvisionerClusterroleYamlBytes() ([]byte, error) {
	return _assetsRbac0000_80_hostpathProvisionerClusterroleYaml, nil
}

func assetsRbac0000_80_hostpathProvisionerClusterroleYaml() (*asset, error) {
	bytes, err := assetsRbac0000_80_hostpathProvisionerClusterroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_80_hostpath-provisioner-clusterrole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_80_hostpathProvisionerClusterrolebindingYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
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

func assetsRbac0000_80_hostpathProvisionerClusterrolebindingYamlBytes() ([]byte, error) {
	return _assetsRbac0000_80_hostpathProvisionerClusterrolebindingYaml, nil
}

func assetsRbac0000_80_hostpathProvisionerClusterrolebindingYaml() (*asset, error) {
	bytes, err := assetsRbac0000_80_hostpathProvisionerClusterrolebindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_80_hostpath-provisioner-clusterrolebinding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_80_openshiftRouterClusterRoleBindingYaml = []byte(`# Binds the router role to its Service Account.
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

func assetsRbac0000_80_openshiftRouterClusterRoleBindingYamlBytes() ([]byte, error) {
	return _assetsRbac0000_80_openshiftRouterClusterRoleBindingYaml, nil
}

func assetsRbac0000_80_openshiftRouterClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsRbac0000_80_openshiftRouterClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_80_openshift-router-cluster-role-binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsRbac0000_80_openshiftRouterClusterRoleYaml = []byte(`# Cluster scoped role for routers. This should be as restrictive as possible.
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

func assetsRbac0000_80_openshiftRouterClusterRoleYamlBytes() ([]byte, error) {
	return _assetsRbac0000_80_openshiftRouterClusterRoleYaml, nil
}

func assetsRbac0000_80_openshiftRouterClusterRoleYaml() (*asset, error) {
	bytes, err := assetsRbac0000_80_openshiftRouterClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/rbac/0000_80_openshift-router-cluster-role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
	"assets/rbac/0000_00_cluster-version-operator_02_roles.yaml":       assetsRbac0000_00_clusterVersionOperator_02_rolesYaml,
	"assets/rbac/0000_50_service-ca-operator_00_roles.yaml":            assetsRbac0000_50_serviceCaOperator_00_rolesYaml,
	"assets/rbac/0000_60_service-ca_00_roles.yaml":                     assetsRbac0000_60_serviceCa_00_rolesYaml,
	"assets/rbac/0000_70_dns-operator_00-cluster-role.yaml":            assetsRbac0000_70_dnsOperator_00ClusterRoleYaml,
	"assets/rbac/0000_70_dns-operator_01-cluster-role-binding.yaml":    assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYaml,
	"assets/rbac/0000_70_dns-operator_01-role-binding.yaml":            assetsRbac0000_70_dnsOperator_01RoleBindingYaml,
	"assets/rbac/0000_70_dns-operator_01-role.yaml":                    assetsRbac0000_70_dnsOperator_01RoleYaml,
	"assets/rbac/0000_70_dns_01-cluster-role-binding.yaml":             assetsRbac0000_70_dns_01ClusterRoleBindingYaml,
	"assets/rbac/0000_70_dns_01-cluster-role.yaml":                     assetsRbac0000_70_dns_01ClusterRoleYaml,
	"assets/rbac/0000_80_hostpath-provisioner-clusterrole.yaml":        assetsRbac0000_80_hostpathProvisionerClusterroleYaml,
	"assets/rbac/0000_80_hostpath-provisioner-clusterrolebinding.yaml": assetsRbac0000_80_hostpathProvisionerClusterrolebindingYaml,
	"assets/rbac/0000_80_openshift-router-cluster-role-binding.yaml":   assetsRbac0000_80_openshiftRouterClusterRoleBindingYaml,
	"assets/rbac/0000_80_openshift-router-cluster-role.yaml":           assetsRbac0000_80_openshiftRouterClusterRoleYaml,
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
		"rbac": {nil, map[string]*bintree{
			"0000_00_cluster-version-operator_02_roles.yaml":       {assetsRbac0000_00_clusterVersionOperator_02_rolesYaml, map[string]*bintree{}},
			"0000_50_service-ca-operator_00_roles.yaml":            {assetsRbac0000_50_serviceCaOperator_00_rolesYaml, map[string]*bintree{}},
			"0000_60_service-ca_00_roles.yaml":                     {assetsRbac0000_60_serviceCa_00_rolesYaml, map[string]*bintree{}},
			"0000_70_dns-operator_00-cluster-role.yaml":            {assetsRbac0000_70_dnsOperator_00ClusterRoleYaml, map[string]*bintree{}},
			"0000_70_dns-operator_01-cluster-role-binding.yaml":    {assetsRbac0000_70_dnsOperator_01ClusterRoleBindingYaml, map[string]*bintree{}},
			"0000_70_dns-operator_01-role-binding.yaml":            {assetsRbac0000_70_dnsOperator_01RoleBindingYaml, map[string]*bintree{}},
			"0000_70_dns-operator_01-role.yaml":                    {assetsRbac0000_70_dnsOperator_01RoleYaml, map[string]*bintree{}},
			"0000_70_dns_01-cluster-role-binding.yaml":             {assetsRbac0000_70_dns_01ClusterRoleBindingYaml, map[string]*bintree{}},
			"0000_70_dns_01-cluster-role.yaml":                     {assetsRbac0000_70_dns_01ClusterRoleYaml, map[string]*bintree{}},
			"0000_80_hostpath-provisioner-clusterrole.yaml":        {assetsRbac0000_80_hostpathProvisionerClusterroleYaml, map[string]*bintree{}},
			"0000_80_hostpath-provisioner-clusterrolebinding.yaml": {assetsRbac0000_80_hostpathProvisionerClusterrolebindingYaml, map[string]*bintree{}},
			"0000_80_openshift-router-cluster-role-binding.yaml":   {assetsRbac0000_80_openshiftRouterClusterRoleBindingYaml, map[string]*bintree{}},
			"0000_80_openshift-router-cluster-role.yaml":           {assetsRbac0000_80_openshiftRouterClusterRoleYaml, map[string]*bintree{}},
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
