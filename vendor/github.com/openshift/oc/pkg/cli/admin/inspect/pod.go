package inspect

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/html"

	corev1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
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
		return fmt.Errorf("one or more errors occurred while gathering container data for pod %s:\n\n    %v", pod.Name, utilerrors.NewAggregate(errs))
	}
	return nil
}

func (o *InspectOptions) gatherContainerInfo(destDir string, pod *corev1.Pod, container corev1.Container) error {
	if err := o.gatherContainerAllLogs(path.Join(destDir, "/"+container.Name), pod, &container); err != nil {
		return err
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

	if o.rotatedPodLogs {
		if err := o.gatherContainerRotatedLogFiles(path.Join(destDir, "/logs/rotated"), pod, container); err != nil {
			errs = append(errs, filterContainerLogsErrors(err))
		}
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

func rotatedLogFilename(pod *corev1.Pod) (string, error) {
	if value, exists := pod.Annotations["kubernetes.io/config.source"]; exists && value == "file" {
		hash, exists := pod.Annotations["kubernetes.io/config.hash"]
		if !exists {
			return "", fmt.Errorf("missing 'kubernetes.io/config.hash' annotation for static pod")
		}
		return pod.Namespace + "_" + pod.Name + "_" + hash, nil
	}
	return pod.Namespace + "_" + pod.Name + "_" + string(pod.GetUID()), nil
}

func (o *InspectOptions) gatherContainerRotatedLogFiles(destDir string, pod *corev1.Pod, container *corev1.Container) error {
	restClient := o.kubeClient.CoreV1().RESTClient()
	var innerErrs []error

	logFileName, err := rotatedLogFilename(pod)
	if err != nil {
		return err
	}

	// Get all container log files from the node
	containerPath := restClient.Get().
		Name(pod.Spec.NodeName).
		Resource("nodes").
		SubResource("proxy", "logs", "pods", logFileName).
		Suffix(container.Name).URL().Path

	req := restClient.Get().RequestURI(containerPath).
		SetHeader("Accept", "text/plain, */*")
	res, err := req.Stream(context.TODO())
	if err != nil {
		return err
	}

	doc, err := html.Parse(res)
	if err != nil {
		return err
	}

	// rotated log files have a suffix added at the end of the file name
	// e.g: 0.log.20211027-082023, 0.log.20211027-082023.gz
	reRotatedLog := regexp.MustCompile(`[0-9]+\.log\..+`)
	var downloadRotatedLogs func(*html.Node)
	downloadRotatedLogs = func(n *html.Node) {
		var fileName string
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					fileName = attr.Val
				}
			}
			if !reRotatedLog.MatchString(fileName) {
				return
			}

			// ensure destination dir exists
			if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
				innerErrs = append(innerErrs, err)
			}

			logsReq := restClient.Get().RequestURI(path.Join(containerPath, fileName)).
				SetHeader("Accept", "text/plain, */*").
				SetHeader("Accept-Encoding", "gzip")

			if err := o.fileWriter.WriteFromSource(path.Join(destDir, fileName), logsReq); err != nil {
				innerErrs = append(innerErrs, err)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			downloadRotatedLogs(c)
		}
	}
	downloadRotatedLogs(doc)
	return utilerrors.NewAggregate(innerErrs)
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
