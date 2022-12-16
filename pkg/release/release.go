/*
Copyright © 2021 MicroShift Contributors

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

package release

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	embedded "github.com/openshift/microshift/assets"
)

var Base = "undefined"

var Image = map[string]string{}

func init() {
	arch_replacer := strings.NewReplacer("amd64", "x86_64", "arm64", "aarch64")
	arch := arch_replacer.Replace(runtime.GOARCH)

	release_file := "release/release-" + arch + ".json"
	data, err := embedded.Asset(release_file)
	if err != nil {
		// If there is no release file for this architecture, work with the generic specs
		return
	}

	var release map[string]any
	if err := json.Unmarshal(data, &release); err != nil {
		panic(fmt.Errorf("unmarshaling %s: %v", release_file, err))
	}

	// Copy in the OCP base version
	metadata := release["release"].(map[string]any)
	Base = metadata["base"].(string)

	// Copy in the pullspecs, translating the keys as used by the OCP release image
	// (with '-'s) into keys we can use in go templates (need to use '_'s instead).
	images := release["images"].(map[string]any)
	for name, pullspec := range images {
		name := strings.Replace(name, "-", "_", -1)
		Image[name] = pullspec.(string)
	}
}
