package cryptomaterial

import (
	"path/filepath"
)

const (
	CACertFileName     = "ca.crt"
	CAKeyFileName      = "ca.key"
	CASerialsFileName  = "serial.txt"
	ServerCertFileName = "server.crt"
	ServerKeyFileName  = "server.key"
	ClientCertFileName = "client.crt"
	ClientKeyFileName  = "client.key"

	ClientCAValidityDays                  = 60
	ClientCertValidityDays                = 30
	AdminKubeconfigClientCertValidityDays = 365 * 10

	ServiceCAValidityDays = 790
)

func CertsDirectory(dataPath string) string { return filepath.Join(dataPath, "certs") }

func CACertPath(dir string) string    { return filepath.Join(dir, CACertFileName) }
func CAKeyPath(dir string) string     { return filepath.Join(dir, CAKeyFileName) }
func CASerialsPath(dir string) string { return filepath.Join(dir, CASerialsFileName) }

func ClientCertPath(dir string) string { return filepath.Join(dir, ClientCertFileName) }
func ClientKeyPath(dir string) string  { return filepath.Join(dir, ClientKeyFileName) }

func ServingCertPath(dir string) string { return filepath.Join(dir, ServerCertFileName) }
func ServingKeyPath(dir string) string  { return filepath.Join(dir, ServerKeyFileName) }

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
	return filepath.Join(KubeletCSRSignerSignerCertDir(certsDir), "kubelet-client")
}

func ServiceCADir(certsDir string) string {
	return filepath.Join(certsDir, "service-ca")
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
