package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// GetPathPrefix calculates the pathPrefix that needs to be trimmed from the output file.
// The generator will generate the output file to the current working directory plus the
// package path name.
// This function calculates what is needed to be trimmed from the package path name to
// make sure the output ends up in the correct directory.
// Eg. if the package is github.com/openshift.io/api/machine/v1,
//   - the current working directory is /home/user/go/src, then the path would be
//     github.com/openshift.io/api/machine/v1 and so the output would be the empty string.
//   - the current working directory is /home/user/go/src/github.com/openshift.io/api, then
//     the path would be machine/v1 and so the output would be github.com/openshift.io/api.
func GetPathPrefix(wd, path, packagePath string) (string, error) {
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "../") {
		return "", errors.New("cannot generate deepcopy functions for a path outside of the working directory")
	}

	if !strings.HasSuffix(packagePath, cleanPath) {
		return "", fmt.Errorf("package path %s does not match with input path %s, expected package path to end with input path", packagePath, cleanPath)
	}

	return strings.TrimSuffix(packagePath, cleanPath), nil
}
