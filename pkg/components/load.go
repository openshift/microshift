/*
Copyright © 2021 Microshift Contributors

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
package components

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	assets "github.com/openshift/microshift/pkg/assets/components"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/release"
)

type ComponentLoader struct {
	manifestDir string
	components  []string
}

func NewComponentLoader(cfg *config.MicroshiftConfig) *ComponentLoader {
	components, _ := assets.AssetDir("assets/components")
	return &ComponentLoader{
		manifestDir: filepath.Join(cfg.DataDir, "manifests"),
		components:  components,
	}
}

func (s *ComponentLoader) Name() string           { return "component-loader" }
func (s *ComponentLoader) Dependencies() []string { return []string{"kube-apiserver"} }

func (s *ComponentLoader) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	// Write embedded component manifests to the manifest folder
	for _, component := range s.components {
		componentDir := filepath.Join(s.manifestDir, component)
		if err := assets.RestoreAssets(s.manifestDir, "assets/components/"+component); err != nil {
			return err
		}
		os.RemoveAll(componentDir)
		if err := os.Rename(filepath.Join(s.manifestDir, "assets", "components", component), componentDir); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(filepath.Join(s.manifestDir, "assets")); err != nil {
		return err
	}

	// Write main kustomization.yaml that includes selected components and maps images according to the release
	f, err := os.OpenFile(filepath.Join(s.manifestDir, "kustomization.yaml"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	f.WriteString("apiVersion: kustomize.config.k8s.io/v1beta1\n")
	f.WriteString("kind: Kustomization\n")
	f.WriteString("\n")
	f.WriteString("bases:\n")
	for _, component := range s.components {
		f.WriteString("  - " + component + "\n")
	}
	f.WriteString("\n")
	f.WriteString("images:\n")
	for k, v := range release.Image {
		f.WriteString("  - name: " + k + "\n")
		f.WriteString("    newName: " + v + "\n")
	}
	f.Close()

	// Signal we're ready and exit
	logrus.Infof("%s is ready", s.Name())
	close(ready)

	return ctx.Err()
}
