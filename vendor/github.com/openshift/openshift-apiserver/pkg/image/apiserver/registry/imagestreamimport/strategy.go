package imagestreamimport

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	"github.com/openshift/library-go/pkg/image/reference"
	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation/whitelist"
)

// strategy implements behavior for ImageStreamImports.
type strategy struct {
	runtime.ObjectTyper
	registryWhitelister whitelist.RegistryWhitelister
}

func NewStrategy(rw whitelist.RegistryWhitelister) *strategy {
	return &strategy{
		ObjectTyper:         legacyscheme.Scheme,
		registryWhitelister: rw,
	}
}

func (s *strategy) NamespaceScoped() bool {
	return true
}

func (s *strategy) GenerateName(string) string {
	return ""
}

func (s *strategy) Canonicalize(runtime.Object) {
}

func (s *strategy) ValidateAllowedRegistries(ctx context.Context, isi *imageapi.ImageStreamImport) field.ErrorList {
	errs := field.ErrorList{}
	validate := func(path *field.Path, name string, insecure bool) field.ErrorList {
		ref, _ := reference.Parse(name)
		registryHost, registryPort := ref.RegistryHostPort(insecure)
		return validation.ValidateRegistryAllowedForImport(ctx, s.registryWhitelister, path.Child("from", "name"), ref.Name, registryHost, registryPort)
	}
	if spec := isi.Spec.Repository; spec != nil && spec.From.Kind == "DockerImage" {
		errs = append(errs, validate(field.NewPath("spec").Child("repository"), spec.From.Name, spec.ImportPolicy.Insecure)...)
	}
	if len(isi.Spec.Images) > 0 {
		for i, image := range isi.Spec.Images {
			errs = append(errs, validate(field.NewPath("spec").Child("images").Index(i), image.From.Name, image.ImportPolicy.Insecure)...)
		}
	}
	return errs
}

func (s *strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	newIST := obj.(*imageapi.ImageStreamImport)
	newIST.Status = imageapi.ImageStreamImportStatus{}
}

func (s *strategy) PrepareImageForCreate(obj runtime.Object) {
	image := obj.(*imageapi.Image)

	// signatures can be added using "images" or "imagesignatures" resources
	image.Signatures = nil

	// Remove the raw manifest as it's very big and this leads to a large memory consumption in etcd.
	image.DockerImageManifest = ""
	image.DockerImageConfig = ""
}

func (s *strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	isi := obj.(*imageapi.ImageStreamImport)
	return validation.ValidateImageStreamImport(isi)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}
