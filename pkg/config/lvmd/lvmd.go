package lvmd

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"
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
	if l.SocketName == "" {
		l.SocketName = defaultSockName
	}
	if len(l.DeviceClasses) == 0 {
		l.DeviceClasses = []*DeviceClass{
			{
				Name:        "default",
				VolumeGroup: defaultRHEL4EdgeVolumeGroup,
				Default:     true,
				SpareGB:     func() *uint64 { s := uint64(defaultSpareGB); return &s }(),
			},
		}
	}
	return l
}

type Config struct {
	Path       string
	LvmdConfig *Lvmd
}

func newLvmdConfigFromFile(p string) (*Lvmd, error) {
	l := new(Lvmd)
	buf, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %v", err)
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

func NewLvmdConfigFromFileOrDefault(path string) (*Lvmd, error) {
	if path == "" {
		l := new(Lvmd)
		return l.withDefaults(), nil
	}
	return newLvmdConfigFromFile(path)
}
