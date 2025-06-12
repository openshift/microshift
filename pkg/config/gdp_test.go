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
			Devices: []deviceplugin.DeviceSpec{}, // empty devices
		}
		assert.NoError(t, cfg.validate())
	})

	t.Run("Devices cannot be empty", func(t *testing.T) {
		cfg := GenericDevicePlugin{
			Status:  "Enabled",
			Devices: []deviceplugin.DeviceSpec{},
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
				Devices: []deviceplugin.DeviceSpec{
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
					Devices: []deviceplugin.DeviceSpec{
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
			Devices: []deviceplugin.DeviceSpec{
				{Name: "serial",
					Groups: []*deviceplugin.Group{
						{
							Paths: []*deviceplugin.Path{
								{Path: "/dev/ttyUSB*"},
							},
							USBSpecs: []*deviceplugin.USBSpec{
								{Vendor: 1, Product: 1, Serial: "s"},
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
}
