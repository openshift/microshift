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
	"os"

	"k8s.io/client-go/util/keyutil"
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
