package prerun

import (
	"encoding/json"
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
				CustomNoUpgrade: config.CustomNoUpgrade{
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
				CustomNoUpgrade: config.CustomNoUpgrade{},
			},
			wantErr: false,
		},
		{
			name: "DevPreviewNoUpgrade",
			lockFile: featureGateLockFile{
				FeatureSet:      config.FeatureSetDevPreviewNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{},
			},
			wantErr: false,
		},
		{
			name: "empty feature gates",
			lockFile: featureGateLockFile{
				FeatureSet:      "",
				CustomNoUpgrade: config.CustomNoUpgrade{},
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
				CustomNoUpgrade: config.CustomNoUpgrade{
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
			want: false,
		},
		{
			name: "CustomNoUpgrade without any enabled/disabled",
			fg: config.FeatureGates{
				FeatureSet:      config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCustomFeatureGatesConfigured(tt.fg)
			if got != tt.want {
				t.Errorf("isCustomFeatureGatesConfigured() = %v, want %v", got, tt.want)
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
				CustomNoUpgrade: config.CustomNoUpgrade{
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

func TestCompareFeatureGates(t *testing.T) {
	tests := []struct {
		name      string
		lockFile  featureGateLockFile
		current   config.FeatureGates
		wantMatch bool
	}{
		{
			name: "identical custom feature gates",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
					Enabled:  []string{"FeatureA", "FeatureB"},
					Disabled: []string{"FeatureC"},
				},
			},
			current: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
					Enabled:  []string{"FeatureA", "FeatureB"},
					Disabled: []string{"FeatureC"},
				},
			},
			wantMatch: true,
		},
		{
			name: "different enabled features",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
					Enabled: []string{"FeatureA"},
				},
			},
			current: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
					Enabled: []string{"FeatureB"},
				},
			},
			wantMatch: false,
		},
		{
			name: "different feature sets",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetTechPreviewNoUpgrade,
			},
			current: config.FeatureGates{
				FeatureSet: config.FeatureSetDevPreviewNoUpgrade,
			},
			wantMatch: false,
		},
		{
			name: "identical TechPreviewNoUpgrade",
			lockFile: featureGateLockFile{
				FeatureSet: config.FeatureSetTechPreviewNoUpgrade,
			},
			current: config.FeatureGates{
				FeatureSet: config.FeatureSetTechPreviewNoUpgrade,
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareFeatureGates(tt.lockFile, tt.current)
			gotMatch := err == nil
			if gotMatch != tt.wantMatch {
				t.Errorf("compareFeatureGates() match = %v, want %v, error = %v", gotMatch, tt.wantMatch, err)
			}
		})
	}
}

func TestFeatureGateLockManagement_FirstRun(t *testing.T) {
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

	// Create a version file to simulate existing data (version file uses JSON format)
	versionData := versionFile{
		Version: versionMetadata{Major: 4, Minor: 18, Patch: 0},
		BootID:  "test-boot",
	}
	versionJSON, _ := json.Marshal(versionData)
	if err := os.WriteFile(versionFilePath, versionJSON, 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
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

	// Create a version file to simulate existing data (version file uses JSON format)
	versionData := versionFile{
		Version: versionMetadata{Major: 4, Minor: 18, Patch: 0},
		BootID:  "test-boot",
	}
	versionJSON, _ := json.Marshal(versionData)
	if err := os.WriteFile(versionFilePath, versionJSON, 0600); err != nil {
		t.Fatal(err)
	}

	// Create lockFile file with initial config
	initialLock := featureGateLockFile{
		FeatureSet: config.FeatureSetCustomNoUpgrade,
		CustomNoUpgrade: config.CustomNoUpgrade{
			Enabled: []string{"FeatureA"},
		},
	}
	if err := writeFeatureGateLockFile(featureGateLockFilePath, initialLock); err != nil {
		t.Fatal(err)
	}

	// Try to run with different config - should fail
	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
					Enabled: []string{"FeatureB"}, // Different feature
				},
			},
		},
	}

	err = FeatureGateLockManagement(cfg)
	if err == nil {
		t.Error("FeatureGateLockManagement() should have failed with config change")
	}
}

func TestFeatureGateLockManagement_VersionChange(t *testing.T) {
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

	// Create a version file with NEW version (simulating upgrade) (version file uses JSON format)
	versionData := versionFile{
		Version: versionMetadata{Major: 4, Minor: 19, Patch: 0}, // Newer version
		BootID:  "test-boot",
	}
	versionJSON, _ := json.Marshal(versionData)
	if err := os.WriteFile(versionFilePath, versionJSON, 0600); err != nil {
		t.Fatal(err)
	}

	// Create lockFile file with OLD version
	lockFile := featureGateLockFile{
		FeatureSet: config.FeatureSetCustomNoUpgrade,
		CustomNoUpgrade: config.CustomNoUpgrade{
			Enabled: []string{"FeatureA"},
		},
		Version: versionMetadata{Major: 4, Minor: 18, Patch: 0}, // Older version
	}
	if err := writeFeatureGateLockFile(featureGateLockFilePath, lockFile); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ApiServer: config.ApiServer{
			FeatureGates: config.FeatureGates{
				FeatureSet: config.FeatureSetCustomNoUpgrade,
				CustomNoUpgrade: config.CustomNoUpgrade{
					Enabled: []string{"FeatureA"},
				},
			},
		},
	}

	err = FeatureGateLockManagement(cfg)
	if err == nil {
		t.Error("FeatureGateLockManagement() should have failed with version change")
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
