package compose

import (
	"fmt"
	"slices"
	"strings"

	"k8s.io/klog/v2"
)

type Ostree interface {
	UpdateSummary() error
	DoesRefExists(ref string) (bool, error)
	CreateAlias(origin string, aliases ...string) error
	ExtractCommit(path string) error
}

var _ Ostree = (*ostree)(nil)

type ostree struct {
	OstreeRepoPath string
}

func NewOstree(repo string) *ostree {
	// TODO: Check if repo exists and create
	return &ostree{OstreeRepoPath: repo}
}

// CreateAlias creates aliases for the reference.
func (o *ostree) CreateAlias(ref string, aliases ...string) error {
	klog.InfoS("Creating aliases for ostree reference", "ref", ref, "aliases", aliases)
	for _, alias := range aliases {
		_, _, err := runCommand(
			"ostree", "refs",
			fmt.Sprintf("--repo=%s", o.OstreeRepoPath),
			"--force",
			"--create", alias, ref)
		if err != nil {
			return fmt.Errorf("failed to create alias %q pointing to %q: %w", alias, ref, err)
		}
	}

	klog.InfoS("Created aliases for ostree reference", "ref", ref, "aliases", aliases)
	return o.UpdateSummary()
}

// DoesRefExists checks if ref is present in the ostree repository.
func (o *ostree) DoesRefExists(ref string) (bool, error) {
	klog.InfoS("Checking if ostree ref exists", "ref", ref)
	sout, _, err := runCommand(
		"ostree", "refs",
		fmt.Sprintf("--repo=%s", o.OstreeRepoPath))
	if err != nil {
		return false, fmt.Errorf("failed to obtain refs of the ostree repository: %w", err)
	}

	refs := slices.DeleteFunc(strings.Split(sout, "\n"), func(line string) bool { return line == "" })
	exists := slices.Contains(refs, ref)

	klog.InfoS("Checked if ostree ref exists", "ref", ref, "exists", exists, "all-refs", refs)
	return exists, nil
}

// UpdateSummary updates summary of the ostree repository.
func (o *ostree) UpdateSummary() error {
	klog.InfoS("Updating ostree repository summary")

	_, _, err := runCommand(
		"ostree", "summary",
		fmt.Sprintf("--repo=%s", o.OstreeRepoPath),
		"--update")
	if err != nil {
		return fmt.Errorf("failed to update ostree repository's summary: %w", err)
	}

	// output is logged by runCommand()
	_, _, err = runCommand(
		"ostree", "summary",
		fmt.Sprintf("--repo=%s", o.OstreeRepoPath),
		"--view")
	if err != nil {
		return fmt.Errorf("failed to view ostree repository's summary: %w", err)
	}

	klog.InfoS("Updated ostree repository summary")
	return nil
}

// ExtractCommit extracts given tar archive to the ostree repository.
func (o *ostree) ExtractCommit(path string) error {
	klog.InfoS("Extracting commit to the repository", "repo", o.OstreeRepoPath, "commit", path)

	_, _, err := runCommand("tar",
		"--strip-components=2",
		"-C", o.OstreeRepoPath,
		"-xf", path,
	)
	if err != nil {
		return fmt.Errorf("failed to extract commit tar archive (%q) to ostree repo (%q): %w", path, o.OstreeRepoPath, err)
	}

	klog.InfoS("Extracted commit to the repository", "repo", o.OstreeRepoPath, "commit", path)
	return o.UpdateSummary()
}

var _ Ostree = (*dryrunOstree)(nil)

type dryrunOstree struct{}

func NewDryRunOstree() *dryrunOstree {
	return &dryrunOstree{}
}

func (d *dryrunOstree) CreateAlias(origin string, aliases ...string) error {
	klog.InfoS("DRY RUN OSTREE: Creating aliases", "origin", origin, "aliases", aliases)
	return nil
}

func (d *dryrunOstree) DoesRefExists(ref string) (bool, error) {
	klog.InfoS("DRY RUN OSTREE: Checking if ostree ref exists", "ref", ref, "exists", false)
	// returning false to keep the dry run flow going instead of skipping some steps
	return false, nil
}

func (d *dryrunOstree) UpdateSummary() error {
	klog.InfoS("DRY RUN OSTREE: Updating ostree repository summary")
	return nil
}

// ExtractCommit implements Ostree.
func (d *dryrunOstree) ExtractCommit(path string) error {
	klog.InfoS("DRY RUN OSTREE: Extracting commit to the repository", "repo", "", "commit", path)
	return nil
}
