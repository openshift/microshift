package config

import "fmt"

const (
	FeatureSetCustomNoUpgrade      = "CustomNoUpgrade"
	FeatureSetTechPreviewNoUpgrade = "TechPreviewNoUpgrade"
	FeatureSetDevPreviewNoUpgrade  = "DevPreviewNoUpgrade"
)

type FeatureGates struct {
	FeatureSet      string          `json:"featureSet"`
	CustomNoUpgrade CustomNoUpgrade `json:"customNoUpgrade"`
}

type CustomNoUpgrade struct {
	Enabled  []string `json:"enabled"`
	Disabled []string `json:"disabled"`
}

func validateFeatureGates(fg *FeatureGates) error {
	if fg == nil {
		return nil
	}
	if err := validateFeatureSet(fg); err != nil {
		return err
	}
	if err := validateCustomNoUpgrade(fg); err != nil {
		return err
	}
	return nil
}

func validateFeatureSets(fg *FeatureGates) error {
	if fg == nil {
		return nil
	}
	// Must use a recognized feature set, or else empty
	if fg.FeatureSet != FeatureSetCustomNoUpgrade && fg.FeatureSet != FeatureSetTechPreviewNoUpgrade && fg.FeatureSet != FeatureSetDevPreviewNoUpgrade {
		return fmt.Errorf("invalid feature set: %s", fg.FeatureSet)
	}
	// Must set FeatureSet to CustomNoUpgrade to use custom feature gates
	if fg.FeatureSet != FeatureSetCustomNoUpgrade && (len(fg.CustomNoUpgrade.Enabled) > 0 || len(fg.CustomNoUpgrade.Disabled) > 0) {
		return fmt.Errorf("CustomNoUpgrade must be empty when FeatureSet is empty")
	}
	// Must not use custom feature gates when FeatureSet is not CustomNoUpgrade
	if fg.FeatureSet != FeatureSetCustomNoUpgrade && (len(fg.CustomNoUpgrade.Enabled) > 0 || len(fg.CustomNoUpgrade.Disabled) > 0) {
		return fmt.Errorf("CustomNoUpgrade must be empty when FeatureSet is empty")
	}
	return nil
}

func validateCustomNoUpgrade(fg *FeatureGates) error {
	if fg == nil {
		return nil
	}
	return nil
}
