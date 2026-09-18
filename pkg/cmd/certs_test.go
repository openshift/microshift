package cmd

import (
	"bytes"
	"crypto/x509"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"

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
	require.Equal(t, certificateStatusAPIVersion, status.APIVersion)
	require.Equal(t, "CertificateStatusList", status.Kind)
	require.Equal(t, "2026-09-08T10:30:00Z", status.GeneratedAt)
	require.Equal(t, certificateStatusConfig{
		ForceRestartOnRedZone: true,
		ServingValidity:       "8760h",
		CAValidity:            "87600h",
	}, status.Config)
	require.Equal(t, []string{}, status.Warnings)
	require.Equal(t, []string{"cert-a", "cert-z", "cert-b"}, []string{status.Items[0].Name, status.Items[1].Name, status.Items[2].Name})
	require.Equal(t, certchains.CertificateZoneYellow, status.Items[0].Zone)
	require.Equal(t, int64((500*time.Hour)/time.Second), status.Items[0].RemainingSeconds)
}

func TestCertStatusOutput(t *testing.T) {
	now := time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC)
	status, err := newCertificateStatusList(certchains.CertificateInventory{
		certificateInventoryEntry("etcd", "etcd-serving", certchains.CertificateRolePeer, certchains.RotationPolicyExtended, now.Add(-900*time.Hour), now.Add(100*time.Hour)),
	}, []string{"configuration warning"}, now)
	require.NoError(t, err)

	var jsonOutput bytes.Buffer
	require.NoError(t, writeCertificateStatusJSON(&jsonOutput, status))
	require.JSONEq(t, `{
		"apiVersion":"microshift.openshift.io/v1alpha1",
		"kind":"CertificateStatusList",
		"generatedAt":"2026-09-08T10:30:00Z",
		"config":{"forceRestartOnRedZone":true,"servingValidity":"8760h","caValidity":"87600h"},
		"items":[{
			"service":"etcd",
			"name":"etcd-serving",
			"role":"peer",
			"rotationPolicy":"extended",
			"zone":"red",
			"notBefore":"2026-08-01T22:30:00Z",
			"notAfter":"2026-09-12T14:30:00Z",
			"remainingSeconds":360000
		}],
		"warnings":["configuration warning"]
	}`, jsonOutput.String())

	var tableOutput bytes.Buffer
	require.NoError(t, writeCertificateStatusTable(&tableOutput, status))
	require.Contains(t, tableOutput.String(), "SERVICE  CERTIFICATE   STATUS  EXPIRY")
	require.Contains(t, tableOutput.String(), "etcd     etcd-serving  Red")
	require.Contains(t, tableOutput.String(), "Expiring  Expires in 5 days")
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

	options.output = "yaml"
	require.EqualError(t, options.run(), `unsupported output format "yaml"; supported formats: json`)
}

func TestCertsCommandRequiresPrivileges(t *testing.T) {
	wantErr := errors.New("privileges required")
	command := newCertsCommand(genericclioptions.IOStreams{}, func() error { return wantErr })
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
