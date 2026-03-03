package assets

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

type RenderedManifests []RenderedManifest

type RenderedManifest struct {
	OriginalFilename string
	Content          []byte

	// use GetDecodedObj to access
	decodedObj runtime.Object
}

func (renderedManifests RenderedManifests) ListManifestOfType(gvk schema.GroupVersionKind) []RenderedManifest {
	ret := []RenderedManifest{}
	for i := range renderedManifests {
		obj, err := renderedManifests[i].GetDecodedObj()
		if err != nil {
			klog.Warningf("failure to read %q: %v", renderedManifests[i].OriginalFilename, err)
			continue
		}
		if obj.GetObjectKind().GroupVersionKind() == gvk {
			ret = append(ret, renderedManifests[i])
		}
	}

	return ret
}

func (renderedManifests RenderedManifests) GetManifest(gvk schema.GroupVersionKind, namespace, name string) (RenderedManifest, error) {
	for i := range renderedManifests {
		obj, err := renderedManifests[i].GetDecodedObj()
		if err != nil {
			klog.Warningf("failure to read %q: %v", renderedManifests[i].OriginalFilename, err)
			continue
		}
		if obj.GetObjectKind().GroupVersionKind() != gvk {
			continue
		}
		objMetadata, err := meta.Accessor(obj)
		if err != nil {
			klog.Warningf("failure to read metadata %q: %v", renderedManifests[i].OriginalFilename, err)
			continue
		}

		// since validation requires that all of these are the same, it doesn't matterwhich one we return
		if objMetadata.GetName() == name && objMetadata.GetNamespace() == namespace {
			return renderedManifests[i], nil
		}
	}

	return RenderedManifest{}, apierrors.NewNotFound(
		schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		},
		name)
}

func (renderedManifests RenderedManifests) GetObject(gvk schema.GroupVersionKind, namespace, name string) (runtime.Object, error) {
	manifest, err := renderedManifests.GetManifest(gvk, namespace, name)
	if err != nil {
		return nil, err
	}
	return manifest.GetDecodedObj()
}

var localScheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(localScheme)

func (c *RenderedManifest) GetDecodedObj() (runtime.Object, error) {
	if c.decodedObj != nil {
		return c.decodedObj, nil
	}

	udi, _, err := codecs.UniversalDecoder().Decode(c.Content, nil, &unstructured.Unstructured{})
	if err != nil {
		return nil, fmt.Errorf("unable to decode %q: %w", c.OriginalFilename, err)
	}
	c.decodedObj = udi

	return c.decodedObj, nil
}
