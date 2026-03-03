package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

func PrintResults(resultsByGroup map[string][]generation.Result) error {
	errs := []error{}

	for _, results := range resultsByGroup {
		resultsByVersion := map[string][]generation.Result{}
		for _, result := range results {
			resultsByVersion[result.Version] = append(resultsByVersion[result.Version], result)
		}

		for _, results := range resultsByVersion {
			resultsByManifest := map[string][]generation.Result{}
			for _, result := range results {
				resultsByManifest[result.Manifest] = append(resultsByManifest[result.Manifest], result)
			}

			for _, results := range resultsByManifest {
				if err := printResults(results); err != nil {
					return err
				}
			}
		}
	}

	return kerrors.NewAggregate(errs)
}

func printResults(results []generation.Result) error {
	for _, result := range results {
		if len(result.Errors) == 0 && len(result.Warnings) == 0 && len(result.Info) == 0 {
			continue
		}

		buf := bytes.NewBuffer(nil)

		buf.WriteString("----- " + colourCyan + result.Group + colourReset + " ----- " + colourCyan + result.Version + colourReset + " -----")
		if result.Manifest != "" {
			buf.WriteString(" " + colourCyan + result.Manifest + colourReset + " -----")
		}
		buf.WriteString("\n")

		buf.WriteString("----- " + colourMagenta + strings.ToUpper(result.Generator) + colourReset + " -----\n")

		for _, err := range result.Errors {
			indentString(buf, err.Error(), 0, "ERROR: ", colourRed)
		}

		for _, warn := range result.Warnings {
			indentString(buf, warn, 0, "WARN: ", colourYellow)
		}

		for _, info := range result.Info {
			indentString(buf, info, 0, "INFO: ", colourGreen)
		}

		if _, err := fmt.Fprint(os.Stdout, buf.String()); err != nil {
			return err
		}

	}
	return nil
}

// indentString prints the string after indenting it by the given amount.
func indentString(buf *bytes.Buffer, msg string, indent int, prefix, colour string) {
	for i := 0; i < indent; i++ {
		buf.WriteString("\t")
	}

	buf.WriteString(colour + prefix + colourReset)
	buf.WriteString(msg + "\n")
}
