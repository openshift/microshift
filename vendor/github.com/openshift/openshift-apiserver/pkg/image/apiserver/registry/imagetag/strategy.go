package imagetag

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

// Strategy implements behavior for ImageTags.
type Strategy struct {
	runtime.ObjectTyper
	registryWhitelister whitelist.RegistryWhitelister
}

// NewStrategy is the default logic that applies when creating and updating
// ImageTag objects via the REST API.
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
	newITag := obj.(*imageapi.ImageTag)
	if newITag.Spec != nil && len(newITag.Spec.Name) == 0 {
		_, tag, _ := imageutil.SplitImageStreamTag(newITag.Name)
		newITag.Spec.Name = tag
	}
	newITag.Status = nil
	newITag.Image = nil
}

func (s Strategy) GenerateName(base string) string {
	return base
}

func (s Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	itag := obj.(*imageapi.ImageTag)

	return validation.ValidateImageTagWithWhitelister(ctx, s.registryWhitelister, itag)
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
	newITag := obj.(*imageapi.ImageTag)
	oldITag := old.(*imageapi.ImageTag)

	// status is explicitly not included to force users to submit it (if in the future
	// we wish to allow status to be removed and spec to be set in the same call)
	newITag.SelfLink = oldITag.SelfLink
	newITag.Image = oldITag.Image
}

func (s Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newITag := obj.(*imageapi.ImageTag)
	oldITag := old.(*imageapi.ImageTag)

	return validation.ValidateImageTagUpdateWithWhitelister(ctx, s.registryWhitelister, newITag, oldITag)
}

// WarningsOnUpdate returns warnings for the given update.
func (Strategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// MatchImageTag returns a generic matcher for a given label and field selector.
func MatchImageTag(label labels.Selector, field fields.Selector) kstorage.SelectionPredicate {
	return kstorage.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(o runtime.Object) (labels.Set, fields.Set, error) {
			obj, ok := o.(*imageapi.ImageTag)
			if !ok {
				return nil, nil, fmt.Errorf("not an ImageTag")
			}
			return labels.Set(obj.Labels), SelectableFields(obj), nil
		},
	}
}

// SelectableFields returns a field set that can be used for filter selection
func SelectableFields(obj *imageapi.ImageTag) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}
