package helpers

import "k8s.io/klog/v2"

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

func (d *dryrunOstree) ExtractCommit(path string) error {
	klog.InfoS("DRY RUN OSTREE: Extracting commit to the repository", "repo", "", "commit", path)
	return nil
}
