package config

import (
	"os"
	"path/filepath"
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

// withStatMock installs a statForTrust that reports paths as root-owned and
// non-group/other-writable by default, using the real filesystem only to learn
// whether a path exists and is a directory. Entries in overrides let individual
// components report a different uid/mode so ownership failures can be exercised
// without running as root.
func withStatMock(t *testing.T, overrides map[string]fakeStat) {
	t.Helper()
	orig := statForTrust
	statForTrust = func(path string) (uint32, os.FileMode, error) {
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
	t.Cleanup(func() { statForTrust = orig })
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
			assert.Equal(t, tt.wantConfig, c.KubeletImageCredentialProviderConfigPath)
			assert.Equal(t, tt.wantBinDir, c.KubeletImageCredentialProviderBinDir)
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

// mkRootFile / mkRootDir create real filesystem objects for path/type checks;
// ownership is asserted through the stat mock, not the real files.
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

func TestValidateKubeletCredentialProvider(t *testing.T) {
	t.Run("neither set is OK", func(t *testing.T) {
		c := &Config{}
		assert.NoError(t, c.validateKubeletCredentialProvider())
	})

	t.Run("only config set", func(t *testing.T) {
		c := &Config{KubeletImageCredentialProviderConfigPath: "/etc/cp.yaml"}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be set together")
	})

	t.Run("only bin dir set", func(t *testing.T) {
		c := &Config{KubeletImageCredentialProviderBinDir: "/usr/libexec/cp"}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be set together")
	})

	t.Run("relative config path", func(t *testing.T) {
		c := &Config{
			KubeletImageCredentialProviderConfigPath: "relative/cp.yaml",
			KubeletImageCredentialProviderBinDir:     "/usr/libexec/cp",
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubelet.imageCredentialProviderConfigPath")
		assert.Contains(t, err.Error(), "must be an absolute path")
	})

	t.Run("relative bin dir", func(t *testing.T) {
		c := &Config{
			KubeletImageCredentialProviderConfigPath: "/etc/cp.yaml",
			KubeletImageCredentialProviderBinDir:     "relative/cp",
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubelet.imageCredentialProviderBinDir")
		assert.Contains(t, err.Error(), "must be an absolute path")
	})

	t.Run("missing config path", func(t *testing.T) {
		dir := t.TempDir()
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: filepath.Join(dir, "does-not-exist.yaml"),
			KubeletImageCredentialProviderBinDir:     dir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file or directory does not exist")
	})

	t.Run("missing bin dir", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     filepath.Join(dir, "missing"),
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file or directory does not exist")
	})

	t.Run("config path descends through a non-directory", func(t *testing.T) {
		dir := t.TempDir()
		file := mkFile(t, dir, "cp.yaml", 0o644)
		withStatMock(t, nil)
		c := &Config{
			// cp.yaml is a regular file, so treating it as a directory is
			// an ENOTDIR mid-path, reported as "does not exist".
			KubeletImageCredentialProviderConfigPath: filepath.Join(file, "extra"),
			KubeletImageCredentialProviderBinDir:     dir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file or directory does not exist")
	})

	t.Run("config path is a FIFO", func(t *testing.T) {
		dir := t.TempDir()
		fifo := filepath.Join(dir, "cp.fifo")
		require.NoError(t, syscall.Mkfifo(fifo, 0o644))
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: fifo,
			KubeletImageCredentialProviderBinDir:     dir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a regular file or a directory")
	})

	t.Run("config path is a regular file", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binDir := mkDir(t, dir, "bin")
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProvider())
	})

	t.Run("config path is a directory", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		binDir := mkDir(t, dir, "bin")
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgDir,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProvider())
	})

	t.Run("config path and bin dir resolve to the same directory", func(t *testing.T) {
		dir := t.TempDir()
		shared := mkDir(t, dir, "shared")
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: shared,
			KubeletImageCredentialProviderBinDir:     shared,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not resolve to the same path")
	})

	t.Run("bin dir is a file", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binFile := mkFile(t, dir, "notadir", 0o644)
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binFile,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a directory")
	})

	t.Run("valid config canonicalizes and stores canonical paths", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binDir := mkDir(t, dir, "bin")
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		require.NoError(t, c.validateKubeletCredentialProvider())
		wantCfg, _ := filepath.EvalSymlinks(cfgFile)
		wantBin, _ := filepath.EvalSymlinks(binDir)
		assert.Equal(t, wantCfg, c.KubeletImageCredentialProviderConfigPath)
		assert.Equal(t, wantBin, c.KubeletImageCredentialProviderBinDir)
	})
}

func TestValidateKubeletCredentialProviderTrustedPath(t *testing.T) {
	// newValidPair returns a config file and bin dir that both pass validation
	// when the default (all-root) stat mock is used.
	newValidPair := func(t *testing.T) (string, string) {
		dir := t.TempDir()
		return mkFile(t, dir, "cp.yaml", 0o644), mkDir(t, dir, "bin")
	}

	t.Run("non-root owner on final object", func(t *testing.T) {
		cfgFile, binDir := newValidPair(t)
		canonical, _ := filepath.EvalSymlinks(binDir)
		withStatMock(t, map[string]fakeStat{canonical: {uid: 1000, mode: 0o755}})
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonical)
		assert.Contains(t, err.Error(), "must be owned by root and not writable by group or others")
	})

	t.Run("group-writable ancestor", func(t *testing.T) {
		cfgFile, binDir := newValidPair(t)
		canonical, _ := filepath.EvalSymlinks(binDir)
		ancestor := filepath.Dir(canonical)
		withStatMock(t, map[string]fakeStat{ancestor: {uid: 0, mode: 0o775}})
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), ancestor)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("world-writable final object", func(t *testing.T) {
		cfgFile, binDir := newValidPair(t)
		canonical, _ := filepath.EvalSymlinks(binDir)
		withStatMock(t, map[string]fakeStat{canonical: {uid: 0, mode: 0o757}})
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("world-writable contained entry", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binDir := mkDir(t, dir, "bin")
		plugin := mkFile(t, binDir, "ecr-credential-provider", 0o755)
		canonicalPlugin, _ := filepath.EvalSymlinks(plugin)
		withStatMock(t, map[string]fakeStat{canonicalPlugin: {uid: 0, mode: 0o757}})
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonicalPlugin)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("compliant root-owned dir and file is OK", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		binDir := mkDir(t, dir, "bin")
		mkFile(t, binDir, "ecr-credential-provider", 0o755)
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProvider())
	})

	t.Run("symlink to compliant target resolves to canonical", func(t *testing.T) {
		dir := t.TempDir()
		realDir := mkDir(t, dir, "real-bin")
		cfgFile := mkFile(t, dir, "cp.yaml", 0o644)
		link := filepath.Join(dir, "link-bin")
		require.NoError(t, os.Symlink(realDir, link))
		withStatMock(t, nil)
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     link,
		}
		require.NoError(t, c.validateKubeletCredentialProvider())
		wantBin, _ := filepath.EvalSymlinks(realDir)
		assert.Equal(t, wantBin, c.KubeletImageCredentialProviderBinDir)
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
		withStatMock(t, map[string]fakeStat{canonicalParent: {uid: 0, mode: 0o777}})
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     binDir,
		}
		err := c.validateKubeletCredentialProvider()
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
		withStatMock(t, map[string]fakeStat{canonical: {uid: 1000, mode: 0o755}})
		c := &Config{
			KubeletImageCredentialProviderConfigPath: cfgFile,
			KubeletImageCredentialProviderBinDir:     link,
		}
		err := c.validateKubeletCredentialProvider()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be owned by root")
	})
}
