package main

import (
	"bytes"
	"flag"
	"fmt"

	"k8s.io/klog/examples/util/require"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	require.NoError(flag.Set("logtostderr", "false"))
	require.NoError(flag.Set("alsologtostderr", "false"))
	flag.Parse()

	buf := new(bytes.Buffer)
	klog.SetOutput(buf)
	klog.Info("nice to meet you")
	klog.Flush()

	fmt.Printf("LOGGED: %s", buf.String())
}
