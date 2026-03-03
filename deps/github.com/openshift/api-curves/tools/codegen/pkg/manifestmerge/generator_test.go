package manifestmerge

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"reflect"
	kyaml "sigs.k8s.io/yaml"
	"testing"
)

func Test_mergeCRD2(t *testing.T) {
	type args struct {
		obj   *unstructured.Unstructured
		patch []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "version-overlay",
			args: args{
				obj:   readCRDYamlOrDie(t, "testdata/field-overlay/01.yaml"),
				patch: readFileOrDie(t, "testdata/field-overlay/02.yaml"),
			},
			want: readFileOrDie(t, "testdata/field-overlay/expected-01-first.yaml"),
		},
		{
			name: "version-invert",
			args: args{
				obj:   readCRDYamlOrDie(t, "testdata/field-overlay/02.yaml"),
				patch: readFileOrDie(t, "testdata/field-overlay/01.yaml"),
			},
			want: readFileOrDie(t, "testdata/field-overlay/expected-02-first.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("left these for convenience, but the field manager time needs to be stripped")
			if got, _ := mergeCRD(tt.args.obj, tt.args.patch, "field-manager-a"); !reflect.DeepEqual(got, tt.want) {
				asMap := got.(*unstructured.Unstructured).Object
				outBytes, err := kyaml.Marshal(asMap)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(outBytes, tt.want) {
					t.Errorf("%v", string(outBytes))
				}
			}
		})
	}
}

func readCRDYamlOrDie(t *testing.T, path string) *unstructured.Unstructured {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := readCRDYaml(data)
	if err != nil {
		t.Fatal(err)
	}
	return ret
}

func readFileOrDie(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
