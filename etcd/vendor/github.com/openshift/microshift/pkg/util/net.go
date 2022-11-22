/*
Copyright © 2021 MicroShift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"crypto/tls"
	"fmt"
	tcpnet "net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	// Used to get gateway IP when default route doesn't exist
	OVNGatewayInterface = "br-ex"
)

func GetHostIP() (string, error) {
	ip, err := net.ChooseHostInterface()
	if err == nil {
		return ip.String(), nil
	}
	klog.V(2).Infof("failed to find default route IP address: %v", err)

	gatewayIP, err := getOVNGatewayIP()
	if err != nil {
		return "", err
	}
	klog.V(2).Infof("ovn gateway IP address: %s", gatewayIP)

	return gatewayIP, nil
}

func getOVNGatewayIP() (string, error) {
	iface, err := tcpnet.InterfaceByName(OVNGatewayInterface)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ip := addr.(*tcpnet.IPNet).IP
		if ip.To4() != nil {
			return ip.To4().String(), nil
		} else {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("failed to get ovn gateway IP address")
}

func RetryInsecureHttpsGet(url string) int {

	status := 0
	err := wait.Poll(5*time.Second, 120*time.Second, func() (bool, error) {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err := http.Get(url)
		if err == nil {
			status = resp.StatusCode
			return true, nil
		}
		return false, nil
	})

	if err != nil && err == wait.ErrWaitTimeout {
		klog.Warningf("Endpoint is not returning any status code")
	}

	return status
}

func RetryTCPConnection(host string, port string) bool {
	status := false
	err := wait.Poll(5*time.Second, 120*time.Second, func() (bool, error) {
		timeout := 30 * time.Second
		_, err := tcpnet.DialTimeout("tcp", tcpnet.JoinHostPort(host, port), timeout)

		if err == nil {
			status = true
			return true, nil
		}
		return false, nil
	})
	if err != nil && err == wait.ErrWaitTimeout {
		klog.Warningf("Endpoint is not returning any status code")
	}
	return status
}

func CreateLocalhostListenerOnPort(port int) (tcpnet.Listener, error) {
	ln, err := tcpnet.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func AddToNoProxyEnv(additionalEntries ...string) error {
	entries := map[string]struct{}{}

	// put both the NO_PROXY and no_proxy elements in a map to avoid duplicates
	addNoProxyEnvVarEntries(entries, "NO_PROXY")
	addNoProxyEnvVarEntries(entries, "no_proxy")

	for _, entry := range additionalEntries {
		entries[entry] = struct{}{}
	}

	noProxyEnv := strings.Join(mapKeys(entries), ",")

	// unset the lower-case one, and keep only upper-case
	os.Unsetenv("no_proxy")
	return errors.Wrap(os.Setenv("NO_PROXY", noProxyEnv), "error updating NO_PROXY")
}

func mapKeys(entries map[string]struct{}) []string {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}

	// sort keys to avoid issues with map key ordering in go future versions on the unit-test side
	sort.Strings(keys)
	return keys
}

func addNoProxyEnvVarEntries(entries map[string]struct{}, envVar string) {
	noProxy := os.Getenv(envVar)

	if noProxy != "" {
		for _, entry := range strings.Split(noProxy, ",") {
			entries[strings.Trim(entry, " ")] = struct{}{}
		}
	}
}
