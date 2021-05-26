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
package controllers

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/openshift/microshift/pkg/constant"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	genericcontrollermanager "k8s.io/controller-manager/app"
	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
)

func KubeAPIServer() error {
	ip, err := util.GetHostIP()
	if err != nil {
		return fmt.Errorf("failed to get host IP: %v", err)
	}
	command := kubeapiserver.NewAPIServerCommand()
	apiArgs := []string{
		//"--openshift-config=/etc/kubernetes/ushift-resources/kube-apiserver/config/config.yaml", //TOOD
		//"--advertise-address=" + ip,
		//"-v=3",
		"--allow-privileged=true",
		"--anonymous-auth=false",
		"--audit-log-path=/var/log/kube-apiserver/audit.log",
		"--audit-policy-file=/etc/kubernetes/ushift-resources/kube-apiserver-audit-policies/default.yaml",
		"--api-audiences=https://kubernetes.svc",
		"--authorization-mode=Node,RBAC",
		"--bind-address=0.0.0.0",
		"--secure-port=6443",
		"--client-ca-file=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt",
		"--enable-admission-plugins=NodeRestriction",
		"--enable-aggregator-routing=true",
		"--etcd-cafile=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt",
		"--etcd-certfile=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client/tls.crt",
		"--etcd-keyfile=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client/tls.key",
		"--etcd-servers=https://127.0.0.1:2379",
		"--insecure-port=0",
		"--kubelet-certificate-authority=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt",
		"--kubelet-client-certificate=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/kubelet-client/tls.crt",
		"--kubelet-client-key=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/kubelet-client/tls.key",
		"--profiling=false",
		"--proxy-client-cert-file=/etc/kubernetes/ushift-certs/kube-apiserver/secrets/aggregator-client/tls.crt",
		"--proxy-client-key-file=/etc/kubernetes/ushift-certs/kube-apiserver/secrets/aggregator-client/tls.key",
		"--requestheader-allowed-names=aggregator,system:aggregator,openshift-apiserver,system:openshift-apiserver,kube-apiserver-proxy,system:kube-apiserver-proxy,openshift-aggregator,system:openshift-aggregator",
		"--requestheader-client-ca-file=/etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt",
		"--requestheader-extra-headers-prefix=X-Remote-Extra-",
		"--requestheader-group-headers=X-Remote-Group",
		"--requestheader-username-headers=X-Remote-User",
		"--service-account-issuer=https://kubernetes.svc",
		"--service-account-key-file=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/service-account-key/service-account.key",
		"--service-account-signing-key-file=/etc/kubernetes/ushift-resources/kube-apiserver/secrets/service-account-signing-key/service-account.key",
		"--service-cluster-ip-range=" + constant.ServiceCIDR,
		"--storage-backend=etcd3",
		"--tls-cert-file=/etc/kubernetes/ushift-certs/kube-apiserver/secrets/service-network-serving-certkey/tls.crt",
		"--tls-private-key-file=/etc/kubernetes/ushift-certs/kube-apiserver/secrets/service-network-serving-certkey/tls.key",
		"--cors-allowed-origins=/127.0.0.1(:[0-9]+)?$,/localhost(:[0-9]+)?$",
		"--log-file=/var/log/kube-apiserver.log",
		"--logtostderr=false",
		"-v=3",
	}
	if err := command.ParseFlags(apiArgs); err != nil {
		logrus.Fatalf("failed to parse flags:%v", err)
	}
	logrus.Infof("starting kube-apiserver %s, args: %v", ip, apiArgs)

	go func() {
		logrus.Fatalf("kube-apiserver exited: %v", command.RunE(command, nil))
	}()

	logrus.Info("waiting for kube-apiserver")

	restConfig, err := clientcmd.BuildConfigFromFlags("", constant.AdminKubeconfigPath)
	if err != nil {
		return err
	}

	versionedClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	err = genericcontrollermanager.WaitForAPIServer(versionedClient, 10*time.Second)
	if err != nil {
		logrus.Warningf("Failed to wait for apiserver being healthy: %v", err)
		return nil
	}
	logrus.Info("kube-apiserver is ready")

	return nil
}
