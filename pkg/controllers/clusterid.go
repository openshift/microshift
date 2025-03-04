/*
Copyright © 2024 MicroShift Contributors

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
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	ClusterIDFilePath = filepath.Join(config.DataDir, "cluster-id")
)

type ClusterID struct {
	cfg *config.Config
}

func NewClusterID(cfg *config.Config) *ClusterID {
	s := &ClusterID{}
	s.cfg = cfg
	return s
}

func (s *ClusterID) Name() string { return "cluster-id-manager" }
func (s *ClusterID) Dependencies() []string {
	return []string{"kube-apiserver"}
}

func (s *ClusterID) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	defer close(ready)

	// Read the 'kube-system' namespace attributes
	restConfig, err := clientcmd.BuildConfigFromFlags("", s.cfg.KubeConfigPath(config.KubeAdmin))
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig admin path: %v", err)
	}
	coreClient := clientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "core-agent"))
	namespace, err := coreClient.Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to read 'kube-system' namespace attributes: %v", err)
	}

	// Use the 'kube-system' namespace metadata UID as the MicroShift Cluster ID
	clusterID := string(namespace.ObjectMeta.UID)
	// Write <config.DataDir>/cluster-id file if it does not already exist
	// or has inconsistent contents
	err = initClusterIDFile(clusterID)
	if err != nil {
		return fmt.Errorf("failed to initialize cluster ID file: %v", err)
	}
	// Log the cluster ID
	klog.Infof("MicroShift Cluster ID: %v", clusterID)

	return ctx.Err()
}

func initClusterIDFile(clusterID string) error {
	// Read and verify the cluster ID file if it already exists,
	// logging a warning if the cluster ID is inconsistent
	data, err := os.ReadFile(ClusterIDFilePath)
	if err != nil && !os.IsNotExist(err) {
		// File exists, but cannot be read
		return err
	}
	if len(data) > 0 {
		if string(data) == clusterID {
			// Consistent cluster ID file exists
			return nil
		}
		klog.Warningf("Overwriting an inconsistent MicroShift Cluster ID '%v' in '%v' file", string(data), ClusterIDFilePath)
	}

	// Write a new cluster ID file
	klog.Infof("Writing MicroShift Cluster ID '%v' to '%v'", clusterID, ClusterIDFilePath)
	return os.WriteFile(ClusterIDFilePath, []byte(clusterID), 0400)
}
