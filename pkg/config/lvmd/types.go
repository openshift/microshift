// DeviceClass type definition and validation is taken from:
// https://github.com/red-hat-storage/topolvm/blob/release-4.10/lvmd/device_class_manager.go

package lvmd

import "regexp"

type DeviceType string

const (
	defaultSpareGB = 10
	TypeThin       = DeviceType("thin")
	TypeThick      = DeviceType("thick")
)

// This regexp is based on the following validation:
//
//	https://github.com/kubernetes/apimachinery/blob/v0.18.3/pkg/util/validation/validation.go#L42
var qualifiedNameRegexp = regexp.MustCompile("^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$")

// This regexp is used to check StripeSize format
var stripeSizeRegexp = regexp.MustCompile("(?i)^([0-9]*)(k|m|g|t|p|e|b|s)?$")

// ThinPoolConfig holds the configuration of thin pool in a volume group
type ThinPoolConfig struct {
	// Name of thinpool
	Name string `json:"name"`
	// OverprovisionRatio signifies the upper bound multiplier for allowing logical volume creation in this pool
	OverprovisionRatio float64 `json:"overprovision-ratio"`
}

// DeviceClass maps between device-classes and target for logical volume creation
// current targets are VolumeGroup for thick-lv and ThinPool for thin-lv
type DeviceClass struct {
	// Name for the device-class name
	Name string `json:"name"`
	// Volume group name for the device-class
	VolumeGroup string `json:"volume-group"`
	// Default is a flag to indicate whether the device-class is the default
	Default bool `json:"default"`
	// SpareGB is storage capacity in GiB to be spared
	SpareGB *uint64 `json:"spare-gb"`
	// Stripe is the number of stripes in the logical volume
	Stripe *uint `json:"stripe"`
	// StripeSize is the amount of data that is written to one device before moving to the next device
	StripeSize string `json:"stripe-size"`
	// LVCreateOptions are extra arguments to pass to lvcreate
	LVCreateOptions []string `json:"lvcreate-options"`
	// Type is the name of logical volume target, supports 'thick' (default) or 'thin' currently
	Type DeviceType `json:"type"`
	// ThinPoolConfig holds the configuration for thinpool in this volume group corresponding to the device-class
	ThinPoolConfig *ThinPoolConfig `json:"thin-pool"`
}
