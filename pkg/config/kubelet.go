package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	kubeletconfigv1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1"
	kubeletconfigv1alpha1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1alpha1"
	kubeletconfigv1beta1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1beta1"
)

// credentialProviderCodec returns the same strict decoder the vendored kubelet
// uses to read the credential provider configuration: see the scheme setup in
// vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/plugin.go and decode()
// in the sibling config.go. Strict decoding rejects unknown fields, and all
// three API versions kubelet accepts (v1alpha1, v1beta1, v1 of
// kubelet.config.k8s.io) are registered together with the internal type and its
// conversions. Building the decoder from the same vendored packages keeps this
// structural check from diverging from the kubelet compiled into the same
// binary; a lenient decoder would let a typo'd field through to the os.Exit at
// registration.
//
// The scheme is built lazily on first use so that importers of pkg/config that
// never touch the feature (generate-config, show-config without it) do not pay
// for four AddToScheme calls at package init.
var credentialProviderCodec = sync.OnceValue(func() runtime.Decoder {
	s := runtime.NewScheme()
	utilruntime.Must(kubeletconfig.AddToScheme(s))
	utilruntime.Must(kubeletconfigv1alpha1.AddToScheme(s))
	utilruntime.Must(kubeletconfigv1beta1.AddToScheme(s))
	utilruntime.Must(kubeletconfigv1.AddToScheme(s))
	return serializer.NewCodecFactory(s, serializer.EnableStrict).UniversalDecoder()
})

// isNotExistErr reports whether err means the path cannot exist. It covers both
// a plain "no such file or directory" and ENOTDIR, which EvalSymlinks returns
// when a non-directory appears mid-path (e.g. "/etc/cp.yaml/extra" where
// cp.yaml is a regular file).
func isNotExistErr(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, syscall.ENOTDIR)
}

const (
	// These are configuration key names, not credentials.
	kubeletImageCredentialProviderConfigPathKey = "imageCredentialProviderConfigPath" //nolint:gosec // G101: not a credential
	kubeletImageCredentialProviderBinDirKey     = "imageCredentialProviderBinDir"     //nolint:gosec // G101: not a credential
)

// kubeletReservedKeys lists the keys under the kubelet: section that MicroShift
// consumes itself (as kubelet startup flags) instead of passing through into the
// generated KubeletConfiguration.
var kubeletReservedKeys = []string{
	kubeletImageCredentialProviderConfigPathKey,
	kubeletImageCredentialProviderBinDirKey,
}

// ownershipFn reports the owning uid and mode of an already symlink-resolved
// path. Production passes lstatOwnership; tests inject a fake so the
// trusted-path ownership rules can be exercised without running as root.
type ownershipFn func(path string) (uid uint32, mode os.FileMode, err error)

// lstatOwnership is the production ownershipFn.
func lstatOwnership(path string) (uint32, os.FileMode, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return 0, 0, err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("unable to determine ownership of %q", path)
	}
	return st.Uid, fi.Mode(), nil
}

// trustChecker applies the trusted-path rule to path chains, memoizing the
// components it has already verified so a shared ancestor (notably every entry
// under one bin directory) is stat'd once per validation.
type trustChecker struct {
	ownership ownershipFn
	verified  map[string]struct{}
}

func newTrustChecker(ownership ownershipFn) *trustChecker {
	return &trustChecker{ownership: ownership, verified: make(map[string]struct{})}
}

// checkChain walks every component of the canonical (already symlink-resolved)
// path from / to the final object and requires each to be owned by root and not
// writable by group or others. Components verified earlier in the same
// validation are skipped.
func (tc *trustChecker) checkChain(canonical string) error {
	for _, component := range trustedPathComponents(canonical) {
		if _, ok := tc.verified[component]; ok {
			continue
		}
		uid, mode, err := tc.ownership(component)
		if err != nil {
			return err
		}
		if uid != 0 || mode&0o022 != 0 {
			return fmt.Errorf("%q must be owned by root and not writable by group or others", component)
		}
		tc.verified[component] = struct{}{}
	}
	return nil
}

// readKubeletCredentialProviderKeys copies the two credential-provider keys from
// the schemaless kubelet map into the raw (unexported) Config fields. It runs on
// every updateComputedValues(); it never touches the canonical exported fields,
// which are owned exclusively by validateKubeletCredentialProvider(), so
// re-running computed-value processing cannot revert a validated path back to the
// unresolved user value. c.Kubelet is left untouched.
func (c *Config) readKubeletCredentialProviderKeys() error {
	configPath, err := kubeletStringValue(c.Kubelet, kubeletImageCredentialProviderConfigPathKey)
	if err != nil {
		return err
	}
	binDir, err := kubeletStringValue(c.Kubelet, kubeletImageCredentialProviderBinDirKey)
	if err != nil {
		return err
	}
	c.kubeletImageCredentialProviderConfigPathRaw = configPath
	c.kubeletImageCredentialProviderBinDirRaw = binDir
	return nil
}

// KubeletImageCredentialProviderPaths returns the canonical (symlink-resolved)
// credential-provider paths computed by validateKubeletCredentialProvider(), and
// whether the feature is active. The kubelet component reads the paths through
// this accessor rather than the fields, so the canonical values it hands to
// kubelet cannot be reverted by a later updateComputedValues().
func (c *Config) KubeletImageCredentialProviderPaths() (configPath, binDir string, enabled bool) {
	if c.KubeletImageCredentialProviderConfigPath == "" {
		return "", "", false
	}
	return c.KubeletImageCredentialProviderConfigPath, c.KubeletImageCredentialProviderBinDir, true
}

// kubeletStringValue reads key from the kubelet map. A missing key, an explicit
// null, or an empty string all mean "unset"; a non-string value is an error.
func kubeletStringValue(m map[string]any, key string) (string, error) {
	if m == nil {
		return "", nil
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("kubelet.%s must be a string, got %T", key, raw)
	}
	return s, nil
}

// KubeletPassthrough returns a copy of the kubelet map with the MicroShift-owned
// keys removed, so only genuine KubeletConfiguration settings are written to the
// generated kubelet config file. A nil map returns nil.
func (c *Config) KubeletPassthrough() map[string]any {
	if c.Kubelet == nil {
		return nil
	}
	out := maps.Clone(c.Kubelet)
	for _, k := range kubeletReservedKeys {
		delete(out, k)
	}
	return out
}

// credentialPathKind selects which object types validateCredentialProviderPath
// accepts for the final object.
type credentialPathKind int

const (
	// credentialProviderConfigKind accepts a regular file or a directory.
	credentialProviderConfigKind credentialPathKind = iota
	// credentialProviderBinDirKind accepts only a directory.
	credentialProviderBinDirKind
)

// validateKubeletCredentialProvider validates the two credential-provider keys
// using the real filesystem for ownership checks.
func (c *Config) validateKubeletCredentialProvider() error {
	return c.validateKubeletCredentialProviderWith(lstatOwnership)
}

// validateKubeletCredentialProviderWith validates the two credential-provider
// keys, resolving ownership through the supplied hook. The rules are applied in
// order and the first failure wins. On success the canonical (symlink-resolved)
// paths are stored in the exported fields the kubelet component reads through
// KubeletImageCredentialProviderPaths(); the raw fields and c.Kubelet are left
// untouched. The exported fields are recomputed from the raw values on every
// call, so validation is idempotent.
func (c *Config) validateKubeletCredentialProviderWith(ownership ownershipFn) error {
	configPath := c.kubeletImageCredentialProviderConfigPathRaw
	binDir := c.kubeletImageCredentialProviderBinDirRaw

	// Clear any previously computed canonical values so a failed (or now-inactive)
	// configuration cannot leave stale paths behind.
	c.KubeletImageCredentialProviderConfigPath = ""
	c.KubeletImageCredentialProviderBinDir = ""

	// Neither set: the feature is inactive.
	if configPath == "" && binDir == "" {
		return nil
	}

	// Both keys must be provided together.
	if configPath == "" || binDir == "" {
		return fmt.Errorf("kubelet.%s and kubelet.%s must be set together",
			kubeletImageCredentialProviderConfigPathKey, kubeletImageCredentialProviderBinDirKey)
	}

	paths := []struct {
		key   string
		value string
		kind  credentialPathKind
	}{
		{kubeletImageCredentialProviderConfigPathKey, configPath, credentialProviderConfigKind},
		{kubeletImageCredentialProviderBinDirKey, binDir, credentialProviderBinDirKind},
	}

	// A single checker verifies both paths (and, for the bin dir, its entries),
	// so shared ancestors are stat'd once. Absolute check and canonicalization
	// happen in one pass; a relative second path is therefore reported only after
	// the first path has resolved.
	checker := newTrustChecker(ownership)
	canonical := make([]string, len(paths))
	for i, p := range paths {
		if !filepath.IsAbs(p.value) {
			return fmt.Errorf("kubelet.%s (%q) must be an absolute path", p.key, p.value)
		}
		resolved, err := validateCredentialProviderPath(p.value, p.kind, checker)
		if err != nil {
			return fmt.Errorf("error validating kubelet.%s (%q): %w", p.key, p.value, err)
		}
		canonical[i] = resolved
	}

	// The two keys must not resolve to the same path. Kubelet would then read the
	// bin dir as the config directory (and vice versa) and fail at registration,
	// so reject it here with a clear message instead.
	if canonical[0] == canonical[1] {
		return fmt.Errorf("kubelet.%s and kubelet.%s must not resolve to the same path (%q)",
			kubeletImageCredentialProviderConfigPathKey, kubeletImageCredentialProviderBinDirKey, canonical[0])
	}

	// Structural pre-validation of the provider configuration, on the canonical
	// paths. Upstream kubelet calls os.Exit(1) when provider registration fails,
	// which in MicroShift terminates the whole process after other components are
	// up. These checks turn the structural conditions that reach that exit into
	// ordinary fail-fast configuration errors. The configured values are used only
	// for the error prefixes; the filesystem work uses the canonical paths.
	if err := validateCredentialProviderStructure(configPath, binDir, canonical[0], canonical[1]); err != nil {
		return err
	}

	// Store canonical paths only once both keys have passed validation.
	c.KubeletImageCredentialProviderConfigPath = canonical[0]
	c.KubeletImageCredentialProviderBinDir = canonical[1]
	return nil
}

// validateCredentialProviderPath resolves path, checks that the final object is
// of an acceptable kind, and applies the trusted-path rule. It returns the
// canonical path. The object type is checked against the real filesystem (so a
// FIFO, socket or device is rejected), while ownership is checked through the
// checker's hook.
func validateCredentialProviderPath(path string, kind credentialPathKind, checker *trustChecker) (string, error) {
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		if isNotExistErr(err) {
			return "", fmt.Errorf("file or directory does not exist")
		}
		return "", err
	}

	fi, err := os.Lstat(canonical)
	if err != nil {
		return "", err
	}
	isDir := fi.IsDir()

	switch kind {
	case credentialProviderConfigKind:
		if !isDir && !fi.Mode().IsRegular() {
			return "", fmt.Errorf("%q must be a regular file or a directory", canonical)
		}
	case credentialProviderBinDirKind:
		if !isDir {
			return "", fmt.Errorf("%q must be a directory", canonical)
		}
	}

	if err := checker.checkChain(canonical); err != nil {
		return "", err
	}

	// If the final object is a directory, every entry it contains must also
	// satisfy the trusted-path rule.
	if isDir {
		if err := validateDirEntries(canonical, checker); err != nil {
			return "", err
		}
	}

	return canonical, nil
}

// validateDirEntries applies the trusted-path rule to every entry in dir.
// Symlinked entries are resolved and the full rule, including the target's
// ancestors, is applied to the target.
func validateDirEntries(dir string, checker *trustChecker) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())
		resolved, err := filepath.EvalSymlinks(entryPath)
		if err != nil {
			if isNotExistErr(err) {
				return fmt.Errorf("%q does not exist", entryPath)
			}
			return err
		}
		if err := checker.checkChain(resolved); err != nil {
			return err
		}
	}
	return nil
}

// validateCredentialProviderStructure verifies the structural conditions that
// would otherwise make kubelet call os.Exit(1) at provider registration: a
// configuration directory with no configuration files, a file that does not
// decode as a CredentialProviderConfig, a file that declares no providers, a
// provider name declared more than once, and a provider name that does not
// resolve to an executable in the bin directory. It does not replicate kubelet's
// semantic validation. configKey and binDirKey are the configured values, used
// only in messages; the checks operate on the symlink-resolved paths.
func validateCredentialProviderStructure(configKey, binDirKey, canonicalConfigPath, canonicalBinDir string) error {
	configPrefix := func(err error) error {
		return fmt.Errorf("error validating kubelet.%s (%q): %w",
			kubeletImageCredentialProviderConfigPathKey, configKey, err)
	}
	binDirPrefix := func(err error) error {
		return fmt.Errorf("error validating kubelet.%s (%q): %w",
			kubeletImageCredentialProviderBinDirKey, binDirKey, err)
	}

	files, err := collectCredentialProviderConfigFiles(canonicalConfigPath)
	if err != nil {
		return configPrefix(err)
	}

	// Record the file each provider name was first declared in, both to reject
	// duplicates across all files (kubelet rejects these in its semantic
	// validation, which exits) and to name the source file in the missing-binary
	// error. Duplicate detection is pure string comparison, so it cannot drift
	// from kubelet.
	declaredIn := make(map[string]string)
	for _, file := range files {
		names, err := decodeCredentialProviderNames(file)
		if err != nil {
			return configPrefix(err)
		}
		for _, name := range names {
			// Kubelet joins the bin dir and the provider name directly; a name
			// containing a separator would escape the bin dir, so reject it.
			if strings.Contains(name, "/") {
				return configPrefix(fmt.Errorf("provider name %q must not contain \"/\"", name))
			}
			if first, ok := declaredIn[name]; ok {
				return configPrefix(fmt.Errorf("provider %q is declared more than once (in %q and %q)", name, first, file))
			}
			declaredIn[name] = file
		}
	}

	// Check the provider binaries in a deterministic order for a stable error.
	names := make([]string, 0, len(declaredIn))
	for name := range declaredIn {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		// Report the joined path. filepath.Join(binDir, name) is exactly what
		// kubelet passes to exec.LookPath at registration; a missing or
		// non-executable binary is what MicroShift is standing in for here, so the
		// error is attributed to the bin dir, whose contents need fixing.
		joined := filepath.Join(canonicalBinDir, name)
		info, err := os.Stat(joined)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
			return binDirPrefix(fmt.Errorf("provider %q (declared in %q) has no executable at %q", name, declaredIn[name], joined))
		}
	}
	return nil
}

// collectCredentialProviderConfigFiles returns the configuration files kubelet
// would read for canonicalConfigPath. A regular file yields itself; a directory
// yields its entries whose extension is .json, .yaml or .yml, sorted
// lexicographically. A real directory entry is skipped (kubelet checks
// DirEntry.IsDir(), which is false for a symlink). Kubelet does not resolve
// symlinks and reads whatever remains with os.ReadFile, so a dangling symlink, a
// symlink to a directory, or any other non-regular entry with a matching
// extension is an error here rather than a skip: kubelet would fail on (or block
// on, for a FIFO) it and reach os.Exit. An empty directory is an error.
func collectCredentialProviderConfigFiles(canonicalConfigPath string) ([]string, error) {
	fi, err := os.Stat(canonicalConfigPath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return []string{canonicalConfigPath}, nil
	}

	entries, err := os.ReadDir(canonicalConfigPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		switch filepath.Ext(entry.Name()) {
		case ".json", ".yaml", ".yml":
		default:
			continue
		}
		// Skip only a real directory, matching kubelet's DirEntry.IsDir() check.
		// DirEntry.IsDir() is false for a symlink, so a symlink named foo.yaml
		// pointing at a directory is NOT skipped here; it is caught below as a
		// non-regular file, the way kubelet would fail os.ReadFile on it.
		if entry.IsDir() {
			continue
		}
		entryPath := filepath.Join(canonicalConfigPath, entry.Name())
		resolved, err := filepath.EvalSymlinks(entryPath)
		if err != nil {
			// Kubelet does not resolve symlinks: it includes any matching entry
			// in configFiles and later os.ReadFile fails, reaching os.Exit. A
			// dangling symlink with a matching extension is therefore an error
			// here, not a skip.
			if isNotExistErr(err) {
				return nil, fmt.Errorf("configuration file %q does not exist (dangling symlink)", entryPath)
			}
			return nil, err
		}
		info, err := os.Lstat(resolved)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			// A symlink to a directory, a FIFO named x.yaml, etc. Kubelet would
			// os.ReadFile it and block or fail at registration; reject it here.
			return nil, fmt.Errorf("configuration file %q is not a regular file", resolved)
		}
		files = append(files, resolved)
	}
	slices.Sort(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("directory contains no .json, .yaml, or .yml configuration files")
	}
	return files, nil
}

// decodeCredentialProviderNames decodes a single configuration file the same way
// kubelet does (strict, via credentialProviderCodec) and returns the declared
// provider names. Decoding with the vendored kubelet packages keeps the check
// aligned with the kubelet in the same build: unknown fields are rejected and all
// three accepted API versions convert to the internal type. It does not replicate
// kubelet's semantic validation beyond kind, group, and the presence of at least
// one provider.
func decodeCredentialProviderNames(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q: %w", file, err)
	}

	obj, gvk, err := credentialProviderCodec().Decode(data, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: %w", file, err)
	}
	if gvk.Kind != "CredentialProviderConfig" {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: unexpected kind %q", file, gvk.Kind)
	}
	if gvk.Group != kubeletconfig.GroupName {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: unexpected group %q", file, gvk.Group)
	}
	cfg, ok := obj.(*kubeletconfig.CredentialProviderConfig)
	if !ok {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: unexpected type %T", file, obj)
	}
	if len(cfg.Providers) == 0 {
		return nil, fmt.Errorf("file %q declares no providers", file)
	}

	names := make([]string, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		names = append(names, p.Name)
	}
	return names, nil
}

// trustedPathComponents returns every path component of abs, ordered from the
// root "/" down to abs itself.
func trustedPathComponents(abs string) []string {
	abs = filepath.Clean(abs)
	var components []string
	for {
		components = append(components, abs)
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	slices.Reverse(components)
	return components
}
