package validation

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	kpath "k8s.io/apimachinery/pkg/api/validation/path"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	kvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kapi "k8s.io/kubernetes/pkg/apis/core"
	kapihelper "k8s.io/kubernetes/pkg/apis/core/helper"
	"k8s.io/kubernetes/pkg/apis/core/validation"

	buildv1 "github.com/openshift/api/build/v1"
	"github.com/openshift/apiserver-library-go/pkg/labelselector"
	"github.com/openshift/library-go/pkg/image/imageutil"
	imageref "github.com/openshift/library-go/pkg/image/reference"
	buildapi "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	buildinternalhelpers "github.com/openshift/openshift-apiserver/pkg/build/apis/build/internal_helpers"
	imageapivalidation "github.com/openshift/openshift-apiserver/pkg/image/apis/image/validation"
)

// ValidateBuild tests required fields for a Build.
func ValidateBuild(build *buildapi.Build) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMeta(&build.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateCommonSpec(&build.Spec.CommonSpec, field.NewPath("spec"))...)
	return allErrs
}

func ValidateBuildUpdate(build *buildapi.Build, older *buildapi.Build) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMetaUpdate(&build.ObjectMeta, &older.ObjectMeta, field.NewPath("metadata"))...)
	allErrs = append(allErrs, validation.ValidateObjectMeta(&build.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)

	if buildinternalhelpers.IsBuildComplete(older) && older.Status.Phase != build.Status.Phase {
		allErrs = append(allErrs, field.Invalid(field.NewPath("status", "phase"), build.Status.Phase, "phase cannot be updated from a terminal state"))
	}

	// lie about the old build's pushsecret value so we can allow it to be updated.
	olderCopy := older.DeepCopy()
	olderCopy.Spec.Output.PushSecret = build.Spec.Output.PushSecret

	if !kapihelper.Semantic.DeepEqual(build.Spec, olderCopy.Spec) {
		diff, err := diffBuildSpec(build.Spec, olderCopy.Spec)
		if err != nil {
			klog.V(2).Infof("Error calculating build spec patch: %v", err)
			diff = "[unknown]"
		}
		detail := fmt.Sprintf("spec is immutable, diff: %s", diff)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), "content of spec is not printed out, please refer to the details", detail))
	}

	return allErrs
}

// refKey returns a key for the given ObjectReference. If the ObjectReference
// doesn't include a namespace, the passed in namespace is used for the reference
func refKey(namespace string, ref *kapi.ObjectReference) string {
	if ref == nil || ref.Kind != "ImageStreamTag" {
		return "nil"
	}
	ns := ref.Namespace
	if ns == "" {
		ns = namespace
	}
	return fmt.Sprintf("%s/%s", ns, ref.Name)
}

// ValidateBuildConfig tests required fields for a Build.
func ValidateBuildConfig(config *buildapi.BuildConfig) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMeta(&config.ObjectMeta, true, apimachineryvalidation.NameIsDNSSubdomain, field.NewPath("metadata"))...)

	// image change triggers that refer
	fromRefs := map[string]struct{}{}
	specPath := field.NewPath("spec")
	triggersPath := specPath.Child("triggers")
	buildFrom := buildinternalhelpers.GetInputReference(config.Spec.Strategy)
	for i, trg := range config.Spec.Triggers {
		allErrs = append(allErrs, validateTrigger(&trg, buildFrom, triggersPath.Index(i))...)
		if trg.Type != buildapi.ImageChangeBuildTriggerType || trg.ImageChange == nil {
			continue
		}
		from := trg.ImageChange.From
		if from == nil {
			from = buildFrom
		}
		fromKey := refKey(config.Namespace, from)
		_, exists := fromRefs[fromKey]
		if exists {
			allErrs = append(allErrs, field.Invalid(triggersPath, config.Spec.Triggers, "multiple ImageChange triggers refer to the same image stream tag"))
		}
		fromRefs[fromKey] = struct{}{}
	}

	switch config.Spec.RunPolicy {
	case buildapi.BuildRunPolicyParallel, buildapi.BuildRunPolicySerial, buildapi.BuildRunPolicySerialLatestOnly:
	default:
		allErrs = append(allErrs, field.Invalid(specPath.Child("runPolicy"), config.Spec.RunPolicy,
			"run policy must Parallel, Serial, or SerialLatestOnly"))
	}

	if config.Spec.SuccessfulBuildsHistoryLimit != nil {
		allErrs = append(allErrs, validation.ValidateNonnegativeField(int64(*config.Spec.SuccessfulBuildsHistoryLimit), specPath.Child("successfulBuildsHistoryLimit"))...)
	}

	if config.Spec.FailedBuildsHistoryLimit != nil {
		allErrs = append(allErrs, validation.ValidateNonnegativeField(int64(*config.Spec.FailedBuildsHistoryLimit), specPath.Child("failedBuildsHistoryLimit"))...)
	}

	allErrs = append(allErrs, validateCommonSpec(&config.Spec.CommonSpec, specPath)...)

	return allErrs
}

func ValidateBuildConfigUpdate(config *buildapi.BuildConfig, older *buildapi.BuildConfig) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validation.ValidateObjectMetaUpdate(&config.ObjectMeta, &older.ObjectMeta, field.NewPath("metadata"))...)
	allErrs = append(allErrs, ValidateBuildConfig(config)...)
	return allErrs
}

// ValidateBuildRequest validates a BuildRequest object
func ValidateBuildRequest(request *buildapi.BuildRequest) field.ErrorList {
	return validation.ValidateObjectMeta(&request.ObjectMeta, true, kpath.ValidatePathSegmentName, field.NewPath("metadata"))
}

func validateCommonSpec(spec *buildapi.CommonSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	s := spec.Strategy

	if s.DockerStrategy != nil && s.JenkinsPipelineStrategy == nil && spec.Source.Git == nil && spec.Source.Binary == nil && spec.Source.Dockerfile == nil && spec.Source.Images == nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("source"), "", "must provide a value for at least one source input(git, binary, dockerfile, images)."))
	}

	allErrs = append(allErrs,
		validateSource(
			&spec.Source,
			s.CustomStrategy != nil,
			s.DockerStrategy != nil,
			s.JenkinsPipelineStrategy != nil && len(s.JenkinsPipelineStrategy.Jenkinsfile) == 0,
			fldPath.Child("source"))...,
	)

	if spec.CompletionDeadlineSeconds != nil {
		if *spec.CompletionDeadlineSeconds <= 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("completionDeadlineSeconds"), spec.CompletionDeadlineSeconds, "completionDeadlineSeconds must be a positive integer greater than 0"))
		}
	}

	allErrs = append(allErrs, validateOutput(&spec.Output, fldPath.Child("output"))...)
	allErrs = append(allErrs, validateStrategy(&spec.Strategy, fldPath.Child("strategy"))...)
	allErrs = append(allErrs, validatePostCommit(spec.PostCommit, fldPath.Child("postCommit"))...)
	allErrs = append(allErrs, ValidateNodeSelector(spec.NodeSelector, fldPath.Child("nodeSelector"))...)

	// TODO: validate resource requirements (prereq: https://github.com/kubernetes/kubernetes/pull/7059)
	return allErrs
}

const (
	maxDockerfileLengthBytes  = 60 * 1000
	maxJenkinsfileLengthBytes = 100 * 1000
)

func validateSource(input *buildapi.BuildSource, isCustomStrategy, isDockerStrategy, isJenkinsPipelineStrategyFromRepo bool, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// Ensure that Git and Binary source types are mutually exclusive.
	if input.Git != nil && input.Binary != nil && !isCustomStrategy {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("git"), "", "may not be set when binary is also set"))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("binary"), "", "may not be set when git is also set"))
		return allErrs
	}

	// Validate individual source type details
	if input.Git != nil {
		allErrs = append(allErrs, validateGitSource(input.Git, fldPath.Child("git"))...)
	}
	if input.Binary != nil {
		allErrs = append(allErrs, validateBinarySource(input.Binary, fldPath.Child("binary"))...)
	}
	if input.Dockerfile != nil {
		allErrs = append(allErrs, validateDockerfile(*input.Dockerfile, fldPath.Child("dockerfile"))...)
	}
	if input.Images != nil {
		for i, image := range input.Images {
			allErrs = append(allErrs, validateImageSource(image, fldPath.Child("images").Index(i))...)
		}
		// validate that no duplicate image sources exist, all other checks happen in validateImageSource
		var set sets.String
		for i, image := range input.Images {
			for j, name := range image.As {
				if len(name) == 0 {
					continue
				}
				if set.Has(name) {
					allErrs = append(allErrs, field.Duplicate(fldPath.Child("images").Index(i).Child("as").Index(j), name))
					continue
				}
				if set == nil {
					set = sets.NewString()
				}
				set.Insert(name)
			}
		}
	}
	if isJenkinsPipelineStrategyFromRepo && input.Git == nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("git"), "", "must be set when using Jenkins Pipeline strategy with Jenkinsfile from a git repo"))
	}

	allErrs = append(allErrs, validateSecrets(input.Secrets, isDockerStrategy, fldPath.Child("secrets"))...)
	allErrs = append(allErrs, validateConfigMaps(input.ConfigMaps, isDockerStrategy, fldPath.Child("configMaps"))...)

	allErrs = append(allErrs, validateSecretRef(input.SourceSecret, fldPath.Child("sourceSecret"))...)

	if len(input.ContextDir) != 0 {
		cleaned := path.Clean(input.ContextDir)
		if strings.HasPrefix(cleaned, "..") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("contextDir"), input.ContextDir, "context dir must not be a relative path"))
		} else {
			if cleaned == "." {
				cleaned = ""
			}
			input.ContextDir = cleaned
		}
	}

	return allErrs
}

func validateDockerfile(dockerfile string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(dockerfile) > maxDockerfileLengthBytes {
		allErrs = append(allErrs, field.Invalid(fldPath, "", fmt.Sprintf("must be smaller than %d bytes", maxDockerfileLengthBytes)))
	}
	return allErrs
}

func validateSecretRef(ref *kapi.LocalObjectReference, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if ref == nil {
		return allErrs
	}
	if len(ref.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
	}
	return allErrs
}

func validateGitSource(git *buildapi.GitBuildSource, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(git.URI) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("uri"), ""))
	} else if _, err := parseGitURL(git.URI); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("uri"), git.URI, err.Error()))
	}
	if git.HTTPProxy != nil && len(*git.HTTPProxy) != 0 {
		if _, err := parseProxyURL(*git.HTTPProxy); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("httpproxy"), *git.HTTPProxy, err.Error()))
		}
	}
	if git.HTTPSProxy != nil && len(*git.HTTPSProxy) != 0 {
		if _, err := parseProxyURL(*git.HTTPSProxy); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("httpsproxy"), *git.HTTPSProxy, err.Error()))
		}
	}
	return allErrs
}

// ParseProxyURL parses a proxy URL and allows fallback to non-URLs like
// myproxy:80 (for example) which url.Parse no longer accepts in Go 1.8.  The
// logic is copied from net/http.ProxyFromEnvironment to try to maintain
// backwards compatibility.
func parseProxyURL(proxy string) (*url.URL, error) {
	proxyURL, err := url.Parse(proxy)

	// logic copied from net/http.ProxyFromEnvironment
	if err != nil || !strings.HasPrefix(proxyURL.Scheme, "http") {
		// proxy was bogus. Try prepending "http://" to it and see if that
		// parses correctly. If not, we fall through and complain about the
		// original one.
		if proxyURL, err := url.Parse("http://" + proxy); err == nil {
			return proxyURL, nil
		}
	}

	return proxyURL, err
}

func validateConfigMaps(configs []buildapi.ConfigMapBuildSource, isDockerStrategy bool, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for i, c := range configs {
		if len(c.ConfigMap.Name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Index(i).Child("configMap"), ""))
		}
		if reasons := validation.ValidateConfigMapName(c.ConfigMap.Name, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("configMap"), c, "must be valid configMap name"))
		}
		if strings.HasPrefix(path.Clean(c.DestinationDir), "..") {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("destinationDir"), c.DestinationDir, "destination dir cannot start with '..'"))
		}
		if isDockerStrategy && filepath.IsAbs(c.DestinationDir) {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("destinationDir"), c.DestinationDir, "for the docker strategy the destinationDir has to be relative path"))
		}
	}
	return allErrs
}

func validateSecrets(secrets []buildapi.SecretBuildSource, isDockerStrategy bool, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for i, s := range secrets {
		if len(s.Secret.Name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Index(i).Child("secret"), ""))
		}
		if reasons := validation.ValidateSecretName(s.Secret.Name, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("secret"), s, "must be valid secret name"))
		}
		if strings.HasPrefix(path.Clean(s.DestinationDir), "..") {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("destinationDir"), s.DestinationDir, "destination dir cannot start with '..'"))
		}
		if isDockerStrategy && filepath.IsAbs(s.DestinationDir) {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(i).Child("destinationDir"), s.DestinationDir, "for the docker strategy the destinationDir has to be relative path"))
		}
	}
	return allErrs
}

func validateImageSource(imageSource buildapi.ImageSource, fldPath *field.Path) field.ErrorList {
	allErrs := validateImageReference(&imageSource.From, fldPath.Child("from"))
	if imageSource.PullSecret != nil {
		allErrs = append(allErrs, validateSecretRef(imageSource.PullSecret, fldPath.Child("pullSecret"))...)
	}
	if len(imageSource.Paths) == 0 && len(imageSource.As) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("paths"), "must specify 'paths' or 'as'"))
		allErrs = append(allErrs, field.Required(fldPath.Child("as"), "must specify 'paths' or 'as'"))
	}
	for i, path := range imageSource.Paths {
		allErrs = append(allErrs, validateImageSourcePath(path, fldPath.Child("paths").Index(i))...)
	}
	for i, name := range imageSource.As {
		if len(name) == 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("as").Index(i), name, "must not be empty"))
			continue
		}
	}
	return allErrs
}

func validateImageSourcePath(imagePath buildapi.ImageSourcePath, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(imagePath.SourcePath) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("sourcePath"), ""))
	}
	if len(imagePath.DestinationDir) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("destinationDir"), ""))
	}
	if !filepath.IsAbs(imagePath.SourcePath) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("sourcePath"), imagePath.SourcePath, "must be an absolute path"))
	}
	if filepath.IsAbs(imagePath.DestinationDir) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("destinationDir"), imagePath.DestinationDir, "must be a relative path"))
	}
	if strings.HasPrefix(path.Clean(imagePath.DestinationDir), "..") {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("destinationDir"), imagePath.DestinationDir, "destination dir cannot start with '..'"))
	}
	return allErrs
}

func validateBinarySource(source *buildapi.BinaryBuildSource, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(source.AsFile) != 0 {
		cleaned := strings.TrimPrefix(path.Clean(source.AsFile), "/")
		if len(cleaned) == 0 || cleaned == "." || strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/") || strings.Contains(cleaned, "\\") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("asFile"), source.AsFile, "file name may not contain slashes or relative path segments and must be a valid POSIX filename"))
		} else {
			source.AsFile = cleaned
		}
	}
	return allErrs
}

func validateToImageReference(reference *kapi.ObjectReference, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	kind, name, namespace := reference.Kind, reference.Name, reference.Namespace
	switch kind {
	case "ImageStreamTag":
		if len(name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
		} else if _, _, ok := imageutil.SplitImageStreamTag(name); !ok {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), name, "ImageStreamTag object references must be in the form <name>:<tag>"))
		} else if name, _, _ := imageutil.SplitImageStreamTag(name); len(imageapivalidation.ValidateImageStreamName(name, false)) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), name, "ImageStreamTag name contains invalid syntax"))
		}
		if len(namespace) != 0 && len(kvalidation.IsDNS1123Subdomain(namespace)) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, "namespace must be a valid subdomain"))
		}

	case "DockerImage":
		if len(namespace) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, "namespace is not valid when used with a 'DockerImage'"))
		}
		if _, err := imageref.Parse(name); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), name, fmt.Sprintf("name is not a valid Docker pull specification: %v", err)))
		}
	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("kind"), ""))
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kind"), kind, "the target of build output must be an 'ImageStreamTag' or 'DockerImage'"))

	}
	return allErrs
}

func validateImageReference(reference *kapi.ObjectReference, fldPath *field.Path) field.ErrorList {
	if reference == nil {
		return nil
	}
	allErrs := field.ErrorList{}
	kind, name, namespace := reference.Kind, reference.Name, reference.Namespace
	switch kind {
	case "ImageStreamTag":
		if len(name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
		} else if _, _, ok := imageutil.SplitImageStreamTag(name); !ok {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), name, "must be <name>:<tag>"))
		} else if name, _, _ := imageutil.SplitImageStreamTag(name); len(imageapivalidation.ValidateImageStreamName(name, false)) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), name, "invalid name syntax"))
		}

		if len(namespace) != 0 && len(kvalidation.IsDNS1123Subdomain(namespace)) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, "must be a valid namespace"))
		}

	case "DockerImage":
		if len(namespace) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, "not valid when used with a 'DockerImage'"))
		}
		if len(name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
		} else if _, err := imageref.Parse(name); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), name, fmt.Sprintf("not a valid Docker pull specification: %v", err)))
		}
	case "ImageStreamImage":
		if len(name) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("name"), ""))
		}
		if len(namespace) != 0 && len(kvalidation.IsDNS1123Subdomain(namespace)) != 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("namespace"), namespace, "must be a valid namespace"))
		}
	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("kind"), ""))
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kind"), kind, "must be 'ImageStreamTag', 'ImageStreamImage', or 'DockerImage'"))

	}
	return allErrs
}

func validateOutput(output *buildapi.BuildOutput, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// TODO: make part of a generic ValidateObjectReference method upstream.
	if output.To != nil {
		allErrs = append(allErrs, validateToImageReference(output.To, fldPath.Child("to"))...)
	}

	allErrs = append(allErrs, validateSecretRef(output.PushSecret, fldPath.Child("pushSecret"))...)
	allErrs = append(allErrs, ValidateImageLabels(output.ImageLabels, fldPath.Child("imageLabels"))...)

	return allErrs
}

func validateStrategy(strategy *buildapi.BuildStrategy, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	strategyCount := 0
	if strategy.SourceStrategy != nil {
		strategyCount++
	}
	if strategy.DockerStrategy != nil {
		strategyCount++
	}
	if strategy.CustomStrategy != nil {
		strategyCount++
	}
	if strategy.JenkinsPipelineStrategy != nil {
		strategyCount++
	}
	if strategyCount != 1 {
		return append(allErrs, field.Invalid(fldPath, strategy, "must provide a value for exactly one of sourceStrategy, customStrategy, dockerStrategy, or jenkinsPipelineStrategy"))
	}

	if strategy.SourceStrategy != nil {
		allErrs = append(allErrs, validateSourceStrategy(strategy.SourceStrategy, fldPath.Child("sourceStrategy"))...)
	}
	if strategy.DockerStrategy != nil {
		allErrs = append(allErrs, validateDockerStrategy(strategy.DockerStrategy, fldPath.Child("dockerStrategy"))...)
	}
	if strategy.CustomStrategy != nil {
		allErrs = append(allErrs, validateCustomStrategy(strategy.CustomStrategy, fldPath.Child("customStrategy"))...)
	}
	if strategy.JenkinsPipelineStrategy != nil {
		allErrs = append(allErrs, validateJenkinsPipelineStrategy(strategy.JenkinsPipelineStrategy, fldPath.Child("jenkinsPipelineStrategy"))...)
	}

	return allErrs
}

func validateDockerStrategy(strategy *buildapi.DockerBuildStrategy, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if strategy.From != nil {
		allErrs = append(allErrs, validateImageReference(strategy.From, fldPath.Child("from"))...)
	}

	allErrs = append(allErrs, validateSecretRef(strategy.PullSecret, fldPath.Child("pullSecret"))...)

	switch t := strategy.ImageOptimizationPolicy; {
	case t == nil:
	case *t == buildapi.ImageOptimizationSkipLayers, *t == buildapi.ImageOptimizationSkipLayersAndWarn,
		*t == buildapi.ImageOptimizationNone:
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("imageOptimizationPolicy"), *t, "must be unset, 'None', 'SkipLayers', or 'SkipLayersAndWarn'"))
	}

	if len(strategy.DockerfilePath) != 0 {
		cleaned, errs := validateRelativePath(strategy.DockerfilePath, "dockerfilePath", fldPath.Child("dockerfilePath"))
		allErrs = append(allErrs, errs...)
		if len(errs) == 0 {
			strategy.DockerfilePath = cleaned
		}
	}

	allErrs = append(allErrs, ValidateStrategyEnv(strategy.Env, fldPath.Child("env"))...)

	return allErrs
}

func validateRelativePath(filePath, fieldName string, fldPath *field.Path) (string, field.ErrorList) {
	allErrs := field.ErrorList{}
	cleaned := path.Clean(filePath)
	switch {
	case filepath.IsAbs(cleaned), cleaned == "..", strings.HasPrefix(cleaned, "../"):
		allErrs = append(allErrs, field.Invalid(fldPath, filePath, fieldName+" must be a relative path within your source location"))
	default:
		if cleaned == "." {
			cleaned = ""
		}
	}
	return cleaned, allErrs
}

func validateSourceStrategy(strategy *buildapi.SourceBuildStrategy, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateImageReference(&strategy.From, fldPath.Child("from"))...)
	allErrs = append(allErrs, validateSecretRef(strategy.PullSecret, fldPath.Child("pullSecret"))...)
	allErrs = append(allErrs, ValidateStrategyEnv(strategy.Env, fldPath.Child("env"))...)
	return allErrs
}

func validateCustomStrategy(strategy *buildapi.CustomBuildStrategy, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateImageReference(&strategy.From, fldPath.Child("from"))...)
	allErrs = append(allErrs, validateSecretRef(strategy.PullSecret, fldPath.Child("pullSecret"))...)
	allErrs = append(allErrs, ValidateStrategyEnv(strategy.Env, fldPath.Child("env"))...)
	return allErrs
}

func validateJenkinsPipelineStrategy(strategy *buildapi.JenkinsPipelineBuildStrategy, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(strategy.JenkinsfilePath) != 0 && len(strategy.Jenkinsfile) != 0 {
		return append(allErrs, field.Invalid(fldPath, strategy, "must provide a value for at most one of jenkinsfilePath, or jenkinsfile"))
	}

	if len(strategy.Jenkinsfile) > maxJenkinsfileLengthBytes {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("jenkinsfile"), "", fmt.Sprintf("must be smaller than %d bytes", maxJenkinsfileLengthBytes)))
	}

	if len(strategy.JenkinsfilePath) != 0 {
		cleaned, errs := validateRelativePath(strategy.JenkinsfilePath, "jenkinsfilePath", fldPath.Child("jenkinsfilePath"))
		allErrs = append(allErrs, errs...)
		if len(errs) == 0 {
			strategy.JenkinsfilePath = cleaned
		}
	}

	allErrs = append(allErrs, ValidateStrategyEnv(strategy.Env, fldPath.Child("env"))...)

	return allErrs
}

func validateTrigger(trigger *buildapi.BuildTriggerPolicy, buildFrom *kapi.ObjectReference, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(trigger.Type) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("type"), ""))
		return allErrs
	}

	// Validate each trigger type
	switch trigger.Type {
	case buildapi.GitHubWebHookBuildTriggerType:
		if trigger.GitHubWebHook == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("github"), ""))
		} else {
			allErrs = append(allErrs, validateWebHook(trigger.GitHubWebHook, fldPath.Child("github"), false)...)
		}
	case buildapi.GitLabWebHookBuildTriggerType:
		if trigger.GitLabWebHook == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("gitlab"), ""))
		} else {
			allErrs = append(allErrs, validateWebHook(trigger.GitLabWebHook, fldPath.Child("gitlab"), false)...)
		}
	case buildapi.BitbucketWebHookBuildTriggerType:
		if trigger.BitbucketWebHook == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("bitbucket"), ""))
		} else {
			allErrs = append(allErrs, validateWebHook(trigger.BitbucketWebHook, fldPath.Child("bitbucket"), false)...)
		}
	case buildapi.GenericWebHookBuildTriggerType:
		if trigger.GenericWebHook == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("generic"), ""))
		} else {
			allErrs = append(allErrs, validateWebHook(trigger.GenericWebHook, fldPath.Child("generic"), true)...)
		}
	case buildapi.ImageChangeBuildTriggerType:
		if trigger.ImageChange == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("imageChange"), ""))
			break
		}
		if trigger.ImageChange.From == nil {
			if buildFrom == nil || buildFrom.Kind != "ImageStreamTag" {
				invalidKindErr := field.Invalid(
					fldPath.Child("imageChange"),
					fmt.Sprintf("build from: %v", buildFrom),
					"a default ImageChange trigger can only be used when the build strategy includes an ImageStreamTag reference.")
				allErrs = append(allErrs, invalidKindErr)
				break
			}

			break
		}
		if kind := trigger.ImageChange.From.Kind; kind != "ImageStreamTag" {
			invalidKindErr := field.Invalid(
				fldPath.Child("imageChange").Child("from").Child("kind"),
				kind,
				"only an ImageStreamTag type of reference is allowed in an ImageChange trigger.")
			allErrs = append(allErrs, invalidKindErr)
			break
		}
		allErrs = append(allErrs, validateImageReference(trigger.ImageChange.From, fldPath.Child("from"))...)
	case buildapi.ConfigChangeBuildTriggerType:
		// doesn't require additional validation
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("type"), trigger.Type, "invalid trigger type"))
	}
	return allErrs
}

func validateWebHook(webHook *buildapi.WebHookTrigger, fldPath *field.Path, isGeneric bool) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(webHook.Secret) == 0 && webHook.SecretReference == nil {
		allErrs = append(allErrs, field.Invalid(fldPath, webHook, "must provide a value for at least one of secret or secretReference"))
	}
	if webHook.SecretReference != nil && len(webHook.SecretReference.Name) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("secretReference").Child("name"), ""))
	}
	if !isGeneric && webHook.AllowEnv {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("allowEnv"), webHook, "git webhooks cannot allow env vars"))
	}
	return allErrs
}

func ValidateBuildLogOptions(opts *buildapi.BuildLogOptions) field.ErrorList {
	allErrs := field.ErrorList{}

	// TODO: Replace by validating PodLogOptions via BuildLogOptions once it's bundled in
	popts := buildinternalhelpers.BuildToPodLogOptions(opts)
	if errs := validation.ValidatePodLogOptions(popts); len(errs) > 0 {
		allErrs = append(allErrs, errs...)
	}

	if opts.Version != nil && *opts.Version <= 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("version"), *opts.Version, "build version must be greater than 0"))
	}
	if opts.Version != nil && opts.Previous {
		allErrs = append(allErrs, field.Invalid(field.NewPath("previous"), opts.Previous, "cannot use previous when a version is specified"))
	}
	return allErrs
}

func ValidateStrategyEnv(vars []kapi.EnvVar, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, ev := range vars {
		idxPath := fldPath.Index(i)
		if len(ev.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), ""))
		} else if errs := kvalidation.IsEnvVarName(ev.Name); len(errs) != 0 {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("name"), ev.Name, strings.Join(errs, "; ")))
		}
		if ev.ValueFrom != nil && ev.ValueFrom.ResourceFieldRef != nil {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("valueFrom").Child("ResourceFieldRef"), ev.Name, "ResourceFieldRef is not valid for valueFrom in build environment variables"))
		}
	}
	return allErrs
}

func validatePostCommit(spec buildapi.BuildPostCommitSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(spec.Script) != 0 && len(spec.Command) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, spec, "cannot use command and script together"))
	}
	return allErrs
}

func ValidateImageLabels(labels []buildapi.ImageLabel, fldPath *field.Path) (allErrs field.ErrorList) {
	for i, lbl := range labels {
		idxPath := fldPath.Index(i)
		if len(lbl.Name) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("name"), ""))
			continue
		}
		for _, msg := range kvalidation.IsConfigMapKey(lbl.Name) {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("name"), lbl.Name, msg))
		}
	}

	// find duplicates
	seen := make(map[string]bool)
	for i, lbl := range labels {
		idxPath := fldPath.Index(i)
		if seen[lbl.Name] {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("name"), lbl.Name, "duplicate name"))
			continue
		}
		seen[lbl.Name] = true
	}

	return
}

func ValidateNodeSelector(nodeSelector map[string]string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for k, v := range nodeSelector {
		_, err := labelselector.Parse(fmt.Sprintf("%s=%s", k, v))
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Key(k),
				nodeSelector[k], "must be a valid node selector"))
		}
	}
	return allErrs
}

func diffBuildSpec(newer, older buildapi.BuildSpec) (string, error) {
	newerObj := &buildapi.Build{Spec: newer}
	olderObj := &buildapi.Build{Spec: older}
	diffBytes, err := CreateBuildPatch(olderObj, newerObj)
	if err != nil {
		return "", err
	}
	return string(diffBytes), nil
}

func CreateBuildPatch(older, newer *buildapi.Build) ([]byte, error) {
	newerJSON, err := runtime.Encode(encoder, newer)
	if err != nil {
		return nil, fmt.Errorf("error encoding newer: %v", err)
	}
	olderJSON, err := runtime.Encode(encoder, older)
	if err != nil {
		return nil, fmt.Errorf("error encoding older: %v", err)
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(olderJSON, newerJSON, &buildv1.Build{})
	if err != nil {
		return nil, fmt.Errorf("error creating a strategic patch: %v", err)
	}
	return patch, nil
}
