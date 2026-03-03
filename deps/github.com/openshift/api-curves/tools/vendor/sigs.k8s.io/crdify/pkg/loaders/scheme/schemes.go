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

package scheme

// Scheme is representation of the schemes that correspond
// to Loader types.
type Scheme string

const (
	// SchemeKubernetes represents the scheme used to
	// signal that a Loader should load from Kubernetes.
	SchemeKubernetes = "kube"

	// SchemeGit represents the scheme used to signal
	// that a Loader should load from a git repository.
	SchemeGit = "git"

	// SchemeFile represents the scheme used to signal
	// that a Loader should load from a file.
	SchemeFile = "file"
)
