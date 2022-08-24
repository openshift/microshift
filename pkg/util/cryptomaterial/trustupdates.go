package cryptomaterial

import (
	"fmt"
	"os"
)

func AddToTotalClientCABundle(certsDir string, cacert []byte) error {
	clientCABundlePath := TotalClientCABundlePath(certsDir)

	f, err := os.OpenFile(clientCABundlePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %q for writing: %w", clientCABundlePath, err)
	}
	defer f.Close()

	f.WriteString("\n")
	f.Write(cacert)

	return nil
}
