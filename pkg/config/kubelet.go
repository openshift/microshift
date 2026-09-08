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

	"golang.org/x/sys/unix"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	kubeletconfigv1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1"
	kubeletconfigv1alpha1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1alpha1"
	kubeletconfigv1beta1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1beta1"
)

// credentialProviderCodecs is the same strict decoder the vendored kubelet uses
// to read the credential provider configuration: see the scheme setup in
// vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/plugin.go (lines
// 69-70, 135-138) and decode() in the sibling config.go. Strict decoding
// rejects unknown fields, and all three API versions kubelet accepts
// (v1alpha1, v1beta1, v1 of kubelet.config.k8s.io) are registered together with
// the internal type and its conversions. Building the decoder from the same
// vendored packages keeps this structural check from diverging from the kubelet
// compiled into the same binary; a lenient decoder would let a typo'd field
// through to the os.Exit at registration.
var credentialProviderCodecs = func() serializer.CodecFactory {
	s := runtime.NewScheme()
	utilruntime.Must(kubeletconfig.AddToScheme(s))
	utilruntime.Must(kubeletconfigv1alpha1.AddToScheme(s))
	utilruntime.Must(kubeletconfigv1beta1.AddToScheme(s))
	utilruntime.Must(kubeletconfigv1.AddToScheme(s))
	return serializer.NewCodecFactory(s, serializer.EnableStrict)
}()

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

// aclForTrust reports whether path carries an extended POSIX access ACL. Mode
// bits do not reveal ACL write grants (for example `setfacl -m u:x:rwx dir`
// leaves the mode at 0755), so the trusted-path rule rejects any component that
// carries one. It is a package-level variable so tests can simulate ACLs without
// setfacl or root. A filesystem that stores no ACL for the object (ENODATA) or
// does not support ACLs (ENOTSUP) reports false.
var aclForTrust = func(path string) (bool, error) {
	sz, err := unix.Lgetxattr(path, "system.posix_acl_access", nil)
	if err != nil {
		if errors.Is(err, unix.ENODATA) || errors.Is(err, unix.ENOTSUP) {
			return false, nil
		}
		return false, err
	}
	return sz > 0, nil
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
		if err := checkNoExtendedACL(component); err != nil {
			return err
		}
	}
	return nil
}

// checkNoExtendedACL rejects a path that carries an extended POSIX ACL, which can
// grant write access that the mode bits do not show.
func checkNoExtendedACL(path string) error {
	hasACL, err := aclForTrust(path)
	if err != nil {
		return err
	}
	if hasACL {
		return fmt.Errorf("%q must not have an extended ACL", path)
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
// yields its entries whose extension is .json, .yaml or .yml, sorted
// lexicographically. Directory entries are skipped (kubelet checks
// !entry.IsDir()). Kubelet does not resolve symlinks and reads whatever remains
// with os.ReadFile, so a dangling symlink or a non-regular entry with a
// matching extension is an error here rather than a skip: kubelet would fail on
// (or block on, for a FIFO) it and reach os.Exit. An empty directory is an
// error.
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
		if info.IsDir() {
			// Kubelet skips directories (it checks !entry.IsDir()); a directory
			// named e.g. foo.yaml is ignored by both.
			continue
		}
		if !info.Mode().IsRegular() {
			// Kubelet would os.ReadFile a non-regular entry (e.g. a FIFO named
			// x.yaml) and block or fail at registration; reject it here.
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
// kubelet does (strict, via credentialProviderCodecs) and returns the declared
// provider names. Decoding with the vendored kubelet packages keeps the check
// aligned with the kubelet in the same build: unknown fields are rejected and all
// three accepted API versions convert to the internal type. It does not replicate
// kubelet's semantic validation beyond kind, group, and the presence of at least
// one provider.
func decodeCredentialProviderNames(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		// A non-root reader (typically `microshift show-config` against a 0600
		// file) cannot read the file; say so rather than reporting it invalid.
		if errors.Is(err, syscall.EACCES) {
			return nil, fmt.Errorf("cannot read %q: permission denied (run as root)", file)
		}
		return nil, fmt.Errorf("unable to read file %q: %w", file, err)
	}

	obj, gvk, err := credentialProviderCodecs.UniversalDecoder().Decode(data, nil, nil)
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
