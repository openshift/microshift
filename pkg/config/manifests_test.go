package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKustomizationPaths(t *testing.T) {
	dataDir, cleanup := setupSuiteDataDir(t)
	defer cleanup()

	kustomizeDirName := func(path string) string {
		return filepath.Join(dataDir, path)
	}

	makeTestKustomize := func(path string) error {
		filename := filepath.Join(kustomizeDirName(path), "kustomization.yaml")
		return os.WriteFile(filename, []byte{}, 0600)
	}

	makeTestKustomizeDir := func(path string) error {
		return os.Mkdir(kustomizeDirName(path), 0700)
	}

	assert.NoError(t, makeTestKustomizeDir("empty"))
	assert.NoError(t, makeTestKustomizeDir("one"))
	assert.NoError(t, makeTestKustomize("one"))
	assert.NoError(t, makeTestKustomizeDir("two"))
	assert.NoError(t, makeTestKustomize("two"))

	var ttests = []struct {
		name        string
		manifests   *Manifests
		results     []string
		expectError bool
	}{
		{
			name: "empty",
			manifests: &Manifests{
				KustomizePaths: []string{},
			},
			results: []string{},
		},
		{
			name: "all",
			manifests: &Manifests{
				KustomizePaths: []string{
					kustomizeDirName("one"),
					kustomizeDirName("two"),
					kustomizeDirName("empty"),
				},
			},
			results: []string{
				kustomizeDirName("one"),
				kustomizeDirName("two"),
			},
		},
		{
			name: "o*",
			manifests: &Manifests{
				KustomizePaths: []string{
					kustomizeDirName("o*"),
				},
			},
			results: []string{
				kustomizeDirName("one"),
			},
		},
		{
			name: "*o*",
			manifests: &Manifests{
				KustomizePaths: []string{
					kustomizeDirName("*o*"),
				},
			},
			results: []string{
				kustomizeDirName("one"),
				kustomizeDirName("two"),
			},
		},
		{
			name: "nomatch",
			manifests: &Manifests{
				KustomizePaths: []string{
					kustomizeDirName("nomatch"),
				},
			},
			results: []string{},
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := tt.manifests.GetKustomizationPaths()
			if tt.expectError {
				assert.EqualError(t, err, "")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, paths, tt.results)
		})
	}
}
