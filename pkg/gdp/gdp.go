package gdp

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/openshift/microshift/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/squat/generic-device-plugin/deviceplugin"
	"k8s.io/klog/v2"

	dp "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type GenericDevicePlugin struct {
}

func NewGenericDevicePlugin(cfg *config.Config) *GenericDevicePlugin {
	return &GenericDevicePlugin{}
}

func (gdp *GenericDevicePlugin) Name() string           { return "generic-device-plugin" }
func (gdp *GenericDevicePlugin) Dependencies() []string { return []string{"kubelet"} }

func (gdp *GenericDevicePlugin) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	devSpec := &deviceplugin.DeviceSpec{
		Name: "squat.ai/serial",
		Groups: []*deviceplugin.Group{{
			Paths: []*deviceplugin.Path{{Path: "/dev/ttyACM*"}},
		}},
	}
	klog.Infof("Created DeviceSpec: %#v | Path: %s", devSpec, devSpec.Groups[0].Paths[0].Path)
	devSpec.Default()
	klog.Infof("DeviceSpec with Defaults: %#v | Path: %s", devSpec, devSpec.Groups[0].Paths[0].Path)

	klog.Info("Creating GenericPlugin")
	gp := deviceplugin.NewGenericPlugin(devSpec, dp.DevicePluginPath, log.NewJSONLogger(&loggerThingy{}), prometheus.NewRegistry(), false)
	klog.Info("Created GenericPlugin")

	close(ready)

	return gp.Run(ctx)
}

type loggerThingy struct {
}

func (lt *loggerThingy) Write(p []byte) (n int, err error) {
	klog.Infof("%s", string(p))
	return len(p), nil
}
