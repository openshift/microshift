package emptypartialschemas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

func createFeatureGatedCRDManifests(allCRDInfo map[string]*CRDInfo, outputDir string, verify bool) error {
	errs := []error{}

	for _, crdInfo := range allCRDInfo {
		crdDir := filepath.Join(outputDir, crdInfo.CRDName)
		if !verify {
			if err := os.MkdirAll(crdDir, 0755); err != nil {
				errs = append(errs, fmt.Errorf("failed creating directory: %w", err))
				continue
			}
		}
		allowedFiles := sets.String{}

		for _, featureGate := range append([]string{""}, crdInfo.FeatureGates...) {
			yamlName := fmt.Sprintf("%s.yaml", featureGate)
			if len(featureGate) == 0 {
				if len(crdInfo.TopLevelFeatureGates) > 0 {
					klog.V(3).Infof("skipping to ungated file because the top level type is featuregated")
					continue
				}
				// we need the directory walk to hit the ungated first so the rest cleanly overlay.
				// not beautiful, but clear.
				yamlName = "AAA_ungated.yaml"
			}
			filename := filepath.Join(crdDir, yamlName)
			allowedFiles.Insert(yamlName)
			klog.V(2).Infof("working with %v in featuregate %q in %v", crdInfo.CRDName, featureGate, yamlName)

			minimalCRD := minimalCRDFor(crdInfo, featureGate)

			var existingCRD *apiextensionsv1.CustomResourceDefinition
			existingContent, err := os.ReadFile(filename)
			switch {
			case err == nil && len(existingContent) > 0:
				existingCRD, err = ReadCustomResourceDefinitionV1(existingContent)
				if err != nil {
					errs = append(errs, fmt.Errorf("unable to parse %q: %w", filename, err))
					continue
				}
			case !verify && os.IsNotExist(err):
			case verify && os.IsNotExist(err):
				errs = append(errs, fmt.Errorf("existing missing %w", err))
				continue
			case err != nil:
				errs = append(errs, fmt.Errorf("failed reading existing %w", err))
				continue
			}

			// if the content doesn't match, stomp it.
			err = ensureNoExtraFields(minimalCRD, existingCRD)
			switch {
			case verify && err != nil:
				errs = append(errs, fmt.Errorf("failed to verify %w", err))
				continue
			case !verify && (err != nil || existingCRD == nil):
				fileContent, err := WriteSpecOnlyCustomResourceDefinitionV1(minimalCRD)
				if err != nil {
					errs = append(errs, fmt.Errorf("%q failed to serialize: %w", filename, err))
					continue
				}
				if err := os.WriteFile(filename, fileContent, 0644); err != nil {
					errs = append(errs, fmt.Errorf("%q failed to write: %w", filename, err))
					continue
				}
				continue
			}
		}

		dirContent, err := os.ReadDir(crdDir)
		if err != nil {
			return err
		}
		for _, dirEntry := range dirContent {
			if !allowedFiles.Has(dirEntry.Name()) {
				fileToRemove := filepath.Join(crdDir, dirEntry.Name())
				if verify {
					errs = append(errs, fmt.Errorf("file %q needs to be removed: %w", fileToRemove, err))
					continue
				}

				klog.Infof("Removing %q", fileToRemove)
				if err := os.Remove(fileToRemove); err != nil {
					return err
				}
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}

func ensureNoExtraFields(minimalCRD, existingCRD *apiextensionsv1.CustomResourceDefinition) error {
	if existingCRD == nil {
		return nil
	}
	existingCRDCopy := existingCRD.DeepCopy()
	// the only diff should be the schema
	// We have exactly one version for each partial schema because they are generated per version.
	if len(existingCRDCopy.Spec.Versions) != 1 {
		return fmt.Errorf("bad versions")
	}
	existingCRDCopy.Spec.Versions[0].Schema = nil

	if !equality.Semantic.DeepEqual(minimalCRD, existingCRDCopy) {
		// TODO could be replaced with a prettier diff printer from schemapatch.
		return fmt.Errorf("unexpected diff: %v", cmp.Diff(minimalCRD, existingCRD))
	}

	return nil
}

func minimalCRDFor(crdInfo *CRDInfo, featureGate string) *apiextensionsv1.CustomResourceDefinition {
	ret := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:        crdInfo.CRDName,
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: crdInfo.GroupName,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:     crdInfo.PluralName,
				Singular:   strings.ToLower(crdInfo.KindName),
				Kind:       crdInfo.KindName,
				ListKind:   fmt.Sprintf("%vList", crdInfo.KindName),
				Categories: nil,
			},
			Scope: apiextensionsv1.ResourceScope(crdInfo.Scope),
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:                     crdInfo.Version,
					Served:                   true,
					Storage:                  true,
					Deprecated:               false,
					DeprecationWarning:       nil,
					Schema:                   nil,
					Subresources:             nil,
					AdditionalPrinterColumns: nil,
				},
			},
			PreserveUnknownFields: false,
		},
	}

	for _, fg := range strings.Split(featureGate, "+") {
		// notice that this produces a "" featuregate to mean ungated
		ret.Annotations[fmt.Sprintf("feature-gate.release.openshift.io/%s", fg)] = "true"
	}

	if len(crdInfo.FilenameRunLevel) > 0 {
		ret.Annotations["api.openshift.io/filename-cvo-runlevel"] = crdInfo.FilenameRunLevel
	}
	if len(crdInfo.FilenameOperatorName) > 0 {
		ret.Annotations["api.openshift.io/filename-operator"] = crdInfo.FilenameOperatorName
	}
	if len(crdInfo.FilenameOperatorOrdering) > 0 {
		ret.Annotations["api.openshift.io/filename-ordering"] = crdInfo.FilenameOperatorOrdering
	}
	if len(crdInfo.ApprovedPRNumber) > 0 {
		ret.Annotations["api-approved.openshift.io"] = crdInfo.ApprovedPRNumber
	}
	if len(crdInfo.Capability) > 0 {
		ret.Annotations["capability.openshift.io/name"] = crdInfo.Capability
	}
	for k, v := range crdInfo.Annotations {
		ret.Annotations[k] = v
	}
	for k, v := range crdInfo.Labels {
		ret.Labels[k] = v
	}

	if len(crdInfo.ShortNames) > 0 {
		ret.Spec.Names.ShortNames = crdInfo.ShortNames
	}
	if len(crdInfo.Category) > 0 {
		ret.Spec.Names.Categories = []string{crdInfo.Category}
	}
	if len(crdInfo.PrinterColumns) > 0 {
		ret.Spec.Versions[0].AdditionalPrinterColumns = crdInfo.PrinterColumns
	}
	if crdInfo.HasStatus {
		ret.Spec.Versions[0].Subresources = &apiextensionsv1.CustomResourceSubresources{
			Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
		}
	}

	ret.GetObjectKind().SetGroupVersionKind(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
	return ret
}
