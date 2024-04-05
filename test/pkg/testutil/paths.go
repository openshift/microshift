package testutil

import (
	"errors"
	"os"
	"path/filepath"
)

type Paths struct {
	MicroShiftRepoRootPath string
	TestDirPath            string
	ArtifactsMainDir       string
	BuildLogsDir           string
	BuildsDir              string
	OSTreeRepoDir          string
	VMStorageDir           string
	BootCImages            string
	RPMRepos               string
}

func NewPaths(microshiftTestDirAbs string) (*Paths, error) {
	microShiftRepoRootPath := filepath.Join(microshiftTestDirAbs, "..")
	artifactsMainDir := filepath.Join(microShiftRepoRootPath, "_output", "test-images")

	paths := &Paths{
		MicroShiftRepoRootPath: microShiftRepoRootPath,
		TestDirPath:            microshiftTestDirAbs,
		ArtifactsMainDir:       artifactsMainDir,
		BuildLogsDir:           filepath.Join(artifactsMainDir, "build-logs"),
		BuildsDir:              filepath.Join(artifactsMainDir, "builds"),
		OSTreeRepoDir:          filepath.Join(artifactsMainDir, "repo"),
		VMStorageDir:           filepath.Join(artifactsMainDir, "vm-storage"),
		BootCImages:            filepath.Join(artifactsMainDir, "bootc-images"),
		RPMRepos:               filepath.Join(artifactsMainDir, "rpm-repos"),
	}

	toCreate := []string{
		paths.ArtifactsMainDir,
		paths.BuildLogsDir,
		paths.BuildsDir,
		paths.OSTreeRepoDir,
		paths.VMStorageDir,
		paths.BootCImages,
		paths.RPMRepos,
	}
	errs := []error{}
	for _, p := range toCreate {
		if err := os.MkdirAll(p, 0755); err != nil {
			errs = append(errs, err)
		}
	}

	return paths, errors.Join(errs...)
}
