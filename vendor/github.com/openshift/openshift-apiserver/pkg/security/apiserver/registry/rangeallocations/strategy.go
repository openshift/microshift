package rangeallocations

import (
	"context"

	securityapi "github.com/openshift/openshift-apiserver/pkg/security/apis/security"
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/core/validation"
)

type strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var strategyInstance = strategy{legacyscheme.Scheme, names.SimpleNameGenerator}

var _ rest.RESTCreateStrategy = strategyInstance
var _ rest.RESTUpdateStrategy = strategyInstance

func (strategy) NamespaceScoped() bool {
	return false
}

func (strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	_ = obj.(*securityapi.RangeAllocation)
}

func (strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	cfg := obj.(*securityapi.RangeAllocation)

	return validation.ValidateObjectMeta(&cfg.ObjectMeta, false, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (strategy) Canonicalize(obj runtime.Object) {
}

func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) PrepareForUpdate(ctx context.Context, newObj, oldObj runtime.Object) {
	_ = oldObj.(*securityapi.RangeAllocation)
	_ = newObj.(*securityapi.RangeAllocation)
}

func (strategy) AllowUnconditionalUpdate() bool {
	return false
}

func (strategy) ValidateUpdate(ctx context.Context, newObj, oldObj runtime.Object) field.ErrorList {
	oldCfg, newCfg := oldObj.(*securityapi.RangeAllocation), newObj.(*securityapi.RangeAllocation)

	return validation.ValidateObjectMetaUpdate(&newCfg.ObjectMeta, &oldCfg.ObjectMeta, field.NewPath("metadata"))
}

// WarningsOnUpdate returns warnings for the given update.
func (strategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
