package cryptomaterial

import "path/filepath"

const (
	CACertFileName     = "ca.crt"
	CAKeyFileName      = "ca.key"
	CASerialsFileName  = "serial.txt"
	ServerCertFileName = "server.crt"
	ServerKeyFileName  = "server.key"
	ClientCertFileName = "client.crt"
	ClientKeyFileName  = "client.key"

	ClientCAValidityDays   = 60
	ClientCertValidityDays = 30
)

func CertsDirectory(dataPath string) string { return filepath.Join(dataPath, "certs") }

func CACertPath(dir string) string    { return filepath.Join(dir, CACertFileName) }
func CAKeyPath(dir string) string     { return filepath.Join(dir, CAKeyFileName) }
func CASerialsPath(dir string) string { return filepath.Join(dir, CASerialsFileName) }

func ClientCertPath(dir string) string { return filepath.Join(dir, ClientCertFileName) }
func ClientKeyPath(dir string) string  { return filepath.Join(dir, ClientKeyFileName) }

func KubeControlPlaneSignerCertDir(certsDir string) string {
	return filepath.Join(certsDir, "kube-control-plane-signer")
}

func KubeSchedulerClientCertDir(certsDir string) string {
	return filepath.Join(KubeControlPlaneSignerCertDir(certsDir), "kube-scheduler")
}

func KubeControllerManagerClientCertDir(certsDir string) string {
	return filepath.Join(KubeControlPlaneSignerCertDir(certsDir), "kube-controller-manager")
}

// TotalClientCABundlePath returns the path to the cert bundle with all client certificate signers
func TotalClientCABundlePath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "client-ca.crt")
}

// UltimateTrustBundlePath returns the path to the cert bundle with the root certificate
func UltimateTrustBundlePath(certsDir string) string {
	return filepath.Join(certsDir, "ca-bundle", "ca-bundle.crt")
}
