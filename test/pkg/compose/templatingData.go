package compose

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

var (
	releaseInfoRpmRx   = regexp.MustCompile("^microshift-release-info-.*.rpm$")
	errNoRemoteRelease = fmt.Errorf("release from remote repo not found")
)

// Release represents metadata of particular set of RPMs.
type Release struct {
	// Repository is where RPM resides. It can be local (on the disk), http (like OpenShift's mirror), or RHOCP.
	Repository string

	// Version is full version string of a RPM, e.g. 4.14.16-202403071942.p0.g4cef5f2.assembly.4.14.16.el9.
	Version string

	// Minor is minor part of the version, e.g. 14.
	Minor int

	// Images is a list of images stored in release-info RPM.
	// Currently only for local repositories.
	// TODO: Extend to remote for local image mirroring.
	Images []string
}

// TemplatingData contains all values needed for templating Composer's Sources & Blueprints,
// and other templated artifacts within a MicroShift's test harness.
type TemplatingData struct {
	Arch string

	// Minor version of current release's RHOCP. If RHOCP is not available yet, it defaults to 0.
	RHOCPMinorY int

	// Minor version of previous release's RHOCP. If RHOCP is not available yet, it defaults to 0.
	RHOCPMinorY1 int

	// Minor version of previous previous release's RHOCP.
	RHOCPMinorY2 int

	// Current stores metadata of current release, i.e. matching currently checked out git branch.
	// If the RHOCP is not available yet, it can point to Release or Engineering candidates present on the OpenShift mirror.
	// If those are also not available, it will be empty and related composer Sources and Blueprints will not be build.
	Current Release

	// Current stores metadata of current release, i.e. matching currently checked out git branch.
	// If the RHOCP is not available yet, it can point to Release or Engineering candidates present on the OpenShift mirror.
	Previous Release

	// Current stores metadata of current release, i.e. matching currently checked out git branch.
	// Usually this should always point to RHOCP because it's two minor versions older than what we're working on currently.
	YMinus2 Release

	// Source stores metadata of RPMs built from currently checked out source code.
	Source Release

	// Source stores metadata of RPMs built from base branch of currently checked out branch.
	// Usually it can be `main` or `release-4.Y`.
	Base Release

	// Source stores metadata of RPMs built from currently checked out source code with minor version overridden
	// to be newer than what we're currently working on.
	// These are needed for various ostree upgrade tests.
	FakeNext Release

	// External stores metadata of RPMs supplied from external source, like private builds.
	External Release
}

func NewTemplatingData(opts *ComposeOpts) (*TemplatingData, error) {
	rpmRepos := path.Join(opts.ArtifactsMainDir, "rpm-repos")
	localRepo := path.Join(rpmRepos, "microshift-local")
	fakeNextRepo := path.Join(rpmRepos, "microshift-fake-next-minor")
	baseRepo := path.Join(rpmRepos, "microshift-base")
	externalRepo := path.Join(rpmRepos, "microshift-external")

	if exists, err := util.PathExists(localRepo); err != nil {
		return nil, fmt.Errorf("failed to check if %s exists: %w", localRepo, err)
	} else if !exists {
		return nil, fmt.Errorf("%s does not exist", localRepo)
	}

	var td *TemplatingData
	var err error

	if opts.TemplatingDataFragmentFilepath != "" {
		td, err = unmarshalTemplatingData(opts.TemplatingDataFragmentFilepath)
		if err != nil {
			return nil, err
		}
	} else {
		td = &TemplatingData{}
	}

	td.Arch = getArch()

	if td.Source.Repository == "" {
		td.Source, err = getReleaseFromLocalFs(localRepo)
		if err != nil {
			return nil, err
		}
	}

	if td.Base.Repository == "" {
		td.Base, err = getReleaseFromLocalFs(baseRepo)
		if err != nil {
			return nil, err
		}
	}

	if td.FakeNext.Repository == "" {
		td.FakeNext, err = getReleaseFromLocalFs(fakeNextRepo)
		if err != nil {
			return nil, err
		}
	}

	if td.External.Repository == "" {
		exists, err := util.PathExistsAndIsNotEmpty(externalRepo)
		if err != nil {
			return nil, err
		}
		if exists {
			td.External, err = getReleaseFromLocalFs(externalRepo)
			if err != nil {
				return nil, err
			}
		}
	}

	if td.Current.Repository == "" {
		td.Current, err = getReleaseFromRemoteRepo(td.Source.Minor)
		if err != nil {
			return nil, err
		}
	}

	if td.Previous.Repository == "" {
		td.Previous, err = getReleaseFromRemoteRepo(td.Source.Minor - 1)
		if err != nil {
			return nil, err
		}
	}

	if td.YMinus2.Repository == "" {
		td.YMinus2, err = getReleaseFromRemoteRepo(td.Source.Minor - 2)
		if err != nil {
			return nil, err
		}
	}

	// If templatingDataInputPath was provided, assume the 0 is
	// already "calculated" value and repo is not available yet.
	if td.RHOCPMinorY == 0 && opts.TemplatingDataFragmentFilepath == "" {
		if isRHOCPAvailable(td.Source.Minor) {
			td.RHOCPMinorY = td.Source.Minor
		}
	}

	if td.RHOCPMinorY1 == 0 {
		if isRHOCPAvailable(td.Previous.Minor) {
			td.RHOCPMinorY1 = td.Previous.Minor
		}
	}
	if td.RHOCPMinorY2 == 0 {
		if isRHOCPAvailable(td.YMinus2.Minor) {
			td.RHOCPMinorY2 = td.YMinus2.Minor
		}
	}

	klog.InfoS("Constructed TemplatingData", "results", td)

	return td, nil
}

func unmarshalTemplatingData(path string) (*TemplatingData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	td := &TemplatingData{}

	err = json.Unmarshal(data, td)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal partial templating data from %q: %w", path, err)
	}

	return td, nil
}

func getArch() string {
	if runtime.GOARCH == "amd64" {
		return "x86_64"
	}
	return "arm64"
}

// getReleaseFromLocalFs creates `Release` from local filesystem repository
func getReleaseFromLocalFs(repo string) (Release, error) {
	releaseInfoFile := ""
	err := filepath.WalkDir(repo, func(path string, d fs.DirEntry, err error) error {
		if releaseInfoFile != "" {
			return nil
		}

		if err != nil {
			fmt.Printf("err: %v\n", err)
			return nil
		}

		if releaseInfoRpmRx.MatchString(d.Name()) {
			releaseInfoFile = path
		}
		return nil
	})
	if err != nil {
		return Release{}, err
	}

	rpmVersion, _, err := runCommand("rpm", "-q", "--queryformat", "%{version}", releaseInfoFile)
	if err != nil {
		return Release{}, err
	}

	minorStr := strings.Split(rpmVersion, ".")[1]
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return Release{}, err
	}

	images, err := getContainerImages(releaseInfoFile)
	if err != nil {
		return Release{}, err
	}

	return Release{
		Repository: repo,
		Version:    rpmVersion,
		Minor:      minor,
		Images:     images,
	}, nil
}

// getReleaseFromRemoteRepo creates `Release` from remote repository.
// It looks for MicroShift RPM in following order:
// RHOCP, Release Candidates on OpenShift mirror, Engineering Candidates on OpenShift mirror.
func getReleaseFromRemoteRepo(minor int) (Release, error) {
	if r, err := getReleaseFromRHOCP(minor); err != nil && !errors.Is(err, errNoRemoteRelease) {
		return Release{}, err
	} else if err == nil {
		klog.InfoS("Found release in RHOCP repository", "minor", minor, "release", r)
		return r, nil
	}

	if r, err := getReleaseFromTheMirror(minor, false); err != nil && !errors.Is(err, errNoRemoteRelease) {
		return Release{}, err
	} else if err == nil {
		klog.InfoS("Found release in RC mirror", "minor", minor, "release", r)
		return r, nil
	}

	if r, err := getReleaseFromTheMirror(minor, true); err != nil && !errors.Is(err, errNoRemoteRelease) {
		return Release{}, err
	} else if err == nil {
		klog.InfoS("Found release in EC mirror", "minor", minor, "release", r)
		return r, nil
	}

	klog.InfoS("No RPMs for the minor found", "minor", minor)

	return Release{}, errNoRemoteRelease
}

// getReleaseFromTheMirror looks for MicroShift RPM in OpenShift mirror
func getReleaseFromTheMirror(minor int, devPreview bool) (Release, error) {
	dp := ""
	if devPreview {
		dp = "-dev-preview"
	}

	repo := fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/%s/microshift/ocp%s/latest-4.%d/el9/os", getArch(), dp, minor)

	resp, err := http.Get(repo)
	if err != nil {
		return Release{}, err
	}
	if resp.StatusCode != 200 {
		return Release{}, errNoRemoteRelease
	}
	sout, _, err := runCommand(
		"sudo", "dnf", "repoquery", "microshift", "--quiet",
		"--queryformat", "%{version}-%{release}",
		"--disablerepo", "*",
		"--repofrompath", fmt.Sprintf("this,%s", repo),
	)
	if err != nil {
		return Release{}, err
	}
	return Release{
		Repository: repo,
		Version:    sout,
		Minor:      minor,
	}, nil
}

// getReleaseFromRHOCP looks for MicroShift RPM in RHOCP
func getReleaseFromRHOCP(minor int) (Release, error) {
	rhocp := fmt.Sprintf("rhocp-4.%d-for-rhel-9-%s-rpms", minor, getArch())
	sout, _, err := runCommand("sudo", "dnf", "repoquery", "microshift",
		"--quiet",
		"--queryformat", "%{version}-%{release}",
		"--repo", rhocp,
		"--latest-limit", "1",
	)
	if err == nil {
		return Release{
			Repository: rhocp,
			Version:    sout,
			Minor:      minor,
		}, nil
	}
	return Release{}, errNoRemoteRelease
}

// isRHOCPAvailable checks if RHOCP of a given `minor` is available for usage by attempting
// to query the repository for cri-o package
func isRHOCPAvailable(minor int) bool {
	repo := fmt.Sprintf("rhocp-4.%d-for-rhel-9-%s-rpms", minor, "x86_64")
	_, _, err := runCommand("sudo", "dnf", "repository-packages", repo, "info", "cri-o")
	return err == nil
}

// getContainerImages extracts list of images from release.json file inside given release-info RPM
func getContainerImages(releaseInfoFilePath string) ([]string, error) {
	sout, _, err := runCommand(
		"bash", "-c",
		fmt.Sprintf("rpm2cpio %s | cpio  -i --to-stdout '*release-%s.json'", releaseInfoFilePath, getArch()),
	)
	if err != nil {
		return []string{}, fmt.Errorf("failed to obtain images from a %q: %w", releaseInfoFilePath, err)
	}

	data := struct {
		Images map[string]string
	}{}
	if err := json.Unmarshal([]byte(sout), &data); err != nil {
		return []string{}, fmt.Errorf("failed to unmarshal images from a %q: %w", releaseInfoFilePath, err)
	}

	images := []string{}
	for _, image := range data.Images {
		images = append(images, image)
	}

	return images, nil
}

// redactOutput overwrites sensitive data when logging command outputs
func redactOutput(output string) string {
	rx := regexp.MustCompile("gpgkeys.*")
	return rx.ReplaceAllString(output, "gpgkeys = REDACTED")
}

func runCommand(c ...string) (string, string, error) {
	cmd := exec.Command(c[0], c[1:]...)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	klog.V(2).InfoS("Running command", "cmd", cmd)
	err := cmd.Run()
	out := strings.Trim(outb.String(), "\n")
	serr := errb.String()
	klog.InfoS("Command complete", "cmd", cmd, "stdout", redactOutput(out), "stderr", redactOutput(serr), "err", err)
	if err != nil {
		return "", "", err
	}

	return out, serr, nil
}
