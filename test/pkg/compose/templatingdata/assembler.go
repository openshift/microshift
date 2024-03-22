package templatingdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/testutil"
	"k8s.io/klog/v2"
)

var (
	releaseInfoRpmRx   = regexp.MustCompile("^microshift-release-info-.*.rpm$")
	errNoRemoteRelease = fmt.Errorf("release from remote repo not found")
)

type TemplatingDataOpts struct {
	ArtifactsMainDir               string
	TemplatingDataFragmentFilepath string
}

func New(opts *TemplatingDataOpts) (*TemplatingData, error) {
	klog.InfoS("Constructing TemplatingData")

	rpmRepos := path.Join(opts.ArtifactsMainDir, "rpm-repos")
	localRepo := path.Join(rpmRepos, "microshift-local")
	fakeNextRepo := path.Join(rpmRepos, "microshift-fake-next-minor")
	baseRepo := path.Join(rpmRepos, "microshift-base")
	externalRepo := path.Join(rpmRepos, "microshift-external")

	if exists, err := util.PathExists(localRepo); err != nil {
		return nil, fmt.Errorf("failed to check if %s exists: %w", localRepo, err)
	} else if !exists {
		return nil, fmt.Errorf("%s does not exist - did you run build_rpms.sh and create_local_repos.sh?", localRepo)
	}

	var td *TemplatingData
	var err error

	if opts.TemplatingDataFragmentFilepath != "" {
		td, err = unmarshalTemplatingData(opts.TemplatingDataFragmentFilepath)
		if err != nil {
			return nil, err
		}
		klog.InfoS("TemplatingData fragment was provided and loaded", "intermediateTeplatingData", td)
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
			klog.ErrorS(err, "Error during WalkDir - ignoring")
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

	if releaseInfoFile == "" {
		return Release{}, fmt.Errorf("could not find microshift-release-info RPM in directory %q", repo)
	}
	klog.InfoS("Found microshift-release-info RPM for local repository", "repo", repo)

	rpmVersion, _, err := testutil.RunCommand("rpm", "-q", "--queryformat", "%{version}", releaseInfoFile)
	if err != nil {
		return Release{}, fmt.Errorf("failed to get version of the microshift-release-info (%q) RPM: %w", releaseInfoFile, err)
	}

	minorStr := strings.Split(rpmVersion, ".")[1]
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return Release{}, fmt.Errorf("failed to convert %q to int: %w", minorStr, err)
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
	klog.InfoS("Looking for a Release for minor version", "minor", minor)

	r, err := getReleaseFromRHOCP(minor)
	if err != nil && !errors.Is(err, errNoRemoteRelease) {
		return Release{}, err
	}
	if err == nil {
		klog.InfoS("Found release in RHOCP repository", "minor", minor, "release", r)
		return r, nil
	}

	r, err = getReleaseFromTheMirror(minor, false)
	if err != nil && !errors.Is(err, errNoRemoteRelease) {
		return Release{}, err
	}
	if err == nil {
		klog.InfoS("Found release in RC mirror", "minor", minor, "release", r)
		return r, nil
	}

	r, err = getReleaseFromTheMirror(minor, true)
	if err != nil && !errors.Is(err, errNoRemoteRelease) {
		return Release{}, err
	}
	if err == nil {
		klog.InfoS("Found release in EC mirror", "minor", minor, "release", r)
		return r, nil
	}

	klog.InfoS("No RPMs for the minor version found", "minor", minor)

	return Release{}, errNoRemoteRelease
}

// getReleaseFromTheMirror looks for MicroShift RPM in OpenShift mirror
func getReleaseFromTheMirror(minor int, devPreview bool) (Release, error) {
	dp := ""
	if devPreview {
		dp = "-dev-preview"
	}

	repo := fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/%s/microshift/ocp%s/latest-4.%d/el9/os/", getArch(), dp, minor)

	resp, err := http.Get(repo)
	if err != nil {
		return Release{}, fmt.Errorf("http.Get(%s) failed: %w", repo, err)
	}
	if resp.StatusCode != 200 { // TODO: Maybe this should compare to 404?
		return Release{}, errNoRemoteRelease
	}
	version, _, err := testutil.RunCommand(
		"sudo", "dnf", "repoquery", "microshift", "--quiet",
		"--queryformat", "%{version}-%{release}",
		"--disablerepo", "*",
		"--repofrompath", fmt.Sprintf("this,%s", repo),
	)
	if err != nil {
		return Release{}, fmt.Errorf("failed to repoquery %q for microshift RPM: %w", repo, err)
	}
	relInfo, err := downloadReleaseInfoRPM(version, "--repofrompath", fmt.Sprintf("this,%s", repo))
	if err != nil {
		return Release{}, err
	}
	imgs, err := getContainerImages(relInfo)
	if err != nil {
		return Release{}, err
	}

	return Release{
		Repository: repo,
		Version:    version,
		Minor:      minor,
		Images:     imgs,
	}, nil
}

// getReleaseFromRHOCP looks for MicroShift RPM in RHOCP
func getReleaseFromRHOCP(minor int) (Release, error) {
	rhocp := fmt.Sprintf("rhocp-4.%d-for-rhel-9-%s-rpms", minor, getArch())
	version, serr, err := testutil.RunCommand("sudo", "dnf", "repoquery", "microshift",
		"--quiet",
		"--queryformat", "%{version}-%{release}",
		"--repo", rhocp,
		"--latest-limit", "1",
	)
	if err == nil {
		relInfo, err := downloadReleaseInfoRPM(version, "--repo", rhocp)
		if err != nil {
			return Release{}, err
		}
		imgs, err := getContainerImages(relInfo)
		if err != nil {
			return Release{}, err
		}
		return Release{
			Repository: rhocp,
			Version:    version,
			Minor:      minor,
			Images:     imgs,
		}, nil
	}
	if strings.Contains(serr, "Cannot download repomd.xml: Cannot download repodata/repomd.xml: All mirrors were tried") {
		return Release{}, errNoRemoteRelease
	}
	return Release{}, err
}

func downloadReleaseInfoRPM(version string, repoOpts ...string) (string, error) {
	klog.InfoS("Downloading microshift-release-info", "version", version)

	destDir := "/tmp"
	cmd := append([]string{"sudo", "dnf", "download", fmt.Sprintf("microshift-release-info-%s", version),
		"--destdir", "/tmp"}, repoOpts...)
	_, _, err := testutil.RunCommand(cmd...)
	if err != nil {
		return "", fmt.Errorf("failed to download %q RPM: %w", fmt.Sprintf("microshift-release-info-%s", version), err)
	}

	// Because of `dnf download` superfluous output we cannot use the stdout.
	// Using --quiet would also cause RPM name to not be printed.
	path := filepath.Join(destDir, "microshift-release-info-"+version+".noarch.rpm")
	klog.InfoS("Downloaded microshift-release-info", "version", version, "destination", path)
	return path, nil
}

// isRHOCPAvailable checks if RHOCP of a given `minor` is available for usage by attempting
// to query the repository for cri-o package
func isRHOCPAvailable(minor int) bool {
	repo := fmt.Sprintf("rhocp-4.%d-for-rhel-9-%s-rpms", minor, "x86_64")
	_, _, err := testutil.RunCommand("sudo", "dnf", "repository-packages", repo, "info", "cri-o")
	return err == nil
}

// getContainerImages extracts list of images from release.json file inside given release-info RPM
func getContainerImages(releaseInfoFilePath string) ([]string, error) {
	sout, _, err := testutil.RunCommand(
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
