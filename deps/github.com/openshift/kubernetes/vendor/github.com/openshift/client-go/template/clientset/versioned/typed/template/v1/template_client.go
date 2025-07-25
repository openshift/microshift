// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	http "net/http"

	templatev1 "github.com/openshift/api/template/v1"
	scheme "github.com/openshift/client-go/template/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type TemplateV1Interface interface {
	RESTClient() rest.Interface
	BrokerTemplateInstancesGetter
	TemplatesGetter
	TemplateInstancesGetter
}

// TemplateV1Client is used to interact with features provided by the template.openshift.io group.
type TemplateV1Client struct {
	restClient rest.Interface
}

func (c *TemplateV1Client) BrokerTemplateInstances() BrokerTemplateInstanceInterface {
	return newBrokerTemplateInstances(c)
}

func (c *TemplateV1Client) Templates(namespace string) TemplateInterface {
	return newTemplates(c, namespace)
}

func (c *TemplateV1Client) TemplateInstances(namespace string) TemplateInstanceInterface {
	return newTemplateInstances(c, namespace)
}

// NewForConfig creates a new TemplateV1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*TemplateV1Client, error) {
	config := *c
	setConfigDefaults(&config)
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new TemplateV1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*TemplateV1Client, error) {
	config := *c
	setConfigDefaults(&config)
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &TemplateV1Client{client}, nil
}

// NewForConfigOrDie creates a new TemplateV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *TemplateV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new TemplateV1Client for the given RESTClient.
func New(c rest.Interface) *TemplateV1Client {
	return &TemplateV1Client{c}
}

func setConfigDefaults(config *rest.Config) {
	gv := templatev1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = rest.CodecFactoryForGeneratedClient(scheme.Scheme, scheme.Codecs).WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *TemplateV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
