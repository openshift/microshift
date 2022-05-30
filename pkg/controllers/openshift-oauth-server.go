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
	"io"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	oauth_apiserver "github.com/openshift/oauth-apiserver/pkg/cmd/oauth-apiserver"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
)

// OAuth OpenShift Service
type OpenShiftOAuth struct {
	options *oauth_apiserver.OAuthAPIServerOptions
	Output  io.Writer
}

func NewOpenShiftOAuth(cfg *config.MicroshiftConfig) *OpenShiftOAuth {
	s := &OpenShiftOAuth{}
	s.configure(cfg)
	return s
}

func (s *OpenShiftOAuth) Name() string { return "oauth-apiserver" }
func (s *OpenShiftOAuth) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-controller-manager", "openshift-apiserver"}
}

func (s *OpenShiftOAuth) configure(cfg *config.MicroshiftConfig) {
	args := []string{
		"start",
		"--secure-port=8443",
		"--kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--authorization-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--authentication-kubeconfig=" + cfg.DataDir + "/resources/kubeadmin/kubeconfig",
		"--audit-log-format=json",
		"--audit-log-maxsize=100",
		"--audit-log-maxbackup=10",
		"--etcd-cafile=" + cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt",
		"--etcd-keyfile=" + cfg.DataDir + "/resources/kube-apiserver/secrets/etcd-client/tls.key",
		"--etcd-certfile=" + cfg.DataDir + "/resources/kube-apiserver/secrets/etcd-client/tls.crt",
		"--shutdown-delay-duration=120s",
		"--tls-cert-file=" + cfg.DataDir + "/resources/openshift-oauth-apiserver/secrets/tls.crt",
		"--tls-private-key-file=" + cfg.DataDir + "/resources/openshift-oauth-apiserver/secrets/tls.key",
		"--cors-allowed-origins='//127.0.0.1(:|$)'",
		"--cors-allowed-origins='//localhost(:|$)'",
		"--etcd-servers=https://127.0.0.1:2379",
		"--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"--tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
		"--tls-cipher-suites=TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		"--tls-min-version=VersionTLS12",
	}

	fs := pflag.NewFlagSet("oauth-apiserver", pflag.PanicOnError)
	opts := oauth_apiserver.NewOAuthAPIServerOptions(os.Stdout)
	opts.AddFlags(fs)

	ls, err := util.CreateLocalhostListenerOnPort(8443)
	if err != nil {
		klog.Errorf("Failed to create listener %v", err)
	}

	opts.RecommendedOptions.SecureServing.Listener = ls
	opts.RecommendedOptions.SecureServing.BindPort = 8443

	if err := fs.Parse(args); err != nil {
		klog.Errorf("failed to parse flags %v", err)
	}
	if err := opts.Complete(); err != nil {
		klog.Errorf("failed to complete options %v", err)
	}
	if err := opts.Validate(args); err != nil {
		klog.Errorf("failed to validate options %v", err)
	}

	s.options = opts
}

func (s *OpenShiftOAuth) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// run readiness check
	go func() {
		healthcheckStatus := util.RetryTCPConnection("127.0.0.1", "8443")
		if !healthcheckStatus {
			klog.Fatalf("%s failed to start", s.Name(), fmt.Errorf("healthcheck failed"))
		}
		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	if err := oauth_apiserver.RunOAuthAPIServer(s.options, ctx.Done()); err != nil {
		klog.Fatalf("Error starting oauth API server: %s", err)
	}

	return ctx.Err()
}
