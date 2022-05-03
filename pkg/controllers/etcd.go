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
	"net/url"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	etcd "go.etcd.io/etcd/server/v3/embed"
	"k8s.io/klog/v2"
)

var (
	tlsCipherSuites = []string{
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
	}
)

const (
	etcdStartupTimeout = 60
)

type EtcdService struct {
	etcdCfg *etcd.Config
}

func NewEtcd(cfg *config.MicroshiftConfig) *EtcdService {
	s := &EtcdService{}
	s.configure(cfg)
	return s
}

func (s *EtcdService) Name() string           { return "etcd" }
func (s *EtcdService) Dependencies() []string { return []string{} }

func (s *EtcdService) configure(cfg *config.MicroshiftConfig) {
	caCertFile := filepath.Join(cfg.DataDir, "certs", "ca-bundle", "ca-bundle.crt")
	certDir := filepath.Join(cfg.DataDir, "certs", s.Name())
	dataDir := filepath.Join(cfg.DataDir, s.Name())

	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	s.etcdCfg = etcd.NewConfig()
	s.etcdCfg.ClusterState = "new"
	//s.etcdCfg.ForceNewCluster = true //TODO
	s.etcdCfg.Logger = "zap"
	s.etcdCfg.Dir = dataDir
	s.etcdCfg.APUrls = setURL([]string{cfg.NodeIP}, ":2380")
	s.etcdCfg.LPUrls = setURL([]string{cfg.NodeIP}, ":2380")
	s.etcdCfg.ACUrls = setURL([]string{cfg.NodeIP}, ":2379")
	s.etcdCfg.LCUrls = setURL([]string{"127.0.0.1", cfg.NodeIP}, ":2379")
	s.etcdCfg.ListenMetricsUrls = setURL([]string{"127.0.0.1"}, ":2381")

	s.etcdCfg.Name = cfg.NodeName
	s.etcdCfg.InitialCluster = fmt.Sprintf("%s=https://%s:2380", cfg.NodeName, cfg.NodeIP)

	s.etcdCfg.CipherSuites = tlsCipherSuites
	s.etcdCfg.ClientTLSInfo.CertFile = filepath.Join(certDir, "etcd-serving.crt")
	s.etcdCfg.ClientTLSInfo.KeyFile = filepath.Join(certDir, "etcd-serving.key")
	s.etcdCfg.ClientTLSInfo.TrustedCAFile = caCertFile
	s.etcdCfg.ClientTLSInfo.ClientCertAuth = false
	s.etcdCfg.ClientTLSInfo.InsecureSkipVerify = true //TODO after fix GenCert to generate client cert

	s.etcdCfg.PeerTLSInfo.CertFile = filepath.Join(certDir, "etcd-peer.crt")
	s.etcdCfg.PeerTLSInfo.KeyFile = filepath.Join(certDir, "etcd-peer.key")
	s.etcdCfg.PeerTLSInfo.TrustedCAFile = caCertFile
	s.etcdCfg.PeerTLSInfo.ClientCertAuth = false
	s.etcdCfg.PeerTLSInfo.InsecureSkipVerify = true //TODO after fix GenCert to generate client cert
}

func (s *EtcdService) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	e, err := etcd.StartEtcd(s.etcdCfg)
	if err != nil {
		return fmt.Errorf("%s failed to start: %v", s.Name(), err)
	}

	// run readiness check
	go func() {
		<-e.Server.ReadyNotify()
		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	<-ctx.Done()
	e.Server.Stop()
	<-e.Server.StopNotify()
	return ctx.Err()
}

func setURL(hostnames []string, port string) []url.URL {
	urls := make([]url.URL, len(hostnames))
	for i, name := range hostnames {
		u, err := url.Parse("https://" + name + port)
		if err != nil {
			return []url.URL{}
		}
		urls[i] = *u
	}
	return urls
}
