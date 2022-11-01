package inspect

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestDirectoryViable(t *testing.T) {
	tc := []struct {
		name          string
		dirName       func() (string, error)
		teardown      func(string) error
		allowOverride bool
		expectedErr   error
	}{
		{
			name:    "ensure non-existent directory is viable",
			dirName: func() (string, error) { return "/foo/bar/baz", nil },
		},
		{
			name: "ensure empty directory is viable",
			dirName: func() (string, error) {
				tmpDir, err := ioutil.TempDir(os.TempDir(), "must-gather-inspect-")
				if err != nil {
					return "", err
				}
				return tmpDir, nil
			},
			teardown: defaultTempDirTeardown,
		},
		{
			name: "ensure non-empty directory not viable",
			dirName: func() (string, error) {
				tmpDir, err := ioutil.TempDir(os.TempDir(), "must-gather-inspect-")
				if err != nil {
					return "", err
				}
				_, err = ioutil.TempFile(tmpDir, "must-gather-inspect-file-")
				return tmpDir, err
			},
			expectedErr: fmt.Errorf("exists and is not empty"),
			teardown:    defaultTempDirTeardown,
		},
		{
			name: "ensure non-empty directory viable with data override",
			dirName: func() (string, error) {
				tmpDir, err := ioutil.TempDir(os.TempDir(), "must-gather-inspect-")
				if err != nil {
					return "", err
				}
				_, err = ioutil.TempFile(tmpDir, "must-gather-inspect-file-")
				return tmpDir, err
			},
			allowOverride: true,
			teardown:      defaultTempDirTeardown,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			dirName, err := test.dirName()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.teardown != nil {
				defer test.teardown(dirName)
			}

			o := InspectOptions{destDir: dirName, overwrite: test.allowOverride}
			err = o.ensureDirectoryViable()
			if !errorFuzzyEquals(err, test.expectedErr) {
				t.Fatalf("unexpected error: expecting %v, but got %v", test.expectedErr, err)
			}
		})
	}
}

// fakeSupportedResourceFinder implements supportedResourceFinder
type fakeSupportedResourceFinder struct {
	withError          error
	supportedResources []*metav1.APIResourceList
}

func (f *fakeSupportedResourceFinder) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return f.supportedResources, f.withError
}

func TestAPIGroupVersionRetrieval(t *testing.T) {
	tc := []struct {
		name              string
		resourceFinder    supportedResourceFinder
		apiGroup          string
		expectedResources []schema.GroupVersionResource
		expectedErr       error
	}{
		{
			name: "ensure unknown apiGroup returns no results",
			resourceFinder: &fakeSupportedResourceFinder{
				supportedResources: fakeAPIResourceList(),
			},
			apiGroup: "unknown",
		},
		{
			name: "ensure known apiGroup returns expected results",
			resourceFinder: &fakeSupportedResourceFinder{
				supportedResources: fakeAPIResourceList(),
			},
			apiGroup:          "test.group.io",
			expectedResources: []schema.GroupVersionResource{{Group: "test.group.io", Version: "v1", Resource: "testresources"}},
		},
		{
			name: "ensure different, but known apiGroup returns expected results",
			resourceFinder: &fakeSupportedResourceFinder{
				supportedResources: fakeAPIResourceList(),
			},
			apiGroup:          "foo.group.io",
			expectedResources: []schema.GroupVersionResource{{Group: "foo.group.io", Version: "v1", Resource: "foos"}},
		},
		{
			name: "ensure known apiGroup with NO `list` verb is omitted from results",
			resourceFinder: &fakeSupportedResourceFinder{
				supportedResources: []*metav1.APIResourceList{
					{
						TypeMeta:     metav1.TypeMeta{Kind: "test", APIVersion: "v1"},
						GroupVersion: schema.GroupVersion{Group: "test.group.io", Version: "v1"}.String(),
						APIResources: []metav1.APIResource{
							{
								Name:         "testresources",
								SingularName: "testresource",
								Group:        "test.group.io",
								Version:      "v1",
								Kind:         "test",
								Verbs:        []string{"get"},
							},
						},
					},
				},
			},
			apiGroup: "test.group.io",
		},
		{
			name:           "ensure only discovery errors are returned when no resources are found",
			resourceFinder: &fakeSupportedResourceFinder{withError: fmt.Errorf("test")},
			apiGroup:       "foo",
			expectedErr:    fmt.Errorf("test"),
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			resources, err := retrieveAPIGroupVersionResourceNames(test.resourceFinder, test.apiGroup)
			if !errorFuzzyEquals(err, test.expectedErr) {
				t.Fatalf("unexpected error: %v", err)
			}

			if !resourceListsEqual(resources, test.expectedResources) {
				t.Fatalf("unexpected result; expected %#v, but got %#v", test.expectedResources, resources)
			}
		})
	}
}

func resourceListsEqual(a, b []schema.GroupVersionResource) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	if len(a) == 1 {
		return a[0].String() == b[0].String()
	}

	for idxA := range a {
		for idxB := range b {
			if a[idxA].String() != b[idxB].String() {
				return false
			}
		}
	}
	return true
}

func fakeAPIResourceList() []*metav1.APIResourceList {
	return []*metav1.APIResourceList{
		{
			TypeMeta:     metav1.TypeMeta{Kind: "test", APIVersion: "v1"},
			GroupVersion: schema.GroupVersion{Group: "test.group.io", Version: "v1"}.String(),
			APIResources: []metav1.APIResource{
				{
					Name:         "testresources",
					SingularName: "testresource",
					Group:        "test.group.io",
					Version:      "v1",
					Kind:         "test",
					Verbs:        []string{"list"},
				},
			},
		},
		{
			TypeMeta:     metav1.TypeMeta{Kind: "foo", APIVersion: "v1"},
			GroupVersion: schema.GroupVersion{Group: "foo.group.io", Version: "v1"}.String(),
			APIResources: []metav1.APIResource{
				{
					Name:         "foos",
					SingularName: "foo",
					Group:        "foo.group.io",
					Version:      "v1",
					Kind:         "foo",
					Verbs:        []string{"list"},
				},
			},
		},
	}
}

func defaultTempDirTeardown(dirName string) error {
	if len(dirName) == 0 {
		return nil
	}
	return os.RemoveAll(dirName)
}

func errorFuzzyEquals(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return strings.Contains(a.Error(), b.Error())
}
