package main

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"

	"github.com/spf13/cobra"
	etcd "go.etcd.io/etcd/server/v3/embed"
	"k8s.io/klog/v2"
)

func NewRunEtcdCommand() *cobra.Command {
	cfg := config.NewMicroshiftConfig()
	cmd := &cobra.Command{
		Use:   "run",
		Short: "TODO",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if err := cfg.ReadAndValidate(config.GetConfigFile(), cmd.Flags()); err != nil {
				klog.Fatalf("Error in reading and validating flags", err)
			}

			e := NewEtcd(cfg)
			ready := make(chan struct{})
			stopped := make(chan struct{})
			return e.Run(cmd.Context(), ready, stopped)
		},
	}
	return cmd
}

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
	certsDir := cryptomaterial.CertsDirectory(config.GetDataDir())

	etcdServingCertDir := cryptomaterial.EtcdServingCertDir(certsDir)
	etcdPeerCertDir := cryptomaterial.EtcdPeerCertDir(certsDir)
	etcdSignerCertPath := cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))
	dataDir := filepath.Join(config.GetDataDir(), s.Name())

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
	s.etcdCfg.ClientTLSInfo.CertFile = cryptomaterial.PeerCertPath(etcdServingCertDir)
	s.etcdCfg.ClientTLSInfo.KeyFile = cryptomaterial.PeerKeyPath(etcdServingCertDir)
	s.etcdCfg.ClientTLSInfo.TrustedCAFile = etcdSignerCertPath

	s.etcdCfg.PeerTLSInfo.CertFile = cryptomaterial.PeerCertPath(etcdPeerCertDir)
	s.etcdCfg.PeerTLSInfo.KeyFile = cryptomaterial.PeerKeyPath(etcdPeerCertDir)
	s.etcdCfg.PeerTLSInfo.TrustedCAFile = etcdSignerCertPath
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
