package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

func TestCertsCommandRunHooks(t *testing.T) {
	for _, hook := range []string{"PersistentPreRun", "PersistentPreRunE"} {
		for _, failure := range []string{"", "root", "privileges"} {
			if hook == "PersistentPreRun" && failure == "root" {
				continue
			}
			t.Run(hook+"/"+failure, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				var calls []string
				options := &certStatusOptions{
					IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
					now:       time.Now,
					loadConfig: func() (*config.Config, error) {
						calls = append(calls, "config")
						return &config.Config{}, nil
					},
					loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
						calls = append(calls, "inventory")
						return nil, nil
					},
				}
				root := &cobra.Command{Use: "microshift"}
				root.AddCommand(newCertsCommand(options, func() error {
					calls = append(calls, "privileges")
					if failure == "privileges" {
						return errors.New("privilege check failed")
					}
					return nil
				}))
				if hook == "PersistentPreRun" {
					root.PersistentPreRun = func(*cobra.Command, []string) { calls = append(calls, "root") }
				} else {
					root.PersistentPreRunE = func(*cobra.Command, []string) error {
						calls = append(calls, "root")
						if failure == "root" {
							return &certificateCommandError{certificatesv1alpha1.ErrorCodeInternalError, errors.New("root hook failed")}
						}
						return nil
					}
				}
				code := RunCertsCommand(root, []string{"certs", "status", "-o", "json"})
				if failure == "" {
					require.Zero(t, code)
					require.Equal(t, []string{"root", "privileges", "config", "inventory"}, calls)
					require.Empty(t, stderr.String())
					return
				}
				require.Equal(t, 1, code)
				require.Empty(t, stdout.String())
				var document certificatesv1alpha1.Error
				require.NoError(t, json.Unmarshal(stderr.Bytes(), &document))
				if failure == "root" {
					require.Equal(t, []string{"root"}, calls)
					require.Equal(t, certificatesv1alpha1.ErrorCodeInternalError, document.Code)
				} else {
					require.Equal(t, []string{"root", "privileges"}, calls)
					require.Equal(t, certificatesv1alpha1.ErrorCodeInsufficientPrivileges, document.Code)
				}
			})
		}
	}
}

func TestCertsCommandRestoresLogging(t *testing.T) {
	state := klog.CaptureState()
	t.Cleanup(state.Restore)
	var logs bytes.Buffer
	klog.ClearLogger()
	klog.LogToStderr(false)
	klog.SetOutput(&logs)
	root := &cobra.Command{Use: "microshift"}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.AddCommand(newCertsCommand(&certStatusOptions{}, func() error { return errors.New("privileges required") }))
	require.Equal(t, 1, RunCertsCommand(root, []string{"certs", "status", "-o", "json"}))
	require.Empty(t, logs.String())
	klog.Info("logging after certificate command")
	klog.Flush()
	require.Contains(t, logs.String(), "logging after certificate command")
}

func TestCertStatusErrors(t *testing.T) {
	const sensitiveConfig = "proxy: http://test-user:certificate-test-secret@proxy.invalid:8080\ninvalid: ["
	const configurationMessage = "failed to load MicroShift configuration; check /etc/microshift/config.yaml and /etc/microshift/config.d"
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
			{name: "flags after help", code: certificatesv1alpha1.ErrorCodeInvalidArguments, arguments: []string{"--help", "--invalid-status-flag"}},
			{name: "flags after short help", code: certificatesv1alpha1.ErrorCodeInvalidArguments, arguments: []string{"-h", "--invalid-status-flag"}},
		} {
			t.Run(format+"/"+tt.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				options := &certStatusOptions{
					IOStreams: genericclioptions.IOStreams{Out: &stdout, ErrOut: &stderr},
					now:       time.Now,
					loadConfig: func() (*config.Config, error) {
						require.NotEqual(t, "privileges", tt.name, "privilege failures must not read configuration")
						if tt.name == "configuration" {
							return nil, fmt.Errorf("failed to convert config yaml (%q) to json: invalid YAML", sensitiveConfig)
						}
						return &config.Config{Warnings: []string{"must not leak alongside an error"}}, nil
					},
					loadInventory: func(*config.Config) (certchains.CertificateInventory, error) {
						require.NotEqual(t, "configuration", tt.name, "configuration failures must not load the inventory")
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
				require.NotContains(t, stderr.String(), "certificate-test-secret")
				require.NotContains(t, stderr.String(), "proxy.invalid")
				require.NotContains(t, stderr.String(), "invalid YAML")
				if format == "" {
					require.True(t, strings.HasPrefix(stderr.String(), "Error: "))
					require.NotContains(t, stderr.String(), "Usage:")
					if tt.name == "configuration" {
						require.Equal(t, "Error: "+configurationMessage+"\n", stderr.String())
					}
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
				if tt.name == "configuration" {
					require.Equal(t, configurationMessage, document.Message)
				}
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
		{[]string{"certs", "status", "--help", "-o", "json"}, "json"},
		{[]string{"certs", "status", "-h", "--output=yaml"}, "yaml"},
		{[]string{"certs", "status", "-o", "json", "--help=invalid"}, "json"},
		{[]string{"certs", "status", "--help=invalid", "-o", "json"}, ""},
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
