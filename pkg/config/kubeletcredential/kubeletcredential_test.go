package kubeletcredential

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

// testConfig lets the moved validation tests keep their original table shape: it
// holds the two raw credential-provider inputs and forwards to validateWith,
// caching the canonical result so KubeletImageCredentialProviderPaths() reads like
// the old (*config.Config) accessor these tests were written against.
type testConfig struct {
	kubeletImageCredentialProviderConfigPathRaw string
	kubeletImageCredentialProviderBinDirRaw     string

	canonicalConfig string
	canonicalBin    string
	enabled         bool
}

func (c *testConfig) validateKubeletCredentialProviderWith(own ownershipFn) error {
	cfg, bin, enabled, err := validateWith(c.kubeletImageCredentialProviderConfigPathRaw, c.kubeletImageCredentialProviderBinDirRaw, own)
	c.canonicalConfig, c.canonicalBin, c.enabled = cfg, bin, enabled
	return err
}

func (c *testConfig) KubeletImageCredentialProviderPaths() (string, string, bool) {
	return c.canonicalConfig, c.canonicalBin, c.enabled
}

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

// cpProvider builds one provider block for buildCredentialProviderConfig. A nil
// pointer field is omitted from the YAML (so a test can exercise a missing
// required field); a non-nil pointer is emitted verbatim. name is always emitted
// (it may be empty). extra holds additional indented YAML lines, e.g. a
// tokenAttributes block.
type cpProvider struct {
	name                 string
	matchImages          *string // YAML value, e.g. `["*.example.com"]`
	defaultCacheDuration *string // e.g. `"12h"`
	apiVersion           *string // e.g. credentialprovider.kubelet.k8s.io/v1
	extra                string
}

// strptr returns a pointer to s, for setting cpProvider fields inline.
func strptr(s string) *string { return &s }

// validCPProvider returns a cpProvider with every required field set to a valid
// value, so a test can override or clear exactly one field.
func validCPProvider(name string) cpProvider {
	return cpProvider{
		name:                 name,
		matchImages:          strptr(`["*.dkr.ecr.*.amazonaws.com"]`),
		defaultCacheDuration: strptr(`"12h"`),
		apiVersion:           strptr("credentialprovider.kubelet.k8s.io/v1"),
	}
}

// buildCredentialProviderConfig renders a CredentialProviderConfig with the
// given providers, omitting any provider field left nil. The config apiVersion is
// always the sole supported value; a provider's own exec apiVersion is set via
// cpProvider.apiVersion.
func buildCredentialProviderConfig(providers ...cpProvider) string {
	var b strings.Builder
	fmt.Fprintf(&b, "apiVersion: %s\n", "kubelet.config.k8s.io/v1")
	b.WriteString("kind: CredentialProviderConfig\n")
	b.WriteString("providers:\n")
	for _, p := range providers {
		fmt.Fprintf(&b, "- name: %q\n", p.name)
		if p.matchImages != nil {
			fmt.Fprintf(&b, "  matchImages: %s\n", *p.matchImages)
		}
		if p.defaultCacheDuration != nil {
			fmt.Fprintf(&b, "  defaultCacheDuration: %s\n", *p.defaultCacheDuration)
		}
		if p.apiVersion != nil {
			fmt.Fprintf(&b, "  apiVersion: %s\n", *p.apiVersion)
		}
		b.WriteString(p.extra)
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
		c := &testConfig{}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("only config set", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &testConfig{kubeletImageCredentialProviderConfigPathRaw: "/etc/cp.yaml"}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be set together")
	})

	t.Run("only bin dir set", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &testConfig{kubeletImageCredentialProviderBinDirRaw: "/usr/libexec/cp"}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be set together")
	})

	t.Run("relative config path", func(t *testing.T) {
		own, _ := newOwnership(nil)
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("config path and bin dir resolve to the same directory", func(t *testing.T) {
		dir := t.TempDir()
		shared := mkDir(t, dir, "shared")
		own, _ := newOwnership(nil)
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("unrelated non-root entry in the config dir is ignored", func(t *testing.T) {
		// kubelet reads only .json/.yaml/.yml files, so a non-root, world-writable
		// README next to a compliant config must not block startup.
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "10-ecr.yaml", "ecr-credential-provider")
		readme := mkFile(t, cfgDir, "README", 0o644)
		canonicalReadme, _ := filepath.EvalSymlinks(readme)
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		// The override would fail the trusted-path rule if README were ever checked.
		own, _ := newOwnership(map[string]fakeStat{canonicalReadme: {uid: 1000, mode: 0o777}})
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("non-root config file in the config dir is rejected naming the file", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		cfgFile := mkConfigFile(t, cfgDir, "10-ecr.yaml", "ecr-credential-provider")
		canonicalFile, _ := filepath.EvalSymlinks(cfgFile)
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(map[string]fakeStat{canonicalFile: {uid: 1000, mode: 0o644}})
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonicalFile)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("writable subdirectory in the config dir is ignored", func(t *testing.T) {
		// A subdirectory is not a file kubelet reads, so its ownership is irrelevant.
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		mkConfigFile(t, cfgDir, "10-ecr.yaml", "ecr-credential-provider")
		sub := mkDir(t, cfgDir, "old")
		canonicalSub, _ := filepath.EvalSymlinks(sub)
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(map[string]fakeStat{canonicalSub: {uid: 1000, mode: 0o777}})
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("unrelated non-root entry in the bin dir is ignored", func(t *testing.T) {
		// kubelet only executes the declared provider binaries, so an unrelated
		// helper.sh in the bin dir must not block startup.
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		helper := mkFile(t, binDir, "helper.sh", 0o755)
		canonicalHelper, _ := filepath.EvalSymlinks(helper)
		own, _ := newOwnership(map[string]fakeStat{canonicalHelper: {uid: 1000, mode: 0o777}})
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		assert.NoError(t, c.validateKubeletCredentialProviderWith(own))
	})

	t.Run("non-root declared provider binary is rejected naming it", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		canonicalBinary, _ := filepath.EvalSymlinks(filepath.Join(binDir, "ecr-credential-provider"))
		own, _ := newOwnership(map[string]fakeStat{canonicalBinary: {uid: 1000, mode: 0o755}})
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), canonicalBinary)
		assert.Contains(t, err.Error(), "must be owned by root")
	})

	t.Run("compliant root-owned dir and file is OK", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &testConfig{
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
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     link,
		}
		require.NoError(t, c.validateKubeletCredentialProviderWith(own))
		wantBin, _ := filepath.EvalSymlinks(realDir)
		_, gotBin, _ := c.KubeletImageCredentialProviderPaths()
		assert.Equal(t, wantBin, gotBin)
	})

	t.Run("declared provider binary symlinked under an unsafe ancestor is rejected naming the ancestor", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "plugin")
		binDir := mkDir(t, dir, "bin")
		// The declared provider binary is a symlink whose target lives outside binDir,
		// under a group-writable ancestor. Resolving it and walking the target's
		// ancestors must catch the unsafe directory.
		unsafeParent := mkDir(t, dir, "unsafe")
		realPlugin := mkFile(t, unsafeParent, "plugin", 0o755)
		require.NoError(t, os.Symlink(realPlugin, filepath.Join(binDir, "plugin")))
		canonicalParent, _ := filepath.EvalSymlinks(unsafeParent)
		own, _ := newOwnership(map[string]fakeStat{canonicalParent: {uid: 0, mode: 0o775}})
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		// The "/" rule now comes from the semantic mirror, so the message reads like
		// kubelet's, with the offending provider named by field path.
		assert.Contains(t, err.Error(), `providers[0].name`)
		assert.Contains(t, err.Error(), "provider name cannot contain '/'")
	})

	t.Run("duplicate provider across two files is rejected", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		fileA := mkConfigFile(t, cfgDir, "a.yaml", "dup")
		fileB := mkConfigFile(t, cfgDir, "b.yaml", "dup")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "dup")
		own, _ := newOwnership(nil)
		c := &testConfig{
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
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		// A within-file duplicate is caught by the per-file semantic mirror
		// (field.Duplicate), mirroring kubelet; the cross-file case (below) keeps the
		// MicroShift message that names both source files.
		assert.Contains(t, err.Error(), `providers[1].name`)
		assert.Contains(t, err.Error(), `Duplicate value: "dup"`)
	})

	t.Run("valid single file and executable provider passes", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := mkConfigFile(t, dir, "cp.yaml", "ecr-credential-provider")
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "ecr-credential-provider")
		own, _ := newOwnership(nil)
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
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
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		assert.Contains(t, err.Error(), fifo)
		assert.Contains(t, err.Error(), "is not a regular file")
	})
}

func TestValidateKubeletCredentialProviderSemantics(t *testing.T) {
	// runSingleFile writes cfgYAML to one config file, installs an executable in the
	// bin dir for each name in binNames, and runs full validation. Semantic
	// validation runs before the provider-binary check, so a semantic-failure case
	// needs no bin names.
	runSingleFile := func(t *testing.T, cfgYAML string, binNames ...string) error {
		t.Helper()
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "cp.yaml")
		require.NoError(t, os.WriteFile(cfgFile, []byte(cfgYAML), 0o644))
		binDir := mkDir(t, dir, "bin")
		for _, n := range binNames {
			mkExecProvider(t, binDir, n)
		}
		own, _ := newOwnership(nil)
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgFile,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		return c.validateKubeletCredentialProviderWith(own)
	}

	t.Run("missing provider apiVersion", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.apiVersion = nil
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].apiVersion: Required value")
	})

	t.Run("unsupported provider apiVersion lists the supported versions", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.apiVersion = strptr("credentialprovider.kubelet.k8s.io/v2")
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].apiVersion")
		assert.Contains(t, err.Error(), "Unsupported value")
		assert.Contains(t, err.Error(), "credentialprovider.kubelet.k8s.io/v1")
		assert.Contains(t, err.Error(), "credentialprovider.kubelet.k8s.io/v1beta1")
		assert.Contains(t, err.Error(), "credentialprovider.kubelet.k8s.io/v1alpha1")
	})

	t.Run("missing matchImages", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.matchImages = nil
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].matchImages: Required value")
	})

	t.Run("invalid matchImages entry", func(t *testing.T) {
		p := validCPProvider("ecr")
		// "[::1" is an unterminated IPv6 host, which ParseSchemelessURL rejects.
		p.matchImages = strptr(`["[::1"]`)
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].matchImages")
		assert.Contains(t, err.Error(), "match image is invalid")
	})

	t.Run("missing defaultCacheDuration", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.defaultCacheDuration = nil
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].defaultCacheDuration: Required value")
	})

	t.Run("negative defaultCacheDuration", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.defaultCacheDuration = strptr(`"-1h"`)
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].defaultCacheDuration")
		assert.Contains(t, err.Error(), "must be greater than or equal to 0")
	})

	t.Run("empty provider name", func(t *testing.T) {
		p := validCPProvider("")
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].name")
		assert.Contains(t, err.Error(), "provider name is required")
	})

	t.Run("provider name with a space", func(t *testing.T) {
		p := validCPProvider("ecr provider")
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider name cannot contain spaces")
	})

	t.Run("provider name is a single dot", func(t *testing.T) {
		p := validCPProvider(".")
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider name cannot be '.'")
	})

	t.Run("provider name is a double dot", func(t *testing.T) {
		p := validCPProvider("..")
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider name cannot be '..'")
	})

	t.Run("tokenAttributes without serviceAccountTokenAudience", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.extra = "  tokenAttributes:\n    requireServiceAccount: true\n    cacheType: Token\n"
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].tokenAttributes.serviceAccountTokenAudience: Required value")
	})

	t.Run("tokenAttributes on a non-v1 provider apiVersion is forbidden", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.apiVersion = strptr("credentialprovider.kubelet.k8s.io/v1beta1")
		p.extra = "  tokenAttributes:\n    serviceAccountTokenAudience: aud\n    requireServiceAccount: true\n    cacheType: Token\n"
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].tokenAttributes")
		assert.Contains(t, err.Error(), "only supported for credentialprovider.kubelet.k8s.io/v1 API version")
	})

	t.Run("tokenAttributes without cacheType", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.extra = "  tokenAttributes:\n    serviceAccountTokenAudience: aud\n    requireServiceAccount: true\n"
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].tokenAttributes.cacheType: Required value")
	})

	t.Run("tokenAttributes with an unsupported cacheType", func(t *testing.T) {
		p := validCPProvider("ecr")
		p.extra = "  tokenAttributes:\n    serviceAccountTokenAudience: aud\n    requireServiceAccount: true\n    cacheType: Bogus\n"
		err := runSingleFile(t, buildCredentialProviderConfig(p))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].tokenAttributes.cacheType")
		assert.Contains(t, err.Error(), "Unsupported value")
	})

	t.Run("multi-file directory names the file with the semantic error", func(t *testing.T) {
		dir := t.TempDir()
		cfgDir := mkDir(t, dir, "cp.d")
		// a.yaml is valid; b.yaml (read second, lexicographically) omits matchImages.
		require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "a.yaml"),
			[]byte(buildCredentialProviderConfig(validCPProvider("provider-a"))), 0o644))
		bad := validCPProvider("provider-b")
		bad.matchImages = nil
		require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "b.yaml"),
			[]byte(buildCredentialProviderConfig(bad)), 0o644))
		binDir := mkDir(t, dir, "bin")
		mkExecProvider(t, binDir, "provider-a")
		mkExecProvider(t, binDir, "provider-b")
		own, _ := newOwnership(nil)
		c := &testConfig{
			kubeletImageCredentialProviderConfigPathRaw: cfgDir,
			kubeletImageCredentialProviderBinDirRaw:     binDir,
		}
		err := c.validateKubeletCredentialProviderWith(own)
		require.Error(t, err)
		canonicalB, _ := filepath.EvalSymlinks(filepath.Join(cfgDir, "b.yaml"))
		assert.Contains(t, err.Error(), canonicalB)
		assert.Contains(t, err.Error(), "providers[0].matchImages: Required value")
	})

	t.Run("fully valid config passes", func(t *testing.T) {
		err := runSingleFile(t, buildCredentialProviderConfig(validCPProvider("ecr-credential-provider")),
			"ecr-credential-provider")
		assert.NoError(t, err)
	})

	t.Run("provider with only a name fails validation before reaching kubelet", func(t *testing.T) {
		// Example from PR review: a v1 CredentialProviderConfig whose provider declares
		// only name. Previously this passed MicroShift's structural check and then made
		// kubelet os.Exit(1); it must now be rejected up front.
		err := runSingleFile(t, buildCredentialProviderConfig(cpProvider{name: "ecr"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "providers[0].apiVersion: Required value")
		assert.Contains(t, err.Error(), "providers[0].matchImages: Required value")
		assert.Contains(t, err.Error(), "providers[0].defaultCacheDuration: Required value")
	})
}
