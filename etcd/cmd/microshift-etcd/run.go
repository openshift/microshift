package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"

	"github.com/spf13/cobra"
	etcd "go.etcd.io/etcd/server/v3/embed"
	"k8s.io/klog/v2"
)

func NewRunEtcdCommand() *cobra.Command {
	cfg := config.NewMicroshiftConfig()
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if err := cfg.ReadAndValidate(config.GetConfigFile()); err != nil {
				klog.Fatalf("Error in reading and validating MicroShift config: %v", err)
			}

			e := NewEtcd(cfg)
			return e.Run()
		},
	}
	return cmd
}

var tlsCipherSuites = []string{
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
}

type EtcdService struct {
	etcdCfg *etcd.Config
}

func NewEtcd(cfg *config.MicroshiftConfig) *EtcdService {
	s := &EtcdService{}
	s.configure(cfg)
	return s
}

func (s *EtcdService) Name() string { return "etcd" }

func (s *EtcdService) configure(cfg *config.MicroshiftConfig) {
	microshiftDataDir := config.GetDataDir()
	certsDir := cryptomaterial.CertsDirectory(microshiftDataDir)

	etcdServingCertDir := cryptomaterial.EtcdServingCertDir(certsDir)
	etcdPeerCertDir := cryptomaterial.EtcdPeerCertDir(certsDir)
	etcdSignerCertPath := cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))
	dataDir := filepath.Join(microshiftDataDir, s.Name())

	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	s.etcdCfg = etcd.NewConfig()
	s.etcdCfg.ClusterState = "new"
	//s.etcdCfg.ForceNewCluster = true //TODO
	s.etcdCfg.Logger = "zap"
	s.etcdCfg.Dir = dataDir
	url2380 := setURL([]string{"localhost"}, "2380")
	url2379 := setURL([]string{"localhost"}, "2379")
	s.etcdCfg.APUrls = url2380
	s.etcdCfg.LPUrls = url2380
	s.etcdCfg.ACUrls = url2379
	s.etcdCfg.LCUrls = url2379
	s.etcdCfg.ListenMetricsUrls = setURL([]string{"localhost"}, "2381")

	s.etcdCfg.Name = cfg.NodeName
	s.etcdCfg.InitialCluster = fmt.Sprintf("%s=https://%s:2380", cfg.NodeName, "localhost")

	s.etcdCfg.CipherSuites = tlsCipherSuites
	s.etcdCfg.ClientTLSInfo.CertFile = cryptomaterial.PeerCertPath(etcdServingCertDir)
	s.etcdCfg.ClientTLSInfo.KeyFile = cryptomaterial.PeerKeyPath(etcdServingCertDir)
	s.etcdCfg.ClientTLSInfo.TrustedCAFile = etcdSignerCertPath

	s.etcdCfg.PeerTLSInfo.CertFile = cryptomaterial.PeerCertPath(etcdPeerCertDir)
	s.etcdCfg.PeerTLSInfo.KeyFile = cryptomaterial.PeerKeyPath(etcdPeerCertDir)
	s.etcdCfg.PeerTLSInfo.TrustedCAFile = etcdSignerCertPath
}

func (s *EtcdService) Run() error {
	e, err := etcd.StartEtcd(s.etcdCfg)
	if err != nil {
		return fmt.Errorf("microshift-etcd failed to start: %v", err)
	}
	<-e.Server.ReadyNotify()

	// Wait to be stopped.
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	sig := <-sigTerm
	klog.Infof("microshift-etcd received signal %v - stopping", sig)

	e.Server.Stop()
	<-e.Server.StopNotify()
	return nil
}

func setURL(hostnames []string, port string) []url.URL {
	urls := make([]url.URL, len(hostnames))
	for i, name := range hostnames {
		u, err := url.Parse("https://" + net.JoinHostPort(name, port))
		if err != nil {
			klog.Errorf("failed to parse url: %v", err)
			return []url.URL{}
		}
		urls[i] = *u
	}
	return urls
}
