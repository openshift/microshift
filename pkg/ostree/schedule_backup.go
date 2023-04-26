package ostree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"k8s.io/klog/v2"
)

type ostreeDeployment struct {
	ID     string `json:"id"`
	Booted bool   `json:"booted"`
}

type ostreeStatus struct {
	Deployments []ostreeDeployment `json:"deployments"`
}

var runOstreeStatus = func() ([]byte, error) {
	cmd := exec.Command("rpm-ostree", "status", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		klog.Errorf("rpm-ostree status --json failed: %s", stderr.String())
		return nil, err
	}

	return stdout.Bytes(), nil
}

func getOstreeStatus() (ostreeStatus, error) {
	status := ostreeStatus{}

	out, err := runOstreeStatus()
	if err != nil {
		return status, err
	}

	err = json.Unmarshal(out, &status)
	return status, err
}

func getBootedOstreeID(status ostreeStatus) (string, error) {
	if len(status.Deployments) == 0 {
		return "", fmt.Errorf("rpm-ostree status --json returned no deployments")
	}

	for _, d := range status.Deployments {
		if d.Booted {
			return d.ID, nil
		}
	}

	return "", fmt.Errorf("output of rpm-ostree status had no active deployments")
}

func ScheduleBackup() error {
	status, err := getOstreeStatus()
	if err != nil {
		return err
	}

	id, err := getBootedOstreeID(status)
	if err != nil {
		return err
	}

	data := nextBoot{
		Action:   actionBackup,
		OstreeID: id,
	}

	return data.Persist()
}
