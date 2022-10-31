package cryptomaterial

import (
	"fmt"
	"os"
	"path/filepath"
)

func AddToTotalClientCABundle(certsDir string, cacerts ...[]byte) error {
	return appendCertsToFile(TotalClientCABundlePath(certsDir), cacerts...)
}

func AddToKubeletClientCABundle(certsDir string, cacerts ...[]byte) error {
	return appendCertsToFile(KubeletClientCAPath(certsDir), cacerts...)
}

func appendCertsToFile(bundlePath string, certs ...[]byte) error {
	// ensure parent dir
	if err := os.MkdirAll(filepath.Dir(bundlePath), os.FileMode(0755)); err != nil {
		return err
	}

	f, err := os.OpenFile(bundlePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %q for writing: %w", bundlePath, err)
	}
	defer f.Close()

	for _, c := range certs {
		f.WriteString("\n")
		f.Write(c)
	}

	return nil
}
