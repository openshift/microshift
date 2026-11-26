package certchains

import (
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/openshift/library-go/pkg/crypto"
)

// serialFileGenerator keeps library-go's counter format, but replaces the file
// atomically and durably before returning a serial for certificate issuance.
// PKI operations are serialized by the certificate operation lock; the mutex
// also protects concurrent callers sharing this generator.
type serialFileGenerator struct {
	file   string
	serial int64
	mu     sync.Mutex
	// Inject failures on either side of the counter replacement in tests.
	rename  func(string, string) error
	syncDir func(*os.File) error
}

func loadSigningCA(certFile, keyFile, serialFile string) (*crypto.CA, error) {
	// Passing the serial path to GetCA would install library-go's in-place
	// writer, including a write while loading a zero counter.
	ca, err := crypto.GetCA(certFile, keyFile, "")
	if err != nil {
		return nil, err
	}
	serial, err := readSerialCounter(serialFile)
	if err != nil {
		return nil, err
	}
	ca.SerialGenerator = &serialFileGenerator{file: serialFile, serial: serial, rename: os.Rename, syncDir: (*os.File).Sync}
	return ca, nil
}

// Initialize only missing counters. Existing corrupt counters must not be
// reset: they record issuance history beyond the certificates still on disk.
func ensureSerialGenerator(serialFile string) (*serialFileGenerator, error) {
	serial, err := readSerialCounter(serialFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	g := &serialFileGenerator{file: serialFile, serial: serial, rename: os.Rename, syncDir: (*os.File).Sync}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(serialFile), 0755); err != nil { //nolint:gosec // Preserve library-go's CA directory permissions.
			return nil, err
		}
		// Zero is unused and one is reserved for the CA.
		if err := g.write(1); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func readSerialCounter(path string) (int64, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	serial, err := strconv.ParseInt(strings.TrimSpace(string(contents)), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid serial counter %q: %w", path, err)
	}
	if serial < 0 {
		return 0, fmt.Errorf("negative serial counter in %q", path)
	}
	return serial, nil
}

func (g *serialFileGenerator) Next(_ *x509.Certificate) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	serial, err := readSerialCounter(g.file)
	if err != nil {
		return 0, err
	}
	if serial != g.serial {
		return 0, fmt.Errorf("serial file %s out of sync ram=%d disk=%d", g.file, g.serial, serial)
	}
	if serial == math.MaxInt64 {
		return 0, fmt.Errorf("serial counter exhausted in %q", g.file)
	}
	next := max(serial, 1) + 1
	if err := g.write(next); err != nil {
		return 0, err
	}
	return next, nil
}

func (g *serialFileGenerator) write(serial int64) error {
	dir, err := os.Open(filepath.Dir(g.file))
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()
	file, err := os.CreateTemp(dir.Name(), ".serial-")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }() //nolint:gosec // CreateTemp chooses the name inside the signer directory.
	// OpenSSL expects an even number of hexadecimal digits and a final newline.
	contents := fmt.Sprintf("%X", serial)
	if len(contents)%2 != 0 {
		contents = "0" + contents
	}
	_, err = file.WriteString(contents + "\n")
	err = errors.Join(err, file.Sync(), file.Close())
	if err != nil {
		return err
	}
	if err := g.rename(file.Name(), g.file); err != nil {
		return err
	}
	// The new counter is visible on disk even if syncing the directory fails.
	g.serial = serial
	return g.syncDir(dir)
}
