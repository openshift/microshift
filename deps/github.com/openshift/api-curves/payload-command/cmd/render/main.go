package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift/api/payload-command/render"
)

// this command injects the initial FeatureGate.status and places some CRDs to be created by the installer during bootstrapping
// remember that these manifests are not maintained in a running cluster.
func main() {
	o := &render.RenderOpts{
		ImageProvidedManifestDir: "/usr/share/bootkube/manifests/manifests",
	}
	o.AddFlags(flag.CommandLine)
	flag.Parse()

	if err := o.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}
