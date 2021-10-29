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
package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	openshift_apiserver "github.com/openshift/openshift-apiserver/pkg/cmd/openshift-apiserver"

	"github.com/openshift/microshift/pkg/config"
)

func newOpenshiftApiServerCommand(stopCh <-chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openshift-apiserver",
		Short: "Command for the OpenShift API Server",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	start := openshift_apiserver.NewOpenShiftAPIServerCommand("start", os.Stdout, os.Stderr, stopCh)
	cmd.AddCommand(start)

	return cmd
}

func OCPAPIServer(cfg *config.MicroshiftConfig) error {
	stopCh := make(chan struct{})
	command := newOpenshiftApiServerCommand(stopCh)
	args := []string{
		"start",
		"--config=" + cfg.DataDir + "/resources/openshift-apiserver/config/config.yaml",
		"--authorization-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--authentication-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--logtostderr=" + strconv.FormatBool(cfg.LogDir == "" || cfg.LogAlsotostderr),
		"--alsologtostderr=" + strconv.FormatBool(cfg.LogAlsotostderr),
		"--v=" + strconv.Itoa(cfg.LogVLevel),
		"--vmodule=" + cfg.LogVModule,
	}
	if cfg.LogDir != "" {
		args = append(args, "--log-dir="+cfg.LogDir)
	}

	command.SetArgs(args)
	logrus.Infof("starting openshift-apiserver, args: %v", args)
	go func() {
		logrus.Fatalf("ocp apiserver exited: %v", command.Execute())
	}()

	// ocp api service registration
	if err := createAPIHeadlessSvc(cfg, "openshift-apiserver", 8444); err != nil {
		logrus.Warningf("failed to apply headless svc %v", err)
		return err
	}
	if err := createAPIHeadlessSvc(cfg, "openshift-oauth-apiserver", 8443); err != nil {
		logrus.Warningf("failed to apply headless svc %v", err)
		return err
	}
	if err := createAPIRegistration(cfg); err != nil {
		logrus.Warningf("failed to register api %v", err)
		return err
	}

	// probe ocp api services
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	err = waitForOCPAPIServer(client, 10*time.Second)
	if err != nil {
		logrus.Warningf("Failed to wait for ocp apiserver: %v", err)
		return nil
	}

	logrus.Info("ocp apiserver is ready")

	return nil
}

func waitForOCPAPIServer(client kubernetes.Interface, timeout time.Duration) error {
	var lastErr error

	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		for _, apiSvc := range []string{
			"apps.openshift.io",
			"authorization.openshift.io",
			"build.openshift.io",
			"image.openshift.io",
			"project.openshift.io",
			"quota.openshift.io",
			"route.openshift.io",
			"security.openshift.io",
			"template.openshift.io", //TODO missing templateinstances
		} {
			status := 0
			url := "/apis/" + apiSvc
			result := client.Discovery().RESTClient().Get().AbsPath(url).Do(context.TODO()).StatusCode(&status)
			if result.Error() != nil {
				lastErr = fmt.Errorf("failed to get apiserver %s status: %v", url, result.Error())
				return false, nil
			}
			if status != http.StatusOK {
				content, _ := result.Raw()
				lastErr = fmt.Errorf("APIServer isn't available: %v", string(content))
				logrus.Warningf("APIServer isn't available yet: %v. Waiting a little while.", string(content))
				return false, nil
			}
		}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf("%v: %v", err, lastErr)
	}

	return nil
}

func OpenShiftAPIServerConfig(cfg *config.MicroshiftConfig) error {
	data := []byte(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftAPIServerConfig
aggregatorConfig:
  allowedNames:
  - kube-apiserver
  - system:kube-apiserver
  - kube-apiserver-proxy
  - system:kube-apiserver-proxy
  - system:openshift-aggregator
  - system:admin
  extraHeaderPrefixes:
  - X-Remote-Extra-
  groupHeaders:
  - X-Remote-Group
  usernameHeaders:
  - X-Remote-User
kubeClientConfig:
  kubeConfig:  ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
apiServerArguments:
  minimal-shutdown-duration:
  - 30s
  anonymous-auth:
  - "false"
  authorization-kubeconfig:
  - ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
  authentication-kubeconfig:
  - ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
  audit-log-format:
  - json
  audit-log-maxbackup:
  - "10"
  audit-log-maxsize:
  - "100"
  authorization-mode:
  - Scope
  - SystemMasters
  - RBAC
  - Node
auditConfig:
  auditFilePath: "` + cfg.LogDir + `/openshift-apiserver/audit.log"
  enabled: true
  logFormat: json
  maximumFileSizeMegabytes: 100
  maximumRetainedFiles: 10
  policyFile: "` + cfg.DataDir + `/resources/openshift-apiserver/config/policy.yaml"
  policyConfiguration:
    apiVersion: audit.k8s.io/v1
    kind: Policy
    omitStages:
    - RequestReceived
    rules:
    - level: None
      resources:
      - group: ''
        resources:
        - events
    - level: None
      resources:
      - group: oauth.openshift.io
        resources:
        - oauthaccesstokens
        - oauthauthorizetokens
    - level: None
      nonResourceURLs:
      - "/api*"
      - "/version"
      - "/healthz"
      userGroups:
      - system:authenticated
      - system:unauthenticated
    - level: Metadata
      omitStages:
      - RequestReceived
imagePolicyConfig:
  internalRegistryHostname: image-registry.openshift-image-registry.svc:5000
projectConfig:
  projectRequestMessage: ''
routingConfig:
  subdomain: ` + cfg.Cluster.Domain + `
servingInfo:
  bindAddress: "0.0.0.0:8444"
  certFile: ` + cfg.DataDir + `/resources/openshift-apiserver/secrets/tls.crt
  keyFile: ` + cfg.DataDir + `/resources/openshift-apiserver/secrets/tls.key
  ca: ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
storageConfig:
  urls:
  - https://127.0.0.1:2379
  certFile: ` + cfg.DataDir + `/resources/kube-apiserver/secrets/etcd-client/tls.crt
  keyFile: ` + cfg.DataDir + `/resources/kube-apiserver/secrets/etcd-client/tls.key
  ca: ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
  `)

	path := filepath.Join(cfg.DataDir, "resources", "openshift-apiserver", "config", "config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}
