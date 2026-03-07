package crdify

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/spf13/afero"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/loaders/composite"
	"sigs.k8s.io/crdify/pkg/loaders/file"
	"sigs.k8s.io/crdify/pkg/loaders/git"
	"sigs.k8s.io/crdify/pkg/loaders/scheme"
	"sigs.k8s.io/crdify/pkg/runner"
	"sigs.k8s.io/crdify/pkg/validations"
	kyaml "sigs.k8s.io/yaml"
)

type generatorOption func(*generator)

// WithComparisonBase configures the git base used by the crdify generator
// to load the "old" CRD.
func WithComparisonBase(base string) generatorOption {
	return func(g *generator) {
		g.comparisonBase = base
	}
}

// WithConfig configures the crdify configuration used by the crdify generator.
func WithConfig(cfg *config.Config) generatorOption {
	return func(g *generator) {
		g.cfg = cfg
	}
}

// WithValidationRegistry configures the validation registry used by the crdify generator.
func WithValidationRegistry(registry validations.Registry) generatorOption {
	return func(g *generator) {
		g.validationRegistry = registry
	}
}

// WithDisabled configures whether or not the crdify generator is disabled.
func WithDisabled(disabled bool) generatorOption {
	return func(g *generator) {
		g.disabled = disabled
	}
}

// generator implements the generation.Generator interface.
// It is designed to verify the CRD schema updates for a particular API group.
type generator struct {
	disabled           bool
	comparisonBase     string
	cfg                *config.Config
	validationRegistry validations.Registry
}

// NewGenerator returns a new generation.Generator
// that runs crdify validations.
// Optionally, callers can provide configuration
// options to change the default behavior of the generator.
func NewGenerator(opts ...generatorOption) generation.Generator {
	defaultGenerator := &generator{
		comparisonBase: "master",
		cfg: &config.Config{
			// Do not fail on currently unhandled changes.
			// This matches the existing usage for crd-schema-checker
			// where it will only fail if it identifies an explicitly
			// failed validation.
			// This is less strict than the default behavior of crdify,
			// but also will result in less noise.
			// We have confidence in our API reviewers to appropriately
			// catch breaking changes that there are not validations for
			// yet, and to check the presubmit jobs that run
			// this generator for any warnings that may be present.
			UnhandledEnforcement: config.EnforcementPolicyWarn,
			Conversion:           config.ConversionPolicyNone,
			Validations: []config.ValidationConfig{
				// Allow addition of new enums
				{
					Name:        "enum",
					Enforcement: config.EnforcementPolicyError,
					Configuration: map[string]interface{}{
						"additionPolicy": "Allow",
					},
				},
				// Ignore any incompatibility detection
				// for descriptions changes.
				{
					Name:        "description",
					Enforcement: config.EnforcementPolicyNone,
				},
			},
		},
		validationRegistry: runner.DefaultRegistry(),
	}

	for _, opt := range opts {
		opt(defaultGenerator)
	}

	return defaultGenerator
}

// Name returns the name of the generator.
func (g *generator) Name() string {
	return "crdify"
}

// ApplyConfig creates returns a new generator based on the configuration passed.
// If the crdify configuration is empty, the existing generator is returned.
func (g *generator) ApplyConfig(cfg *generation.Config) generation.Generator {
	if cfg == nil || cfg.Crdify == nil {
		return g
	}

	return NewGenerator(
		WithDisabled(cfg.Crdify.Disabled),
		WithConfig(cfg.Crdify.Config),
		WithComparisonBase(g.comparisonBase),
	)
}

// GenGroup runs the crdify generator against the given group context.
func (g *generator) GenGroup(groupCtx generation.APIGroupContext) ([]generation.Result, error) {
	if g.disabled {
		klog.V(2).Infof("Skipping crdify check for %s", groupCtx.Name)
		return nil, nil
	}

	errs := []error{}
	results := []generation.Result{}

	for _, version := range groupCtx.Versions {
		klog.V(1).Infof("Verifying API schema for %s/%s", groupCtx.Name, version.Name)

		r, err := g.genGroupVersion(groupCtx.Name, version)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not run crdify generator for group/version %s/%s: %w", groupCtx.Name, version.Name, err))
		}

		results = append(results, r...)
	}

	if len(errs) > 0 {
		return results, kerrors.NewAggregate(errs)
	}

	return results, nil
}

// genGroupVersion runs the crdify generator against a particular version of the API group.
func (g *generator) genGroupVersion(name string, version generation.APIVersionContext) ([]generation.Result, error) {
	crdPaths, err := getCRDPathsForVersion(version)
	if err != nil {
		return nil, fmt.Errorf("getting CRD paths: %w", err)
	}

	results := []generation.Result{}
	errs := []error{}

	loader := composite.NewComposite(
		map[string]composite.Loader{
			scheme.SchemeFile: file.New(afero.OsFs{}),
			scheme.SchemeGit:  git.New(),
		},
	)

	for _, crdPath := range crdPaths {
		oldCrd, err := loader.Load(context.TODO(), fmt.Sprintf("git://%s?path=%s", g.comparisonBase, crdPath))
		if err != nil {
			if errors.Is(err, gitobject.ErrFileNotFound) {
				// No previous CRD existed for this path, ignore this path.
				continue
			}

			errs = append(errs, fmt.Errorf("loading old CRD: %w", err))
			continue
		}

		newCrd, err := loader.Load(context.TODO(), fmt.Sprintf("file://%s", crdPath))
		if err != nil {
			errs = append(errs, fmt.Errorf("loading new CRD: %w", err))
			continue
		}

		run, err := runner.New(g.cfg, g.validationRegistry)
		if err != nil {
			errs = append(errs, fmt.Errorf("creating validation runner: %w", err))
		}

		runResults := run.Run(oldCrd, newCrd)

		result := generation.Result{
			Generator: g.Name(),
			Group:     name,
			Version:   version.Name,
			Manifest:  crdPath,
		}

		for _, crdResults := range runResults.CRDValidation {
			for _, err := range crdResults.Errors {
				resErr := fmt.Errorf("%s: %w", crdResults.Name, errors.New(err))
				result.Errors = append(result.Errors, resErr)
				errs = append(errs, resErr)
			}

			for _, warn := range crdResults.Warnings {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %s", crdResults.Name, warn))
			}
		}

		versionedResultFunc := func(version, property string, compResult validations.ComparisonResult) {
			for _, err := range compResult.Errors {
				resErr := fmt.Errorf("(%s) %s - %s: %w", version, property, compResult.Name, errors.New(err))
				result.Errors = append(result.Errors, resErr)
				errs = append(errs, resErr)
			}

			for _, warn := range compResult.Warnings {
				result.Warnings = append(result.Warnings, fmt.Sprintf("(%s) %s - %s: %s", version, property, compResult.Name, warn))
			}
		}

		processVersionedPropertyResults(runResults.SameVersionValidation, versionedResultFunc)
		processVersionedPropertyResults(runResults.ServedVersionValidation, versionedResultFunc)

		results = append(results, result)
	}

	if len(errs) > 0 {
		return results, kerrors.NewAggregate(errs)
	}

	return results, nil
}

// processVersionedPropertyResults iterates over all version-property pairs and calls the provided processing function
func processVersionedPropertyResults(vpr map[string]map[string][]validations.ComparisonResult, processFunc func(version, property string, cr validations.ComparisonResult)) {
	for version, versionResults := range vpr {
		for property, propertyResults := range versionResults {
			for _, versionedPropertyResult := range propertyResults {
				processFunc(version, property, versionedPropertyResult)
			}
		}
	}
}

const (
	// featureGatedCRDManifests is the folder name we use to generate
	// partial CRD manifests.
	featureGatedCRDManifests   = "zz_generated.featuregated-crd-manifests"
	manualOverrideCRDManifests = "manual-override-crd-manifests"
)

// getCRDPathsForVersion returns a string slice containing all the relative paths
// for CRDs in the provided generation.APIVersionContext.
// A nil slice and an error are returned if any error occurs while building the
// set of relative paths.
func getCRDPathsForVersion(version generation.APIVersionContext) ([]string, error) {
	contexts := []string{}

	currDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current working directory: %w", err)
	}

	relativePath, err := filepath.Rel(currDir, version.Path)
	if err != nil {
		return nil, fmt.Errorf("determining relative path for version context: %w", err)
	}

	err = recursiveCRDFilesForPath(relativePath, sets.New(featureGatedCRDManifests, manualOverrideCRDManifests), func(path string) {
		contexts = append(contexts, path)
	})
	if err != nil {
		return nil, fmt.Errorf("recursively building CRD file paths: %w", err)
	}

	return contexts, nil
}

func recursiveCRDFilesForPath(baseDir string, directoryIgnoreSet sets.Set[string], handleFunc func(path string)) error {
	dirEntries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("getting entries for directory %q: %w", baseDir, err)
	}

	for _, fileInfo := range dirEntries {
		filePath := filepath.Join(baseDir, fileInfo.Name())
		if fileInfo.IsDir() {
			if directoryIgnoreSet.Has(fileInfo.Name()) {
				continue
			}

			err := recursiveCRDFilesForPath(filePath, directoryIgnoreSet, handleFunc)
			if err != nil {
				return fmt.Errorf("recursing into directory %s: %w", filePath, err)
			}
		}

		if filepath.Ext(fileInfo.Name()) != ".yaml" {
			continue
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read file %s: %v", filePath, err)
		}

		partialObject := &metav1.PartialObjectMetadata{}
		if err := kyaml.Unmarshal(data, partialObject); err != nil {
			return fmt.Errorf("could not unmarshal YAML for type meta inspection: %v", err)
		}

		// Ignore any file that doesn't have a kind of CustomResourceDefinition.
		if !isCustomResourceDefinition(partialObject) {
			continue
		}

		handleFunc(filePath)
	}

	return nil
}

// isCustomResourceDefinition returns true if the object is a CustomResourceDefinition.
// This is determined by the object having a Kind of CustomResourceDefinition and the
// correct APIVersion.
func isCustomResourceDefinition(partialObject *metav1.PartialObjectMetadata) bool {
	return partialObject.APIVersion == apiextensionsv1.SchemeGroupVersion.String() && partialObject.Kind == "CustomResourceDefinition"
}
