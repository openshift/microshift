package originimagereferencemutators

import (
	"fmt"

	kapiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kapi "k8s.io/kubernetes/pkg/apis/core"

	"github.com/openshift/apiserver-library-go/pkg/admission/imagepolicy/imagereferencemutators"
	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
)

type OriginImageMutators struct {
	imagereferencemutators.KubeImageMutators
}

// GetImageReferenceMutator returns a mutator for the provided object, or an error if no
// such mutator is defined. Only references that are different between obj and old will
// be returned unless old is nil.
func (o OriginImageMutators) GetImageReferenceMutator(obj, old runtime.Object) (imagereferencemutators.ImageReferenceMutator, error) {
	oldAnnotations, _ := o.GetAnnotationAccessor(old)
	annotations, _ := o.GetAnnotationAccessor(obj)
	resolveAnnotationChanged := imagereferencemutators.ResolveAllNames(annotations) != imagereferencemutators.ResolveAllNames(oldAnnotations)

	switch t := obj.(type) {
	case *buildapi.Build:
		if oldT, ok := old.(*buildapi.Build); ok && oldT != nil {
			return &buildSpecMutator{
				spec:                     &t.Spec.CommonSpec,
				oldSpec:                  &oldT.Spec.CommonSpec,
				path:                     field.NewPath("spec"),
				resolveAnnotationChanged: resolveAnnotationChanged}, nil
		}
		return &buildSpecMutator{spec: &t.Spec.CommonSpec, path: field.NewPath("spec")}, nil
	case *buildapi.BuildConfig:
		if oldT, ok := old.(*buildapi.BuildConfig); ok && oldT != nil {
			return &buildSpecMutator{
				spec:                     &t.Spec.CommonSpec,
				oldSpec:                  &oldT.Spec.CommonSpec,
				path:                     field.NewPath("spec"),
				resolveAnnotationChanged: resolveAnnotationChanged}, nil
		}
		return &buildSpecMutator{spec: &t.Spec.CommonSpec, path: field.NewPath("spec")}, nil
	}
	if spec, path, err := getPodSpec(obj); err == nil {
		var oldSpec *kapi.PodSpec
		if old != nil {
			oldSpec, _, err = getPodSpec(old)
			if err != nil {
				return nil, fmt.Errorf("old and new pod spec objects were not of the same type %T != %T: %v", obj, old, err)
			}
		}
		return imagereferencemutators.NewPodSpecMutator(spec, oldSpec, path, resolveAnnotationChanged), nil
	}
	if spec, path, err := getPodSpecV1(obj); err == nil {
		var oldSpec *kapiv1.PodSpec
		if old != nil {
			oldSpec, _, err = getPodSpecV1(old)
			if err != nil {
				return nil, fmt.Errorf("old and new pod spec objects were not of the same type %T != %T: %v", obj, old, err)
			}
		}
		return imagereferencemutators.NewPodSpecV1Mutator(spec, oldSpec, path, resolveAnnotationChanged), nil
	}
	return o.KubeImageMutators.GetImageReferenceMutator(obj, old)
}

type annotationsAccessor struct {
	object   metav1.Object
	template metav1.Object
}

func (a annotationsAccessor) Annotations() map[string]string {
	return a.object.GetAnnotations()
}

func (a annotationsAccessor) TemplateAnnotations() (map[string]string, bool) {
	if a.template == nil {
		return nil, false
	}
	return a.template.GetAnnotations(), true
}

func (a annotationsAccessor) SetAnnotations(annotations map[string]string) {
	a.object.SetAnnotations(annotations)
}

func (a annotationsAccessor) SetTemplateAnnotations(annotations map[string]string) bool {
	if a.template == nil {
		return false
	}
	a.template.SetAnnotations(annotations)
	return true
}

// GetAnnotationAccessor returns an accessor for the provided object or false if the object
// does not support accessing annotations.
func (o OriginImageMutators) GetAnnotationAccessor(obj runtime.Object) (imagereferencemutators.AnnotationAccessor, bool) {
	switch t := obj.(type) {
	case metav1.Object:
		templateObject, _ := getTemplateMetaObject(obj)
		return annotationsAccessor{object: t, template: templateObject}, true
	default:
		return o.KubeImageMutators.GetAnnotationAccessor(obj)
	}
}
