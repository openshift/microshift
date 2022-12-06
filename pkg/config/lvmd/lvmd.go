package lvmd

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"
)

const (
	LvmdConfigFileName          = "lvmd.yaml"
	defaultSockName             = "/run/lvmd/lvmd.socket"
	defaultRHEL4EdgeVolumeGroup = "rhel"
)

// Lvmd stores the read-in or defaulted values of the lvmd configuration and provides the topolvm-node process information
// about its host's storage environment.
type Lvmd struct {
	DeviceClasses []*DeviceClass `json:"device-classes"`
	SocketName    string         `json:"socket-name"`
}

func (l *Lvmd) WithDefaults() *Lvmd {
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

func NewLvmdConfigFromFile(p string) (*Lvmd, error) {
	l := new(Lvmd)
	buf, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read lvmd file: %v", err)
	}

	err = yaml.Unmarshal(buf, &l)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling lvmd file: %v", err)
	}
	if l.SocketName == "" {
		l.SocketName = defaultSockName
	}
	return l, nil
}
