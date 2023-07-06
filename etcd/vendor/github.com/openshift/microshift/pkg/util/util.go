package util

import (
	"errors"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/sets"
)

func Must(err error) {
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
}

func Default(s string, defaultS string) string {
	if s == "" {
		return defaultS
	}
	return s
}

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking if path (%s) exists failed: %w", path, err)
	}
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0700)
}

func PathExistsAndIsNotEmpty(path string, ignores ...string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("checking if path (%s) exists failed: %w", path, err)
		}
	}

	if !fi.IsDir() {
		return fi.Size() != 0, nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to ReadDir %q: %w", path, err)
	}

	iset := sets.New[string](ignores...)
	for _, f := range files {
		if iset.Has(f.Name()) {
			continue
		}
		return true, nil
	}

	return false, nil
}
