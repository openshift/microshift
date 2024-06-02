package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewPaths(t *testing.T) {
	createdDirs := []string{}

	mkdirAll := func(p string, _ os.FileMode) error {
		createdDirs = append(createdDirs, p)
		return nil
	}

	testDir := "/home/user/microshift/test"

	paths, err := newPaths(testDir, mkdirAll)
	assert.NoError(t, err)
	assert.NotNil(t, paths)

	assert.Equal(t, "/home/user/microshift", paths.MicroShiftRepoRootPath)
	assert.Equal(t, "/home/user/microshift/test", paths.TestDirPath)
	assert.Equal(t, "/home/user/microshift/test/image-blueprints", paths.ImageBlueprintsPath)
	assert.Equal(t, "/home/user/microshift/_output/test-images", paths.ArtifactsMainDir)
	assert.Equal(t, "/home/user/microshift/_output/test-images/build-logs", paths.BuildLogsDir)
	assert.Equal(t, "/home/user/microshift/_output/test-images/builds", paths.BuildsDir)
	assert.Equal(t, "/home/user/microshift/_output/test-images/repo", paths.OSTreeRepoDir)
	assert.Equal(t, "/home/user/microshift/_output/test-images/vm-storage", paths.VMStorageDir)
	assert.Equal(t, "/home/user/microshift/_output/test-images/bootc-images", paths.BootCImages)
	assert.Equal(t, "/home/user/microshift/_output/test-images/rpm-repos", paths.RPMRepos)

	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images")
	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images/build-logs")
	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images/builds")
	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images/repo")
	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images/vm-storage")
	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images/bootc-images")
	assert.Contains(t, createdDirs, "/home/user/microshift/_output/test-images/rpm-repos", paths.RPMRepos)
}
