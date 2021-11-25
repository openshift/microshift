/*
Copyright © 2021 Microshift Contributors

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
	tcpnet "net"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
)

func GetHostIP() (string, error) {
	ip, err := net.ChooseHostInterface()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
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
		logrus.Warningf("Endpoint is not returning any status code")
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
		logrus.Warningf("Endpoint is not returning any status code")
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
