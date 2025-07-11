package config

import (
	"fmt"
	"testing"

	"github.com/squat/generic-device-plugin/deviceplugin"
	"github.com/stretchr/testify/assert"
)

func Test_GDP_Validate(t *testing.T) {
	t.Run("Disabled feature means no validation", func(t *testing.T) {
		cfg := GenericDevicePlugin{
			Status:  "Disabled",
			Domain:  "invalid_domain_bacause_of_the_underscores",
			Devices: []DeviceSpec{}, // empty devices
		}
		assert.NoError(t, cfg.validate())
	})

	t.Run("Devices cannot be empty", func(t *testing.T) {
		cfg := GenericDevicePlugin{
			Status:  "Enabled",
			Devices: []DeviceSpec{},
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "at least one device must be specified")
	})

	t.Run("Status other than Enabled or Disabled is invalid", func(t *testing.T) {
		cfg := GenericDevicePlugin{
			Status: "managed",
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Status must be either 'Enabled' or 'Disabled'")
	})

	t.Run("device name mustn't start or end with -", func(t *testing.T) {
		for _, dn := range []string{"-ok", "ok-"} {
			cfg := GenericDevicePlugin{
				Status: "Enabled",
				Domain: "valid-domain.io",
				Devices: []DeviceSpec{
					{Name: dn},
				},
			}
			err := cfg.validate()
			assert.Error(t, err)
			assert.ErrorContains(t, err, "device name must match the regular expression")
		}
	})

	t.Run("device name cannot include special chars beside -", func(t *testing.T) {
		invalidChars := `! @#$%^&*()_+=[]{};'\:"|<>?,./~` + "`"
		for _, char := range invalidChars {
			t.Run(fmt.Sprintf("device name with %q is invalid", string(char)), func(t *testing.T) {
				cfg := GenericDevicePlugin{
					Status: "Enabled",
					Domain: "valid-domain.io",
					Devices: []DeviceSpec{
						{Name: "ok" + string(char) + "ok"},
					},
				}
				err := cfg.validate()
				assert.Error(t, err)
				assert.ErrorContains(t, err, "device name must match the regular expression")
			})
		}
	})

	t.Run("device cannot specify both path and usb", func(t *testing.T) {
		cfg := GenericDevicePlugin{
			Status: "Enabled",
			Domain: "valid-domain.io",
			Devices: []DeviceSpec{
				{Name: "serial",
					Groups: []*Group{
						{
							Paths: []*Path{
								{Path: "/dev/ttyUSB*"},
							},
							USBSpecs: []*USBSpec{
								{Vendor: "1", Product: "1", Serial: "s"},
							},
						},
					},
				},
			},
		}
		err := cfg.validate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "cannot define both path and usb at the same time")
	})

	t.Run("empty usb is ignored and removed during validation", func(t *testing.T) {
		cfg := GenericDevicePlugin{
			Status: "Enabled",
			Domain: "valid-domain.io",
			Devices: []DeviceSpec{
				{Name: "serial",
					Groups: []*Group{
						{
							Paths: []*Path{
								{Path: "/dev/ttyUSB*"},
							},
							USBSpecs: []*USBSpec{
								{Vendor: "", Product: "", Serial: ""},
							},
						},
					},
				},
			},
		}
		err := cfg.validate()
		assert.NoError(t, err)

		// validate() also removes empty usb specs
		assert.Equal(t, 1, len(cfg.Devices[0].Groups[0].Paths))
		assert.Equal(t, "/dev/ttyUSB*", cfg.Devices[0].Groups[0].Paths[0].Path)
		assert.Equal(t, 0, len(cfg.Devices[0].Groups[0].USBSpecs))
	})
}

func Test_GDP_Path_Validate(t *testing.T) {
	t.Run("lower case type is invalid", func(t *testing.T) {
		for _, invalidType := range []string{"device", "mount"} {
			path := &Path{
				Type: deviceplugin.PathType(invalidType),
			}
			errs := path.validate()
			assert.Error(t, errs[0])
			assert.ErrorContains(t, errs[0], "invalid type for path")
		}
	})

	t.Run("readOnly is only allowed for paths of type Mount", func(t *testing.T) {
		path := &Path{
			Type:     deviceplugin.DevicePathType,
			ReadOnly: true,
		}
		errs := path.validate()
		assert.Error(t, errs[0])
		assert.ErrorContains(t, errs[0], "readOnly is only allowed for paths of type Mount")
	})

	t.Run("permissions are not allowed for paths of type Mount", func(t *testing.T) {
		path := &Path{
			Type:        deviceplugin.MountPathType,
			Permissions: "mrw",
		}
		errs := path.validate()
		assert.Error(t, errs[0])
		assert.ErrorContains(t, errs[0], "permissions are not allowed for paths of type Mount")
	})

	t.Run("only mrw permissions are allowed", func(t *testing.T) {
		path := &Path{
			Type:        deviceplugin.DevicePathType,
			Permissions: "qetyuiopasdfghjklzxcvbn",
		}
		errs := path.validate()
		assert.Error(t, errs[0])
		assert.ErrorContains(t, errs[0], "invalid characters 'q, e, t, y, u, i, o, p, a, s, d, f, g, h, j, k, l, z, x, c, v, b, n'")
	})

	t.Run("empty permissions are ok", func(t *testing.T) {
		path := &Path{
			Type:        deviceplugin.DevicePathType,
			Permissions: "",
		}
		errs := path.validate()
		assert.Len(t, errs, 0)
	})
}

func Test_USBSpec_toGDP(t *testing.T) {
	t.Run("hexadecimals are converted properly", func(t *testing.T) {
		spec := USBSpec{
			Vendor:  "0x0451",
			Product: "0x16a8",
		}
		gdp := spec.toGDP()
		assert.Equal(t, deviceplugin.USBID(0x0451), gdp.Vendor)
		assert.Equal(t, deviceplugin.USBID(0x16a8), gdp.Product)
	})

	t.Run("hexadecimals without leading 0x are converted properly", func(t *testing.T) {
		spec := USBSpec{
			Vendor:  "0451",
			Product: "16a8",
		}
		gdp := spec.toGDP()
		assert.Equal(t, deviceplugin.USBID(0x0451), gdp.Vendor)
		assert.Equal(t, deviceplugin.USBID(0x16a8), gdp.Product)
	})

	t.Run("decimals are converted properly", func(t *testing.T) {
		spec := USBSpec{
			Vendor:  "451",
			Product: "781",
		}
		gdp := spec.toGDP()
		assert.Equal(t, deviceplugin.USBID(0x0451), gdp.Vendor)
		assert.Equal(t, deviceplugin.USBID(0x0781), gdp.Product)
	})

	t.Run("decimals with leading zeros are converted properly", func(t *testing.T) {
		spec := USBSpec{
			Vendor:  "0451",
			Product: "0781",
		}
		gdp := spec.toGDP()
		assert.Equal(t, deviceplugin.USBID(0x0451), gdp.Vendor)
		assert.Equal(t, deviceplugin.USBID(0x0781), gdp.Product)
	})

	t.Run("serial number is copied as-is", func(t *testing.T) {
		{
			spec := USBSpec{Serial: "s"}
			gdp := spec.toGDP()
			assert.Equal(t, "s", gdp.Serial)
		}
		{
			spec := USBSpec{Serial: ""}
			gdp := spec.toGDP()
			assert.Equal(t, "", gdp.Serial)
		}
	})
}
