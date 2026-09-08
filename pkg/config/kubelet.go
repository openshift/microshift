package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	kubeletconfigv1 "k8s.io/kubelet/config/v1"
	kubeletconfigv1alpha1 "k8s.io/kubelet/config/v1alpha1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/yaml"
)

// acceptedCredentialProviderConfigAPIVersions mirrors the CredentialProviderConfig
// apiVersions the vendored kubelet registers and accepts when it decodes the
// provider configuration (see decode() in
// vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/config.go, which accepts
// any version registered in its scheme for group kubelet.config.k8s.io). Deriving
// the set from the vendored SchemeGroupVersions keeps this structural check from
// drifting from the kubelet built into the same binary.
var acceptedCredentialProviderConfigAPIVersions = map[string]struct{}{
	kubeletconfigv1.SchemeGroupVersion.String():       {},
	kubeletconfigv1beta1.SchemeGroupVersion.String():  {},
	kubeletconfigv1alpha1.SchemeGroupVersion.String(): {},
}

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

// statForTrust returns the owning uid and mode of an already symlink-resolved
// path. It is a package-level variable so tests can exercise the trusted-path
// ownership rules without running as root.
var statForTrust = func(path string) (uid uint32, mode os.FileMode, err error) {
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

// readKubeletCredentialProviderKeys copies the two credential-provider keys from
// the schemaless kubelet map into the typed Config fields. c.Kubelet is left
// untouched; canonicalization happens later, during validation.
func (c *Config) readKubeletCredentialProviderKeys() error {
	configPath, err := kubeletStringValue(c.Kubelet, kubeletImageCredentialProviderConfigPathKey)
	if err != nil {
		return err
	}
	binDir, err := kubeletStringValue(c.Kubelet, kubeletImageCredentialProviderBinDirKey)
	if err != nil {
		return err
	}
	c.KubeletImageCredentialProviderConfigPath = configPath
	c.KubeletImageCredentialProviderBinDir = binDir
	return nil
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
	out := make(map[string]any, len(c.Kubelet))
	for k, v := range c.Kubelet {
		if slices.Contains(kubeletReservedKeys, k) {
			continue
		}
		out[k] = v
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

// validateKubeletCredentialProvider validates the two credential-provider keys.
// The rules are applied in order and the first failure wins. On success the
// typed fields are replaced with the canonical (symlink-resolved) paths that are
// handed to the kubelet; c.Kubelet is left untouched.
func (c *Config) validateKubeletCredentialProvider() error {
	configPath := c.KubeletImageCredentialProviderConfigPath
	binDir := c.KubeletImageCredentialProviderBinDir

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
		dst   *string
	}{
		{kubeletImageCredentialProviderConfigPathKey, configPath, credentialProviderConfigKind, &c.KubeletImageCredentialProviderConfigPath},
		{kubeletImageCredentialProviderBinDirKey, binDir, credentialProviderBinDirKind, &c.KubeletImageCredentialProviderBinDir},
	}

	// Check that both keys are absolute before touching the filesystem.
	for _, p := range paths {
		if !filepath.IsAbs(p.value) {
			return fmt.Errorf("kubelet.%s (%q) must be an absolute path", p.key, p.value)
		}
	}

	canonical := make([]string, len(paths))
	for i, p := range paths {
		resolved, err := validateCredentialProviderPath(p.value, p.kind)
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
	// up. These checks turn the three structural conditions that reach that exit
	// (empty config directory, undecodable config, unresolvable provider name)
	// into ordinary fail-fast configuration errors. configPath (the configured
	// value) is used only for the error prefix; the filesystem work uses the
	// canonical paths.
	if err := validateCredentialProviderStructure(configPath, canonical[0], canonical[1]); err != nil {
		return err
	}

	// Store canonical paths only once both keys have passed validation.
	for i, p := range paths {
		*p.dst = canonical[i]
	}
	return nil
}

// validateCredentialProviderPath resolves path, checks that the final object is
// of an acceptable kind, and applies the trusted-path rule. It returns the
// canonical path. The object type is checked against the real filesystem (so a
// FIFO, socket or device is rejected), while ownership is checked through the
// statForTrust hook.
func validateCredentialProviderPath(path string, kind credentialPathKind) (string, error) {
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

	if err := validateTrustedChain(canonical); err != nil {
		return "", err
	}

	// If the final object is a directory, every entry it contains must also
	// satisfy the trusted-path rule.
	if isDir {
		if err := validateDirEntries(canonical); err != nil {
			return "", err
		}
	}

	return canonical, nil
}

// validateDirEntries applies the trusted-path rule to every entry in dir.
// Symlinked entries are resolved and the full rule, including the target's
// ancestors, is applied to the target.
func validateDirEntries(dir string) error {
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
		if err := validateTrustedChain(resolved); err != nil {
			return err
		}
	}
	return nil
}

// validateTrustedChain walks every component of the canonical (already
// symlink-resolved) path from / to the final object and requires each to be
// owned by root and not writable by group or others.
func validateTrustedChain(canonical string) error {
	for _, component := range trustedPathComponents(canonical) {
		uid, mode, err := statForTrust(component)
		if err != nil {
			return err
		}
		if uid != 0 || mode&0o022 != 0 {
			return fmt.Errorf("%q must be owned by root and not writable by group or others", component)
		}
	}
	return nil
}

// validateCredentialProviderStructure verifies the structural conditions that
// would otherwise make kubelet call os.Exit(1) at provider registration: a
// configuration directory with no configuration files, a file that does not
// decode as a CredentialProviderConfig, a file that declares no providers, and a
// provider name that does not resolve to an executable in the bin directory. It
// does not replicate kubelet's semantic validation. configKey is the configured
// value, used only in messages; canonicalConfigPath and canonicalBinDir are the
// symlink-resolved paths the checks operate on.
func validateCredentialProviderStructure(configKey, canonicalConfigPath, canonicalBinDir string) error {
	prefix := func(err error) error {
		return fmt.Errorf("error validating kubelet.%s (%q): %w",
			kubeletImageCredentialProviderConfigPathKey, configKey, err)
	}

	files, err := collectCredentialProviderConfigFiles(canonicalConfigPath)
	if err != nil {
		return prefix(err)
	}

	for _, file := range files {
		names, err := decodeCredentialProviderNames(file)
		if err != nil {
			return prefix(err)
		}
		for _, name := range names {
			// Kubelet joins the bin dir and the provider name directly; a name
			// containing a separator would escape the bin dir, so reject it.
			if strings.Contains(name, "/") {
				return prefix(fmt.Errorf("provider name %q must not contain \"/\"", name))
			}
			// Report the joined path, never exec.LookPath's return value: on
			// error LookPath returns an empty string, which is the upstream
			// defect that prints "plugin binary executable  did not exist".
			joined := filepath.Join(canonicalBinDir, name)
			if _, err := exec.LookPath(joined); err != nil {
				return prefix(fmt.Errorf("provider %q has no executable at %q", name, joined))
			}
		}
	}
	return nil
}

// collectCredentialProviderConfigFiles returns the configuration files kubelet
// would read for canonicalConfigPath. A regular file yields itself; a directory
// yields its regular entries (symlinks resolved) whose extension is .json,
// .yaml or .yml, sorted lexicographically. An empty directory is an error.
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
		resolved, err := filepath.EvalSymlinks(filepath.Join(canonicalConfigPath, entry.Name()))
		if err != nil {
			if isNotExistErr(err) {
				// A dangling symlink is not a configuration file kubelet reads.
				continue
			}
			return nil, err
		}
		info, err := os.Lstat(resolved)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		files = append(files, resolved)
	}
	slices.Sort(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("directory contains no .json, .yaml, or .yml configuration files")
	}
	return files, nil
}

// decodeCredentialProviderNames decodes a single configuration file as a
// CredentialProviderConfig using the vendored k8s.io/kubelet/config/v1 types and
// returns the declared provider names. Decoding with the vendored types keeps the
// check aligned with the kubelet in the same build. It does not validate provider
// contents beyond apiVersion, kind, and the presence of at least one provider.
func decodeCredentialProviderNames(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %q: %w", file, err)
	}

	var cfg kubeletconfigv1.CredentialProviderConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: %w", file, err)
	}
	if cfg.Kind != "CredentialProviderConfig" {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: unexpected kind %q", file, cfg.Kind)
	}
	if _, ok := acceptedCredentialProviderConfigAPIVersions[cfg.APIVersion]; !ok {
		return nil, fmt.Errorf("file %q is not a valid CredentialProviderConfig: unsupported apiVersion %q", file, cfg.APIVersion)
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
