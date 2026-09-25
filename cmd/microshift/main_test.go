//nolint:testpackage // Exercise the unexported executable dispatch boundary.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

	stdout, stderr, code = executeRunCommand(t, "version", "--invalid-version-flag", "-o", "json")
	require.Equal(t, 1, code)
	require.Empty(t, stdout)
	require.Contains(t, stderr, "unknown flag: --invalid-version-flag")
	require.NotContains(t, stderr, certificatesv1alpha1.APIVersion)
}

func executeRunCommand(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, append([]string{"-test.run=^TestRunCommandHelperProcess$", "--"}, args...)...)
	command.Env = append(os.Environ(), "MICROSHIFT_TEST_RUN_COMMAND=1")
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
	os.Exit(runCommand(newCommand(), os.Args[separator+1:]))
}
