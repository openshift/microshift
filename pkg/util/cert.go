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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
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
	key, err := PrivateKey()
	if err != nil {
		return err
	}
	pub := &key.PublicKey
	pubBuff, err := PublicKeyToPem(pub)
	if err != nil {
		return err
	}
	keyBuff := PrivateKeyToPem(key)
	os.MkdirAll(dir, 0700)
	pubPath := filepath.Join(dir, pubFilename)
	keyPath := filepath.Join(dir, keyFilename)
	ioutil.WriteFile(pubPath, pubBuff, 0644)
	ioutil.WriteFile(keyPath, keyBuff, 0644)
	return err

}

// PrivateKey generates an RSA Private key and returns the value
func PrivateKey() (*rsa.PrivateKey, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, errors.Wrap(err, "error generating RSA private key")
	}

	return rsaKey, nil
}

// PrivateKeyToPem converts an rsa.PrivateKey object to pem string
func PrivateKeyToPem(key *rsa.PrivateKey) []byte {
	keyInBytes := x509.MarshalPKCS1PrivateKey(key)
	keyinPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyInBytes,
		},
	)
	return keyinPem
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
