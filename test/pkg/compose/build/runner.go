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
	for _, group := range toBuild {
		if err := ib.buildGroup(group); err != nil {
			return err
		}
	}
	return nil
}

func (ib *Runner) buildGroup(group Group) error {
	eg, _ := errgroup.WithContext(context.TODO())

	for _, build := range group {
		build := build
		eg.Go(func() error {
			err := build.Execute(ib.Opts)
			if err != nil {
				klog.ErrorS(err, "Build error")
			}
			return err
		})
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
