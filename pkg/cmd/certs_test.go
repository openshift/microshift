package cmd

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

func TestNewCertificateStatusList(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 500, time.UTC)
	inventory := certchains.CertificateInventory{
		certificateInventoryEntry("service-b", "cert-b", certchains.CertificateRoleCA, certchains.RotationPolicyExtended, now.Add(-900*time.Hour), now.Add(100*time.Hour)),
		certificateInventoryEntry("service-a", "cert-z", certchains.CertificateRoleServing, certchains.RotationPolicyStandard, now.Add(-400*time.Hour), now.Add(600*time.Hour)),
		certificateInventoryEntry("service-a", "cert-a", certchains.CertificateRoleClient, certchains.RotationPolicyStandard, now.Add(-500*time.Hour), now.Add(500*time.Hour)),
	}

	status, err := newCertificateStatusList(inventory, nil, now)
	require.NoError(t, err)
	require.Equal(t, certificatesv1alpha1.APIVersion, status.APIVersion)
	require.Equal(t, "CertificateStatusList", status.Kind)
	require.Equal(t, "2026-09-08T10:30:00Z", status.GeneratedAt.Format(time.RFC3339))
	require.Equal(t, certificatesv1alpha1.CertificateStatusConfig{
		ForceRestartOnExpirationImminent: true,
		ServingValidity:                  "8760h",
		CAValidity:                       "87600h",
	}, status.Config)
	require.Equal(t, []string{}, status.Warnings)
	require.Equal(t, []string{"cert-a", "cert-z", "cert-b"}, []string{status.Items[0].Name, status.Items[1].Name, status.Items[2].Name})
	require.Equal(t, certificatesv1alpha1.CertificateStatusExpiresSoon, status.Items[0].Status)
	require.Equal(t, int64((500*time.Hour)/time.Second), status.Items[0].RemainingSeconds)
	require.Equal(t, "cert-b", inventory[0].Name, "building status must not reorder the inventory")
}

func TestCertStatusOutput(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	status, err := newCertificateStatusList(certchains.CertificateInventory{
		certificateInventoryEntry("etcd", "etcd-serving", certchains.CertificateRolePeer, certchains.RotationPolicyExtended, now.Add(-900*time.Hour), now.Add(100*time.Hour)),
	}, []string{"configuration warning"}, now)
	require.NoError(t, err)

	var jsonOutput bytes.Buffer
	require.NoError(t, writeCertificateObject(&jsonOutput, &status, certificateOutputJSON))
	require.JSONEq(t, `{
		"apiVersion":"microshift.openshift.io/v1alpha1",
		"kind":"CertificateStatusList",
		"generatedAt":"2026-09-08T10:30:00Z",
		"config":{"forceRestartOnExpirationImminent":true,"servingValidity":"8760h","caValidity":"87600h"},
		"items":[{
			"service":"etcd",
			"name":"etcd-serving",
			"role":"peer",
			"rotationPolicy":"extended",
			"status":"ExpirationImminent",
			"notBefore":"2026-08-01T22:30:00Z",
			"notAfter":"2026-09-12T14:30:00Z",
			"remainingSeconds":360000
		}],
		"warnings":["configuration warning"]
	}`, jsonOutput.String())

	var yamlOutput bytes.Buffer
	require.NoError(t, writeCertificateObject(&yamlOutput, &status, certificateOutputYAML))
	yamlAsJSON, err := yaml.YAMLToJSONStrict(yamlOutput.Bytes())
	require.NoError(t, err)
	require.JSONEq(t, jsonOutput.String(), string(yamlAsJSON))

	// A caller can deserialize the exported type and render the same table
	// without access to the internal inventory or its x509 certificates.
	var decoded certificatesv1alpha1.CertificateStatusList
	require.NoError(t, yaml.UnmarshalStrict(yamlOutput.Bytes(), &decoded))
	var decodedTable bytes.Buffer
	require.NoError(t, writeCertificateStatusTable(&decodedTable, decoded))

	var tableOutput bytes.Buffer
	require.NoError(t, writeCertificateStatusTable(&tableOutput, status))
	require.Equal(t, tableOutput.String(), decodedTable.String())
	require.Contains(t, tableOutput.String(), "SERVICE  CERTIFICATE   STATUS              EXPIRY")
	require.Contains(t, tableOutput.String(), "etcd     etcd-serving  ExpirationImminent")
	require.Contains(t, tableOutput.String(), "Expires in 5 days, which is at or below the critical threshold of 5 days")
}

func TestCertStatusOptionsRun(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	warnings := []string(nil)
	options := certStatusOptions{
		IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
		output:    "json",
		now:       func() time.Time { return now },
		loadConfig: func() (*config.Config, error) {
			return &config.Config{Warnings: warnings}, nil
		},
		loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
			return certchains.CertificateInventory{
				certificateInventoryEntry("etcd", "etcd-signer", certchains.CertificateRoleCA, certchains.RotationPolicyExtended, now.Add(-100*time.Hour), now.Add(900*time.Hour)),
			}, nil
		},
	}
	require.NoError(t, options.run())
	require.Contains(t, stdout.String(), `"kind": "CertificateStatusList"`)

	stdout.Reset()
	options.output = ""
	warnings = []string{"configuration warning"}
	require.NoError(t, options.run())
	require.NotContains(t, stdout.String(), "configuration warning")
	require.Equal(t, "WARNING: configuration warning\n", stderr.String())

	stdout.Reset()
	options.output = "xml"
	require.EqualError(t, options.run(), `unsupported output format "xml"; supported formats: json, yaml`)
	require.Empty(t, stdout.String())
}

func TestCertStatusStates(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	for _, tt := range []struct {
		name                string
		notBefore, notAfter time.Duration
		want                certificatesv1alpha1.CertificateStatus
		message             string
	}{
		{"healthy", -100 * time.Hour, 900 * time.Hour, certificatesv1alpha1.CertificateStatusHealthy, "Valid for 38 days"},
		{"expires soon", -500 * time.Hour, 500 * time.Hour, certificatesv1alpha1.CertificateStatusExpiresSoon, "Expires in 21 days, which is at or below the warning threshold of 25 days"},
		{"expiration imminent", -800 * time.Hour, 200 * time.Hour, certificatesv1alpha1.CertificateStatusExpirationImminent, "Expires in 9 days, which is at or below the critical threshold of 14 days"},
		{"expired", -1000 * time.Hour, -time.Hour, certificatesv1alpha1.CertificateStatusExpired, "Expired 1 days ago"},
		{"exactly at expiry", -1000 * time.Hour, 0, certificatesv1alpha1.CertificateStatusExpired, "Expired 0 days ago"},
		{"not yet valid", time.Hour, 1000 * time.Hour, certificatesv1alpha1.CertificateStatusExpirationImminent, "Valid in 1 days"},
	} {
		for _, format := range []string{"", certificateOutputJSON, certificateOutputYAML} {
			t.Run(tt.name+"/"+format, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				options := &certStatusOptions{
					IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
					now:       func() time.Time { return now },
					loadConfig: func() (*config.Config, error) {
						return &config.Config{}, nil
					},
					loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
						return certchains.CertificateInventory{
							certificateInventoryEntry("test", "test-serving", certchains.CertificateRoleServing,
								certchains.RotationPolicyStandard, now.Add(tt.notBefore), now.Add(tt.notAfter)),
						}, nil
					},
				}
				command := newCertsCommand(options, func() error { return nil })
				args := []string{"status"}
				if format != "" {
					args = append(args, "-o", format)
				}
				require.Zero(t, RunCertsCommand(command, args), "certificate state must not change the exit code")
				require.Empty(t, stderr.String())
				if format == "" {
					rows := strings.Split(strings.TrimSpace(stdout.String()), "\n")
					require.Len(t, rows, 2)
					require.Equal(t, []string{"SERVICE", "CERTIFICATE", "STATUS", "EXPIRY", "MESSAGE"}, strings.Fields(rows[0]))
					fields := strings.Fields(rows[1])
					require.GreaterOrEqual(t, len(fields), 5)
					require.Equal(t, string(tt.want), fields[2])
					require.Equal(t, tt.message, strings.Join(fields[4:], " "))
					return
				}
				decoded, gvk, err := certificateCodecs.UniversalDeserializer().Decode(stdout.Bytes(), nil, nil)
				require.NoError(t, err)
				require.NotContains(t, stdout.String(), "zone", "the legacy zone field must not be emitted")
				require.NotContains(t, stdout.String(), "forceRestartOnRedZone")
				require.Equal(t, certificatesv1alpha1.GroupVersion.WithKind(certificatesv1alpha1.CertificateStatusListKind), *gvk)
				status, ok := decoded.(*certificatesv1alpha1.CertificateStatusList)
				require.True(t, ok)
				require.Len(t, status.Items, 1)
				require.Equal(t, tt.want, status.Items[0].Status)
				require.Equal(t, int64(tt.notAfter/time.Second), status.Items[0].RemainingSeconds)
			})
		}
	}
}

func TestCertStatusStructuredOutputFlags(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	for _, args := range [][]string{{"-o", "json"}, {"--output=json"}, {"-o", "yaml"}, {"--output=yaml"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			options := &certStatusOptions{
				IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
				now:       func() time.Time { return now },
				loadConfig: func() (*config.Config, error) {
					return &config.Config{Warnings: []string{"configuration warning"}}, nil
				},
				loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
					return certchains.CertificateInventory{}, nil
				},
			}
			command := newCertsCommand(options, func() error { return nil })
			require.Equal(t, 0, RunCertsCommand(command, append([]string{"status"}, args...)))
			require.Empty(t, stderr.String())
			var status certificatesv1alpha1.CertificateStatusList
			if options.output == certificateOutputJSON {
				decoder := json.NewDecoder(&stdout)
				require.NoError(t, decoder.Decode(&status))
				require.ErrorIs(t, decoder.Decode(new(any)), io.EOF)
			} else {
				require.NoError(t, yaml.UnmarshalStrict(stdout.Bytes(), &status))
			}
			require.Equal(t, certificatesv1alpha1.CertificateStatusListKind, status.Kind)
			require.NotNil(t, status.Items, "empty inventories must serialize as []")
			require.Equal(t, []string{"configuration warning"}, status.Warnings)
		})
	}
}

func TestCertificateStatusThresholdMessages(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	const day = 24 * time.Hour
	for _, tt := range []struct {
		name                string
		policy              certchains.RotationPolicy
		validity, remaining time.Duration
		message             string
	}{
		{"standard healthy", certchains.RotationPolicyStandard, 365 * day, 300 * day,
			"Valid for 300 days"},
		{"standard warning boundary", certchains.RotationPolicyStandard, 1000 * day, 583 * day,
			"Expires in 583 days, which is at or below the warning threshold of 583 days"},
		{"standard critical boundary", certchains.RotationPolicyStandard, 1000 * day, 333 * day,
			"Expires in 333 days, which is at or below the critical threshold of 333 days"},
		{"extended healthy", certchains.RotationPolicyExtended, 3650 * day, 1000 * day,
			"Valid for 1000 days"},
		{"extended warning boundary", certchains.RotationPolicyExtended, 1000 * day, 150 * day,
			"Expires in 150 days, which is at or below the warning threshold of 150 days"},
		{"extended critical boundary", certchains.RotationPolicyExtended, 1000 * day, 100 * day,
			"Expires in 100 days, which is at or below the critical threshold of 100 days"},
		{"custom six week lifetime", certchains.RotationPolicyStandard, 42 * day, 20 * day,
			"Expires in 20 days, which is at or below the warning threshold of 25 days"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			status, err := newCertificateStatusList(certchains.CertificateInventory{
				certificateInventoryEntry("test", "test-certificate", certchains.CertificateRoleServing, tt.policy,
					now.Add(tt.remaining-tt.validity), now.Add(tt.remaining)),
			}, nil, now)
			require.NoError(t, err)
			message, err := humanCertificateStatus(status.Items[0], now)
			require.NoError(t, err)
			require.Equal(t, tt.message, message)
			status.Items[0].RotationPolicy = "invalid"
			var out bytes.Buffer
			require.ErrorContains(t, writeCertificateStatusTable(&out, status), "unknown rotation policy")
			require.Empty(t, out.String(), "invalid policy must not produce a partial table")
		})
	}
}

func TestCertStatusOutputWriteFailure(t *testing.T) {
	wantErr := errors.New("output unavailable")
	status, err := newCertificateStatusList(nil, nil, time.Now())
	require.NoError(t, err)
	require.ErrorIs(t, writeCertificateObject(failingCertificateWriter{wantErr}, &status, certificateOutputJSON), wantErr)
	require.ErrorIs(t, writeCertificateObject(failingCertificateWriter{wantErr}, &status, certificateOutputYAML), wantErr)
}

type failingCertificateWriter struct {
	err error
}

func (w failingCertificateWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestCertsCommandRequiresPrivileges(t *testing.T) {
	wantErr := errors.New("privileges required")
	command := newCertsCommand(&certStatusOptions{}, func() error { return wantErr })
	command.SetArgs([]string{"status"})
	require.ErrorIs(t, command.Execute(), wantErr)

	statusCommand, _, err := command.Find([]string{"status"})
	require.NoError(t, err)
	require.NotNil(t, statusCommand.Flags().Lookup("output"))
	require.Equal(t, "o", statusCommand.Flags().Lookup("output").Shorthand)
}

func certificateInventoryEntry(service, name string, role certchains.CertificateRole, policy certchains.RotationPolicy, notBefore, notAfter time.Time) certchains.CertificateInventoryEntry {
	return certchains.CertificateInventoryEntry{
		Path:           strings.Split(name, "/"),
		Name:           name,
		Service:        service,
		Role:           role,
		RotationPolicy: policy,
		Certificate: x509.Certificate{
			NotBefore: notBefore,
			NotAfter:  notAfter,
		},
	}
}
