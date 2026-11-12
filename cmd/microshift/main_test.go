//nolint:testpackage // Exercise the unexported executable dispatch boundary.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
)

func TestRunCommandCertificateErrors(t *testing.T) {
	for _, format := range []string{"", "json", "yaml"} {
		for _, arguments := range [][]string{
			{"certs", "status", "unexpected"},
			{"certs", "status", "--invalid-status-flag"},
			{"certs", "status", "--help", "--invalid-status-flag"},
			{"certs", "status", "-h", "--invalid-status-flag"},
			{"certs", "renew", "unexpected", "--serving"},
			{"certs", "renew", "--ca", "--invalid-renew-flag"},
			{"certs", "renew", "--help", "--invalid-renew-flag"},
			{"certs", "renew", "-h", "--invalid-renew-flag"},
		} {
			t.Run(format+"/"+strings.Join(arguments, " "), func(t *testing.T) {
				args := slices.Clone(arguments)
				if format != "" {
					args = append(args, "-o", format)
				}
				stdout, stderr, code := executeRunCommand(t, args...)
				require.Equal(t, 1, code)
				require.Empty(t, stdout)
				require.NotContains(t, stderr, "Usage:")
				if format == "" {
					require.True(t, strings.HasPrefix(stderr, "Error: "))
					return
				}
				data := []byte(stderr)
				if format == "yaml" {
					var err error
					data, err = yaml.YAMLToJSONStrict(data)
					require.NoError(t, err)
				}
				decoder := json.NewDecoder(bytes.NewReader(data))
				var document certificatesv1alpha1.Error
				require.NoError(t, decoder.Decode(&document))
				require.ErrorIs(t, decoder.Decode(new(any)), io.EOF)
				require.Equal(t, certificatesv1alpha1.APIVersion, document.APIVersion)
				require.Equal(t, certificatesv1alpha1.ErrorKind, document.Kind)
				require.Equal(t, certificatesv1alpha1.ErrorCodeInvalidArguments, document.Code)
				require.NotEmpty(t, document.Message)
				require.False(t, document.GeneratedAt.IsZero())
				require.Nil(t, document.Details)
			})
		}
	}
}

func TestRunCommandOtherCommands(t *testing.T) {
	stdout, stderr, code := executeRunCommand(t, "certs", "status", "--help")
	require.Zero(t, code)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "Report the status of managed MicroShift certificates")
	stdout, stderr, code = executeRunCommand(t, "certs", "renew", "--help")
	require.Zero(t, code)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "--dry-run")
	require.Contains(t, stdout, "--serving")
	require.Contains(t, stdout, "--ca")

	stdout, stderr, code = executeRunCommand(t, "version", "--invalid-version-flag", "-o", "json")
	require.Equal(t, 1, code)
	require.Empty(t, stdout)
	require.Contains(t, stderr, "unknown flag: --invalid-version-flag")
	require.NotContains(t, stderr, certificatesv1alpha1.APIVersion)
}

func TestRunCommandCertificateHooks(t *testing.T) {
	for _, hook := range []string{"injected", "root", "root-error"} {
		for _, format := range []string{"json", "yaml"} {
			t.Run(hook+"/"+format, func(t *testing.T) {
				stdout, stderr, code := executeRunCommandWithEnv(t,
					[]string{"MICROSHIFT_TEST_CERTIFICATE_HOOK=" + hook}, "certs", "status", "-o", format)
				if hook != "root-error" {
					require.Zero(t, code)
					require.Equal(t, "command ran\n", stdout)
					require.Empty(t, stderr, "root initialization must not add logs to structured output")
					return
				}
				require.Equal(t, 1, code)
				require.Empty(t, stdout)
				data := []byte(stderr)
				if format == "yaml" {
					var err error
					data, err = yaml.YAMLToJSONStrict(data)
					require.NoError(t, err)
				}
				decoder := json.NewDecoder(bytes.NewReader(data))
				var document certificatesv1alpha1.Error
				require.NoError(t, decoder.Decode(&document))
				require.ErrorIs(t, decoder.Decode(new(any)), io.EOF)
				require.Equal(t, certificatesv1alpha1.ErrorKind, document.Kind)
				require.Equal(t, "root hook failed", document.Message)
			})
		}
	}
}

func executeRunCommand(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	return executeRunCommandWithEnv(t, nil, args...)
}

func executeRunCommandWithEnv(t *testing.T, env []string, args ...string) (string, string, int) {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, append([]string{"-test.run=^TestRunCommandHelperProcess$", "--"}, args...)...)
	command.Env = append(os.Environ(), "MICROSHIFT_TEST_RUN_COMMAND=1")
	command.Env = append(command.Env, env...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	require.NoError(t, ctx.Err())
	if err != nil {
		var exitError *exec.ExitError
		require.ErrorAs(t, err, &exitError)
	}
	return stdout.String(), stderr.String(), command.ProcessState.ExitCode()
}

func TestRunCommandHelperProcess(t *testing.T) {
	if os.Getenv("MICROSHIFT_TEST_RUN_COMMAND") != "1" {
		return
	}
	separator := slices.Index(os.Args, "--")
	require.NotEqual(t, -1, separator)
	command := newCommand()
	if hook := os.Getenv("MICROSHIFT_TEST_CERTIFICATE_HOOK"); hook != "" {
		configureCertificateHookTest(t, command, hook)
	}
	os.Exit(runCommand(command, os.Args[separator+1:]))
}

func configureCertificateHookTest(t *testing.T, root *cobra.Command, hook string) {
	t.Helper()
	log.SetFlags(log.LstdFlags)
	rootCalls := 0
	if hook != "injected" {
		root.PersistentPreRunE = func(*cobra.Command, []string) error {
			rootCalls++
			klog.Info("root initialization diagnostic")
			if hook == "root-error" {
				return errors.New("root hook failed")
			}
			return nil
		}
	}
	status, _, err := root.Find([]string{"certs", "status"})
	require.NoError(t, err)
	require.NotNil(t, status.PreRunE, "privilege checks belong to the executable subcommand")
	// Stub privilege and inventory work so this subprocess test is independent
	// of the invoking user's permissions and host configuration.
	status.PreRunE = func(*cobra.Command, []string) error {
		require.NotEqual(t, "root-error", hook, "root failure must stop execution")
		require.Zero(t, log.Flags(), "the injected root hook must initialize logging first")
		if hook == "root" {
			require.Equal(t, 1, rootCalls)
		}
		return nil
	}
	status.RunE = func(command *cobra.Command, _ []string) error {
		command.Println("command ran")
		return nil
	}
}
