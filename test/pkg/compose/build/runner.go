package build

import (
	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"github.com/openshift/microshift/test/pkg/testutil"
	"k8s.io/klog/v2"
)

type Opts struct {
	Composer helpers.Composer
	Ostree   helpers.Ostree

	Force            bool
	DryRun           bool
	ArtifactsMainDir string

	Junit *testutil.JUnit
}

type Runner struct {
	Opts *Opts
}

func (ib *Runner) Build(toBuild Plan) error {
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
		if err := ib.executeGroup(group); err != nil {
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

func (ib *Runner) executeGroup(group Group) error {
	aeg := testutil.NewAllErrGroup()

	for _, build := range group {
		build := build
		aeg.Go(func() error {
			err := build.Execute(ib.Opts)
			if err != nil {
				klog.ErrorS(err, "Build execution error")
			}
			return err
		})
	}

	return aeg.Wait()
}
