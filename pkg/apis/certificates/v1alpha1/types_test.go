package v1alpha1_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
)

func TestCertificateStatusRoundTrip(t *testing.T) {
	const document = `{
		"apiVersion":"microshift.openshift.io/v1alpha1",
		"kind":"CertificateStatusList",
		"generatedAt":"2026-09-08T10:30:00Z",
		"config":{"forceRestartOnRedZone":false,"servingValidity":"8760h","caValidity":"87600h"},
		"items":[{
			"service":"etcd","name":"etcd-serving","role":"peer","rotationPolicy":"extended","zone":"red",
			"notBefore":"2016-09-07T10:30:00Z","notAfter":"2026-09-07T10:30:00Z","remainingSeconds":-86400
		}],
		"warnings":[]
	}`
	var status certificatesv1alpha1.CertificateStatusList
	require.NoError(t, json.Unmarshal([]byte(document), &status))
	require.Equal(t, certificatesv1alpha1.CertificateRolePeer, status.Items[0].Role)
	require.Equal(t, certificatesv1alpha1.RotationPolicyExtended, status.Items[0].RotationPolicy)
	require.Equal(t, certificatesv1alpha1.CertificateZoneRed, status.Items[0].Zone)
	require.False(t, status.Config.ForceRestartOnRedZone)
	require.Equal(t, int64(-86400), status.Items[0].RemainingSeconds)
	require.NotNil(t, status.Warnings)

	encoded, err := yaml.Marshal(status)
	require.NoError(t, err)
	var decoded certificatesv1alpha1.CertificateStatusList
	require.NoError(t, yaml.UnmarshalStrict(encoded, &decoded))
	roundTrip, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.JSONEq(t, document, string(roundTrip))
}

func TestCertificateRenewalResultRoundTrip(t *testing.T) {
	now := metav1.NewTime(time.Date(2026, time.September, 8, 10, 30, 0, 0, time.UTC))
	for _, tt := range []struct {
		status  certificatesv1alpha1.RenewalStatus
		dryRun  bool
		changed bool
	}{
		{certificatesv1alpha1.RenewalStatusValidated, true, false},
		{certificatesv1alpha1.RenewalStatusCompleted, false, true},
	} {
		t.Run(string(tt.status), func(t *testing.T) {
			result := certificatesv1alpha1.CertificateRenewalResult{
				APIVersion:  certificatesv1alpha1.APIVersion,
				Kind:        certificatesv1alpha1.CertificateRenewalResultKind,
				GeneratedAt: now,
				Mode:        certificatesv1alpha1.RenewalModeCA,
				Status:      tt.status,
				DryRun:      tt.dryRun,
				Items: []certificatesv1alpha1.CertificateRenewalItem{{
					Service:         "service-ca",
					Name:            "service-ca",
					Role:            certificatesv1alpha1.CertificateRoleCA,
					CurrentNotAfter: now,
					NewNotAfter:     metav1.NewTime(now.Add(24 * time.Hour)),
					Changed:         tt.changed,
				}},
				Impact:   certificatesv1alpha1.CertificateRenewalImpact{ServiceRestartRequired: true},
				Warnings: []string{},
			}
			encodedJSON, err := json.Marshal(result)
			require.NoError(t, err)
			require.Contains(t, string(encodedJSON), `"status":"`+string(tt.status)+`"`)
			require.Contains(t, string(encodedJSON), `"parentCA":null`)
			require.NotContains(t, string(encodedJSON), `"planned"`)
			encodedYAML, err := yaml.Marshal(result)
			require.NoError(t, err)
			var decoded certificatesv1alpha1.CertificateRenewalResult
			require.NoError(t, yaml.UnmarshalStrict(encodedYAML, &decoded))
			require.Equal(t, tt.status, decoded.Status)
			require.Equal(t, tt.dryRun, decoded.DryRun)
			require.Equal(t, tt.changed, decoded.Items[0].Changed)
			roundTrip, err := json.Marshal(decoded)
			require.NoError(t, err)
			require.JSONEq(t, string(encodedJSON), string(roundTrip))
		})
	}
}

func TestCertificateErrorRoundTrip(t *testing.T) {
	for _, details := range []string{`null`, `{"service":"microshift.service","state":"active"}`} {
		document := `{
			"apiVersion":"microshift.openshift.io/v1alpha1","kind":"Error","generatedAt":"2026-09-08T10:30:00Z",
			"code":"MicroShiftRunning","message":"MicroShift must be stopped before renewal.","details":` + details + `
		}`
		var failure certificatesv1alpha1.Error
		require.NoError(t, json.Unmarshal([]byte(document), &failure))
		require.Equal(t, certificatesv1alpha1.ErrorCodeMicroShiftRunning, failure.Code)
		encoded, err := yaml.Marshal(failure)
		require.NoError(t, err)
		var decoded certificatesv1alpha1.Error
		require.NoError(t, yaml.UnmarshalStrict(encoded, &decoded))
		roundTrip, err := json.Marshal(decoded)
		require.NoError(t, err)
		require.JSONEq(t, document, string(roundTrip))
	}
}
