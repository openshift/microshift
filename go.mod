module github.com/openshift/microshift

go 1.16

replace (
	github.com/docker/distribution => github.com/openshift/docker-distribution v0.0.0-20180925154709-d4c35485a70d
	github.com/docker/docker => github.com/openshift/moby-moby v0.0.0-20190308215630-da810a85109d
	github.com/moby/buildkit => github.com/dmcgowan/buildkit v0.0.0-20170731200553-da2b9dc7dab9
	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.4
	k8s.io/apiserver => github.com/openshift/kubernetes-apiserver v0.0.0-20210222103016-a023730276fe
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.4
	k8s.io/code-generator => k8s.io/code-generator v0.20.4
	k8s.io/component-base => k8s.io/component-base v0.20.4
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.4
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.4
	k8s.io/cri-api => k8s.io/cri-api v0.20.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.4
	k8s.io/kubectl => k8s.io/kubectl v0.20.4
	k8s.io/kubelet => k8s.io/kubelet v0.20.4
	k8s.io/kubernetes => github.com/openshift/kubernetes v1.20.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.4
	k8s.io/metrics => k8s.io/metrics v0.20.4
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.4
)

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/containers/image v3.0.2+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/fsouza/go-dockerclient v0.0.0-20171004212419-da3951ba2e9e // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/errors v0.19.2 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.1 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/jteeuwen/go-bindata v3.0.8-0.20151023091102-a0ff2567cfb7+incompatible // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6 // indirect
	github.com/openshift/api v0.0.0-20210412200117-f60a0b2883ea // indirect
	github.com/openshift/apiserver-library-go v0.0.0-20210426120049-59b0e972bfb7 // indirect
	github.com/openshift/build-machinery-go v0.0.0-20210423112049-9415d7ebd33e // indirect
	github.com/openshift/client-go v0.0.0-20210331195552-cf6c2669e01f // indirect
	github.com/openshift/library-go v0.0.0-20210331235027-66936e2fcc52 // indirect
	github.com/openshift/openshift-apiserver v0.0.0-alpha.0.0.20210324042139-a0f41548179e
	github.com/openshift/openshift-controller-manager v0.0.0-alpha.0.0.20210204125221-e0755cd0dca5
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.21.0-rc.0 // indirect
	k8s.io/apiextensions-apiserver v0.21.0-rc.0 // indirect
	k8s.io/apimachinery v0.21.0-rc.0 // indirect
	k8s.io/apiserver v0.21.0-rc.0 // indirect
	k8s.io/client-go v0.21.0-rc.0 // indirect
	k8s.io/cloud-provider v0.20.4 // indirect
	k8s.io/code-generator v0.21.0-rc.0 // indirect
	k8s.io/component-base v0.21.0-rc.0
	k8s.io/component-helpers v0.20.4 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/kube-aggregator v0.21.0-rc.0 // indirect
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd // indirect
	k8s.io/kubectl v0.20.4 // indirect
	k8s.io/kubernetes v1.21.0-rc.0
)
