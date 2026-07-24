// Example demonstrating the new stderr threshold behavior
package main

import (
	"flag"

	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Info("This is an INFO message")
	klog.Warning("This is a WARNING message")
	klog.Error("This is an ERROR message")

	klog.Flush()
}

// Run examples:
//
// 1. Legacy behavior (default) - all logs to stderr:
//    go run main.go -logtostderr=true -stderrthreshold=ERROR
//    Result: All three messages appear
//
// 2. New behavior - filter by severity:
//    go run main.go -logtostderr=true -legacy_stderr_threshold_behavior=false -stderrthreshold=ERROR
//    Result: Only ERROR message appears
//
// 3. New behavior - show WARNING and above:
//    go run main.go -logtostderr=true -legacy_stderr_threshold_behavior=false -stderrthreshold=WARNING
//    Result: WARNING and ERROR messages appear
//
// 4. Using alsologtostderrthreshold with file logging:
//    go run main.go -logtostderr=false -alsologtostderr=true -alsologtostderrthreshold=ERROR -log_dir=/tmp/logs
//    Result: All logs in files, only ERROR to stderr
