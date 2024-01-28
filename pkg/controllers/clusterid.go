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
	"os"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
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
		panic(err)
	}
	coreClient := clientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "core-agent"))
	namespace, err := coreClient.Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	// Use the 'kube-system' namespace metadata UID as the MicroShift Cluster ID
	clusterID := string(namespace.ObjectMeta.UID)
	// Write <config.DataDir>/cluster-id file if it does not already exist
	initClusterIDFile(clusterID)
	// Log the cluster ID
	klog.Infof("MicroShift Cluster ID: %v", clusterID)

	return ctx.Err()
}

func initClusterIDFile(clusterID string) {
	// The location of the cluster ID file
	fileName := filepath.Join(config.DataDir, "cluster-id")

	// Do not create the cluster ID file if it already exists
	_, err := os.Stat(fileName)
	if !os.IsNotExist(err) {
		return
	}

	// Write the cluster ID to a new file
	klog.Infof("Writing MicroShift Cluster ID '%v' to '%v'", clusterID, fileName)
	err = os.WriteFile(fileName, []byte(clusterID), 0400)
	if err != nil {
		panic(err)
	}
}
