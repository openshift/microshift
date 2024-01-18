/*
Copyright © 2021 MicroShift Contributors

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
package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	tcpnet "net"
	"os"
	"strings"

	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"
)

const keySize = 2048

func EnsureKeyPair(pubKeyPath, privKeyPath string) error {
	if _, err := getKeyPair(pubKeyPath, privKeyPath); err == nil {
		return nil
	}

	return GenKeys(pubKeyPath, privKeyPath)
}

// GenKeys generates and save rsa keys
func GenKeys(pubPath, keyPath string) error {
	rsaKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	keyPEM, err := keyutil.MarshalPrivateKeyToPEM(rsaKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key to PEM: %w", err)
	}

	pubPEM, err := PublicKeyToPem(&rsaKey.PublicKey)
	if err != nil {
		return err
	}

	if err := keyutil.WriteKey(keyPath, keyPEM); err != nil {
		return fmt.Errorf("failed to write the private key to %s: %w", keyPath, err)
	}

	if err := os.WriteFile(pubPath, pubPEM, 0400); err != nil {
		return fmt.Errorf("failed to write public key to %s: %w", pubPath, err)
	}

	return nil
}

// PublicKeyToPem converts an rsa.PublicKey object to pem string
func PublicKeyToPem(key *rsa.PublicKey) ([]byte, error) {
	keyInBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to MarshalPKIXPublicKey: %w", err)
	}
	keyinPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: keyInBytes,
		},
	)
	return keyinPem, nil
}

func getKeyPair(pubKeyPath, privKeyPath string) (*rsa.PrivateKey, error) {
	pubKeys, err := keyutil.PublicKeysFromFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}
	if len(pubKeys) > 1 {
		return nil, fmt.Errorf("too many pub keys in file %s", pubKeyPath)
	}

	privKey, err := keyutil.PrivateKeyFromFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("only RSA private keys are currently supported")
	}

	if !rsaPrivKey.PublicKey.Equal(pubKeys[0].(*rsa.PublicKey)) {
		return nil, fmt.Errorf("public and private keys don't match")
	}

	return rsaPrivKey, nil
}

func IsCertAllowed(clusterNetwork []string, serviceNetwork []string, certPath string, extraNames []string) (bool, error) {
	certsSNIs, err := GetSNIsFromCert(certPath, extraNames)
	if err != nil {
		return false, err
	}

	// iterate over the SNIs and generate kubeconfig files
	// TODO: do we want to read cert.IPAddresses?
	for _, dns := range certsSNIs {
		// check if SNI is allowed (non local or wildcard)
		if !VerifyAllowedSNI(clusterNetwork, serviceNetwork, dns) {
			klog.Infof("Certificate SNI is not allowed %s", dns)
			return false, nil
		}
	}

	return true, nil
}

// GetSNIsFromCert - get list of unique SNIs from certificate
func GetSNIsFromCert(certPath string, extraNames []string) ([]string, error) {
	pemByte, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemByte)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	klog.Infof("CN contains %s", cert.Subject.CommonName)
	klog.Infof("SAN contains %v DNSNames", cert.DNSNames)

	// create one list to hold all the SNIs
	certsSNIs := append(cert.DNSNames, cert.Subject.CommonName)

	// add  Certificate SAN ipaddress
	for _, ipaddress := range cert.IPAddresses {
		certsSNIs = append(certsSNIs, ipaddress.String())
	}
	// add Configuration names
	if len(extraNames) > 0 {
		certsSNIs = append(certsSNIs, extraNames...)
	}
	return certsSNIs, nil
}

// IsWildcardDNS - check if DNS is a wildcard
func IsWildcardDNS(val string) bool {
	return strings.Contains(val, "*.")
}

func IsWildcardEntry(val string) bool {
	return strings.Contains(val, "*.")
}

// verifyAllowedSNI checks if sni is allowed
// return bool: true, false
func VerifyAllowedSNI(clusterNetwork []string, serviceNetwork []string, sni string) bool {
	forbiddenValues := []string{"localhost", "svc.cluster.local", "openshift.default.svc", "kubernetes.default.svc"}

	ipAddress := tcpnet.ParseIP(sni)

	//check if IPAddress or DNS record
	if ipAddress == nil {
		for _, val := range forbiddenValues {
			if strings.Contains(sni, val) {
				return false
			}
		}
	} else {
		if ContainIPANetwork(ipAddress, clusterNetwork) ||
			ContainIPANetwork(ipAddress, serviceNetwork) ||
			ContainIPANetwork(ipAddress, []string{"127.0.0.1/8", "169.254.169.2/29"}) {
			return false
		}
	}
	return true
}
