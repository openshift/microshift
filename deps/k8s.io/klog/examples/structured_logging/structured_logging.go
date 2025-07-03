package main

import (
	"flag"

	"k8s.io/klog/v2"
)

// MyStruct will be logged via %+v
type MyStruct struct {
	Name     string
	Data     string
	internal int
}

// MyStringer will be logged as string, with String providing that string.
type MyString MyStruct

func (m MyString) String() string {
	return m.Name + ": " + m.Data
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	someData := MyStruct{
		Name:     "hello",
		Data:     "world",
		internal: 42,
	}

	longData := MyStruct{
		Name: "long",
		Data: `Multiple
lines
with quite a bit
of text.`,
	}

	logData := MyStruct{
		Name: "log output from some program",
		Data: `I0000 12:00:00.000000  123456 main.go:42] Starting
E0000 12:00:01.000000  123456 main.go:43] Failed for some reason
`,
	}

	stringData := MyString(longData)

	klog.Infof("someData printed using InfoF: %v", someData)
	klog.Infof("longData printed using InfoF: %v", longData)
	klog.Infof(`stringData printed using InfoF,
with the message across multiple lines:
%v`, stringData)
	klog.Infof("logData printed using InfoF:\n%v", logData)

	klog.Info("=============================================")

	klog.InfoS("using InfoS", "someData", someData)
	klog.InfoS("using InfoS", "longData", longData)
	klog.InfoS(`using InfoS with
the message across multiple lines`,
		"int", 1,
		"stringData", stringData,
		"str", "another value")
	klog.InfoS("using InfoS", "logData", logData)
	klog.InfoS("using InfoS", "boolean", true, "int", 1, "float", 0.1)

	// The Kubernetes recommendation is to start the message with uppercase
	// and not end with punctuation. See
	// https://github.com/kubernetes/community/blob/HEAD/contributors/devel/sig-instrumentation/migration-to-structured-logging.md
	klog.InfoS("Did something", "item", "foobar")
	// Not recommended, but also works.
	klog.InfoS("This is a full sentence.", "item", "foobar")
}
