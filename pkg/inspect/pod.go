package inspect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func (o *InspectOptions) gatherPodData(destDir, namespace string, pod *corev1.Pod) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.yaml", pod.Name)
	if err := o.fileWriter.WriteFromResource(path.Join(destDir, "/"+filename), pod); err != nil {
		return err
	}

	errs := []error{}

	// gather data for each container in the given pod
	for _, container := range pod.Spec.Containers {
		if err := o.gatherContainerInfo(path.Join(destDir, "/"+container.Name), pod, container); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	for _, container := range pod.Spec.InitContainers {
		if err := o.gatherContainerInfo(path.Join(destDir, "/"+container.Name), pod, container); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("one or more errors ocurred while gathering container data for pod %s:\n\n    %v", pod.Name, utilerrors.NewAggregate(errs))
	}
	return nil
}

func (o *InspectOptions) gatherContainerInfo(destDir string, pod *corev1.Pod, container corev1.Container) error {
	if err := o.gatherContainerAllLogs(path.Join(destDir, "/"+container.Name), pod, &container); err != nil {
		return err
	}
	if len(o.restConfig.BearerToken) > 0 {
		// token authentication is vulnerable to replays if the token is sent to a potentially untrustworthy source.
		klog.V(1).Infof("        Skipping container endpoint collection for pod %q container %q: Using token authentication\n", pod.Name, container.Name)
		return nil
	}
	if len(container.Ports) == 0 {
		klog.V(1).Infof("        Skipping container endpoint collection for pod %q container %q: No ports\n", pod.Name, container.Name)
		return nil
	}
	port := &RemoteContainerPort{
		Protocol: "https",
		Port:     container.Ports[0].ContainerPort,
	}

	if err := o.gatherContainerEndpoints(path.Join(destDir, "/"+container.Name), pod, port); err != nil {
		klog.V(1).Infof("        Skipping one or more container endpoint collection for pod %q container %q: %v\n", pod.Name, container.Name, err)
	}

	return nil
}

func (o *InspectOptions) gatherContainerAllLogs(destDir string, pod *corev1.Pod, container *corev1.Container) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	errs := []error{}
	if err := o.gatherContainerLogs(path.Join(destDir, "/logs"), pod, container); err != nil {
		errs = append(errs, filterContainerLogsErrors(err))
	}

	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	return nil
}

func (o *InspectOptions) gatherContainerEndpoints(destDir string, pod *corev1.Pod, metricsPort *RemoteContainerPort) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	errs := []error{}
	if err := o.gatherContainerHealthz(path.Join(destDir, "/healthz"), pod, metricsPort); err != nil {
		errs = append(errs, fmt.Errorf("unable to gather container /healthz: %v", err))
	}
	if err := o.gatherContainerVersion(destDir, pod, metricsPort); err != nil {
		errs = append(errs, fmt.Errorf("unable to gather container /version: %v", err))
	}
	if err := o.gatherContainerMetrics(destDir, pod, metricsPort); err != nil {
		errs = append(errs, fmt.Errorf("unable to gather container /metrics: %v", err))
	}
	if err := o.gatherContainerDebug(destDir, pod, metricsPort); err != nil && !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("unable to gather container /debug : %v", err))
	}

	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	return nil
}

func filterContainerLogsErrors(err error) error {
	if strings.Contains(err.Error(), "previous terminated container") && strings.HasSuffix(err.Error(), "not found") {
		klog.V(1).Infof("        Unable to gather previous container logs: %v\n", err)
		return nil
	}
	return err
}

func (o *InspectOptions) gatherContainerVersion(destDir string, pod *corev1.Pod, metricsPort *RemoteContainerPort) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	hasVersionPath := false

	// determine if a /version endpoint exists
	paths, err := getAvailablePodEndpoints(o.podUrlGetter, pod, o.restConfig, metricsPort)
	if err != nil {
		return err
	}
	for _, p := range paths {
		if p != "/version" {
			continue
		}
		hasVersionPath = true
		break
	}
	if !hasVersionPath {
		klog.V(1).Infof("        Skipping /version info gathering for pod %q. Endpoint not found...\n", pod.Name)
		return nil
	}

	result, err := o.podUrlGetter.Get("/version", pod, o.restConfig, metricsPort)

	return o.fileWriter.WriteFromSource(path.Join(destDir, "version.json"), &TextWriterSource{Text: result})
}

// gatherContainerDebug invokes an asynchronous network call to gather pprof profile and heap
func (o *InspectOptions) gatherContainerDebug(destDir string, pod *corev1.Pod, debugPort *RemoteContainerPort) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}
	endpoints := []string{"heap", "profile", "trace"}

	for _, endpoint := range endpoints {
		// we need a token in order to access the /debug endpoint
		result, err := o.podUrlGetter.Get("/debug/pprof/"+endpoint, pod, o.restConfig, debugPort)
		if err != nil {
			return err
		}
		if err := o.fileWriter.WriteFromSource(path.Join(destDir, endpoint), &TextWriterSource{Text: result}); err != nil {
			return err
		}
	}

	return nil
}

// gatherContainerMetrics invokes an asynchronous network call
func (o *InspectOptions) gatherContainerMetrics(destDir string, pod *corev1.Pod, metricsPort *RemoteContainerPort) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	// we need a token in order to access the /metrics endpoint
	result, err := o.podUrlGetter.Get("/metrics", pod, o.restConfig, metricsPort)
	if err != nil {
		return err
	}

	return o.fileWriter.WriteFromSource(path.Join(destDir, "metrics.json"), &TextWriterSource{Text: result})
}

func (o *InspectOptions) gatherContainerHealthz(destDir string, pod *corev1.Pod, metricsPort *RemoteContainerPort) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	paths, err := getAvailablePodEndpoints(o.podUrlGetter, pod, o.restConfig, metricsPort)
	if err != nil {
		return err
	}

	healthzSeparator := "/healthz"
	healthzPaths := []string{}
	for _, p := range paths {
		if !strings.HasPrefix(p, healthzSeparator) {
			continue
		}
		healthzPaths = append(healthzPaths, p)
	}
	if len(healthzPaths) == 0 {
		return fmt.Errorf("unable to find any available /healthz paths hosted in pod %q", pod.Name)
	}

	err = o.podUrlGetter.ForwardRequests(pod, o.restConfig, metricsPort, func(restClient rest.Interface) error {
		for _, healthzPath := range healthzPaths {
			ioCloser, err := restClient.Get().RequestURI(path.Join("/", healthzPath)).Stream(context.TODO())
			if err != nil {
				return err
			}
			defer ioCloser.Close()

			data := bytes.NewBuffer(nil)
			if _, err := io.Copy(data, ioCloser); err != nil {
				return err
			}

			if len(healthzSeparator) > len(healthzPath) {
				continue
			}
			filename := healthzPath[len(healthzSeparator):]
			if len(filename) == 0 {
				filename = "index"
			} else {
				filename = strings.TrimPrefix(filename, "/")
			}

			filenameSegs := strings.Split(filename, "/")
			if len(filenameSegs) > 1 {
				// ensure directory structure for nested paths exists
				filenameSegs = filenameSegs[:len(filenameSegs)-1]
				if err := os.MkdirAll(path.Join(destDir, "/"+strings.Join(filenameSegs, "/")), os.ModePerm); err != nil {
					return err
				}
			}

			if err := o.fileWriter.WriteFromSource(path.Join(destDir, filename), &TextWriterSource{Text: data.String()}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func getAvailablePodEndpoints(urlGetter *PortForwardURLGetter, pod *corev1.Pod, config *rest.Config, port *RemoteContainerPort) ([]string, error) {
	result, err := urlGetter.Get("/", pod, config, port)
	if err != nil {
		return nil, err
	}

	resultBuffer := bytes.NewBuffer([]byte(result))
	pathInfo := map[string][]string{}

	// first, unmarshal result into json object and obtain all available /healthz endpoints
	if err := json.Unmarshal(resultBuffer.Bytes(), &pathInfo); err != nil {
		return nil, err
	}
	paths, ok := pathInfo["paths"]
	if !ok {
		return nil, fmt.Errorf("unable to extract path information for pod %q", pod.Name)
	}

	return paths, nil
}

func (o *InspectOptions) gatherContainerLogs(destDir string, pod *corev1.Pod, container *corev1.Container) error {
	// ensure destination path exists
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}
	errs := []error{}
	wg := sync.WaitGroup{}
	errLock := sync.Mutex{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		innerErrs := []error{}
		logOptions := &corev1.PodLogOptions{
			Container:  container.Name,
			Follow:     false,
			Previous:   false,
			Timestamps: true,
		}
		if len(o.sinceTime) > 0 {
			logOptions.SinceTime = &o.sinceTimestamp
		}
		if o.since != 0 {
			logOptions.SinceSeconds = &o.sinceInt
		}
		filename := "current.log"
		logsReq := o.kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
		if err := o.fileWriter.WriteFromSource(path.Join(destDir, "/"+filename), logsReq); err != nil {
			innerErrs = append(innerErrs, err)

			// if we had an error, we will try again with an insecure backendproxy flag set
			logOptions.InsecureSkipTLSVerifyBackend = true
			logsReq = o.kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
			filename = "current.insecure.log"
			if err := o.fileWriter.WriteFromSource(path.Join(destDir, "/"+filename), logsReq); err != nil {
				innerErrs = append(innerErrs, err)
			}
		}

		errLock.Lock()
		defer errLock.Unlock()
		errs = append(errs, innerErrs...)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()

		innerErrs := []error{}
		logOptions := &corev1.PodLogOptions{
			Container:  container.Name,
			Follow:     false,
			Previous:   true,
			Timestamps: true,
		}
		logsReq := o.kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
		filename := "previous.log"
		if err := o.fileWriter.WriteFromSource(path.Join(destDir, "/"+filename), logsReq); err != nil {
			innerErrs = append(innerErrs, err)

			// if we had an error, we will try again with an insecure backendproxy flag set
			logOptions.InsecureSkipTLSVerifyBackend = true
			logsReq = o.kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
			filename = "previous.insecure.log"
			if err := o.fileWriter.WriteFromSource(path.Join(destDir, "/"+filename), logsReq); err != nil {
				innerErrs = append(innerErrs, err)
			}
		}

		errLock.Lock()
		defer errLock.Unlock()
		errs = append(errs, innerErrs...)
	}()
	wg.Wait()
	return utilerrors.NewAggregate(errs)
}
