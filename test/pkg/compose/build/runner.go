package build

import (
	"context"

	"github.com/openshift/microshift/test/pkg/compose/helpers"
	"golang.org/x/sync/errgroup"
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
		eg.Go(func() error { return build.Execute(ib.Opts) })
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
