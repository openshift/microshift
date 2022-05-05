package buildconfiginstantiate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	v1 "github.com/openshift/openshift-apiserver/pkg/build/apis/build/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/registry/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	kapi "k8s.io/kubernetes/pkg/apis/core"

	buildv1 "github.com/openshift/api/build/v1"
	buildtypedclient "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	buildinternalhelpers "github.com/openshift/openshift-apiserver/pkg/build/apis/build/internal_helpers"
	"github.com/openshift/openshift-apiserver/pkg/build/apiserver/apiserverbuildutil"
	"github.com/openshift/openshift-apiserver/pkg/build/apiserver/buildgenerator"
	buildwait "github.com/openshift/openshift-apiserver/pkg/build/apiserver/registry/wait"
)

var (
	cancelPollInterval = 500 * time.Millisecond
	cancelPollDuration = 30 * time.Second
)

// NewStorage creates a new storage object for build generation
func NewStorage(generator *buildgenerator.BuildGenerator) *InstantiateREST {
	return &InstantiateREST{generator: generator}
}

// InstantiateREST is a RESTStorage implementation for a BuildGenerator which supports only
// the Create operation (as the generator has no underlying storage object).
type InstantiateREST struct {
	generator *buildgenerator.BuildGenerator
}

var _ rest.Creater = &InstantiateREST{}
var _ rest.StorageMetadata = &InstantiateREST{}

// New creates a new build generation request
func (s *InstantiateREST) New() runtime.Object {
	return &buildapi.BuildRequest{}
}

// Create instantiates a new build from a build configuration
func (s *InstantiateREST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	if err := rest.BeforeCreate(Strategy, ctx, obj); err != nil {
		return nil, err
	}
	if err := createValidation(ctx, obj); err != nil {
		return nil, err
	}

	request := obj.(*buildapi.BuildRequest)
	if len(request.TriggeredBy) == 0 {
		buildTriggerCauses := []buildapi.BuildTriggerCause{}
		request.TriggeredBy = append(buildTriggerCauses,
			buildapi.BuildTriggerCause{
				Message: buildapi.BuildTriggerCauseManualMsg,
			},
		)
	}
	return s.generator.InstantiateInternal(ctx, request, *options)
}

func (s *InstantiateREST) ProducesObject(verb string) interface{} {
	// for documentation purposes
	return buildv1.Build{}
}

func (s *InstantiateREST) ProducesMIMETypes(verb string) []string {
	return nil // no additional mime types
}

func NewBinaryStorage(generator *buildgenerator.BuildGenerator, buildClient buildtypedclient.BuildsGetter, inClientConfig *restclient.Config) *BinaryInstantiateREST {
	clientConfig := restclient.CopyConfig(inClientConfig)
	clientConfig.APIPath = "/api"
	clientConfig.GroupVersion = &schema.GroupVersion{Version: "v1"}
	clientConfig.NegotiatedSerializer = legacyscheme.Codecs

	return &BinaryInstantiateREST{
		Generator:    generator,
		BuildClient:  buildClient,
		ClientConfig: clientConfig,
		Timeout:      5 * time.Minute,
	}
}

type BinaryInstantiateREST struct {
	Generator    *buildgenerator.BuildGenerator
	BuildClient  buildtypedclient.BuildsGetter
	ClientConfig *restclient.Config
	Timeout      time.Duration
}

var _ rest.Connecter = &BinaryInstantiateREST{}
var _ rest.StorageMetadata = &InstantiateREST{}

// New creates a new build generation request
func (r *BinaryInstantiateREST) New() runtime.Object {
	return &buildapi.BinaryBuildRequestOptions{}
}

// Connect returns a ConnectHandler that will handle the request/response for a request
func (r *BinaryInstantiateREST) Connect(ctx context.Context, name string, options runtime.Object, responder rest.Responder) (http.Handler, error) {
	return &binaryInstantiateHandler{
		r:         r,
		responder: responder,
		ctx:       ctx,
		name:      name,
		options:   options.(*buildapi.BinaryBuildRequestOptions),
	}, nil
}

// NewConnectOptions prepares a binary build request.
func (r *BinaryInstantiateREST) NewConnectOptions() (runtime.Object, bool, string) {
	return &buildapi.BinaryBuildRequestOptions{}, false, ""
}

// ConnectMethods returns POST, the only supported binary method.
func (r *BinaryInstantiateREST) ConnectMethods() []string {
	return []string{"POST"}
}

func (r *BinaryInstantiateREST) ProducesObject(verb string) interface{} {
	// for documentation purposes
	return buildv1.Build{}
}

func (r *BinaryInstantiateREST) ProducesMIMETypes(verb string) []string {
	return nil // no additional mime types
}

// binaryInstantiateHandler responds to upload requests
type binaryInstantiateHandler struct {
	r *BinaryInstantiateREST

	responder rest.Responder
	ctx       context.Context
	name      string
	options   *buildapi.BinaryBuildRequestOptions
}

var _ http.Handler = &binaryInstantiateHandler{}

func (h *binaryInstantiateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	build, err := h.handle(r.Body)
	if err != nil {
		h.responder.Error(err)
		return
	}
	h.responder.Object(http.StatusCreated, build)
}

func (h *binaryInstantiateHandler) handle(r io.Reader) (runtime.Object, error) {
	h.options.Name = h.name
	if err := rest.BeforeCreate(BinaryStrategy, h.ctx, h.options); err != nil {
		klog.Infof("failed to validate binary: %#v", h.options)
		return nil, err
	}

	request := &buildapi.BuildRequest{}
	request.Name = h.name
	if len(h.options.Commit) > 0 {
		request.Revision = &buildapi.SourceRevision{
			Git: &buildapi.GitSourceRevision{
				Committer: buildapi.SourceControlUser{
					Name:  h.options.CommitterName,
					Email: h.options.CommitterEmail,
				},
				Author: buildapi.SourceControlUser{
					Name:  h.options.AuthorName,
					Email: h.options.AuthorEmail,
				},
				Message: h.options.Message,
				Commit:  h.options.Commit,
			},
		}
	}
	request.Binary = &buildapi.BinaryBuildSource{
		AsFile: h.options.AsFile,
	}

	var build, result *buildapi.Build
	var instErr error
	start := time.Now()
	if err := wait.Poll(time.Second, h.r.Timeout, func() (bool, error) {
		result, instErr = h.r.Generator.InstantiateInternal(h.ctx, request, metav1.CreateOptions{})
		if instErr != nil {
			if errors.IsNotFound(instErr) {
				if s, ok := instErr.(errors.APIStatus); ok {
					if s.Status().Kind == "imagestreamtags" {
						return false, nil
					}
				}
			}
			klog.V(2).Infof("failed to instantiate %#v with error %v", request, instErr)
			return false, instErr
		}
		build = result
		return true, nil
	}); err != nil {
		klog.Warningf("giving up trying to instantiate %#v due to: %v", request, instErr)
		return nil, instErr
	}
	remaining := h.r.Timeout - time.Since(start)

	// Attempt to cancel the build if it did not start running
	// before we gave up.
	cancel := true
	defer func() {
		if !cancel {
			return
		}
		h.cancelBuild(build)
	}()

	latest, ok, err := buildwait.WaitForRunningBuild(h.r.BuildClient, build.Namespace, build.Name, remaining)

	switch {
	// err checks, no ok check, needs to occur before ref to latest
	case err == buildwait.ErrBuildDeleted:
		errLog := fmt.Sprintf("build %s was deleted before it started: %s", build.Name, apiserverbuildutil.NoBuildLogsMessage)
		klog.Warningf(errLog)
		return nil, errors.NewBadRequest(errLog)
	case err != nil:
		errLog := fmt.Sprintf("unable to wait for build %s to run: %v", build.Name, err)
		klog.Warningf(errLog)
		return nil, errors.NewBadRequest(errLog)
	case !ok:
		errLog := fmt.Sprintf("timed out waiting for build %s to start after %s", build.Name, h.r.Timeout)
		klog.Warningf(errLog)
		return nil, errors.NewTimeoutError(errLog, 0)
	case latest.Status.Phase == buildv1.BuildPhaseError:
		// don't cancel the build if it reached a terminal state on its own
		cancel = false
		errLog := fmt.Sprintf("build %s encountered an error: %s", build.Name, apiserverbuildutil.NoBuildLogsMessage)
		klog.Warningf(errLog)
		return nil, errors.NewBadRequest(errLog)
	case latest.Status.Phase == buildv1.BuildPhaseFailed:
		// don't cancel the build if it reached a terminal state on its own
		cancel = false
		errLog := fmt.Sprintf("build %s failed: %s: %s", build.Name, latest.Status.Reason, latest.Status.Message)
		klog.Warningf(errLog)
		return nil, errors.NewBadRequest(errLog)
	case latest.Status.Phase == buildv1.BuildPhaseCancelled:
		// don't cancel the build if it reached a terminal state on its own
		cancel = false
		errLog := fmt.Sprintf("build %s was cancelled: %s", build.Name, apiserverbuildutil.NoBuildLogsMessage)
		klog.Warningf(errLog)
		return nil, errors.NewBadRequest(errLog)
	case latest.Status.Phase != buildv1.BuildPhaseRunning:
		errLog := fmt.Sprintf("cannot upload file to build %s with status %s", build.Name, latest.Status.Phase)
		klog.Warningf(errLog)
		return nil, errors.NewBadRequest(errLog)
	}

	buildPodName := buildinternalhelpers.GetBuildPodName(build)
	opts := &kapi.PodAttachOptions{
		Stdin:     true,
		Container: apiserverbuildutil.GitCloneContainer,
	}
	// Custom builds don't have a gitclone container, so we inject the source
	// directly into the main container.
	if build.Spec.Strategy.CustomStrategy != nil {
		opts.Container = apiserverbuildutil.CustomBuild
	}

	restClient, err := restclient.RESTClientFor(h.r.ClientConfig)
	if err != nil {
		return nil, err
	}

	// TODO: consider abstracting into a client invocation or client helper
	req := restClient.Post().
		Resource("pods").
		Name(buildPodName).
		Namespace(build.Namespace).
		SubResource("attach")
	req.VersionedParams(opts, legacyscheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(h.r.ClientConfig, "POST", req.URL())
	if err != nil {
		klog.Warningf("error with SPDY POST binary content to build pod %s/%s: %#v", build.Namespace, buildPodName, err)
		return nil, err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin: r,
	})
	if err != nil {
		klog.Warningf("error streaming binary content with build pod %s/%s: %#v", build.Namespace, buildPodName, err)
		return nil, errors.NewInternalError(err)
	}
	cancel = false
	return latest, nil
}

// cancelBuild will mark a build for cancellation unless
// cancel is false in which case it is a no-op.
func (h *binaryInstantiateHandler) cancelBuild(build *buildapi.Build) {
	var versionedBuild = &buildv1.Build{}
	if err := v1.Convert_build_Build_To_v1_Build(build, versionedBuild, nil); err != nil {
		klog.Errorf("Unable to convert build to versioned build: %v", err)
		return
	}
	versionedBuild.Status.Cancelled = true
	h.r.Generator.Client.UpdateBuild(h.ctx, versionedBuild, metav1.UpdateOptions{})
	wait.Poll(cancelPollInterval, cancelPollDuration, func() (bool, error) {
		versionedBuild.Status.Cancelled = true
		err := h.r.Generator.Client.UpdateBuild(h.ctx, versionedBuild, metav1.UpdateOptions{})
		switch {
		case err != nil && errors.IsConflict(err):
			versionedBuild, err = h.r.Generator.Client.GetBuild(h.ctx, versionedBuild.Name, metav1.GetOptions{})
			return false, err
		default:
			return true, err
		}
	})
}
