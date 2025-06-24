package cryptomaterial

import (
	"fmt"
	"os"
	"path/filepath"
)

func AppendCertsToFile(bundlePath string, certs ...[]byte) error {
	// ensure parent dir
	if err := os.MkdirAll(filepath.Dir(bundlePath), os.FileMode(0755)); err != nil {
		return err
	}

	f, err := os.OpenFile(bundlePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %q for writing: %w", bundlePath, err)
	}
	defer func() { _ = f.Close() }()

	for _, c := range certs {
		if _, err = f.WriteString("\n"); err != nil {
			return err
		}
		if _, err = f.Write(c); err != nil {
			return err
		}
	}
	return nil
}
