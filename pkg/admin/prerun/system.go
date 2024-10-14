package prerun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

func getCurrentBootID() (string, error) {
	path := "/proc/sys/kernel/random/boot_id"
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %q: %w", path, err)
	}
	// removing dashes to get the same format as `journalctl --list-boots`
	return strings.ReplaceAll(strings.TrimSpace(string(content)), "-", ""), nil
}

type deployment struct {
	ID     string `json:"id"`
	Booted bool   `json:"booted"`
	Staged bool   `json:"staged"`
	Pinned bool   `json:"pinned"`
}

func getDeploymentsFromOSTree() ([]deployment, error) {
	cmd := exec.Command("rpm-ostree", "status", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%q failed: %q: %w", cmd, stderr.String(), err)
	}

	outputToLog := stdout.String()
	outputToLog = strings.ReplaceAll(outputToLog, "\n", "")
	outputToLog = strings.ReplaceAll(outputToLog, " ", "")
	klog.InfoS("rpm-ostree status", "output", outputToLog)

	status := struct {
		Deployments []deployment `json:"deployments"`
	}{}

	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", stdout.Bytes(), err)
	}

	if len(status.Deployments) == 0 {
		return nil, fmt.Errorf("unexpected amount (0) of deployments from rpm-ostree status output")
	}

	klog.InfoS("OSTree deployments",
		"deployments", status.Deployments)
	return status.Deployments, nil
}

func GetCurrentDeploymentID() (string, error) {
	deployments, err := getDeploymentsFromOSTree()
	if err != nil {
		return "", fmt.Errorf("failed to get deployments: %w", err)
	}

	for _, deployment := range deployments {
		if deployment.Booted {
			return deployment.ID, nil
		}
	}

	return "", fmt.Errorf("could not find booted deployment in %#v", deployments)
}

func getAllDeploymentIDs() ([]string, error) {
	deployments, err := getDeploymentsFromOSTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	ids := make([]string, 0, len(deployments))
	for _, deployment := range deployments {
		ids = append(ids, deployment.ID)
	}

	return ids, nil
}
