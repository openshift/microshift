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
	"net"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/controllers"
	"github.com/openshift/microshift/pkg/util"
	ctrl "k8s.io/kubernetes/pkg/controlplane"
)

func initAll(cfg *config.MicroshiftConfig) error {
	// create CA and keys
	if err := initCerts(cfg); err != nil {
		return err
	}
	// create configs
	if err := initServerConfig(cfg); err != nil {
		return err
	}
	// create kubeconfig for kube-scheduler, kubelet, openshift-apiserver,controller-manager
	if err := initKubeconfig(cfg); err != nil {
		return err
	}

	return nil
}

func loadCA(cfg *config.MicroshiftConfig) error {
	return util.LoadRootCA(cfg.DataDir+"/certs/ca-bundle", "ca-bundle.crt", "ca-bundle.key")
}

func initCerts(cfg *config.MicroshiftConfig) error {
	_, svcNet, err := net.ParseCIDR(cfg.Cluster.ServiceCIDR)
	if err != nil {
		return err
	}

	_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
	if err != nil {
		return err
	}

	// store root CA for all
	//TODO generate ca bundles for each component
	if err := util.StoreRootCA("https://kubernetes.svc", cfg.DataDir+"/certs/ca-bundle",
		"ca-bundle.crt", "ca-bundle.key",
		[]string{"https://kubernetes.svc"}); err != nil {
		return err
	}

	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	if err := util.GenCerts("etcd-server", cfg.DataDir+"/certs/etcd",
		"etcd-serving.crt", "etcd-serving.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return err
	}

	if err := util.GenCerts("etcd-peer", cfg.DataDir+"/certs/etcd",
		"etcd-peer.crt", "etcd-peer.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return err
	}

	// kube-apiserver
	if err := util.GenCerts("etcd-client", cfg.DataDir+"/resources/kube-apiserver/secrets/etcd-client",
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return err
	}
	if err := util.GenCerts("kube-apiserver", cfg.DataDir+"/certs/kube-apiserver/secrets/service-network-serving-certkey",
		"tls.crt", "tls.key",
		[]string{"kube-apiserver", cfg.NodeIP, cfg.NodeName, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes",
			"localhost",
			apiServerServiceIP.String()}); err != nil {
		return err
	}
	if err := util.GenKeys(cfg.DataDir+"/resources/kube-apiserver/secrets/service-account-key",
		"service-account.crt", "service-account.key"); err != nil {
		return err
	}
	if err := util.GenCerts("system:masters", cfg.DataDir+"/certs/kube-apiserver/secrets/aggregator-client",
		"tls.crt", "tls.key",
		[]string{"system:admin", "system:masters"}); err != nil {
		return err
	}
	if err := util.GenCerts("system:masters", cfg.DataDir+"/resources/kube-apiserver/secrets/kubelet-client",
		"tls.crt", "tls.key",
		[]string{"kube-apiserver", "system:kube-apiserver", "system:masters"}); err != nil {
		return err
	}
	if err := util.GenKeys(cfg.DataDir+"/resources/kube-apiserver/sa-public-key",
		"serving-ca.pub", "serving-ca.key"); err != nil {
		return err
	}

	if err := util.GenCerts("kubelet", cfg.DataDir+"/resources/kubelet/secrets/kubelet-client",
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return err
	}

	// ocp
	if err := util.GenCerts("openshift-apiserver", cfg.DataDir+"/resources/openshift-apiserver/secrets",
		"tls.crt", "tls.key",
		[]string{"openshift-apiserver", cfg.NodeIP, cfg.NodeName, "openshift-apiserver.default.svc", "openshift-apiserver.default",
			"127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}); err != nil {
		return err
	}
	if err := util.GenCerts("openshift-controller-manager", cfg.DataDir+"/resources/openshift-controller-manager/secrets",
		"tls.crt", "tls.key",
		[]string{"openshift-controller-manager", cfg.NodeName, cfg.NodeIP, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default",
			"kubernetes", "localhost"}); err != nil {
		return err
	}
	if err := util.GenCerts("service-ca", cfg.DataDir+"/resources/service-ca/secrets/service-ca",
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName, apiServerServiceIP.String()}); err != nil {
		return err
	}
	if err := util.GenCerts("openshift-oauth-apiserver", cfg.DataDir+"/resources/openshift-oauth-apiserver/secrets",
		"tls.crt", "tls.key",
		[]string{"openshift-oauth-apiserver", cfg.NodeIP, cfg.NodeName, "127.0.0.1", "openshift-oauth-apiserver.default.svc",
			"openshift-oauth-apiserver.svc", "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}); err != nil {
		return err
	}
	return nil
}

func initServerConfig(cfg *config.MicroshiftConfig) error {
	return controllers.OpenShiftAPIServerConfig(cfg)
}

func initKubeconfig(cfg *config.MicroshiftConfig) error {
	if err := util.Kubeconfig(cfg.DataDir+"/resources/kubeadmin/kubeconfig", "system:admin", []string{"system:masters"}, cfg.Cluster.URL); err != nil {
		return err
	}
	if err := util.Kubeconfig(cfg.DataDir+"/resources/kube-apiserver/kubeconfig", "kube-apiserver", []string{"kube-apiserver", "system:kube-apiserver", "system:masters"}, cfg.Cluster.URL); err != nil {
		return err
	}
	if err := util.Kubeconfig(cfg.DataDir+"/resources/kube-controller-manager/kubeconfig", "system:kube-controller-manager", []string{"system:kube-controller-manager"}, cfg.Cluster.URL); err != nil {
		return err
	}
	if err := util.Kubeconfig(cfg.DataDir+"/resources/kube-scheduler/kubeconfig", "system:kube-scheduler", []string{"system:kube-scheduler"}, cfg.Cluster.URL); err != nil {
		return err
	}
	// per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
	if err := util.Kubeconfig(cfg.DataDir+"/resources/kubelet/kubeconfig", "system:node:"+cfg.NodeName, []string{"system:nodes"}, cfg.Cluster.URL); err != nil {
		return err
	}
	if err := util.Kubeconfig(cfg.DataDir+"/resources/kube-proxy/kubeconfig", "system:kube-proxy", []string{"system:nodes"}, cfg.Cluster.URL); err != nil {
		return err
	}
	return nil
}
