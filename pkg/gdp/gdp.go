package gdp

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/squat/generic-device-plugin/deviceplugin"
	"k8s.io/klog/v2"

	dp "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type GenericDevicePlugin struct {
	configuration config.GenericDevicePlugin
}

func NewGenericDevicePlugin(cfg *config.Config) *GenericDevicePlugin {
	return &GenericDevicePlugin{configuration: cfg.GenericDevicePlugin}
}

func (gdp *GenericDevicePlugin) Name() string           { return "generic-device-plugin" }
func (gdp *GenericDevicePlugin) Dependencies() []string { return []string{"kubelet"} }

func (gdp *GenericDevicePlugin) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	if gdp.configuration.Status == "Disabled" {
		klog.Infof("Generic Device Plugin is Disabled")
		close(ready)
		close(stopped)
		return nil
	}
	klog.Infof("Generic Device Plugin is Enabled")

	defer close(stopped)

	// From https://github.com/squat/generic-device-plugin/blob/main/main.go
	deviceSpecs := gdp.configuration.Devices
	for i := range deviceSpecs {
		deviceSpecs[i].Default()
		trim := strings.TrimSpace(deviceSpecs[i].Name)
		deviceSpecs[i].Name = path.Join(gdp.configuration.Domain, trim)
		for j := range deviceSpecs[i].Groups {
			for k := range deviceSpecs[i].Groups[j].Paths {
				deviceSpecs[i].Groups[j].Paths[k].Path = strings.TrimSpace(deviceSpecs[i].Groups[j].Paths[k].Path)
				deviceSpecs[i].Groups[j].Paths[k].MountPath = strings.TrimSpace(deviceSpecs[i].Groups[j].Paths[k].MountPath)
			}
		}
	}
	if len(deviceSpecs) == 0 {
		return fmt.Errorf("at least one device must be specified")
	}

	aeg := &util.AllErrGroup{}
	for i, deviceSpec := range deviceSpecs {
		enableUSBDiscovery := false
		for _, g := range deviceSpec.Groups {
			if len(g.USBSpecs) > 0 {
				enableUSBDiscovery = true
				break
			}
		}
		gp := deviceplugin.NewGenericPlugin(&deviceSpec,
			dp.DevicePluginPath,
			log.NewJSONLogger(&logPassthrough{}),
			nil,
			enableUSBDiscovery)

		aeg.Go(func() error {
			if err := gp.Run(ctx); err != nil {
				klog.Errorf("Generic Plugin for %s failed: %v", deviceSpecs[i].Name, err)
				return err
			}
			return nil
		})
	}

	close(ready)
	errs := aeg.Wait()
	if errs != nil {
		return errs
	}

	return ctx.Err()
}

type logPassthrough struct{}

func (lt *logPassthrough) Write(p []byte) (n int, err error) {
	klog.Infof("%s", string(p))
	return len(p), nil
}
