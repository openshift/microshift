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
	"context"
	"crypto/tls"
	"fmt"
	tcpnet "net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var previousGatewayIP string = ""

func GetHostIP() (string, error) {
	ip, err := net.ChooseHostInterface()
	if err == nil {
		return ip.String(), nil
	}
	hostIP := ip.String()
	if hostIP != previousGatewayIP {
		previousGatewayIP = hostIP
		klog.V(2).Infof("host gateway IP address: %s", hostIP)
	}

	return hostIP, nil
}

func RetryInsecureGet(ctx context.Context, url string) int {
	status := 0
	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
		c := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec
				},
			},
		}
		resp, err := c.Get(url)
		if err == nil {
			status = resp.StatusCode
			return true, nil
		}
		defer resp.Body.Close() //nolint:errcheck
		return false, nil
	})

	if err != nil && err == context.DeadlineExceeded {
		klog.Warningf("Endpoint is not returning any status code")
	}

	return status
}

func RetryTCPConnection(ctx context.Context, host string, port string) bool {
	status := false
	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 120*time.Second, false, func(ctx context.Context) (bool, error) {
		timeout := 30 * time.Second
		_, err := tcpnet.DialTimeout("tcp", tcpnet.JoinHostPort(host, port), timeout)

		if err == nil {
			status = true
			return true, nil
		}
		return false, nil
	})
	if err != nil && err == context.DeadlineExceeded {
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
	if err := os.Setenv("NO_PROXY", noProxyEnv); err != nil {
		return fmt.Errorf("failed to update NO_PROXY: %w", err)
	}
	return nil
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
