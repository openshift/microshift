package imagestream

import (
	"context"
	"fmt"
	"strings"

	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/internalimageutil"

	authorizationapi "k8s.io/api/authorization/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/authentication/user"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/storage/names"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	kapi "k8s.io/kubernetes/pkg/apis/core"
	kapihelper "k8s.io/kubernetes/pkg/apis/core/helper"

	"github.com/openshift/library-go/pkg/authorization/authorizationutil"
	"github.com/openshift/library-go/pkg/image/reference"
	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation"
	"github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation/whitelist"
	imageadmission "github.com/openshift/openshift-apiserver/pkg/image/apiserver/admission/limitrange"
	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/registryhostname"
)

type ResourceGetter interface {
	Get(context.Context, string, *metav1.GetOptions) (runtime.Object, error)
}

// Strategy implements behavior for ImageStreams.
type Strategy struct {
	runtime.ObjectTyper
	names.NameGenerator
	registryHostnameRetriever registryhostname.RegistryHostnameRetriever
	tagVerifier               *TagVerifier
	limitVerifier             imageadmission.LimitVerifier
	registryWhitelister       whitelist.RegistryWhitelister
	imageStreamGetter         ResourceGetter
}

// NewStrategy is the default logic that applies when creating and updating
// ImageStream objects via the REST API.
func NewStrategy(
	registryHostname registryhostname.RegistryHostnameRetriever,
	subjectAccessReviewClient authorizationclient.SubjectAccessReviewInterface,
	limitVerifier imageadmission.LimitVerifier,
	registryWhitelister whitelist.RegistryWhitelister,
	imageStreamGetter ResourceGetter,
) Strategy {
	return Strategy{
		ObjectTyper:               legacyscheme.Scheme,
		NameGenerator:             names.SimpleNameGenerator,
		registryHostnameRetriever: registryHostname,
		tagVerifier:               &TagVerifier{subjectAccessReviewClient},
		limitVerifier:             limitVerifier,
		registryWhitelister:       registryWhitelister,
		imageStreamGetter:         imageStreamGetter,
	}
}

// NamespaceScoped is true for image streams.
func (s Strategy) NamespaceScoped() bool {
	return true
}

// collapseEmptyStatusTags removes status tags that are completely empty.
func collapseEmptyStatusTags(stream *imageapi.ImageStream) {
	for tag, ref := range stream.Status.Tags {
		if len(ref.Items) == 0 && len(ref.Conditions) == 0 {
			delete(stream.Status.Tags, tag)
		}
	}
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (s Strategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	stream := obj.(*imageapi.ImageStream)
	stream.Status = imageapi.ImageStreamStatus{
		DockerImageRepository:       s.dockerImageRepository(ctx, stream, false),
		PublicDockerImageRepository: s.publicDockerImageRepository(stream),
		Tags:                        make(map[string]imageapi.TagEventList),
	}
	stream.Generation = 1
	for tag, ref := range stream.Spec.Tags {
		ref.Generation = &stream.Generation
		stream.Spec.Tags[tag] = ref
	}
	collapseEmptyStatusTags(stream)
}

// Validate validates a new image stream and verifies the current user is
// authorized to access any image streams newly referenced in spec.tags.
func (s Strategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	stream := obj.(*imageapi.ImageStream)
	var errs field.ErrorList
	if err := s.validateTagsAndLimits(ctx, nil, stream); err != nil {
		errs = append(errs, field.InternalError(field.NewPath(""), err))
	}
	errs = append(errs, validation.ValidateImageStreamWithWhitelister(ctx, s.registryWhitelister, stream)...)
	return errs
}

func (s Strategy) validateTagsAndLimits(ctx context.Context, oldStream, newStream *imageapi.ImageStream) error {
	user, ok := apirequest.UserFrom(ctx)
	if !ok {
		return kerrors.NewForbidden(schema.GroupResource{Resource: "imagestreams"}, newStream.Name, fmt.Errorf("no user context available"))
	}

	errs := s.tagVerifier.Verify(ctx, oldStream, newStream, user)
	errs = append(errs, s.tagsChanged(ctx, oldStream, newStream)...)
	if len(errs) > 0 {
		return kerrors.NewInvalid(schema.GroupKind{Kind: "imagestreams"}, newStream.Name, errs)
	}

	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		ns = newStream.Namespace
	}
	return s.limitVerifier.VerifyLimits(ns, newStream)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (Strategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// AllowCreateOnUpdate is false for image streams.
func (s Strategy) AllowCreateOnUpdate() bool {
	return false
}

func (Strategy) AllowUnconditionalUpdate() bool {
	return false
}

// dockerImageRepository determines the container image stream for stream.
// If stream.DockerImageRepository is set, that value is returned. Otherwise,
// if a default registry exists, the value returned is of the form
// <default registry>/<namespace>/<stream name>.
func (s Strategy) dockerImageRepository(ctx context.Context, stream *imageapi.ImageStream, allowNamespaceDefaulting bool) string {
	registry, ok := s.registryHostnameRetriever.InternalRegistryHostname(ctx)
	if !ok {
		return stream.Spec.DockerImageRepository
	}

	if len(stream.Namespace) == 0 && allowNamespaceDefaulting {
		stream.Namespace = metav1.NamespaceDefault
	}
	ref := imageapi.DockerImageReference{
		Registry:  registry,
		Namespace: stream.Namespace,
		Name:      stream.Name,
	}
	return ref.String()
}

// publicDockerImageRepository determines the public location of given image
// stream. If the ExternalRegistryHostname is set in the master config, the
// value of this property is used as a hostname part for the container image
// reference.
func (s Strategy) publicDockerImageRepository(stream *imageapi.ImageStream) string {
	externalHostname, ok := s.registryHostnameRetriever.ExternalRegistryHostname()
	if !ok {
		return ""
	}
	ref := imageapi.DockerImageReference{
		Registry:  externalHostname,
		Namespace: stream.Namespace,
		Name:      stream.Name,
	}
	return ref.String()
}

func parseFromReference(stream *imageapi.ImageStream, from *kapi.ObjectReference) (string, string, error) {
	splitChar := ""
	refType := ""

	switch from.Kind {
	case "ImageStreamTag":
		splitChar = ":"
		refType = "tag"
	case "ImageStreamImage":
		splitChar = "@"
		refType = "id"
	default:
		return "", "", fmt.Errorf("invalid from.kind %q - only ImageStreamTag and ImageStreamImage are allowed", from.Kind)
	}

	parts := strings.Split(from.Name, splitChar)
	switch len(parts) {
	case 1:
		// <tag> or <id>
		return stream.Name, from.Name, nil
	case 2:
		// <stream>:<tag> or <stream>@<id>
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("invalid from.name %q - it must be of the form <%s> or <stream>%s<%s>", from.Name, refType, splitChar, refType)
	}
}

// tagsChanged updates stream.Status.Tags based on the old and new image stream.
// if the old stream is nil, all tags are considered additions.
func (s Strategy) tagsChanged(ctx context.Context, old, stream *imageapi.ImageStream) field.ErrorList {
	internalRegistry, hasInternalRegistry := s.registryHostnameRetriever.InternalRegistryHostname(ctx)

	var errs field.ErrorList

	oldTags := map[string]imageapi.TagReference{}
	if old != nil && old.Spec.Tags != nil {
		oldTags = old.Spec.Tags
	}

	for tag, tagRef := range stream.Spec.Tags {
		if oldRef, ok := oldTags[tag]; ok && !tagRefChanged(oldRef, tagRef, stream.Namespace) {
			continue
		}

		if tagRef.From == nil {
			continue
		}

		klog.V(5).Infof("Detected changed tag %s in %s/%s", tag, stream.Namespace, stream.Name)

		generation := stream.Generation
		tagRef.Generation = &generation

		fromPath := field.NewPath("spec", "tags").Key(tag).Child("from")
		if tagRef.From.Kind == "DockerImage" && len(tagRef.From.Name) > 0 {
			if tagRef.Reference {
				event, err := tagReferenceToTagEvent(stream, tagRef, "")
				if err != nil {
					errs = append(errs, field.Invalid(fromPath, tagRef.From, err.Error()))
					continue
				}
				stream.Spec.Tags[tag] = tagRef
				internalimageutil.AddTagEventToImageStream(stream, tag, *event)
			}
			continue
		}

		tagRefStreamName, tagOrID, err := parseFromReference(stream, tagRef.From)
		if err != nil {
			errs = append(errs, field.Invalid(fromPath.Child("name"), tagRef.From.Name, "must be of the form <tag>, <repo>:<tag>, <id>, or <repo>@<id>"))
			continue
		}

		streamRef := stream
		streamRefNamespace := tagRef.From.Namespace
		if len(streamRefNamespace) == 0 {
			streamRefNamespace = stream.Namespace
		}
		if streamRefNamespace != stream.Namespace || tagRefStreamName != stream.Name {
			obj, err := s.imageStreamGetter.Get(apirequest.WithNamespace(apirequest.NewContext(), streamRefNamespace), tagRefStreamName, &metav1.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					errs = append(errs, field.NotFound(fromPath.Child("name"), tagRef.From.Name))
				} else {
					errs = append(errs, field.Invalid(fromPath.Child("name"), tagRef.From.Name, fmt.Sprintf("unable to retrieve image stream: %v", err)))
				}
				continue
			}

			streamRef = obj.(*imageapi.ImageStream)
		}

		event, err := tagReferenceToTagEvent(streamRef, tagRef, tagOrID)
		if err != nil {
			errs = append(errs, field.Invalid(fromPath.Child("name"), tagRef.From.Name, fmt.Sprintf("error generating tag event: %v", err)))
			continue
		}
		if event == nil {
			// referenced tag or ID doesn't exist, which is ok
			continue
		}

		// if this is not a reference tag, and the tag points to the internal registry for the other namespace, alter it to
		// point to this stream so that pulls happen from this stream in the future.
		if !tagRef.Reference {
			if ref, err := reference.Parse(event.DockerImageReference); err == nil {
				if hasInternalRegistry && ref.Registry == internalRegistry && ref.Namespace == streamRef.Namespace && ref.Name == streamRef.Name {
					ref.Namespace = stream.Namespace
					ref.Name = stream.Name
					event.DockerImageReference = ref.Exact()
				}
			}
		}

		stream.Spec.Tags[tag] = tagRef
		internalimageutil.AddTagEventToImageStream(stream, tag, *event)
	}

	updateChangedTrackingTags(stream, old)

	// use a consistent timestamp on creation
	if old == nil && !stream.CreationTimestamp.IsZero() {
		for tag, list := range stream.Status.Tags {
			for _, event := range list.Items {
				event.Created = stream.CreationTimestamp
			}
			stream.Status.Tags[tag] = list
		}
	}

	return errs
}

// UpdateChangedTrackingTags identifies any tags in the status that have changed and
// ensures any referenced tracking tags are also updated. It returns the number of
// updates applied.
func updateChangedTrackingTags(new, old *imageapi.ImageStream) int {
	changes := 0
	for newTag, newImages := range new.Status.Tags {
		if len(newImages.Items) == 0 {
			continue
		}
		if old != nil {
			oldImages := old.Status.Tags[newTag]
			changed, deleted := tagsChanged(newImages.Items, oldImages.Items)
			if !changed || deleted {
				continue
			}
		}
		changes += internalimageutil.UpdateTrackingTags(new, newTag, newImages.Items[0])
	}
	return changes
}

// tagsChanged returns true if the two lists differ, and if the newer list is empty
// then deleted is returned true as well.
func tagsChanged(new, old []imageapi.TagEvent) (changed bool, deleted bool) {
	switch {
	case len(old) == 0 && len(new) == 0:
		return false, false
	case len(new) == 0:
		return true, true
	case len(old) == 0:
		return true, false
	default:
		return new[0] != old[0], false
	}
}

func tagReferenceToTagEvent(stream *imageapi.ImageStream, tagRef imageapi.TagReference, tagOrID string) (*imageapi.TagEvent, error) {
	var (
		event *imageapi.TagEvent
		err   error
	)
	switch tagRef.From.Kind {
	case "DockerImage":
		event = &imageapi.TagEvent{
			Created:              metav1.Now(),
			DockerImageReference: tagRef.From.Name,
		}

	case "ImageStreamImage":
		event, err = internalimageutil.ResolveImageID(stream, tagOrID)
	case "ImageStreamTag":
		event, err = internalimageutil.LatestTaggedImage(stream, tagOrID), nil
	default:
		err = fmt.Errorf("invalid from.kind %q: it must be DockerImage, ImageStreamImage or ImageStreamTag", tagRef.From.Kind)
	}
	if err != nil {
		return nil, err
	}
	if event != nil && tagRef.Generation != nil {
		event.Generation = *tagRef.Generation
	}
	return event, nil
}

// tagRefChanged returns true if the tag ref changed between two spec updates.
func tagRefChanged(old, next imageapi.TagReference, streamNamespace string) bool {
	if next.From == nil {
		// both fields in next are empty
		return false
	}
	if len(next.From.Kind) == 0 && len(next.From.Name) == 0 {
		// invalid
		return false
	}
	oldFrom := old.From
	if oldFrom == nil {
		oldFrom = &kapi.ObjectReference{}
	}
	oldNamespace := oldFrom.Namespace
	if len(oldNamespace) == 0 {
		oldNamespace = streamNamespace
	}
	nextNamespace := next.From.Namespace
	if len(nextNamespace) == 0 {
		nextNamespace = streamNamespace
	}
	if oldNamespace != nextNamespace {
		return true
	}
	if oldFrom.Name != next.From.Name {
		return true
	}
	if old.Reference != next.Reference {
		return true
	}
	return tagRefGenerationChanged(old, next)
}

// tagRefGenerationChanged returns true if and only the values were set and the new generation
// is at zero.
func tagRefGenerationChanged(old, next imageapi.TagReference) bool {
	switch {
	case old.Generation != nil && next.Generation != nil:
		if *old.Generation == *next.Generation {
			return false
		}
		if *next.Generation == 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func tagEventChanged(old, next imageapi.TagEvent) bool {
	return old.Image != next.Image || old.DockerImageReference != next.DockerImageReference || old.Generation > next.Generation
}

// updateSpecTagGenerationsForUpdate ensures that new spec updates always have a generation set, and that the value
// cannot be updated by an end user (except by setting generation 0).
func updateSpecTagGenerationsForUpdate(stream, oldStream *imageapi.ImageStream) {
	for tag, ref := range stream.Spec.Tags {
		if ref.Generation != nil && *ref.Generation == 0 {
			continue
		}
		if oldRef, ok := oldStream.Spec.Tags[tag]; ok {
			ref.Generation = oldRef.Generation
			stream.Spec.Tags[tag] = ref
		}
	}
}

// ensureSpecTagGenerationsAreSet ensures that all spec tags have a generation set to either 0 or the
// current stream value.
func ensureSpecTagGenerationsAreSet(stream, oldStream *imageapi.ImageStream) {
	oldTags := map[string]imageapi.TagReference{}
	if oldStream != nil && oldStream.Spec.Tags != nil {
		oldTags = oldStream.Spec.Tags
	}

	// set the generation for any spec tags that have changed, are nil, or are zero
	for tag, ref := range stream.Spec.Tags {
		if oldRef, ok := oldTags[tag]; !ok || tagRefChanged(oldRef, ref, stream.Namespace) {
			ref.Generation = nil
		}

		if ref.Generation != nil && *ref.Generation != 0 {
			continue
		}
		ref.Generation = &stream.Generation
		stream.Spec.Tags[tag] = ref
	}
}

// updateObservedGenerationForStatusUpdate ensures every status item has a generation set.
func updateObservedGenerationForStatusUpdate(stream, oldStream *imageapi.ImageStream) {
	for tag, newer := range stream.Status.Tags {
		if len(newer.Items) == 0 || newer.Items[0].Generation != 0 {
			// generation is set, continue
			continue
		}

		older := oldStream.Status.Tags[tag]
		if len(older.Items) == 0 || !tagEventChanged(older.Items[0], newer.Items[0]) {
			// if the tag wasn't changed by the status update
			newer.Items[0].Generation = stream.Generation
			stream.Status.Tags[tag] = newer
			continue
		}

		spec, ok := stream.Spec.Tags[tag]
		if !ok || spec.Generation == nil {
			// if the spec tag has no generation
			newer.Items[0].Generation = stream.Generation
			stream.Status.Tags[tag] = newer
			continue
		}

		// set the status tag from the spec tag generation
		newer.Items[0].Generation = *spec.Generation
		stream.Status.Tags[tag] = newer
	}
}

type TagVerifier struct {
	subjectAccessReviewClient authorizationclient.SubjectAccessReviewInterface
}

func (v *TagVerifier) Verify(ctx context.Context, old, stream *imageapi.ImageStream, user user.Info) field.ErrorList {
	var errors field.ErrorList
	oldTags := map[string]imageapi.TagReference{}
	if old != nil && old.Spec.Tags != nil {
		oldTags = old.Spec.Tags
	}
	// Store tags by the ImageStreamTags they reference on creation to minimize the number of SAR checks to make
	tagsByNamespacedStream := make(map[string]map[string][]string)
	for tag, tagRef := range stream.Spec.Tags {
		if tagRef.From == nil {
			continue
		}
		if len(tagRef.From.Namespace) == 0 {
			continue
		}
		if stream.Namespace == tagRef.From.Namespace {
			continue
		}
		if oldRef, ok := oldTags[tag]; ok && !tagRefChanged(oldRef, tagRef, stream.Namespace) {
			continue
		}

		streamName, _, err := parseFromReference(stream, tagRef.From)
		fromPath := field.NewPath("spec", "tags").Key(tag).Child("from")
		if err != nil {
			errors = append(errors, field.Invalid(fromPath.Child("name"), tagRef.From.Name, "must be of the form <tag>, <repo>:<tag>, <id>, or <repo>@<id>"))
			continue
		}

		mapping, ok := tagsByNamespacedStream[tagRef.From.Namespace]
		if !ok {
			mapping = make(map[string][]string)
		}
		mapping[streamName] = append(mapping[streamName], tag)
		tagsByNamespacedStream[tagRef.From.Namespace] = mapping
	}
	for namespace, mapping := range tagsByNamespacedStream {
		for streamName, tags := range mapping {
			// Make sure this user can pull the specified image before allowing them to tag it into another imagestream
			subjectAccessReview := authorizationutil.AddUserToSAR(user, &authorizationapi.SubjectAccessReview{
				Spec: authorizationapi.SubjectAccessReviewSpec{
					ResourceAttributes: &authorizationapi.ResourceAttributes{
						Namespace:   namespace,
						Verb:        "get",
						Group:       imageapi.GroupName,
						Resource:    "imagestreams",
						Subresource: "layers",
						Name:        streamName,
					},
				},
			})
			klog.V(4).Infof("Performing SubjectAccessReview for user=%s, groups=%v to %s/%s", user.GetName(), user.GetGroups(), namespace, streamName)
			resp, err := v.subjectAccessReviewClient.Create(ctx, subjectAccessReview, metav1.CreateOptions{})
			if err != nil || resp == nil || (resp != nil && !resp.Status.Allowed) {
				message := fmt.Sprintf("%s/%s", namespace, streamName)
				if resp != nil {
					message = message + fmt.Sprintf(": %q %q", resp.Status.Reason, resp.Status.EvaluationError)
				}
				if err != nil {
					message = message + fmt.Sprintf("- %v", err)
				}
				for _, tag := range tags {
					fromPath := field.NewPath("spec", "tags").Key(tag).Child("from")
					errors = append(errors, field.Forbidden(fromPath, message))
				}
				continue
			}
		}
	}
	return errors
}

// Canonicalize normalizes the object after validation.
func (Strategy) Canonicalize(obj runtime.Object) {
}

func (s Strategy) prepareForUpdate(ctx context.Context, obj, old runtime.Object, resetStatus bool) {
	oldStream := old.(*imageapi.ImageStream)
	stream := obj.(*imageapi.ImageStream)

	collapseEmptyStatusTags(stream)
	stream.Generation = oldStream.Generation
	if resetStatus {
		stream.Status = oldStream.Status
	}
	stream.Status.DockerImageRepository = s.dockerImageRepository(ctx, stream, true)
	stream.Status.PublicDockerImageRepository = s.publicDockerImageRepository(stream)

	// ensure that users cannot change spec tag generation to any value except 0
	updateSpecTagGenerationsForUpdate(stream, oldStream)

	// Any changes to the spec increment the generation number.
	//
	// TODO: Any changes to a part of the object that represents desired state (labels,
	// annotations etc) should also increment the generation.
	if !kapihelper.Semantic.DeepEqual(oldStream.Spec, stream.Spec) || stream.Generation == 0 {
		stream.Generation = oldStream.Generation + 1
	}

	// default spec tag generations afterwards (to avoid updating the generation for legacy objects)
	ensureSpecTagGenerationsAreSet(stream, oldStream)
}

func (s Strategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	s.prepareForUpdate(ctx, obj, old, true)
}

// ValidateUpdate is the default update validation for an end user.
func (s Strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	stream := obj.(*imageapi.ImageStream)
	oldStream := old.(*imageapi.ImageStream)
	var errs field.ErrorList
	if err := s.validateTagsAndLimits(ctx, oldStream, stream); err != nil {
		errs = append(errs, field.InternalError(field.NewPath(""), err))
	}
	errs = append(errs, validation.ValidateImageStreamUpdateWithWhitelister(ctx, s.registryWhitelister, stream, oldStream)...)
	return errs
}

// WarningsOnUpdate returns warnings for the given update.
func (Strategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Decorate decorates stream.Status.DockerImageRepository using the logic from
// dockerImageRepository().
func (s Strategy) Decorate(obj runtime.Object) {
	switch t := obj.(type) {
	case *imageapi.ImageStream:
		t.Status.DockerImageRepository = s.dockerImageRepository(context.TODO(), t, true)
		t.Status.PublicDockerImageRepository = s.publicDockerImageRepository(t)
	case *imageapi.ImageStreamList:
		for i := range t.Items {
			is := &t.Items[i]
			is.Status.DockerImageRepository = s.dockerImageRepository(context.TODO(), is, true)
			is.Status.PublicDockerImageRepository = s.publicDockerImageRepository(is)
		}
	default:
		// This was not an object we can decorate.  This is not an error, as the
		// caching layer can pass through here, too.
	}
}

type StatusStrategy struct {
	Strategy
}

// NewStatusStrategy creates a status update strategy around an existing stream
// strategy.
func NewStatusStrategy(strategy Strategy) StatusStrategy {
	return StatusStrategy{strategy}
}

// Canonicalize normalizes the object after validation.
func (StatusStrategy) Canonicalize(obj runtime.Object) {
}

func (StatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	oldStream := old.(*imageapi.ImageStream)
	stream := obj.(*imageapi.ImageStream)

	stream.Spec.Tags = oldStream.Spec.Tags
	stream.Spec.DockerImageRepository = oldStream.Spec.DockerImageRepository

	updateObservedGenerationForStatusUpdate(stream, oldStream)
}

func (s StatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newIS := obj.(*imageapi.ImageStream)
	errs := field.ErrorList{}

	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		ns = newIS.Namespace
	}
	err := s.limitVerifier.VerifyLimits(ns, newIS)
	if err != nil {
		errs = append(errs, field.Forbidden(field.NewPath("imageStream"), err.Error()))
	}

	// TODO: merge valid fields after update
	errs = append(errs, validation.ValidateImageStreamStatusUpdateWithWhitelister(ctx, s.registryWhitelister, newIS, old.(*imageapi.ImageStream))...)
	return errs
}

// InternalStrategy implements behavior for updating both the spec and status
// of an image stream
type InternalStrategy struct {
	Strategy
}

// NewInternalStrategy creates an update strategy around an existing stream
// strategy.
func NewInternalStrategy(strategy Strategy) InternalStrategy {
	return InternalStrategy{strategy}
}

// Canonicalize normalizes the object after validation.
func (InternalStrategy) Canonicalize(obj runtime.Object) {
}

func (s InternalStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	stream := obj.(*imageapi.ImageStream)

	stream.Status.DockerImageRepository = s.dockerImageRepository(ctx, stream, false)
	stream.Status.PublicDockerImageRepository = s.publicDockerImageRepository(stream)
	stream.Generation = 1
	for tag, ref := range stream.Spec.Tags {
		ref.Generation = &stream.Generation
		stream.Spec.Tags[tag] = ref
	}
}

func (s InternalStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	s.prepareForUpdate(ctx, obj, old, false)
}
