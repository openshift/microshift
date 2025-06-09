/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/printf"
)

// TestGoVet checks that "go vet" detects incorrect klog calls like
// mismatched format specifiers and arguments.
func TestGoVet(t *testing.T) {
	testdata := analysistest.TestData()
	src := path.Join(testdata, "src")
	t.Cleanup(func() {
		os.RemoveAll(src)
	})

	// analysistest doesn't support using existing code
	// via modules (https://github.com/golang/go/issues/37054).
	// Populating the "testdata/src" directory with the
	// result of "go mod vendor" is a workaround.
	cmd := exec.Command("go", "mod", "vendor", "-o", src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\nOutput: %s", cmd, err, string(out))
	}

	analyzer := printf.Analyzer
	analysistest.Run(t, testdata, analyzer, "")
}
