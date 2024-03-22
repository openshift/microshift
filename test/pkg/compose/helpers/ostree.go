package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/testutil"

	"k8s.io/klog/v2"
)

type Ostree interface {
	DoesRefExists(ref string) (bool, error)
	CreateAlias(origin string, aliases ...string) error
	ExtractCommit(path string) error
}

var _ Ostree = (*ostree)(nil)

type ostree struct {
	OstreeRepoPath string
	mutex          sync.Mutex
}

func NewOstree(repo string) (*ostree, error) {
	if exists, err := util.PathExists(repo); err != nil {
		return nil, err
	} else if !exists {
		if err := os.MkdirAll(filepath.Join(repo, ".."), 0755); err != nil {
			return nil, fmt.Errorf("failed to create path %q: %w", filepath.Join(repo, ".."), err)
		}
		_, _, err := testutil.RunCommand("ostree", "init", "--repo", repo)
		if err != nil {
			return nil, fmt.Errorf("failed to create ostree repo %q: %w", repo, err)
		}
	}

	return &ostree{OstreeRepoPath: repo}, nil
}

// CreateAlias creates aliases for the reference.
func (o *ostree) CreateAlias(ref string, aliases ...string) error {
	klog.InfoS("Creating aliases for ostree reference", "ref", ref, "aliases", aliases)

	o.mutex.Lock()
	defer o.mutex.Unlock()

	for _, alias := range aliases {
		_, _, err := testutil.RunCommand(
			"ostree", "refs",
			"--repo", o.OstreeRepoPath,
			"--force",
			"--create", alias, ref)
		if err != nil {
			return fmt.Errorf("failed to create alias %q pointing to %q: %w", alias, ref, err)
		}
	}

	klog.InfoS("Created aliases for ostree reference", "ref", ref, "aliases", aliases)
	return o.updateSummary()
}

// DoesRefExists checks if ref is present in the ostree repository.
func (o *ostree) DoesRefExists(ref string) (bool, error) {
	klog.InfoS("Checking if ostree ref exists", "ref", ref)

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if exists, err := util.PathExistsAndIsNotEmpty(o.OstreeRepoPath); err != nil {
		return false, fmt.Errorf("failed to check if ostree repo %q exists: %w", o.OstreeRepoPath, err)
	} else if !exists {
		// If repo doesn't exist (yet), the ref also doesn't exist
		klog.InfoS("Ostree repository doesn't exist yet, therefore ref also doesn't exist", "ref", ref, "exists", false)
		return false, nil
	}

	sout, _, err := testutil.RunCommand("ostree", "refs", "--repo", o.OstreeRepoPath)
	if err != nil {
		return false, fmt.Errorf("failed to obtain refs of the ostree repository: %w", err)
	}

	refs := slices.DeleteFunc(strings.Split(sout, "\n"), func(line string) bool { return line == "" })
	exists := slices.Contains(refs, ref)

	klog.InfoS("Checked if ostree ref exists", "ref", ref, "exists", exists, "all-refs", refs)
	return exists, nil
}

// ExtractCommit extracts given tar archive to the ostree repository.
func (o *ostree) ExtractCommit(path string) error {
	klog.InfoS("Extracting commit to the repository", "repo", o.OstreeRepoPath, "commit", path)

	o.mutex.Lock()
	defer o.mutex.Unlock()

	_, _, err := testutil.RunCommand("tar",
		"--strip-components=2",
		"-C", o.OstreeRepoPath,
		"-xf", path,
	)
	if err != nil {
		return fmt.Errorf("failed to extract commit tar archive (%q) to ostree repo (%q): %w", path, o.OstreeRepoPath, err)
	}

	klog.InfoS("Extracted commit to the repository", "repo", o.OstreeRepoPath, "commit", path)
	return o.updateSummary()
}

// updateSummary updates summary of the ostree repository.
func (o *ostree) updateSummary() error {
	klog.InfoS("Updating ostree repository summary")

	_, _, err := testutil.RunCommand(
		"ostree", "summary",
		"--repo", o.OstreeRepoPath,
		"--update")
	if err != nil {
		return fmt.Errorf("failed to update ostree repository's summary: %w", err)
	}

	// output is logged by testutil.RunCommand()
	_, _, err = testutil.RunCommand(
		"ostree", "summary",
		"--repo", o.OstreeRepoPath,
		"--view")
	if err != nil {
		return fmt.Errorf("failed to view ostree repository's summary: %w", err)
	}

	klog.InfoS("Updated ostree repository summary")
	return nil
}
