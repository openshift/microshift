package build

import (
	"context"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/testutil"
	"k8s.io/klog/v2"
)

type Opts struct {
	Composer helpers.Composer
	Ostree   helpers.Ostree
	Podman   helpers.Podman
	Utils    UtilProxy

	Force  bool
	DryRun bool

	Retries       int
	RetryInterval time.Duration

	Paths  *testutil.Paths
	Events testutil.EventManager
}

type Runner struct {
	Opts *Opts
}

func (ib *Runner) Build(ctx context.Context, toBuild Plan) error {
	klog.InfoS("Running preparation phase of all the builds in all the groups")
	for idx, group := range toBuild {
		klog.InfoS("Running Prepare phase group", "index", idx)
		if err := ib.prepareGroup(group); err != nil {
			return err
		}
	}
	klog.InfoS("Completed preparation phase of all the builds in all the groups")

	for idx, group := range toBuild {
		klog.InfoS("Running Execute phase group", "index", idx)
		if err := ib.executeGroup(ctx, group); err != nil {
			return err
		}
	}
	return nil
}

func (ib *Runner) prepareGroup(group Group) error {
	aeg := testutil.NewAllErrGroup()

	for _, build := range group {
		build := build
		aeg.Go(func() error {
			err := build.Prepare(ib.Opts)
			if err != nil {
				klog.ErrorS(err, "Build preparation error")
			}
			return err
		})
	}

	return aeg.Wait()
}

func (ib *Runner) executeGroup(ctx context.Context, group Group) error {
	aeg := testutil.NewAllErrGroup()

	for _, build := range group {
		build := build
		aeg.Go(func() error {
			err := build.Execute(ctx, ib.Opts)
			if err != nil {
				klog.ErrorS(err, "Build execution error")
			}
			return err
		})
	}

	return aeg.Wait()
}

type UtilProxy interface {
	PathExistsAndIsNotEmpty(path string) (bool, error)
	Rename(oldpath, newpath string) error
}

var _ UtilProxy = (*utilProxy)(nil)

func NewUtilProxy() UtilProxy {
	return &utilProxy{}
}

type utilProxy struct{}

func (u *utilProxy) Rename(oldpath string, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (*utilProxy) PathExistsAndIsNotEmpty(path string) (bool, error) {
	return util.PathExistsAndIsNotEmpty(path)
}

func NewDryRunUtilProxy() UtilProxy {
	return &dryRunUtilProxy{}
}

type dryRunUtilProxy struct{}

func (u *dryRunUtilProxy) Rename(oldpath string, newpath string) error {
	return nil
}

func (*dryRunUtilProxy) PathExistsAndIsNotEmpty(path string) (bool, error) {
	return true, nil
}
