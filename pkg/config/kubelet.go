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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	credentialproviderv1 "k8s.io/kubelet/pkg/apis/credentialprovider/v1"
	credentialproviderv1alpha1 "k8s.io/kubelet/pkg/apis/credentialprovider/v1alpha1"
	credentialproviderv1beta1 "k8s.io/kubelet/pkg/apis/credentialprovider/v1beta1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	kubeletconfigv1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1"
	kubeletconfigv1alpha1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1alpha1"
	kubeletconfigv1beta1 "k8s.io/kubernetes/pkg/kubelet/apis/config/v1beta1"
)

// credentialProviderAPIVersions is the set of provider exec API versions kubelet
// accepts, mirroring the apiVersions map built in
// vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/plugin.go from the same
// three packages. Re-check against upstream on every kubernetes rebase.
var credentialProviderAPIVersions = sets.New[string](
	credentialproviderv1alpha1.SchemeGroupVersion.String(),
	credentialproviderv1beta1.SchemeGroupVersion.String(),
	credentialproviderv1.SchemeGroupVersion.String(),
)

// validCredentialProviderCacheTypes mirrors validCacheTypes in
// vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/config.go.
var validCredentialProviderCacheTypes = sets.New[string](
	string(kubeletconfig.ServiceAccountServiceAccountTokenCacheType),
	string(kubeletconfig.TokenServiceAccountTokenCacheType),
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
	if err := validateCredentialProviderStructure(configPath, binDir, canonical[0], canonical[1], checker); err != nil {
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

	// If the final object is a directory, apply the trusted-path rule only to the
	// entries kubelet will actually consume.
	if isDir {
		switch kind {
		case credentialProviderConfigKind:
			// kubelet's readCredentialProviderConfig reads only the .json/.yaml/.yml
			// files in the directory; every other entry (a README, a .bak, an editor
			// swap file, a subdirectory) is ignored. Check exactly those files, so an
			// unrelated non-root entry does not block startup when kubelet would never
			// read it.
			if err := validateConfigDirEntries(canonical, checker); err != nil {
				return "", err
			}
		case credentialProviderBinDirKind:
			// kubelet only executes filepath.Join(binDir, provider.Name) for the
			// declared providers, so the bin dir's entries are not walked here. The
			// trusted-path rule is applied to each declared provider binary in
			// validateCredentialProviderStructure; unrelated files in the bin dir are
			// ignored, matching kubelet. The bin dir itself is still verified above,
			// which is what prevents an unprivileged user from dropping a file named
			// after a provider.
		}
	}

	return canonical, nil
}

// isCredentialProviderConfigEntry reports whether entry is one kubelet's
// readCredentialProviderConfig would read from a configuration directory: a
// non-directory whose extension is .json, .yaml or .yml. DirEntry.IsDir() is
// false for a symlink, matching kubelet, so a symlink named x.yaml is a config
// entry (and is caught downstream if it does not resolve to a regular file).
func isCredentialProviderConfigEntry(entry os.DirEntry) bool {
	if entry.IsDir() {
		return false
	}
	switch filepath.Ext(entry.Name()) {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

// validateConfigDirEntries applies the trusted-path rule to the configuration
// files in dir that kubelet would read, ignoring every other entry. Symlinked
// entries are resolved and the full rule, including the target's ancestors, is
// applied to the target.
func validateConfigDirEntries(dir string, checker *trustChecker) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !isCredentialProviderConfigEntry(entry) {
			continue
		}
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

// validateCredentialProviderStructure verifies the structural and semantic
// conditions that would otherwise make kubelet call os.Exit(1) at provider
// registration: a configuration directory with no configuration files, a file
// that does not decode as a CredentialProviderConfig, a file that declares no
// providers, a provider that fails kubelet's semantic validation (see
// validateCredentialProviderSemantics), a provider name declared more than once
// across files, and a provider name that does not resolve to an executable in the
// bin directory. configKey and binDirKey are the configured values, used only in
// messages; the checks operate on the symlink-resolved paths. checker applies the
// trusted-path rule to each declared provider binary (the only bin-dir entries
// kubelet consumes).
func validateCredentialProviderStructure(configKey, binDirKey, canonicalConfigPath, canonicalBinDir string, checker *trustChecker) error {
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
	// duplicates across all files and to name the source file in the missing-binary
	// error. Cross-file duplicate detection is pure string comparison at the merged
	// level, so it cannot drift from kubelet; within-file duplicates are caught by
	// the semantic validation below.
	declaredIn := make(map[string]string)
	for _, file := range files {
		cfg, err := decodeCredentialProviderConfig(file)
		if err != nil {
			return configPrefix(err)
		}
		// Semantic validation mirrors kubelet's own, run per file so its field-path
		// messages read like kubelet's. Kubelet validates the merged provider list;
		// validating per file gives the same coverage with a message that can name
		// the offending file when the config path is a directory.
		if err := validateCredentialProviderSemantics(cfg.Providers); err != nil {
			if len(files) > 1 {
				return configPrefix(fmt.Errorf("file %q: %w", file, err))
			}
			return configPrefix(err)
		}
		for _, p := range cfg.Providers {
			if first, ok := declaredIn[p.Name]; ok {
				return configPrefix(fmt.Errorf("provider %q is declared more than once (in %q and %q)", p.Name, first, file))
			}
			declaredIn[p.Name] = file
		}
	}

	// Check the provider binaries in a deterministic order for a stable error.
	names := make([]string, 0, len(declaredIn))
	for name := range declaredIn {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		// filepath.Join(binDir, name) is exactly what kubelet passes to
		// exec.LookPath at registration; a missing or non-executable binary is what
		// MicroShift is standing in for here, so the error is attributed to the bin
		// dir, whose contents need fixing.
		joined := filepath.Join(canonicalBinDir, name)
		resolved, err := filepath.EvalSymlinks(joined)
		if err != nil {
			return binDirPrefix(fmt.Errorf("provider %q (declared in %q) has no executable at %q", name, declaredIn[name], joined))
		}
		info, err := os.Lstat(resolved)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
			return binDirPrefix(fmt.Errorf("provider %q (declared in %q) has no executable at %q", name, declaredIn[name], joined))
		}
		// The provider binary runs with kubelet's privileges, so the resolved binary
		// (and, for a symlinked binary, its target's ancestors) must satisfy the
		// trusted-path rule. Ancestors already verified when the bin dir was checked
		// are memoized and skipped.
		if err := checker.checkChain(resolved); err != nil {
			return binDirPrefix(err)
		}
	}
	return nil
}

// validateCredentialProviderSemantics mirrors validateCredentialProviderConfig in
// vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/config.go, which is
// unexported. Re-check against upstream on every kubernetes rebase. The structure
// and rule order are kept identical to upstream so a diff is trivial; the only
// deliberate deviations are:
//   - field paths use providers[i] (Index) so a message names the offending
//     provider; upstream uses the bare "providers" path.
//   - the KubeletServiceAccountTokenForCredentialProviders feature-gate check is
//     not mirrored (see the tokenAttributes block below).
func validateCredentialProviderSemantics(providers []kubeletconfig.CredentialProvider) error {
	allErrs := field.ErrorList{}

	if len(providers) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("providers"), "at least 1 item in plugins is required"))
	}

	seenProviderNames := sets.New[string]()
	for i := range providers {
		provider := providers[i]
		fieldPath := field.NewPath("providers").Index(i)

		// Upstream has no explicit empty-name rule, but filepath.Join(binDir, "")
		// is the bin dir itself and registration then fails; report it clearly.
		if len(provider.Name) == 0 {
			allErrs = append(allErrs, field.Required(fieldPath.Child("name"), "provider name is required"))
		}

		if strings.Contains(provider.Name, "/") {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("name"), provider.Name, "provider name cannot contain '/'"))
		}

		if strings.Contains(provider.Name, " ") {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("name"), provider.Name, "provider name cannot contain spaces"))
		}

		if provider.Name == "." {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("name"), provider.Name, "provider name cannot be '.'"))
		}

		if provider.Name == ".." {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("name"), provider.Name, "provider name cannot be '..'"))
		}

		if seenProviderNames.Has(provider.Name) {
			allErrs = append(allErrs, field.Duplicate(fieldPath.Child("name"), provider.Name))
		}
		seenProviderNames.Insert(provider.Name)

		if provider.APIVersion == "" {
			allErrs = append(allErrs, field.Required(fieldPath.Child("apiVersion"), ""))
		} else if !credentialProviderAPIVersions.Has(provider.APIVersion) {
			allErrs = append(allErrs, field.NotSupported(fieldPath.Child("apiVersion"), provider.APIVersion, sets.List(credentialProviderAPIVersions)))
		}

		if len(provider.MatchImages) == 0 {
			allErrs = append(allErrs, field.Required(fieldPath.Child("matchImages"), "at least 1 item in matchImages is required"))
		}

		for _, matchImage := range provider.MatchImages {
			if _, err := credentialprovider.ParseSchemelessURL(matchImage); err != nil {
				allErrs = append(allErrs, field.Invalid(fieldPath.Child("matchImages"), matchImage, fmt.Sprintf("match image is invalid: %s", err.Error())))
			}
		}

		if provider.DefaultCacheDuration == nil {
			allErrs = append(allErrs, field.Required(fieldPath.Child("defaultCacheDuration"), ""))
		}

		if provider.DefaultCacheDuration != nil && provider.DefaultCacheDuration.Duration < 0 {
			allErrs = append(allErrs, field.Invalid(fieldPath.Child("defaultCacheDuration"), provider.DefaultCacheDuration, "must be greater than or equal to 0"))
		}

		allErrs = append(allErrs, validateCredentialProviderTokenAttributes(fieldPath, provider)...)
	}

	return allErrs.ToAggregate()
}

// validateCredentialProviderTokenAttributes mirrors the tokenAttributes block of
// validateCredentialProviderConfig. It is factored out of
// validateCredentialProviderSemantics only to keep that function within the
// project's cyclomatic-complexity limit; the rules and their order are unchanged
// from upstream. A nil provider.TokenAttributes yields no errors.
func validateCredentialProviderTokenAttributes(fieldPath *field.Path, provider kubeletconfig.CredentialProvider) field.ErrorList {
	if provider.TokenAttributes == nil {
		return nil
	}

	allErrs := field.ErrorList{}
	fldPath := fieldPath.Child("tokenAttributes")
	// Upstream also forbids tokenAttributes when the
	// KubeletServiceAccountTokenForCredentialProviders feature gate is
	// disabled. That gate is Beta/on by default in the vendored kubelet, and
	// its effective value at registration comes from the kubelet passthrough
	// (kubelet.featureGates), which this validation cannot see. The gate check
	// is therefore deliberately not mirrored; it is the one documented residual.
	if len(provider.TokenAttributes.ServiceAccountTokenAudience) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("serviceAccountTokenAudience"), ""))
	}
	if provider.TokenAttributes.RequireServiceAccount == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("requireServiceAccount"), ""))
	}
	if provider.APIVersion != credentialproviderv1.SchemeGroupVersion.String() {
		allErrs = append(allErrs, field.Forbidden(fldPath, fmt.Sprintf("tokenAttributes is only supported for %s API version", credentialproviderv1.SchemeGroupVersion.String())))
	}

	if provider.TokenAttributes.RequireServiceAccount != nil && !*provider.TokenAttributes.RequireServiceAccount && len(provider.TokenAttributes.RequiredServiceAccountAnnotationKeys) > 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("requiredServiceAccountAnnotationKeys"), "requireServiceAccount cannot be false when requiredServiceAccountAnnotationKeys is set"))
	}

	allErrs = append(allErrs, validateCredentialProviderAnnotationKeys(fldPath.Child("requiredServiceAccountAnnotationKeys"), provider.TokenAttributes.RequiredServiceAccountAnnotationKeys)...)
	allErrs = append(allErrs, validateCredentialProviderAnnotationKeys(fldPath.Child("optionalServiceAccountAnnotationKeys"), provider.TokenAttributes.OptionalServiceAccountAnnotationKeys)...)

	requiredServiceAccountAnnotationKeys := sets.New[string](provider.TokenAttributes.RequiredServiceAccountAnnotationKeys...)
	optionalServiceAccountAnnotationKeys := sets.New[string](provider.TokenAttributes.OptionalServiceAccountAnnotationKeys...)
	duplicateAnnotationKeys := requiredServiceAccountAnnotationKeys.Intersection(optionalServiceAccountAnnotationKeys)
	if duplicateAnnotationKeys.Len() > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, sets.List(duplicateAnnotationKeys), "annotation keys cannot be both required and optional"))
	}

	switch {
	case len(provider.TokenAttributes.CacheType) == 0:
		allErrs = append(allErrs, field.Required(fldPath.Child("cacheType"), fmt.Sprintf("cacheType is required to be set when tokenAttributes is specified. Supported values are: %s", strings.Join(sets.List(validCredentialProviderCacheTypes), ", "))))
	case validCredentialProviderCacheTypes.Has(string(provider.TokenAttributes.CacheType)):
		// ok
	default:
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("cacheType"), provider.TokenAttributes.CacheType, sets.List(validCredentialProviderCacheTypes)))
	}

	return allErrs
}

// validateCredentialProviderAnnotationKeys mirrors validateServiceAccountAnnotationKeys
// in vendor/k8s.io/kubernetes/pkg/credentialprovider/plugin/config.go.
func validateCredentialProviderAnnotationKeys(fldPath *field.Path, keys []string) field.ErrorList {
	allErrs := field.ErrorList{}

	seenAnnotationKeys := sets.New[string]()
	for _, k := range keys {
		// The rule is QualifiedName except that case doesn't matter, so convert to
		// lowercase before checking.
		for _, msg := range validation.IsQualifiedName(strings.ToLower(k)) {
			allErrs = append(allErrs, field.Invalid(fldPath, k, msg))
		}
		if seenAnnotationKeys.Has(k) {
			allErrs = append(allErrs, field.Duplicate(fldPath, k))
		}
		seenAnnotationKeys.Insert(k)
	}
	return allErrs
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
		// isCredentialProviderConfigEntry applies the same .json/.yaml/.yml +
		// DirEntry.IsDir() filter kubelet uses, so this loop and the trusted-path
		// walk in validateConfigDirEntries cannot disagree about which entries
		// kubelet consumes. A real directory named foo.yaml is skipped; a symlink
		// named foo.yaml is not (IsDir() is false for a symlink) and is caught below
		// as a non-regular file, the way kubelet would fail os.ReadFile on it.
		if !isCredentialProviderConfigEntry(entry) {
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

// decodeCredentialProviderConfig decodes a single configuration file the same way
// kubelet does (strict, via credentialProviderCodec) and returns the internal
// CredentialProviderConfig. Decoding with the vendored kubelet packages keeps the
// check aligned with the kubelet in the same build: unknown fields are rejected
// and all three accepted API versions convert to the internal type. It checks
// kind, group, type, and the presence of at least one provider; the field-level
// semantic rules are applied by validateCredentialProviderSemantics.
func decodeCredentialProviderConfig(file string) (*kubeletconfig.CredentialProviderConfig, error) {
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
	return cfg, nil
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
