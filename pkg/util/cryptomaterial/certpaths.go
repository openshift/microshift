package cryptomaterial

import (
	"path/filepath"
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

	AdminKubeconfigCAValidityDays                      = 365 * 10
	AdminKubeconfigClientCertValidityDays              = 365 * 10
	AggregatorFrontProxySignerCAValidityDays           = 30
	KubeAPIServerToKubeletCAValidityDays               = 365
	KubeControlPlaneSignerCAValidityDays               = 365
	KubeControllerManagerCSRSignerSignerCAValidityDays = 60
	KubeControllerManagerCSRSignerCAValidityDays       = 30
	EtcdSignerCAValidityDays                           = 365 * 10

	ClientCertValidityDays  = 30
	ServingCertValidityDays = 30

	ServiceCAValidityDays            = 790
	ServiceCAServingCertValidityDays = 730

	KubeAPIServerServingSignerCAValidityDays = 365 * 10
	KubeAPIServerServingCertValidityDays     = 365

	IngressSignerCAValidityDays    = 365 * 2
	IngressServingCertValidityDays = 365
)

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

func KubeControlPlaneSignerCertDir(certsDir string) string {
	return filepath.Join(certsDir, "kube-control-plane-signer")
}

func KubeSchedulerClientCertDir(certsDir string) string {
	return filepath.Join(KubeControlPlaneSignerCertDir(certsDir), "kube-scheduler")
}

func KubeControllerManagerClientCertDir(certsDir string) string {
	return filepath.Join(KubeControlPlaneSignerCertDir(certsDir), "kube-controller-manager")
}

func KubeAPIServerToKubeletSignerCertDir(certsDir string) string {
	return filepath.Join(certsDir, "kube-apiserver-to-kubelet-client-signer")
}

func KubeAPIServerToKubeletClientCertDir(certsDir string) string {
	return filepath.Join(KubeAPIServerToKubeletSignerCertDir(certsDir), "kube-apiserver-to-kubelet-client")
}

func AdminKubeconfigSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "admin-kubeconfig-signer")
}

func AdminKubeconfigClientCertDir(certsDir string) string {
	return filepath.Join(AdminKubeconfigSignerDir(certsDir), "admin-kubeconfig-client")
}

// KubeletCSRSignerSignerCertDir returns path to the signer that signs kubelet CSRs
// and the signer that signs CSRs of the CSR API
func KubeletCSRSignerSignerCertDir(certsDir string) string {
	return filepath.Join(certsDir, "kubelet-csr-signer-signer")
}

func CSRSignerCertDir(certsDir string) string {
	return filepath.Join(KubeletCSRSignerSignerCertDir(certsDir), "csr-signer")
}

func KubeletClientCertDir(certsDir string) string {
	return filepath.Join(CSRSignerCertDir(certsDir), "kubelet-client")
}

func KubeletServingCertDir(certsDir string) string {
	return filepath.Join(CSRSignerCertDir(certsDir), "kubelet-server")
}

func ServiceCADir(certsDir string) string {
	return filepath.Join(certsDir, "service-ca")
}

func RouteControllerManagerServingCertDir(certsDir string) string {
	return filepath.Join(ServiceCADir(certsDir), "route-controller-manager-serving")
}

func IngressCADir(certsDir string) string {
	return filepath.Join(certsDir, "ingress-ca")
}

func AggregatorSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "aggregator-signer")
}

func AggregatorClientCertDir(certsDir string) string {
	return filepath.Join(AggregatorSignerDir(certsDir), "aggregator-client")
}

func EtcdSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "etcd-signer")
}

func EtcdPeerCertDir(certsDir string) string {
	return filepath.Join(EtcdSignerDir(certsDir), "etcd-peer")
}

func EtcdAPIServerClientCertDir(certsDir string) string {
	return filepath.Join(EtcdSignerDir(certsDir), "apiserver-etcd-client")
}

func EtcdServingCertDir(certsDir string) string {
	return filepath.Join(EtcdSignerDir(certsDir), "etcd-serving")
}

func KubeAPIServerExternalSigner(certsDir string) string {
	return filepath.Join(certsDir, "kube-apiserver-external-signer")
}

func KubeAPIServerExternalServingCertDir(certsDir string) string {
	return filepath.Join(KubeAPIServerExternalSigner(certsDir), "kube-external-serving")
}

func KubeAPIServerLocalhostSigner(certsDir string) string {
	return filepath.Join(certsDir, "kube-apiserver-localhost-signer")
}

func KubeAPIServerLocalhostServingCertDir(certsDir string) string {
	return filepath.Join(KubeAPIServerLocalhostSigner(certsDir), "kube-apiserver-localhost-serving")
}

func KubeAPIServerServiceNetworkSigner(certsDir string) string {
	return filepath.Join(certsDir, "kube-apiserver-service-network-signer")
}

func KubeAPIServerServiceNetworkServingCertDir(certsDir string) string {
	return filepath.Join(KubeAPIServerServiceNetworkSigner(certsDir), "kube-apiserver-service-network-serving")
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

func ServiceAccountTokenCABundlePath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "service-account-token-ca.crt")
}
