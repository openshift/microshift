module github.com/openshift/microshift

go 1.21

require (
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // openshift-controller-manager
	github.com/google/go-cmp v0.6.0
	github.com/miekg/dns v1.1.35 // microshift
	github.com/openshift/api v0.0.0-20231218131639-7a5aa77cc72d
	github.com/openshift/build-machinery-go v0.0.0-20230824093055-6a18da01283c
	github.com/openshift/client-go v0.0.0-20231218155125-ff7d9f9bf415
	github.com/openshift/cluster-policy-controller v0.0.0-20240125161249-15c068ab02d7
	github.com/openshift/library-go v0.0.0-20231218143352-99cedb2a141c
	github.com/openshift/route-controller-manager v0.0.0-20240105122523-95aa561e9387
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	github.com/vishvananda/netlink v1.1.0
	go.etcd.io/etcd/client/pkg/v3 v3.5.11
	go.etcd.io/etcd/client/v3 v3.5.10
	golang.org/x/sys v0.15.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.29.0
	k8s.io/apiextensions-apiserver v0.29.0
	k8s.io/apimachinery v0.29.0
	k8s.io/apiserver v0.29.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v0.29.0
	k8s.io/cloud-provider v0.28.3
	k8s.io/component-base v0.29.0
	k8s.io/klog/v2 v2.110.1
	k8s.io/kube-aggregator v0.29.0
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.29.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	sigs.k8s.io/kube-storage-version-migrator v0.0.6-0.20230721195810-5c8923c5ff96
	sigs.k8s.io/kustomize/api v0.13.5-0.20230601165947-6ce0bf390ce3
)

require github.com/google/s2a-go v0.1.7 // indirect

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/mocks v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20211209120228-48547f28849e // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v1.18.1-0.20220218231025-f11817397a1b // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Microsoft/hcsshim v0.8.25 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/container-storage-interface/spec v1.8.0 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/ttrpc v1.2.2 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.4 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-ldap/ldap/v3 v3.4.3 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/cadvisor v0.48.1 // indirect
	github.com/google/cel-go v0.17.7 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/mrunalp/fileutils v0.5.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.1.10 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220909204839-494a5a6aca78 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/openshift/apiserver-library-go v0.0.0-20231218150122-47b436d2f389 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/profile v1.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rubiojr/go-vhd v0.0.0-20200706105327-02e210299021 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/seccomp/libseccomp-golang v0.10.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/vmware/govmomi v0.30.6 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.11 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.42.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.42.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.44.0 // indirect
	go.opentelemetry.io/otel v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.19.0 // indirect
	go.opentelemetry.io/otel/sdk v1.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.starlark.net v0.0.0-20230525235612-a134d8f9ddca // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.19.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/oauth2 v0.11.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/term v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	google.golang.org/api v0.126.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/gcfg.v1 v1.2.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/cluster-bootstrap v0.0.0 // indirect
	k8s.io/component-helpers v0.29.0 // indirect
	k8s.io/controller-manager v0.28.3 // indirect
	k8s.io/cri-api v0.0.0 // indirect
	k8s.io/csi-translation-lib v0.0.0 // indirect
	k8s.io/dynamic-resource-allocation v0.0.0 // indirect
	k8s.io/endpointslice v0.0.0 // indirect
	k8s.io/gengo v0.0.0-20230829151522-9cce18d56c01 // indirect
	k8s.io/kms v0.29.0 // indirect
	k8s.io/kube-controller-manager v0.0.0 // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/kubelet v0.28.3 // indirect
	k8s.io/legacy-cloud-providers v0.0.0 // indirect
	k8s.io/metrics v0.0.0 // indirect
	k8s.io/mount-utils v0.0.0 // indirect
	k8s.io/pod-security-admission v0.28.3 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.28.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/kyaml v0.14.3-0.20230601165947-6ce0bf390ce3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace (
	github.com/onsi/ginkgo/v2 => github.com/openshift/ginkgo/v2 v2.6.1-0.20231031162821-c5e24be53ea7 // from kubernetes
	k8s.io/api => github.com/openshift/kubernetes/staging/src/k8s.io/api v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/apiextensions-apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/apimachinery => github.com/openshift/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/cli-runtime => github.com/openshift/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/client-go => github.com/openshift/kubernetes/staging/src/k8s.io/client-go v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/cloud-provider => github.com/openshift/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/cluster-bootstrap => github.com/openshift/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/code-generator => github.com/openshift/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/component-base => github.com/openshift/kubernetes/staging/src/k8s.io/component-base v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/component-helpers => github.com/openshift/kubernetes/staging/src/k8s.io/component-helpers v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/cri-api => github.com/openshift/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/csi-translation-lib => github.com/openshift/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/dynamic-resource-allocation => github.com/openshift/kubernetes/staging/src/k8s.io/dynamic-resource-allocation v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/kms => github.com/openshift/kubernetes/staging/src/k8s.io/kms v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/kube-aggregator => github.com/openshift/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
	k8s.io/kube-controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/kube-proxy => github.com/openshift/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/kube-scheduler => github.com/openshift/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/kubectl => github.com/openshift/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/kubelet => github.com/openshift/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/kubernetes => github.com/openshift/kubernetes v0.0.0-20240126174038-5a4819c6048e // release kubernetes
	k8s.io/legacy-cloud-providers => github.com/openshift/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/metrics => github.com/openshift/kubernetes/staging/src/k8s.io/metrics v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/mount-utils => github.com/openshift/kubernetes/staging/src/k8s.io/mount-utils v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/pod-security-admission => github.com/openshift/kubernetes/staging/src/k8s.io/pod-security-admission v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/sample-apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/sample-cli-plugin => github.com/openshift/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20240126174038-5a4819c6048e // from kubernetes
	k8s.io/sample-controller => github.com/openshift/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20240126174038-5a4819c6048e // from kubernetes
)

replace (
	go.etcd.io/etcd/api/v3 => github.com/openshift/etcd/api/v3 v3.5.1-0.20240105153547-66af0e5d532b // from etcd
	go.etcd.io/etcd/client/pkg/v3 => github.com/openshift/etcd/client/pkg/v3 v3.5.1-0.20240105153547-66af0e5d532b // from etcd
	go.etcd.io/etcd/client/v3 => github.com/openshift/etcd/client/v3 v3.5.1-0.20240105153547-66af0e5d532b // from etcd
)

replace sigs.k8s.io/kube-storage-version-migrator => github.com/openshift/kubernetes-kube-storage-version-migrator v0.0.3-0.20240125051406-969a60e9e246 // release kube-storage-version-migrator via kubernetes-kube-storage-version-migrator

replace k8s.io/endpointslice => github.com/openshift/kubernetes/staging/src/k8s.io/endpointslice v0.0.0-20240126174038-5a4819c6048e // staging kubernetes
