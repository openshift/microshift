package gdp

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
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
	// TODO: Add skipped bits

	deviceSpecs := gdp.configuration.Devices
	for i, dsr := range deviceSpecs {
		deviceSpecs[i].Default()
		trim := strings.TrimSpace(deviceSpecs[i].Name)
		deviceSpecs[i].Name = path.Join(gdp.configuration.Domain, trim)
		for j, g := range deviceSpecs[i].Groups {
			if len(g.Paths) > 0 && len(g.USBSpecs) > 0 {
				return fmt.Errorf(
					"failed to parse device %q; cannot define both path and usb at the same time",
					dsr.Name,
				)
			}
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
	for i := range deviceSpecs {
		gp := deviceplugin.NewGenericPlugin(&deviceSpecs[i], dp.DevicePluginPath, log.NewJSONLogger(&loggerThingy{}), prometheus.NewRegistry(), false)
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

type loggerThingy struct {
}

func (lt *loggerThingy) Write(p []byte) (n int, err error) {
	klog.Infof("%s", string(p))
	return len(p), nil
}
