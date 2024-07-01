package components

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/openshift/microshift/pkg/config/lvmd"
)

// runtimeCfgUpdateDelay is the delay to wait for the runtime config to be updated via the watch
// in the tests. This is a bit of a hack to avoid having to wait for the watch to trigger.
// In case of failures on loads, consider increasing this value.
// In a normal I/O environment, this value should never be a problem.
var runtimeCfgUpdateDelay = 50 * time.Millisecond

func Test_loadCSIPluginConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testUsrCfg := filepath.Join(t.TempDir(), "lvmd-usr.yaml")
	testRuntimeCfg := filepath.Join(t.TempDir(), "lvmd-runtime.yaml")
	testDefaultCfgLvmd := &lvmd.Lvmd{
		DeviceClasses: []*lvmd.DeviceClass{
			{
				Name:        "ssd",
				VolumeGroup: "imaginary-vg",
				Default:     true,
			},
		},
		SocketName: "the socket should not matter!",
	}

	getRuntimeCfgAfterDelay := func() *lvmd.Lvmd {
		// wait for the runtime config to be updated via the watch
		time.Sleep(runtimeCfgUpdateDelay)
		runtimeLvmd, err := lvmd.NewLvmdConfigFromFile(testRuntimeCfg)
		if err != nil {
			t.Fatalf("Failed to load user config: %v", err)
		}
		runtimeLvmd.Message = "" // ignore message field
		return runtimeLvmd
	}

	previousCfgLoader := defaultCfgLoader
	defaultCfgLoader = func() (*lvmd.Lvmd, error) {
		return testDefaultCfgLvmd, nil
	}
	t.Cleanup(func() {
		defaultCfgLoader = previousCfgLoader
	})

	var runtimeLvmd *lvmd.Lvmd
	t.Run("default config loading and runtime file init should work when no files exist", func(t *testing.T) {
		cfg, err := loadCSIPluginConfig(ctx, testUsrCfg, testRuntimeCfg)
		if err != nil {
			t.Fatalf("Failed to load csi plugin config: %v", err)
		}

		if !cfg.IsEnabled() {
			t.Fatalf("Expected csi plugin to be enabled")
		}
		if len(cfg.SocketName) == 0 {
			t.Fatalf("Expected socket name to be set")
		}

		if _, err := os.Stat(testRuntimeCfg); err != nil {
			t.Fatalf("Failed to open runtime config file: %v", err)
		}

		runtimeLvmd = getRuntimeCfgAfterDelay()

		if !reflect.DeepEqual(runtimeLvmd, testDefaultCfgLvmd) {
			t.Fatalf("Expected runtime config to match default config")
		}
	})

	var usrLvmd *lvmd.Lvmd
	t.Run("user config loading", func(t *testing.T) {
		usrLvmd = &lvmd.Lvmd{
			DeviceClasses: []*lvmd.DeviceClass{
				{
					Name:        "ssd",
					VolumeGroup: "imaginary-vg",
					Default:     true,
				},
				{
					Name:        "hdd",
					VolumeGroup: "user-vg",
					Default:     false,
				},
			},
			SocketName: "the socket should not matter!",
		}

		if err := lvmd.SaveLvmdConfigToFile(usrLvmd, testUsrCfg); err != nil {
			t.Fatalf("Failed to save user config: %v", err)
		}

		t.Run("chmod to write only should not affect the watch", func(t *testing.T) {
			// wait for the runtime config to be updated via the watch but it should not trigger anything
			if err := os.Chmod(testUsrCfg, 0400); err != nil {
				t.Fatalf("Failed to change user config permissions: %v", err)
			}

			runtimeLvmd = getRuntimeCfgAfterDelay()
			if !reflect.DeepEqual(runtimeLvmd, usrLvmd) {
				t.Fatalf("Expected runtime config to match user config")
			}

			if err := lvmd.SaveLvmdConfigToFile(usrLvmd, testUsrCfg); err == nil {
				t.Fatalf("expected usr config to have been updated to Readonly from Chmod (400) and not be writable: %v", err)
			}

			// wait for the runtime config to be updated via the watch but it should not trigger anything
			if err := os.Chmod(testUsrCfg, 0600); err != nil {
				t.Fatalf("Failed to change user config permissions: %v", err)
			}
		})

		t.Run("user config update should trigger a reload", func(t *testing.T) {
			// switch the defaults to trigger a reload
			usrLvmd.DeviceClasses[1].Default = true
			usrLvmd.DeviceClasses[0].Default = false

			if err := lvmd.SaveLvmdConfigToFile(usrLvmd, testUsrCfg); err != nil {
				t.Fatalf("Failed to save user config: %v", err)
			}

			runtimeLvmd = getRuntimeCfgAfterDelay()
			if !reflect.DeepEqual(runtimeLvmd, usrLvmd) {
				t.Fatalf("Expected runtime config to match user config")
			}
		})
	})

	t.Run("user config deletion should trigger a reload to defaults", func(t *testing.T) {
		if err := os.Remove(testUsrCfg); err != nil {
			t.Fatalf("Failed to remove user config: %v", err)
		}
		runtimeLvmd = getRuntimeCfgAfterDelay()
		if reflect.DeepEqual(runtimeLvmd, usrLvmd) {
			t.Fatalf("Expected runtime config to now match default config again, but was equal to old usr config")
		}
		if !reflect.DeepEqual(runtimeLvmd, testDefaultCfgLvmd) {
			t.Fatalf("Expected runtime config to now match default config again, but was not equal to the default cfg")
		}
	})

	t.Run("user config watch should be stopped after shutdown", func(t *testing.T) {
		cancel()

		usrLvmd.SocketName = "new-socket-name-after-shutdown"
		if err := lvmd.SaveLvmdConfigToFile(usrLvmd, testUsrCfg); err != nil {
			t.Fatalf("Failed to save user config after watch shutdown: %v", err)
		}

		runtimeLvmd = getRuntimeCfgAfterDelay()
		if reflect.DeepEqual(runtimeLvmd, usrLvmd) {
			t.Fatalf("Expected runtime config to not have been updated after watch shutdown, but was equal to new usr config")
		}
	})
}
