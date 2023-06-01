package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/openshift/microshift/pkg/util"
)

type rpmOstreeDeployment struct {
	ID     DeploymentID
	Booted bool
}

type rpmOstreeStatus struct {
	Deployments []rpmOstreeDeployment
}

func getRPMOSTreeStatus() (*rpmOstreeStatus, error) {
	cmd := exec.Command("rpm-ostree", "status", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rpm-ostree status --json failed: %s", stderr.String())
	}

	status := &rpmOstreeStatus{}
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		return nil, err
	}

	return status, nil
}

func getCurrentDeploymentID() (DeploymentID, error) {
	status, err := getRPMOSTreeStatus()
	if err != nil {
		return "", fmt.Errorf("getting current deployment ID failed: %w", err)
	}

	if len(status.Deployments) == 0 {
		return "", fmt.Errorf("rpm-ostree reported 0 deployments")
	}

	for _, d := range status.Deployments {
		if d.Booted {
			return d.ID, nil
		}
	}

	return "", fmt.Errorf("rpm-ostree reported 0 booted deployments")
}

func isOSTree() (bool, error) {
	return util.PathExists("/run/ostree-booted")
}
