package v1

import (
	"net/url"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/openshift/api/apps/v1"
	"github.com/openshift/library-go/pkg/image/imageutil"
	newer "github.com/openshift/openshift-apiserver/pkg/apps/apis/apps"
)

func Convert_v1_DeploymentTriggerImageChangeParams_To_apps_DeploymentTriggerImageChangeParams(in *v1.DeploymentTriggerImageChangeParams, out *newer.DeploymentTriggerImageChangeParams, s conversion.Scope) error {
	if err := autoConvert_v1_DeploymentTriggerImageChangeParams_To_apps_DeploymentTriggerImageChangeParams(in, out, s); err != nil {
		return err
	}
	switch in.From.Kind {
	case "ImageStreamTag":
	case "ImageStream", "ImageRepository":
		out.From.Kind = "ImageStreamTag"
		if !strings.Contains(out.From.Name, ":") {
			out.From.Name = imageutil.JoinImageStreamTag(out.From.Name, imageutil.DefaultImageTag)
		}
	default:
		// Will be handled by validation
	}
	return nil
}

func Convert_apps_DeploymentTriggerImageChangeParams_To_v1_DeploymentTriggerImageChangeParams(in *newer.DeploymentTriggerImageChangeParams, out *v1.DeploymentTriggerImageChangeParams, s conversion.Scope) error {
	if err := autoConvert_apps_DeploymentTriggerImageChangeParams_To_v1_DeploymentTriggerImageChangeParams(in, out, s); err != nil {
		return err
	}
	switch in.From.Kind {
	case "ImageStreamTag":
	case "ImageStream", "ImageRepository":
		out.From.Kind = "ImageStreamTag"
		if !strings.Contains(out.From.Name, ":") {
			out.From.Name = imageutil.JoinImageStreamTag(out.From.Name, imageutil.DefaultImageTag)
		}
	default:
		// Will be handled by validation
	}
	return nil
}

func Convert_v1_RollingDeploymentStrategyParams_To_apps_RollingDeploymentStrategyParams(in *v1.RollingDeploymentStrategyParams, out *newer.RollingDeploymentStrategyParams, s conversion.Scope) error {
	SetDefaults_RollingDeploymentStrategyParams(in)

	out.UpdatePeriodSeconds = in.UpdatePeriodSeconds
	out.IntervalSeconds = in.IntervalSeconds
	out.TimeoutSeconds = in.TimeoutSeconds

	if in.Pre != nil {
		out.Pre = &newer.LifecycleHook{}
		if err := Convert_v1_LifecycleHook_To_apps_LifecycleHook(in.Pre, out.Pre, s); err != nil {
			return err
		}
	}
	if in.Post != nil {
		out.Post = &newer.LifecycleHook{}
		if err := Convert_v1_LifecycleHook_To_apps_LifecycleHook(in.Post, out.Post, s); err != nil {
			return err
		}
	}
	if in.MaxUnavailable != nil {
		if err := metav1.Convert_intstr_IntOrString_To_intstr_IntOrString(in.MaxUnavailable, &out.MaxUnavailable, s); err != nil {
			return err
		}
	}
	if in.MaxSurge != nil {
		if err := metav1.Convert_intstr_IntOrString_To_intstr_IntOrString(in.MaxSurge, &out.MaxSurge, s); err != nil {
			return err
		}
	}

	return nil
}

func Convert_apps_RollingDeploymentStrategyParams_To_v1_RollingDeploymentStrategyParams(in *newer.RollingDeploymentStrategyParams, out *v1.RollingDeploymentStrategyParams, s conversion.Scope) error {
	out.UpdatePeriodSeconds = in.UpdatePeriodSeconds
	out.IntervalSeconds = in.IntervalSeconds
	out.TimeoutSeconds = in.TimeoutSeconds

	if in.Pre != nil {
		out.Pre = &v1.LifecycleHook{}
		if err := Convert_apps_LifecycleHook_To_v1_LifecycleHook(in.Pre, out.Pre, s); err != nil {
			return err
		}
	}
	if in.Post != nil {
		out.Post = &v1.LifecycleHook{}
		if err := Convert_apps_LifecycleHook_To_v1_LifecycleHook(in.Post, out.Post, s); err != nil {
			return err
		}
	}

	if out.MaxUnavailable == nil {
		out.MaxUnavailable = &intstr.IntOrString{}
	}
	if out.MaxSurge == nil {
		out.MaxSurge = &intstr.IntOrString{}
	}

	if err := metav1.Convert_intstr_IntOrString_To_intstr_IntOrString(&in.MaxUnavailable, out.MaxUnavailable, s); err != nil {
		return err
	}
	if err := metav1.Convert_intstr_IntOrString_To_intstr_IntOrString(&in.MaxSurge, out.MaxSurge, s); err != nil {
		return err
	}

	return nil
}

func Convert_url_Values_To_v1_DeploymentLogOptions(in *url.Values, out *v1.DeploymentLogOptions, s conversion.Scope) error {
	if in == nil || out == nil {
		return nil
	}

	*out = v1.DeploymentLogOptions{}

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

	return nil
}

// AddCustomConversionFuncs adds conversion functions which cannot be automatically generated.
// This is typically due to the objects not having 1:1 field mappings.
func AddCustomConversionFuncs(scheme *runtime.Scheme) error {
	return scheme.AddConversionFunc((*url.Values)(nil), (*v1.DeploymentLogOptions)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_url_Values_To_v1_DeploymentLogOptions(a.(*url.Values), b.(*v1.DeploymentLogOptions), scope)
	})
}
