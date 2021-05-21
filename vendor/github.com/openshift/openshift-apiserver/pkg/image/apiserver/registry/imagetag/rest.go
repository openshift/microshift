package imagetag

import (
	"context"
	"fmt"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metainternal "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/printers"
	printerstorage "k8s.io/kubernetes/pkg/printers/storage"

	imagegroup "github.com/openshift/api/image"
	"github.com/openshift/library-go/pkg/image/imageutil"

	"github.com/openshift/openshift-apiserver/pkg/api/apihelpers"
	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation/whitelist"
	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/internalimageutil"
	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/registry/image"
	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/registry/imagestream"
	imageprinters "github.com/openshift/openshift-apiserver/pkg/image/printers/internalversion"
)

// REST implements the RESTStorage interface for ImageTag
// It only supports the Get method and is used to simplify retrieving an Image by tag from an ImageStream
type REST struct {
	imageRegistry       image.Registry
	imageStreamRegistry imagestream.Registry
	strategy            Strategy
	rest.TableConvertor
}

// NewREST returns a new REST.
func NewREST(imageRegistry image.Registry, imageStreamRegistry imagestream.Registry, registryWhitelister whitelist.RegistryWhitelister) *REST {
	return &REST{
		imageRegistry:       imageRegistry,
		imageStreamRegistry: imageStreamRegistry,
		strategy:            NewStrategy(registryWhitelister),
		TableConvertor:      printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(imageprinters.AddImageOpenShiftHandlers)},
	}
}

var _ rest.Getter = &REST{}
var _ rest.Lister = &REST{}
var _ rest.CreaterUpdater = &REST{}
var _ rest.GracefulDeleter = &REST{}
var _ rest.ShortNamesProvider = &REST{}
var _ rest.Scoper = &REST{}

// ShortNames implements the ShortNamesProvider interface. Returns a list of short names for a resource.
func (r *REST) ShortNames() []string {
	return []string{"itag"}
}

// New is only implemented to make REST implement RESTStorage
func (r *REST) New() runtime.Object {
	return &imageapi.ImageTag{}
}

// NewList returns a new list object
func (r *REST) NewList() runtime.Object {
	return &imageapi.ImageTagList{}
}

func (s *REST) NamespaceScoped() bool {
	return true
}

// nameAndTag splits a string into its name component and tag component, and returns an error
// if the string is not in the right form.
func nameAndTag(id string) (name string, tag string, err error) {
	name, tag, err = imageutil.ParseImageStreamTagName(id)
	if err != nil {
		err = kapierrors.NewBadRequest("ImageTags must be retrieved with <name>:<tag>")
	}
	return
}

func (r *REST) List(ctx context.Context, options *metainternal.ListOptions) (runtime.Object, error) {
	imageStreams, err := r.imageStreamRegistry.ListImageStreams(ctx, options)
	if err != nil {
		return nil, err
	}

	matcher := MatchImageTag(apihelpers.InternalListOptionsToSelectors(options))
	tagNames := sets.NewString()

	list := &imageapi.ImageTagList{}
	for _, currIS := range imageStreams.Items {
		// prepare a list of all possible tags, then add each one in order
		for tag := range tagNames {
			delete(tagNames, tag)
		}
		for tag := range currIS.Spec.Tags {
			tagNames.Insert(tag)
		}
		for tag := range currIS.Status.Tags {
			tagNames.Insert(tag)
		}
		for _, currTag := range tagNames.List() {
			itag, err := newITag(currTag, &currIS, nil, false)
			if err != nil {
				if kapierrors.IsNotFound(err) {
					continue
				}
				return nil, err
			}
			matches, err := matcher.Matches(itag)
			if err != nil {
				return nil, err
			}

			if matches {
				list.Items = append(list.Items, *itag)
			}
		}
	}

	return list, nil
}

// Get retrieves an image that has been tagged by stream and tag. `id` is of the format <stream name>:<tag>.
func (r *REST) Get(ctx context.Context, id string, options *metav1.GetOptions) (runtime.Object, error) {
	name, tag, err := nameAndTag(id)
	if err != nil {
		return nil, err
	}

	imageStream, err := r.imageStreamRegistry.GetImageStream(ctx, name, options)
	if err != nil {
		return nil, err
	}

	image, err := r.imageFor(ctx, tag, imageStream)
	if err != nil {
		if !kapierrors.IsNotFound(err) {
			return nil, err
		}
		image = nil
	}

	return newITag(tag, imageStream, image, false)
}

func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	itag, ok := obj.(*imageapi.ImageTag)
	if !ok {
		return nil, kapierrors.NewBadRequest(fmt.Sprintf("obj is not an ImageTag: %#v", obj))
	}
	if err := rest.BeforeCreate(r.strategy, ctx, obj); err != nil {
		return nil, err
	}
	if err := createValidation(ctx, obj.DeepCopyObject()); err != nil {
		return nil, err
	}
	namespace, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, kapierrors.NewBadRequest("a namespace must be specified to import images")
	}

	imageStreamName, imageTag, ok := imageutil.SplitImageStreamTag(itag.Name)
	if !ok {
		return nil, fmt.Errorf("%q must be of the form <stream_name>:<tag>", itag.Name)
	}

	for i := 10; i > 0; i-- {
		target, err := r.imageStreamRegistry.GetImageStream(ctx, imageStreamName, &metav1.GetOptions{})
		if err != nil {
			if !kapierrors.IsNotFound(err) {
				return nil, err
			}

			// try to create the target if it doesn't exist
			target = &imageapi.ImageStream{}
			target.Namespace = namespace
			target.Name = imageStreamName
		}

		if target.Spec.Tags == nil {
			target.Spec.Tags = make(map[string]imageapi.TagReference)
		}

		// The user wants to create a spec tag.
		_, exists := target.Spec.Tags[imageTag]
		if exists {
			return nil, kapierrors.NewAlreadyExists(imagegroup.Resource("imagetags"), itag.Name)
		}
		if itag.Spec != nil {
			target.Spec.Tags[imageTag] = *itag.Spec
		}

		// Check the stream creation timestamp and make sure we will not
		// create a new image stream while deleting.
		if target.CreationTimestamp.IsZero() {
			target, err = r.imageStreamRegistry.CreateImageStream(ctx, target, &metav1.CreateOptions{})
		} else {
			target, err = r.imageStreamRegistry.UpdateImageStream(ctx, target, false, &metav1.UpdateOptions{})
		}
		if kapierrors.IsAlreadyExists(err) || kapierrors.IsConflict(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		image, _ := r.imageFor(ctx, imageTag, target)
		return newITag(imageTag, target, image, true)
	}
	// We tried to update resource, but we kept conflicting. Inform the client that we couldn't complete
	// the operation but that they may try again.
	return nil, kapierrors.NewServerTimeout(imagegroup.Resource("imagetags"), "create", 2)
}

func (r *REST) Update(ctx context.Context, tagName string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	name, tag, err := nameAndTag(tagName)
	if err != nil {
		return nil, false, err
	}

	namespace, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, false, kapierrors.NewBadRequest("namespace is required on ImageTags")
	}

	create := false
	imageStream, err := r.imageStreamRegistry.GetImageStream(ctx, name, &metav1.GetOptions{})
	if err != nil {
		if !kapierrors.IsNotFound(err) {
			return nil, false, err
		}
		imageStream = &imageapi.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
		}
		rest.FillObjectMetaSystemFields(&imageStream.ObjectMeta)
		create = true
	}

	// create the synthetic old itag
	old, err := newITag(tag, imageStream, nil, true)
	if err != nil {
		return nil, false, err
	}

	obj, err := objInfo.UpdatedObject(ctx, old)
	if err != nil {
		return nil, false, err
	}

	itag, ok := obj.(*imageapi.ImageTag)
	if !ok {
		return nil, false, kapierrors.NewBadRequest(fmt.Sprintf("obj is not an ImageTag: %#v", obj))
	}

	// check for conflict
	switch {
	case len(itag.ResourceVersion) == 0:
		// we allow blind PUT because it is useful for the most common tag action - "I want this tag to equal this, no matter what the current value"
		itag.ResourceVersion = imageStream.ResourceVersion
	case len(imageStream.ResourceVersion) == 0:
		// image stream did not exist, cannot update
		return nil, false, kapierrors.NewNotFound(imagegroup.Resource("imagetags"), tagName)
	case imageStream.ResourceVersion != itag.ResourceVersion:
		// conflicting input and output
		return nil, false, kapierrors.NewConflict(imagegroup.Resource("imagetags"), itag.Name, fmt.Errorf("another caller has updated the resource version to %s", imageStream.ResourceVersion))
	}

	if create {
		if err := rest.BeforeCreate(r.strategy, ctx, obj); err != nil {
			return nil, false, err
		}
		if err := createValidation(ctx, obj.DeepCopyObject()); err != nil {
			return nil, false, err
		}
	} else {
		if err := rest.BeforeUpdate(r.strategy, ctx, obj, old); err != nil {
			return nil, false, err
		}
		if err := updateValidation(ctx, obj.DeepCopyObject(), old.DeepCopyObject()); err != nil {
			return nil, false, err
		}
	}

	// if !exists && itag.Spec == nil {
	// 	return nil, false, kapierrors.NewBadRequest(fmt.Sprintf("imagetag %s is not a spec or status tag in imagestream %s/%s, cannot be updated", tag, imageStream.Namespace, imageStream.Name))
	// }

	tagRef, exists := imageStream.Spec.Tags[tag]
	if itag.Spec != nil {
		if imageStream.Spec.Tags == nil {
			imageStream.Spec.Tags = map[string]imageapi.TagReference{}
		}

		tagRef = *itag.Spec
		tagRef.Name = tag
		imageStream.Spec.Tags[tag] = tagRef
	} else {
		// TODO: should error if it doesn't already exist?
		delete(imageStream.Spec.Tags, tag)
	}

	// mutate the image stream
	var newImageStream *imageapi.ImageStream
	if create {
		newImageStream, err = r.imageStreamRegistry.CreateImageStream(ctx, imageStream, &metav1.CreateOptions{})
	} else {
		newImageStream, err = r.imageStreamRegistry.UpdateImageStream(ctx, imageStream, false, &metav1.UpdateOptions{})
	}
	if err != nil {
		return nil, false, err
	}

	image, err := r.imageFor(ctx, tag, newImageStream)
	if err != nil {
		if !kapierrors.IsNotFound(err) {
			return nil, false, err
		}
	}

	newITag, err := newITag(tag, newImageStream, image, true)
	return newITag, !exists, err
}

// Delete removes a tag from a stream. `id` is of the format <stream name>:<tag>.
// The associated image that the tag points to is *not* deleted.
// The tag history is removed.
func (r *REST) Delete(ctx context.Context, id string, objectFunc rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	name, tag, err := nameAndTag(id)
	if err != nil {
		return nil, false, err
	}

	for i := 10; i > 0; i-- {
		stream, err := r.imageStreamRegistry.GetImageStream(ctx, name, &metav1.GetOptions{})
		if err != nil {
			return nil, false, err
		}
		if options != nil {
			if pre := options.Preconditions; pre != nil {
				if pre.UID != nil && *pre.UID != stream.UID {
					return nil, false, kapierrors.NewConflict(imagegroup.Resource("imagetags"), id, fmt.Errorf("the UID precondition was not met"))
				}
			}
		}

		notFound := true

		// Try to delete the status tag
		if _, ok := stream.Status.Tags[tag]; ok {
			delete(stream.Status.Tags, tag)
			notFound = false
		}

		// Try to delete the spec tag
		if _, ok := stream.Spec.Tags[tag]; ok {
			delete(stream.Spec.Tags, tag)
			notFound = false
		}

		if notFound {
			return nil, false, kapierrors.NewNotFound(imagegroup.Resource("imagetags"), id)
		}

		_, err = r.imageStreamRegistry.UpdateImageStream(ctx, stream, false, &metav1.UpdateOptions{})
		if kapierrors.IsConflict(err) {
			continue
		}
		if err != nil && !kapierrors.IsNotFound(err) {
			return nil, false, err
		}
		return &metav1.Status{Status: metav1.StatusSuccess}, true, nil
	}
	// We tried to update resource, but we kept conflicting. Inform the client that we couldn't complete
	// the operation but that they may try again.
	return nil, false, kapierrors.NewServerTimeout(imagegroup.Resource("imagetags"), "delete", 2)
}

// imageFor retrieves the most recent image for a tag in a given imageStreem.
func (r *REST) imageFor(ctx context.Context, tag string, imageStream *imageapi.ImageStream) (*imageapi.Image, error) {
	event := internalimageutil.LatestTaggedImage(imageStream, tag)
	if event == nil || len(event.Image) == 0 {
		return nil, kapierrors.NewNotFound(imagegroup.Resource("imagetags"), imageutil.JoinImageStreamTag(imageStream.Name, tag))
	}

	return r.imageRegistry.GetImage(ctx, event.Image, &metav1.GetOptions{})
}

// newITag initializes an image tag from an image stream and image.
func newITag(tag string, imageStream *imageapi.ImageStream, image *imageapi.Image, allowEmpty bool) (*imageapi.ImageTag, error) {
	itagName := imageutil.JoinImageStreamTag(imageStream.Name, tag)

	itag := &imageapi.ImageTag{
		ObjectMeta: imageStream.ObjectMeta,
	}
	itag.Name = itagName

	var event *imageapi.TagEvent
	for name, tagEvents := range imageStream.Status.Tags {
		if name != tag {
			continue
		}
		itag.Status = &imageapi.NamedTagEventList{
			Tag:        name,
			Conditions: tagEvents.Conditions,
			Items:      tagEvents.Items,
		}
		if len(tagEvents.Items) > 0 {
			event = &tagEvents.Items[0]
		}
		break
	}

	if tagRef, ok := imageStream.Spec.Tags[tag]; ok {
		copiedTagRef := tagRef
		itag.Spec = &copiedTagRef
	}

	if image != nil {
		if err := internalimageutil.InternalImageWithMetadata(image); err != nil {
			return nil, err
		}
		image.DockerImageManifest = ""
		image.DockerImageConfig = ""
		itag.Image = image
		itag.Image.DockerImageReference = internalimageutil.ResolveReferenceForTagEvent(imageStream, tag, event)
	}

	if !allowEmpty && itag.Spec == nil && itag.Status == nil {
		return nil, kapierrors.NewNotFound(imagegroup.Resource("imagetags"), itagName)
	}

	return itag, nil
}
