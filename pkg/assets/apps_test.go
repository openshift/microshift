package assets

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestAppsPodAnnotations(t *testing.T) {
	for _, kind := range []string{"Deployment", "DaemonSet"} {
		t.Run(kind, func(t *testing.T) {
			manifest := []byte(fmt.Sprintf(`apiVersion: apps/v1
kind: %s
metadata:
  name: managed
  namespace: managed
spec:
  template:
    metadata:
      annotations:
        existing: retained
    spec:
      containers:
      - name: test
        image: test
`, kind))
			read := func(annotations map[string]string) corev1.PodTemplateSpec {
				if kind == "Deployment" {
					handler := &dpApplier{podAnnotations: annotations}
					handler.Read(manifest, nil, nil)
					return handler.dp.Spec.Template
				}
				handler := &dsApplier{podAnnotations: annotations}
				handler.Read(manifest, nil, nil)
				return handler.ds.Spec.Template
			}
			original := read(nil)
			if !reflect.DeepEqual(original.Annotations, map[string]string{"existing": "retained"}) {
				t.Fatalf("unselected workload annotations changed: %v", original.Annotations)
			}
			annotations := map[string]string{"certificate-generation": "first"}
			first := read(annotations)
			if first.Annotations["existing"] != "retained" || first.Annotations["certificate-generation"] != "first" {
				t.Fatalf("annotations not merged into pod template: %v", first.Annotations)
			}
			if !reflect.DeepEqual(first, read(annotations)) {
				t.Fatal("unchanged generation changes the pod template on restart")
			}
			annotations["certificate-generation"] = "second"
			second := read(annotations)
			if reflect.DeepEqual(first, second) || first.Annotations["certificate-generation"] != "first" {
				t.Fatal("renewal must change only the new pod template")
			}
		})
	}
}

func TestAnnotatePodTemplateWithoutAnnotations(t *testing.T) {
	template := &corev1.PodTemplateSpec{}
	annotatePodTemplate(template, nil)
	if template.Annotations != nil {
		t.Fatal("nil annotations must leave the template unchanged")
	}
	annotatePodTemplate(template, map[string]string{"generation": "one"})
	if template.Annotations["generation"] != "one" {
		t.Fatal("annotation missing")
	}
}
