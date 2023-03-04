package config

import (
	"net/url"
	"strconv"
)

type ApiServer struct {
	// SubjectAltNames added to API server certs
	SubjectAltNames []string `json:"subjectAltNames"`
	// Kube apiserver advertise address to work around the certificates issue
	// when requiring external access using the node IP. This will turn into
	// the IP configured in the endpoint slice for kubernetes service. Must be
	// a reachable IP from pods. Defaults to service network CIDR first
	// address.
	AdvertiseAddress string `json:"advertiseAddress,omitempty"`
	// Determines if kube-apiserver controller should configure the
	// AdvertiseAddress in the loopback interface. Automatically computed.
	SkipInterface bool `json:"-"`

	// The URL of the API server
	URL string `json:"-"`
}

// extract the api server port from the cluster URL
func (c *Config) ApiServerPort() (int, error) {
	var port string

	parsed, err := url.Parse(c.ApiServer.URL)
	if err != nil {
		return 0, err
	}

	// default empty URL to port 6443
	port = parsed.Port()
	if port == "" {
		port = "6443"
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, err
	}
	return portNum, nil
}
