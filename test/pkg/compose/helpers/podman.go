package helpers

import (
	"context"
	"fmt"
	"os"

	"github.com/openshift/microshift/test/pkg/testutil"
	"k8s.io/klog/v2"
)

type Podman interface {
	BuildAndSave(ctx context.Context, tag, containerfilePath, contextDir, output string) error
}

var _ Podman = (*podman)(nil)

type podman struct{}

func NewPodman() *podman {
	return &podman{}
}

func (p *podman) BuildAndSave(ctx context.Context, tag, containerfilePath, contextDir, output string) error {
	klog.InfoS("Building Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)
	_, _, err := testutil.RunCommandWithContext(ctx, "podman", "build", "--tag", tag, "--file", containerfilePath, contextDir)
	if err != nil {
		return fmt.Errorf("failed to build: %v", err)
	}
	klog.InfoS("Built Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)

	klog.InfoS("Saving Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)
	_, _, err = testutil.RunCommandWithContext(ctx, "podman", "save", "--format", "oci-dir", "-o", output, tag)
	if err != nil {
		if ctx.Err() != nil {
			if err := os.RemoveAll(output); err != nil {
				klog.ErrorS(err, "Failed to remove path", "path", output)
			}
			return ctx.Err()
		}
		return fmt.Errorf("failed to save: %v", err)
	}
	klog.InfoS("Saved Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)

	return nil
}

var _ Podman = (*dryrunPodman)(nil)

type dryrunPodman struct{}

func NewDryRunPodman() *dryrunPodman {
	return &dryrunPodman{}
}

func (p *dryrunPodman) BuildAndSave(ctx context.Context, tag, containerfilePath, contextDir, outputFile string) error {
	klog.InfoS("DRYRUN: Building and saving Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", outputFile, "contextDir", contextDir)
	return nil
}
