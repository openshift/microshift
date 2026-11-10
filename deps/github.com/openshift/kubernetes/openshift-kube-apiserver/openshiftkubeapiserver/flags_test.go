package openshiftkubeapiserver

import (
	"reflect"
	"testing"

	"github.com/openshift/api/config/v1"
	kubecontrolplanev1 "github.com/openshift/api/kubecontrolplane/v1"
	"github.com/spf13/pflag"
	kubeapiserveroptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

func TestSNICertKeys(t *testing.T) {
	testCases := []struct {
		names    []string
		expected string
	}{
		{names: []string{"foo"}, expected: "secret.crt,secret.key:foo"},
		{names: []string{"foo", "bar"}, expected: "secret.crt,secret.key:foo,bar"},
		{expected: "secret.crt,secret.key"},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := sniCertKeys([]v1.NamedCertificate{{Names: tc.names, CertInfo: v1.CertInfo{CertFile: "secret.crt", KeyFile: "secret.key"}}})
			if len(result) != 1 || result[0] != tc.expected {
				t.Errorf("expected: %v, actual: %v", []string{tc.expected}, result)
			}
		})
	}
}

func TestConfigToFlagsCurvePreferences(t *testing.T) {
	testCases := []struct {
		name               string
		curvePreferences   []int32
		apiServerArguments map[string]kubecontrolplanev1.Arguments
		wantArgs           []string
		wantParsed         []int32
	}{
		{
			name:     "nil uses default curves",
			wantArgs: []string{},
		},
		{
			name:             "empty uses default curves",
			curvePreferences: []int32{},
			wantArgs:         []string{},
		},
		{
			name:             "one curve",
			curvePreferences: []int32{23},
			wantArgs:         []string{"--tls-curve-preferences=23"},
			wantParsed:       []int32{23},
		},
		{
			name:             "multiple curves preserve order",
			curvePreferences: []int32{23, 24},
			wantArgs:         []string{"--tls-curve-preferences=23", "--tls-curve-preferences=24"},
			wantParsed:       []int32{23, 24},
		},
		{
			name:             "explicit argument overrides config",
			curvePreferences: []int32{23, 24},
			apiServerArguments: map[string]kubecontrolplanev1.Arguments{
				"tls-curve-preferences": {"29"},
			},
			wantArgs:   []string{"--tls-curve-preferences=29"},
			wantParsed: []int32{29},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &kubecontrolplanev1.KubeAPIServerConfig{
				GenericAPIServerConfig: v1.GenericAPIServerConfig{
					ServingInfo: v1.HTTPServingInfo{
						ServingInfo: v1.ServingInfo{
							BindAddress:      "127.0.0.1:6443",
							CipherSuites:     []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
							CurvePreferences: tc.curvePreferences,
							MinTLSVersion:    "VersionTLS12",
						},
					},
				},
				APIServerArguments: tc.apiServerArguments,
			}

			args, err := ConfigToFlags(config)
			if err != nil {
				t.Fatalf("ConfigToFlags() error = %v", err)
			}

			curveArgs := flagArguments(args, "tls-curve-preferences")
			if !reflect.DeepEqual(curveArgs, tc.wantArgs) {
				t.Errorf("ConfigToFlags() curve arguments = %v, want %v", curveArgs, tc.wantArgs)
			}
			if got, want := flagArguments(args, "tls-cipher-suites"), []string{"--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}; !reflect.DeepEqual(got, want) {
				t.Errorf("ConfigToFlags() cipher suite arguments = %v, want %v", got, want)
			}
			if got, want := flagArguments(args, "tls-min-version"), []string{"--tls-min-version=VersionTLS12"}; !reflect.DeepEqual(got, want) {
				t.Errorf("ConfigToFlags() minimum TLS version arguments = %v, want %v", got, want)
			}

			options := kubeapiserveroptions.NewSecureServingOptions()
			flagSet := pflag.NewFlagSet("secure serving", pflag.ContinueOnError)
			options.AddFlags(flagSet)
			if err := flagSet.Parse(curveArgs); err != nil {
				t.Fatalf("parsing curve arguments: %v", err)
			}
			if !reflect.DeepEqual(options.CurvePreferences, tc.wantParsed) {
				t.Errorf("parsed curve preferences = %v, want %v", options.CurvePreferences, tc.wantParsed)
			}
		})
	}
}

func flagArguments(args []string, name string) []string {
	prefix := "--" + name + "="
	values := make([]string, 0, len(args))
	for _, arg := range args {
		if len(arg) >= len(prefix) && arg[:len(prefix)] == prefix {
			values = append(values, arg)
		}
	}
	return values
}
