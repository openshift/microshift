package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"

	gengogenerator "k8s.io/gengo/v2/generator"
)

type jsonFile struct {
}

func NewGengoJSONFile() gengogenerator.FileType {
	return &jsonFile{}
}

func (a jsonFile) AssembleFile(f *gengogenerator.File, path string) error {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, &f.Body)
	return err
}

func (a jsonFile) VerifyFile(f *gengogenerator.File, path string) error {
	if path == "-" {
		// Nothing to verify against.
		return nil
	}

	formatted := f.Body.Bytes()
	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read file %q for comparison: %v", path, err)
	}
	if bytes.Equal(formatted, existing) {
		return nil
	}

	// Be nice and find the first place where they differ
	// (Copied from gengo's default file type)
	i := 0
	for i < len(formatted) && i < len(existing) && formatted[i] == existing[i] {
		i++
	}
	eDiff, fDiff := existing[i:], formatted[i:]
	if len(eDiff) > 100 {
		eDiff = eDiff[:100]
	}
	if len(fDiff) > 100 {
		fDiff = fDiff[:100]
	}
	return fmt.Errorf("output for %q differs; first existing/expected diff: \n  %q\n  %q", path, string(eDiff), string(fDiff))
}
