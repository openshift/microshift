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
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/net"
)

func GetHostIP() (string, error) {
	ip, err := net.ChooseHostInterface()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

func RetryHttpGet(url string) int {
	var timers = []time.Duration{
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
		20 * time.Second,
	}

	status := 0
	for _, timer := range timers {
		resp, err := http.Get(url)
		if err == nil {
			status = resp.StatusCode
			break
		}

		logrus.Infof("Request error: %+v\n", err)
		logrus.Infof("Retrying in %v\n", timer)
		time.Sleep(timer)
	}

	return status
}
