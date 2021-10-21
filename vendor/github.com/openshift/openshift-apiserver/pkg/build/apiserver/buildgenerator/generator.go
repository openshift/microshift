package buildgenerator

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	v1 "github.com/openshift/openshift-apiserver/pkg/build/apis/build/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/credentialprovider"
	credentialprovidersecrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	buildv1clienttyped "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1clienttyped "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"github.com/openshift/library-go/pkg/build/buildutil"
	"github.com/openshift/library-go/pkg/build/naming"
	"github.com/openshift/library-go/pkg/image/imageutil"
	"github.com/openshift/openshift-apiserver/pkg/bootstrappolicy"
	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
)

const conflictRetries = 3

// BuildGenerator is a central place responsible for generating new Build objects
// from BuildConfigs and other Builds.
type BuildGenerator struct {
	Client          GeneratorClient
	ServiceAccounts corev1client.ServiceAccountsGetter
	Secrets         corev1client.SecretsGetter
}

// GeneratorClient is the API client used by the generator
type GeneratorClient interface {
	GetBuildConfig(ctx context.Context, name string, options metav1.GetOptions) (*buildv1.BuildConfig, error)
	UpdateBuildConfig(ctx context.Context, buildConfig *buildv1.BuildConfig, options metav1.UpdateOptions) error
	GetBuild(ctx context.Context, name string, options metav1.GetOptions) (*buildv1.Build, error)
	CreateBuild(ctx context.Context, build *buildv1.Build, options metav1.CreateOptions) error
	UpdateBuild(ctx context.Context, build *buildv1.Build, options metav1.UpdateOptions) error
	GetImageStream(ctx context.Context, name string, options metav1.GetOptions) (*imagev1.ImageStream, error)
	GetImageStreamImage(ctx context.Context, name string, options metav1.GetOptions) (*imagev1.ImageStreamImage, error)
	GetImageStreamTag(ctx context.Context, name string, options metav1.GetOptions) (*imagev1.ImageStreamTag, error)
}

// Client is an implementation of the GeneratorClient interface
type Client struct {
	BuildConfigs      buildv1clienttyped.BuildConfigsGetter
	Builds            buildv1clienttyped.BuildsGetter
	ImageStreams      imagev1clienttyped.ImageStreamsGetter
	ImageStreamImages imagev1clienttyped.ImageStreamImagesGetter
	ImageStreamTags   imagev1clienttyped.ImageStreamTagsGetter
}

// GetBuildConfig retrieves a named build config
func (c Client) GetBuildConfig(ctx context.Context, name string, options metav1.GetOptions) (*buildv1.BuildConfig, error) {
	return c.BuildConfigs.BuildConfigs(apirequest.NamespaceValue(ctx)).Get(ctx, name, options)
}

// UpdateBuildConfig updates a named build config
func (c Client) UpdateBuildConfig(ctx context.Context, buildConfig *buildv1.BuildConfig, options metav1.UpdateOptions) error {
	_, err := c.BuildConfigs.BuildConfigs(apirequest.NamespaceValue(ctx)).Update(ctx, buildConfig, options)
	return err
}

// GetBuild retrieves a build
func (c Client) GetBuild(ctx context.Context, name string, options metav1.GetOptions) (*buildv1.Build, error) {
	return c.Builds.Builds(apirequest.NamespaceValue(ctx)).Get(ctx, name, options)
}

// CreateBuild creates a new build
func (c Client) CreateBuild(ctx context.Context, build *buildv1.Build, options metav1.CreateOptions) error {
	_, err := c.Builds.Builds(apirequest.NamespaceValue(ctx)).Create(ctx, build, options)
	return err
}

// UpdateBuild updates a build
func (c Client) UpdateBuild(ctx context.Context, build *buildv1.Build, options metav1.UpdateOptions) error {
	_, err := c.Builds.Builds(apirequest.NamespaceValue(ctx)).Update(ctx, build, options)
	return err
}

// GetImageStream retrieves a named image stream
func (c Client) GetImageStream(ctx context.Context, name string, options metav1.GetOptions) (*imagev1.ImageStream, error) {
	return c.ImageStreams.ImageStreams(apirequest.NamespaceValue(ctx)).Get(ctx, name, options)
}

// GetImageStreamImage retrieves an image stream image
func (c Client) GetImageStreamImage(ctx context.Context, name string, options metav1.GetOptions) (*imagev1.ImageStreamImage, error) {
	return c.ImageStreamImages.ImageStreamImages(apirequest.NamespaceValue(ctx)).Get(ctx, name, options)
}

// GetImageStreamTag retrieves and image stream tag
func (c Client) GetImageStreamTag(ctx context.Context, name string, options metav1.GetOptions) (*imagev1.ImageStreamTag, error) {
	return c.ImageStreamTags.ImageStreamTags(apirequest.NamespaceValue(ctx)).Get(ctx, name, options)
}

// fetchServiceAccountSecrets retrieves the Secrets used for pushing and pulling
// images from private container image registries.
func fetchServiceAccountSecrets(ctx context.Context, secrets corev1client.SecretsGetter, serviceAccounts corev1client.ServiceAccountsGetter, namespace, serviceAccount string) ([]corev1.Secret, error) {
	var result []corev1.Secret
	sa, err := serviceAccounts.ServiceAccounts(namespace).Get(ctx, serviceAccount, metav1.GetOptions{})
	if err != nil {
		return result, fmt.Errorf("error getting push/pull secrets for service account %s/%s: %v", namespace, serviceAccount, err)
	}
	for _, ref := range sa.Secrets {
		secret, err := secrets.Secrets(namespace).Get(ctx, ref.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}
		result = append(result, *secret)
	}
	return result, nil
}

func getTriggerNamespaceName(bc *buildv1.BuildConfig, triggerFrom buildv1.ImageStreamTagReference) k8stypes.NamespacedName {
	if len(triggerFrom.Name) == 0 {
		strategyRef := buildutil.GetInputReference(bc.Spec.Strategy)
		if strategyRef == nil || strategyRef.Kind != "ImageStreamTag" {
			return k8stypes.NamespacedName{}
		}
		triggerFrom.Namespace = strategyRef.Namespace
		triggerFrom.Name = strategyRef.Name
	}
	triggerNs := triggerFrom.Namespace
	if triggerNs == "" {
		triggerNs = bc.Namespace
	}
	return k8stypes.NamespacedName{Namespace: triggerNs, Name: triggerFrom.Name}
}

// findImageChangeTrigger finds an image change trigger that has a from that matches the passed in ref
// if no match is found but there is an image change trigger with a null from, that trigger is returned
func findImageChangeTrigger(bc *buildv1.BuildConfig, ref *corev1.ObjectReference) (*buildv1.ImageChangeTrigger, *buildv1.ImageChangeTriggerStatus) {
	if ref == nil {
		return nil, nil
	}
	var requestTrigger *buildv1.ImageChangeTrigger
	var responseTrigger *buildv1.ImageChangeTriggerStatus
	refNs := ref.Namespace
	if refNs == "" {
		refNs = bc.Namespace
	}
	for _, trigger := range bc.Spec.Triggers {
		if trigger.Type != buildv1.ImageChangeBuildTriggerType {
			continue
		}
		imageChange := trigger.ImageChange
		istRef := buildv1.ImageStreamTagReference{}
		from := trigger.ImageChange.From
		if from != nil {
			istRef = buildv1.ImageStreamTagReference{Namespace: from.Namespace, Name: from.Name}
		}
		triggerNSN := getTriggerNamespaceName(bc, istRef)
		if triggerNSN.Name == ref.Name && triggerNSN.Namespace == refNs {
			requestTrigger = imageChange
			break
		}
	}
	for _, trigger := range bc.Status.ImageChangeTriggers {
		triggerNSN := getTriggerNamespaceName(bc, trigger.From)
		if triggerNSN.Name == ref.Name && triggerNSN.Namespace == refNs {
			responseTrigger = &trigger
			break
		}

	}
	return requestTrigger, responseTrigger
}

func describeBuildRequest(request *buildv1.BuildRequest) string {
	desc := fmt.Sprintf("BuildConfig: %s/%s", request.Namespace, request.Name)
	if request.Revision != nil {
		desc += fmt.Sprintf(", Revision: %#v", request.Revision.Git)
	}
	if request.TriggeredByImage != nil {
		desc += fmt.Sprintf(", TriggeredBy: %s/%s with stream: %s/%s",
			request.TriggeredByImage.Kind, request.TriggeredByImage.Name,
			request.From.Kind, request.From.Name)
	}
	if request.LastVersion != nil {
		desc += fmt.Sprintf(", LastVersion: %d", *request.LastVersion)
	}
	return desc
}

// Adds new Build Args to existing Build Args. Overwrites existing ones
func updateBuildArgs(oldArgs *[]corev1.EnvVar, newArgs []corev1.EnvVar) []corev1.EnvVar {
	combined := make(map[string]string)

	// Change oldArgs into a map
	for _, o := range *oldArgs {
		combined[o.Name] = o.Value
	}

	// Add new args, this overwrites old
	for _, n := range newArgs {
		combined[n.Name] = n.Value
	}

	// Change back into an array
	var result []corev1.EnvVar
	for k, v := range combined {
		result = append(result, corev1.EnvVar{Name: k, Value: v})
	}

	return result
}

// DEPRECATED: Use only by apiserver
func (g *BuildGenerator) InstantiateInternal(ctx context.Context, request *buildapi.BuildRequest, opts metav1.CreateOptions) (*buildapi.Build, error) {
	versionedRequest := &buildv1.BuildRequest{}
	if err := v1.Convert_build_BuildRequest_To_v1_BuildRequest(request, versionedRequest, nil); err != nil {
		return nil, fmt.Errorf("failed to convert internal BuildRequest to external: %v", err)
	}
	build, err := g.Instantiate(ctx, versionedRequest, opts)
	if err != nil {
		return nil, err
	}
	internalBuild := &buildapi.Build{}
	if err := v1.Convert_v1_Build_To_build_Build(build, internalBuild, nil); err != nil {
		return nil, fmt.Errorf("failed to convert external Build to internal: %v", err)
	}
	return internalBuild, nil
}

// Instantiate returns a new Build object based on a BuildRequest object
func (g *BuildGenerator) Instantiate(ctx context.Context, request *buildv1.BuildRequest, opts metav1.CreateOptions) (*buildv1.Build, error) {
	var build *buildv1.Build
	var err error
	for i := 0; i < conflictRetries; i++ {
		build, err = g.instantiate(ctx, request, opts)
		if errors.IsConflict(err) {
			klog.V(2).Infof("instantiate returned conflict, try %d/%d", i+1, conflictRetries)
			continue
		}
		if err != nil {
			return nil, err
		}
		if err == nil {
			break
		}
	}
	return build, err
}

func (g *BuildGenerator) instantiate(ctx context.Context, request *buildv1.BuildRequest, opts metav1.CreateOptions) (*buildv1.Build, error) {
	klog.V(4).Infof("Generating Build from %s", describeBuildRequest(request))
	bc, err := g.Client.GetBuildConfig(ctx, request.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if isPaused(bc) {
		return nil, errors.NewBadRequest(fmt.Sprintf("can't instantiate from BuildConfig %s/%s: BuildConfig is paused", bc.Namespace, bc.Name))
	}

	if err := g.checkLastVersion(bc, request.LastVersion); err != nil {
		return nil, errors.NewBadRequest(err.Error())
	}

	if err := g.updateImageTriggers(ctx, bc, request.From, request.TriggeredByImage); err != nil {
		if _, ok := err.(errors.APIStatus); ok {
			return nil, err
		}
		return nil, errors.NewInternalError(err)
	}

	newBuild, err := g.generateBuildFromConfig(ctx, bc, request.Revision, request.Binary)
	if err != nil {
		if _, ok := err.(errors.APIStatus); ok {
			return nil, err
		}
		return nil, errors.NewInternalError(err)
	}

	// Add labels and annotations from the buildrequest.  Existing
	// label/annotations will take precedence because we don't want system
	// annotations/labels (eg buildname) to get stomped on.
	newBuild.Annotations = mergeMaps(request.Annotations, newBuild.Annotations)
	newBuild.Labels = mergeMaps(request.Labels, newBuild.Labels)

	// Copy build trigger information and build arguments to the build object.
	newBuild.Spec.TriggeredBy = request.TriggeredBy

	if len(request.Env) > 0 {
		updateBuildEnv(newBuild, request.Env)
	}

	// Update the Docker strategy options
	if request.DockerStrategyOptions != nil {
		dockerOpts := request.DockerStrategyOptions

		// Update the Docker build args
		if dockerOpts.BuildArgs != nil && len(dockerOpts.BuildArgs) > 0 {
			if newBuild.Spec.Strategy.DockerStrategy == nil {
				return nil, errors.NewBadRequest(fmt.Sprintf("Cannot specify Docker build specific options on %s/%s, not a Docker build.", bc.Namespace, bc.ObjectMeta.Name))
			}
			newBuild.Spec.Strategy.DockerStrategy.BuildArgs = updateBuildArgs(&newBuild.Spec.Strategy.DockerStrategy.BuildArgs, dockerOpts.BuildArgs)
		}

		// Update the Docker noCache option
		if dockerOpts.NoCache != nil {
			if newBuild.Spec.Strategy.DockerStrategy == nil {
				return nil, errors.NewBadRequest(fmt.Sprintf("Cannot specify Docker build specific options on %s/%s, not a Docker build.", bc.Namespace, bc.ObjectMeta.Name))
			}
			newBuild.Spec.Strategy.DockerStrategy.NoCache = *dockerOpts.NoCache
		}
	}

	// Update the Source strategy options
	if request.SourceStrategyOptions != nil {
		sourceOpts := request.SourceStrategyOptions

		// Update the Source incremental option
		if sourceOpts.Incremental != nil {
			if newBuild.Spec.Strategy.SourceStrategy == nil {
				return nil, errors.NewBadRequest(fmt.Sprintf("Cannot specify Source build specific options on %s/%s, not a Source build.", bc.Namespace, bc.ObjectMeta.Name))
			}
			newBuild.Spec.Strategy.SourceStrategy.Incremental = sourceOpts.Incremental
		}
	}
	klog.V(4).Infof("Build %s/%s has been generated from %s/%s BuildConfig", newBuild.Namespace, newBuild.ObjectMeta.Name, bc.Namespace, bc.ObjectMeta.Name)

	// need to update the BuildConfig because LastVersion and possibly
	// LastTriggeredImageID changed
	if err := g.Client.UpdateBuildConfig(ctx, bc, metav1.UpdateOptions{}); err != nil {
		klog.V(2).Infof("Failed to update BuildConfig %s/%s so no Build will be created", bc.Namespace, bc.Name)
		return nil, err
	}

	// Ideally we would create the build *before* updating the BC to ensure
	// that we don't set the LastTriggeredImageID on the BC and then fail to
	// create the corresponding build, however doing things in that order
	// allows for a race condition in which two builds get kicked off.  Doing
	// it in this order ensures that we catch the race while updating the BC.
	return g.createBuild(ctx, newBuild, opts)
}

// checkLastVersion will return an error if the BuildConfig's LastVersion doesn't match the passed in lastVersion
// when lastVersion is not nil
func (g *BuildGenerator) checkLastVersion(bc *buildv1.BuildConfig, lastVersion *int64) error {
	if lastVersion != nil && bc.Status.LastVersion != *lastVersion {
		klog.V(2).Infof("Aborting version triggered build for BuildConfig %s/%s because the BuildConfig LastVersion (%d) does not match the requested LastVersion (%d)", bc.Namespace, bc.Name, bc.Status.LastVersion, *lastVersion)
		return fmt.Errorf("the LastVersion(%v) on build config %s/%s does not match the build request LastVersion(%d)",
			bc.Status.LastVersion, bc.Namespace, bc.Name, *lastVersion)
	}
	return nil
}

// updateImageTriggers sets the LastTriggeredImageID on all the ImageChangeTriggers on the BuildConfig and
// updates the From reference of the strategy if the strategy uses an ImageStream or ImageStreamTag reference.
// Also updates the LastTriggerTime for the trigger that correlates to the triggeredBy parameter.
func (g *BuildGenerator) updateImageTriggers(ctx context.Context, bc *buildv1.BuildConfig, from, triggeredBy *corev1.ObjectReference) error {
	var requestTrigger *buildv1.ImageChangeTrigger
	var responseTrigger *buildv1.ImageChangeTriggerStatus
	strategyImageRef := buildutil.GetInputReference(bc.Spec.Strategy)
	if from != nil {
		requestTrigger, responseTrigger = findImageChangeTrigger(bc, from)
	}
	if triggeredBy != nil &&
		//TODO we still update spec until deprecated field (drepcated in 4.8) is removed (presumably 4.9)
		((requestTrigger != nil && requestTrigger.LastTriggeredImageID == triggeredBy.Name) ||
			(responseTrigger != nil && responseTrigger.LastTriggeredImageID == triggeredBy.Name)) {
		klog.V(2).Infof("Aborting imageid triggered build for BuildConfig %s/%s with imageid %s because the BuildConfig already matches this imageid", bc.Namespace, bc.Name, triggeredBy.Name)
		return fmt.Errorf("build config %s/%s has already instantiated a build for imageid %s", bc.Namespace, bc.Name, triggeredBy.Name)
	}
	// Update last triggered image id for all image change triggers
	// reset the status field in case the list of ICTs change in the spec; we'll then repopulate all of them
	bc.Status.ImageChangeTriggers = []buildv1.ImageChangeTriggerStatus{}
	for _, trigger := range bc.Spec.Triggers {
		if trigger.Type != buildv1.ImageChangeBuildTriggerType {
			continue
		}
		ictsFrom := trigger.ImageChange.From
		if ictsFrom == nil {
			ictsFrom = strategyImageRef
		}
		imageChangeTriggerStatus := buildv1.ImageChangeTriggerStatus{
			From: buildv1.ImageStreamTagReference{Namespace: ictsFrom.Namespace, Name: ictsFrom.Name},
		}
		// Use the requested image id for the trigger that caused the build, otherwise resolve to the latest
		if triggeredBy != nil && trigger.ImageChange == requestTrigger {
			//TODO we still update spec until deprecated field (deprecated in 4.8) is ignored (presumably 4.9)
			trigger.ImageChange.LastTriggeredImageID = triggeredBy.Name
			imageChangeTriggerStatus.LastTriggeredImageID = triggeredBy.Name
			// we only update the trigger time for trigger associated with triggeredBy
			imageChangeTriggerStatus.LastTriggerTime = metav1.Now()
			bc.Status.ImageChangeTriggers = append(bc.Status.ImageChangeTriggers, imageChangeTriggerStatus)
			continue
		}

		// the reason we update LastTriggeredImageID for triggers whose From field do not match triggeredBy is to
		// minimize triggering multiple builds when several image changes arrive concurrently.  When those events
		// come in, the check prior to the for loop should prevent additional build generation.

		triggerImageRef := trigger.ImageChange.From
		if triggerImageRef == nil {
			triggerImageRef = strategyImageRef
		}
		if triggerImageRef == nil {
			klog.Warningf("Could not get ImageStream reference for default ImageChangeTrigger on BuildConfig %s/%s", bc.Namespace, bc.Name)
			continue
		}
		image, err := g.resolveImageStreamReference(ctx, *triggerImageRef, bc.Namespace)
		if err != nil {
			// If the trigger is for the strategy from ref, return an error
			if trigger.ImageChange.From == nil {
				return err
			}
			// Otherwise, warn that an error occurred, but continue
			klog.Warningf("Could not resolve trigger reference for build config %s/%s: %#v", bc.Namespace, bc.Name, triggerImageRef)
		}
		//TODO we still update spec until deprecated field (deprecated in 4.8) is ignored (presumably 4.9)
		trigger.ImageChange.LastTriggeredImageID = image
		imageChangeTriggerStatus.LastTriggeredImageID = image
		// reminder: we do not update the LastTriggeredTime here
		bc.Status.ImageChangeTriggers = append(bc.Status.ImageChangeTriggers, imageChangeTriggerStatus)
	}
	return nil
}

// Clone returns clone of a Build
// DEPRECATED: Use only in apiserver
func (g *BuildGenerator) CloneInternal(ctx context.Context, request *buildapi.BuildRequest) (*buildapi.Build, error) {
	versionedRequest := &buildv1.BuildRequest{}
	if err := v1.Convert_build_BuildRequest_To_v1_BuildRequest(request, versionedRequest, nil); err != nil {
		return nil, err
	}
	build, err := g.Clone(ctx, versionedRequest)
	if err != nil {
		return nil, err
	}
	internalBuild := &buildapi.Build{}
	if err := v1.Convert_v1_Build_To_build_Build(build, internalBuild, nil); err != nil {
		return nil, err
	}
	return internalBuild, nil
}

// Clone returns clone of a Build
func (g *BuildGenerator) Clone(ctx context.Context, request *buildv1.BuildRequest) (*buildv1.Build, error) {
	var build *buildv1.Build
	var err error

	for i := 0; i < conflictRetries; i++ {
		build, err = g.clone(ctx, request)
		if err == nil || !errors.IsConflict(err) {
			break
		}
		klog.V(4).Infof("clone returned conflict, try %d/%d", i+1, conflictRetries)
	}

	return build, err
}

func (g *BuildGenerator) clone(ctx context.Context, request *buildv1.BuildRequest) (*buildv1.Build, error) {
	klog.V(4).Infof("Generating build from build %s/%s", request.Namespace, request.Name)
	build, err := g.Client.GetBuild(ctx, request.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var buildConfig *buildv1.BuildConfig
	if build.Status.Config != nil {
		buildConfig, err = g.Client.GetBuildConfig(ctx, build.Status.Config.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if isPaused(buildConfig) {
			return nil, errors.NewInternalError(&generatorFatalError{Reason: fmt.Sprintf(
				"can't instantiate from BuildConfig %s/%s: BuildConfig is paused", buildConfig.Namespace, buildConfig.Name)})
		}
	}

	newBuild := generateBuildFromBuild(build, buildConfig)
	klog.V(4).Infof("Build %s/%s has been generated from Build %s/%s", newBuild.Namespace, newBuild.ObjectMeta.Name, build.Namespace, build.ObjectMeta.Name)

	// Copy build trigger information to the build object.
	newBuild.Spec.TriggeredBy = request.TriggeredBy

	if len(request.Env) > 0 {
		updateBuildEnv(newBuild, request.Env)
	}

	// Update the Docker build args
	if request.DockerStrategyOptions != nil {
		dockerOpts := request.DockerStrategyOptions
		if dockerOpts.BuildArgs != nil && len(dockerOpts.BuildArgs) > 0 {
			if newBuild.Spec.Strategy.DockerStrategy == nil {
				return nil, errors.NewBadRequest(fmt.Sprintf("Cannot specify build args on %s/%s, not a Docker build.", buildConfig.Namespace, buildConfig.ObjectMeta.Name))
			}
			newBuild.Spec.Strategy.DockerStrategy.BuildArgs = updateBuildArgs(&newBuild.Spec.Strategy.DockerStrategy.BuildArgs, dockerOpts.BuildArgs)
		}
	}

	// need to update the BuildConfig because LastVersion changed
	if buildConfig != nil {
		if err := g.Client.UpdateBuildConfig(ctx, buildConfig, metav1.UpdateOptions{}); err != nil {
			klog.V(4).Infof("Failed to update BuildConfig %s/%s so no Build will be created", buildConfig.Namespace, buildConfig.Name)
			return nil, err
		}
	}

	return g.createBuild(ctx, newBuild, metav1.CreateOptions{})
}

// GeneratorFatalError represents a fatal error while generating a build.
// An operation that fails because of a fatal error should not be retried.
type generatorFatalError struct {
	// Reason the fatal error occurred
	Reason string
}

// Error returns the error string for this fatal error
func (e *generatorFatalError) Error() string {
	return fmt.Sprintf("fatal error generating Build from BuildConfig: %s", e.Reason)
}

// createBuild is responsible for validating build object and saving it and returning newly created object
func (g *BuildGenerator) createBuild(ctx context.Context, build *buildv1.Build, opts metav1.CreateOptions) (*buildv1.Build, error) {
	if !rest.ValidNamespace(ctx, &build.ObjectMeta) {
		return nil, errors.NewConflict(buildv1.Resource("build"), build.Namespace, fmt.Errorf("Build.Namespace does not match the provided context"))
	}
	rest.FillObjectMetaSystemFields(&build.ObjectMeta)
	err := g.Client.CreateBuild(ctx, build, opts)
	if err != nil {
		return nil, err
	}
	return g.Client.GetBuild(ctx, build.Name, metav1.GetOptions{})
}

// generateBuildFromConfig generates a build definition based on the current imageid
// from any ImageStream that is associated to the BuildConfig by From reference in
// the Strategy, or uses the Image field of the Strategy. If binary is provided, override
// the current build strategy with a binary artifact for this specific build.
// Takes a BuildConfig to base the build on, and an optional SourceRevision to build.
func (g *BuildGenerator) generateBuildFromConfig(ctx context.Context, bc *buildv1.BuildConfig, revision *buildv1.SourceRevision, binary *buildv1.BinaryBuildSource) (*buildv1.Build, error) {

	// Need to copy the buildConfig here so that it doesn't share pointers with
	// the build object which could be (will be) modified later.
	buildName := getNextBuildName(bc)
	bcCopy := bc.DeepCopy()
	now := metav1.Now()
	serviceAccount := bcCopy.Spec.ServiceAccount
	if len(serviceAccount) == 0 {
		serviceAccount = bootstrappolicy.BuilderServiceAccountName
	}
	t := true
	build := &buildv1.Build{
		Spec: buildv1.BuildSpec{
			CommonSpec: buildv1.CommonSpec{
				ServiceAccount:            serviceAccount,
				Source:                    bcCopy.Spec.Source,
				Strategy:                  bcCopy.Spec.Strategy,
				Output:                    bcCopy.Spec.Output,
				Revision:                  revision,
				Resources:                 bcCopy.Spec.Resources,
				PostCommit:                bcCopy.Spec.PostCommit,
				CompletionDeadlineSeconds: bcCopy.Spec.CompletionDeadlineSeconds,
				NodeSelector:              bcCopy.Spec.NodeSelector,
				MountTrustedCA:            bcCopy.Spec.MountTrustedCA,
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   buildName,
			Labels: bcCopy.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: buildv1.GroupVersion.String(), // BuildConfig.APIVersion is not populated
					Kind:       "BuildConfig",                 // BuildConfig.Kind is not populated
					Name:       bcCopy.Name,
					UID:        bcCopy.UID,
					Controller: &t,
				},
			},
		},
		Status: buildv1.BuildStatus{
			Phase: buildv1.BuildPhaseNew,
			Conditions: []buildv1.BuildCondition{{
				Type:               buildv1.BuildConditionType(buildv1.BuildPhaseNew),
				Status:             corev1.ConditionTrue,
				LastUpdateTime:     now,
				LastTransitionTime: now,
			}},
			Config: &corev1.ObjectReference{
				Kind:      "BuildConfig",
				Name:      bcCopy.Name,
				Namespace: bcCopy.Namespace,
			},
		},
	}

	setBuildSource(binary, build)
	setBuildAnnotationAndLabel(bcCopy, build)

	var builderSecrets []corev1.Secret
	var err error
	if builderSecrets, err = fetchServiceAccountSecrets(ctx, g.Secrets, g.ServiceAccounts, bcCopy.Namespace, serviceAccount); err != nil {
		return nil, err
	}

	// Resolve image source if present
	if err = g.setBuildSourceImage(ctx, builderSecrets, bcCopy, &build.Spec.Source); err != nil {
		return nil, err
	}
	if err = g.setBaseImageAndPullSecretForBuildStrategy(ctx, builderSecrets, bcCopy, &build.Spec.Strategy); err != nil {
		return nil, err
	}

	return build, nil
}

// setBuildSourceImage set BuildSource Image item for new build
func (g *BuildGenerator) setBuildSourceImage(ctx context.Context, builderSecrets []corev1.Secret, bcCopy *buildv1.BuildConfig, Source *buildv1.BuildSource) error {
	var err error

	strategyImageChangeTrigger := getStrategyImageChangeTrigger(bcCopy)
	for i, sourceImage := range Source.Images {
		if sourceImage.PullSecret == nil {
			sourceImage.PullSecret = g.resolveImageSecret(ctx, builderSecrets, &sourceImage.From, bcCopy.Namespace)
		}

		var sourceImageSpec string
		// if the imagesource matches the strategy from, and we have a trigger for the strategy from,
		// use the imageid from the trigger rather than resolving it.
		if strategyFrom := buildutil.GetInputReference(bcCopy.Spec.Strategy); strategyFrom != nil &&
			reflect.DeepEqual(sourceImage.From, *strategyFrom) &&
			strategyImageChangeTrigger != nil {
			sourceImageSpec = strategyImageChangeTrigger.LastTriggeredImageID
		} else {
			refImageChangeTrigger := getImageChangeTriggerForRef(bcCopy, &sourceImage.From)
			// if there is no trigger associated with this imagesource, resolve the imagesource reference now.
			// otherwise use the imageid from the imagesource trigger.
			if refImageChangeTrigger == nil {
				sourceImageSpec, err = g.resolveImageStreamReference(ctx, sourceImage.From, bcCopy.Namespace)
				if err != nil {
					return err
				}
			} else {
				sourceImageSpec = refImageChangeTrigger.LastTriggeredImageID
			}
		}

		sourceImage.From.Kind = "DockerImage"
		sourceImage.From.Name = sourceImageSpec
		sourceImage.From.Namespace = ""
		Source.Images[i] = sourceImage
	}

	return nil
}

// setBaseImageAndPullSecretForBuildStrategy sets base image and pullSecret items used in buildStrategy for new builds
func (g *BuildGenerator) setBaseImageAndPullSecretForBuildStrategy(ctx context.Context, builderSecrets []corev1.Secret, bcCopy *buildv1.BuildConfig, strategy *buildv1.BuildStrategy) error {
	var err error
	var image string

	if strategyImageChangeTrigger := getStrategyImageChangeTrigger(bcCopy); strategyImageChangeTrigger != nil {
		image = strategyImageChangeTrigger.LastTriggeredImageID
	}
	// If the Build is using a From reference instead of a resolved image, we need to resolve that From
	// reference to a valid image so we can run the build.  Builds do not consume ImageStream references,
	// only image specs.
	switch {
	case strategy.SourceStrategy != nil:
		if image == "" {
			image, err = g.resolveImageStreamReference(ctx, strategy.SourceStrategy.From, bcCopy.Namespace)
			if err != nil {
				return err
			}
		}
		strategy.SourceStrategy.From = corev1.ObjectReference{
			Kind: "DockerImage",
			Name: image,
		}
		if strategy.SourceStrategy.PullSecret == nil {
			strategy.SourceStrategy.PullSecret = g.resolveImageSecret(ctx, builderSecrets, &strategy.SourceStrategy.From, bcCopy.Namespace)
		}
	case strategy.DockerStrategy != nil &&
		strategy.DockerStrategy.From != nil:
		if image == "" {
			image, err = g.resolveImageStreamReference(ctx, *strategy.DockerStrategy.From, bcCopy.Namespace)
			if err != nil {
				return err
			}
		}
		strategy.DockerStrategy.From = &corev1.ObjectReference{
			Kind: "DockerImage",
			Name: image,
		}
		if strategy.DockerStrategy.PullSecret == nil {
			strategy.DockerStrategy.PullSecret = g.resolveImageSecret(ctx, builderSecrets, strategy.DockerStrategy.From, bcCopy.Namespace)
		}
	case strategy.CustomStrategy != nil:
		if image == "" {
			image, err = g.resolveImageStreamReference(ctx, strategy.CustomStrategy.From, bcCopy.Namespace)
			if err != nil {
				return err
			}
		}
		strategy.CustomStrategy.From = corev1.ObjectReference{
			Kind: "DockerImage",
			Name: image,
		}
		if strategy.CustomStrategy.PullSecret == nil {
			strategy.CustomStrategy.PullSecret = g.resolveImageSecret(ctx, builderSecrets, &strategy.CustomStrategy.From, bcCopy.Namespace)
		}
		updateCustomImageEnv(strategy.CustomStrategy, image)
	}
	return nil
}

// resolveImageStreamReference looks up the ImageStream[Tag/Image] and converts it to a
// docker pull spec that can be used in an Image field.
func (g *BuildGenerator) resolveImageStreamReference(ctx context.Context, from corev1.ObjectReference, defaultNamespace string) (string, error) {
	var namespace string
	if len(from.Namespace) != 0 {
		namespace = from.Namespace
	} else {
		namespace = defaultNamespace
	}

	klog.V(4).Infof("Resolving ImageStreamReference %s of Kind %s in namespace %s", from.Name, from.Kind, namespace)
	switch from.Kind {
	case "ImageStreamImage":
		name, id, err := imageutil.ParseImageStreamImageName(from.Name)
		if err != nil {
			err = resolveError(from.Kind, namespace, from.Name, err)
			klog.V(2).Info(err)
			return "", err
		}
		stream, err := g.Client.GetImageStream(apirequest.WithNamespace(ctx, namespace), name, metav1.GetOptions{})
		if err != nil {
			err = resolveError(from.Kind, namespace, from.Name, err)
			klog.V(2).Info(err)
			return "", err
		}
		reference, ok := dockerImageReferenceForImage(stream, id)
		if !ok {
			err = resolveError(from.Kind, namespace, from.Name, fmt.Errorf("unable to find corresponding tag for image %q", id))
			klog.V(2).Info(err)
			return "", err
		}
		klog.V(4).Infof("Resolved ImageStreamImage %s to image %q", from.Name, reference)
		return reference, nil

	case "ImageStreamTag":
		name, tag, err := imageutil.ParseImageStreamTagName(from.Name)
		if err != nil {
			err = resolveError(from.Kind, namespace, from.Name, err)
			klog.V(2).Info(err)
			return "", err
		}
		stream, err := g.Client.GetImageStream(apirequest.WithNamespace(ctx, namespace), name, metav1.GetOptions{})
		if err != nil {
			err = resolveError(from.Kind, namespace, from.Name, err)
			klog.V(2).Info(err)
			return "", err
		}
		reference, ok := imageutil.ResolveLatestTaggedImage(stream, tag)
		if !ok {
			err = resolveError(from.Kind, namespace, from.Name, fmt.Errorf("unable to find latest tagged image"))
			klog.V(2).Info(err)
			return "", err
		}
		klog.V(4).Infof("Resolved ImageStreamTag %s to image %q", from.Name, reference)
		return reference, nil
	case "DockerImage":
		return from.Name, nil
	default:
		return "", fmt.Errorf("unknown From Kind %s", from.Kind)
	}
}

// resolveImageStreamDockerRepository looks up the ImageStream[Tag/Image] and converts it to a
// the docker repository reference with no tag information
func (g *BuildGenerator) resolveImageStreamDockerRepository(ctx context.Context, from corev1.ObjectReference, defaultNamespace string) (string, error) {
	namespace := defaultNamespace
	if len(from.Namespace) > 0 {
		namespace = from.Namespace
	}

	klog.V(4).Infof("Resolving ImageStreamReference %s of Kind %s in namespace %s", from.Name, from.Kind, namespace)
	switch from.Kind {
	case "ImageStreamImage":
		imageStreamImage, err := g.Client.GetImageStreamImage(apirequest.WithNamespace(ctx, namespace), from.Name, metav1.GetOptions{})
		if err != nil {
			err = resolveError(from.Kind, namespace, from.Name, err)
			klog.V(2).Info(err)
			return "", err
		}
		image := imageStreamImage.Image
		klog.V(4).Infof("Resolved ImageStreamReference %s to image %s with reference %s in namespace %s", from.Name, image.Name, image.DockerImageReference, namespace)
		return image.DockerImageReference, nil
	case "ImageStreamTag":
		name := strings.Split(from.Name, ":")[0]
		is, err := g.Client.GetImageStream(apirequest.WithNamespace(ctx, namespace), name, metav1.GetOptions{})
		if err != nil {
			err = resolveError("ImageStream", namespace, from.Name, err)
			klog.V(2).Info(err)
			return "", err
		}
		image, err := dockerImageReferenceForStream(is)
		if err != nil {
			klog.V(2).Infof("Error resolving container image reference for %s/%s: %v", namespace, name, err)
			return "", err
		}
		klog.V(4).Infof("Resolved ImageStreamTag %s/%s to repository %s", namespace, from.Name, image)
		return image.String(), nil
	case "DockerImage":
		return from.Name, nil
	default:
		return "", fmt.Errorf("unknown From Kind %s", from.Kind)
	}
}

// dockerImageReferenceForStream returns a DockerImageReference that represents
// the ImageStream or false, if no valid reference exists.
func dockerImageReferenceForStream(stream *imagev1.ImageStream) (imagev1.DockerImageReference, error) {
	spec := stream.Status.DockerImageRepository
	if len(spec) == 0 {
		spec = stream.Spec.DockerImageRepository
	}
	if len(spec) == 0 {
		return imagev1.DockerImageReference{}, fmt.Errorf("no possible pull spec for %s/%s", stream.Namespace, stream.Name)
	}
	return imageutil.ParseDockerImageReference(spec)
}

// resolveImageSecret looks up the Secrets provided by the Service Account and
// attempt to find a best match for given image.
func (g *BuildGenerator) resolveImageSecret(ctx context.Context, secrets []corev1.Secret, imageRef *corev1.ObjectReference, buildNamespace string) *corev1.LocalObjectReference {
	if len(secrets) == 0 || imageRef == nil {
		return nil
	}
	// Get the image pull spec from the image stream reference
	imageSpec, err := g.resolveImageStreamDockerRepository(ctx, *imageRef, buildNamespace)
	if err != nil {
		klog.V(2).Infof("Unable to resolve the image name for %s/%s: %v", buildNamespace, imageRef, err)
		return nil
	}
	s := findDockerSecretAsReference(secrets, imageSpec)
	if s == nil {
		klog.V(4).Infof("No secrets found for pushing or pulling the %s  %s/%s", imageRef.Kind, buildNamespace, imageRef.Name)
	}
	return s
}

// findDockerSecretAsInternalReference looks through a set of k8s Secrets to find one that represents Docker credentials
// and which contains credentials that are associated with the registry identified by the image.  It returns
// a LocalObjectReference to the Secret, or nil if no match was found.
func findDockerSecretAsReference(secrets []corev1.Secret, image string) *corev1.LocalObjectReference {
	emptyKeyring := credentialprovider.BasicDockerKeyring{}
	for _, secret := range secrets {
		secretList := []corev1.Secret{*secret.DeepCopy()}
		keyring, err := credentialprovidersecrets.MakeDockerKeyring(secretList, &emptyKeyring)
		if err != nil {
			klog.V(2).Infof("Unable to make the Docker keyring for %s/%s secret: %v", secret.Name, secret.Namespace, err)
			continue
		}
		if _, found := keyring.Lookup(image); found {
			return &corev1.LocalObjectReference{Name: secret.Name}
		}
	}
	return nil
}

func resolveError(kind string, namespace string, name string, err error) error {
	msg := fmt.Sprintf("Error resolving %s %s in namespace %s: %v", kind, name, namespace, err)
	return &errors.StatusError{ErrStatus: metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusUnprocessableEntity,
		Reason:  metav1.StatusReasonInvalid,
		Message: msg,
		Details: &metav1.StatusDetails{
			Kind: kind,
			Name: name,
			Causes: []metav1.StatusCause{{
				Field:   "from",
				Message: msg,
			}},
		},
	}}
}

// getNextBuildName returns name of the next build and increments BuildConfig's LastVersion.
func getNextBuildName(buildConfig *buildv1.BuildConfig) string {
	buildConfig.Status.LastVersion++
	return naming.GetName(buildConfig.Name, strconv.FormatInt(buildConfig.Status.LastVersion, 10), kvalidation.DNS1123SubdomainMaxLength)
}

// updateCustomImageEnv updates base image env variable reference with the new image for a custom build strategy.
// If no env variable reference exists, create a new env variable.
func updateCustomImageEnv(strategy *buildv1.CustomBuildStrategy, newImage string) {
	if strategy.Env == nil {
		strategy.Env = make([]corev1.EnvVar, 1)
		strategy.Env[0] = corev1.EnvVar{Name: buildv1.CustomBuildStrategyBaseImageKey, Value: newImage}
	} else {
		found := false
		for i := range strategy.Env {
			klog.V(4).Infof("Checking env variable %s %s", strategy.Env[i].Name, strategy.Env[i].Value)
			if strategy.Env[i].Name == buildv1.CustomBuildStrategyBaseImageKey {
				found = true
				strategy.Env[i].Value = newImage
				klog.V(4).Infof("Updated env variable %s to %s", strategy.Env[i].Name, strategy.Env[i].Value)
				break
			}
		}
		if !found {
			strategy.Env = append(strategy.Env, corev1.EnvVar{Name: buildv1.CustomBuildStrategyBaseImageKey, Value: newImage})
		}
	}
}

// generateBuildFromBuild creates a new build based on a given Build.
func generateBuildFromBuild(build *buildv1.Build, buildConfig *buildv1.BuildConfig) *buildv1.Build {
	buildCopy := build.DeepCopy()
	now := metav1.Now()
	newBuild := &buildv1.Build{
		Spec: buildCopy.Spec,
		ObjectMeta: metav1.ObjectMeta{
			Name:            getNextBuildNameFromBuild(buildCopy, buildConfig),
			Labels:          buildCopy.ObjectMeta.Labels,
			Annotations:     buildCopy.ObjectMeta.Annotations,
			OwnerReferences: buildCopy.ObjectMeta.OwnerReferences,
		},
		Status: buildv1.BuildStatus{
			Phase: buildv1.BuildPhaseNew,
			Conditions: []buildv1.BuildCondition{{
				Type:               buildv1.BuildConditionType(buildv1.BuildPhaseNew),
				Status:             corev1.ConditionTrue,
				LastUpdateTime:     now,
				LastTransitionTime: now,
			}},
			Config: buildCopy.Status.Config,
		},
	}
	// TODO remove/update this when we support cloning binary builds
	// we need to explicitly set type to empty string so that this does not get
	// defaulted to non-empty binary build
	newBuild.Spec.Source.Type = ""
	newBuild.Spec.Source.Binary = nil
	if newBuild.Annotations == nil {
		newBuild.Annotations = make(map[string]string)
	}
	newBuild.Annotations[buildv1.BuildCloneAnnotation] = build.Name
	if buildConfig != nil {
		newBuild.Annotations[buildv1.BuildNumberAnnotation] = strconv.FormatInt(buildConfig.Status.LastVersion, 10)
	} else {
		// builds without a buildconfig don't have build numbers.
		delete(newBuild.Annotations, buildv1.BuildNumberAnnotation)
	}

	// if they exist, Jenkins reporting annotations must be removed when cloning.
	delete(newBuild.Annotations, buildv1.BuildJenkinsStatusJSONAnnotation)
	delete(newBuild.Annotations, buildv1.BuildJenkinsLogURLAnnotation)
	delete(newBuild.Annotations, buildv1.BuildJenkinsConsoleLogURLAnnotation)
	delete(newBuild.Annotations, buildv1.BuildJenkinsBlueOceanLogURLAnnotation)
	delete(newBuild.Annotations, buildv1.BuildJenkinsBuildURIAnnotation)

	// remove the BuildPodNameAnnotation for good measure.
	delete(newBuild.Annotations, buildv1.BuildPodNameAnnotation)

	return newBuild
}

// getNextBuildNameFromBuild returns name of the next build with random uuid added at the end
func getNextBuildNameFromBuild(build *buildv1.Build, buildConfig *buildv1.BuildConfig) string {
	var buildName string
	if buildConfig != nil {
		return getNextBuildName(buildConfig)
	}
	// for builds created by hand, append a timestamp when cloning/rebuilding them
	// because we don't have a sequence number to bump.
	buildName = build.Name
	// remove the old timestamp if we're cloning a build that is itself a clone.
	if matched, _ := regexp.MatchString(`^.+-\d{10}$`, buildName); matched {
		nameElems := strings.Split(buildName, "-")
		buildName = strings.Join(nameElems[:len(nameElems)-1], "-")
	}
	suffix := fmt.Sprintf("%v", metav1.Now().UnixNano())
	if len(suffix) > 10 {
		suffix = suffix[len(suffix)-10:]
	}
	return naming.GetName(buildName, suffix, kvalidation.DNS1123SubdomainMaxLength)

}

// getStrategyImageChangeTrigger returns the ImageChangeTrigger that corresponds to the BuildConfig's strategy
func getStrategyImageChangeTrigger(bc *buildv1.BuildConfig) *buildv1.ImageChangeTriggerStatus {
	strategyRef := buildutil.GetInputReference(bc.Spec.Strategy)
	for _, trigger := range bc.Status.ImageChangeTriggers {
		if reflect.DeepEqual(strategyRef, trigger.From) {
			return &trigger
		}
	}
	return nil
}

// getImageChangeTriggerForRef returns the ImageChangeTrigger that is triggered by a change to
// the provided object reference, if any
func getImageChangeTriggerForRef(bc *buildv1.BuildConfig, ref *corev1.ObjectReference) *buildv1.ImageChangeTriggerStatus {
	if ref == nil || ref.Kind != "ImageStreamTag" {
		return nil
	}
	for _, trigger := range bc.Status.ImageChangeTriggers {
		if trigger.From.Name == ref.Name && trigger.From.Namespace == ref.Namespace {
			return &trigger
		}
	}
	return nil
}

// setBuildSource update build source by binary status
func setBuildSource(binary *buildv1.BinaryBuildSource, build *buildv1.Build) {
	if binary != nil {
		build.Spec.Source.Git = nil
		build.Spec.Source.Binary = binary
		if build.Spec.Source.Dockerfile != nil && binary.AsFile == "Dockerfile" {
			build.Spec.Source.Dockerfile = nil
		}
	} else {
		// must explicitly set this because we copied the source values from the buildconfig.
		// we need to explicitly set type to empty string so that this does not get
		// defaulted to non-empty binary build
		build.Spec.Source.Type = ""
		build.Spec.Source.Binary = nil
	}
}

// setBuildAnnotationAndLabel set annotations and label info of this build
func setBuildAnnotationAndLabel(bcCopy *buildv1.BuildConfig, build *buildv1.Build) {
	if build.Annotations == nil {
		build.Annotations = make(map[string]string)
	}
	// bcCopy.Status.LastVersion has been increased
	build.Annotations[buildv1.BuildNumberAnnotation] = strconv.FormatInt(bcCopy.Status.LastVersion, 10)
	build.Annotations[buildv1.BuildConfigAnnotation] = bcCopy.Name
	if build.Labels == nil {
		build.Labels = make(map[string]string)
	}
	build.Labels[buildv1.BuildConfigLabelDeprecated] = labelValue(bcCopy.Name)
	build.Labels[buildv1.BuildConfigLabel] = labelValue(bcCopy.Name)
	build.Labels[buildv1.BuildRunPolicyLabel] = string(bcCopy.Spec.RunPolicy)
}

func labelValue(name string) string {
	end := len(name)
	newName := name
	// first, try to truncate from the end to find a valid
	// label
	for end > 0 {
		errStrs := validation.IsValidLabelValue(newName)
		if len(errStrs) == 0 {
			return newName
		}
		end--
		newName = newName[:end]
	}
	klog.Warningf("In creating the value of the build label in the build pod, several attempts at manipulating %s to meet k8s label name requirements failed", name)
	return name
}

// mergeMaps will merge to map[string]string instances, with
// keys from the second argument overwriting keys from the
// first argument, in case of duplicates.
func mergeMaps(a, b map[string]string) map[string]string {
	if a == nil && b == nil {
		return nil
	}

	res := make(map[string]string)

	for k, v := range a {
		res[k] = v
	}

	for k, v := range b {
		res[k] = v
	}

	return res
}

// isPaused returns true if the provided BuildConfig is paused and cannot be used to create a new Build
func isPaused(bc *buildv1.BuildConfig) bool {
	return strings.ToLower(bc.Annotations[buildv1.BuildConfigPausedAnnotation]) == "true"
}

// UpdateBuildEnv updates the strategy environment
// This will replace the existing variable definitions with provided env
func updateBuildEnv(build *buildv1.Build, env []corev1.EnvVar) {
	// TODO moving to library-go
	buildEnv := buildutil.GetBuildEnv(build)

	newEnv := []corev1.EnvVar{}
	for _, e := range buildEnv {
		exists := false
		for _, n := range env {
			if e.Name == n.Name {
				exists = true
				break
			}
		}
		if !exists {
			newEnv = append(newEnv, e)
		}
	}
	newEnv = append(newEnv, env...)
	// TODO moving to library-go
	buildutil.SetBuildEnv(build, newEnv)
}

// dockerImageReferenceForImage returns the docker reference for specified image. Assuming
// the image stream contains the image and the image has corresponding tag, this function
// will try to find this tag and take the reference policy into the account.
// If the image stream does not reference the image or the image does not have
// corresponding tag event, this function will return false.
func dockerImageReferenceForImage(stream *imagev1.ImageStream, imageID string) (string, bool) {
	tag, event := latestImageTagEvent(stream, imageID)
	if len(tag) == 0 {
		return "", false
	}
	var ref *imagev1.TagReference
	for _, t := range stream.Spec.Tags {
		if t.Name == tag {
			ref = &t
			break
		}
	}
	if ref == nil {
		return event.DockerImageReference, true
	}
	switch ref.ReferencePolicy.Type {
	case imagev1.LocalTagReferencePolicy:
		ref, err := imageutil.ParseDockerImageReference(stream.Status.DockerImageRepository)
		if err != nil {
			return event.DockerImageReference, true
		}
		ref.Tag = ""
		ref.ID = event.Image
		return dockerImageReferenceExact(ref), true
	default:
		return event.DockerImageReference, true
	}
}

// dockerImageReferenceNameString returns the name of the reference with its tag or ID.
func dockerImageReferenceNameString(r imagev1.DockerImageReference) string {
	switch {
	case len(r.Name) == 0:
		return ""
	case len(r.Tag) > 0:
		return r.Name + ":" + r.Tag
	case len(r.ID) > 0:
		var ref string
		if _, err := imageutil.ParseDigest(r.ID); err == nil {
			// if it parses as a digest, its v2 pull by id
			ref = "@" + r.ID
		} else {
			// if it doesn't parse as a digest, it's presumably a v1 registry by-id tag
			ref = ":" + r.ID
		}
		return r.Name + ref
	default:
		return r.Name
	}
}

// dockerImageReferenceExact returns a string representation of the set fields on the DockerImageReference
func dockerImageReferenceExact(r imagev1.DockerImageReference) string {
	name := dockerImageReferenceNameString(r)
	if len(name) == 0 {
		return name
	}
	s := r.Registry
	if len(s) > 0 {
		s += "/"
	}
	if len(r.Namespace) != 0 {
		s += r.Namespace + "/"
	}
	return s + name
}

// latestImageTagEvent returns the most recent TagEvent and the tag for the specified
// image.
// Copied from v3.7 github.com/openshift/openshift-apiserver/pkg/image/apis/image/v1/helpers.go
func latestImageTagEvent(stream *imagev1.ImageStream, imageID string) (string, *imagev1.TagEvent) {
	var (
		latestTagEvent *imagev1.TagEvent
		latestTag      string
	)
	for _, events := range stream.Status.Tags {
		if len(events.Items) == 0 {
			continue
		}
		tag := events.Tag
		for i, event := range events.Items {
			if imageutil.DigestOrImageMatch(event.Image, imageID) &&
				(latestTagEvent == nil || latestTagEvent != nil && event.Created.After(latestTagEvent.Created.Time)) {
				latestTagEvent = &events.Items[i]
				latestTag = tag
			}
		}
	}
	return latestTag, latestTagEvent
}
