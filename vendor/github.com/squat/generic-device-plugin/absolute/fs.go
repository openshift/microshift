// Copyright 2023 the generic-device-plugin authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package absolute

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var _ fs.FS = (*FS)(nil)
var _ fs.GlobFS = (*FS)(nil)
var _ fs.ReadFileFS = (*FS)(nil)
var _ fs.StatFS = (*FS)(nil)
var _ fs.ReadDirFS = (*FS)(nil)
var _ fs.SubFS = (*FS)(nil)

type FS struct {
	fs.FS
	prefix string
}

func New(fsys fs.FS, prefix string) fs.FS {
	return &FS{fsys, prefix}
}

func (f *FS) Open(name string) (fs.File, error) {
	name, err := filepath.Rel(f.prefix, name)
	if err != nil {
		return nil, err
	}
	return f.FS.Open(strings.TrimPrefix(name, f.prefix))
}

func (f *FS) Glob(pattern string) ([]string, error) {
	name, err := filepath.Rel(f.prefix, pattern)
	if err != nil {
		return nil, err
	}
	matches, err := fs.Glob(f.FS, name)
	if err != nil {
		return nil, err
	}
	for i := range matches {
		matches[i] = filepath.Join(f.prefix, matches[i])
	}
	return matches, nil
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	name, err := filepath.Rel(f.prefix, name)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(f.FS, name)
}

func (f *FS) Stat(name string) (fs.FileInfo, error) {
	name, err := filepath.Rel(f.prefix, name)
	if err != nil {
		return nil, err
	}
	return fs.Stat(f.FS, name)
}

func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	name, err := filepath.Rel(f.prefix, name)
	if err != nil {
		return nil, err
	}
	return fs.ReadDir(f.FS, name)
}

func (f *FS) Sub(name string) (fs.FS, error) {
	name, err := filepath.Rel(f.prefix, name)
	if err != nil {
		return nil, err
	}
	return fs.Sub(f.FS, name)
}
