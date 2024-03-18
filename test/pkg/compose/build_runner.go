package compose

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type BuildRunner struct {
	Composer Composer
}

func (ib *BuildRunner) Build(toBuild BuildPlan) error {
	for _, group := range toBuild {
		if err := ib.buildGroup(group); err != nil {
			return err
		}
	}
	return nil
}

func (ib *BuildRunner) buildGroup(group BuildGroup) error {
	eg, _ := errgroup.WithContext(context.TODO())

	for _, build := range group {
		build := build
		eg.Go(func() error { return build.Execute(ib.Composer) })
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
