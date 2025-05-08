package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/klog/v2"

	"github.com/squat/generic-device-plugin/deviceplugin"

	"k8s.io/apimachinery/pkg/util/validation"
)

type GenericDevicePlugin struct {
	// Generic Device Plugin status, can be Enabled or Disabled
	// +kubebuilder:default=Disabled
	// +kubebuilder:validation:Enum:=Enabled;Disabled;""
	Status string `json:"status"`

	// Domain with which devices will be present in the cluster,
	// e.g. device.microshift.io/serial.
	// +kubebuilder:default=device.microshift.io
	Domain string `json:"domain"`

	// TODO: Fix schemaless - multiple nested slices do not play nicely with example configs and produce invalid default yaml config

	// Devices configuration
	// +kubebuilder:validation:Schemaless
	Devices []deviceplugin.DeviceSpec `json:"devices"`
}

func (gdp GenericDevicePlugin) incorporateUserSettings(c *Config) {
	klog.Info("gdp.incorporateUserSettings")

	if gdp.Status != "" {
		klog.Info("gdp.incorporateUserSettings - update status")
		c.GenericDevicePlugin.Status = gdp.Status
	}

	if gdp.Domain != "" {
		klog.Info("gdp.incorporateUserSettings - update domain")
		c.GenericDevicePlugin.Domain = gdp.Domain
	}

	if len(gdp.Devices) > 0 {
		klog.Info("gdp.incorporateUserSettings - update devices")
		c.GenericDevicePlugin.Devices = gdp.Devices
	}

	klog.Info("gdp.incorporateUserSettings - end")
}

func (gdp GenericDevicePlugin) validate() error {
	if gdp.Status != "Enabled" && gdp.Status != "Disabled" {
		return fmt.Errorf("genericDevicePlugin.Status must be either 'Enabled' or 'Disabled'")
	}

	if gdp.Status != "Enabled" {
		return nil
	}

	errs := []error{}

	// From https://github.com/squat/generic-device-plugin/blob/main/main.go
	if domainErrs := validation.IsDNS1123Subdomain(gdp.Domain); len(domainErrs) > 0 {
		errs = append(errs, fmt.Errorf("failed to parse domain %q: %s", gdp.Domain, strings.Join(domainErrs, ", ")))
	}

	if len(gdp.Devices) == 0 {
		errs = append(errs, fmt.Errorf("genericDevicePlugin.Devices is empty - at least one device must be specified"))
	}

	deviceTypeFmt := "[a-z0-9][-a-z0-9]*[a-z0-9]"
	deviceTypeRegexp := regexp.MustCompile("^" + deviceTypeFmt + "$")

	for _, deviceSpec := range gdp.Devices {
		if !deviceTypeRegexp.MatchString(strings.TrimSpace(deviceSpec.Name)) {
			errs = append(errs, fmt.Errorf("failed to parse device %q; device name must match the regular expression %q", deviceSpec.Name, deviceTypeFmt))
		}

		for _, g := range deviceSpec.Groups {
			if len(g.Paths) > 0 && len(g.USBSpecs) > 0 {
				errs = append(errs, fmt.Errorf(
					"failed to parse device %q; cannot define both path and usb at the same time",
					deviceSpec.Name))
			}
		}
	}

	return errors.Join(errs...)
}

func genericDevicePluginDefaults() GenericDevicePlugin {
	return GenericDevicePlugin{
		Status: "Disabled",
		Domain: "device.microshift.io",
		Devices: []deviceplugin.DeviceSpec{
			{
				Name: "serial",
				Groups: []*deviceplugin.Group{
					{Paths: []*deviceplugin.Path{{Path: "/dev/ttyUSB*"}}},
					{Paths: []*deviceplugin.Path{{Path: "/dev/ttyACM*"}}},
					{Paths: []*deviceplugin.Path{{Path: "/dev/tty.usb*"}}},
					{Paths: []*deviceplugin.Path{{Path: "/dev/cu.*"}}},
					{Paths: []*deviceplugin.Path{{Path: "/dev/cuaU*"}}},
					{Paths: []*deviceplugin.Path{{Path: "/dev/rfcomm*"}}},
				},
			},
			{
				Name: "video",
				Groups: []*deviceplugin.Group{
					{
						Paths: []*deviceplugin.Path{{Path: "/dev/video0"}},
					},
				},
			},
			{
				Name: "fuse",
				Groups: []*deviceplugin.Group{
					{
						Paths: []*deviceplugin.Path{{Path: "/dev/fuse"}},
						Count: 10,
					},
				},
			},
			{
				Name: "audio",
				Groups: []*deviceplugin.Group{
					{
						Paths: []*deviceplugin.Path{{Path: "/dev/snd"}},
						Count: 10,
					},
				},
			},
			{
				Name: "capture",
				Groups: []*deviceplugin.Group{
					{
						Paths: []*deviceplugin.Path{
							{Path: "/dev/snd/controlC0"},
							{Path: "/dev/snd/pcmC0D0c"},
						},
					},
					{
						Paths: []*deviceplugin.Path{
							{Path: "/dev/snd/controlC1", MountPath: "/dev/snd/controlC0"},
							{Path: "/dev/snd/pcmC1D0c", MountPath: "/dev/snd/pcmC0D0c"},
						},
					},
					{
						Paths: []*deviceplugin.Path{
							{Path: "/dev/snd/controlC2", MountPath: "/dev/snd/controlC0"},
							{Path: "/dev/snd/pcmC2D0c", MountPath: "/dev/snd/pcmC0D0c"},
						},
					},
					{
						Paths: []*deviceplugin.Path{
							{Path: "/dev/snd/controlC3", MountPath: "/dev/snd/controlC0"},
							{Path: "/dev/snd/pcmC3D0c", MountPath: "/dev/snd/pcmC0D0c"},
						},
					},
				},
			},
		},
	}
}
