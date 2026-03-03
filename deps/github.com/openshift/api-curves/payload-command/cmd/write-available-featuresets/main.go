package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift/api/payload-command/render"
)

// this command writes manifests for each available featureset so that the cluster-config-operator can read them
// in order to maintain the list.
func main() {
	o := &render.WriteFeatureSets{}
	o.AddFlags(flag.CommandLine)
	flag.Parse()

	if err := o.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}
