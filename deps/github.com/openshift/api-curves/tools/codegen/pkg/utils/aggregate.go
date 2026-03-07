package utils

import (
	"bytes"
	"errors"
	"strings"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// AggregatePrinter is an error wrapper that prints out aggregate and wrapped
// errors as individual errors in individual lines.
type AggregatePrinter struct {
	inner error
}

// NewAggregatePrinter creates a new Aggregate error.
func NewAggregatePrinter(err error) AggregatePrinter {
	return AggregatePrinter{
		inner: err,
	}
}

// Error prints the error out as a string.
// It unwraps wrapped errors and iterates over aggregate errors
// to print errors on invidividual lines.
func (a AggregatePrinter) Error() string {
	buf := bytes.NewBuffer([]byte("\n"))

	handleError(buf, a.inner, 0)

	return buf.String()
}

// handleError checks if the error given is an aggregate and iterates
// over its sub-errors if it is, else it checks if it is a wrapped
// error, in which case it prints out the context and unwraps the error.
// At 100 levels of indent the function stops to avoid infinite loops.
func handleError(buf *bytes.Buffer, err error, indent int) {
	if agg, ok := err.(kerrors.Aggregate); ok && indent < 100 {
		for _, aggErr := range agg.Errors() {
			handleError(buf, aggErr, indent)
		}
	} else if u := errors.Unwrap(err); u != nil && indent < 100 {
		errorContext := strings.Split(err.Error(), u.Error())[0]
		printError(buf, errors.New(errorContext), indent)
		handleError(buf, u, indent+1)
	} else {
		printError(buf, err, indent)
	}
}

// printError prints the error after indenting it by the given amount.
func printError(buf *bytes.Buffer, err error, indent int) {
	buf.WriteString(colourRed)

	for i := 0; i < indent; i++ {
		buf.WriteString("\t")
	}

	buf.WriteString(err.Error())
	buf.WriteString("\n" + colourReset)
}
