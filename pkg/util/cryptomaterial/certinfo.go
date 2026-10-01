package cryptomaterial

import (
	"crypto/x509"
	"path/filepath"
	"time"
)

const (
	CACertFileName     = "ca.crt"
	CAKeyFileName      = "ca.key"
	CABundleFileName   = "ca-bundle.crt"
	CASerialsFileName  = "serial.txt"
	ServerCertFileName = "server.crt"
	ServerKeyFileName  = "server.key"
	ClientCertFileName = "client.crt"
	ClientKeyFileName  = "client.key"
	PeerCertFileName   = "peer.crt"
	PeerKeyFileName    = "peer.key"

	ShortLivedCertificateValidity = 365 * 24 * time.Hour
	LongLivedCertificateValidity  = 10 * ShortLivedCertificateValidity
)

func IsCertShortLived(c *x509.Certificate) bool {
	totalTime := c.NotAfter.Sub(c.NotBefore)

	// certs under 5 years are considered short-lived
	return totalTime < 5*365*time.Hour*24
}

func CertsDirectory(dataPath string) string { return filepath.Join(dataPath, "certs") }

func CACertPath(dir string) string    { return filepath.Join(dir, CACertFileName) }
func CAKeyPath(dir string) string     { return filepath.Join(dir, CAKeyFileName) }
func CASerialsPath(dir string) string { return filepath.Join(dir, CASerialsFileName) }

func CABundlePath(dir string) string { return filepath.Join(dir, CABundleFileName) }

func ClientCertPath(dir string) string { return filepath.Join(dir, ClientCertFileName) }
func ClientKeyPath(dir string) string  { return filepath.Join(dir, ClientKeyFileName) }

func ServingCertPath(dir string) string { return filepath.Join(dir, ServerCertFileName) }
func ServingKeyPath(dir string) string  { return filepath.Join(dir, ServerKeyFileName) }

func PeerCertPath(dir string) string { return filepath.Join(dir, PeerCertFileName) }
func PeerKeyPath(dir string) string  { return filepath.Join(dir, PeerKeyFileName) }

// ClientCADir returns the path to the consolidated client CA directory.
func ClientCADir(certsDir string) string { return filepath.Join(certsDir, "client-ca") }

// ServingCADir returns the path to the consolidated serving CA directory.
func ServingCADir(certsDir string) string { return filepath.Join(certsDir, "serving-ca") }

// PeerCADir returns the path to the peer (etcd) CA directory.
func PeerCADir(certsDir string) string { return filepath.Join(certsDir, "peer-ca") }

// legacyKubeControlPlaneSignerDir returns the old CA path used only for migration detection.
func legacyKubeControlPlaneSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "kube-control-plane-signer")
}

// Client CA leaf directories (all under client-ca/).
func KubeSchedulerClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kube-scheduler")
}
func KubeControllerManagerClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kube-controller-manager")
}
func KubeAPIServerToKubeletClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kube-apiserver-to-kubelet-client")
}
func MetricsServerKubeletClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "metrics-server-kubelet-client")
}
func AdminKubeconfigClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "admin-kubeconfig-client")
}
func KubeletClientCertDir(certsDir string) string {
	return filepath.Join(ClientCADir(certsDir), "kubelet-client")
}

// Serving CA leaf directories (all under serving-ca/).
func KASServingCertDir(certsDir string) string {
	return filepath.Join(ServingCADir(certsDir), "kube-apiserver-serving")
}
func KubeletServingCertDir(certsDir string) string {
	return filepath.Join(ServingCADir(certsDir), "kubelet-server")
}

// Unchanged CA directories.
func ServiceCADir(certsDir string) string { return filepath.Join(certsDir, "service-ca") }
func RouteControllerManagerServingCertDir(certsDir string) string {
	return filepath.Join(ServiceCADir(certsDir), "route-controller-manager-serving")
}
func IngressCADir(certsDir string) string { return filepath.Join(certsDir, "ingress-ca") }
func AggregatorSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "aggregator-signer")
}
func AggregatorClientCertDir(certsDir string) string {
	return filepath.Join(AggregatorSignerDir(certsDir), "aggregator-client")
}

// Peer CA leaf directories (all under peer-ca/).
func EtcdAPIServerClientCertDir(certsDir string) string {
	return filepath.Join(PeerCADir(certsDir), "apiserver-etcd-client")
}
func EtcdPeerCertDir(certsDir string) string {
	return filepath.Join(PeerCADir(certsDir), "etcd-peer")
}
func EtcdServingCertDir(certsDir string) string {
	return filepath.Join(PeerCADir(certsDir), "etcd-serving")
}

// TotalClientCABundlePath returns the path to the cert bundle with all client certificate signers
func TotalClientCABundlePath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "client-ca.crt")
}

// UltimateTrustBundlePath returns the path to the cert bundle with the root certificate
func UltimateTrustBundlePath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "ca-bundle.crt")
}

// KubeletClientCAPath returns the path to the cert bundle with all client certificate signers that kubelet should respect
func KubeletClientCAPath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "kubelet-ca.crt")
}

func KubeletServingCAPath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "kubelet-serving-ca.crt")
}

func ServiceAccountTokenCABundlePath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "service-account-token-ca.crt")
}
