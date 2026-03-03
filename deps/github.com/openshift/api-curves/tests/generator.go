package tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	yamlpatch "github.com/vmware-archive/yaml-patch"

	"github.com/ghodss/yaml"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var (
	timeout = 5 * time.Second
)

// LoadTestSuiteSpecs recursively walks the given paths looking for any file with the suffix `.testsuite.yaml`.
// It then loads these files in SuiteSpec structs ready for the generator to generate the test cases.
func LoadTestSuiteSpecs(paths ...string) ([]SuiteSpec, error) {
	suiteFiles := make(map[string]struct{})

	for _, path := range paths {
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			dirPath := filepath.Base(filepath.Dir(filepath.Dir(path)))
			if !info.IsDir() && strings.HasSuffix(path, ".yaml") && dirPath == "tests" {
				suiteFiles[path] = struct{}{}
			}

			return nil
		}); err != nil {
			return nil, fmt.Errorf("could not load files from path %q: %w", path, err)
		}
	}

	out := []SuiteSpec{}
	for path := range suiteFiles {
		suite, err := loadSuiteFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not set up test suite: %w", err)
		}

		out = append(out, suite)
	}

	return out, nil
}

// loadSuiteFile loads an individual SuiteSpec from the given file name.
func loadSuiteFile(path string) (SuiteSpec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return SuiteSpec{}, fmt.Errorf("could not read file %q: %w", path, err)
	}

	s := SuiteSpec{}
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return SuiteSpec{}, fmt.Errorf("could not unmarshal YAML file %q: %w", path, err)
	}

	if len(s.CRDName) == 0 {
		return SuiteSpec{}, fmt.Errorf("test suite spec %q is invalid: missing required field `crdName`", path)
	}

	s.PerTestRuntimeInfo, err = perTestRuntimeInfo(filepath.Dir(path), s.CRDName, s.FeatureGates)
	if err != nil {
		return SuiteSpec{}, fmt.Errorf("unable to determine which CRD files to use: %w", err)
	}
	if len(s.PerTestRuntimeInfo.CRDFilenames) == 0 {
		return SuiteSpec{}, fmt.Errorf("missing CRD files to use for test %v", path)
	}

	if s.Version == "" {
		version, err := getSuiteSpecTestVersion(s)
		if err != nil {
			return SuiteSpec{}, fmt.Errorf("could not determine test suite CRD version for %q: %w", path, err)
		}

		s.Version = version
	}

	return s, nil
}

// GenerateTestSuite generates a Ginkgo test suite from the provided SuiteSpec.
func GenerateTestSuite(suiteSpec SuiteSpec) {
	for i := range suiteSpec.PerTestRuntimeInfo.CRDFilenames {
		crdFilename := suiteSpec.PerTestRuntimeInfo.CRDFilenames[i]

		baseCRD, err := loadVersionedCRD(suiteSpec, crdFilename)
		Expect(err).ToNot(HaveOccurred())

		suiteName, err := generateSuiteName(suiteSpec, crdFilename)
		Expect(err).ToNot(HaveOccurred())

		Describe(suiteName, Ordered, func() {
			var crdOptions envtest.CRDInstallOptions
			var crd *apiextensionsv1.CustomResourceDefinition

			BeforeEach(OncePerOrdered, func() {
				Expect(k8sClient).ToNot(BeNil(), "Kubernetes client is not initialised")

				crdOptions = envtest.CRDInstallOptions{
					CRDs: []*apiextensionsv1.CustomResourceDefinition{
						baseCRD.DeepCopy(),
					},
				}

				crds, err := envtest.InstallCRDs(cfg, crdOptions)
				Expect(err).ToNot(HaveOccurred())

				Expect(crds).To(HaveLen(1), "Only one CRD should have been installed")
				crd = crds[0]

				Expect(envtest.WaitForCRDs(cfg, crds, crdOptions)).To(Succeed())
			})

			AfterEach(func() {
				// Remove all of the resources we created during the test.
				for _, u := range newUnstructuredsFor(crd) {
					Expect(k8sClient.DeleteAllOf(ctx, u, client.InNamespace("default")))
				}
			})

			AfterEach(OncePerOrdered, func() {
				// Remove the CRD and wait for it to be removed from the API.
				// If we don't wait then subsequent tests may fail.
				Expect(envtest.UninstallCRDs(cfg, crdOptions)).ToNot(HaveOccurred())
				Eventually(komega.Get(crd), timeout).Should(Not(Succeed()))
			})

			generateOnCreateTable(suiteSpec.Tests.OnCreate)
			generateOnUpdateTable(suiteSpec.Tests.OnUpdate, crdFilename)
		})
	}
}

// generateOnCreateTable generates a table of tests from the defined OnCreate tests
// within the test suite test spec.
func generateOnCreateTable(onCreateTests []OnCreateTestSpec) {
	type onCreateTableInput struct {
		featureGate   string
		initial       []byte
		expected      []byte
		expectedError string
	}

	// assertOnCreate runs the actual test for each table entry
	var assertOnCreate interface{} = func(in onCreateTableInput) {
		initialObj, err := newUnstructuredFrom(in.initial)
		Expect(err).ToNot(HaveOccurred(), "initial data should be a valid Kubernetes YAML resource")

		err = k8sClient.Create(ctx, initialObj)
		if in.expectedError != "" {
			Expect(err).To(MatchError(ContainSubstring(in.expectedError)))
			return
		}
		Expect(err).ToNot(HaveOccurred())

		// Fetch the object we just created from the API.
		gotObj := newEmptyUnstructuredFrom(initialObj)
		Expect(k8sClient.Get(ctx, objectKey(initialObj), gotObj))

		expectedObj, err := newUnstructuredFrom(in.expected)
		Expect(err).ToNot(HaveOccurred(), "expected data should be a valid Kubernetes YAML resource when no expected error is provided")

		// Ensure the name and namespace match.
		// The IgnoreAutogeneratedMetadata will ignore any additional meta set in the API.
		expectedObj.SetName(gotObj.GetName())
		expectedObj.SetNamespace(gotObj.GetNamespace())

		Expect(gotObj).To(komega.EqualObject(expectedObj, komega.IgnoreAutogeneratedMetadata))
	}

	// First argument to the table is the test function.
	tableEntries := []interface{}{assertOnCreate}

	// Convert the test specs into table entries
	for _, testEntry := range onCreateTests {
		tableEntries = append(tableEntries, Entry(testEntry.Name, onCreateTableInput{
			initial:       []byte(testEntry.Initial),
			expected:      []byte(testEntry.Expected),
			expectedError: testEntry.ExpectedError,
		}))
	}

	if len(tableEntries) > 1 {
		DescribeTable("On Create", tableEntries...)
	}
}

// generateOnUpdateTable generates a table of tests from the defined OnUpdate tests
// within the test suite test spec.
func generateOnUpdateTable(onUpdateTests []OnUpdateTestSpec, crdFileName string) {
	type onUpdateTableInput struct {
		featureGate         string
		crdPatches          []Patch
		initial             []byte
		updated             []byte
		expected            []byte
		expectedError       string
		expectedStatusError string
	}

	var assertOnUpdate interface{} = func(in onUpdateTableInput) {
		var originalCRDObjectKey client.ObjectKey
		var originalCRDSpec apiextensionsv1.CustomResourceDefinitionSpec

		initialObj, err := newUnstructuredFrom(in.initial)
		Expect(err).ToNot(HaveOccurred(), "initial data should be a valid Kubernetes YAML resource")

		if len(in.crdPatches) > 0 {
			patchedCRD, err := getPatchedCRD(crdFileName, in.crdPatches)
			Expect(err).ToNot(HaveOccurred(), "could not load patched crd")

			originalCRDObjectKey = objectKey(patchedCRD)

			originalCRD := &apiextensionsv1.CustomResourceDefinition{}
			Expect(k8sClient.Get(ctx, originalCRDObjectKey, originalCRD))

			originalCRDSpec = *originalCRD.Spec.DeepCopy()
			originalCRD.Spec = patchedCRD.Spec

			// Add a sentinel field so that we can check that the schema update has persisted.
			originalCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["sentinel"] = apiextensionsv1.JSONSchemaProps{
				Type: "string",
				Enum: []apiextensionsv1.JSON{
					{Raw: []byte(fmt.Sprintf(`"%s+patched"`, initialObj.GetUID()))},
				},
			}
			initialObj.Object["sentinel"] = initialObj.GetUID() + "+patched"

			Expect(k8sClient.Update(ctx, originalCRD)).To(Succeed(), "failed updating patched CRD schema")
		}

		initialStatus, ok, err := unstructured.NestedFieldNoCopy(initialObj.Object, "status")
		Expect(err).ToNot(HaveOccurred())
		if ok {
			Expect(initialStatus).ToNot(BeNil())
		}

		// Use an eventually here, so that we retry until the sential correctly applies.
		Eventually(func() error {
			return k8sClient.Create(ctx, initialObj)
		}, timeout).Should(Succeed(), "initial object should create successfully")

		if initialStatus != nil {
			Expect(unstructured.SetNestedField(initialObj.Object, initialStatus, "status")).To(Succeed(), "should be able to restore initial status")
			Expect(k8sClient.Status().Update(ctx, initialObj)).ToNot(HaveOccurred(), "initial object status should update successfully")
		}

		if len(in.crdPatches) > 0 {
			originalCRD := &apiextensionsv1.CustomResourceDefinition{}
			Expect(k8sClient.Get(ctx, originalCRDObjectKey, originalCRD))

			originalCRD.Spec = originalCRDSpec

			// Add a sentinel field so that we can check that the schema update has persisted.
			originalCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["sentinel"] = apiextensionsv1.JSONSchemaProps{
				Type: "string",
				Enum: []apiextensionsv1.JSON{
					{Raw: []byte(fmt.Sprintf(`"%s+restored"`, initialObj.GetUID()))},
				},
			}

			Expect(k8sClient.Update(ctx, originalCRD)).To(Succeed())

			Eventually(func() error {
				updatedObj := initialObj.DeepCopy()
				updatedObj.Object["sentinel"] = initialObj.GetUID() + "+restored"

				return k8sClient.Update(ctx, updatedObj)
			}, timeout).Should(Succeed(), "Sentinel should be persisted")

			// Drop the sentinel field now we know the rest of the CRD schema is up to date.
			originalCRD.Spec = originalCRDSpec
			Expect(k8sClient.Update(ctx, originalCRD)).To(Succeed())
		}

		// Fetch the object we just created from the API.
		gotObj := newEmptyUnstructuredFrom(initialObj)
		Expect(k8sClient.Get(ctx, objectKey(initialObj), gotObj))

		updatedObj, err := newUnstructuredFrom(in.updated)
		Expect(err).ToNot(HaveOccurred(), "updated data should be a valid Kubernetes YAML resource")

		updatedObjStatus, ok, err := unstructured.NestedFieldNoCopy(updatedObj.Object, "status")
		Expect(err).ToNot(HaveOccurred())
		if ok {
			Expect(updatedObjStatus).ToNot(BeNil())
		}

		// The updated object needs the following fields copied over.
		updatedObj.SetName(gotObj.GetName())
		updatedObj.SetNamespace(gotObj.GetNamespace())
		updatedObj.SetResourceVersion(gotObj.GetResourceVersion())

		err = k8sClient.Update(ctx, updatedObj)
		if in.expectedError != "" {
			Expect(err).To(MatchError(ContainSubstring(in.expectedError)))
			return
		}
		Expect(err).ToNot(HaveOccurred(), "unexpected error updating spec")

		if updatedObjStatus != nil {
			Expect(unstructured.SetNestedField(updatedObj.Object, updatedObjStatus, "status")).To(Succeed(), "should be able to restore updated status")

			err := k8sClient.Status().Update(ctx, updatedObj)
			if in.expectedStatusError != "" {
				Expect(err).To(MatchError(ContainSubstring(in.expectedStatusError)))
				return
			}
			Expect(err).ToNot(HaveOccurred(), "unexpected error updating status")
		}

		Expect(k8sClient.Get(ctx, objectKey(initialObj), gotObj))

		expectedObj, err := newUnstructuredFrom(in.expected)
		Expect(err).ToNot(HaveOccurred(), "expected data should be a valid Kubernetes YAML resource when no expected error is provided")

		// Ensure the name and namespace match.
		// The IgnoreAutogeneratedMetadata will ignore any additional meta set in the API.
		expectedObj.SetName(gotObj.GetName())
		expectedObj.SetNamespace(gotObj.GetNamespace())

		Expect(gotObj).To(komega.EqualObject(expectedObj, komega.IgnoreAutogeneratedMetadata))
	}

	// First argument to the table is the test function.
	tableEntries := []interface{}{assertOnUpdate}

	// Convert the test specs into table entries
	for _, testEntry := range onUpdateTests {
		tableEntries = append(tableEntries, Entry(testEntry.Name, onUpdateTableInput{
			crdPatches:          testEntry.InitialCRDPatches,
			initial:             []byte(testEntry.Initial),
			updated:             []byte(testEntry.Updated),
			expected:            []byte(testEntry.Expected),
			expectedError:       testEntry.ExpectedError,
			expectedStatusError: testEntry.ExpectedStatusError,
		}))
	}

	if len(tableEntries) > 1 {
		DescribeTable("On Update", tableEntries...)
	}
}

// newUnstructuredsFor creates a set of unstructured resources for each version of the CRD.
// This allows us to ensure all CR instances are deleted after each test.
func newUnstructuredsFor(crd *apiextensionsv1.CustomResourceDefinition) []*unstructured.Unstructured {
	out := []*unstructured.Unstructured{}

	for _, version := range crd.Spec.Versions {
		out = append(out, newUnstructuredsForVersion(crd, version.Name))
	}

	return out
}

// newUnstructuredsForVersion creates an unstructured resource for the CRD at a given version.
func newUnstructuredsForVersion(crd *apiextensionsv1.CustomResourceDefinition, version string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}

	u.SetAPIVersion(fmt.Sprintf("%s/%s", crd.Spec.Group, version))
	u.SetKind(crd.Spec.Names.Kind)

	return u
}

// newUnstructuredFrom unmarshals the raw YAML data into an unstructured,
// and then sets the namespace and generateName ahead of the test.
func newUnstructuredFrom(raw []byte) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}

	if err := k8syaml.Unmarshal(raw, &u.Object); err != nil {
		return nil, fmt.Errorf("could not unmarshal raw YAML: %w", err)
	}

	// Names should be unique for each test so ensure we generate a name
	u.SetGenerateName("test-")
	// We need to have a namespace, use the default.
	u.SetNamespace("default")

	return u, nil
}

// newEmptyUnstructuredFrom creates a new unstructured with the same GVK as the input object,
// all other fields are cleared.
func newEmptyUnstructuredFrom(initial *unstructured.Unstructured) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}

	if initial != nil {
		u.GetObjectKind().SetGroupVersionKind(initial.GetObjectKind().GroupVersionKind())
	}

	return u
}

// objectKey extracts a client.ObjectKey from the given object.
func objectKey(obj client.Object) client.ObjectKey {
	return client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

func loadCRDFromFile(filename string) (*apiextensionsv1.CustomResourceDefinition, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not load CRD: %w", err)
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(raw, crd); err != nil {
		return nil, fmt.Errorf("could not unmarshal CRD: %w", err)
	}

	return crd, nil
}

// loadVersionedCRD loads the CRD and removes any version schema that is not the current suite
// version. This allows testing of CRDs for versions that are not currently the storage version.
func loadVersionedCRD(suiteSpec SuiteSpec, crdFilename string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd, err := loadCRDFromFile(crdFilename)
	if err != nil {
		return nil, fmt.Errorf("could not load CRD: %w", err)
	}

	if suiteSpec.Version == "" {
		return crd, nil
	}

	crdVersions := []apiextensionsv1.CustomResourceDefinitionVersion{}

	for _, version := range crd.Spec.Versions {
		if version.Name != suiteSpec.Version {
			continue
		}

		version.Storage = true
		version.Served = true

		crdVersions = append(crdVersions, version)
	}

	if len(crdVersions) == 0 {
		return nil, fmt.Errorf("could not find CRD version matching version %s", suiteSpec.Version)
	}

	crd.Spec.Versions = crdVersions

	return crd, nil
}

// generateSuiteName prepends the specified suite name with the GVR string
// for the CRD under test.
func generateSuiteName(suiteSpec SuiteSpec, crdFilename string) (string, error) {
	crd, err := loadCRDFromFile(crdFilename)
	if err != nil {
		return "", fmt.Errorf("could not load CRD: %w", err)
	}
	featureSet := crd.Annotations["release.openshift.io/feature-set"]
	clusterProfiles := clusterProfilesShortNamesFrom(crd.Annotations)
	filename := filepath.Base(crdFilename)

	gvr := schema.GroupVersionResource{
		Group:    crd.Spec.Group,
		Resource: crd.Spec.Names.Plural,
		Version:  suiteSpec.Version,
	}

	return fmt.Sprintf(
		"[%s][ClusterProfiles=%v][FeatureSet=%q][FeatureGate=%v][File=%v] %s",
		gvr.String(),
		strings.Join(clusterProfiles.List(), ","),
		featureSet,
		strings.Join(suiteSpec.FeatureGates, ","),
		filename,
		suiteSpec.Name,
	), nil
}

// getSuiteSpecTestVersion is used to populate the test suites version
// field when not set.
// This is then used to set storage and served versions as well as
// to generate the test suite name.
func getSuiteSpecTestVersion(suiteSpec SuiteSpec) (string, error) {
	version := ""
	for _, file := range suiteSpec.PerTestRuntimeInfo.CRDFilenames {
		crd, err := loadCRDFromFile(file)
		if err != nil {
			return "", err
		}
		if len(crd.Spec.Versions) > 1 {
			return "", fmt.Errorf("too many versions, specify one in the suite")
		}
		if len(version) == 0 {
			version = crd.Spec.Versions[0].Name
			continue
		}

		if version != crd.Spec.Versions[0].Name {
			return "", fmt.Errorf("too many versions, specify one in the suite.  Saw %v and %v", version, crd.Spec.Versions[0].Name)
		}
	}

	return version, nil
}

func getPatchedCRD(crdFileName string, patches []Patch) (*apiextensionsv1.CustomResourceDefinition, error) {
	patch := yamlpatch.Patch{}

	for _, p := range patches {
		patch = append(patch, yamlpatch.Operation{
			Op:    yamlpatch.Op(p.Op),
			Path:  yamlpatch.OpPath(p.Path),
			Value: yamlpatch.NewNode(p.Value),
		})
	}

	baseDoc, err := os.ReadFile(crdFileName)
	if err != nil {
		return nil, fmt.Errorf("could not read file %q: %w", crdFileName, err)
	}

	patchedDoc, err := patch.Apply(baseDoc)
	if err != nil {
		return nil, fmt.Errorf("could not apply patch: %w", err)
	}

	placeholderWrapper := yamlpatch.NewPlaceholderWrapper("{{", "}}")
	patchedData := bytes.NewBuffer(placeholderWrapper.Unwrap(patchedDoc))

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(patchedData.Bytes(), crd); err != nil {
		return nil, fmt.Errorf("could not unmarshal CRD: %w", err)
	}

	return crd, nil
}
