package config

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func HideUnsupportedFlags(flags *pflag.FlagSet) {
	// hide logging flags that we do not use/support
	loggingFlags := pflag.NewFlagSet("logging-flags", pflag.ContinueOnError)
	logs.AddFlags(loggingFlags)

	supportedLoggingFlags := sets.NewString("v")

	loggingFlags.VisitAll(func(pf *pflag.Flag) {
		if !supportedLoggingFlags.Has(pf.Name) {
			if err := flags.MarkHidden(pf.Name); err != nil {
				klog.Errorf("failed to hide flag %q: %v", pf.Name, err)
			}
		}
	})
	if err := flags.MarkHidden("version"); err != nil {
		klog.Errorf("failed to hide flag %q: %v", "version", err)
	}
}
