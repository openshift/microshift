// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"

	"github.com/spf13/afero"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// File is a Loader implementation to load a CustomResourceDefinition
// from a file.
type File struct {
	// filesystem is the filesystem used to load the file containing
	// the CustomResourceDefinition.
	filesystem afero.Fs
}

// New returns a new instance of the File Loader
// using the provided afero.Fs as the underlying file system.
func New(filesystem afero.Fs) *File {
	return &File{
		filesystem: filesystem,
	}
}

// Load parses the hostname and path of the provided URL to determine the file containing the CustomResourceDefinition
// and reads it into a new CustomResourceDefinition object.
func (f *File) Load(_ context.Context, location *url.URL) (*apiextensionsv1.CustomResourceDefinition, error) {
	filePath, err := filepath.Abs(path.Join(location.Hostname(), location.Path))
	if err != nil {
		return nil, fmt.Errorf("ensuring absolute path: %w", err)
	}

	file, err := f.filesystem.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", filePath, err)
	}
	//nolint:errcheck
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}

	err = yaml.Unmarshal(fileBytes, crd)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling contents of file %q: %w", filePath, err)
	}

	return crd, nil
}
