package manifestmerge

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/sets"

	"path/filepath"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/utils/pointer"
	kyaml "sigs.k8s.io/yaml"
)

var (
	DefaultPayloadFeatureGatePath = filepath.Join("payload-manifests", "featuregates")

	allClusterProfiles = []string{
		"include.release.openshift.io/ibm-cloud-managed",
		"include.release.openshift.io/self-managed-high-availability",
	}
)

// Options contains the configuration required for the schemapatch generator.
type Options struct {
	// Disabled indicates whether the schemapatch generator is disabled or not.
	// This default to false as the schemapatch generator is enabled by default.
	Disabled bool

	// Verify determines whether the generator should verify the content instead
	// of updating the generated file.
	Verify bool

	// PayloadFeatureGatePath is a specified path for the featuregate CRD to inform whether particular
	// gates are off or on.
	// If not set, the default "payload-manifests/featuregates" is used.
	PayloadFeatureGatePath string
}

// generator implements the generation.Generator interface.
// It is designed to generate schemapatch updates for a particular API group.
type generator struct {
	disabled               bool
	verify                 bool
	payloadFeatureGatePath string
	allKnownFeatureSets    sets.String
}

// NewGenerator builds a new schemapatch generator.
func NewGenerator(opts Options) generation.Generator {
	payloadFeatureGatePath := DefaultPayloadFeatureGatePath
	if opts.PayloadFeatureGatePath != "" {
		payloadFeatureGatePath = opts.PayloadFeatureGatePath
	}

	allKnownFeatureSets, err := AllKnownFeatureSets(payloadFeatureGatePath)
	if err != nil {
		panic(err)
	}

	return &generator{
		disabled:               opts.Disabled,
		verify:                 opts.Verify,
		payloadFeatureGatePath: payloadFeatureGatePath,
		allKnownFeatureSets:    allKnownFeatureSets,
	}
}

// ApplyConfig creates returns a new generator based on the configuration passed.
// If the schemapatch configuration is empty, the existing generation is returned.
func (g *generator) ApplyConfig(config *generation.Config) generation.Generator {
	if config == nil || config.ManifestMerge == nil {
		return g
	}

	return NewGenerator(
		Options{
			Disabled: config.ManifestMerge.Disabled,
			Verify:   g.verify,
		},
	)
}

// Name returns the name of the generator.
func (g *generator) Name() string {
	return "manifestMerge"
}

// GenGroup runs the schemapatch generator against the given group context.
func (g *generator) GenGroup(groupCtx generation.APIGroupContext) ([]generation.Result, error) {
	if g.disabled {
		klog.V(2).Infof("Skipping %q for %s", g.Name(), groupCtx.Name)
		return nil, nil
	}

	versionPaths := allVersionPaths(groupCtx.Versions)

	errs := []error{}

	for _, version := range groupCtx.Versions {
		action := "Generating"
		if g.verify {
			action = "Verifying"
		}

		klog.Infof("%s %q for for %s/%s", action, g.Name(), groupCtx.Name, version.Name)

		if err := g.genGroupVersion(groupCtx.Name, version, versionPaths); err != nil {
			errs = append(errs, fmt.Errorf("could not run %q generator for group/version %s/%s: %w", g.Name(), groupCtx.Name, version.Name, err))
		}
	}

	if len(errs) > 0 {
		return nil, kerrors.NewAggregate(errs)
	}

	return nil, nil
}

// genGroupVersion runs the schemapatch generator against a particular version of the API group.
func (g *generator) genGroupVersion(group string, version generation.APIVersionContext, versionPaths []string) error {
	errs := []error{}

	for _, versionPath := range versionPaths {
		resourcePaths := []string{}

		manualCRDOverridesPath := filepath.Join(versionPath, "manual-override-crd-manifests")
		byFeatureGatePath := filepath.Join(versionPath, "zz_generated.featuregated-crd-manifests")
		generatedOutputPath := filepath.Join(versionPath, "zz_generated.crd-manifests")

		possibleResources, err := os.ReadDir(byFeatureGatePath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		for _, path := range possibleResources {
			if path.IsDir() {
				resourcePaths = append(resourcePaths, filepath.Join(byFeatureGatePath, path.Name()))
			}
		}

		allCRDsToRender := []crdForFeatureSet{}

		for _, resourcePath := range resourcePaths {
			// at this point we have a few paths.  In the end, we want to generate manifests for every (crd, clusterprofile, featureset) tuple.
			// a prefix can be specified to associate with an image that creates these, but we create one file each.
			// to start, we need to be able to diff the result against what we already have to do a meaningful review,
			// so to do this we'll allow an all the tuples to a single file option.
			crdName := filepath.Base(resourcePath)

			// allResourcePaths has our generated path and the manual overrides if they exist
			allResourcePaths := []string{resourcePath}
			manualCRDOverridesForCRDPath := filepath.Join(manualCRDOverridesPath, filepath.Base(resourcePath))
			if _, err := os.ReadDir(manualCRDOverridesForCRDPath); err == nil {
				allResourcePaths = append(allResourcePaths, manualCRDOverridesForCRDPath)
			}

			// again in the future we'll expand to clusterprofile, featureset tuples, but for now all clusterprofiles are considered combined
			// this assumption works for everything *except* for authentication.
			resultingCRDs := []crdForFeatureSet{}
			crdFilenamePattern := ""
			for _, clusterProfile := range allClusterProfiles {
				for _, featureSetName := range g.allKnownFeatureSets.List() {
					partialManifestFilter, err := FilterForFeatureSet(g.payloadFeatureGatePath, clusterProfile, featureSetName)
					if err != nil {
						errs = append(errs, err)
						continue
					}

					var mergeErrors []error
					resultingCRD, mergeErrors := mergeAllPertinentCRDsInDirs(allResourcePaths, partialManifestFilter)
					if len(mergeErrors) > 0 {
						errs = append(errs, mergeErrors...)
						continue
					}

					// TODO the filename is carried on the CRD, need to work out how to clean up. probably easier once we have a dedicated directory
					if resultingCRD == nil { // this means we didn't find any file that matched the filter this is ok, we have nothing to do.
						unstructuredResultingCRD := &unstructured.Unstructured{}
						unstructuredResultingCRD.GetObjectKind().SetGroupVersionKind(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
						unstructuredResultingCRD.SetAnnotations(map[string]string{
							"CRDNotPresent": fmt.Sprintf("%d", rand.Int31()), // ensures they won't be equal
						})

						resultingCRDs = append(resultingCRDs, crdForFeatureSet{
							crd:            unstructuredResultingCRD,
							featureSet:     featureSetName,
							clusterProfile: clusterProfile,
							outputFile:     "",
							noData:         true,
						})
						continue
					}

					pluralCRDName, _, _ := unstructured.NestedString(resultingCRD.Object, "spec", "names", "plural")
					fileCVORunLevel := resultingCRD.GetAnnotations()["api.openshift.io/filename-cvo-runlevel"]
					fileOperatorName := resultingCRD.GetAnnotations()["api.openshift.io/filename-operator"]
					fileOperatorOrdering := resultingCRD.GetAnnotations()["api.openshift.io/filename-ordering"]
					outputFilePattern := ""
					switch {
					case len(fileCVORunLevel) > 0 && len(fileOperatorName) > 0 && len(fileOperatorOrdering) > 0:
						outputFilePattern = fmt.Sprintf("%s_%s_%s_%sMARKERS.crd.yaml", fileCVORunLevel, fileOperatorName, fileOperatorOrdering, pluralCRDName)
					case len(fileOperatorName) > 0 && len(fileOperatorOrdering) > 0:
						outputFilePattern = fmt.Sprintf("%s_%s_%sMARKERS.crd.yaml", fileOperatorName, fileOperatorOrdering, pluralCRDName)
					case len(fileOperatorName) > 0:
						outputFilePattern = fmt.Sprintf("%s_%sMARKERS.crd.yaml", fileOperatorName, pluralCRDName)
					case len(fileOperatorOrdering) > 0:
						outputFilePattern = fmt.Sprintf("%s_%sMARKERS.crd.yaml", fileOperatorOrdering, pluralCRDName)
					default:
						outputFilePattern = fmt.Sprintf("%sMARKERS.crd.yaml", pluralCRDName)
					}
					fileMarker := fmt.Sprintf("-%s-%s", clusterProfile, featureSetName)
					outputFileBaseName := strings.ReplaceAll(outputFilePattern, "MARKERS", fileMarker)

					if len(crdFilenamePattern) == 0 {
						crdFilenamePattern = outputFilePattern
					}
					if len(outputFileBaseName) == 0 {
						errs = append(errs, fmt.Errorf("crd %q needs '// +openshift:file-pattern=' %v", crdName, resultingCRD.GetAnnotations()))
						continue
					}

					resultingCRD.SetManagedFields(nil)

					annotations := resultingCRD.GetAnnotations()
					for key := range annotations {
						if strings.HasPrefix(key, "api.openshift.io/filename") {
							delete(annotations, key)
						}
						if strings.HasPrefix(key, "feature-gate.release.openshift.io/") {
							delete(annotations, key)
						}
						if strings.HasPrefix(key, "include.release.openshift.io/") {
							delete(annotations, key)
						}
						if strings.HasPrefix(key, "partial-filename.release.openshift.io/") {
							delete(annotations, key)
						}
					}
					for key := range annotations {
						if strings.HasSuffix(key, "-") {
							toRemove := key[:len(key)-1]
							delete(annotations, toRemove)
							delete(annotations, key)
						}
					}
					resultingCRD.SetAnnotations(annotations)

					resultingCRDs = append(resultingCRDs, crdForFeatureSet{
						crd:            resultingCRD,
						featureSet:     featureSetName,
						clusterProfile: clusterProfile,
					})
				}
			}

			// check to see if all the resultingCRDs are the same
			crdsToRender, err := getCRDsToRender(resultingCRDs, crdFilenamePattern, generatedOutputPath, g.allKnownFeatureSets)
			if err != nil {
				errs = append(errs, fmt.Errorf("crd %q failed to compute CRDs to render: %w", crdName, err))
				continue
			}
			allCRDsToRender = append(allCRDsToRender, crdsToRender...)
		}

		// write this doc.go out so we can vendor these directories and copy content
		depGoFilename := filepath.Join(generatedOutputPath, "doc.go")
		versionFromPath := filepath.Base(versionPath)
		groupFromPath := filepath.Base(filepath.Dir(versionPath))
		simplestGoFile := []byte(fmt.Sprintf("package %s_%s_crdmanifests\n", groupFromPath, versionFromPath))
		if !g.verify {
			if err := os.MkdirAll(generatedOutputPath, 0755); err != nil {
				errs = append(errs, fmt.Errorf("failed creating directory: %w", err))
				continue
			}
			if err := os.WriteFile(depGoFilename, simplestGoFile, 0644); err != nil {
				errs = append(errs, fmt.Errorf("unable to write dep file: %w", err))
				continue
			}
		} else {
			existingContent, err := os.ReadFile(depGoFilename)
			switch {
			case os.IsNotExist(err):
				errs = append(errs, fmt.Errorf("missing doc.go in %v: %w", depGoFilename, err))
			case err != nil:
				errs = append(errs, fmt.Errorf("unable to read dep file: %w", err))
			default:
				if !bytes.Equal(simplestGoFile, existingContent) {
					errs = append(errs, fmt.Errorf("%s content does not match: %v", simplestGoFile, cmp.Diff(simplestGoFile, existingContent)))
				}
			}
		}

		for _, resultingCRD := range allCRDsToRender {
			manifestData, err := kyaml.Marshal(resultingCRD.crd.Object)
			if err != nil {
				errs = append(errs, fmt.Errorf("could not encode file %s: %v", resultingCRD.outputFile, err))
				continue
			}

			if g.verify {
				existingBytes, err := os.ReadFile(resultingCRD.outputFile)
				if err != nil {
					errs = append(errs, fmt.Errorf("could not read file %s: %v", resultingCRD.outputFile, err))
					continue
				}
				if !bytes.Equal(manifestData, existingBytes) {
					diff := utils.Diff(existingBytes, manifestData, resultingCRD.outputFile)

					return fmt.Errorf("API schema for %s is out of date, please regenerate the API schema:\n%s", resultingCRD.outputFile, diff)
				}

				continue
			}

			if err := os.WriteFile(resultingCRD.outputFile, manifestData, 0644); err != nil {
				return fmt.Errorf("could not write manifest %s: %w", resultingCRD.outputFile, err)
			}
		}

		// remove extra content.
		outputResources, err := os.ReadDir(generatedOutputPath)
		switch {
		case g.verify && os.IsNotExist(err):
			// do nothing, may not be a failure if there's nothing to put here
		case err != nil:
			errs = append(errs, fmt.Errorf("failed to read generated output: %w", err))
		}
		for _, curr := range outputResources {
			if curr.IsDir() {
				errs = append(errs, fmt.Errorf("unexpected directory: %q", curr.Name()))
			}
			found := false
			for _, expectedCRD := range allCRDsToRender {
				filename := filepath.Base(expectedCRD.outputFile)
				if curr.Name() == filename {
					found = true
					break
				}
			}
			// always expect the doc.go
			if curr.Name() == "doc.go" {
				found = true
			}

			switch {
			case !found && g.verify:
				errs = append(errs, fmt.Errorf("need to remove: %q", curr.Name()))
			case !found && !g.verify:
				if err := os.Remove(filepath.Join(generatedOutputPath, curr.Name())); err != nil {
					errs = append(errs, fmt.Errorf("failed to remove: %q", curr.Name()))
				}
			}
		}
	}

	return kerrors.NewAggregate(errs)
}

func getCRDsToRender(resultingCRDs []crdForFeatureSet, crdFilenamePattern, outputPath string, allKnownFeatureSets sets.String) ([]crdForFeatureSet, error) {
	allCRDsWithData := filterCRDs(resultingCRDs, &HasData{})
	sameSchemaInAllCRDs := areCRDsTheSame(allCRDsWithData)
	hasAllFeatureSets := featureSetsFromCRDs(allCRDsWithData).Equal(allKnownFeatureSets)
	if sameSchemaInAllCRDs && hasAllFeatureSets {
		crdFilename := strings.ReplaceAll(crdFilenamePattern, "MARKERS", "")
		crdFullPath := filepath.Join(outputPath, crdFilename)
		crdToWrite := allCRDsWithData[0].crd.DeepCopy()

		clusterProfilesToAdd := clusterProfilesFromCRDs(allCRDsWithData)
		if len(clusterProfilesToAdd) == 0 {
			clusterProfilesToAdd = sets.NewString(allClusterProfiles...)
		}
		annotations := crdToWrite.GetAnnotations()
		for _, clusterProfile := range clusterProfilesToAdd.List() {
			annotations[clusterProfile] = "true"
		}
		crdToWrite.SetAnnotations(annotations)

		return []crdForFeatureSet{
			{
				crd:        crdToWrite,
				outputFile: crdFullPath,
			},
		}, nil
	}

	// so they aren't all the same. Check first to see if they're the same for FeatureSet across all ClusterProfiles
	// then check if they're the same for all featuresets on clusterProfile.
	// if they only vary by featureset, then featureset files only
	// if they only vary by clusterprofile, then clusterprofile files only
	// if they vary by both, slice by clusterprofile first, then by featureset
	eachFeatureSetTheSameForAllClusterProfiles := true
	for _, featureSet := range allKnownFeatureSets.List() {
		filter := &AndCRDFilter{
			filters: []CRDFilter{
				&HasData{},
				&FeatureSetFilter{featureSetName: featureSet},
			},
		}
		filteredCRDs := filterCRDs(resultingCRDs, filter)
		sameSchema := areCRDsTheSame(filteredCRDs)
		if len(filteredCRDs) > 0 && !sameSchema {
			eachFeatureSetTheSameForAllClusterProfiles = false
		}
	}
	if eachFeatureSetTheSameForAllClusterProfiles {
		crdsToWrite := []crdForFeatureSet{}
		for _, featureSet := range allKnownFeatureSets.List() {
			filter := &AndCRDFilter{
				filters: []CRDFilter{
					&HasData{},
					&FeatureSetFilter{featureSetName: featureSet},
				},
			}
			filteredCRDs := filterCRDs(resultingCRDs, filter)
			if len(filteredCRDs) == 0 {
				continue
			}

			crdFilename := strings.ReplaceAll(crdFilenamePattern, "MARKERS", fmt.Sprintf("-%s", featureSet))
			crdFullPath := filepath.Join(outputPath, crdFilename)
			crdToWrite := filteredCRDs[0].crd.DeepCopy()

			clusterProfilesToAdd := clusterProfilesFromCRDs(filteredCRDs)
			annotations := crdToWrite.GetAnnotations()
			annotations["release.openshift.io/feature-set"] = featureSet
			for _, clusterProfile := range clusterProfilesToAdd.List() {
				annotations[clusterProfile] = "true"
			}
			crdToWrite.SetAnnotations(annotations)

			crdsToWrite = append(crdsToWrite, crdForFeatureSet{
				crd:        crdToWrite,
				featureSet: featureSet,
				outputFile: crdFullPath,
			})
		}
		return crdsToWrite, nil
	}

	eachClusterProfiletheSameForAllFeatureSets := true
	notHandled := []crdForFeatureSet{}
	crdsToWrite := []crdForFeatureSet{}
	for _, clusterProfile := range allClusterProfiles {
		filter := &AndCRDFilter{
			filters: []CRDFilter{
				&HasData{},
				&ClusterProfileFilter{clusterProfile: clusterProfile},
			},
		}
		filteredCRDs := filterCRDs(resultingCRDs, filter)
		sameSchema := areCRDsTheSame(filteredCRDs)
		if !sameSchema {
			eachClusterProfiletheSameForAllFeatureSets = false
			notHandled = append(notHandled, filteredCRDs...)
			continue
		}

		clusterProfileShortName, err := utils.ClusterProfileToShortName(clusterProfile)
		if err != nil {
			return nil, fmt.Errorf("unrecognized clusterprofile name %q: %w", clusterProfile, err)
		}
		crdFilename := strings.ReplaceAll(crdFilenamePattern, "MARKERS", fmt.Sprintf("-%s", clusterProfileShortName))
		crdFullPath := filepath.Join(outputPath, crdFilename)
		crdToWrite := filteredCRDs[0].crd.DeepCopy()

		annotations := crdToWrite.GetAnnotations()
		annotations[clusterProfile] = "true"
		crdToWrite.SetAnnotations(annotations)

		crdsToWrite = append(crdsToWrite, crdForFeatureSet{
			crd:            crdToWrite,
			clusterProfile: clusterProfile,
			outputFile:     crdFullPath,
		})
	}

	if eachClusterProfiletheSameForAllFeatureSets {
		return crdsToWrite, nil
	}

	// at this point, write each clusterProfile that IS unique, then write the remainder

	for i, curr := range notHandled {
		if curr.noData {
			continue
		}
		clusterProfileShortName, err := utils.ClusterProfileToShortName(curr.clusterProfile)
		if err != nil {
			return nil, fmt.Errorf("unrecognized clusterprofile name %q: %w", curr.clusterProfile, err)
		}
		crdFilename := strings.ReplaceAll(crdFilenamePattern, "MARKERS", fmt.Sprintf("-%s-%s", clusterProfileShortName, curr.featureSet))
		crdFullPath := filepath.Join(outputPath, crdFilename)

		crdToWrite := notHandled[i].crd.DeepCopy()
		annotations := crdToWrite.GetAnnotations()
		annotations["release.openshift.io/feature-set"] = curr.featureSet
		annotations[curr.clusterProfile] = "true"
		crdToWrite.SetAnnotations(annotations)
		crdsToWrite = append(crdsToWrite, crdForFeatureSet{
			crd:            crdToWrite,
			featureSet:     curr.featureSet,
			clusterProfile: curr.clusterProfile,
			outputFile:     crdFullPath,
		})
	}
	return crdsToWrite, nil
}

func clusterProfilesFromCRDs(resultingCRDs []crdForFeatureSet) sets.String {
	ret := sets.String{}
	for _, currCRD := range resultingCRDs {
		ret.Insert(currCRD.clusterProfile)
	}

	return ret
}

func featureSetsFromCRDs(resultingCRDs []crdForFeatureSet) sets.String {
	ret := sets.String{}
	for _, currCRD := range resultingCRDs {
		ret.Insert(currCRD.featureSet)
	}

	return ret
}

func filterCRDs(resultingCRDs []crdForFeatureSet, filter CRDFilter) []crdForFeatureSet {
	ret := []crdForFeatureSet{}
	for i, currCRD := range resultingCRDs {
		if ok := filter.UseCRD(currCRD); ok {
			ret = append(ret, resultingCRDs[i])
		}
	}

	return ret
}

func areCRDsTheSame(resultingCRDs []crdForFeatureSet) bool {
	if len(resultingCRDs) == 0 {
		return false
	}

	var prevCRDMinusIdentifier *unstructured.Unstructured
	for _, currCRD := range resultingCRDs {
		currCRDMinusIdentifier := currCRD.crd.DeepCopy()

		if prevCRDMinusIdentifier == nil {
			prevCRDMinusIdentifier = currCRDMinusIdentifier
			continue
		}
		if !equality.Semantic.DeepEqual(*prevCRDMinusIdentifier, *currCRDMinusIdentifier) {
			return false
		}
	}

	return true
}

type crdForFeatureSet struct {
	crd            *unstructured.Unstructured
	featureSet     string
	clusterProfile string
	outputFile     string
	noData         bool
}

// pertinent is determined by the `filter`. If it passes the filter, it's pertinent.
// filters commonly include clusterprofile and featureset-to-feature-gate mapping.
// for example, the TechPreviewNoUpgrade featureset produces a filter that looks to see if the featuregate specified is
// enabled when TechPreviewNoUpgrade is set.
func mergeAllPertinentCRDsInDirs(resourcePaths []string, filter ManifestFilter) (*unstructured.Unstructured, []error) {
	var resultingCRD *unstructured.Unstructured
	var errs []error

	for _, resourcePath := range resourcePaths {
		var currErrs []error
		resultingCRD, currErrs = mergeAllPertinentCRDsInDir(resourcePath, filter, resultingCRD)
		errs = append(errs, currErrs...)
	}

	return resultingCRD, errs
}

// pertinent is determined by the `filter`. If it passes the filter, it's pertinent.
// filters commonly include clusterprofile and featureset-to-feature-gate mapping.
// for example, the TechPreviewNoUpgrade featureset produces a filter that looks to see if the featuregate specified is
// enabled when TechPreviewNoUpgrade is set.
func mergeAllPertinentCRDsInDir(resourcePath string, filter ManifestFilter, startingCRD *unstructured.Unstructured) (*unstructured.Unstructured, []error) {

	var unstructuredResultingCRD *unstructured.Unstructured
	if startingCRD != nil {
		unstructuredResultingCRD = startingCRD.DeepCopy()

	} else {
		annotations := map[string]string{
			"api.openshift.io/merged-by-featuregates": "true",
		}

		unstructuredResultingCRD = &unstructured.Unstructured{}
		unstructuredResultingCRD.GetObjectKind().SetGroupVersionKind(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
		unstructuredResultingCRD.SetAnnotations(annotations)
	}

	errs := []error{}
	var resultingCRD runtime.Object
	resultingCRD = unstructuredResultingCRD

	partialManifestFiles, err := os.ReadDir(resourcePath)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	// Sort the manifests such that any combination of feature gates is applied after the gates that it combines.
	// This means that if we have a file that is "foo+bar" and a file that is "foo", the "foo+bar" file will be applied last.
	// This enables more speicfic handling for combinations of feature gates that affect the same field.
	slices.SortStableFunc(partialManifestFiles, func(a, b os.DirEntry) int {
		aBase := strings.TrimSuffix(filepath.Base(a.Name()), ".yaml")
		bBase := strings.TrimSuffix(filepath.Base(b.Name()), ".yaml")

		if strings.Contains(bBase, "+") && strings.Contains(bBase, aBase) {
			return -1
		}
		if strings.Contains(aBase, "+") && strings.Contains(aBase, bBase) {
			return 1
		}

		return strings.Compare(aBase, bBase)
	})

	foundAFile := false
	for _, partialManifest := range partialManifestFiles {
		path := filepath.Join(resourcePath, partialManifest.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not read file %s: %w", path, err))
			continue
		}
		useManifest, err := filter.UseManifest(data)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not determine whether to use file %s: %w", path, err))
			continue
		}
		if !useManifest {
			continue
		}

		foundAFile = true
		newResult, err := mergeCRD(resultingCRD, data, path)
		if err != nil {
			errs = append(errs, fmt.Errorf("error applying %q: %w", path, err))
			continue
		}

		// I added this to debug which files were being combined part way through the generation
		annotations := newResult.(*unstructured.Unstructured).GetAnnotations()
		annotations[fmt.Sprintf("partial-filename.release.openshift.io/%s", path)] = "true"
		newResult.(*unstructured.Unstructured).SetAnnotations(annotations)

		resultingCRD = newResult
	}

	if !foundAFile {
		return nil, errs
	}

	return resultingCRD.(*unstructured.Unstructured), errs
}

// allVersionPaths creates a list of all version paths for the group.
func allVersionPaths(versions []generation.APIVersionContext) []string {
	out := []string{}

	for _, version := range versions {
		out = append(out, version.Path)
	}

	return out
}

func mergeCRD(obj runtime.Object, patchBytes []byte, fieldManager string) (runtime.Object, error) {
	ssaFieldManager, err := getApplyFieldManager()
	if err != nil {
		return nil, err
	}

	serverSideApplyPatcher := &applyPatcher{
		patch: patchBytes,
		options: &metav1.PatchOptions{
			Force:        pointer.BoolPtr(true),
			FieldManager: fieldManager,
		},
		fieldManager: ssaFieldManager,
	}

	ret := obj.DeepCopyObject()
	output, err := serverSideApplyPatcher.applyPatchToCurrentObject(ret)
	if err != nil {
		return nil, err
	}

	return output, nil
}

//go:embed crd-schema.json
var crdSchemaJSON []byte

var (
	schemaReadOnce sync.Once
	schemaReadErr  error
	schemaSchema   map[string]*spec.Schema
)

func getCRDSchema() (map[string]*spec.Schema, error) {
	schemaReadOnce.Do(func() {
		openapiV3CRD := &spec3.OpenAPI{}
		err := json.Unmarshal(crdSchemaJSON, openapiV3CRD)
		if err != nil {
			schemaReadErr = err
			return
		}
		schemaSchema = openapiV3CRD.Components.Schemas
	})
	return schemaSchema, schemaReadErr
}

var (
	fieldManagerOnce sync.Once
	fieldManagerErr  error
	fieldManagerRet  *managedfields.FieldManager
)

func getApplyFieldManager() (*managedfields.FieldManager, error) {
	openAPIModels, err := getCRDSchema()
	if err != nil {
		return nil, err
	}

	fieldManagerOnce.Do(func() {
		typeConverter, err := managedfields.NewTypeConverter(openAPIModels, false)
		if err != nil {
			fieldManagerErr = err
			return
		}
		fieldManager, err := managedfields.NewDefaultCRDFieldManager(
			typeConverter,
			noopConverter{},
			noopDefaulter{},
			noopCreator{},
			apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
			apiextensionsv1.SchemeGroupVersion,
			"",
			nil,
		)
		if err != nil {
			fieldManagerErr = err
			return
		}
		fieldManagerRet = fieldManager
	})

	return fieldManagerRet, fieldManagerErr
}

type applyPatcher struct {
	patch        []byte
	options      *metav1.PatchOptions
	fieldManager *managedfields.FieldManager
}

func (p *applyPatcher) applyPatchToCurrentObject(obj runtime.Object) (runtime.Object, error) {
	force := false
	if p.options.Force != nil {
		force = *p.options.Force
	}
	if p.fieldManager == nil {
		panic("FieldManager must be installed to run apply")
	}

	patchObj := &unstructured.Unstructured{Object: map[string]interface{}{}}
	if err := kyaml.Unmarshal(p.patch, &patchObj.Object); err != nil {
		return nil, fmt.Errorf("unable to unmarshal the patch: %w", err)
	}

	obj, err := p.fieldManager.Apply(obj, patchObj, p.options.FieldManager, force)
	if err != nil {
		return obj, err
	}

	return obj, nil
}

func readCRDYaml(data []byte) (*unstructured.Unstructured, error) {
	json, err := kyaml.YAMLToJSON(data)
	if err != nil {
		json = data
	}
	obj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, json)
	if err != nil {
		return nil, err
	}

	return obj.(*unstructured.Unstructured), nil
}
