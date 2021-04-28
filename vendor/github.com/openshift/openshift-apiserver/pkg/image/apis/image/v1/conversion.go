package v1

import (
	"sort"
	"strings"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	core "k8s.io/kubernetes/pkg/apis/core"
	corev1 "k8s.io/kubernetes/pkg/apis/core/v1"

	v1 "github.com/openshift/api/image/v1"
	"github.com/openshift/library-go/pkg/image/reference"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	newer "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
)

var (
	dockerImageScheme = runtime.NewScheme()
	dockerImageCodecs = serializer.NewCodecFactory(dockerImageScheme)
)

func init() {
	Install(dockerImageScheme)
}

// The docker metadata must be cast to a version
func Convert_image_Image_To_v1_Image(in *newer.Image, out *v1.Image, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.DockerImageReference = in.DockerImageReference
	out.DockerImageManifest = in.DockerImageManifest
	out.DockerImageManifestMediaType = in.DockerImageManifestMediaType
	out.DockerImageConfig = in.DockerImageConfig

	gvString := in.DockerImageMetadataVersion
	if len(gvString) == 0 {
		gvString = "1.0"
	}
	if !strings.Contains(gvString, "/") {
		gvString = "image.openshift.io/" + gvString
	}

	version, err := schema.ParseGroupVersion(gvString)
	if err != nil {
		return err
	}
	data, err := runtime.Encode(dockerImageCodecs.LegacyCodec(version), &in.DockerImageMetadata)
	if err != nil {
		return err
	}
	out.DockerImageMetadata.Raw = data
	out.DockerImageMetadataVersion = version.Version

	if in.DockerImageLayers != nil {
		out.DockerImageLayers = make([]v1.ImageLayer, len(in.DockerImageLayers))
		for i := range in.DockerImageLayers {
			out.DockerImageLayers[i].MediaType = in.DockerImageLayers[i].MediaType
			out.DockerImageLayers[i].Name = in.DockerImageLayers[i].Name
			out.DockerImageLayers[i].LayerSize = in.DockerImageLayers[i].LayerSize
		}
	} else {
		out.DockerImageLayers = nil
	}

	if in.Signatures != nil {
		out.Signatures = make([]v1.ImageSignature, len(in.Signatures))
		for i := range in.Signatures {
			if err := s.Convert(&in.Signatures[i], &out.Signatures[i]); err != nil {
				return err
			}
		}
	} else {
		out.Signatures = nil
	}

	if in.DockerImageSignatures != nil {
		out.DockerImageSignatures = make([][]byte, 0, len(in.DockerImageSignatures))
		out.DockerImageSignatures = append(out.DockerImageSignatures, in.DockerImageSignatures...)
	} else {
		out.DockerImageSignatures = nil
	}

	return nil
}

func Convert_v1_Image_To_image_Image(in *v1.Image, out *newer.Image, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	out.DockerImageReference = in.DockerImageReference
	out.DockerImageManifest = in.DockerImageManifest
	out.DockerImageManifestMediaType = in.DockerImageManifestMediaType
	out.DockerImageConfig = in.DockerImageConfig

	version := in.DockerImageMetadataVersion
	if len(version) == 0 {
		version = "1.0"
	}
	if len(in.DockerImageMetadata.Raw) > 0 {
		// TODO: add a way to default the expected kind and version of an object if not set
		obj, err := dockerImageScheme.New(schema.GroupVersionKind{Group: "image.openshift.io", Version: version, Kind: "DockerImage"})
		if err != nil {
			return err
		}
		if err := runtime.DecodeInto(dockerImageCodecs.UniversalDecoder(), in.DockerImageMetadata.Raw, obj); err != nil {
			return err
		}
		if err := s.Convert(obj, &out.DockerImageMetadata); err != nil {
			return err
		}
	}
	out.DockerImageMetadataVersion = version

	if in.DockerImageLayers != nil {
		out.DockerImageLayers = make([]newer.ImageLayer, len(in.DockerImageLayers))
		for i := range in.DockerImageLayers {
			out.DockerImageLayers[i].MediaType = in.DockerImageLayers[i].MediaType
			out.DockerImageLayers[i].Name = in.DockerImageLayers[i].Name
			out.DockerImageLayers[i].LayerSize = in.DockerImageLayers[i].LayerSize
		}
	} else {
		out.DockerImageLayers = nil
	}

	if in.Signatures != nil {
		out.Signatures = make([]newer.ImageSignature, len(in.Signatures))
		for i := range in.Signatures {
			if err := s.Convert(&in.Signatures[i], &out.Signatures[i]); err != nil {
				return err
			}
		}
	} else {
		out.Signatures = nil
	}

	if in.DockerImageSignatures != nil {
		out.DockerImageSignatures = nil
		for _, v := range in.DockerImageSignatures {
			out.DockerImageSignatures = append(out.DockerImageSignatures, v)
		}
	} else {
		out.DockerImageSignatures = nil
	}

	return nil
}

func Convert_runtime_RawExtension_To_image_DockerImage(in *runtime.RawExtension, out *image.DockerImage, s conversion.Scope) error {
	return s.Convert(in, out)
}

func Convert_image_DockerImage_To_runtime_RawExtension(in *image.DockerImage, out *runtime.RawExtension, s conversion.Scope) error {
	return s.Convert(in, out)
}

func Convert_v1_ImageStreamSpec_To_image_ImageStreamSpec(in *v1.ImageStreamSpec, out *newer.ImageStreamSpec, s conversion.Scope) error {
	out.LookupPolicy = newer.ImageLookupPolicy{Local: in.LookupPolicy.Local}
	out.DockerImageRepository = in.DockerImageRepository
	out.Tags = make(map[string]newer.TagReference)
	return s.Convert(&in.Tags, &out.Tags)
}

func Convert_image_ImageStreamSpec_To_v1_ImageStreamSpec(in *newer.ImageStreamSpec, out *v1.ImageStreamSpec, s conversion.Scope) error {
	out.LookupPolicy = v1.ImageLookupPolicy{Local: in.LookupPolicy.Local}
	out.DockerImageRepository = in.DockerImageRepository
	if len(in.DockerImageRepository) > 0 {
		// ensure that stored image references have no tag or ID, which was possible from 1.0.0 until 1.0.7
		if ref, err := reference.Parse(in.DockerImageRepository); err == nil {
			if len(ref.Tag) > 0 || len(ref.ID) > 0 {
				ref.Tag, ref.ID = "", ""
				out.DockerImageRepository = ref.Exact()
			}
		}
	}
	out.Tags = make([]v1.TagReference, 0, 0)
	return s.Convert(&in.Tags, &out.Tags)
}

func Convert_v1_ImageStreamStatus_To_image_ImageStreamStatus(in *v1.ImageStreamStatus, out *newer.ImageStreamStatus, s conversion.Scope) error {
	out.DockerImageRepository = in.DockerImageRepository
	out.PublicDockerImageRepository = in.PublicDockerImageRepository
	out.Tags = make(map[string]newer.TagEventList)
	return Convert_v1_NamedTagEventListArray_to_api_TagEventListArray(&in.Tags, &out.Tags, s)
}

func Convert_image_ImageStreamStatus_To_v1_ImageStreamStatus(in *newer.ImageStreamStatus, out *v1.ImageStreamStatus, s conversion.Scope) error {
	out.DockerImageRepository = in.DockerImageRepository
	out.PublicDockerImageRepository = in.PublicDockerImageRepository
	if len(in.DockerImageRepository) > 0 {
		// ensure that stored image references have no tag or ID, which was possible from 1.0.0 until 1.0.7
		if ref, err := reference.Parse(in.DockerImageRepository); err == nil {
			if len(ref.Tag) > 0 || len(ref.ID) > 0 {
				ref.Tag, ref.ID = "", ""
				out.DockerImageRepository = ref.Exact()
			}
		}
	}
	out.Tags = make([]v1.NamedTagEventList, 0, 0)
	return Convert_image_TagEventListArray_to_v1_NamedTagEventListArray(&in.Tags, &out.Tags, s)
}

func Convert_image_TagEventConditionArray_to_v1_TagEventConditionArray(in *[]newer.TagEventCondition, out *[]v1.TagEventCondition, s conversion.Scope) error {
	for _, o := range *in {
		n := v1.TagEventCondition{}
		if err := Convert_image_TagEventCondition_To_v1_TagEventCondition(&o, &n, s); err != nil {
			return err
		}
		*out = append(*out, n)
	}
	return nil
}

func Convert_v1_TagEventConditionArray_to_image_TagEventConditionArray(in *[]v1.TagEventCondition, out *[]newer.TagEventCondition, s conversion.Scope) error {
	for _, o := range *in {
		n := newer.TagEventCondition{}
		if err := Convert_v1_TagEventCondition_To_image_TagEventCondition(&o, &n, s); err != nil {
			return err
		}
		*out = append(*out, n)
	}
	return nil
}

func Convert_image_TagEventArray_to_v1_TagEventArray(in *[]newer.TagEvent, out *[]v1.TagEvent, s conversion.Scope) error {
	for _, o := range *in {
		n := v1.TagEvent{}
		if err := Convert_image_TagEvent_To_v1_TagEvent(&o, &n, s); err != nil {
			return err
		}
		*out = append(*out, n)
	}
	return nil
}

func Convert_v1_TagEventArray_to_image_TagEventArray(in *[]v1.TagEvent, out *[]newer.TagEvent, s conversion.Scope) error {
	for _, o := range *in {
		n := newer.TagEvent{}
		if err := Convert_v1_TagEvent_To_image_TagEvent(&o, &n, s); err != nil {
			return err
		}
		*out = append(*out, n)
	}
	return nil
}

func Convert_v1_NamedTagEventListArray_to_api_TagEventListArray(in *[]v1.NamedTagEventList, out *map[string]newer.TagEventList, s conversion.Scope) error {
	for _, curr := range *in {
		newTagEventList := newer.TagEventList{}
		if err := Convert_v1_TagEventConditionArray_to_image_TagEventConditionArray(&curr.Conditions, &newTagEventList.Conditions, s); err != nil {
			return err
		}
		if err := Convert_v1_TagEventArray_to_image_TagEventArray(&curr.Items, &newTagEventList.Items, s); err != nil {
			return err
		}
		(*out)[curr.Tag] = newTagEventList
	}

	return nil
}
func Convert_image_TagEventListArray_to_v1_NamedTagEventListArray(in *map[string]newer.TagEventList, out *[]v1.NamedTagEventList, s conversion.Scope) error {
	allKeys := make([]string, 0, len(*in))
	for key := range *in {
		allKeys = append(allKeys, key)
	}
	sort.Strings(allKeys)

	for _, key := range allKeys {
		newTagEventList := (*in)[key]
		oldTagEventList := &v1.NamedTagEventList{Tag: key}
		if err := Convert_image_TagEventConditionArray_to_v1_TagEventConditionArray(&newTagEventList.Conditions, &oldTagEventList.Conditions, s); err != nil {
			return err
		}
		if err := Convert_image_TagEventArray_to_v1_TagEventArray(&newTagEventList.Items, &oldTagEventList.Items, s); err != nil {
			return err
		}

		*out = append(*out, *oldTagEventList)
	}

	return nil
}
func Convert_v1_TagReferenceArray_to_api_TagReferenceMap(in *[]v1.TagReference, out *map[string]newer.TagReference, s conversion.Scope) error {
	for _, curr := range *in {
		r := newer.TagReference{}
		if err := s.Convert(&curr, &r); err != nil {
			return err
		}
		(*out)[curr.Name] = r
	}
	return nil
}
func Convert_image_TagReferenceMap_to_v1_TagReferenceArray(in *map[string]newer.TagReference, out *[]v1.TagReference, s conversion.Scope) error {
	allTags := make([]string, 0, len(*in))
	for tag := range *in {
		allTags = append(allTags, tag)
	}
	sort.Strings(allTags)

	for _, tag := range allTags {
		newTagReference := (*in)[tag]
		oldTagReference := v1.TagReference{}
		if err := s.Convert(&newTagReference, &oldTagReference); err != nil {
			return err
		}
		oldTagReference.Name = tag
		*out = append(*out, oldTagReference)
	}
	return nil
}

// Convert_image_v1_SecretList_To_v1_SecretList is an autogenerated conversion function.
func Convert_image_v1_SecretList_To_v1_SecretList(in *v1.SecretList, out *core.SecretList, s conversion.Scope) error {
	v := apicorev1.SecretList(*in)
	return corev1.Convert_v1_SecretList_To_core_SecretList(&v, out, s)
}

// Convert_v1_SecretList_To_image_v1_SecretList is an autogenerated conversion function.
func Convert_v1_SecretList_To_image_v1_SecretList(in *core.SecretList, out *v1.SecretList, s conversion.Scope) error {
	apiCorev1SecretList := &apicorev1.SecretList{}
	if err := corev1.Convert_core_SecretList_To_v1_SecretList(in, apiCorev1SecretList, s); err != nil {
		return err
	}
	*out = v1.SecretList(*apiCorev1SecretList)
	return nil
}

func AddConversionFuncs(s *runtime.Scheme) error {
	if err := s.AddConversionFunc((*[]newer.TagEventCondition)(nil), (*[]v1.TagEventCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_image_TagEventConditionArray_to_v1_TagEventConditionArray(a.(*[]newer.TagEventCondition), b.(*[]v1.TagEventCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]v1.TagEventCondition)(nil), (*[]newer.TagEventCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_TagEventConditionArray_to_image_TagEventConditionArray(a.(*[]v1.TagEventCondition), b.(*[]newer.TagEventCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]newer.TagEvent)(nil), (*[]v1.TagEvent)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_image_TagEventArray_to_v1_TagEventArray(a.(*[]newer.TagEvent), b.(*[]v1.TagEvent), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]v1.TagEvent)(nil), (*[]newer.TagEvent)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_TagEventArray_to_image_TagEventArray(a.(*[]v1.TagEvent), b.(*[]newer.TagEvent), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]v1.NamedTagEventList)(nil), (*map[string]newer.TagEventList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_NamedTagEventListArray_to_api_TagEventListArray(a.(*[]v1.NamedTagEventList), b.(*map[string]newer.TagEventList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*map[string]newer.TagEventList)(nil), (*[]v1.NamedTagEventList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_image_TagEventListArray_to_v1_NamedTagEventListArray(a.(*map[string]newer.TagEventList), b.(*[]v1.NamedTagEventList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*[]v1.TagReference)(nil), (*map[string]newer.TagReference)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_TagReferenceArray_to_api_TagReferenceMap(a.(*[]v1.TagReference), b.(*map[string]newer.TagReference), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*map[string]newer.TagReference)(nil), (*[]v1.TagReference)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_image_TagReferenceMap_to_v1_TagReferenceArray(a.(*map[string]newer.TagReference), b.(*[]v1.TagReference), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v1.SecretList)(nil), (*core.SecretList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_image_v1_SecretList_To_v1_SecretList(a.(*v1.SecretList), b.(*core.SecretList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*core.SecretList)(nil), (*v1.SecretList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_SecretList_To_image_v1_SecretList(a.(*core.SecretList), b.(*v1.SecretList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*runtime.RawExtension)(nil), (*image.DockerImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_runtime_RawExtension_To_image_DockerImage(a.(*runtime.RawExtension), b.(*image.DockerImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*image.DockerImage)(nil), (*runtime.RawExtension)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_image_DockerImage_To_runtime_RawExtension(a.(*image.DockerImage), b.(*runtime.RawExtension), scope)
	}); err != nil {
		return err
	}
	return nil
}

func addFieldSelectorKeyConversions(scheme *runtime.Scheme) error {
	if err := scheme.AddFieldLabelConversionFunc(v1.GroupVersion.WithKind("ImageStream"), imageStreamFieldSelectorKeyConversionFunc); err != nil {
		return err
	}
	return nil
}

// because field selectors can vary in support by version they are exposed under, we have one function for each
// groupVersion we're registering for

func imageStreamFieldSelectorKeyConversionFunc(label, value string) (internalLabel, internalValue string, err error) {
	switch label {
	case "spec.dockerImageRepository",
		"status.dockerImageRepository":
		return label, value, nil
	default:
		return runtime.DefaultMetaV1FieldSelectorConversion(label, value)
	}
}
