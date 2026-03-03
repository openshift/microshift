// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"sigs.k8s.io/yaml"
)

// EnforcementPolicy is a representation of how validations
// should be enforced.
type EnforcementPolicy string

const (
	// EnforcementPolicyError is used to represent that a validation
	// should report an error when identifying incompatible changes.
	EnforcementPolicyError EnforcementPolicy = "Error"

	// EnforcementPolicyWarn is used to represent that a validation
	// should report a warning when identifying incompatible changes.
	EnforcementPolicyWarn EnforcementPolicy = "Warn"

	// EnforcementPolicyNone is used to represent that a validation
	// should not report anything when identifying incompatible changes.
	EnforcementPolicyNone EnforcementPolicy = "None"
)

// ConversionPolicy is a representation of how the served version
// validator should react when a CRD specifies a conversion
// strategy.
type ConversionPolicy string

const (
	// ConversionPolicyNone is used to represent that the served
	// version validator should not treat CRDs with a conversion strategy
	// differently and should run all the validations on detected changes
	// across served versions of the CRD.
	ConversionPolicyNone ConversionPolicy = "None"

	// ConversionPolicyIgnore is used to represent that the served
	// version validator should treat CRDs with a conversion strategy
	// specified as a valid reason to skip running the validations
	// on changes across served versions of the CRD.
	ConversionPolicyIgnore ConversionPolicy = "Ignore"
)

// Config is the configuration used for dictating how validations
// and validators should be configured.
type Config struct {
	// validations is an optional field used to configure the set of
	// validations that should be run during comparisons.
	//
	// Configuration of validations is strictly additive.
	// Default behaviors of validations will be used in the
	// event they are not included in the set of configured validations.
	Validations []ValidationConfig `json:"validations"`

	// unhandledEnforcement is an optional field used to configure
	// how changes that have not been handled by an existing validation
	// should be treated.
	//
	// Allowed values are Error, Warn, and None.
	//
	// When set to Error, any unhandled changes will result in an error.
	//
	// When set to Warn, any unhandled changes will result in a warning.
	//
	// When set to None, unhandled changes will be ignored.
	// Defaults to Error.
	UnhandledEnforcement EnforcementPolicy `json:"unhandledEnforcement"`

	// conversion is an optional field used to configure how validations
	// are run against served versions of the CRD.
	//
	// Allowed values are None and Ignore.
	//
	// When set to Ignore, if a conversion strategy of "Webhook" is specified served
	// versions will not be validated.
	//
	// When set to None, even if a conversion strategy of "Webhook" is specified served
	// versions will be validated.
	// Defaults to None.
	Conversion ConversionPolicy `json:"conversion"`
}

// ValidationConfig is used to dictate how individual validations
// should be configured.
type ValidationConfig struct {
	// name is a required field used to specify a validation.
	// name must be a known validation.
	Name string `json:"name"`

	// enforcement is a required field used to specify how a validation
	// should be enforced.
	//
	// Allowed values are Error, Warn, and None.
	//
	// When set to Error, any incompatibilities found by the validation
	// will result in an error message.
	//
	// When set to Warn, any incompatibilities found by the validation
	// will result in a warning message.
	//
	// When set to None, the validation will not be run.
	Enforcement EnforcementPolicy `json:"enforcement"`

	// configuration is an optional field used to specify a configuration
	// for the validation.
	Configuration map[string]interface{} `json:"configuration"`
}

// Load reads a file into a Config object and validates it.
// If there are any errors encountered while loading the file contents
// into the Config object a nil value and error will be returned.
// Otherwise, a pointer to the Config object will be returned alongside a nil error.
func Load(configFile string) (*Config, error) {
	cfg := &Config{}

	if configFile != "" {
		//nolint:gosec
		file, err := os.Open(configFile)
		if err != nil {
			return nil, fmt.Errorf("loading config file %q: %w", configFile, err)
		}

		configBytes, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("reading config file %q: %w", configFile, err)
		}

		err = file.Close()
		if err != nil {
			log.Print("failed to close config file after reading", configFile, err)
		}

		err = yaml.Unmarshal(configBytes, cfg)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling config file %q contents: %w", configFile, err)
		}
	}

	err := ValidateConfig(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateConfig ensures a valid Config object.
// It will set defaults where appropriate and return an error
// if user specified values are invalid.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		// nothing to validate
		return nil
	}

	validationErr := ValidateValidations(cfg.Validations...)
	unhandledEnforcementErr := ValidateEnforcementPolicy(&cfg.UnhandledEnforcement, false)
	conversionErr := ValidateConversionPolicy(&cfg.Conversion)

	return errors.Join(validationErr, unhandledEnforcementErr, conversionErr)
}

// ValidateConversionPolicy ensures the provided ConversionPolicy
// is valid.
// It will modify the ConversionPolicy to set it to the
// default value of "None" if it is the empty string ("").
// Returns an error if an invalid ConversionPolicy is specified.
func ValidateConversionPolicy(policy *ConversionPolicy) error {
	if policy == nil {
		// nothing to validate
		return nil
	}

	var err error

	switch *policy {
	case ConversionPolicyNone, ConversionPolicyIgnore:
		// do nothing, valid values
	case ConversionPolicy(""):
		// default to None
		*policy = ConversionPolicyNone
	default:
		err = fmt.Errorf("%w: %q", errUnknownConversionPolicy, *policy)
	}

	return err
}

var errUnknownConversionPolicy = errors.New("unknown conversion policy")

// ValidateEnforcementPolicy ensures the provided EnforcementPolicy
// is valid.
// It will modify the EnforcementPolicy to set it to the
// default value of "Error" if it is the empty string ("") and the EnforcementPolicy
// is not required.
// Returns an error if an invalid EnforcementPolicy is specified.
func ValidateEnforcementPolicy(policy *EnforcementPolicy, required bool) error {
	if policy == nil {
		// nothing to validate
		return nil
	}

	var err error

	switch *policy {
	case EnforcementPolicyError, EnforcementPolicyWarn, EnforcementPolicyNone:
		// do nothing, valid values
	case EnforcementPolicy(""):
		if required {
			err = errEnforcementRequired
			break
		}
		// default to error
		*policy = EnforcementPolicyError
	default:
		err = fmt.Errorf("%w: %q", errUnknownEnforcement, string(*policy))
	}

	return err
}

var (
	errEnforcementRequired = errors.New("enforcement is required")
	errUnknownEnforcement  = errors.New("unknown enforcement")
)

// ValidateValidations loops through the provided ValidationConfig
// items to ensure they are valid.
// Returns an aggregated error of the invalid ValidationConfig items.
func ValidateValidations(validations ...ValidationConfig) error {
	errs := []error{}

	for i, validation := range validations {
		if validation.Name == "" {
			errs = append(errs, fmt.Errorf("validations[%d] is invalid: %w", i, errNameRequired))
		}

		errs = append(errs, ValidateEnforcementPolicy(&validation.Enforcement, true))
	}

	return errors.Join(errs...)
}

var errNameRequired = errors.New("name is required")
