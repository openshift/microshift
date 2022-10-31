package lvmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"k8s.io/klog/v2"
)

const (
	defaultSockName             = "/run/lvmd/lvmd.socket"
	defaultRHEL4EdgeVolumeGroup = "rhel"
)

// Lvmd stores the read-in or defaulted values of the lvmd configuration and provides the topolvm-node process information
// about its host's storage environment.
type Lvmd struct {
	DeviceClasses []*DeviceClass `json:"device-classes"`
	SocketName    string         `json:"socket-name"`
}

func (l *Lvmd) withDefaults() *Lvmd {
	l.SocketName = defaultSockName
	l.DeviceClasses = []*DeviceClass{
		{
			Name:        "default",
			VolumeGroup: defaultRHEL4EdgeVolumeGroup,
			Default:     true,
			SpareGB:     func() *uint64 { s := uint64(defaultSpareGB); return &s }(),
		},
	}
	return l
}

func newLvmdConfigFromFile(p string) (*Lvmd, error) {
	l := new(Lvmd)
	buf, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &l)
	if err != nil {
		return nil, fmt.Errorf("parsing lvmd config: %v", err)
	}
	if l.SocketName == "" {
		l.SocketName = defaultSockName
	}
	return l, nil
}

// NewLvmdConfigFromFileOrDefault takes a path to a lvmd config file.  If the file exists and is readable, returns the
// unmarshalled config *Lvmd and no error.  If the file does not exist, is inaccessible, or fails to unmarshall,
// a nil ptr and the error are returned.
// The defaulting behavior exists for cases where a microshift config file was found, but the user has not specified
// a lvmd file.
func NewLvmdConfigFromFileOrDefault(path string) (*Lvmd, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			klog.Infof("lvmd file not found, assuming default values")
			return new(Lvmd).withDefaults(), nil
		}
		return nil, fmt.Errorf("failed to get lvmd config file: %v", err)
	}

	l, err := newLvmdConfigFromFile(path)
	if err == nil {
		klog.Infof("got lvmd config from file %q", path)
		return l, nil
	}
	return nil, fmt.Errorf("getting lvmd config: %v", err)
}
