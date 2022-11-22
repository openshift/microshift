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
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/keyutil"
)

const (
	defaultDurationDays = 365
	defaultDuration     = defaultDurationDays * 24 * time.Hour
	defaultHostname     = "localhost"

	keySize = 2048

	ValidityOneDay   = 24 * time.Hour
	ValidityOneYear  = 365 * ValidityOneDay
	ValidityTenYears = 10 * ValidityOneYear
)

// GenKeys generates and save rsa keys
func GenKeys(dir, pubFilename, keyFilename string) error {
	rsaKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return errors.Wrap(err, "error generating RSA private key")
	}

	keyPEM, err := keyutil.MarshalPrivateKeyToPEM(rsaKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key to PEM: %v", err)
	}

	pubPEM, err := PublicKeyToPem(&rsaKey.PublicKey)
	if err != nil {
		return err
	}

	keyPath := filepath.Join(dir, keyFilename)
	pubPath := filepath.Join(dir, pubFilename)

	if err := keyutil.WriteKey(keyPath, keyPEM); err != nil {
		return fmt.Errorf("failed to write the private key to %s: %v", keyPath, err)
	}

	ioutil.WriteFile(pubPath, pubPEM, 0644)

	return nil
}

// PublicKeyToPem converts an rsa.PublicKey object to pem string
func PublicKeyToPem(key *rsa.PublicKey) ([]byte, error) {
	keyInBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to MarshalPKIXPublicKey")
	}
	keyinPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: keyInBytes,
		},
	)
	return keyinPem, nil
}
