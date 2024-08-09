package healthcheck

import (
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

const (
	microshiftWaitTimeoutSecVar = "MICROSHIFT_WAIT_TIMEOUT_SEC"
	greenbootMaxBootAttemptsVar = "GREENBOOT_MAX_BOOT_ATTEMPTS"
	bootCounterVar              = "boot_counter"

	greenbootConfFilepath = "/etc/greenboot/greenboot.conf"
	grubEnvs              = "/boot/grub2/grubenv"

	defaultMicroshiftWaitTimeoutSec = 300
	defaultGreenbootMaxBootAttempts = 3
)

// getGreenbootTimeoutDuration obtains the max duration of the healthchecks
// based on greenboot's setting and GRUB's env variables.
func getGreenbootTimeoutDuration() time.Duration {
	timeout := time.Duration(getGreenbootTimeout()) * time.Second
	klog.Infof("Calculated timeout for the healthchecks: %v", timeout)
	return timeout
}

// getGreenbootTimeout calculates the timeout in seconds for the healthchecks.
func getGreenbootTimeout() int {
	greenbootConfig := readGreenbootConfig()
	if len(greenbootConfig) > 0 {
		klog.Infof("Greenboot variables: %v", greenbootConfig)
	}

	grubVars := readGrubVars()
	if len(grubVars) > 0 {
		klog.Infof("GRUB boot variables: %v", grubVars)
	}

	configuredTimeout := getIntValueFromMap(greenbootConfig, microshiftWaitTimeoutSecVar, defaultMicroshiftWaitTimeoutSec)
	if configuredTimeout < 60 {
		klog.Infof("Configured %q is less than 60 - using 60", microshiftWaitTimeoutSecVar)
		configuredTimeout = 60
	}
	maxBoots := getIntValueFromMap(greenbootConfig, greenbootMaxBootAttemptsVar, defaultGreenbootMaxBootAttempts)
	bootCounter := getIntValueFromMap(grubVars, bootCounterVar, maxBoots-1)

	return calculateTimeout(configuredTimeout, maxBoots, bootCounter)
}

func readGreenbootConfig() map[string]string {
	return readBashlikeConfigFile(greenbootConfFilepath, "GREENBOOT_", "MICROSHIFT_")
}

func readGrubVars() map[string]string {
	return readBashlikeConfigFile(grubEnvs, "boot_")
}

func calculateTimeout(configuredTimeout, maxBoots, bootCounter int) int {
	result := configuredTimeout * (maxBoots - bootCounter)
	klog.V(2).Infof("Calculated timeout based on '%s * (%s - %s)': %d * (%d - %d) = %d",
		microshiftWaitTimeoutSecVar, greenbootMaxBootAttemptsVar, bootCounterVar,
		configuredTimeout, maxBoots, bootCounter, result)
	return result
}

func readBashlikeConfigFile(path string, prefixes ...string) map[string]string {
	dat, err := os.ReadFile(path)
	if err != nil {
		klog.Errorf("Error: %v", err)
		return nil
	}
	return bashlikeVarsToMap(strings.Split(string(dat), "\n"), prefixes...)
}

func bashlikeVarsToMap(lines []string, prefixes ...string) map[string]string {
	m := make(map[string]string)

	for _, line := range lines {
		for _, prefix := range prefixes {
			if strings.HasPrefix(line, prefix) {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "#") {
					continue
				}

				fields := strings.Split(trimmed, "=")
				if len(fields) != 2 {
					klog.Error(nil, "Unexpected entry conf file: %q", line)
					continue
				}

				key := fields[0]
				value := fields[1]
				if strings.Contains(value, "#") {
					value_fields := strings.Split(value, "#")
					value = strings.TrimSpace(value_fields[0])
				}

				m[key] = value
			}
		}
	}
	return m
}

func getIntValueFromMap(m map[string]string, key string, def int) int {
	valueStr, ok := m[key]
	if !ok {
		klog.V(2).Infof("Configuration doesn't contain key %q - using default value: %d", key, def)
		return def
	}

	klog.V(2).Infof("Found value for %q: %q", key, valueStr)
	if valueStr == "" {
		klog.V(2).Infof("Using default value for %q: %d", key, def)
		return def
	}
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		klog.Errorf("Failed to parse string to int - using default: %d. Error: %v", def, err)
		return def
	}

	return val
}
