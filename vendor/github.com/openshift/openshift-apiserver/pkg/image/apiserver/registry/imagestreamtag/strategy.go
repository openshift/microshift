package imagestreamtag

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	kstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	"github.com/openshift/library-go/pkg/image/imageutil"
	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation/whitelist"
)

// Strategy implements behavior for ImageStreamTags.
type Strategy struct {
	runtime.ObjectTyper
	registryWhitelister whitelist.RegistryWhitelister
}

// NewStrategy is the default logic that applies when creating and updating
// ImageStreamTag objects via the REST API.
func NewStrategy(registryWhitelister whitelist.RegistryWhitelister) Strategy {
	return Strategy{
		ObjectTyper:         legacyscheme.Scheme,
		registryWhitelister: registryWhitelister,
	}
}

func (s Strategy) NamespaceScoped() bool {
	return true
}

func (s Strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	newIST := obj.(*imageapi.ImageStreamTag)
	if newIST.Tag != nil && len(newIST.Tag.Name) == 0 {
		_, tag, _ := imageutil.SplitImageStreamTag(newIST.Name)
		newIST.Tag.Name = tag
	}
	newIST.Conditions = nil
	newIST.Image = imageapi.Image{}
}

func (s Strategy) GenerateName(base string) string {
	return base
}

func (s Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	istag := obj.(*imageapi.ImageStreamTag)

	return validation.ValidateImageStreamTagWithWhitelister(ctx, s.registryWhitelister, istag)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (Strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (s Strategy) AllowCreateOnUpdate() bool {
	return false
}

func (Strategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (Strategy) Canonicalize(obj runtime.Object) {
}

func (s Strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newIST := obj.(*imageapi.ImageStreamTag)
	oldIST := old.(*imageapi.ImageStreamTag)

	// for backwards compatibility, callers can't be required to set both annotation locations when
	// doing a GET and then update.
	if newIST.Tag != nil {
		newIST.Tag.Annotations = newIST.Annotations
	}
	newIST.Conditions = oldIST.Conditions
	newIST.SelfLink = oldIST.SelfLink
	newIST.Image = oldIST.Image
}

func (s Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newIST := obj.(*imageapi.ImageStreamTag)
	oldIST := old.(*imageapi.ImageStreamTag)

	return validation.ValidateImageStreamTagUpdateWithWhitelister(ctx, s.registryWhitelister, newIST, oldIST)
}

// WarningsOnUpdate returns warnings for the given update.
func (Strategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// MatchImageStreamTag returns a generic matcher for a given label and field selector.
func MatchImageStreamTag(label labels.Selector, field fields.Selector) kstorage.SelectionPredicate {
	return kstorage.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(o runtime.Object) (labels.Set, fields.Set, error) {
			obj, ok := o.(*imageapi.ImageStreamTag)
			if !ok {
				return nil, nil, fmt.Errorf("not an ImageStreamTag")
			}
			return labels.Set(obj.Labels), SelectableFields(obj), nil
		},
	}
}

// SelectableFields returns a field set that can be used for filter selection
func SelectableFields(obj *imageapi.ImageStreamTag) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}
