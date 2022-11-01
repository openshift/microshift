package inspect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type defaultPortForwarder struct {
	restConfig *rest.Config

	StopChannel  chan struct{}
	ReadyChannel chan struct{}
}

func NewDefaultPortForwarder(adminConfig *rest.Config) *defaultPortForwarder {
	return &defaultPortForwarder{
		restConfig:   adminConfig,
		StopChannel:  make(chan struct{}, 1),
		ReadyChannel: make(chan struct{}, 1),
	}
}

func (f *defaultPortForwarder) ForwardPortsAndExecute(pod *corev1.Pod, ports []string, toExecute func()) error {
	if len(ports) < 1 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}

	restClient, err := rest.RESTClientFor(setRESTConfigDefaults(*f.restConfig))
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	stdout := bytes.NewBuffer(nil)
	req := restClient.Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(dialer, ports, f.StopChannel, f.ReadyChannel, stdout, ioutil.Discard)
	if err != nil {
		return err
	}

	go func() {
		if f.StopChannel != nil {
			defer close(f.StopChannel)
		}

		<-f.ReadyChannel
		toExecute()
	}()

	return fw.ForwardPorts()
}

func setRESTConfigDefaults(config rest.Config) *rest.Config {
	if config.GroupVersion == nil {
		config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	}
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = scheme.Codecs
	}
	if len(config.UserAgent) == 0 {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	config.APIPath = "/api"
	return &config
}

func newInsecureRESTClientForHost(host string) (rest.Interface, error) {
	insecure := true

	configFlags := &genericclioptions.ConfigFlags{}
	configFlags.Insecure = &insecure
	configFlags.APIServer = &host

	newConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return rest.RESTClientFor(setRESTConfigDefaults(*newConfig))
}

type RemoteContainerPort struct {
	Port     int32
	Protocol string
}

type PortForwardURLGetter struct {
	Protocol  string
	Host      string
	LocalPort string
}

func (c *PortForwardURLGetter) Get(urlPath string, pod *corev1.Pod, config *rest.Config, containerPort *RemoteContainerPort) (string, error) {
	var result string

	return result, c.ForwardRequests(pod, config, containerPort, func(restClient rest.Interface) error {
		ioCloser, err := restClient.Get().RequestURI(urlPath).Stream(context.TODO())
		if err != nil {
			return err
		}
		defer ioCloser.Close()

		data := bytes.NewBuffer(nil)
		if _, err := io.Copy(data, ioCloser); err != nil {
			return err
		}
		result = data.String()
		return nil
	})
}

func (c *PortForwardURLGetter) ForwardRequests(pod *corev1.Pod, config *rest.Config, containerPort *RemoteContainerPort, fn func(rest.Interface) error) error {
	var lastErr error
	forwarder := NewDefaultPortForwarder(config)

	if err := forwarder.ForwardPortsAndExecute(pod, []string{fmt.Sprintf("%v:%v", c.LocalPort, containerPort.Port)}, func() {
		url := fmt.Sprintf("%s://%s:%s", containerPort.Protocol, c.Host, c.LocalPort)
		restClient, err := newInsecureRESTClientForHost(url)
		if err != nil {
			lastErr = err
			return
		}

		if err := fn(restClient); err != nil {
			lastErr = err
			return
		}
	}); err != nil {
		return err
	}
	return lastErr
}
