package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStat overrides the ownership/mode reported for a specific path.
type fakeStat struct {
	uid  uint32
	mode os.FileMode
}

// newOwnership returns an ownershipFn that reports paths as root-owned and
// non-group/other-writable by default, using the real filesystem only to learn
// whether a path is a directory. Entries in overrides let individual components
// report a different uid/mode so ownership failures can be exercised without
// running as root. The returned map records how many times each path was
// queried, so tests can assert the trusted-path walk is not repeated.
func newOwnership(overrides map[string]fakeStat) (ownershipFn, map[string]int) {
	calls := map[string]int{}
	fn := func(path string) (uint32, os.FileMode, error) {
		calls[path]++
		if o, ok := overrides[path]; ok {
			return o.uid, o.mode, nil
		}
		// Default: root-owned, non-group/other-writable, with a mode that
		// matches whether the real path is a directory.
		fi, err := os.Lstat(path)
		if err != nil {
			return 0, 0, err
		}
		mode := os.FileMode(0o644)
		if fi.IsDir() {
			mode = 0o755
		}
		return 0, mode, nil
	}
	return fn, calls
}

func TestReadKubeletCredentialProviderKeys(t *testing.T) {
	ttests := []struct {
		name        string
		kubelet     map[string]any
		wantConfig  string
		wantBinDir  string
		expectError string
	}{
		{
			name:    "nil map",
			kubelet: nil,
		},
		{
			name:    "keys absent",
			kubelet: map[string]any{"cpuManagerPolicy": "static"},
		},
		{
			name: "both present",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
				"imageCredentialProviderBinDir":     "/usr/libexec/cp",
			},
			wantConfig: "/etc/microshift/cp.yaml",
			wantBinDir: "/usr/libexec/cp",
		},
		{
			name: "only config present",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
			},
			wantConfig: "/etc/microshift/cp.yaml",
		},
		{
			name: "empty string is unset",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "",
				"imageCredentialProviderBinDir":     "",
			},
		},
		{
			name: "explicit null is unset",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": nil,
				"imageCredentialProviderBinDir":     nil,
			},
		},
		{
			name: "int type is rejected",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": 42,
			},
			expectError: "kubelet.imageCredentialProviderConfigPath must be a string, got int",
		},
		{
			name: "bool type is rejected",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
				"imageCredentialProviderBinDir":     true,
			},
			expectError: "kubelet.imageCredentialProviderBinDir must be a string, got bool",
		},
		{
			name: "map type is rejected",
			kubelet: map[string]any{
				"imageCredentialProviderConfigPath": map[string]any{"a": "b"},
			},
			expectError: "kubelet.imageCredentialProviderConfigPath must be a string, got map[string]interface {}",
		},
	}

	for _, tt := range ttests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Kubelet: tt.kubelet}
			err := c.readKubeletCredentialProviderKeys()
			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
				return
			}
			require.NoError(t, err)
			// Reading populates the raw fields, not the canonical exported ones.
			assert.Equal(t, tt.wantConfig, c.kubeletImageCredentialProviderConfigPathRaw)
			assert.Equal(t, tt.wantBinDir, c.kubeletImageCredentialProviderBinDirRaw)
			assert.Empty(t, c.KubeletImageCredentialProviderConfigPath)
			assert.Empty(t, c.KubeletImageCredentialProviderBinDir)
			// c.Kubelet must never be modified during reading.
			assert.Equal(t, tt.kubelet, c.Kubelet)
		})
	}
}

func TestKubeletPassthrough(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		c := &Config{Kubelet: nil}
		assert.Nil(t, c.KubeletPassthrough())
	})

	t.Run("empty map returns empty, non-nil map", func(t *testing.T) {
		c := &Config{Kubelet: map[string]any{}}
		got := c.KubeletPassthrough()
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("map with only reserved keys returns empty map", func(t *testing.T) {
		c := &Config{Kubelet: map[string]any{
			"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
			"imageCredentialProviderBinDir":     "/usr/libexec/cp",
		}}
		assert.Empty(t, c.KubeletPassthrough())
	})

	t.Run("drops exactly the reserved keys and preserves the rest", func(t *testing.T) {
		c := &Config{Kubelet: map[string]any{
			"imageCredentialProviderConfigPath": "/etc/microshift/cp.yaml",
			"imageCredentialProviderBinDir":     "/usr/libexec/cp",
			"cpuManagerPolicy":                  "static",
			"kubeReserved":                      map[string]any{"memory": "500Mi"},
		}}
		got := c.KubeletPassthrough()
		assert.Equal(t, map[string]any{
			"cpuManagerPolicy": "static",
			"kubeReserved":     map[string]any{"memory": "500Mi"},
		}, got)
		// The original map is untouched.
		assert.Contains(t, c.Kubelet, "imageCredentialProviderConfigPath")
		assert.Contains(t, c.Kubelet, "imageCredentialProviderBinDir")
	})
}

// mkFile / mkDir create real filesystem objects for path/type checks; ownership
// is asserted through the injected ownership hook, not the real files.
func mkFile(t *testing.T, dir, name string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte("x"), mode))
	return p
}

func mkDir(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.Mkdir(p, 0o755))
	return p
}

// credentialProviderConfigYAML returns a structurally valid
// CredentialProviderConfig (kubelet.config.k8s.io/v1) declaring one provider per
// name.
func credentialProviderConfigYAML(names ...string) string {
	return credentialProviderConfigYAMLAt("kubelet.config.k8s.io/v1", names...)
}

// credentialProviderConfigYAMLAt is credentialProviderConfigYAML with an explicit
// config apiVersion, so the v1beta1 and v1alpha1 code paths can be exercised.
func credentialProviderConfigYAMLAt(apiVersion string, names ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "apiVersion: %s\n", apiVersion)
	b.WriteString("kind: CredentialProviderConfig\n")
	b.WriteString("providers:\n")
	for _, n := range names {
		fmt.Fprintf(&b, "- name: %s\n", n)
		b.WriteString("  matchImages: [\"*.dkr.ecr.*.amazonaws.com\"]\n")
		b.WriteString("  defaultCacheDuration: \"12h\"\n")
		b.WriteString("  apiVersion: credentialprovider.kubelet.k8s.io/v1\n")
	}
	return b.String()
}

// mkConfigFile writes a structurally valid config file naming the given
// providers and returns its path.
func mkConfigFile(t *testing.T, dir, name string, providers ...string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(credentialProviderConfigYAML(providers...)), 0o644))
	return p
}

// mkExecProvider writes an executable file (0o755) in binDir named name, so the
// provider-binary check resolves it.
func mkExecProvider(t *testing.T, binDir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"), 0o755))
}

func TestValidateKubeletCredentialProvider(t *testing.T) {
	t.Run("neither set is OK", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &Config{}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("only config set", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &Config{kubeletImageCredentialProviderConfigPathRaw: "/etc/cp.yaml"}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be set together")
	})

	t.Run("only bin dir set", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &Config{kubeletImageCredentialProviderBinDirRaw: "/usr/libexec/cp"}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be set together")
	})

	t.Run("relative config path", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: "relative/cp.yaml",
			kubeletImageCredentialProviderBinDirRaw:     "/usr/libexec/cp",
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubelet.imageCredentialProviderConfigPath")
		assert.Contains(t, err.Error(), "must be an absolute path")
	})

	t.Run("relative bin dir", func(t *testing.T) {
		// The config path is validated first (merged loop), so it must be a valid
		// absolute path for the relative bin-dir error to be the one reported.
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     "relative/cp",
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubelet.imageCredentialProviderBinDir")
		assert.Contains(t, err.Error(), "must be an absolute path")
	})

	t.Run("missing config path", func(t *testing.T) {
		dir := t.TempDir()
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: filepath.Join(dir, "does-not-exist.yaml"),
			kubeletImageCredentialProviderBinDirRaw:     dir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file or directory does not exist")
	})

	t.Run("missing bin dir", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     filepath.Join(dir, "missing"),
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file or directory does not exist")
	})

	t.Run("config path descends through a non-directory", func(t *testing.T) {
		dir := t.TempDir()
		file := mkFile(t, dir, "cp.yaml", 0o644)
		own, _ := newOwnership(nil)
		c := &Config{
			// cp.yaml is a regular file, so treating it as a directory is
			// an ENOTDIR mid-path, reported as "does not exist".
			kubeletImageCredentialProviderConfigPathRaw: filepath.Join(file, "extra"),
			kubeletImageCredentialProviderBinDirRaw:     dir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file or directory does not exist")
	})

	t.Run("config path is a FIFO", func(t *testing.T) {
		dir := t.TempDir()
		fifo := filepath.Join(dir, "cp.fifo")
		require.NoError(t, syscall.Mkfifo(fifo, 0o644))
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: fifo,
			kubeletImageCredentialProviderBinDirRaw:     dir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a regular file or a directory")
	})

	t.Run("config path is a regular file", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("config path is a directory", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("config path and bin dir resolve to the same directory", func(t *testing.T) {
		dir := t.TempDir()
		shared := mkDir(t, dir, "shared")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: shared,
			kubeletImageCredentialProviderBinDirRaw:     shared,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not resolve to the same path")
	})

	t.Run("bin dir is a file", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binFile := mkFile(t, dir, "notadir", 0o644)
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binFile,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a directory")
	})

	t.Run("valid config canonicalizes and stores canonical paths", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		require.NoError(t, c.validateKubeletCredentialProviderWith(own))
		wantCfg, _ := filepath.EvalSymlinks(cfgFile)
		wantBin, _ := filepath.EvalSymlinks(binDir)
		gotCfg, gotBin, enabled := c.KubeletImageCredentialProviderPaths()
		assert.True(t, enabled)
		assert.Equal(t, wantCfg, gotCfg)
		assert.Equal(t, wantBin, gotBin)
	})

	t.Run("canonical paths survive a later updateComputedValues", func(t *testing.T) {
		dir := t.TempDir()
		realDir := mkDir(t, dir, "real-bin")
		mkExecProvider(t, realDir, "ecr-credential-provider")
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		// A symlinked bin dir so canonical differs from the configured value.
		link := filepath.Join(dir, "link-bin")
		require.NoError(t, os.Symlink(realDir, link))

		c := NewDefault()
		c.Kubelet = map[string]any{
			"imageCredentialProviderConfigPath": cfgFile,
			"imageCredentialProviderBinDir":     link,
		}
		require.NoError(t, c.readKubeletCredentialProviderKeys())

		own, _ := newOwnership(nil)
		require.NoError(t, c.validateKubeletCredentialProviderWith(own))
		wantBin, _ := filepath.EvalSymlinks(realDir)
		_, gotBin, enabled := c.KubeletImageCredentialProviderPaths()
		require.True(t, enabled)
		require.Equal(t, wantBin, gotBin)

		// updateComputedValues() re-reads the raw values from c.Kubelet; it must
		// not revert the canonical paths the accessor returns.
		require.NoError(t, c.updateComputedValues())
		_, gotBin, enabled = c.KubeletImageCredentialProviderPaths()
		assert.True(t, enabled)
		assert.Equal(t, wantBin, gotBin, "accessor reverted to the unresolved path after updateComputedValues")
	})
}

func TestValidateKubeletCredentialProviderTrustedPath(t *testing.T) {
	// newValidPair returns a config file and bin dir that both pass validation
	// when the default (all-root) ownership hook is used.
	newValidPair := func(t *testing.T) (string, string) {
		dir := t.TempDir()
		return mkFile(t, dir, "cp.yaml", 0o644), mkDir(t, dir, "bin")
	}

	t.Run("non-root owner on final object", func(t *testing.T) {
		cfgFile, binDir := newValidPair(t)
		canonical, _ := filepath.EvalSymlinks(binDir)
		own, _ := newOwnership(map[string]fakeStat{canonical: {uid: 1000, mode: 0o755}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonical)
		assert.Contains(t, err.Error(), "must be owned by root and not writable by group or others")
	})

	t.Run("group-writable ancestor", func(t *testing.T) {
		cfgFile, binDir := newValidPair(t)
		canonical, _ := filepath.EvalSymlinks(binDir)
		ancestor := filepath.Dir(canonical)
		own, _ := newOwnership(map[string]fakeStat{ancestor: {uid: 0, mode: 0o775}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), ancestor)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("world-writable final object", func(t *testing.T) {
		cfgFile, binDir := newValidPair(t)
		canonical, _ := filepath.EvalSymlinks(binDir)
		own, _ := newOwnership(map[string]fakeStat{canonical: {uid: 0, mode: 0o757}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("world-writable contained entry", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binDir := mkDir(t, dir, "bin")
		plugin := mkFile(t, binDir, "ecr-credential-provider", 0o755)
		canonicalPlugin, _ := filepath.EvalSymlinks(plugin)
		own, _ := newOwnership(map[string]fakeStat{canonicalPlugin: {uid: 0, mode: 0o757}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonicalPlugin)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("compliant root-owned dir and file is OK", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("symlink to compliant target resolves to canonical", func(t *testing.T) {
		dir := t.TempDir()
		realDir := mkDir(t, dir, "real-bin")
		mkExecProvider(t, realDir, "ecr-credential-provider")
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		link := filepath.Join(dir, "link-bin")
		require.NoError(t, os.Symlink(realDir, link))
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     link,
		}
		require.NoError(t, c.validateKubeletCredentialProviderWith(own))
		wantBin, _ := filepath.EvalSymlinks(realDir)
		_, gotBin, _ := c.KubeletImageCredentialProviderPaths()
		assert.Equal(t, wantBin, gotBin)
	})

	t.Run("symlinked bin-dir entry is checked at its target including ancestors", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binDir := mkDir(t, dir, "bin")
		// The real plugin lives outside binDir, under an unsafe ancestor.
		unsafeParent := mkDir(t, dir, "unsafe")
		realPlugin := mkFile(t, unsafeParent, "plugin", 0o755)
		require.NoError(t, os.Symlink(realPlugin, filepath.Join(binDir, "plugin")))
		canonicalParent, _ := filepath.EvalSymlinks(unsafeParent)
		own, _ := newOwnership(map[string]fakeStat{canonicalParent: {uid: 0, mode: 0o777}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonicalParent)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("symlink to unsafe target is rejected", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		realDir := mkDir(t, dir, "real-bin")
		link := filepath.Join(dir, "link-bin")
		require.NoError(t, os.Symlink(realDir, link))
		canonical, _ := filepath.EvalSymlinks(realDir)
		own, _ := newOwnership(map[string]fakeStat{canonical: {uid: 1000, mode: 0o755}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     link,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("each ancestor is stat'd only once across the whole validation", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "a", "b", "c")
		binDir := mkDir(t, dir, "bin")
		// Three entries under one bin dir: their shared ancestors must not be
		// re-walked once memoized.
		mkExecProvider(t, binDir, "a")
		mkExecProvider(t, binDir, "b")
		mkExecProvider(t, binDir, "c")
		own, calls := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		require.NoError(t, c.validateKubeletCredentialProviderWith(own))
		require.NotEmpty(t, calls)
		for path, n := range calls {
			assert.Equalf(t, 1, n, "path %q was stat'd %d times, expected exactly once", path, n)
		}
	})
}

func TestValidateKubeletCredentialProviderStructure(t *testing.T) {
	t.Run("directory with no matching files names the directory", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgDir)
		assert.Contains(t, err.Error(), "contains no .json, .yaml, or .yml")
	})

	t.Run("directory with only a .txt names the directory", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "readme.txt"), []byte("x"), 0o644))
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgDir)
		assert.Contains(t, err.Error(), "contains no .json, .yaml, or .yml")
	})

	t.Run("wrong kind names the file", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile, []byte(
			"apiVersion: kubelet.config.k8s.io/v1\nkind: NotThatKind\nproviders: []\n"), 0o644))
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgFile)
		assert.Contains(t, err.Error(), "is not a valid CredentialProviderConfig")
	})

	t.Run("wrong apiVersion names the file", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile, []byte(
			"apiVersion: example.com/v1\nkind: CredentialProviderConfig\nproviders: []\n"), 0o644))
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgFile)
		assert.Contains(t, err.Error(), "is not a valid CredentialProviderConfig")
	})

	t.Run("malformed YAML names the file and includes the decode error", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile, []byte("providers: [ this is : not : yaml\n"), 0o644))
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgFile)
		assert.Contains(t, err.Error(), "is not a valid CredentialProviderConfig")
	})

	t.Run("empty providers reports declares no providers", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile, []byte(
			"apiVersion: kubelet.config.k8s.io/v1\nkind: CredentialProviderConfig\nproviders: []\n"), 0o644))
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgFile)
		assert.Contains(t, err.Error(), "declares no providers")
	})

	t.Run("provider with no file in bin dir names provider, source file and joined path", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		canonicalBin, _ := filepath.EvalSymlinks(binDir)
		canonicalCfg, _ := filepath.EvalSymlinks(cfgFile)
		// The error is attributed to the bin dir, whose contents need fixing.
		assert.Contains(t, err.Error(), "kubelet.imageCredentialProviderBinDir")
		assert.Contains(t, err.Error(), `provider "ecr-credential-provider"`)
		assert.Contains(t, err.Error(), fmt.Sprintf("declared in %q", canonicalCfg))
		assert.Contains(t, err.Error(), "has no executable at")
		assert.Contains(t, err.Error(), filepath.Join(canonicalBin, "ecr-credential-provider"))
	})

	t.Run("provider file present but not executable is rejected", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		// 0o644: present but not executable, so the execute-bit check rejects it.
		require.NoError(t, os.WriteFile(filepath.Join(binDir, "ecr-credential-provider"), []byte("x"), 0o644))
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `provider "ecr-credential-provider"`)
		assert.Contains(t, err.Error(), "has no executable at")
	})

	t.Run("provider name containing a slash is rejected", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "sub/provider")
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `provider name "sub/provider" must not contain "/"`)
	})

	t.Run("duplicate provider across two files is rejected", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		fileA := mkConfigFile(t, cfgDir, "a.yaml", "dup")
		fileB := mkConfigFile(t, cfgDir, "b.yaml", "dup")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "dup")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `provider "dup" is declared more than once`)
		canonicalA, _ := filepath.EvalSymlinks(fileA)
		canonicalB, _ := filepath.EvalSymlinks(fileB)
		assert.Contains(t, err.Error(), canonicalA)
		assert.Contains(t, err.Error(), canonicalB)
	})

	t.Run("duplicate provider within one file is rejected", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "dup", "dup")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "dup")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `provider "dup" is declared more than once`)
	})

	t.Run("valid single file and executable provider passes", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("valid directory with two files and both providers present passes", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "a.yaml", "provider-a")
		mkConfigFile(t, cfgDir, "b.json", "provider-b")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "provider-a")
		mkExecProvider(t, binDir, "provider-b")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("v1beta1 config decodes and passes", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile,
			[]byte(credentialProviderConfigYAMLAt("kubelet.config.k8s.io/v1beta1", "ecr-credential-provider")), 0o644))
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("v1alpha1 config decodes and passes", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile,
			[]byte(credentialProviderConfigYAMLAt("kubelet.config.k8s.io/v1alpha1", "ecr-credential-provider")), 0o644))
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("unknown field is rejected by strict decoding", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		// matchImage (singular) is not a field of CredentialProvider; strict
		// decoding rejects it, the way kubelet does at registration.
		require.NoError(t, os.WriteFile(cfgFile, []byte(
			"apiVersion: kubelet.config.k8s.io/v1\n"+
				"kind: CredentialProviderConfig\n"+
				"providers:\n"+
				"- name: ecr-credential-provider\n"+
				"  matchImage: [\"*.example.com\"]\n"), 0o644))
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), cfgFile)
		assert.Contains(t, err.Error(), "is not a valid CredentialProviderConfig")
	})

	t.Run("world-writable bin dir wins over unresolvable provider", func(t *testing.T) {
		dir := t.TempDir()
		// Config names a provider that does not exist, but the bin dir is
		// world-writable; the trusted-path check runs first, so the permission
		// error is reported, not the provider error.
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "no-such-provider")
		binDir := mkDir(t, dir, "bin")
		canonicalBin, _ := filepath.EvalSymlinks(binDir)
		own, _ := newOwnership(map[string]fakeStat{canonicalBin: {uid: 0, mode: 0o757}})
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be owned by root and not writable by group or others")
		assert.NotContains(t, err.Error(), "has no executable")
	})

	t.Run("dangling .yaml symlink alongside a valid file is an error naming the link", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "a.yaml", "provider-a")
		// A symlink with a matching extension whose target does not exist:
		// kubelet would include it and fail at os.ReadFile, reaching os.Exit.
		// The per-entry trusted-path walk (validateDirEntries) resolves every
		// directory entry and rejects the broken link before the structural
		// stage runs, so the observable message is "does not exist"; the
		// collectCredentialProviderConfigFiles branch (asserted directly below)
		// is the same defense at the structural stage.
		dangling := filepath.Join(cfgDir, "b.yaml")
		require.NoError(t, os.Symlink(filepath.Join(dir, "nonexistent-target.yaml"), dangling))
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "provider-a")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), dangling)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("collectCredentialProviderConfigFiles rejects a dangling symlink", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "a.yaml", "provider-a")
		dangling := filepath.Join(cfgDir, "b.yaml")
		require.NoError(t, os.Symlink(filepath.Join(dir, "nonexistent-target.yaml"), dangling))
		_, err := collectCredentialProviderConfigFiles(cfgDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), dangling)
		assert.Contains(t, err.Error(), "dangling symlink")
	})

	t.Run("symlink to a directory named x.yaml is rejected as not a regular file", func(t *testing.T) {
		// DirEntry.IsDir() is false for a symlink, so kubelet does not skip a
		// symlink-to-directory: it os.ReadFile's it and fails, reaching os.Exit.
		// collectCredentialProviderConfigFiles must treat it as an error, not a
		// skip.
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "a.yaml", "provider-a")
		targetDir := mkDir(t, dir, "target-dir")
		link := filepath.Join(cfgDir, "b.yaml")
		require.NoError(t, os.Symlink(targetDir, link))
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "provider-a")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		canonicalTarget, _ := filepath.EvalSymlinks(targetDir)
		assert.Contains(t, err.Error(), canonicalTarget)
		assert.Contains(t, err.Error(), "is not a regular file")
	})

	t.Run("collectCredentialProviderConfigFiles skips a real directory named x.yaml", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		validFile := mkConfigFile(t, cfgDir, "a.yaml", "provider-a")
		// A real subdirectory named b.yaml: kubelet checks DirEntry.IsDir() and
		// skips it; so must we.
		require.NoError(t, os.Mkdir(filepath.Join(cfgDir, "b.yaml"), 0o755))
		files, err := collectCredentialProviderConfigFiles(cfgDir)
		require.NoError(t, err)
		canonicalValid, _ := filepath.EvalSymlinks(validFile)
		assert.Equal(t, []string{canonicalValid}, files)
	})

	t.Run("FIFO named x.yaml is rejected as not a regular file", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		fifo := filepath.Join(cfgDir, "x.yaml")
		require.NoError(t, syscall.Mkfifo(fifo, 0o644))
		binDir := mkDir(t, dir, "bin")
		own, _ := newOwnership(nil)
		c := &Config{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), fifo)
		assert.Contains(t, err.Error(), "is not a regular file")
	})
}
