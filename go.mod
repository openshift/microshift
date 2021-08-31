module github.com/openshift/microshift

go 1.16

replace (
	github.com/docker/distribution => github.com/openshift/docker-distribution v0.0.0-20180925154709-d4c35485a70d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/golang/protobuf => github.com/golang/protobuf v1.4.3
	github.com/google/cadvisor => github.com/google/cadvisor v0.38.7
	github.com/moby/buildkit => github.com/dmcgowan/buildkit v0.0.0-20170731200553-da2b9dc7dab9
	github.com/onsi/ginkgo => github.com/openshift/ginkgo v4.5.0-origin.1+incompatible
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a
	k8s.io/api => k8s.io/api v0.20.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.9
	// k8s.io/apiserver => k8s.io/apiserver v0.20.9
	k8s.io/apiserver => github.com/openshift/kubernetes-apiserver v0.0.0-20210222103016-a023730276fe // points to openshift-apiserver-4.7-kubernetes-1.20.4 branch
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.9
	k8s.io/client-go => k8s.io/client-go v0.20.9
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.9
	k8s.io/code-generator => k8s.io/code-generator v0.20.9
	k8s.io/component-base => k8s.io/component-base v0.20.9
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.9
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.9
	k8s.io/cri-api => k8s.io/cri-api v0.20.9
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.9
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.9
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.9
	k8s.io/kubectl => k8s.io/kubectl v0.20.9
	k8s.io/kubelet => k8s.io/kubelet v0.20.9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.9
	k8s.io/metrics => k8s.io/metrics v0.20.9
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.9
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.9
)

require (
	github.com/auth0/go-jwt-middleware v1.0.1 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/openshift/api v0.0.0-20210428205234-a8389931bee7
	github.com/openshift/build-machinery-go v0.0.0-20210209125900-0da259a2c359
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/openshift/openshift-apiserver v0.0.0-alpha.0.0.20210618182237-f9ac08715d1b
	github.com/openshift/openshift-controller-manager v0.0.0-alpha.0.0.20210730000055-c93745bf6898
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.20.9
	k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery v0.20.9
	k8s.io/cli-runtime v0.20.9
	k8s.io/client-go v0.20.9
	k8s.io/component-base v0.20.9
	k8s.io/controller-manager v0.20.9
	k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kubectl v0.20.9
	k8s.io/kubernetes v1.20.9
	sigs.k8s.io/yaml v1.2.0
)
