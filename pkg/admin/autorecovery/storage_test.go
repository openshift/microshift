package autorecovery

import (
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Backups(t *testing.T) {
	previouslyRestoredBackup := Backup{CreationTime: time.Date(2024, 10, 1, 1, 0, 0, 0, time.UTC), Version: "4.18.1"}
	oldVersion := Backup{CreationTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC), Version: "4.18.0"}
	mostRecent := Backup{CreationTime: time.Date(2024, 10, 1, 3, 0, 0, 0, time.UTC), Version: "4.18.1"}
	initialBackups := Backups{
		oldVersion,
		previouslyRestoredBackup,
		{
			CreationTime: time.Date(2024, 10, 1, 2, 0, 0, 0, time.UTC),
			Version:      "4.18.1",
		},
		mostRecent,
	}

	backupsWithoutPrevious := initialBackups.RemoveBackup(previouslyRestoredBackup.Name())
	assert.Len(t, backupsWithoutPrevious, 3)
	assert.NotContains(t, backupsWithoutPrevious, previouslyRestoredBackup)

	backupsFilteredByVersion := backupsWithoutPrevious.FilterByVersion("4.18.1")
	assert.Len(t, backupsFilteredByVersion, 2)
	assert.NotContains(t, backupsFilteredByVersion, oldVersion)

	candidate := backupsFilteredByVersion.GetMostRecent()
	assert.Equal(t, mostRecent, candidate)
}

func Test_dirListToBackups(t *testing.T) {
	expectedBackups := Backups{
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 11, 00, 0, time.UTC), Version: "4.18.0"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 22, 00, 0, time.UTC), Version: "4.18.0"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 33, 00, 0, time.UTC), Version: "4.18.1"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 44, 00, 0, time.UTC), Version: "default-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 55, 00, 0, time.UTC), Version: "rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0"},
	}

	files := []string{
		"20241001111100_4.18.0",
		"20241001112200_4.18.0",
		"20241001113300_4.18.1",
		"202410011139004.18.1", // Missing delimiter `_`` - should be ignored
		"1239832asd_4.18.1",    // Invalid datetime - should be ignored
		"20241001114400_default-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0",
		"20241001115500_rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0",
	}

	backups := dirListToBackups(files)
	assert.Len(t, backups, 5)
	assert.Equal(t, expectedBackups, backups)
}

func Test_getBackups(t *testing.T) {
	expectedBackups := Backups{
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 11, 00, 0, time.UTC), Version: "4.18.0"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 22, 00, 0, time.UTC), Version: "4.18.0"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 33, 00, 0, time.UTC), Version: "4.18.1"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 44, 00, 0, time.UTC), Version: "default-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0"},
		Backup{CreationTime: time.Date(2024, 10, 01, 11, 55, 00, 0, time.UTC), Version: "rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0"},
	}

	storageFS := fstest.MapFS{
		"20241001111100_4.18.0": &fstest.MapFile{Mode: fs.ModeDir},
		"20241001112200_4.18.0": &fstest.MapFile{Mode: fs.ModeDir},
		"20241001113300_4.18.1": &fstest.MapFile{Mode: fs.ModeDir},
		"20241001114400_default-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0": &fstest.MapFile{Mode: fs.ModeDir},
		"20241001115500_rhel-35d7b5c80f0f1378d6846f6dc1304bbf1dcdc5847198fcd4e6099364eaf99048.0":    &fstest.MapFile{Mode: fs.ModeDir},
	}

	bs, err := getBackups(storageFS)
	assert.NoError(t, err)
	assert.Equal(t, expectedBackups, bs)
}
