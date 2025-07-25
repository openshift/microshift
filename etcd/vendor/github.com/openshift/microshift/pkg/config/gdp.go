package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/squat/generic-device-plugin/deviceplugin"

	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	deviceTypeFmt = "[a-z0-9][-a-z0-9]*[a-z0-9]"
)

var (
	deviceTypeRegexp = regexp.MustCompile("^" + deviceTypeFmt + "$")
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

	// Devices configuration
	Devices []DeviceSpec `json:"devices"`
}

// GetGDPDevices transforms the Devices to be of a type compatible with the Generic Device Plugin implementation.
func (gdp *GenericDevicePlugin) GetGDPDevices() []deviceplugin.DeviceSpec {
	deviceSpecs := make([]deviceplugin.DeviceSpec, 0, len(gdp.Devices))
	for _, deviceSpec := range gdp.Devices {
		deviceSpecs = append(deviceSpecs, deviceSpec.toGDP())
	}
	return deviceSpecs
}

func (gdp GenericDevicePlugin) incorporateUserSettings(c *Config) {
	if gdp.Status != "" {
		c.GenericDevicePlugin.Status = gdp.Status
	}

	if gdp.Domain != "" {
		c.GenericDevicePlugin.Domain = gdp.Domain
	}

	if len(gdp.Devices) > 0 {
		c.GenericDevicePlugin.Devices = gdp.Devices
	}
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

	paths := make(map[string][]string)

	for i, deviceSpec := range gdp.Devices {
		if !deviceTypeRegexp.MatchString(strings.TrimSpace(deviceSpec.Name)) {
			errs = append(errs, fmt.Errorf("failed to parse device %q; device name must match the regular expression %q", deviceSpec.Name, deviceTypeFmt))
		}

		for j, g := range deviceSpec.Groups {
			validUSBs := make([]*USBSpec, 0, len(g.USBSpecs))
			for _, usb := range g.USBSpecs {
				// If USBSpec exists but is empty (e.g., comes from the config.yaml.default),
				// we treat it like nonexistent.
				if usb == nil || (usb.Vendor == "" && usb.Product == "" && usb.Serial == "") {
					continue
				}
				validUSBs = append(validUSBs, usb)
			}
			gdp.Devices[i].Groups[j].USBSpecs = validUSBs

			if len(g.Paths) > 0 && len(gdp.Devices[i].Groups[j].USBSpecs) > 0 {
				errs = append(errs, fmt.Errorf(
					"failed to parse device %q; cannot define both path and usb at the same time",
					deviceSpec.Name))
			}

			for _, path := range g.Paths {
				paths[path.Path] = append(paths[path.Path], deviceSpec.Name)
				errs = append(errs, path.validate()...)
			}

			for _, usb := range g.USBSpecs {
				errs = append(errs, usb.validate()...)
			}
		}
	}

	for path, devices := range paths {
		if len(devices) > 1 {
			errs = append(errs, fmt.Errorf("path '%s' is specified multiple times for devices: %s", path, strings.Join(devices, ", ")))
		}
	}

	return errors.Join(errs...)
}

func genericDevicePluginDefaults() GenericDevicePlugin {
	return GenericDevicePlugin{
		Status: "Disabled",
		Domain: "device.microshift.io",
		Devices: []DeviceSpec{
			{
				Name: "serial",
				Groups: []*Group{
					{
						Count: 1,
						Paths: []*Path{{
							Path:        "/dev/ttyACM0",
							MountPath:   "/dev/ttyACM0",
							Permissions: "mrw",
							ReadOnly:    false,
							Type:        deviceplugin.DevicePathType,
							Limit:       1,
						}},
					},
				},
			},
		},
	}
}

// Following DeviceSpec, Group, and Path are copied from Generic Device Plugin code and
// changed to look prettier when priting a configuration using 'show-config' command.
// Changes:
// - 'omitempty' tag was added to Group's Paths and USBSpecs
// - 'omitempty' tag was added to Path's Type
//
// Also USBSpec's Vendor and Product are strings instead of uint16 to avoid YAML unmarshalling problems
// (0451 is interpreted as octal instead of hexadecimal because of the leading 0 which led to USB devices being not exposed)

// DeviceSpec defines a device that should be discovered and scheduled.
// DeviceSpec allows multiple host devices to be selected and scheduled fungibly under the same name.
// Furthermore, host devices can be composed into groups of device nodes that should be scheduled
// as an atomic unit.
type DeviceSpec struct {
	// Name is a unique string representing the kind of device this specification describes.
	// +kubebuilder:default=serial
	Name string `json:"name"`
	// Groups is a list of groups of devices that should be scheduled under the same name.
	Groups []*Group `json:"groups"`
}

func (d DeviceSpec) toGDP() deviceplugin.DeviceSpec {
	groups := make([]*deviceplugin.Group, len(d.Groups))
	for i, group := range d.Groups {
		groups[i] = group.toGDP()
	}
	return deviceplugin.DeviceSpec{
		Name:   d.Name,
		Groups: groups,
	}
}

// Group represents a set of devices that should be grouped and mounted into a container together as one single meta-device.
type Group struct {
	// 'paths' is the list of devices of which the device group consists.
	// Paths can be globs, in which case each device matched by the path will be schedulable `Count` times.
	// When the paths have differing cardinalities, that is, the globs match different numbers of devices,
	// the cardinality of each path is capped at the lowest cardinality.
	// 'paths' is exclusive with 'usbs'.
	Paths []*Path `json:"paths,omitempty"`
	// 'usbs' is the list of USB specifications that this device group consists of.
	// A USB device must match exactly on all the given attributes to pass.
	// 'usbs' is exclusive with 'paths'.
	USBSpecs []*USBSpec `json:"usbs,omitempty"`
	// Count specifies how many times this group can be mounted concurrently.
	// When unspecified, Count defaults to 1.
	// +kubebuilder:default=1
	Count uint `json:"count,omitempty"`
}

func (g *Group) toGDP() *deviceplugin.Group {
	paths := make([]*deviceplugin.Path, 0, len(g.Paths))
	for _, path := range g.Paths {
		paths = append(paths, path.toGDP())
	}
	usbSpecs := make([]*deviceplugin.USBSpec, 0, len(g.USBSpecs))
	for _, usb := range g.USBSpecs {
		usbSpecs = append(usbSpecs, usb.toGDP())
	}
	return &deviceplugin.Group{
		Paths:    paths,
		USBSpecs: usbSpecs,
		Count:    g.Count,
	}
}

// Path represents a file path that should be discovered.
type Path struct {
	// Path is the file path of a device in the host.
	// +kubebuilder:default=/dev/ttyACM0
	Path string `json:"path"`
	// MountPath is the file path at which the host device should be mounted within the container.
	// When unspecified, MountPath defaults to the Path.
	// +kubebuilder:default=/dev/ttyACM0
	MountPath string `json:"mountPath,omitempty"`
	// Permissions is the file-system permissions given to the mounted device.
	// Permissions apply only to mounts of type `Device`.
	// This can be one or more of:
	// * r - allows the container to read from the specified device.
	// * w - allows the container to write to the specified device.
	// * m - allows the container to create device files that do not yet exist.
	// When unspecified, Permissions defaults to mrw.
	// +kubebuilder:default=mrw
	Permissions string `json:"permissions,omitempty"`
	// ReadOnly specifies whether the path should be mounted read-only.
	// ReadOnly applies only to mounts of type `Mount`.
	// +kubebuilder:default=false
	ReadOnly bool `json:"readOnly,omitempty"`
	// Type describes what type of file-system node this Path represents and thus how it should be mounted.
	// Allowed values: "Device", "Mount". When unspecified, type defaults to Device.
	// +kubebuilder:default=Device
	// +kubebuilder:validation:Enum:=Device;Mount;""
	Type deviceplugin.PathType `json:"type,omitempty"`
	// Limit specifies up to how many times this device can be used in the group concurrently when other devices
	// in the group yield more matches.
	// For example, if one path in the group matches 5 devices and another matches 1 device but has a limit of 10,
	// then the group will provide 5 pairs of devices.
	// When unspecified, Limit defaults to 1.
	// +kubebuilder:default=1
	Limit uint `json:"limit,omitempty"`
}

func (p *Path) toGDP() *deviceplugin.Path {
	return &deviceplugin.Path{
		Path:        p.Path,
		MountPath:   p.MountPath,
		Permissions: p.Permissions,
		ReadOnly:    p.ReadOnly,
		Type:        p.Type,
		Limit:       p.Limit,
	}
}

func (p *Path) validate() []error {
	errs := []error{}

	if p.Type != "" &&
		p.Type != deviceplugin.DevicePathType &&
		p.Type != deviceplugin.MountPathType {
		errs = append(errs, fmt.Errorf(
			"invalid type for path '%s': '%s' - accepted values are '' (will default to %s), %s, or %s",
			p.Path, p.Type, deviceplugin.DevicePathType, deviceplugin.DevicePathType, deviceplugin.MountPathType))
	}

	if p.ReadOnly && p.Type != deviceplugin.MountPathType {
		errs = append(errs, fmt.Errorf(
			"invalid readOnly value for path '%s': readOnly is only allowed for paths of type Mount",
			p.Path))
	}

	if p.Type == deviceplugin.MountPathType && p.Permissions != "" {
		errs = append(errs, fmt.Errorf(
			"invalid permissions for path '%s': permissions are not allowed for paths of type Mount",
			p.Path))
	}

	// Although Permissions are only for 'Device' type, we can always check them
	// and offer as many error messages as possible in case the type was mistyped.
	invalidChars := []string{}
	for _, c := range p.Permissions {
		if c != 'm' && c != 'r' && c != 'w' {
			invalidChars = append(invalidChars, string(c))
		}
	}
	if len(invalidChars) > 0 {
		errs = append(errs, fmt.Errorf(
			"invalid permissions for path %q: invalid characters '%s' in permissions string '%s' (only 'm', 'r', and 'w' are allowed)",
			p.Path, strings.Join(invalidChars, ", "), p.Permissions))
	}

	return errs
}

// USBSpec represents a USB device specification that should be discovered.
// A USB device must match exactly on all the given attributes to pass.
type USBSpec struct {
	// Vendor is the USB Vendor ID of the device to match on.
	// It should have the format "0x0123" or "0123".
	Vendor string `json:"vendor"`
	// Product is the USB Product ID of the device to match on.
	// It should have the format "0x0123" or "0123".
	Product string `json:"product"`
	// Serial is the serial number of the device to match on.
	// If specified, it must match exactly.
	Serial string `json:"serial"`
}

func (usb *USBSpec) toGDP() *deviceplugin.USBSpec {
	spec := &deviceplugin.USBSpec{
		Serial: usb.Serial,
	}
	// Ignoring errors because unmarshalling is checked in validate() during config validation.
	_ = spec.Vendor.UnmarshalJSON([]byte(strings.TrimPrefix(usb.Vendor, "0x")))
	_ = spec.Product.UnmarshalJSON([]byte(strings.TrimPrefix(usb.Product, "0x")))
	return spec
}

func (usb *USBSpec) validate() []error {
	errs := []error{}
	spec := &deviceplugin.USBSpec{}
	if err := spec.Vendor.UnmarshalJSON([]byte(strings.TrimPrefix(usb.Vendor, "0x"))); err != nil {
		errs = append(errs, fmt.Errorf("failed to parse USB Vendor ID %q: %v", usb.Vendor, err))
	}
	if err := spec.Product.UnmarshalJSON([]byte(strings.TrimPrefix(usb.Product, "0x"))); err != nil {
		errs = append(errs, fmt.Errorf("failed to parse USB Product ID %q: %v", usb.Product, err))
	}
	return errs
}
