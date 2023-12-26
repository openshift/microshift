/*
Copyright © 2022 MicroShift Contributors

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
package controllers

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/version"

	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

type VersionManager struct {
	cfg *config.Config
}

func NewVersionManager(cfg *config.Config) *VersionManager {
	s := &VersionManager{}
	s.cfg = cfg
	return s
}

func (s *VersionManager) Name() string { return "version-manager" }
func (s *VersionManager) Dependencies() []string {
	return []string{"kube-apiserver"}
}

func (s *VersionManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	var cm = "version/microshift-version.yaml"

	defer close(stopped)
	defer close(ready)

	versionInfo := version.Get()
	// cluster ID read from <config.DataDir>/cluster-id file
	clusterIDFromDB := initClusterID()
	var data = map[string]string{
		"major":     versionInfo.Major,
		"minor":     versionInfo.Minor,
		"patch":     versionInfo.Patch,
		"version":   versionInfo.String(),
		"clusterid": clusterIDFromDB,
	}

	kubeConfigPath := s.cfg.KubeConfigPath(config.KubeAdmin)
	if err := assets.ApplyConfigMapWithData(ctx, cm, data, kubeConfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v, %v", cm, err)
		return err
	}

	return ctx.Err()
}

func createClusterIDFile(fileName string) {
	uuidBytes, err := uuid.New().MarshalText()
	if err != nil {
		klog.Fatal(err)
	}

	// Write the UUID to a new file
	err = os.WriteFile(fileName, uuidBytes, 0400)
	if err != nil {
		klog.Fatal(err)
	}
}

func initClusterID() string {
	// The location of the cluster ID file
	fileName := filepath.Join(config.DataDir, "cluster-id")

	// The default cluster ID is empty, all zeros
	uuidBytes, err := uuid.Nil.MarshalText()
	if err != nil {
		klog.Fatal(err)
	}

	// Cluster ID can only be initialized or read when running as root
	if os.Geteuid() != 0 {
		return string(uuidBytes)
	}

	// Cannot create cluster ID file if the MicroShift DB directory does not exist
	var info os.FileInfo
	info, err = os.Stat(config.DataDir)
	if err != nil || !info.IsDir() {
		klog.Fatalf(
			"Cannot create MicroShift Cluster ID file before the '%s' directory is created",
			config.DataDir)
	}

	// Create the cluster ID file if it does not already exist
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		createClusterIDFile(fileName)
	}

	// Read the cluster ID from the disk
	uuidBytes, err = os.ReadFile(fileName)
	if err != nil {
		klog.Fatal(err)
	}

	return strings.TrimSpace(string(uuidBytes))
}
