package config

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

// CSIStorageDriver is an enum value that determines whether MicroShift deploys LVMS.
// +kubebuilder:validation:Enum:="";none;lvms
type CSIStorageDriver string

const (
	// CsiDriverUnset exists to support backwards compatibility with existing MicroShift clusters. When .storage.driver is
	// "", MicroShift will default to deploying LVMS. This preserves the current deployment behavior of existing
	// clusters.
	CsiDriverUnset CSIStorageDriver = ""
	//  CsiDriverNone signals MicroShift to not deploy the LVMS components. Setting the value for a cluster that has already
	//  deployed LVMS will not cause LVMS to be deleted. Otherwise, volumes already deployed on the cluster would be
	//  orphaned once their workloads stop or restart.
	CsiDriverNone CSIStorageDriver = "none"
	// CsiDriverLVMS is equivalent to CsiDriverUnset, and explicitly tells MicroShift to deploy LVMS. This option exists to
	// provide a differentiation between LVMS and potential future driver options.
	CsiDriverLVMS CSIStorageDriver = "lvms"
)

// OptionalCsiComponent values determine which CSI components MicroShift should deploy. Currently only csi snapshot components
// are supported.
// +kubebuilder:validation:Enum:=none;snapshot-controller;""
type OptionalCsiComponent string

const (
	// CsiComponentNone exists to support backwards compatibility with existing MicroShift clusters. By default,
	// MicroShift will deploy snapshot controller when no components are specified. This preserves the
	// current deployment behavior of existing clusters. Users must set .storage.with-csi-components: [ "none" ] to
	// explicitly tell MicroShift not to deploy any CSI components. The CSI Driver is excluded as it is typically
	// deployed via the same manifest as the accompanying storage driver. Like DriverOpt, uninstallation is not
	// supported as this can lead to orphaned storage volumes. Mutually exclusive with all other ComponentOpt values.
	CsiComponentNone OptionalCsiComponent = "none"
	// CsiComponentSnapshot causes MicroShift to deploy the CSI Snapshot controller.
	CsiComponentSnapshot OptionalCsiComponent = "snapshot-controller"
	// CsiComponentNullAlias is equivalent to not specifying a value. It exists because controller-gen generates
	// default empty-array values as [""], instead of []. Failing to include this odd value would mean the generated
	// /etc/microshift/config.default.yaml would break if passed to MicroShift.
	CsiComponentNullAlias OptionalCsiComponent = ""
)

// Storage represents a subfield of the MicroShift config data structure. Its purpose to provide a user
// facing interface to control whether MicroShift should deploy LVMS on startup.
type Storage struct {
	// Driver is a user defined string value matching one of the above CSIStorageDriver values. MicroShift uses this
	// value to decide whether to deploy the LVMS operator. An unset field defaults to "" during yaml parsing, and thus
	// could mean that the cluster has been upgraded. In order to support the existing out-of-box behavior, MicroShift
	// assumes an empty string to mean the storage driver should be deployed.
	// Allowed values are: unset or one of ["", "lvms", "none"]
	// +kubebuilder:validation:Optional
	Driver CSIStorageDriver `json:"driver,omitempty"`
	// OptionalCSIComponents is a user defined slice of CSIComponent values. These value tell MicroShift which
	// additional, non-driver, CSI controllers to deploy on start. MicroShift will deploy snapshot controller
	// when no components are specified. This preserves the current deployment behavior of existing
	// clusters. Users must set `.storage.optionalCsiComponents: []` to explicitly tell MicroShift not to deploy any CSI
	// components. The CSI Driver is excluded as it is typically deployed via the same manifest as the accompanying
	// storage driver. Like CSIStorageDriver, uninstallation is not supported as this can lead to orphaned storage
	// objects.
	// Allowed values are: unset, [], or one or more of ["snapshot-controller"]
	// +kubebuilder:validation:Optional
	// +kubebuilder:example={"snapshot-controller"}
	OptionalCSIComponents []OptionalCsiComponent `json:"optionalCsiComponents,omitempty"`
}

func (s Storage) driverIsValid() (isSupported bool) {
	return sets.New[CSIStorageDriver](CsiDriverNone, CsiDriverLVMS, CsiDriverUnset).Has(s.Driver)
}

func (s Storage) csiComponentsAreValid() []string {
	supported := sets.New[OptionalCsiComponent](CsiComponentSnapshot, CsiComponentNone,
		CsiComponentNullAlias)
	unsupported := sets.New[string]()

	for _, cfgComp := range s.OptionalCSIComponents {
		if !supported.Has(cfgComp) {
			unsupported.Insert(string(cfgComp))
		}
	}
	return unsupported.UnsortedList()
}

func (s Storage) noneIsMutuallyExclusive() ([]string, bool) {
	components := sets.New[OptionalCsiComponent](s.OptionalCSIComponents...)
	if components.Has(CsiComponentNone) && components.Len() > 1 {
		components.Delete(CsiComponentNone)
		ret := sets.New[string]()
		for v := range components {
			ret.Insert(string(v))
		}
		return ret.UnsortedList(), false
	}
	return nil, true
}

// IsValid checks all sub-fields of a Storage object. The data is considered valid when
// both .driver and .csi-components are either empty strings or legitimate values. For
// .csi-components, the string "none" is mutually exclusive with all other values, valid or not.
func (s Storage) IsValid() []error {
	errs := sets.New[error]()
	if !s.driverIsValid() {
		errs.Insert(fmt.Errorf("invalid driver %q", s.Driver))
	}
	if comps := s.csiComponentsAreValid(); len(comps) > 0 {
		errs.Insert(fmt.Errorf("invalid CSI components: %v", comps))
	}
	return errs.UnsortedList()
}

// IsEnabled returns false only when .storage.driver: "none". An empty value is considered "enabled"
// for backwards compatibility. Otherwise, the meaning of the config would silently change after an
// upgrade from enabled-by-default to disabled-by-default.
func (s Storage) IsEnabled() bool {
	return s.Driver != CsiDriverNone
}
