package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

func TestCertificateRenewalApply(t *testing.T) {
	for _, mode := range []string{"serving", "ca"} {
		t.Run(mode, func(t *testing.T) {
			cfg := &config.Config{
				Network:   config.Network{ServiceNetwork: []string{"10.43.0.0/16"}},
				Node:      config.Node{HostnameOverride: "test-node", NodeIP: "192.0.2.10"},
				DNS:       config.DNS{BaseDomain: "example.test"},
				ApiServer: config.ApiServer{AdvertiseAddress: "10.44.0.0", URL: "https://127.0.0.1:6443", Port: 6443},
			}
			dataDir := t.TempDir()
			builder, err := certificateChainsSetup(cfg, dataDir)
			require.NoError(t, err)
			chains, err := builder.Complete()
			require.NoError(t, err)
			require.NoError(t, initKubeconfigs(cfg, chains, dataDir))
			unrelated := filepath.Join(dataDir, "resources", "unrelated")
			require.NoError(t, os.Mkdir(unrelated, 0700))
			require.NoError(t, os.WriteFile(filepath.Join(unrelated, "kubeconfig"), []byte("unrelated component data"), 0600))
			before, err := builder.LoadInventory()
			require.NoError(t, err)
			beforeFiles := certificateFileDigests(t, dataDir)
			plan, err := builder.PlanRenewal(mode == "ca", time.Now())
			require.NoError(t, err)
			require.Equal(t, beforeFiles, certificateFileDigests(t, dataDir))
			wantCount := 19
			if mode == "ca" {
				wantCount = 31
			}
			require.Len(t, plan, wantCount)
			renewed, err := applyCertificateRenewal(cfg, dataDir, mode == "ca")
			require.NoError(t, err)
			require.Len(t, renewed, wantCount)
			contents, err := os.ReadFile(filepath.Join(unrelated, "kubeconfig"))
			require.NoError(t, err)
			require.Equal(t, "unrelated component data", string(contents))
			_, err = validateCertificateRenewal(cfg, dataDir, before, mode == "ca")
			require.NoError(t, err)
			_, err = os.Stat(filepath.Join(dataDir, ".cert-renewal"))
			require.ErrorIs(t, err, os.ErrNotExist)
			if mode == "serving" {
				afterFiles := certificateFileDigests(t, dataDir)
				for path, digest := range beforeFiles {
					if filepath.Base(path) == "ca.crt" || filepath.Base(path) == "ca.key" {
						require.Equal(t, digest, afterFiles[path], "CA material must remain unchanged: %s", path)
					}
				}
			}
		})
	}
}

func TestCertRenewCommand(t *testing.T) {
	for _, output := range []string{"", certificateOutputJSON, certificateOutputYAML} {
		for _, dryRun := range []bool{true, false} {
			t.Run(output+"/"+map[bool]string{true: "dry-run", false: "apply"}[dryRun], func(t *testing.T) {
				now := time.Now().UTC().Truncate(time.Second)
				entry := certificateInventoryEntry("test", "client", certchains.CertificateRoleClient, certchains.RotationPolicyStandard, now.Add(-time.Hour), now.Add(time.Hour))
				entry.ParentCA = "test-ca"
				plan := []certchains.CertificateRenewalPlanEntry{{CertificateInventoryEntry: entry, NewNotAfter: now.Add(24 * time.Hour)}}
				var stdout, stderr bytes.Buffer
				applied, released := false, false
				options := &certRenewOptions{
					IOStreams:  genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
					now:        func() time.Time { return now },
					loadConfig: func() (*config.Config, error) { return &config.Config{}, nil },
					prepare: func(writing bool) (func(), error) {
						require.Equal(t, !dryRun, writing)
						return func() { released = true }, nil
					},
					plan: func(*config.Config, bool, time.Time) ([]certchains.CertificateRenewalPlanEntry, error) {
						return plan, nil
					},
					apply: func(*config.Config, bool) (certchains.CertificateInventory, error) {
						applied = true
						entry.Certificate.NotAfter = plan[0].NewNotAfter
						return certchains.CertificateInventory{entry}, nil
					},
				}
				root := newCertsCommand(&certStatusOptions{IOStreams: options.IOStreams}, func() error { return nil })
				root.AddCommand(newCertsRenewCommand(options))
				args := []string{"renew", "--serving"}
				if dryRun {
					args = append(args, "--dry-run")
				}
				if output != "" {
					args = append(args, "-o", output)
				}
				require.Zero(t, RunCertsCommand(root, args), stderr.String())
				require.Equal(t, !dryRun, applied)
				require.True(t, released)
				if output == "" {
					require.Contains(t, stdout.String(), "client")
					require.Contains(t, stderr.String(), "WARNING:")
					return
				}
				require.Empty(t, stderr.String())
				var result certificatesv1alpha1.CertificateRenewalResult
				object, gvk, err := certificateCodecs.UniversalDeserializer().Decode(stdout.Bytes(), nil, nil)
				require.NoError(t, err)
				require.IsType(t, &certificatesv1alpha1.CertificateRenewalResult{}, object)
				require.Equal(t, certificatesv1alpha1.GroupVersion.WithKind(certificatesv1alpha1.CertificateRenewalResultKind), *gvk)
				if output == certificateOutputJSON {
					require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
				} else {
					require.NoError(t, yaml.UnmarshalStrict(stdout.Bytes(), &result))
				}
				require.Equal(t, certificatesv1alpha1.CertificateRenewalResultKind, result.Kind)
				require.Equal(t, dryRun, result.DryRun)
				require.Equal(t, !dryRun, result.Items[0].Changed)
				require.Equal(t, "test-ca", *result.Items[0].ParentCA)
				require.True(t, plan[0].NewNotAfter.Equal(result.Items[0].NewNotAfter.Time))
				if dryRun {
					require.Equal(t, certificatesv1alpha1.RenewalStatusValidated, result.Status)
				} else {
					require.Equal(t, certificatesv1alpha1.RenewalStatusCompleted, result.Status)
				}
				require.True(t, result.Impact.ServiceRestartRequired)
				require.False(t, result.Impact.KubeconfigRedistributionRequired)
			})
		}
	}
}

func TestCertRenewErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		code certificatesv1alpha1.ErrorCode
	}{
		{"missing mode", nil, certificatesv1alpha1.ErrorCodeInvalidArguments},
		{"both modes", []string{"--ca", "--serving"}, certificatesv1alpha1.ErrorCodeInvalidArguments},
		{"configuration", []string{"--ca", "--dry-run"}, certificatesv1alpha1.ErrorCodeInvalidConfiguration},
		{"running", []string{"--ca"}, certificatesv1alpha1.ErrorCodeMicroShiftRunning},
		{"planning", []string{"--serving", "--dry-run"}, certificatesv1alpha1.ErrorCodeRenewalFailed},
		{"inventory", []string{"--serving", "--dry-run"}, certificatesv1alpha1.ErrorCodeCertificateInventoryFailed},
		{"recovery", []string{"--serving"}, certificatesv1alpha1.ErrorCodeRecoveryFailed},
		{"apply", []string{"--serving"}, certificatesv1alpha1.ErrorCodeRenewalFailed},
		{"rollback", []string{"--serving"}, certificatesv1alpha1.ErrorCodeRecoveryFailed},
	} {
		for _, output := range []string{"", certificateOutputJSON, certificateOutputYAML} {
			t.Run(tt.name+"/"+output, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				options := &certRenewOptions{
					IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr}, now: time.Now,
					loadConfig: func() (*config.Config, error) {
						if tt.name == "configuration" {
							return nil, errors.New("credential-bearing-configuration")
						}
						return &config.Config{}, nil
					},
					prepare: func(bool) (func(), error) {
						if tt.name == "running" || tt.name == "recovery" {
							return nil, &certificateCommandError{tt.code, errors.New("operation refused")}
						}
						return func() {}, nil
					},
					plan: func(*config.Config, bool, time.Time) ([]certchains.CertificateRenewalPlanEntry, error) {
						if tt.name == "inventory" {
							return nil, &certchains.InventoryReadError{Err: errors.New("cannot read inventory")}
						}
						if tt.name == "apply" || tt.name == "rollback" {
							return []certchains.CertificateRenewalPlanEntry{{}}, nil
						}
						return nil, errors.New("invalid key pair")
					},
					apply: func(*config.Config, bool) (certchains.CertificateInventory, error) {
						if tt.name == "rollback" {
							return nil, &certificates.RecoveryError{Err: errors.New("cannot restore backup")}
						}
						return nil, errors.New("cannot stage certificates")
					},
				}
				root := newCertsCommand(&certStatusOptions{IOStreams: options.IOStreams}, func() error { return nil })
				root.AddCommand(newCertsRenewCommand(options))
				args := append([]string{"renew", "-o", output}, tt.args...)
				require.Equal(t, 1, RunCertsCommand(root, args))
				require.Empty(t, stdout.String())
				require.NotContains(t, stderr.String(), "credential-bearing-configuration")
				if output == "" {
					require.Contains(t, stderr.String(), "Error:")
					return
				}
				var failure certificatesv1alpha1.Error
				if output == certificateOutputJSON {
					require.NoError(t, json.Unmarshal(stderr.Bytes(), &failure))
				} else {
					require.NoError(t, yaml.UnmarshalStrict(stderr.Bytes(), &failure))
				}
				require.Equal(t, tt.code, failure.Code)
			})
		}
	}
}

func certificateFileDigests(t *testing.T, dir string) map[string][32]byte {
	t.Helper()
	files := map[string][32]byte{}
	require.NoError(t, filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = sha256.Sum256(contents)
		}
		return nil
	}))
	return files
}
