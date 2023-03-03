package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"

	"github.com/spf13/cobra"
	etcd "go.etcd.io/etcd/server/v3/embed"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
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
	etcdCfg                 *etcd.Config
	minDefragBytes          int64
	maxFragmentedPercentage float64
	defragCheckFreq         time.Duration
	doStartupDefrag         bool
}

func NewEtcd(cfg *config.MicroshiftConfig) *EtcdService {
	s := &EtcdService{}
	s.configure(cfg)
	return s
}

func (s *EtcdService) Name() string { return "etcd" }

func (s *EtcdService) configure(cfg *config.MicroshiftConfig) {
	s.minDefragBytes = cfg.Etcd.MinDefragBytes
	s.maxFragmentedPercentage = cfg.Etcd.MaxFragmentedPercentage
	s.defragCheckFreq = cfg.Etcd.DefragCheckDuration
	s.doStartupDefrag = cfg.Etcd.DoStartupDefrag

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
	s.etcdCfg.QuotaBackendBytes = cfg.Etcd.QuotaBackendBytes
	url2380 := setURL([]string{"localhost"}, "2380")
	url2379 := setURL([]string{"localhost"}, "2379")
	s.etcdCfg.APUrls = url2380
	s.etcdCfg.LPUrls = url2380
	s.etcdCfg.ACUrls = url2379
	s.etcdCfg.LCUrls = url2379
	s.etcdCfg.ListenMetricsUrls = setURL([]string{"localhost"}, "2381")

	s.etcdCfg.Name = cfg.Node.HostnameOverride
	s.etcdCfg.InitialCluster = fmt.Sprintf("%s=https://%s:2380", cfg.Node.HostnameOverride, "localhost")

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
	defer func() {
		e.Server.Stop()
		<-e.Server.StopNotify()
	}()

	// If we were told to, go ahead and do a defragment now.
	if s.doStartupDefrag {
		if err := e.Server.Backend().Defrag(); err != nil {
			err = fmt.Errorf("initial defragmentation failed: %v", err)
			klog.Error(err)
			return err
		}
	}

	// Start up the defrag controller.
	defragCtx, defragShutdown := context.WithCancel(context.Background())
	go s.defragController(defragCtx, e.Server.Backend())

	// Wait to be stopped.
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)
	sig := <-sigTerm
	klog.Infof("microshift-etcd received signal %v - stopping", sig)

	// Shutdown the defrag controller.
	defragShutdown()

	return nil
}

func (s *EtcdService) defragController(ctx context.Context, be backend.Backend) {
	// Stop the controller if defrags are disabled.
	if s.defragCheckFreq == 0 {
		klog.Warning("defragmentation has been disabled")
		return
	}

	// This timer will check the fragmented conditions periodically.
	timer := time.NewTimer(s.defragCheckFreq)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case start := <-timer.C:
			if isBackendFragmented(be, s.maxFragmentedPercentage, s.minDefragBytes) {
				klog.Info("attempting to defragment backend")
				if err := be.Defrag(); err != nil {
					klog.Errorf("defragmentation failed: %v", err)
				} else {
					klog.Infof("defragmentation took %v", time.Since(start))
				}
			}
			timer.Reset(s.defragCheckFreq)
		}
	}
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

func isBackendFragmented(b backend.Backend, maxFragmentedPercentage float64, minDefragBytes int64) bool {
	fragmentedPercentage := checkFragmentationPercentage(b.Size(), b.SizeInUse())
	if fragmentedPercentage > 0.00 {
		klog.Infof("backend store fragmented: %.2f %%, dbSize: %d", fragmentedPercentage, b.Size())
	}
	return fragmentedPercentage >= maxFragmentedPercentage && b.Size() >= minDefragBytes
}

func checkFragmentationPercentage(ondisk, inuse int64) float64 {
	diff := float64(ondisk - inuse)
	fragmentedPercentage := (diff / float64(ondisk)) * 100
	return math.Round(fragmentedPercentage*100) / 100
}
