package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

func TestCertStatusErrors(t *testing.T) {
	for _, format := range []string{"", "json", "yaml"} {
		for _, tt := range []struct {
			name      string
			code      certificatesv1alpha1.ErrorCode
			arguments []string
		}{
			{name: "privileges", code: certificatesv1alpha1.ErrorCodeInsufficientPrivileges},
			{name: "configuration", code: certificatesv1alpha1.ErrorCodeInvalidConfiguration},
			{name: "inventory", code: certificatesv1alpha1.ErrorCodeCertificateInventoryFailed},
			{name: "certificate data", code: certificatesv1alpha1.ErrorCodeCertificateInventoryFailed},
			{name: "arguments", code: certificatesv1alpha1.ErrorCodeInvalidArguments, arguments: []string{"unexpected"}},
			{name: "flags", code: certificatesv1alpha1.ErrorCodeInvalidArguments, arguments: []string{"--invalid-status-flag"}},
		} {
			t.Run(format+"/"+tt.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				options := &certStatusOptions{
					IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
					now:       time.Now,
					loadConfig: func() (*config.Config, error) {
						require.NotEqual(t, "privileges", tt.name, "privilege failures must not read configuration")
						if tt.name == "configuration" {
							return nil, errors.New("invalid configuration")
						}
						return &config.Config{Warnings: []string{"must not leak alongside an error"}}, nil
					},
					loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
						if tt.name == "inventory" {
							return nil, errors.New("certificate file is unreadable")
						}
						if tt.name == "certificate data" {
							return certchains.CertificateInventory{{Name: "invalid-certificate"}}, nil
						}
						t.Fatal("argument and privilege failures must not load the inventory")
						return nil, nil
					},
				}
				root := &cobra.Command{Use: "microshift"}
				root.SetOut(&stdout)
				root.SetErr(&stderr)
				root.AddCommand(newCertsCommand(options, func() error {
					if tt.name == "privileges" {
						return errors.New("command requires root privileges")
					}
					return nil
				}))
				args := append([]string{"certs", "status"}, tt.arguments...)
				if format != "" {
					// Put output after the invalid flag to exercise early parse failures.
					args = append(args, "-o", format)
				}
				require.Equal(t, 1, RunCertsCommand(root, args))
				require.Empty(t, stdout.String())
				if format == "" {
					require.True(t, strings.HasPrefix(stderr.String(), "Error: "))
					require.NotContains(t, stderr.String(), "Usage:")
					return
				}
				data := stderr.Bytes()
				if format == "yaml" {
					var err error
					data, err = yaml.YAMLToJSONStrict(data)
					require.NoError(t, err)
				}
				decoder := json.NewDecoder(bytes.NewReader(data))
				var document certificatesv1alpha1.Error
				require.NoError(t, decoder.Decode(&document))
				require.ErrorIs(t, decoder.Decode(new(any)), io.EOF, "stderr must contain exactly one document")
				require.Equal(t, tt.code, document.Code)
				require.NotEmpty(t, document.Message)
				require.False(t, document.GeneratedAt.IsZero())
				require.Nil(t, document.Details)
				var fields map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(data, &fields))
				require.JSONEq(t, "null", string(fields["details"]))
				decoded, gvk, err := certificateCodecs.UniversalDeserializer().Decode(data, nil, nil)
				require.NoError(t, err)
				require.IsType(t, &certificatesv1alpha1.Error{}, decoded)
				require.Equal(t, certificatesv1alpha1.GroupVersion.WithKind(certificatesv1alpha1.ErrorKind), *gvk)
			})
		}
	}
}

func TestCertificateOutputFormat(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"certs", "status", "-o", "json"}, "json"},
		{[]string{"certs", "status", "--output=yaml"}, "yaml"},
		{[]string{"certs", "status", "--invalid", "-ojson"}, "json"},
		{[]string{"certs", "status", "-o=yaml", "--invalid"}, "yaml"},
		{[]string{"certs", "status", "--", "-o", "json"}, ""},
		{[]string{"certs", "status", "-o", "json", "--output=xml"}, "xml"},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			require.Equal(t, tt.want, certificateOutputFormat(tt.args))
		})
	}
}

func TestCertificateObjectEncoding(t *testing.T) {
	var out bytes.Buffer
	failure := &certificatesv1alpha1.Error{}
	for _, format := range []string{"json", "yaml"} {
		// An unregistered GVK must not accidentally be emitted as a valid document.
		failure.APIVersion = "unsupported/v1"
		failure.Kind = "Unknown"
		// Versioned encoders use the registered Go type to supply the correct GVK.
		require.NoError(t, writeCertificateObject(&out, failure, format))
		decoded, gvk, err := certificateCodecs.UniversalDeserializer().Decode(out.Bytes(), nil, nil)
		require.NoError(t, err)
		require.IsType(t, failure, decoded)
		require.Equal(t, certificatesv1alpha1.GroupVersion.WithKind(certificatesv1alpha1.ErrorKind), *gvk)
		out.Reset()
		failure.Details = &runtime.RawExtension{Raw: []byte("invalid JSON")}
		require.Error(t, writeCertificateObject(&out, failure, format))
		require.Empty(t, out.String(), "serialization failures must not write a partial document")
		failure.Details = nil
	}
	require.Error(t, writeCertificateObject(&out, failure, "xml"))
	require.Empty(t, out.String())
}
