package importer

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/containers/image/pkg/sysregistriesv2"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/api/errcode"
	v2 "github.com/docker/distribution/registry/api/v2"
	godigest "github.com/opencontainers/go-digest"

	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"

	"github.com/openshift/api/image"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/library-go/pkg/image/imageutil"
	imageref "github.com/openshift/library-go/pkg/image/reference"
	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	"github.com/openshift/openshift-apiserver/pkg/image/apiserver/internalimageutil"
)

// Interface loads images into an image stream import request.
type Interface interface {
	Import(ctx context.Context, isi *imageapi.ImageStreamImport, stream *imageapi.ImageStream) error
}

// RepositoryRetriever fetches a Docker distribution.Repository.
type RepositoryRetriever interface {
	// Repository returns a properly authenticated distribution.Repository for the given
	// docker image reference and insecure toleration behavior.
	Repository(ctx context.Context, ref imageref.DockerImageReference, insecure bool) (distribution.Repository, error)
}

// ImageStreamImporter implements an import strategy for container images. It keeps a cache of images
// per distinct auth context to reduce duplicate loads. This type is not thread safe.
type ImageStreamImporter struct {
	maximumTagsPerRepo int

	retriever RepositoryRetriever
	limiter   flowcontrol.RateLimiter
	regConf   *sysregistriesv2.V2RegistriesConf

	digestToRepositoryCache map[context.Context]map[manifestKey]*imageapi.Image

	// digestToLayerSizeCache maps layer digests to size.
	digestToLayerSizeCache *ImageStreamLayerCache
}

// NewImageStreamImporter creates an importer that will load images from a remote container image
// registry into an ImageStreamImport object. Limiter may be nil.
func NewImageStreamImporter(
	retriever RepositoryRetriever,
	regConf *sysregistriesv2.V2RegistriesConf,
	maximumTagsPerRepo int,
	limiter flowcontrol.RateLimiter,
	cache *ImageStreamLayerCache,
) *ImageStreamImporter {
	if limiter == nil {
		limiter = flowcontrol.NewFakeAlwaysRateLimiter()
	}
	if cache == nil {
		klog.V(5).Infof("the global layer cache is disabled")
	}
	return &ImageStreamImporter{
		maximumTagsPerRepo: maximumTagsPerRepo,

		retriever: retriever,
		limiter:   limiter,
		regConf:   regConf,

		// XXX The context used to index here is the client request's context. We might
		// be able to get rid of this index at all as ImageStreamImporter is instantiated
		// once per request.
		digestToRepositoryCache: make(map[context.Context]map[manifestKey]*imageapi.Image),
		digestToLayerSizeCache:  cache,
	}
}

// Import tries to complete the provided isi object with images loaded from remote registries.
func (imp *ImageStreamImporter) Import(ctx context.Context, isi *imageapi.ImageStreamImport, stream *imageapi.ImageStream) error {
	// Initialize layer size cache if not given.
	if imp.digestToLayerSizeCache == nil {
		cache, err := NewImageStreamLayerCache(DefaultImageStreamLayerCacheSize)
		if err != nil {
			return err
		}
		imp.digestToLayerSizeCache = &cache
	}
	// Initialize the image cache entry for a context.
	if _, ok := imp.digestToRepositoryCache[ctx]; !ok {
		imp.digestToRepositoryCache[ctx] = make(map[manifestKey]*imageapi.Image)
	}

	imp.importImages(ctx, isi, stream)
	imp.importFromRepository(ctx, isi)
	return nil
}

// allowRegistryInsecureAccess returns true if access to image may be done insecurely (ignoring
// invalid certificate). Returns true if either global registries configuration (regConf) or
// image specific policy is true.
func (imp *ImageStreamImporter) allowRegistryInsecureAccess(
	policy imageapi.TagImportPolicy, ref imageapi.DockerImageReference,
) bool {
	if imp.regConf == nil {
		return policy.Insecure
	}
	regURL := ref.DockerClientDefaults().Registry
	for _, reg := range imp.regConf.Registries {
		if reg.Location == regURL {
			return policy.Insecure || reg.Insecure
		}
	}
	return policy.Insecure
}

// blockedRegistry returns if registry hosting the image to be imported has been blocked.
func (imp *ImageStreamImporter) blockedRegistry(ref imageapi.DockerImageReference) bool {
	if imp.regConf == nil {
		return false
	}
	regURL := ref.DockerClientDefaults().Registry
	for _, reg := range imp.regConf.Registries {
		if reg.Location == regURL {
			return reg.Blocked
		}
	}
	return false
}

// importImages updates the passed ImageStreamImport object and sets Status for each image based on
// whether the import succeeded or failed. Cache is updated with any loaded images.
func (imp *ImageStreamImporter) importImages(ctx context.Context, isi *imageapi.ImageStreamImport, stream *imageapi.ImageStream) {
	tags := make(map[manifestKey][]int)
	ids := make(map[manifestKey][]int)
	repositories := make(map[repositoryKey]*importRepository)
	cache := imp.digestToRepositoryCache[ctx]

	isi.Status.Images = make([]imageapi.ImageImportStatus, len(isi.Spec.Images))
	for i := range isi.Spec.Images {
		spec := &isi.Spec.Images[i]
		from := spec.From
		if from.Kind != "DockerImage" {
			continue
		}
		// TODO: This should be removed in 1.6
		// See for more info: https://github.com/openshift/origin/pull/11774#issuecomment-258905994
		var (
			err error
			ref imageapi.DockerImageReference
		)
		if from.Name != "*" {
			ref, err = imageref.Parse(from.Name)
			if err != nil {
				isi.Status.Images[i].Status = invalidStatus("", field.Invalid(field.NewPath("from", "name"), from.Name, fmt.Sprintf("invalid name: %v", err)))
				continue
			}
		} else {
			ref = imageapi.DockerImageReference{Name: from.Name}
		}

		if imp.blockedRegistry(ref) {
			isi.Status.Images[i].Status = forbiddenStatus(
				fmt.Errorf("registry %s blocked", ref.Registry),
			)
			continue
		}

		defaultRef := ref.DockerClientDefaults()
		repoName := defaultRef.RepositoryName()
		registryURL := defaultRef.RegistryURL()

		key := repositoryKey{url: *registryURL, name: repoName}
		repo, ok := repositories[key]
		if !ok {
			repo = &importRepository{
				Ref:      ref,
				Registry: &key.url,
				Name:     key.name,
				Insecure: imp.allowRegistryInsecureAccess(spec.ImportPolicy, ref),
			}
			repositories[key] = repo
		}

		if len(defaultRef.ID) > 0 {
			id := manifestKey{repositoryKey: key}
			id.value = defaultRef.ID
			ids[id] = append(ids[id], i)
			if len(ids[id]) == 1 {
				repo.Digests = append(repo.Digests, importDigest{
					Name:  defaultRef.ID,
					Image: cache[id],
				})
			}
		} else {
			var toName string
			if spec.To != nil {
				toName = spec.To.Name
			} else {
				toName = defaultRef.Tag
			}
			tagReference := stream.Spec.Tags[toName]

			preferArch := tagReference.Annotations[imagev1.ImporterPreferArchAnnotation]
			preferOS := tagReference.Annotations[imagev1.ImporterPreferOSAnnotation]

			tag := manifestKey{repositoryKey: key, preferArch: preferArch, preferOS: preferOS}
			tag.value = defaultRef.Tag
			tags[tag] = append(tags[tag], i)
			if len(tags[tag]) == 1 {
				repo.Tags = append(repo.Tags, importTag{
					Name:       defaultRef.Tag,
					PreferArch: preferArch,
					PreferOS:   preferOS,
					Image:      cache[tag],
				})
			}
		}
	}

	// for each repository we found, import all tags and digests
	for key, repo := range repositories {
		imp.importRepositoryFromDocker(ctx, repo)
		for _, tag := range repo.Tags {
			j := manifestKey{repositoryKey: key, preferArch: tag.PreferArch, preferOS: tag.PreferOS}
			j.value = tag.Name
			if tag.Image != nil {
				cache[j] = tag.Image
			}
			for _, index := range tags[j] {
				if tag.Err != nil {
					setImageImportStatus(isi, index, tag.Name, tag.Err)
					continue
				}
				copied := *tag.Image
				image := &isi.Status.Images[index]
				ref := repo.Ref
				ref.Tag, ref.ID = tag.Name, copied.Name
				copied.DockerImageReference = ref.MostSpecific().Exact()
				image.Tag = tag.Name
				image.Image = &copied
				image.Status.Status = metav1.StatusSuccess
			}
		}
		for _, digest := range repo.Digests {
			j := manifestKey{repositoryKey: key}
			j.value = digest.Name
			if digest.Image != nil {
				cache[j] = digest.Image
			}
			for _, index := range ids[j] {
				if digest.Err != nil {
					setImageImportStatus(isi, index, "", digest.Err)
					continue
				}
				image := &isi.Status.Images[index]
				copied := *digest.Image
				ref := repo.Ref
				ref.Tag, ref.ID = "", copied.Name
				copied.DockerImageReference = ref.MostSpecific().Exact()
				image.Image = &copied
				image.Status.Status = metav1.StatusSuccess
			}
		}
	}
}

// importFromRepository imports the repository named on the ImageStreamImport, if any, importing
// up to maximumTagsPerRepo, and reporting status on each image that is attempted to be imported.
// If the repository cannot be found or tags cannot be retrieved, the repository status field is
// set.
func (imp *ImageStreamImporter) importFromRepository(ctx context.Context, isi *imageapi.ImageStreamImport) {
	if isi.Spec.Repository == nil {
		return
	}
	cache := imp.digestToRepositoryCache[ctx]
	isi.Status.Repository = &imageapi.RepositoryImportStatus{}
	status := isi.Status.Repository

	spec := isi.Spec.Repository
	from := spec.From
	if from.Kind != "DockerImage" {
		return
	}
	// TODO: This should be removed in 1.6
	// See for more info: https://github.com/openshift/origin/pull/11774#issuecomment-258905994
	var (
		err error
		ref imageapi.DockerImageReference
	)
	if from.Name != "*" {
		ref, err = imageref.Parse(from.Name)
		if err != nil {
			status.Status = invalidStatus("", field.Invalid(field.NewPath("from", "name"), from.Name, fmt.Sprintf("invalid name: %v", err)))
			return
		}
	} else {
		ref = imageapi.DockerImageReference{Name: from.Name}
	}

	if imp.blockedRegistry(ref) {
		status.Status = forbiddenStatus(fmt.Errorf("registry %s blocked", ref.Registry))
		return
	}

	defaultRef := ref.DockerClientDefaults()
	repoName := defaultRef.RepositoryName()
	registryURL := defaultRef.RegistryURL()

	key := repositoryKey{url: *registryURL, name: repoName}
	repo := &importRepository{
		Ref:         ref,
		Registry:    &key.url,
		Name:        key.name,
		Insecure:    imp.allowRegistryInsecureAccess(spec.ImportPolicy, ref),
		MaximumTags: imp.maximumTagsPerRepo,
	}
	imp.importRepositoryFromDocker(ctx, repo)

	if repo.Err != nil {
		status.Status = imageImportStatus(repo.Err, "", "repository")
		return
	}

	additional := []string{}
	tagKey := manifestKey{repositoryKey: key}
	for _, s := range repo.AdditionalTags {
		tagKey.value = s
		if image, ok := cache[tagKey]; ok {
			repo.Tags = append(repo.Tags, importTag{
				Name:  s,
				Image: image,
			})
		} else {
			additional = append(additional, s)
		}
	}
	status.AdditionalTags = additional

	failures := 0
	status.Status.Status = metav1.StatusSuccess
	status.Images = make([]imageapi.ImageImportStatus, len(repo.Tags))
	for i, tag := range repo.Tags {
		status.Images[i].Tag = tag.Name
		if tag.Err != nil {
			failures++
			status.Images[i].Status = imageImportStatus(tag.Err, "", "repository")
			continue
		}
		status.Images[i].Status.Status = metav1.StatusSuccess

		copied := *tag.Image
		ref.Tag, ref.ID = tag.Name, copied.Name
		copied.DockerImageReference = ref.MostSpecific().Exact()
		status.Images[i].Image = &copied
	}
	if failures > 0 {
		status.Status.Status = metav1.StatusFailure
		status.Status.Reason = metav1.StatusReason("ImportFailed")
		switch failures {
		case 1:
			status.Status.Message = "one of the images from this repository failed to import"
		default:
			status.Status.Message = fmt.Sprintf("%d of the images from this repository failed to import", failures)
		}
	}
}

func applyErrorToRepository(repository *importRepository, err error) {
	repository.Err = err
	for i := range repository.Tags {
		repository.Tags[i].Err = err
	}

	// repository.Digests are handled independedly because they might be
	// accessed from a mirror.
}

func formatPingError(imageRef imageref.DockerImageReference, insecure bool, err error) error {
	switch {
	case err == reference.ErrReferenceInvalidFormat:
		err = field.Invalid(field.NewPath("from", "name"), imageRef.RepositoryName(), "the provided repository name is not valid")
	case isDockerError(err, v2.ErrorCodeNameUnknown):
		err = kapierrors.NewNotFound(image.Resource("dockerimage"), imageRef.Exact())
	case isDockerError(err, errcode.ErrorCodeUnauthorized):
		err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the container image %q", imageRef.Exact()))
	case strings.Contains(err.Error(), "tls: oversized record received with length") && !insecure:
		err = kapierrors.NewBadRequest(fmt.Sprintf("the repository %s is HTTP only and requires the insecure flag to import", imageRef.RepositoryName()))
	case strings.HasSuffix(err.Error(), "no basic auth credentials"):
		err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the container image %q and did not have credentials to the repository", imageRef.Exact()))
	default:
		err = fmt.Errorf("%s: %s", imageRef.Exact(), err)
	}
	return err
}

func formatRepositoryError(ref reference.Named, err error) error {
	switch {
	case isDockerError(err, v2.ErrorCodeManifestUnknown):
		err = kapierrors.NewNotFound(image.Resource("dockerimage"), ref.String())
	case isDockerError(err, errcode.ErrorCodeUnauthorized):
		err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the container image %q", ref.String()))
	case strings.HasSuffix(err.Error(), "no basic auth credentials"):
		err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the container image %q", ref.String()))
	case strings.HasSuffix(err.Error(), "incorrect username or password"):
		err = kapierrors.NewUnauthorized(fmt.Sprintf("incorrect username or password for image %q", ref.String()))
	default:
		err = fmt.Errorf("%s: %s", ref.String(), err)
	}
	return err
}

// calculateImageSize gets and updates size of each image layer. If manifest v2 is converted to v1,
// then it loses information about layers size. We have to get this information from server again.
func (imp *ImageStreamImporter) calculateImageSize(ctx context.Context, bs distribution.BlobStore, image *imageapi.Image) error {
	blobSet := sets.NewString()
	size := int64(0)
	for i := range image.DockerImageLayers {
		layer := &image.DockerImageLayers[i]

		if blobSet.Has(layer.Name) {
			continue
		}
		blobSet.Insert(layer.Name)

		if layerSize, ok := imp.digestToLayerSizeCache.Get(layer.Name); ok {
			layerSize := layerSize.(int64)
			layer.LayerSize = layerSize
			size += layerSize
			continue
		}

		desc, err := bs.Stat(ctx, godigest.Digest(layer.Name))
		if err != nil {
			return err
		}

		imp.digestToLayerSizeCache.Add(layer.Name, desc.Size)
		layer.LayerSize = desc.Size
		size += desc.Size
	}

	if len(image.DockerImageConfig) > 0 && !blobSet.Has(image.DockerImageMetadata.ID) {
		blobSet.Insert(image.DockerImageMetadata.ID)
		size += int64(len(image.DockerImageConfig))
	}

	image.DockerImageMetadata.Size = size
	return nil
}

func isSubrepo(repo, ancestor string) bool {
	if repo == ancestor {
		return true
	}
	if len(repo) > len(ancestor) {
		return strings.HasPrefix(repo, ancestor) && repo[len(ancestor)] == '/'
	}
	return false
}

func (imp *ImageStreamImporter) findRegistryConfiguration(ref reference.Named) *sysregistriesv2.Registry {
	if imp.regConf == nil || len(imp.regConf.Registries) == 0 {
		return nil
	}

	repoName := ref.Name()

	var bestMatch *sysregistriesv2.Registry
	for i := range imp.regConf.Registries {
		// XXX We cannot use `i, reg := range` here and `&reg` later.
		// Otherwise the value will be changed on the next loop iteration.
		reg := imp.regConf.Registries[i]

		if bestMatch == nil || len(reg.Prefix) > len(bestMatch.Prefix) {
			if isSubrepo(repoName, reg.Prefix) {
				bestMatch = &reg
			}
		}
	}
	return bestMatch
}

// getPullSources returns a list of possible sources for ref. In case of errors, the returned list
// still can be used and contains at least one element.
func (imp *ImageStreamImporter) getPullSources(ref reference.Named) ([]sysregistriesv2.PullSource, error) {
	reg := imp.findRegistryConfiguration(ref)
	if reg == nil {
		return []sysregistriesv2.PullSource{{Reference: ref}}, nil
	}

	pullSources, err := reg.PullSourcesFromReference(ref)
	if err != nil {
		return []sysregistriesv2.PullSource{{Reference: ref}}, fmt.Errorf("unable to get all pull sources for %s: %v", ref.String(), err)
	}

	return pullSources, nil
}

// getManifestFromSource pulls a manifest from the source that is referenced by ref. Also it
// returns the manifest service and the blob store so that additional resources (an image
// configuration for schema 2 manifests or another manifest for manifest lists) can be pulled.
func (imp *ImageStreamImporter) getManifestFromSource(
	ctx context.Context, ref reference.Named, insecure bool,
) (distribution.Manifest, distribution.ManifestService, distribution.BlobStore, error) {
	imageRef, err := imageref.Parse(ref.String())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to parse reference %q: %v", ref.String(), err)
	}

	repo, err := imp.retriever.Repository(ctx, imageRef, insecure)
	if err != nil {
		return nil, nil, nil, formatPingError(imageRef, insecure, err)
	}

	ms, err := repo.Manifests(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create manifests service for %s: %v", ref.String(), err)
	}

	var dgst godigest.Digest
	var opts []distribution.ManifestServiceOption
	if digestRef, ok := ref.(reference.Digested); ok {
		dgst = digestRef.Digest()
	} else if taggedRef, ok := ref.(reference.Tagged); ok {
		opts = append(opts, distribution.WithTag(taggedRef.Tag()))
	} else {
		opts = append(opts, distribution.WithTag("latest"))
	}

	manifest, err := ms.Get(ctx, dgst, opts...)
	if err != nil {
		return nil, nil, nil, formatRepositoryError(ref, err)
	}

	return manifest, ms, repo.Blobs(ctx), nil
}

// getManifestFromSource pulls a manifest from the source respecting
// V2RegistriesConf.
func (imp *ImageStreamImporter) getManifest(
	ctx context.Context, ref reference.Named, insecure bool,
) (distribution.Manifest, distribution.ManifestService, distribution.BlobStore, error) {
	var errs []error

	pullSources, err := imp.getPullSources(ref)
	if err != nil {
		errs = append(errs, err)
	}

	if klog.V(5).Enabled() {
		out := make([]string, len(pullSources))
		for i, pullSource := range pullSources {
			out[i] = pullSource.Reference.String()
		}
		klog.Infof("importing %s: going to try next pull sources in order: %v", ref, out)
	}

	for _, pullSource := range pullSources {
		klog.V(5).Infof("importing %s: trying to fetch manifest from %s...", ref, pullSource.Reference)

		manifest, ms, bs, err := imp.getManifestFromSource(ctx, pullSource.Reference, insecure)
		if err != nil {
			klog.V(5).Infof("importing %s: failed to get manifest from %s: %s", ref, pullSource.Reference, err)
			errs = append(errs, err)
			continue
		}

		klog.V(5).Infof("importing %s: got manifest from %s", ref.String(), pullSource.Reference)

		return manifest, ms, bs, nil
	}

	if len(errs) == 1 {
		// getManifestFromSource can return StatusError (for example,
		// Unauthorized) that provides a meaningful Reason. NewAggregate will
		// hide this Reason, so it's better to not aggregate a single error.
		// Eventually we need a better way to aggregate errors.
		return nil, nil, nil, errs[0]
	}
	return nil, nil, nil, utilerrors.NewAggregate(errs)
}

func manifestFromManifestList(
	ctx context.Context,
	manifestList *manifestlist.DeserializedManifestList,
	ref reference.Named,
	s distribution.ManifestService,
	preferArch, preferOS string,
) (distribution.Manifest, godigest.Digest, error) {
	if len(manifestList.Manifests) == 0 {
		return nil, "", fmt.Errorf("no manifests in manifest list")
	}

	if preferArch == "" {
		preferArch = runtime.GOARCH
	}
	if preferOS == "" {
		preferOS = runtime.GOOS
	}

	var manifestDigest godigest.Digest
	for _, manifestDescriptor := range manifestList.Manifests {
		if manifestDescriptor.Platform.Architecture == preferArch && manifestDescriptor.Platform.OS == preferOS {
			manifestDigest = manifestDescriptor.Digest
			break
		}
	}

	// if we couldn't match the preferred arch/os, and we couldn't match the platform's
	// arch/os, prefer x86/linux before falling back to "first image in the manifestlist"
	// as a last resort.
	if manifestDigest == "" {
		for _, manifestDescriptor := range manifestList.Manifests {
			if manifestDescriptor.Platform.Architecture == "amd64" && manifestDescriptor.Platform.OS == "linux" {
				manifestDigest = manifestDescriptor.Digest
				break
			}
		}
	}

	if manifestDigest == "" {
		klog.V(5).Infof("unable to find %s/%s manifest in manifest list %s, doing conservative fail by switching to the first one: %#+v", preferOS, preferArch, ref.String(), manifestList.Manifests[0])
		manifestDigest = manifestList.Manifests[0].Digest
	}

	manifest, err := s.Get(ctx, manifestDigest)
	if err != nil {
		klog.V(5).Infof("unable to get %s/%s manifest by digest %q for image %s: %#v", preferOS, preferArch, manifestDigest, ref.String(), err)
		return nil, "", err
	}

	return manifest, manifestDigest, err
}

func (imp *ImageStreamImporter) importManifest(
	ctx context.Context,
	manifest distribution.Manifest,
	ref reference.Named,
	d godigest.Digest,
	s distribution.ManifestService,
	b distribution.BlobStore,
	preferArch, preferOS string,
) (image *imageapi.Image, err error) {
	if manifestList, ok := manifest.(*manifestlist.DeserializedManifestList); ok {
		manifest, d, err = manifestFromManifestList(ctx, manifestList, ref, s, preferArch, preferOS)
		if err != nil {
			return nil, formatRepositoryError(ref, err)
		}
	}

	if signedManifest, isSchema1 := manifest.(*schema1.SignedManifest); isSchema1 {
		image, err = schema1ToImage(signedManifest, d)
	} else if deserializedManifest, isSchema2 := manifest.(*schema2.DeserializedManifest); isSchema2 {
		imageConfig, getImportConfigErr := b.Get(ctx, deserializedManifest.Config.Digest)
		if getImportConfigErr != nil {
			klog.V(5).Infof("unable to get image config by digest %q for image %s: %#v", d, ref.String(), getImportConfigErr)
			return image, formatRepositoryError(ref, getImportConfigErr)
		}
		image, err = schema2OrOCIToImage(deserializedManifest, imageConfig, d)
	} else if deserializedManifest, isOCISchema := manifest.(*ocischema.DeserializedManifest); isOCISchema {
		imageConfig, getImportConfigErr := b.Get(ctx, deserializedManifest.Config.Digest)
		if getImportConfigErr != nil {
			klog.V(5).Infof("unable to get image config by digest %q for image %s: %#v", d, ref.String(), getImportConfigErr)
			return image, formatRepositoryError(ref, getImportConfigErr)
		}
		image, err = schema2OrOCIToImage(manifest, imageConfig, d)
	} else {
		err = fmt.Errorf("unsupported image manifest type: %T", manifest)
		klog.V(5).Info(err)
	}
	if err != nil {
		return
	}

	if err := internalimageutil.InternalImageWithMetadata(image); err != nil {
		return image, err
	}

	if image.DockerImageMetadata.Size == 0 {
		if err := imp.calculateImageSize(ctx, b, image); err != nil {
			return image, err
		}
	}
	return
}

// importRepositoryFromDocker loads the tags and images requested in the passed importRepository, obeying the
// optional rate limiter.  Errors are set onto the individual tags and digest objects.
func (imp *ImageStreamImporter) importRepositoryFromDocker(ctx context.Context, repository *importRepository) {
	klog.V(5).Infof("importing remote Docker repository registry=%s repository=%s insecure=%t", repository.Registry, repository.Name, repository.Insecure)

	// load digests
	for i := range repository.Digests {
		importDigest := &repository.Digests[i]
		if importDigest.Err != nil || importDigest.Image != nil {
			continue
		}

		d, err := godigest.Parse(importDigest.Name)
		if err != nil {
			importDigest.Err = err
			continue
		}

		ref := repository.Ref
		ref.Tag = ""
		ref.ID = string(d)

		dockerRef, err := reference.ParseNormalizedNamed(ref.Exact())
		if err != nil {
			importDigest.Err = fmt.Errorf("unable to parse docker reference %s: %v", ref.Exact(), err)
			continue
		}

		imp.limiter.Accept()

		manifest, ms, bs, err := imp.getManifest(ctx, dockerRef, repository.Insecure)
		if err != nil {
			klog.V(5).Infof("unable to get manifest by digest %s for image %s: %v", d, ref.Exact(), err)
			importDigest.Err = err
			continue
		}

		importDigest.Image, importDigest.Err = imp.importManifest(ctx, manifest, dockerRef, d, ms, bs, "", "")
	}

	// if repository import is requested (MaximumTags), attempt to load the tags, sort them, and request the first N
	if count := repository.MaximumTags; count > 0 || count == -1 {
		// retrieve the repository
		repo, err := imp.retriever.Repository(ctx, repository.Ref, repository.Insecure)
		if err != nil {
			klog.V(5).Infof("unable to access repository %#v: %#v", repository, err)
			if strings.HasSuffix(err.Error(), "does not support v2 API") {
				err := kapierrors.NewForbidden(image.Resource(""), "", fmt.Errorf("registry %q does not support the v2 Registry API", repository.Registry.Host))
				err.ErrStatus.Reason = "NotV2Registry"
				applyErrorToRepository(repository, err)
				return
			}
			err = formatPingError(repository.Ref, repository.Insecure, err)
			applyErrorToRepository(repository, err)
			return
		}

		tags, err := repo.Tags(ctx).All(ctx)
		if err != nil {
			klog.V(5).Infof("unable to access tags for repository %#v: %#v", repository, err)
			switch {
			case isDockerError(err, v2.ErrorCodeNameUnknown):
				err = kapierrors.NewNotFound(image.Resource("dockerimage"), repository.Ref.Exact())
			case isDockerError(err, errcode.ErrorCodeUnauthorized):
				err = kapierrors.NewUnauthorized(fmt.Sprintf("you may not have access to the container image %q", repository.Ref.Exact()))
			}
			repository.Err = err
			return
		}
		// some images on the Hub have empty tags - treat those as "latest"
		set := sets.NewString(tags...)
		if set.Has("") {
			set.Delete("")
			set.Insert(imagev1.DefaultImageTag)
		}
		tags = set.List()
		// include only the top N tags in the result, put the rest in AdditionalTags
		imageutil.PrioritizeTags(tags)
		for _, s := range tags {
			if count <= 0 && repository.MaximumTags != -1 {
				repository.AdditionalTags = append(repository.AdditionalTags, s)
				continue
			}
			count--
			repository.Tags = append(repository.Tags, importTag{
				Name: s,
			})
		}
	}

	for i := range repository.Tags {
		importTag := &repository.Tags[i]
		if importTag.Err != nil || importTag.Image != nil {
			continue
		}

		ref := repository.Ref
		ref.Tag = importTag.Name
		ref.ID = ""

		dockerRef, err := reference.ParseNormalizedNamed(ref.Exact())
		if err != nil {
			importTag.Err = fmt.Errorf("unable to parse docker reference %s: %v", ref.Exact(), err)
			continue
		}

		imp.limiter.Accept()

		manifest, ms, bs, err := imp.getManifest(ctx, dockerRef, repository.Insecure)
		if err != nil {
			klog.V(5).Infof("unable to get manifest by tag %q for image %s: %#v", importTag.Name, ref.Exact(), err)
			importTag.Err = err
			continue
		}

		importTag.Image, importTag.Err = imp.importManifest(ctx, manifest, dockerRef, "", ms, bs, importTag.PreferArch, importTag.PreferOS)
	}
}

type importTag struct {
	Name       string
	PreferArch string
	PreferOS   string
	Image      *imageapi.Image
	Err        error
}

type importDigest struct {
	Name  string
	Image *imageapi.Image
	Err   error
}

type importRepository struct {
	Ref      imageapi.DockerImageReference
	Registry *url.URL
	Name     string
	Insecure bool

	Tags    []importTag
	Digests []importDigest

	MaximumTags    int
	AdditionalTags []string
	Err            error
}

// repositoryKey is the key used to cache information loaded from a remote Docker repository.
type repositoryKey struct {
	// The URL of the server
	url url.URL
	// The name of the image repository (contains both namespace and path)
	name string
}

// manifestKey is a key for a map between a container image tag or image ID and a retrieved imageapi.Image, used
// to ensure we don't fetch the same image multiple times.
type manifestKey struct {
	repositoryKey
	// The tag or ID of the image, not used within the same map
	value string
	// An architecture of the image which should be selected from a manifest list.
	preferArch string
	// An operation system of the image which should be selected from a manifest list.
	preferOS string
}

func imageImportStatus(err error, kind, position string) metav1.Status {
	switch t := err.(type) {
	case kapierrors.APIStatus:
		return t.Status()
	case *field.Error:
		return kapierrors.NewInvalid(image.Kind(kind), position, field.ErrorList{t}).ErrStatus
	default:
		return kapierrors.NewInternalError(err).ErrStatus
	}
}

func setImageImportStatus(images *imageapi.ImageStreamImport, i int, tag string, err error) {
	images.Status.Images[i].Tag = tag
	images.Status.Images[i].Status = imageImportStatus(err, "", "")
}

func invalidStatus(position string, errs ...*field.Error) metav1.Status {
	return kapierrors.NewInvalid(schema.GroupKind{Group: "", Kind: ""}, position, errs).ErrStatus
}

func forbiddenStatus(err error) metav1.Status {
	return kapierrors.NewForbidden(schema.GroupResource{Group: "", Resource: ""}, "", err).ErrStatus
}
