package version

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

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
	// cluster ID read from /usr/share/microshift/cluster-id file
	clusterIDFromDisk string
)

type Info struct {
	version.Info
	Patch     string `json:"patch"`
	ClusterID string `json:"clusterid"`
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
		Patch:     patchFromGit,
		ClusterID: clusterIDFromDisk,
	}
}

func init() {
	fileName := "/usr/share/microshift/cluster-id"
	var content []byte
	var err error

	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		content = []byte("unknown-cluster-id")
	} else {
		// Read the cluster ID from the disk
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			log.Fatal(err)
		}
	}
	clusterIDFromDisk = strings.TrimSpace(string(content))
}
