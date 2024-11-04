package autorecovery

import (
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"k8s.io/klog/v2"
)

type Backup struct {
	CreationTime time.Time

	// Version is either Deployment ID (bootc/ostree systems) or version of the MicroShift binary (RPM systems).
	// Used for selecting version-wise compatible backups for restoring.
	Version string
}

func (b Backup) Name() data.BackupName {
	t := b.CreationTime.Format(backupCreationTimeFormat)
	return data.BackupName(fmt.Sprintf("%s_%s", t, b.Version))
}

type Backups []Backup

func (bs Backups) RemoveBackup(bb data.BackupName) Backups {
	match := false
	filtered := slices.DeleteFunc(bs, func(b Backup) bool {
		if b.Name() == bb {
			match = true
			return true
		}
		return false
	})
	if match {
		klog.InfoS("Filtered list of backups - removed previously restored backup", "removed", bb, "newList", filtered)
	}
	return filtered
}

func (bs Backups) FilterByVersion(ver string) Backups {
	filtered := slices.DeleteFunc(bs, func(b Backup) bool {
		return b.Version != ver
	})
	klog.InfoS("Filtered list of backups by version", "version", ver, "newList", filtered)
	return filtered
}

func (bs Backups) GetMostRecent() Backup {
	slices.SortFunc(bs, func(a, b Backup) int {
		return b.CreationTime.Compare(a.CreationTime)
	})
	return bs[0]
}

func GetBackups(storagePath data.StoragePath) (Backups, error) {
	fsys := os.DirFS(string(storagePath))
	return getBackups(fsys)
}

func getBackups(fsys fs.FS) (Backups, error) {
	dirEntries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			dirs = append(dirs, dirEntry.Name())
		}
	}
	bs := dirListToBackups(dirs)
	klog.InfoS("Auto-recovery backup storage read and parsed", "dirs", dirs, "backups", bs)

	return bs, nil
}

func dirListToBackups(files []string) Backups {
	backups := make([]Backup, 0, len(files))
	for _, file := range files {
		splitName := strings.Split(file, "_")
		if len(splitName) != 2 {
			continue
		}
		dt := splitName[0]
		ver := splitName[1]
		creationTime, err := time.Parse(backupCreationTimeFormat, dt)
		if err != nil {
			klog.ErrorS(err, "Failed to parse datetime part of the backup name", "name", file)
			continue
		}
		backups = append(backups, Backup{
			CreationTime: creationTime,
			Version:      ver,
		})
	}
	return backups
}
