package main

import (
	"flag"

	"k8s.io/klog/examples/util/require"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	// By default klog writes to stderr. Setting logtostderr to false makes klog
	// write to a log file.
	require.NoError(flag.Set("logtostderr", "false"))
	require.NoError(flag.Set("log_file", "myfile.log"))
	flag.Parse()
	klog.Info("nice to meet you")
	klog.Flush()
}
