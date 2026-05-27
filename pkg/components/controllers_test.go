package components

import (
	"strings"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func newTestConfig() *config.Config {
	return &config.Config{
		DNS: config.DNS{BaseDomain: "example.com"},
		Network: config.Network{
			ClusterNetwork: []string{"10.42.0.0/16"},
			ServiceNetwork: []string{"10.43.0.0/16"},
		},
		Ingress: config.IngressConfig{
			Ports: config.IngressPortsConfig{
				Http:  ptr.To(80),
				Https: ptr.To(443),
			},
			TuningOptions: config.IngressControllerTuningOptions{
				HeaderBufferBytes:           32768,
				HeaderBufferMaxRewriteBytes: 8192,
				HealthCheckInterval:         &metav1.Duration{Duration: 5 * time.Second},
				ClientTimeout:               &metav1.Duration{Duration: 30 * time.Second},
				ClientFinTimeout:            &metav1.Duration{Duration: 1 * time.Second},
				ServerTimeout:               &metav1.Duration{Duration: 30 * time.Second},
				ServerFinTimeout:            &metav1.Duration{Duration: 1 * time.Second},
				TunnelTimeout:               &metav1.Duration{Duration: 1 * time.Hour},
				TLSInspectDelay:             &metav1.Duration{Duration: 5 * time.Second},
				ThreadCount:                 4,
				MaxConnections:              50000,
			},
			ForwardedHeaderPolicy: "Append",
			TLSSecurityProfile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			},
			ServingCertificateSecret: "router-certs-default",
			DefaultHttpVersionPolicy: 1,
			LogEmptyRequests:         "Log",
			HTTPEmptyRequestsPolicy:  "Respond",
			AccessLogging: config.AccessLogging{
				Status: config.AccessLoggingDisabled,
			},
		},
	}
}

func requireStringParam(t *testing.T, params assets.RenderParams, key string) string {
	t.Helper()
	v, ok := params[key]
	if !ok {
		t.Fatalf("missing param %q", key)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("param %q has type %T, want string", key, v)
	}
	return s
}

func TestGenerateIngressParamsFIPSCiphers(t *testing.T) {
	t.Run("FIPS enabled filters non-FIPS TLS 1.3 ciphers", func(t *testing.T) {
		cfg := newTestConfig()
		params, err := generateIngressParams(cfg, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cipherSuites := requireStringParam(t, params, "RouterCiphersSuites")
		if strings.Contains(cipherSuites, "TLS_CHACHA20_POLY1305_SHA256") {
			t.Errorf("FIPS mode should filter out TLS_CHACHA20_POLY1305_SHA256, got: %s", cipherSuites)
		}
		if !strings.Contains(cipherSuites, "TLS_AES_128_GCM_SHA256") {
			t.Errorf("FIPS mode should keep TLS_AES_128_GCM_SHA256, got: %s", cipherSuites)
		}
		if !strings.Contains(cipherSuites, "TLS_AES_256_GCM_SHA384") {
			t.Errorf("FIPS mode should keep TLS_AES_256_GCM_SHA384, got: %s", cipherSuites)
		}
	})

	t.Run("non-FIPS keeps all TLS 1.3 ciphers", func(t *testing.T) {
		cfg := newTestConfig()
		params, err := generateIngressParams(cfg, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cipherSuites := requireStringParam(t, params, "RouterCiphersSuites")
		if !strings.Contains(cipherSuites, "TLS_CHACHA20_POLY1305_SHA256") {
			t.Errorf("non-FIPS mode should keep TLS_CHACHA20_POLY1305_SHA256, got: %s", cipherSuites)
		}
	})
}

func TestGenerateIngressParamsFIPSCurves(t *testing.T) {
	t.Run("FIPS enabled uses only NIST curves", func(t *testing.T) {
		cfg := newTestConfig()
		params, err := generateIngressParams(cfg, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		curves := requireStringParam(t, params, "RouterTLSCurves")
		if strings.Contains(curves, "X25519MLKEM768") {
			t.Errorf("FIPS mode should exclude X25519MLKEM768, got: %s", curves)
		}
		for _, c := range strings.Split(curves, ":") {
			if c == "X25519" {
				t.Errorf("FIPS mode should exclude X25519, got: %s", curves)
			}
		}
		if curves != "P-256:P-384:P-521" {
			t.Errorf("FIPS mode curves should be P-256:P-384:P-521, got: %s", curves)
		}
	})

	t.Run("non-FIPS includes PQC hybrid curve", func(t *testing.T) {
		cfg := newTestConfig()
		params, err := generateIngressParams(cfg, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		curves := requireStringParam(t, params, "RouterTLSCurves")
		if !strings.Contains(curves, "X25519MLKEM768") {
			t.Errorf("non-FIPS mode should include X25519MLKEM768, got: %s", curves)
		}
		if curves != "X25519MLKEM768:X25519:P-256:P-384:P-521" {
			t.Errorf("non-FIPS mode curves should be X25519MLKEM768:X25519:P-256:P-384:P-521, got: %s", curves)
		}
	})
}
