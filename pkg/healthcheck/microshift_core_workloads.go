package healthcheck

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

// getCoreMicroShiftWorkloads assembles a structure with the core MicroShift
// workloads that the healthcheck should verify.
func getCoreMicroShiftWorkloads() (map[string]NamespaceWorkloads, error) {
	cfg, err := config.ActiveConfig()
	if err != nil {
		return nil, err
	}

	workloads := map[string]NamespaceWorkloads{
		"openshift-service-ca": {
			Deployments: []string{"service-ca"},
		},
		"openshift-ingress": {
			Deployments: []string{"router-default"},
		},
		"openshift-dns": {
			DaemonSets: []string{
				"dns-default",
				"node-resolver",
			},
		},
	}
	if cfg.Network.IsEnabled() {
		workloads["openshift-ovn-kubernetes"] = NamespaceWorkloads{
			DaemonSets: []string{"ovnkube-master", "ovnkube-node"},
		}
	}
	if err := fillOptionalWorkloadsIfApplicable(cfg, workloads); err != nil {
		return nil, err
	}

	return workloads, nil
}

func lvmsIsExpected(cfg *config.Config) (bool, error) {
	cfgFile := "/etc/microshift/lvmd.yaml"
	if exists, err := util.PathExists(cfgFile); err != nil {
		return false, err
	} else if exists {
		klog.Infof("%s exists - expecting LVMS to be deployed", cfgFile)
		return true, nil
	}

	if !cfg.Storage.IsEnabled() {
		klog.Infof("LVMS is disabled via config. Configured value: %q", string(cfg.Storage.Driver))
		return false, nil
	}

	cmd := exec.Command("vgs", "--readonly", "--options=name", "--reportformat=json")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	out := &bytes.Buffer{}
	err = json.Compact(out, output)
	if err != nil {
		klog.Errorf("Failed to compact 'vgs' output: %s", string(output))
	} else {
		klog.V(2).Infof("vgs reported: %s", out.String())
	}

	report := struct {
		Report []struct {
			VGs []struct {
				VGName string `json:"vg_name"`
			} `json:"vg"`
		} `json:"report"`
	}{}

	err = json.Unmarshal(output, &report)
	if err != nil {
		return false, err
	}

	if len(report.Report) == 0 || len(report.Report[0].VGs) == 0 {
		klog.Infof("Detected 0 volume groups - LVMS is not expected")
		return false, nil
	}

	if len(report.Report[0].VGs) == 1 {
		klog.Infof("Detected 1 volume group (%s) - LVMS is expected", report.Report[0].VGs[0].VGName)
		return true, nil
	}

	for _, vg := range report.Report[0].VGs {
		if vg.VGName == "microshift" {
			klog.Infof("Found volume group named 'microshift' - LVMS is expected")
			return true, nil
		}
	}

	return false, nil
}

func fillOptionalWorkloadsIfApplicable(cfg *config.Config, workloads map[string]NamespaceWorkloads) error {
	if expected, err := lvmsIsExpected(cfg); err != nil {
		return err
	} else if expected {
		workloads["openshift-storage"] = NamespaceWorkloads{
			DaemonSets:  []string{"vg-manager"},
			Deployments: []string{"lvms-operator"},
		}
	}
	if comps := getExpectedCSIComponents(cfg); len(comps) != 0 {
		klog.Infof("At least one CSI Component is enabled")
		workloads["kube-system"] = NamespaceWorkloads{
			Deployments: comps,
		}
	}
	return nil
}

func getExpectedCSIComponents(cfg *config.Config) []string {
	klog.V(2).Infof("Configured optional CSI components: %v", cfg.Storage.OptionalCSIComponents)

	if len(cfg.Storage.OptionalCSIComponents) == 0 {
		return []string{"csi-snapshot-controller"}
	}

	// Validation fails when there's more than one component provided and one of them is "None".
	// In other words: if "None" is used, it can be the only element.
	if len(cfg.Storage.OptionalCSIComponents) == 1 && cfg.Storage.OptionalCSIComponents[0] == config.CsiComponentNone {
		return nil
	}

	deployments := []string{}
	for _, comp := range cfg.Storage.OptionalCSIComponents {
		if comp == config.CsiComponentSnapshot {
			deployments = append(deployments, "csi-snapshot-controller")
		}
	}
	return deployments
}
