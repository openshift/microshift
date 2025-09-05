package healthcheck

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/config"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/utils/ptr"
)

func printPostFailureDebugInfo(ctx context.Context, coreClient *coreclientv1.CoreV1Client) {
	output := strings.Builder{}

	unpulledOrFailedImages(ctx, coreClient, &output)
	allPodsAndEvents(&output)

	klog.Infof("DEBUG INFORMATION\n%s", output.String())
}

func allPodsAndEvents(output *strings.Builder) {
	cliOptions := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cliOptions.KubeConfig = ptr.To(filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig"))
	if homedir.HomeDir() == "" {
		// By default client writes cache to $HOME/.kube/cache.
		// However, when healthcheck is executed by greenboot, the $HOME is empty,
		// so discovery client wants to write to /.kube which is immutable on ostre
		// causing flood of warnings (and is not elegant to create new root level directory).
		cliOptions.CacheDir = ptr.To(filepath.Join("tmp", ".kube", "cache"))
	}
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(cliOptions)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: output, ErrOut: output}

	cmdGet := get.NewCmdGet("", f, ioStreams)
	opts := get.NewGetOptions("", ioStreams)
	opts.AllNamespaces = true
	opts.PrintFlags.OutputFormat = ptr.To("wide")
	if err := opts.Complete(f, cmdGet, []string{"DUMMY"}); err != nil {
		klog.Errorf("Failed to complete get's options: %v", err)
		return
	}

	if err := opts.Validate(); err != nil {
		klog.Errorf("Failed to validate get's options: %v", err)
		return
	}

	output.WriteString("---------- PODS:\n")
	if err := opts.Run(f, []string{"pods"}); err != nil {
		klog.Errorf("Failed to run 'get pods': %v", err)
		return
	}
	output.WriteString("\n")
	output.WriteString("\n---------- EVENTS:\n")
	opts.SortBy = ".metadata.creationTimestamp"
	if err := opts.Run(f, []string{"events"}); err != nil {
		klog.Errorf("Failed to run 'get events': %v", err)
		return
	}
	output.WriteString("\n")
}

// unpulledOrFailedImages prepares a debug log with information about images that are still being pulled or failed to be pulled.
func unpulledOrFailedImages(ctx context.Context, coreClient *coreclientv1.CoreV1Client, output *strings.Builder) {
	// Get list of existing Pods to skip Events belonging to non-existing Pods to avoid false positives:
	// If someone creates and deletes a lot of workloads, there might be "Pulling" events for each Pod without
	// the corresponding "Pulled" event.
	pods, err := coreClient.Pods("").List(ctx, v1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to retrieve pods: %v", err)
		return
	}
	existingPodsNames := sets.New[string]()
	for _, pod := range pods.Items {
		existingPodsNames.Insert(pod.Name)
	}

	var pullingEvents, pulledEvents, failedEvents *corev1.EventList
	if pullingEvents, err = coreClient.Events("").List(ctx, v1.ListOptions{FieldSelector: "reportingComponent=kubelet,reason=Pulling"}); err != nil {
		klog.Errorf("Failed to retrieve Pulling events: %v", err)
		return
	}
	if pulledEvents, err = coreClient.Events("").List(ctx, v1.ListOptions{FieldSelector: "reportingComponent=kubelet,reason=Pulled"}); err != nil {
		klog.Errorf("Failed to retrieve Pulled events: %v", err)
		return
	}
	if failedEvents, err = coreClient.Events("").List(ctx, v1.ListOptions{FieldSelector: "reportingComponent=kubelet,reason=Failed"}); err != nil {
		klog.Errorf("Failed to retrieve Failed events: %v", err)
		return
	}

	unpulledImages, failedImages := analyzeEventsLookingForUnpulledOrFailedImages(existingPodsNames, pullingEvents, pulledEvents, failedEvents)

	if len(unpulledImages) > 0 {
		output.WriteString("---------- IMAGES THAT ARE STILL BEING PULLED:\n")
		for _, unpulledImage := range unpulledImages {
			output.WriteString(fmt.Sprintf("- %q for Pod %q in namespace %q\n", unpulledImage.Image, unpulledImage.PodName, unpulledImage.Namespace))
		}
		output.WriteString("\n")
	}

	if len(failedImages) > 0 {
		output.WriteString("---------- IMAGES THAT FAILED TO BE PULLED:\n")
		for _, failedImage := range failedImages {
			output.WriteString(fmt.Sprintf("- %q for Pod %q in namespace %q: %s\n", failedImage.Image, failedImage.PodName, failedImage.Namespace, failedImage.Message))
		}
		output.WriteString("\n")
	}
}

type unpulledImage struct {
	Namespace string
	PodName   string
	Image     string
}

type failedImage struct {
	unpulledImage
	Message string
}

// analyzeEventsLookingForUnpulledOrFailedImages goes through and tries to match
// image related Events to find images that are still being pulled
// and images that failed to be pulled.
func analyzeEventsLookingForUnpulledOrFailedImages(existingPodsNames sets.Set[string], pullingEvents, pulledEvents, failedEvents *corev1.EventList) ([]unpulledImage, []failedImage) {
	getImageInfo := func(event corev1.Event) (string, string, string) {
		pod := event.InvolvedObject.Name
		ns := event.InvolvedObject.Namespace
		img := strings.Split(event.Message, "\"")[1]
		return ns, pod, img
	}

	unpulledImages := sets.New[unpulledImage]()

	for _, event := range pullingEvents.Items {
		ns, pod, img := getImageInfo(event)
		if !existingPodsNames.Has(pod) {
			continue
		}
		unpulledImages.Insert(unpulledImage{Namespace: ns, PodName: pod, Image: img})
	}

	for _, event := range pulledEvents.Items {
		ns, pod, img := getImageInfo(event)
		unpulledImages.Delete(unpulledImage{Namespace: ns, PodName: pod, Image: img})
	}

	failedImages := sets.New[failedImage]()

	for _, event := range failedEvents.Items {
		if !strings.HasPrefix(event.Message, "Failed to pull image") {
			continue
		}
		ns, pod, img := getImageInfo(event)
		if !existingPodsNames.Has(pod) {
			continue
		}
		unpulledImages.Delete(unpulledImage{Namespace: ns, PodName: pod, Image: img})

		failedImages.Insert(failedImage{
			unpulledImage: unpulledImage{Namespace: ns, PodName: pod, Image: img},
			Message:       event.Message,
		})
	}

	return unpulledImages.UnsortedList(), failedImages.UnsortedList()
}
