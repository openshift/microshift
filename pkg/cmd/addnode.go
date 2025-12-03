package cmd

import (
	"context"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	librarycrypto "github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/microshift/pkg/components"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	// Default timeout for operations
	joinDefaultTimeout = 10 * time.Minute
)

type AddNodeOptions struct {
	KubeconfigPath string
	Timeout        time.Duration
	Learner        bool
}

func NewAddNodeCommand() *cobra.Command {
	opts := &AddNodeOptions{
		KubeconfigPath: "",
		Timeout:        joinDefaultTimeout,
	}

	cmd := &cobra.Command{
		Use:   "add-node",
		Short: "Adds a new node to an existing MicroShift cluster",
		Long: `This command joins a node to an existing MicroShift cluster by:
1. Loading the MicroShift configuration for current node.
2. Fetch Certificate Authorities from the cluster using provided kubeconfig.
4. Configuring etcd cluster to add the new member.
5. Configuring kubelet to bootstrap into the cluster.
6. Restarting the MicroShift systemd unit.
7. Verifying the node is ready in the cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddNode(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.KubeconfigPath, "kubeconfig", opts.KubeconfigPath,
		"Path to kubeconfig file for connecting to the cluster")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", opts.Timeout,
		"Timeout for cluster join operations")
	cmd.Flags().BoolVar(&opts.Learner, "learner", true,
		"Join the cluster as a learner node")

	if version.Get().BuildVariant != version.BuildVariantCommunity {
		cmd.Hidden = true
	}

	return cmd
}

func runAddNode(ctx context.Context, opts *AddNodeOptions) error {
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	klog.Info("Starting cluster join process...")
	if opts.Learner {
		klog.Info("Will add etcd node as learner")
	}

	cfg, err := config.ActiveConfig()
	if err != nil {
		return fmt.Errorf("failed to load MicroShift configuration: %w", err)
	}
	klog.Info("MicroShift configuration loaded successfully")

	client, err := createKubernetesClient(opts.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	nodeName := cfg.CanonicalNodeName()
	if isNodeInKubernetesCluster(ctx, client, nodeName) {
		klog.Infof("Node %s is already part of the cluster. Skipping join process.", nodeName)
		return nil
	}

	if err := cleanupMicroShiftData(cfg); err != nil {
		return fmt.Errorf("failed to cleanup MicroShift data directories: %w", err)
	}
	klog.Info("MicroShift data directories cleaned up successfully")

	for _, resource := range components.CertificateAuthorityResources {
		if err := fetchCertificateAuthority(ctx, client, resource.Name, resource.Dir); err != nil {
			return fmt.Errorf("failed to fetch certificate authority %s: %w", resource.Name, err)
		}
	}
	for _, resource := range components.ServiceAccountKeyResources {
		if err := fetchServiceAccountKey(ctx, client, resource.Name, resource.Dir); err != nil {
			return fmt.Errorf("failed to fetch service account key %s: %w", resource.Name, err)
		}
	}
	klog.Info("Certificate authorities fetched and written successfully")

	if err := generateEtcdCertificates(cfg); err != nil {
		return fmt.Errorf("failed to generate etcd certificates: %w", err)
	}
	klog.Info("Etcd certificates generated successfully")

	etcdMembers, err := getEtcdClusterNodes(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get cluster information: %w", err)
	}

	if err := configureEtcdForCluster(ctx, cfg, etcdMembers, opts.Learner); err != nil {
		return fmt.Errorf("failed to configure etcd for cluster: %w", err)
	}

	if err := configureBootstrapKubeconfig(cfg, opts.KubeconfigPath); err != nil {
		return fmt.Errorf("failed to configure bootstrap kubeconfig: %w", err)
	}

	if err := restartMicroShift(); err != nil {
		return fmt.Errorf("failed to restart MicroShift service: %w", err)
	}
	klog.Info("MicroShift service restarted")

	if err := waitForNodeReady(ctx, client, cfg.CanonicalNodeName()); err != nil {
		return fmt.Errorf("node failed to become ready: %w", err)
	}

	klog.Info("Node successfully joined the cluster")
	return nil
}

func createKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	exists, err := util.PathExists(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if kubeconfig file exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("kubeconfig file %s does not exist", kubeconfigPath)
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}

func fetchCertificateAuthority(ctx context.Context, client kubernetes.Interface, name, dir string) error {
	secret, err := client.CoreV1().Secrets("kube-system").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get %s secret: %w", name, err)
	}

	caCert, exists := secret.Data["ca.crt"]
	if !exists {
		return fmt.Errorf("ca.crt not found in secret")
	}

	caKey, exists := secret.Data["ca.key"]
	if !exists {
		return fmt.Errorf("ca.key not found in secret")
	}

	serial, exists := secret.Data["serial.txt"]
	if !exists {
		return fmt.Errorf("serial.txt not found in secret")
	}

	caBundle, caBundleExists := secret.Data["ca-bundle.crt"]

	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.WriteFile(cryptomaterial.CACertPath(dir), caCert, 0600); err != nil {
		return fmt.Errorf("failed to write ca.crt: %w", err)
	}

	if err := os.WriteFile(cryptomaterial.CAKeyPath(dir), caKey, 0600); err != nil {
		return fmt.Errorf("failed to write ca.key: %w", err)
	}

	if err := os.WriteFile(cryptomaterial.CASerialsPath(dir), serial, 0600); err != nil {
		return fmt.Errorf("failed to write serial.txt: %w", err)
	}

	if caBundleExists {
		if err := os.WriteFile(cryptomaterial.CABundlePath(dir), caBundle, 0600); err != nil {
			return fmt.Errorf("failed to write ca-bundle.crt: %w", err)
		}
	}

	return nil
}

func fetchServiceAccountKey(ctx context.Context, client kubernetes.Interface, name, dir string) error {
	secret, err := client.CoreV1().Secrets("kube-system").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get %s secret: %w", name, err)
	}

	serviceAccountKey, exists := secret.Data["service-account.key"]
	if !exists {
		return fmt.Errorf("service-account.key not found in secret")
	}

	serviceAccountPubKey, exists := secret.Data["service-account.pub"]
	if !exists {
		return fmt.Errorf("service-account.pub not found in secret")
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "service-account.key"), serviceAccountKey, 0600); err != nil {
		return fmt.Errorf("failed to write service-account.key: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "service-account.pub"), serviceAccountPubKey, 0400); err != nil {
		return fmt.Errorf("failed to write service-account.pub: %w", err)
	}

	return nil
}

func generateEtcdCertificates(cfg *config.Config) error {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	etcdSignerDir := cryptomaterial.EtcdSignerDir(certsDir)

	caCertPath := cryptomaterial.CACertPath(etcdSignerDir)
	caKeyPath := cryptomaterial.CAKeyPath(etcdSignerDir)

	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	caKey, err := os.ReadFile(caKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA key: %w", err)
	}

	// Create CA config from the provided cert and key
	caTLSConfig, err := librarycrypto.GetTLSCertificateConfigFromBytes(caCert, caKey)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate config: %w", err)
	}

	// Create CA object for signing
	caConfig := &librarycrypto.CA{
		Config:          caTLSConfig,
		SerialGenerator: &librarycrypto.RandomSerialGenerator{},
	}

	// Create directories for etcd certificates
	servingCertDir := cryptomaterial.EtcdServingCertDir(certsDir)
	if err := os.MkdirAll(servingCertDir, 0750); err != nil {
		return fmt.Errorf("failed to create serving cert directory: %w", err)
	}

	peerCertDir := cryptomaterial.EtcdPeerCertDir(certsDir)
	if err := os.MkdirAll(peerCertDir, 0750); err != nil {
		return fmt.Errorf("failed to create peer cert directory: %w", err)
	}

	clientCertDir := cryptomaterial.EtcdAPIServerClientCertDir(certsDir)
	if err := os.MkdirAll(clientCertDir, 0750); err != nil {
		return fmt.Errorf("failed to create client cert directory: %w", err)
	}

	// Prepare hostnames and IPs for etcd certificates
	hostnames := []string{"localhost", cfg.Node.HostnameOverride}
	ips := []net.IP{net.ParseIP("127.0.0.1")}
	if cfg.Node.NodeIP != "" {
		if ip := net.ParseIP(cfg.Node.NodeIP); ip != nil {
			ips = append(ips, ip)
		}
	}

	// Generate serving certificate
	servingTLS, err := caConfig.MakeServerCertForDuration(
		sets.New[string](hostnames...),
		cryptomaterial.ShortLivedCertificateValidity,
		func(certTemplate *x509.Certificate) error {
			certTemplate.Subject.CommonName = "system:etcd-server:etcd-client"
			certTemplate.Subject.Organization = []string{"system:etcd-servers"}
			certTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
			certTemplate.IPAddresses = ips
			certTemplate.SerialNumber = big.NewInt(4)
			serialNumberPath := filepath.Join(servingCertDir, "serial.txt")
			if err := os.WriteFile(serialNumberPath, []byte(certTemplate.SerialNumber.String()), 0600); err != nil {
				return fmt.Errorf("failed to write serial number to disk: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to generate serving certificate: %w", err)
	}

	servingCertPath := cryptomaterial.PeerCertPath(servingCertDir)
	servingKeyPath := cryptomaterial.PeerKeyPath(servingCertDir)
	if err := servingTLS.WriteCertConfigFile(servingCertPath, servingKeyPath); err != nil {
		return fmt.Errorf("failed to write serving certificate: %w", err)
	}

	peerTLS, err := caConfig.MakeServerCertForDuration(
		sets.New[string](hostnames...),
		cryptomaterial.ShortLivedCertificateValidity,
		func(certTemplate *x509.Certificate) error {
			certTemplate.Subject.CommonName = "system:etcd-peer:etcd-client"
			certTemplate.Subject.Organization = []string{"system:etcd-peers"}
			certTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
			certTemplate.IPAddresses = ips
			certTemplate.SerialNumber = big.NewInt(4)
			serialNumberPath := filepath.Join(peerCertDir, "serial.txt")
			if err := os.WriteFile(serialNumberPath, []byte(certTemplate.SerialNumber.String()), 0600); err != nil {
				return fmt.Errorf("failed to write serial number to disk: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to generate peer certificate: %w", err)
	}

	peerCertPath := cryptomaterial.PeerCertPath(peerCertDir)
	peerKeyPath := cryptomaterial.PeerKeyPath(peerCertDir)
	if err := peerTLS.WriteCertConfigFile(peerCertPath, peerKeyPath); err != nil {
		return fmt.Errorf("failed to write peer certificate: %w", err)
	}

	// Generate client certificate
	clientUserInfo := &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}}
	clientTLS, err := caConfig.MakeClientCertificateForDuration(
		clientUserInfo,
		cryptomaterial.ShortLivedCertificateValidity,
	)
	if err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	clientCertPath := cryptomaterial.ClientCertPath(clientCertDir)
	clientKeyPath := cryptomaterial.ClientKeyPath(clientCertDir)
	if err := clientTLS.WriteCertConfigFile(clientCertPath, clientKeyPath); err != nil {
		return fmt.Errorf("failed to write client certificate: %w", err)
	}

	klog.Info("All etcd certificates generated successfully with proper signatures and SAN entries")
	return nil
}

func getEtcdClusterNodes(ctx context.Context, client kubernetes.Interface) ([]string, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var members []string
	for _, node := range nodes.Items {
		nodeIP := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeIP = addr.Address
				break
			}
		}
		if nodeIP != "" {
			//TODO net.JoinHostPort
			members = append(members, fmt.Sprintf("%s=https://%s:2380", node.Name, nodeIP))
		}
	}

	return members, nil
}

func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func configureEtcdForCluster(ctx context.Context, cfg *config.Config, clusterMembers []string, isLearner bool) error {
	dataDir := filepath.Dir(cfg.EtcdConfigPath())
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return fmt.Errorf("failed to create etcd data directory: %w", err)
	}

	nodeIP := cfg.Node.NodeIP
	if nodeIP == "" {
		nodeIP = "127.0.0.1" // fallback
	}
	currentNodeMember := fmt.Sprintf("%s=https://%s:2380", cfg.CanonicalNodeName(), nodeIP)
	cfgInitialCluster := append(clusterMembers, currentNodeMember)

	clusterConfig := fmt.Sprintf("ETCD_INITIAL_CLUSTER=%s\nETCD_INITIAL_CLUSTER_STATE=existing", strings.Join(cfgInitialCluster, ","))

	if err := os.WriteFile(cfg.EtcdConfigPath(), []byte(clusterConfig), 0600); err != nil {
		return fmt.Errorf("failed to write etcd cluster configuration: %w", err)
	}

	klog.Infof("Etcd configuration written to %s", cfg.EtcdConfigPath())

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

	memberResponse, err := client.MemberList(ctx)
	if err != nil {
		return fmt.Errorf("failed to list etcd members: %v", err)
	}

	var filteredEndpoints []string
	initialCluster := fmt.Sprintf("%s=https://%s:2380", cfg.Node.HostnameOverride, cfg.Node.NodeIP)
	for _, member := range memberResponse.Members {
		if member.Name == cfg.Node.HostnameOverride {
			klog.Infof("etcd member %s already exists", cfg.Node.HostnameOverride)
			continue
		}
		if !member.IsLearner {
			filteredEndpoints = append(filteredEndpoints, member.ClientURLs[0])
		}
		initialCluster = fmt.Sprintf("%s,%s=%s", initialCluster, member.Name, member.PeerURLs[0])
	}

	klog.Infof("initial cluster: %v. Member endpoints: %v", initialCluster, filteredEndpoints)

	client, err = clientv3.New(clientv3.Config{
		Endpoints:   filteredEndpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
		Context:     ctx,
	})
	if err != nil {
		return fmt.Errorf("failed to create etcd client with filtered endpoints: %v", err)
	}

	addFunction := client.MemberAdd
	if isLearner {
		addFunction = client.MemberAddAsLearner
	}
	response, err := addFunction(ctx, []string{fmt.Sprintf("https://%s", net.JoinHostPort(cfg.Node.NodeIP, "2380"))})
	if err != nil {
		return fmt.Errorf("failed to add etcd node: %v", err)
	}
	klog.Infof("Successfully added etcd node: %v", response)
	return nil
}

func configureBootstrapKubeconfig(cfg *config.Config, kubeconfigPath string) error {
	bootstrapKubeConfigPath := cfg.BootstrapKubeConfigPath()
	if err := os.MkdirAll(filepath.Dir(bootstrapKubeConfigPath), 0750); err != nil {
		return fmt.Errorf("failed to create kubelet directory: %w", err)
	}

	if err := copyFile(kubeconfigPath, bootstrapKubeConfigPath); err != nil {
		return fmt.Errorf("failed to copy kubeconfig for kubelet: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

func restartMicroShift() error {
	cmd := exec.Command("systemctl", "restart", "microshift")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart microshift service: %w", err)
	}
	return nil
}

func isNodeInKubernetesCluster(ctx context.Context, client kubernetes.Interface, nodeName string) bool {
	_, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func waitForNodeReady(ctx context.Context, client kubernetes.Interface, nodeName string) error {
	klog.Infof("Waiting for node %s to become ready...", nodeName)

	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for node to become ready")
		case <-ticker.C:
			node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			if err != nil {
				klog.Warningf("Failed to get node %s: %v", nodeName, err)
				continue
			}

			if isNodeReady(node) {
				klog.Infof("Node %s is ready!", nodeName)
				return nil
			}

			klog.Infof("Node %s is not ready yet, waiting...", nodeName)
		}
	}
}

func cleanupMicroShiftData(cfg *config.Config) error {
	directoriesToClean := []string{
		filepath.Dir(cfg.EtcdConfigPath()),
		cryptomaterial.CertsDirectory(config.DataDir),
		filepath.Join(config.DataDir, "resources"),
		filepath.Join(config.DataDir, "bootstrap"),
	}

	for _, dir := range directoriesToClean {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", dir, err)
		}
	}
	return nil
}
