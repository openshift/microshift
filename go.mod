module github.com/openshift/microshift

go 1.16

replace (
	github.com/docker/distribution => github.com/openshift/docker-distribution v0.0.0-20180925154709-d4c35485a70d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5
	github.com/google/cadvisor => github.com/google/cadvisor v0.38.7
	github.com/moby/buildkit => github.com/dmcgowan/buildkit v0.0.0-20170731200553-da2b9dc7dab9
	github.com/onsi/ginkgo => github.com/openshift/ginkgo v4.5.0-origin.1+incompatible
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200224152610-e50cd9704f63
	k8s.io/api => k8s.io/api v0.20.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.0
	k8s.io/apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20210430024024-e3fdce453fbd //github.com/openshift/kubernetes-apiserver v0.0.0-20210222103016-a023730276fe
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.0
	k8s.io/client-go => k8s.io/client-go v0.20.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.0
	k8s.io/code-generator => k8s.io/code-generator v0.20.0
	k8s.io/component-base => k8s.io/component-base v0.20.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.0
	k8s.io/controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20210430024024-e3fdce453fbd
	k8s.io/cri-api => k8s.io/cri-api v0.20.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.0
	k8s.io/kube-aggregator => github.com/openshift/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20210430024024-e3fdce453fbd
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.0
	k8s.io/kubectl => k8s.io/kubectl v0.20.0
	k8s.io/kubelet => k8s.io/kubelet v0.20.0
	k8s.io/kubernetes => github.com/openshift/kubernetes v1.20.1-0.20210426184014-5feb30e1bd36
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.0
	k8s.io/metrics => k8s.io/metrics v0.20.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.0
)

require (
	github.com/auth0/go-jwt-middleware v1.0.1 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/containerd/containerd v1.4.8 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f
	github.com/go-ozzo/ozzo-validation v3.5.0+incompatible // indirect
	github.com/google/cadvisor v0.38.7 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/heketi/tests v0.0.0-20151005000721-f3775cbcefd6 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lpabon/godbc v0.1.1 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/openshift/api v0.0.0-20210416104836-15c7ae797f3f
	github.com/openshift/apiserver-library-go v0.0.0-20210426120049-59b0e972bfb7 // indirect
	github.com/openshift/build-machinery-go v0.0.0-20210423112049-9415d7ebd33e
	github.com/openshift/client-go v0.0.0-20210409155308-a8e62c60e930
	github.com/openshift/oauth-apiserver v0.0.0-20210622161626-6e0f92194d5a
	github.com/openshift/openshift-apiserver v0.0.0-alpha.0.0.20210324042139-a0f41548179e
	github.com/openshift/openshift-controller-manager v0.0.0-alpha.0.0.20210204125221-e0755cd0dca5
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20171018203845-0dec1b30a021 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/apiserver v0.21.0
	k8s.io/cli-runtime v0.20.4
	k8s.io/client-go v0.21.0
	k8s.io/component-base v0.21.0
	k8s.io/controller-manager v0.20.4
	k8s.io/csi-translation-lib v0.20.4 // indirect
	k8s.io/kube-aggregator v0.21.0
	k8s.io/kubectl v0.20.4
	k8s.io/kubernetes v1.21.0
	k8s.io/metrics v0.20.4 // indirect
	sigs.k8s.io/yaml v1.2.0
)
