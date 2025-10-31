package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/spf13/cobra"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"

	"k8s.io/klog/v2"
)

func NewPromoteLearnerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promote-learner",
		Short: "Promote learner node to member in the etcd cluster",
		Long: `This command promotes the local etcd instance from a learner node to a full voting member within an existing MicroShift etcd cluster.
It:
  - Connects to the etcd cluster using the current node's configuration.
  - Verifies that the local etcd instance is currently a learner.
  - Issues a promote request to the cluster.
  - Restarts the MicroShift service to make the membership change effective to apiserver.
Use this command only after the learner node has fully caught up with the cluster and you are ready for it to become a voting member.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPromoteLearner(cmd.Context())
		},
	}
	return cmd
}

func runPromoteLearner(ctx context.Context) error {
	klog.Info("Starting learner promotion process")

	cfg, err := config.ActiveConfig()
	if err != nil {
		return fmt.Errorf("failed to load MicroShift configuration: %w", err)
	}

	klog.Infof("Creating Kubernetes client from %s", cfg.KubeConfigPath(config.KubeAdmin))
	client, err := createKubernetesClient(cfg.KubeConfigPath(config.KubeAdmin))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	etcdMembers, err := getEtcdClusterNodes(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get cluster information: %w", err)
	}

	klog.Info("Promoting etcd learner to member")
	if err := promoteEtcdLearner(ctx, etcdMembers, cfg); err != nil {
		return fmt.Errorf("failed to promote etcd learner: %w", err)
	}

	klog.Info("Restarting MicroShift service")
	if err := restartMicroShift(); err != nil {
		return fmt.Errorf("failed to restart MicroShift service: %w", err)
	}
	klog.Info("Learner promotion process completed successfully")
	return nil
}

func promoteEtcdLearner(ctx context.Context, clusterMembers []string, cfg *config.Config) error {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	etcdPeerClientCertDir := cryptomaterial.EtcdPeerCertDir(certsDir)

	tlsInfo := transport.TLSInfo{
		CertFile:      cryptomaterial.PeerCertPath(etcdPeerClientCertDir),
		KeyFile:       cryptomaterial.PeerKeyPath(etcdPeerClientCertDir),
		TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir)),
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to create etcd client TLS config: %v", err)
	}
	var endpoints []string
	for _, member := range clusterMembers {
		parts := strings.SplitN(member, "=", 2)
		if len(parts) == 2 {
			endpoint := strings.Replace(parts[1], ":2380", ":2379", 1)
			endpoints = append(endpoints, endpoint)
		}
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
		Context:     ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %v", err)
	}
	defer client.Close()

	memberResponse, err := client.MemberList(ctx)
	if err != nil {
		return fmt.Errorf("failed to list etcd members: %v", err)
	}

	found, learner := false, false
	var id uint64 = 0

	for _, member := range memberResponse.Members {
		if member.Name == cfg.CanonicalNodeName() {
			found = true
			if member.IsLearner {
				learner = true
				id = member.ID
			}
		}
	}

	if !found {
		return fmt.Errorf("node %s is not in the etcd cluster", cfg.CanonicalNodeName())
	}
	if !learner {
		return fmt.Errorf("node %s is not a learner", cfg.CanonicalNodeName())
	}

	response, err := client.MemberPromote(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to promote etcd learner: %v", err)
	}
	klog.Infof("Successfully promoted etcd learner %s with response: %v", cfg.CanonicalNodeName(), response)

	return nil
}
