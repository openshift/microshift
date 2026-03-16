package prerun

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openshift/microshift/pkg/config"
	"sigs.k8s.io/yaml"
)

func TestFeatureGateLockFile_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		lockFile featureGateLockFile
		wantErr  bool
	}{
		{
			name: "custom feature gates with enabled and disabled",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{
					Enabled:  []string{"FeatureA", "FeatureB"},
					Disabled: []string{"FeatureC"},
				},
			},
			wantErr: false,
		},
		{
			name: "TechPreviewNoUpgrade",
			lockFile: featureGateLockFile{
				FeatureSet:      config.FeatureSetTechPreviewNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{},
			},
			wantErr: false,
		},
		{
			name: "DevPreviewNoUpgrade",
			lockFile: featureGateLockFile{
				FeatureSet:      config.FeatureSetDevPreviewNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{},
			},
			wantErr: false,
		},
		{
			name: "empty feature gates",
			lockFile: featureGateLockFile{
				FeatureSet:      "",
				CustomNoUpgrade: config.EnableDisableFeatures{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.lockFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var unmarshaled featureGateLockFile
				if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
					t.Errorf("Unmarshal() error = %v", err)
					return
				}
				if !reflect.DeepEqual(tt.lockFile, unmarshaled) {
					t.Errorf("Marshal/Unmarshal roundtrip failed: got %#v, want %#v", unmarshaled, tt.lockFile)
				}
			}
		})
	}
}

func TestIsCustomFeatureGatesConfigured(t *testing.T) {
	tests := []struct {
		name string
		fg   config.FeatureGates
		want bool
	}{
		{
			name: "CustomNoUpgrade with enabled features",
			fg: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{
					Enabled: []string{"FeatureA"},
				},
			},
			want: true,
		},
		{
			name: "TechPreviewNoUpgrade",
			fg: config.FeatureGates{
				FeatureSet: config.FeatureSetTechPreviewNoUpgrade,
			},
			want: true,
		},
		{
			name: "DevPreviewNoUpgrade",
			fg: config.FeatureGates{
				FeatureSet: config.FeatureSetDevPreviewNoUpgrade,
			},
			want: true,
		},
		{
			name: "empty feature gates",
			fg: config.FeatureGates{
				FeatureSet: "",
			},
			want: true, // validation passes when prev and current both have no feature set
		},
		{
			name: "CustomNoUpgrade without any enabled/disabled",
			fg: config.FeatureGates{
				FeatureSet:      config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{},
			},
			want: true, // validation passes when prev and current match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configValidationChecksPass(featureGateLockFile{
				FeatureSet:      tt.fg.FeatureSet,
				CustomNoUpgrade: tt.fg.CustomNoUpgrade,
			}, &tt.fg)
			got := err == nil
			if got != tt.want {
				t.Errorf("configValidationChecksPass() got pass = %v, want %v (err = %v)", got, tt.want, err)
			}
		})
	}
}

func TestFeatureGateLockFile_ReadWrite(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "featuregate-lockFile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockFilePath := filepath.Join(tmpDir, "no-upgrade")

	tests := []struct {
		name     string
		lockFile featureGateLockFile
	}{
		{
			name: "write and read custom feature gates",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{
					Enabled:  []string{"FeatureA", "FeatureB"},
					Disabled: []string{"FeatureC"},
				},
			},
		},
		{
			name: "write and read TechPreviewNoUpgrade",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetTechPreviewNoUpgrade,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write lockFile file
			if err := writeFeatureGateLockFile(lockFilePath, tt.lockFile); err != nil {
				t.Errorf("writeFeatureGateLockFile() error = %v", err)
				return
			}

			// Read lockFile file
			got, err := readFeatureGateLockFile(lockFilePath)
			if err != nil {
				t.Errorf("readFeatureGateLockFile() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.lockFile) {
				t.Errorf("readFeatureGateLockFile() = %#v, want %#v", got, tt.lockFile)
			}

			// Clean up for next test
			os.Remove(lockFilePath)
		})
	}
}

func TestFeatureGateLockFile_ReadNonExistent(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "featuregate-lockFile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockFilePath := filepath.Join(tmpDir, "no-upgrade")

	_, err = readFeatureGateLockFile(lockFilePath)
	if !errors.Is(err, errLockFileDoesNotExist) {
		t.Errorf("readFeatureGateLockFile() error = %v, want %v", err, errLockFileDoesNotExist)
	}
}

func TestConfigValidationChecksPass(t *testing.T) {
	tests := []struct {
		name     string
		lockFile featureGateLockFile
		current  config.FeatureGates
		wantErr  bool
	}{
		{
			name: "unset any feature set",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
			},
			current: config.FeatureGates{
				FeatureSet: "",
			},
			wantErr: true,
		},
		{
			name: "change CustomNoUpgrade to any other feature set",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
			},
			current: config.FeatureGates{
				FeatureSet: config.FeatureSetTechPreviewNoUpgrade,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configValidationChecksPass(tt.lockFile, &tt.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("configValidationChecksPass() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatureGateLockManagement_FirstRun(t *testing.T) {
	// Use a fixed test version (doesn't depend on ldflags)
	testVersion := versionMetadata{Major: 4, Minor: 21, Patch: 0}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "featuregate-lockFile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the lockFile file path for testing
	originalPath := featureGateLockFilePath
	featureGateLockFilePath = filepath.Join(tmpDir, "no-upgrade")
	defer func() { featureGateLockFilePath = originalPath }()

	// Override getExecutableVersion for testing
	originalGetExecutableVersion := getExecutableVersion
	getExecutableVersion = func() (versionMetadata, error) {
		return testVersion, nil
	}
	defer func() { getExecutableVersion = originalGetExecutableVersion }()

	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{
					Enabled: []string{"FeatureA"},
				},
			},
		},
	}

	// First run - should create lockFile file
	if err := FeatureGateLockManagement(cfg); err != nil {
		t.Errorf("FeatureGateLockManagement() first run error = %v", err)
	}

	// Verify lockFile file was created
	if _, err := os.Stat(featureGateLockFilePath); os.IsNotExist(err) {
		t.Error("Lock file was not created")
	}

	// Second run with same config - should succeed
	if err := FeatureGateLockManagement(cfg); err != nil {
		t.Errorf("FeatureGateLockManagement() second run error = %v", err)
	}
}

func TestFeatureGateLockManagement_ConfigChange(t *testing.T) {
	// Use a fixed test version (doesn't depend on ldflags)
	testVersion := versionMetadata{Major: 4, Minor: 21, Patch: 0}

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "featuregate-lockFile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the lockFile file path for testing
	originalPath := featureGateLockFilePath
	featureGateLockFilePath = filepath.Join(tmpDir, "no-upgrade")
	defer func() { featureGateLockFilePath = originalPath }()

	// Override getExecutableVersion for testing
	originalGetExecutableVersion := getExecutableVersion
	getExecutableVersion = func() (versionMetadata, error) {
		return testVersion, nil
	}
	defer func() { getExecutableVersion = originalGetExecutableVersion }()

	// Create lockFile file with initial config (CustomNoUpgrade feature set)
	initialLock := featureGateLockFile{
		FeatureSet: config.FeatureSetCustomNoUpgrade,
		CustomNoUpgrade: config.EnableDisableFeatures{
			Enabled: []string{"FeatureA"},
		},
		Version: testVersion,
	}
	if err := writeFeatureGateLockFile(featureGateLockFilePath, initialLock); err != nil {
		t.Fatal(err)
	}

	// Try to run with no feature gates configured - should fail
	// (configValidationChecksPass blocks unsetting a feature set)
	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: "", // Trying to unset feature gates
			},
		},
	}

	err = FeatureGateLockManagement(cfg)
	if err == nil {
		t.Error("FeatureGateLockManagement() should have failed when trying to unset feature gates")
	}
}

func TestFeatureGateLockManagement_VersionChange(t *testing.T) {
	// Use a fixed base version for testing (doesn't depend on ldflags)
	baseVersion := versionMetadata{Major: 4, Minor: 21, Patch: 0}

	// getVersion creates a version with offsets from the base version
	getVersion := func(majorOffset, minorOffset, patchOffset int) versionMetadata {
		return versionMetadata{
			Major: baseVersion.Major + majorOffset,
			Minor: baseVersion.Minor + minorOffset,
			Patch: baseVersion.Patch + patchOffset,
		}
	}

	tests := []struct {
		name                            string
		lockFileVer                     versionMetadata
		currentVer                      versionMetadata
		customNoUpgrade                 *config.EnableDisableFeatures
		specialHandlingSupportException *config.EnableDisableFeatures
		wantErr                         bool
		description                     string
	}{
		{
			name:                            "minor version upgrade should fail",
			lockFileVer:                     getVersion(0, 0, 0),
			currentVer:                      getVersion(0, 1, 0),
			wantErr:                         true,
			specialHandlingSupportException: &config.EnableDisableFeatures{},
			description:                     "Minor version upgrade (4.21.0 -> 4.22.0) should be blocked",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
		},
		{
			name:        "major version upgrade should fail",
			lockFileVer: getVersion(0, 0, 0),
			currentVer:  getVersion(1, 0, 0),
			wantErr:     true,
			description: "Major version upgrade (4.21.0 -> 5.0.0) should be blocked",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{},
		},
		{
			name:        "patch version change should succeed",
			lockFileVer: getVersion(0, 0, 0),
			currentVer:  getVersion(0, 0, 1),
			wantErr:     false,
			description: "Patch version change (4.21.0 -> 4.21.1) should be allowed",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{},
		},
		{
			name:        "same version should succeed",
			lockFileVer: getVersion(0, 0, 0),
			currentVer:  getVersion(0, 0, 0),
			wantErr:     false,
			description: "Same version (4.21.0 -> 4.21.0) should be allowed",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{},
		},
		{
			name:        "minor version downgrade should fail",
			lockFileVer: getVersion(0, 1, 0),
			currentVer:  getVersion(0, 0, 0),
			wantErr:     true,
			description: "Minor version downgrade (4.22.0 -> 4.21.0) should be blocked",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{},
		},
		{
			name:        "major version downgrade should fail",
			lockFileVer: getVersion(1, -21, 0),
			currentVer:  getVersion(0, 0, 0),
			wantErr:     true,
			description: "Major version downgrade (5.0.0 -> 4.21.0) should be blocked",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{},
		},
		{
			name:        "major version upgrade with special handling support exception should succeed",
			lockFileVer: getVersion(0, 0, 0),
			currentVer:  getVersion(1, -21, 0),
			wantErr:     false,
			description: "major version upgrade (4.21.0 -> 5.0.0) with special handling support exception should succeed",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
		},
		{
			name:        "minor version upgrade with special handling support exception should succeed",
			lockFileVer: getVersion(0, 0, 0),
			currentVer:  getVersion(0, 1, 0),
			wantErr:     false,
			description: "minor version upgrade (4.21.0 -> 4.22.0) with special handling support exception should succeed",
			customNoUpgrade: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
			specialHandlingSupportException: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
		},
		{
			name:            "minor version upgrade with feature only in special handling should succeed",
			lockFileVer:     getVersion(0, 0, 0),
			currentVer:      getVersion(0, 1, 0),
			wantErr:         false,
			description:     "feature only in SpecialHandlingSupportExceptionRequired should not block upgrades",
			customNoUpgrade: &config.EnableDisableFeatures{},
			specialHandlingSupportException: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
		},
		{
			name:            "major version upgrade with feature only in special handling should succeed",
			lockFileVer:     getVersion(0, 0, 0),
			currentVer:      getVersion(1, -21, 0),
			wantErr:         false,
			description:     "feature only in SpecialHandlingSupportExceptionRequired should not block major upgrades",
			customNoUpgrade: &config.EnableDisableFeatures{},
			specialHandlingSupportException: &config.EnableDisableFeatures{
				Enabled: []string{"FeatureA"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := os.MkdirTemp("", "featuregate-lockFile-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Override the lockFile file path for testing
			originalPath := featureGateLockFilePath
			featureGateLockFilePath = filepath.Join(tmpDir, "no-upgrade")
			defer func() { featureGateLockFilePath = originalPath }()

			// Override version file path for testing
			originalVersionPath := versionFilePath
			versionFilePath = filepath.Join(tmpDir, "version")
			defer func() { versionFilePath = originalVersionPath }()

			// Create lockFile file with locked version. Lock file does not store the special handling support exception.
			customNoUpgrade := config.EnableDisableFeatures{}
			if tt.customNoUpgrade != nil {
				customNoUpgrade = *tt.customNoUpgrade
			}
			specialHandling := config.EnableDisableFeatures{}
			if tt.specialHandlingSupportException != nil {
				specialHandling = *tt.specialHandlingSupportException
			}
			lockFile := featureGateLockFile{
				FeatureSet:      config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: customNoUpgrade,
				Version:         tt.lockFileVer,
			}
			if err := writeFeatureGateLockFile(featureGateLockFilePath, lockFile); err != nil {
				t.Fatal(err)
			}

			cfg := &config.Config{
				ApiServer: config.ApiServer{
					FeatureGates: config.FeatureGates{
						FeatureSet:                              config.FeatureSetCustomNoUpgrade,
						CustomNoUpgrade:                         customNoUpgrade,
						SpecialHandlingSupportExceptionRequired: specialHandling,
					},
				},
			}

			err = FeatureGateLockManagement(cfg)
			if err == nil {
				t.Error("FeatureGateLockManagement() should have failed with version change")
			}
		})
	}
}

func TestFeatureGateLockManagement_NoCustomFeatureGates(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "featuregate-lockFile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the lockFile file path for testing
	originalPath := featureGateLockFilePath
	featureGateLockFilePath = filepath.Join(tmpDir, "no-upgrade")
	defer func() { featureGateLockFilePath = originalPath }()

	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: "", // No custom feature gates
			},
		},
	}

	// Should succeed and not create lockFile file
	if err := FeatureGateLockManagement(cfg); err != nil {
		t.Errorf("FeatureGateLockManagement() with no custom feature gates error = %v", err)
	}

	// Verify lockFile file was not created
	if _, err := os.Stat(featureGateLockFilePath); !os.IsNotExist(err) {
		t.Error("Lock file should not have been created without custom feature gates")
	}
}

// TestUpgradeThenRollbackWithFeatureGateLock simulates the scenario where:
// 1. MicroShift 4.21.0 runs with CustomNoUpgrade + FeatureA
// 2. RPM is upgraded to 4.22.0 and MicroShift is restarted
// 3. FeatureGateLockManagement blocks startup due to version mismatch
// 4. User rolls back to 4.21.0
//
// The test verifies that VersionMetadataManagement does not write the
// version file, so when FeatureGateLockManagement fails the rollback
// path described in the error message remains viable.
func TestUpgradeThenRollbackWithFeatureGateLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prerun-rollback-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalVersionPath := versionFilePath
	versionFilePath = filepath.Join(tmpDir, "version")
	defer func() { versionFilePath = originalVersionPath }()

	originalLockPath := featureGateLockFilePath
	featureGateLockFilePath = filepath.Join(tmpDir, "no-upgrade")
	defer func() { featureGateLockFilePath = originalLockPath }()

	originalGetExecVer := getExecutableVersion
	defer func() { getExecutableVersion = originalGetExecVer }()

	v421 := versionMetadata{Major: 4, Minor: 21, Patch: 0}
	v422 := versionMetadata{Major: 4, Minor: 22, Patch: 0}

	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.EnableDisableFeatures{
					Enabled: []string{"FeatureA"},
				},
			},
		},
	}

	// --- Phase 1: Initial run at 4.21.0 ---
	getExecutableVersion = func() (versionMetadata, error) { return v421, nil }

	if err := os.WriteFile(versionFilePath, []byte(`{"version":"4.21.0","boot_id":"boot-1"}`), 0600); err != nil {
		t.Fatal(err)
	}

	if err := FeatureGateLockManagement(cfg); err != nil {
		t.Fatalf("Phase 1: FeatureGateLockManagement should succeed on first run: %v", err)
	}

	lockFile, err := readFeatureGateLockFile(featureGateLockFilePath)
	if err != nil {
		t.Fatalf("Phase 1: failed to read lock file: %v", err)
	}
	if lockFile.Version != v421 {
		t.Fatalf("Phase 1: lock file version = %s, want %s", lockFile.Version.String(), v421.String())
	}

	// --- Phase 2: Upgrade to 4.22.0 ---
	getExecutableVersion = func() (versionMetadata, error) { return v422, nil }

	// VersionMetadataManagement validates but does not write the version file
	if err := VersionMetadataManagement(); err != nil {
		t.Fatalf("Phase 2: VersionMetadataManagement should succeed: %v", err)
	}

	// FeatureGateLockManagement should FAIL (locked=4.21, exec=4.22)
	err = FeatureGateLockManagement(cfg)
	if err == nil {
		t.Fatal("Phase 2: FeatureGateLockManagement should fail due to version mismatch")
	}

	// Because FeatureGateLockManagement failed, WriteVersionMetadata is NOT
	// called. Verify the version file still contains 4.21.0.
	dataVer, err := getVersionOfData()
	if err != nil {
		t.Fatalf("Phase 2: failed to read version file: %v", err)
	}
	if dataVer != v421 {
		t.Fatalf("Phase 2: version file should still be %s after failed pre-run, got %s", v421.String(), dataVer.String())
	}

	// --- Phase 3: Roll back to 4.21.0 ---
	getExecutableVersion = func() (versionMetadata, error) { return v421, nil }

	if err := VersionMetadataManagement(); err != nil {
		t.Fatalf("Phase 3: VersionMetadataManagement should succeed after rollback: %v", err)
	}

	if err := FeatureGateLockManagement(cfg); err != nil {
		t.Fatalf("Phase 3: FeatureGateLockManagement should succeed after rollback: %v", err)
	}

	if err := WriteVersionMetadata(); err != nil {
		t.Fatalf("Phase 3: WriteVersionMetadata should succeed: %v", err)
	}

	dataVer, err = getVersionOfData()
	if err != nil {
		t.Fatalf("Phase 3: failed to read version file after write: %v", err)
	}
	if dataVer != v421 {
		t.Fatalf("Phase 3: version file should be %s after rollback, got %s", v421.String(), dataVer.String())
	}
}

// TestVersionMetadataManagementDoesNotWriteVersionFile verifies that
// VersionMetadataManagement is read-only and WriteVersionMetadata is
// required to persist the version.
func TestVersionMetadataManagementDoesNotWriteVersionFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prerun-no-write-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	originalVersionPath := versionFilePath
	versionFilePath = filepath.Join(tmpDir, "version")
	defer func() { versionFilePath = originalVersionPath }()

	originalGetExecVer := getExecutableVersion
	defer func() { getExecutableVersion = originalGetExecVer }()

	v414 := versionMetadata{Major: 4, Minor: 14, Patch: 0}
	v415 := versionMetadata{Major: 4, Minor: 15, Patch: 0}

	if err := os.WriteFile(versionFilePath, []byte(`{"version":"4.14.0","boot_id":"b"}`), 0600); err != nil {
		t.Fatal(err)
	}

	getExecutableVersion = func() (versionMetadata, error) { return v415, nil }

	if err := VersionMetadataManagement(); err != nil {
		t.Fatalf("VersionMetadataManagement should succeed: %v", err)
	}

	// Version file should still be 4.14.0 — no write happened yet
	dataVer, err := getVersionOfData()
	if err != nil {
		t.Fatalf("failed to read version file: %v", err)
	}
	if dataVer != v414 {
		t.Fatalf("version file should still be %s after management, got %s", v414.String(), dataVer.String())
	}

	if err := WriteVersionMetadata(); err != nil {
		t.Fatalf("WriteVersionMetadata should succeed: %v", err)
	}

	dataVer, err = getVersionOfData()
	if err != nil {
		t.Fatalf("failed to read version file after write: %v", err)
	}
	if dataVer != v415 {
		t.Fatalf("version file should be %s after write, got %s", v415.String(), dataVer.String())
	}
}
