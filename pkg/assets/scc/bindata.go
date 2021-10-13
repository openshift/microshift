// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml
// assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml
// assets/scc/0000_80_hostpath-provisioner-securitycontextconstraints.yaml
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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

	info := bindataFileInfo{name: "assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYaml = []byte(`kind: SecurityContextConstraints
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

func assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYamlBytes() ([]byte, error) {
	return _assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYaml, nil
}

func assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYaml() (*asset, error) {
	bytes, err := assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/scc/0000_80_hostpath-provisioner-securitycontextconstraints.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml":           assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml":       assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml": assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml":      assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml":          assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml":       assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml,
	"assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml":       assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml,
	"assets/scc/0000_80_hostpath-provisioner-securitycontextconstraints.yaml": assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYaml,
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
		"scc": {nil, map[string]*bintree{
			"0000_20_kube-apiserver-operator_00_scc-anyuid.yaml":           {assetsScc0000_20_kubeApiserverOperator_00_sccAnyuidYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml":       {assetsScc0000_20_kubeApiserverOperator_00_sccHostaccessYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml": {assetsScc0000_20_kubeApiserverOperator_00_sccHostmountAnyuidYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml":      {assetsScc0000_20_kubeApiserverOperator_00_sccHostnetworkYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-nonroot.yaml":          {assetsScc0000_20_kubeApiserverOperator_00_sccNonrootYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-privileged.yaml":       {assetsScc0000_20_kubeApiserverOperator_00_sccPrivilegedYaml, map[string]*bintree{}},
			"0000_20_kube-apiserver-operator_00_scc-restricted.yaml":       {assetsScc0000_20_kubeApiserverOperator_00_sccRestrictedYaml, map[string]*bintree{}},
			"0000_80_hostpath-provisioner-securitycontextconstraints.yaml": {assetsScc0000_80_hostpathProvisionerSecuritycontextconstraintsYaml, map[string]*bintree{}},
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
