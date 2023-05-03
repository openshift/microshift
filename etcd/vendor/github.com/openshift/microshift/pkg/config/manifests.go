package config

const (
	// for files managed via management system in /etc, i.e. user applications
	defaultManifestDirEtc = "/etc/microshift/manifests"
	// for files embedded in ostree. i.e. cni/other component customizations
	defaultManifestDirLib = "/usr/lib/microshift/manifests"
)

type Manifests struct {
	// The locations on the filesystem to scan for kustomization.yaml
	// files to use to load manifests. Set to a list of paths to scan
	// only those paths. Set to an empty list to disable loading
	// manifests.
	//
	// +kubebuilder:default={"/usr/lib/microshift/manifests","/etc/microshift/manifests"}
	KustomizePaths []string `json:"kustomizePaths"`
}
