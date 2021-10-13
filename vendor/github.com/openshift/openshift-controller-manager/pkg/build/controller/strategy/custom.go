package strategy

import (
	"errors"
	"fmt"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	buildv1 "github.com/openshift/api/build/v1"
	buildutil "github.com/openshift/openshift-controller-manager/pkg/build/buildutil"
)

var (
	customBuildEncodingScheme       = runtime.NewScheme()
	customBuildEncodingCodecFactory = serializer.NewCodecFactory(customBuildEncodingScheme)
)

func init() {
	utilruntime.Must(buildv1.Install(customBuildEncodingScheme))
	utilruntime.Must(buildv1.DeprecatedInstallWithoutGroup(customBuildEncodingScheme))
	customBuildEncodingCodecFactory = serializer.NewCodecFactory(customBuildEncodingScheme)
}

// CustomBuildStrategy creates a build using a custom builder image.
type CustomBuildStrategy struct {
}

// CreateBuildPod creates the pod to be used for the Custom build
func (bs *CustomBuildStrategy) CreateBuildPod(build *buildv1.Build, additionalCAs map[string]string, internalRegistryHost string) (*corev1.Pod, error) {
	strategy := build.Spec.Strategy.CustomStrategy
	if strategy == nil {
		return nil, errors.New("CustomBuildStrategy cannot be executed without CustomStrategy parameters")
	}

	codec := customBuildEncodingCodecFactory.LegacyCodec(buildv1.GroupVersion)
	if len(strategy.BuildAPIVersion) != 0 {
		gv, err := schema.ParseGroupVersion(strategy.BuildAPIVersion)
		if err != nil {
			return nil, &FatalError{fmt.Sprintf("failed to parse buildAPIVersion specified in custom build strategy (%q): %v", strategy.BuildAPIVersion, err)}
		}
		codec = customBuildEncodingCodecFactory.LegacyCodec(gv)
	}

	data, err := runtime.Encode(codec, build)
	if err != nil {
		return nil, fmt.Errorf("failed to encode the build: %v", err)
	}

	containerEnv := []corev1.EnvVar{
		{Name: "BUILD", Value: string(data)},
		{Name: "LANG", Value: "C.utf8"},
	}

	if build.Spec.Source.Git != nil {
		addSourceEnvVars(build.Spec.Source, &containerEnv)
	}

	if build.Spec.Output.To != nil {
		addOutputEnvVars(build.Spec.Output.To, &containerEnv)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the output docker tag %q: %v", build.Spec.Output.To.Name, err)
		}
	}

	if len(strategy.From.Name) == 0 {
		return nil, errors.New("CustomBuildStrategy cannot be executed without image")
	}

	if len(strategy.Env) > 0 {
		containerEnv = append(containerEnv, strategy.Env...)
	}

	if strategy.ExposeDockerSocket {
		klog.V(2).Infof("ExposeDockerSocket is enabled for %s build", build.Name)
		containerEnv = append(containerEnv, corev1.EnvVar{Name: "DOCKER_SOCKET", Value: dockerSocketPath})
	}

	serviceAccount := build.Spec.ServiceAccount
	if len(serviceAccount) == 0 {
		serviceAccount = buildutil.BuilderServiceAccountName
	}

	privileged := true
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildutil.GetBuildPodName(build),
			Namespace: build.Namespace,
			Labels:    getPodLabels(build),
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount,
			Containers: []corev1.Container{
				{
					Name:  CustomBuild,
					Image: strategy.From.Name,
					Env:   containerEnv,
					// TODO: run unprivileged https://github.com/openshift/origin/issues/662
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			NodeSelector:  build.Spec.NodeSelector,
		},
	}

	pod = setupActiveDeadline(pod, build)

	if !strategy.ForcePull {
		pod.Spec.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
	} else {
		klog.V(2).Infof("ForcePull is enabled for %s build", build.Name)
		pod.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	}
	pod.Spec.Containers[0].Resources = build.Spec.Resources
	if build.Spec.Source.Binary != nil {
		pod.Spec.Containers[0].Stdin = true
		pod.Spec.Containers[0].StdinOnce = true
	}

	if strategy.ExposeDockerSocket {
		setupDockerSocket(pod)
	}
	setupDockerSecrets(pod, &pod.Spec.Containers[0], build.Spec.Output.PushSecret, strategy.PullSecret, build.Spec.Source.Images)
	setOwnerReference(pod, build)
	setupSourceSecrets(pod, &pod.Spec.Containers[0], build.Spec.Source.SourceSecret)
	setupInputSecrets(pod, &pod.Spec.Containers[0], build.Spec.Source.Secrets)
	setupAdditionalSecrets(pod, &pod.Spec.Containers[0], build.Spec.Strategy.CustomStrategy.Secrets)
	setupContainersConfigs(build, pod)
	setupBuildCAs(build, pod, additionalCAs, internalRegistryHost)
	setupContainersStorage(pod, &pod.Spec.Containers[0]) // for unprivileged builds
	return pod, nil
}
