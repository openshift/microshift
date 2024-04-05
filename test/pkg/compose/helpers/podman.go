package helpers

import (
	"fmt"

	"github.com/openshift/microshift/test/pkg/testutil"
	"k8s.io/klog/v2"
)

type Podman interface {
	BuildAndSave(tag, containerfilePath, contextDir, output string) error
}

var _ Podman = (*podman)(nil)

type podman struct{}

func NewPodman() *podman {
	return &podman{}
}

func (p *podman) BuildAndSave(tag, containerfilePath, contextDir, output string) error {
	klog.InfoS("Building Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)
	_, _, err := testutil.RunCommand("podman", "build", "--tag", tag, "--file", containerfilePath, contextDir)
	if err != nil {
		return fmt.Errorf("failed to build: %v", err)
	}
	klog.InfoS("Built Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)

	klog.InfoS("Saving Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", output, "contextDir", contextDir)
	_, _, err = testutil.RunCommand("podman", "save", "--format", "oci-dir", "-o", output, tag)
	if err != nil {
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

func (p *dryrunPodman) BuildAndSave(tag, containerfilePath, contextDir, outputFile string) error {
	klog.InfoS("DRYRUN: Building and saving Containerfile", "tag", tag, "containerfile", containerfilePath, "destination", outputFile, "contextDir", contextDir)
	return nil
}
