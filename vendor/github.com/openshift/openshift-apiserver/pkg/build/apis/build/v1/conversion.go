package v1

import (
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/openshift/api/build/v1"
	"github.com/openshift/library-go/pkg/image/imageutil"
	newer "github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	buildinternalhelpers "github.com/openshift/openshift-apiserver/pkg/build/apis/build/internal_helpers"
)

func Convert_v1_BuildConfig_To_build_BuildConfig(in *v1.BuildConfig, out *newer.BuildConfig, s conversion.Scope) error {
	if err := autoConvert_v1_BuildConfig_To_build_BuildConfig(in, out, s); err != nil {
		return err
	}

	newTriggers := []newer.BuildTriggerPolicy{}
	// Strip off any default imagechange triggers where the buildconfig's
	// "from" is not an ImageStreamTag, because those triggers
	// will never be invoked.
	imageRef := buildinternalhelpers.GetInputReference(out.Spec.Strategy)
	hasIST := imageRef != nil && imageRef.Kind == "ImageStreamTag"
	for _, trigger := range out.Spec.Triggers {
		if trigger.Type != newer.ImageChangeBuildTriggerType {
			newTriggers = append(newTriggers, trigger)
			continue
		}
		if (trigger.ImageChange == nil || trigger.ImageChange.From == nil) && !hasIST {
			continue
		}
		newTriggers = append(newTriggers, trigger)
	}
	out.Spec.Triggers = newTriggers
	return nil
}

func Convert_v1_SourceBuildStrategy_To_build_SourceBuildStrategy(in *v1.SourceBuildStrategy, out *newer.SourceBuildStrategy, s conversion.Scope) error {
	if err := autoConvert_v1_SourceBuildStrategy_To_build_SourceBuildStrategy(in, out, s); err != nil {
		return err
	}
	switch in.From.Kind {
	case "ImageStream":
		out.From.Kind = "ImageStreamTag"
		out.From.Name = imageutil.JoinImageStreamTag(in.From.Name, "")
	}
	return nil
}

func Convert_v1_DockerBuildStrategy_To_build_DockerBuildStrategy(in *v1.DockerBuildStrategy, out *newer.DockerBuildStrategy, s conversion.Scope) error {
	if err := autoConvert_v1_DockerBuildStrategy_To_build_DockerBuildStrategy(in, out, s); err != nil {
		return err
	}
	if in.From != nil {
		switch in.From.Kind {
		case "ImageStream":
			out.From.Kind = "ImageStreamTag"
			out.From.Name = imageutil.JoinImageStreamTag(in.From.Name, "")
		}
	}
	return nil
}

func Convert_v1_CustomBuildStrategy_To_build_CustomBuildStrategy(in *v1.CustomBuildStrategy, out *newer.CustomBuildStrategy, s conversion.Scope) error {
	if err := autoConvert_v1_CustomBuildStrategy_To_build_CustomBuildStrategy(in, out, s); err != nil {
		return err
	}
	switch in.From.Kind {
	case "ImageStream":
		out.From.Kind = "ImageStreamTag"
		out.From.Name = imageutil.JoinImageStreamTag(in.From.Name, "")
	}
	return nil
}

func Convert_v1_BuildOutput_To_build_BuildOutput(in *v1.BuildOutput, out *newer.BuildOutput, s conversion.Scope) error {
	if err := autoConvert_v1_BuildOutput_To_build_BuildOutput(in, out, s); err != nil {
		return err
	}
	if in.To != nil && (in.To.Kind == "ImageStream" || len(in.To.Kind) == 0) {
		out.To.Kind = "ImageStreamTag"
		out.To.Name = imageutil.JoinImageStreamTag(in.To.Name, "")
	}
	return nil
}

func Convert_v1_BuildTriggerPolicy_To_build_BuildTriggerPolicy(in *v1.BuildTriggerPolicy, out *newer.BuildTriggerPolicy, s conversion.Scope) error {
	if err := autoConvert_v1_BuildTriggerPolicy_To_build_BuildTriggerPolicy(in, out, s); err != nil {
		return err
	}

	switch in.Type {
	case v1.ImageChangeBuildTriggerTypeDeprecated:
		out.Type = newer.ImageChangeBuildTriggerType
	case v1.GenericWebHookBuildTriggerTypeDeprecated:
		out.Type = newer.GenericWebHookBuildTriggerType
	case v1.GitHubWebHookBuildTriggerTypeDeprecated:
		out.Type = newer.GitHubWebHookBuildTriggerType
	}
	return nil
}

func Convert_build_SourceRevision_To_v1_SourceRevision(in *newer.SourceRevision, out *v1.SourceRevision, s conversion.Scope) error {
	if err := autoConvert_build_SourceRevision_To_v1_SourceRevision(in, out, s); err != nil {
		return err
	}
	out.Type = v1.BuildSourceGit
	return nil
}

func Convert_build_BuildSource_To_v1_BuildSource(in *newer.BuildSource, out *v1.BuildSource, s conversion.Scope) error {
	if err := autoConvert_build_BuildSource_To_v1_BuildSource(in, out, s); err != nil {
		return err
	}
	switch {
	// It is legal for a buildsource to have both a git+dockerfile source, but in v1 that was represented
	// as type git.
	case in.Git != nil:
		out.Type = v1.BuildSourceGit
	// It is legal for a buildsource to have both a binary+dockerfile source, but in v1 that was represented
	// as type binary.
	case in.Binary != nil:
		out.Type = v1.BuildSourceBinary
	case in.Dockerfile != nil:
		out.Type = v1.BuildSourceDockerfile
	case len(in.Images) > 0:
		out.Type = v1.BuildSourceImage
	default:
		out.Type = v1.BuildSourceNone
	}
	return nil
}

func Convert_build_BuildStrategy_To_v1_BuildStrategy(in *newer.BuildStrategy, out *v1.BuildStrategy, s conversion.Scope) error {
	if err := autoConvert_build_BuildStrategy_To_v1_BuildStrategy(in, out, s); err != nil {
		return err
	}
	switch {
	case in.SourceStrategy != nil:
		out.Type = v1.SourceBuildStrategyType
	case in.DockerStrategy != nil:
		out.Type = v1.DockerBuildStrategyType
	case in.CustomStrategy != nil:
		out.Type = v1.CustomBuildStrategyType
	case in.JenkinsPipelineStrategy != nil:
		out.Type = v1.JenkinsPipelineBuildStrategyType
	default:
		out.Type = ""
	}
	return nil
}

func Convert_url_Values_To_v1_BinaryBuildRequestOptions(in *url.Values, out *v1.BinaryBuildRequestOptions, s conversion.Scope) error {
	if in == nil || out == nil {
		return nil
	}
	out.AsFile = in.Get("asFile")
	out.Commit = in.Get("revision.commit")
	out.Message = in.Get("revision.message")
	out.AuthorName = in.Get("revision.authorName")
	out.AuthorEmail = in.Get("revision.authorEmail")
	out.CommitterName = in.Get("revision.committerName")
	out.CommitterEmail = in.Get("revision.committerEmail")
	return nil
}

func Convert_v1_BinaryBuildRequestOptions_To_url_Values(in *v1.BinaryBuildRequestOptions, out *url.Values, s conversion.Scope) error {
	if in == nil || out == nil {
		return nil
	}
	out.Set("asFile", in.AsFile)
	out.Set("revision.commit", in.Commit)
	out.Set("revision.message", in.Message)
	out.Set("revision.authorName", in.AuthorName)
	out.Set("revision.authorEmail", in.AuthorEmail)
	out.Set("revision.committerName", in.CommitterName)
	out.Set("revision.committerEmail", in.CommitterEmail)
	return nil
}

func Convert_url_Values_To_v1_BuildLogOptions(in *url.Values, out *v1.BuildLogOptions, s conversion.Scope) error {
	if in == nil || out == nil {
		return nil
	}

	*out = v1.BuildLogOptions{}

	if values, ok := map[string][]string(*in)["container"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_string(&values, &out.Container, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["follow"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.Follow, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["previous"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.Previous, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["sinceSeconds"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.SinceSeconds, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["sinceTime"]; ok && len(values) > 0 {
		if err := metav1.Convert_Slice_string_To_Pointer_v1_Time(&values, &out.SinceTime, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["timestamps"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.Timestamps, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["tailLines"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.TailLines, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["limitBytes"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.LimitBytes, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["nowait"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.NoWait, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["version"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_Pointer_int64(&values, &out.Version, s); err != nil {
			return err
		}
	}
	if values, ok := map[string][]string(*in)["insecureSkipTLSVerifyBackend"]; ok && len(values) > 0 {
		if err := runtime.Convert_Slice_string_To_bool(&values, &out.InsecureSkipTLSVerifyBackend, s); err != nil {
			return err
		}
	}
	return nil
}

// AddCustomConversionFuncs adds conversion functions which cannot be automatically generated.
// This is typically due to the objects not having 1:1 field mappings.
func AddCustomConversionFuncs(scheme *runtime.Scheme) error {
	if err := scheme.AddConversionFunc((*url.Values)(nil), (*v1.BinaryBuildRequestOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_BinaryBuildRequestOptions(a.(*url.Values), b.(*v1.BinaryBuildRequestOptions), scope)
	}); err != nil {
		return err
	}
	if err := scheme.AddConversionFunc((*v1.BinaryBuildRequestOptions)(nil), (*url.Values)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1_BinaryBuildRequestOptions_To_url_Values(a.(*v1.BinaryBuildRequestOptions), b.(*url.Values), scope)
	}); err != nil {
		return err
	}
	return scheme.AddConversionFunc((*url.Values)(nil), (*v1.BuildLogOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_BuildLogOptions(a.(*url.Values), b.(*v1.BuildLogOptions), scope)
	})
}

func AddFieldSelectorKeyConversions(scheme *runtime.Scheme) error {
	return scheme.AddFieldLabelConversionFunc(v1.GroupVersion.WithKind("Build"), buildFieldSelectorKeyConversionFunc)
}

func buildFieldSelectorKeyConversionFunc(label, value string) (internalLabel, internalValue string, err error) {
	switch label {
	case "status",
		"podName":
		return label, value, nil
	default:
		return runtime.DefaultMetaV1FieldSelectorConversion(label, value)
	}
}
