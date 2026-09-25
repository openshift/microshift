package cmd

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/cli"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
)

var certificateCodecs = func() serializer.CodecFactory {
	scheme := runtime.NewScheme()
	utilruntime.Must(certificatesv1alpha1.AddToScheme(scheme))
	return serializer.NewCodecFactory(scheme)
}()

type certificateCommandError struct {
	code certificatesv1alpha1.ErrorCode
	err  error
}

func (e *certificateCommandError) Error() string { return e.err.Error() }
func (e *certificateCommandError) Unwrap() error { return e.err }

// RunCertsCommand executes a certificate command through the root command.
// Unlike cli.Run, it formats errors itself so diagnostics cannot be appended
// to a machine-readable Error document, including failures before RunE.
func RunCertsCommand(root *cobra.Command, args []string) int {
	root.SetArgs(args)
	root.SilenceUsage = true
	command, _, _ := root.Find(args)
	if command == nil {
		command = root
	}
	if err := cli.RunNoErrOutput(root); err != nil {

		// If no output desired, just print stdout
		out := command.ErrOrStderr()
		format := certificateOutputFormat(args)
		if format != certificateOutputJSON && format != certificateOutputYAML {
			_, _ = fmt.Fprintf(out, "Error: %v\n", err)
			return 1
		}

		code := certificatesv1alpha1.ErrorCodeInvalidArguments
		var failure *certificateCommandError
		if errors.As(err, &failure) {
			code = failure.code
		}

		document := &certificatesv1alpha1.Error{
			GeneratedAt: metav1.NewTime(time.Now().UTC()),
			Code:        code,
			Message:     err.Error(),
		}
		// If stderr itself cannot be written, there is no other diagnostic
		// channel that preserves the output contract. Still exit non-zero.
		_ = writeCertificateObject(out, document, format)
		return 1
	}
	return 0
}

func certificateOutputFormat(args []string) string {
	// Parse only the output selector, independently of the command parser:
	// an invalid flag may occur before -o and stop normal flag parsing early.
	flags := pflag.NewFlagSet("certificate output", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.ParseErrorsAllowlist.UnknownFlags = true
	output := flags.StringP("output", "o", "", "")
	_ = flags.Parse(args)
	return *output
}

func writeCertificateObject(out io.Writer, object runtime.Object, format string) error {
	info, ok := runtime.SerializerInfoForMediaType(certificateCodecs.SupportedMediaTypes(), "application/"+format)
	if !ok {
		return fmt.Errorf("unsupported certificate output format %q", format)
	}
	encoder := info.Serializer
	if info.PrettySerializer != nil {
		encoder = info.PrettySerializer
	}
	// Encode fully before writing so serialization failures leave stdout empty.
	contents, err := runtime.Encode(certificateCodecs.EncoderForVersion(encoder, certificatesv1alpha1.GroupVersion), object)
	if err != nil {
		return err
	}
	_, err = out.Write(contents)
	return err
}
