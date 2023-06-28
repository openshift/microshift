package config

import (
	"fmt"
	"path/filepath"
	"sort"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kustomize/api/konfig"
)

const (
	// for files managed via management system in /etc, i.e. user applications
	defaultManifestDirEtc     = "/etc/microshift/manifests"
	defaultManifestDirEtcGlob = "/etc/microshift/manifests.d/*"
	// for files embedded in ostree. i.e. cni/other component customizations
	defaultManifestDirLib     = "/usr/lib/microshift/manifests"
	defaultManifestDirLibGlob = "/usr/lib/microshift/manifests.d/*"
)

type Manifests struct {
	// The locations on the filesystem to scan for kustomization
	// files to use to load manifests. Set to a list of paths to scan
	// only those paths. Set to an empty list to disable loading
	// manifests. The entries in the list can be glob patterns to
	// match multiple subdirectories.
	//
	// +kubebuilder:default={"/usr/lib/microshift/manifests","/usr/lib/microshift/manifests.d/*","/etc/microshift/manifests","/etc/microshift/manifests.d/*"}
	KustomizePaths []string `json:"kustomizePaths"`
}

// GetKustomizationPaths returns the list of configured paths for
// which there are actual kustomization files to be loaded. The paths
// are returned in the order given in the configuration file. The
// results of any glob patterns are sorted lexicographically.
func (m *Manifests) GetKustomizationPaths() ([]string, error) {
	kustomizationFileNames := konfig.RecognizedKustomizationFileNames()
	results := []string{}
	for _, path := range m.KustomizePaths {
		for _, filename := range kustomizationFileNames {
			pattern := filepath.Join(path, filename)
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("Could not understand kustomizePath value %v: %w", path, err)
			}
			if len(matches) == 0 {
				klog.Infof("No kustomize path matches %v", pattern)
				continue
			}
			// We add the filename to the pattern so we only return
			// directories where there is something to apply, but the
			// results we need are the directory names, so convert the
			// full match string back to a directory.
			//
			// Glob() does not explicitly say it sorts its return
			// value, so we do it to ensure deterministic behavior.
			sort.Strings(matches)
			for _, match := range matches {
				klog.Infof("Adding kustomize path %v", filepath.Dir(match))
				results = append(results, filepath.Dir(match))
			}
		}
	}
	return results, nil
}
