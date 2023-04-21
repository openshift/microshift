package config

import (
	"fmt"
	"path/filepath"
	"sort"

	"k8s.io/klog/v2"
)

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
	// manifests. The entries in the list can be glob patterns to
	// match multiple subdirectories.
	//
	// +kubebuilder:default={"/usr/lib/microshift/manifests","/etc/microshift/manifests"}
	KustomizePaths []string `json:"kustomizePaths"`
}

func (m *Manifests) GetKustomizationPaths() ([]string, error) {
	results := []string{}
	for _, path := range m.KustomizePaths {
		pattern := filepath.Join(path, "kustomization.yaml")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("Could not understand kustomizePath value %v: %w", path, err)
		}
		if len(matches) == 0 {
			klog.Infof("No kustomize path matches %v", pattern)
			continue
		}
		// We add kustomization.yaml to the pattern so we only return
		// directories where there is something to apply, but the
		// results we need are the directory names, so convert the
		// full match string back to a directory.
		//
		// Glob() does not explicitly say it sorts its return value,
		// so we do it to ensure deterministic behavior.
		sort.Strings(matches)
		for _, match := range matches {
			klog.Infof("Adding kustomize path %v", filepath.Dir(match))
			results = append(results, filepath.Dir(match))
		}
	}
	return results, nil
}
