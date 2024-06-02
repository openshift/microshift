package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	MicroShiftRepoRootPath string
	TestDirPath            string
	ImageBlueprintsPath    string
	ArtifactsMainDir       string
	BuildLogsDir           string
	BuildsDir              string
	OSTreeRepoDir          string
	VMStorageDir           string
	BootCImages            string
	RPMRepos               string
}

func NewPaths() (*Paths, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working dir: %w", err)
	}
	return newPaths(wd, os.MkdirAll)
}

func newPaths(testDirPath string, mkdirAll func(string, os.FileMode) error) (*Paths, error) {
	microShiftRepoRootPath := filepath.Join(testDirPath, "..")
	artifactsMainDir := filepath.Join(microShiftRepoRootPath, "_output", "test-images")

	paths := &Paths{
		MicroShiftRepoRootPath: microShiftRepoRootPath,
		TestDirPath:            testDirPath,
		ImageBlueprintsPath:    filepath.Join(testDirPath, "image-blueprints"),
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
		if err := mkdirAll(p, 0755); err != nil {
			errs = append(errs, err)
		}
	}

	return paths, errors.Join(errs...)
}
