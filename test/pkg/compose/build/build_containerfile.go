package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/test/pkg/testutil"
	"k8s.io/klog/v2"
)

var _ Build = (*ContainerfileBuild)(nil)

type ContainerfileBuild struct {
	build
}

func NewContainerfileBuild(path string, opts *PlannerOpts) (*ContainerfileBuild, error) {
	klog.InfoS("Constructing BlueprintBuild", "path", path)
	start := time.Now()

	filename := filepath.Base(path)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	opts.Events.AddEvent(&testutil.Event{
		Name:      name,
		Suite:     "render",
		ClassName: "containerfile",
		Start:     start,
		End:       time.Now(),
	})

	cb := &ContainerfileBuild{
		build: build{
			Name: name,
			Path: path,
		},
	}

	return cb, nil
}

func (c *ContainerfileBuild) Prepare(opts *Opts) error {
	return nil
}

func (c *ContainerfileBuild) Execute(opts *Opts) error {
	// podman's libpod doesn't give enough value but brings a lot of dependencies
	// https://pkg.go.dev/github.com/containers/podman/libpod#Runtime.Build
	// https://github.com/containers/buildah/tree/main/imagebuildah is an overkill - gives too much manual control.

	start := time.Now()

	outputDir := filepath.Join(opts.ArtifactsMainDir, "bootc-images")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	output := filepath.Join(outputDir, c.Name)
	if exists, err := util.PathExists(output); err != nil {
		return err
	} else if exists {
		if !opts.Force {
			klog.InfoS("Containerfile already exists and --force wasn't present - skipping", "path", output)
			opts.Events.AddEvent(&testutil.SkippedEvent{
				Event: testutil.Event{
					Name:      c.Name,
					Suite:     "compose",
					ClassName: "containerfile",
					Start:     start,
					End:       time.Now(),
				},
				Message: "Containerfile already exists and --force wasn't present - skipping",
			})

			return nil
		}

		if err := os.RemoveAll(output); err != nil {
			return err
		}
	}

	err := opts.Podman.BuildAndSave(c.Name, c.Path, filepath.Join(opts.ArtifactsMainDir, "rpm-repos"), output)
	if err != nil {
		klog.ErrorS(err, "Failed to build Containerfile", "name", c.Name)
		opts.Events.AddEvent(&testutil.FailedEvent{
			Event: testutil.Event{
				Name:      c.Name,
				Suite:     "compose",
				ClassName: "containerfile",
				Start:     start,
				End:       time.Now(),
			},
			Message: "Failed to build Containerfile",
			Content: err.Error(),
		})
		return fmt.Errorf("failed to build Containerfile: %v", err)
	}

	opts.Events.AddEvent(&testutil.Event{
		Name:      c.Name,
		Suite:     "compose",
		ClassName: "containerfile",
		Start:     start,
		End:       time.Now(),
	})
	return nil
}
