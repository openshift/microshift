package version

import (
	"fmt"
	"runtime"

	// "k8s.io/component-base/metrics"
	// "k8s.io/component-base/metrics/legacyregistry"

	"k8s.io/apimachinery/pkg/version"
)

var (
	// commitFromGit is a constant representing the source version that
	// generated this build. It should be set during build via -ldflags.
	commitFromGit string
	// versionFromGit is a constant representing the version tag that
	// generated this build. It should be set during build via -ldflags.
	versionFromGit = "unknown"
	// major version
	majorFromGit string
	// minor version
	minorFromGit string
	// patch version
	patchFromGit string
	// build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate string
	// state of git tree, either "clean" or "dirty"
	gitTreeState string
	// releaseStream specifies whether microshift is compiled from OCP or OKD
	variant string
)

const VariantCommunity = "community"

type Info struct {
	version.Info
	Variant string
	Patch   string `json:"patch"`
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() Info {
	return Info{
		Info: version.Info{
			Major:        majorFromGit,
			Minor:        minorFromGit,
			GitCommit:    commitFromGit,
			GitVersion:   versionFromGit,
			GitTreeState: gitTreeState,
			BuildDate:    buildDate,
			GoVersion:    runtime.Version(),
			Compiler:     runtime.Compiler,
			Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Patch:   patchFromGit,
		Variant: variant,
	}
}

func init() {
	// buildInfo := metrics.NewGaugeVec(
	// 	&metrics.GaugeOpts{
	// 		Name: "openshift_build_info",
	// 		Help: "A metric with a constant '1' value labeled by major, minor, git commit & git version from which OpenShift was built.",
	// 	},
	// 	[]string{"major", "minor", "gitCommit", "gitVersion"},
	// )
	// buildInfo.WithLabelValues(majorFromGit, minorFromGit, commitFromGit, versionFromGit).Set(1)

	// // we're ok with an error here for now because test-integration illegally runs the same process
	// legacyregistry.Register(buildInfo)
}
