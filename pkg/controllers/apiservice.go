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
	"context"
	"io/ioutil"
	"strings"

	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/util/intstr"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	apiregistrationclientv1 "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/typed/apiregistration/v1"
)

func createAPIHeadlessSvc(cfg *config.MicroshiftConfig, svcName string, svcPort int) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}

	client := coreclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "core-agent"))
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(443),
				},
			},
		},
	}
	_, err = client.Services("default").Get(context.TODO(), svc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		klog.Infof("Creating service %s", svc.Name)
		_, err = client.Services("default").Create(context.TODO(), svc, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		endpoints := &corev1.Endpoints{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Endpoints",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: "default",
			},
		}

		k8s_endpoints, err := client.Endpoints("default").Get(context.TODO(), "kubernetes", metav1.GetOptions{})
		if err != nil {
			klog.Infof("Failed to find kubernetes endpoints")
		}
		subsets := endpoints.Subsets
		for _, sub := range k8s_endpoints.Subsets {
			addr := sub.Addresses
			ports := []corev1.EndpointPort{
				{
					Port: int32(svcPort),
				},
			}
			subsets = append(subsets,
				corev1.EndpointSubset{
					Addresses: addr,
					Ports:     ports,
				})
		}
		endpoints.Subsets = subsets
		_, err = client.Endpoints("default").Get(context.TODO(), endpoints.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			klog.Infof("Creating endpoints %s", endpoints.Name)
			_, err = client.Endpoints("default").Create(context.TODO(), endpoints, metav1.CreateOptions{})
			return err
		}
	}
	return nil
}
func trimFirst(s string, sep string) string {
	parts := strings.Split(s, sep)
	return strings.Join(parts[1:], sep)
}

func createAPIRegistration(cfg *config.MicroshiftConfig) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}
	caFile, err := ioutil.ReadFile(cfg.DataDir + "/certs/ca-bundle/ca-bundle.crt")
	if err != nil {
		klog.Errorf("Error loading CA bundle certificate %v", err)
	}
	client := apiregistrationclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "apiregistration-agent"))
	for _, apiSvc := range []string{
		"v1.apps.openshift.io",
		"v1.authorization.openshift.io",
		"v1.build.openshift.io",
		"v1.image.openshift.io",
		//"v1.oauth.openshift.io", //TODO check if they exist
		"v1.project.openshift.io",
		"v1.quota.openshift.io",
		"v1.route.openshift.io",
		"v1.security.openshift.io",
		"v1.template.openshift.io", //TODO missing templateinstances
	} {
		api := &apiregistrationv1.APIService{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIService",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: apiSvc,
			},
			Spec: apiregistrationv1.APIServiceSpec{
				Service: &apiregistrationv1.ServiceReference{
					Name:      "openshift-apiserver",
					Namespace: "default",
				},
				Group:                trimFirst(apiSvc, "."),
				GroupPriorityMinimum: 9900,
				Version:              "v1",
				CABundle:             caFile,
				VersionPriority:      15,
			},
		}
		_, err = client.APIServices().Get(context.TODO(), api.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			klog.Infof("Creating api registration %s", api.Name)
			_, _ = client.APIServices().Create(context.TODO(), api, metav1.CreateOptions{})
		}
	}

	for _, oauthApiSvc := range []string{
		"v1.user.openshift.io",
	} {
		oauthApi := &apiregistrationv1.APIService{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIService",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: oauthApiSvc,
			},
			Spec: apiregistrationv1.APIServiceSpec{
				Service: &apiregistrationv1.ServiceReference{
					Name:      "openshift-oauth-apiserver",
					Namespace: "default",
				},
				Group:                trimFirst(oauthApiSvc, "."),
				GroupPriorityMinimum: 9900,
				Version:              "v1",
				CABundle:             caFile,
				VersionPriority:      15,
			},
		}
		_, err = client.APIServices().Get(context.TODO(), oauthApi.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			klog.Infof("creating api registration %s", oauthApi.Name)
			_, _ = client.APIServices().Create(context.TODO(), oauthApi, metav1.CreateOptions{})
		}
	}
	return nil
}

func ApplyDefaultSCCs(cfg *config.MicroshiftConfig) error {
	kubeconfigPath := cfg.DataDir + "/resources/kubeadmin/kubeconfig"
	var (
		sccs = []string{
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-anyuid.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostaccess.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostmount-anyuid.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-hostnetwork.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-nonroot.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-privileged.yaml",
			"assets/scc/0000_20_kube-apiserver-operator_00_scc-restricted.yaml",
		}
	)
	if err := assets.ApplySCCs(sccs, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sccs %v", err)
		return err
	}
	return nil
}
