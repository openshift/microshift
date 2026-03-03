package schemapatch

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	yamlpatch "github.com/vmware-archive/yaml-patch"

	"k8s.io/klog/v2"
)

// loadPatch loads and parses the patch file from disk.
func loadPatch(placeholderWrapper *yamlpatch.PlaceholderWrapper, path string) (yamlpatch.Patch, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	patch, err := yamlpatch.DecodePatch(placeholderWrapper.Wrap(data))
	if err != nil {
		return nil, fmt.Errorf("could not decode patch: %v", err)
	}

	return patch, nil
}

// executeYAMLPatchForManifest applies a YAML patch to the manifest within the buffer
// and updates the buffer with the result.
func executeYAMLPatchForManifest(gc schemaPatchGenerationContext, buf *bytes.Buffer) error {
	if gc.patchPath == "" {
		return nil
	}

	manifestName := filepath.Base(gc.manifestPath)
	patchName := filepath.Base(gc.patchPath)

	klog.V(2).Infof("Patching CRD %s with patch file %s", manifestName, patchName)

	placeholderWrapper := yamlpatch.NewPlaceholderWrapper("{{", "}}")
	patch, err := loadPatch(placeholderWrapper, gc.patchPath)
	if err != nil {
		return fmt.Errorf("could not load patch %s: %v", gc.patchPath, err)
	}

	baseDoc := placeholderWrapper.Wrap(buf.Bytes())

	patchedDoc, err := patch.Apply(baseDoc)
	if err != nil {
		return fmt.Errorf("could not apply patch: %v", err)
	}

	patchedData := bytes.NewBuffer(placeholderWrapper.Unwrap(patchedDoc))
	*buf = *patchedData

	return nil
}
