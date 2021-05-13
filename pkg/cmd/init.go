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
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/constant"
	"github.com/openshift/microshift/pkg/util"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "openshift init",
	Long:  `openshift init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initAll(args)
	},
}

func initAll(args []string) error {
	// create CA and keys
	if err := initCerts(); err != nil {
		return err
	}
	// create configs
	if err := initServerConfig(); err != nil {
		return err
	}
	if err := initNodeConfig(); err != nil {
		return err
	}
	// create kubeconfig for kube-scheduler, kubelet, openshift-apiserver,controller-manager
	if err := initKubeconfig(); err != nil {
		return err
	}

	return nil
}

func initCerts() error {
	// etcd
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	ip, err := util.GetHostIP()
	if err != nil {
		return fmt.Errorf("failed to get host IP: %v", err)
	}
	// store root CA for all
	//TODO generate ca bundles for each component
	if err := util.StoreRootCA("https://kubernetes.svc", "/etc/kubernetes/ushift-certs/ca-bundle",
		"ca-bundle.crt", "ca-bundle.key",
		[]string{"https://kubernetes.svc"}); err != nil {
		return err
	}

	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	if err := util.GenCerts("etcd-server", "/etc/kubernetes/ushift-certs/secrets/etcd-all-serving",
		"etcd-serving.crt", "etcd-serving.key",
		[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}

	if err := util.GenCerts("etcd-peer", "/etc/kubernetes/ushift-certs/secrets/etcd-all-peer",
		"etcd-peer.crt", "etcd-peer.key",
		[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}

	// kube-apiserver
	if err := util.GenCerts("etcd-client", "/etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client",
		"tls.crt", "tls.key",
		[]string{"localhost", ip, "127.0.0.1", hostname}); err != nil {
		return err
	}
	if err := util.GenCerts("kube-apiserver", "/etc/kubernetes/ushift-certs/kube-apiserver/secrets/service-network-serving-certkey",
		"tls.crt", "tls.key",
		[]string{"kube-apiserver", ip, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}); err != nil {
		return err
	}
	if err := util.GenKeys("/etc/kubernetes/ushift-resources/kube-apiserver/secrets/service-account-key",
		"service-account.crt", "service-account.key"); err != nil {
		return err
	}
	if err := util.GenKeys("/etc/kubernetes/ushift-resources/kube-apiserver/secrets/service-account-signing-key",
		"service-account.crt", "service-account.key"); err != nil {
		return err
	}
	if err := util.GenCerts("system:masters", "/etc/kubernetes/ushift-certs/kube-apiserver/secrets/aggregator-client",
		"tls.crt", "tls.key",
		[]string{"system:admin", "system:masters"}); err != nil {
		return err
	}
	if err := util.GenCerts("system:masters", "/etc/kubernetes/ushift-resources/kube-apiserver/secrets/kubelet-client",
		"tls.crt", "tls.key",
		[]string{"kube-apiserver", "system:kube-apiserver", "system:masters"}); err != nil {
		return err
	}
	if err := util.GenKeys("/etc/kubernetes/ushift-resources/kube-apiserver/sa-public-key",
		"serving-ca.pub", "serving-ca.key"); err != nil {
		return err
	}

	// ocp
	if err := util.GenCerts("openshift-apiserver", "/etc/kubernetes/ushift-resources/ocp-apiserver/secrets",
		"tls.crt", "tls.key",
		[]string{"openshift-apiserver", ip, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}); err != nil {
		return err
	}
	if err := util.GenCerts("openshift-controller-manager", "/etc/kubernetes/ushift-resources/ocp-controller-manager/secrets",
		"tls.crt", "tls.key",
		[]string{"openshift-controller-manager", ip, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}); err != nil {
		return err
	}
	return nil
}

func initServerConfig() error {
	if err := config.KubeAPIServerConfig("/etc/kubernetes/ushift-resources/kube-apiserver/config/config.yaml", "" /*svc CIDR*/); err != nil {
		return err
	}
	if err := config.KubeControllerManagerConfig("/etc/kubernetes/ushift-resources/kube-controller-manager/config/config.yaml"); err != nil {
		return err
	}
	if err := config.OpenShiftAPIServerConfig("/etc/kubernetes/ushift-resources/openshift-apiserver/config/config.yaml"); err != nil {
		return err
	}
	if err := config.OpenShiftControllerManagerConfig("/etc/kubernetes/ushift-resources/openshift-controller-manager/config/config.yaml"); err != nil {
		return err
	}
	if err := config.KubeSchedulerConfig("/etc/kubernetes/ushift-resources/kube-scheduler/config/config.yaml"); err != nil {
		return err
	}
	return nil
}

func initNodeConfig() error {
	if err := config.KubeletConfig("/etc/kubernetes/ushift-resources/kubelet/config/config.yaml"); err != nil {
		return err
	}
	if err := config.OpenShiftSDNConfig("/etc/kubernetes/ushift-resources/openshift-sdn/config/config.yaml"); err != nil {
		return err
	}
	return nil
}

func initKubeconfig() error {
	if err := util.Kubeconfig(constant.AdminKubeconfigPath, "system:admin", []string{"system:masters"}); err != nil {
		return err
	}
	if err := util.Kubeconfig(constant.KubeAPIKubeconfigPath, "kube-apiserver", []string{"system:kube-apiserver", "system:masters"}); err != nil {
		return err
	}
	if err := util.Kubeconfig(constant.KubeControllerManagerKubeconfigPath, "kube-controller-manager", []string{"system:kube-controller-manager"}); err != nil {
		return err
	}
	if err := util.Kubeconfig(constant.KubeSchedulerKubeconfigPath, "kube-scheduler", []string{"system:kube-scheduler"}); err != nil {
		return err
	}
	if err := util.Kubeconfig(constant.KubeletKubeconfigPath, "kubelet", []string{"system:node", "system:node-bootstrapper"}); err != nil {
		return err
	}
	return nil
}
