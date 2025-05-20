package lvmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	RuntimeLvmdConfigFile       = "/var/lib/microshift/lvms/lvmd.yaml"
	LvmdConfigFileName          = "lvmd.yaml"
	defaultSockName             = "/run/lvmd/lvmd.socket"
	defaultRHEL4EdgeVolumeGroup = "microshift"

	errorMessageNoVolumeGroups       = "No volume groups found"
	errorMessageMultipleVolumeGroups = "Multiple volume groups are available, but no configuration file was provided."
	statusMessageFoundDefault        = "Found default volume group \"microshift\""
	statusMessageDefaultAvailable    = "Defaulting to the only available volume group"
)

// Lvmd stores the read-in or defaulted values of the lvmd configuration and provides the topolvm-node process information
// about its host's storage environment.
type Lvmd struct {
	DeviceClasses []*DeviceClass `json:"device-classes"`
	SocketName    string         `json:"socket-name"`
	Message       string         `json:"-"`
}

// IsEnabled returns a boolean indicating whether the CSI driver
// should be enabled for this host.
func (l *Lvmd) IsEnabled() bool {
	return len(l.DeviceClasses) > 0
}

func uint64Ptr(val uint64) *uint64 {
	return &val
}

func getLvmdConfigForVGs(vgNames []string) (*Lvmd, error) {
	response := &Lvmd{
		SocketName: defaultSockName,
	}
	vgName := ""
	if len(vgNames) == 0 {
		response.Message = errorMessageNoVolumeGroups
		klog.V(2).Info(errorMessageNoVolumeGroups)
		return response, nil
	} else if len(vgNames) == 1 {
		vgName = vgNames[0]
		klog.V(2).Infof("Using volume group %q", vgName)
		response.Message = statusMessageDefaultAvailable
	} else {
		for _, name := range vgNames {
			if name == defaultRHEL4EdgeVolumeGroup {
				klog.V(2).Infof("Using default volume group %q", defaultRHEL4EdgeVolumeGroup)
				vgName = name
				response.Message = statusMessageFoundDefault
				break
			}
		}

		// If the default volume group is not found and there are
		// multiple volume groups, disable the CSI driver.
		if vgName == "" {
			klog.V(2).Infof("Multiple volume groups available but no configuration file is present, disabling CSI. %v", vgNames)
			response.Message = errorMessageMultipleVolumeGroups
			return response, nil
		}
	}

	// Fill in the default device class using the selected volume
	// group.
	response.DeviceClasses = []*DeviceClass{
		{
			Name:        "default",
			VolumeGroup: vgName,
			Default:     true,
			SpareGB:     uint64Ptr(defaultSpareGB),
		},
	}
	return response, nil
}

// DefaultLvmdConfig returns a configuration struct for Lvmd with
// default settings based on the current host. If a single volume
// group is found, that value is used. If multiple volume groups are
// available and one is named "rhel", that group is used. Otherwise,
// the configuration returned will report that it is not enabled (see
// IsEnabled()).
func DefaultLvmdConfig() (*Lvmd, error) {
	vgNames, err := getVolumeGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to discover local volume groups: %w", err)
	}
	return getLvmdConfigForVGs(vgNames)
}

// getVolumeGroups returns a slice of volume group names.
func getVolumeGroups() ([]string, error) {
	cmd := exec.Command("vgs", "--readonly", "--options=name", "--noheadings")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running vgs: %w", err)
	}
	names := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		newName := strings.Trim(line, " \t\n")
		if newName != "" {
			names = append(names, newName)
		}
	}
	return names, nil
}

func NewLvmdConfigFromFile(p string) (*Lvmd, error) {
	l := new(Lvmd)
	buf, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read lvmd file: %w", err)
	}

	err = yaml.Unmarshal(buf, &l)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling lvmd file: %w", err)
	}
	if l.SocketName == "" {
		l.SocketName = defaultSockName
	}
	l.Message = fmt.Sprintf("Read from %s", p)
	return l, nil
}

func SaveLvmdConfigToFile(l *Lvmd, p string) error {
	buf, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshalling lvmd config: %w", err)
	}
	err = os.WriteFile(p, buf, 0600)
	if err != nil {
		return fmt.Errorf("writing lvmd config: %w", err)
	}
	return nil
}

func LvmPresentOnMachine() error {
	if _, err := exec.LookPath("lvm"); err != nil {
		return fmt.Errorf("failed to find 'vgs' command line tool: %w", err)
	}
	return nil
}

// DefaultDeviceClassName scans []DeviceClasses and returns the name for the default device class per these rules:
// 1) If the DeviceClass is marked as default, return this name, else
// 2) If the there is not default DeviceClass and a device named "microshift" exists, return "microshift", else
// 3) Return the name of the first DeviceClass in the list.
func DefaultDeviceClassName(l *Lvmd) (string, error) {
	if l == nil {
		return "", fmt.Errorf("nil Lvmd")
	}
	if len(l.DeviceClasses) == 0 {
		return "", fmt.Errorf("no device classes found in Lvmd")
	}
	var n string
	for _, d := range l.DeviceClasses {
		if d.Default {
			n = d.Name
			break
		} else if d.Name == "microshift" {
			n = d.Name
		}
	}
	if n == "" {
		n = l.DeviceClasses[0].Name
	}
	return n, nil
}
