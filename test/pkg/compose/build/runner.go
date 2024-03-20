package build

import (
	"context"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
)

type Opts struct {
	Composer helpers.Composer
	Ostree   helpers.Ostree

	Force            bool
	DryRun           bool
	ArtifactsMainDir string
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
	eg, _ := errgroup.WithContext(context.TODO())

	for _, build := range group {
		build := build
		eg.Go(func() error {
			err := build.Prepare(ib.Opts)
			if err != nil {
				klog.ErrorS(err, "Build preparation error")
			}
			return err
		})
	}

	return eg.Wait()
}

func (ib *Runner) executeGroup(group Group) error {
	eg, _ := errgroup.WithContext(context.TODO())

	for _, build := range group {
		build := build
		eg.Go(func() error {
			err := build.Execute(ib.Opts)
			if err != nil {
				klog.ErrorS(err, "Build execution error")
			}
			return err
		})
	}

	return eg.Wait()
}
