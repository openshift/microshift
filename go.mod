module github.com/openshift/microshift

go 1.24.0

require (
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // openshift-controller-manager
	github.com/google/go-cmp v0.7.0
	github.com/miekg/dns v1.1.63 // microshift
	github.com/openshift/api v0.0.0-20250710004639-926605d3338b
	github.com/openshift/build-machinery-go v0.0.0-20250602125535-1b6d00b8c37c
	github.com/openshift/client-go v0.0.0-20250710075018-396b36f983ee
	github.com/openshift/library-go v0.0.0-20250710130336-73c7662bc565
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.10.0
	github.com/vishvananda/netlink v1.3.1-0.20250206174618-62fb240731fa
	go.etcd.io/etcd/client/pkg/v3 v3.5.21
	go.etcd.io/etcd/client/v3 v3.5.21
	golang.org/x/sys v0.31.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/fsnotify/fsnotify v1.8.0
	github.com/go-kit/kit v0.9.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/snappy v0.0.4
	github.com/openshift/cluster-policy-controller v0.0.0-20250310152427-748524784686
	github.com/openshift/route-controller-manager v0.0.0-20250709131101-e148fabc13f7
	github.com/prometheus/client_model v0.6.1
	github.com/prometheus/common v0.62.0
	github.com/prometheus/prometheus v0.302.1
	github.com/squat/generic-device-plugin v0.0.0-20250710162141-0f7fddf166f1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v1.33.2
	k8s.io/apiextensions-apiserver v1.33.2
	k8s.io/apimachinery v1.33.2
	k8s.io/apiserver v1.33.2
	k8s.io/cli-runtime v1.33.2
	k8s.io/client-go v1.33.2
	k8s.io/cloud-provider v1.33.2
	k8s.io/component-base v1.33.2
	k8s.io/kube-aggregator v1.33.2
	k8s.io/kubectl v1.33.2
	k8s.io/kubelet v1.33.2
	k8s.io/utils v0.0.0-20241210054802-24370beab758
	sigs.k8s.io/kube-storage-version-migrator v0.0.6-0.20230721195810-5c8923c5ff96
	sigs.k8s.io/kustomize/api v0.19.0
	sigs.k8s.io/kustomize/kyaml v0.19.0
)

require (
	cel.dev/expr v0.19.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20211209120228-48547f28849e // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hnslib v0.1.1 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/armon/circbuf v0.0.0-20190214190532-5111143e8da2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/container-storage-interface/spec v1.9.0 // indirect
	github.com/containerd/containerd/api v1.8.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/ttrpc v1.2.6 // indirect
	github.com/containerd/typeurl/v2 v2.2.2 // indirect
	github.com/coreos/go-oidc v2.3.0+incompatible // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/fgprof v0.9.4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.4 // indirect
	github.com/go-ldap/ldap/v3 v3.4.3 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cadvisor v0.52.1 // indirect
	github.com/google/cel-go v0.23.2 // indirect
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/opencontainers/cgroups v0.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.2.5 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/opencontainers/selinux v1.11.1 // indirect
	github.com/openshift/apiserver-library-go v0.0.0-20250710132015-f0d44ef6e53b // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.42.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	k8s.io/cluster-bootstrap v1.33.2 // indirect
	k8s.io/component-helpers v1.33.2 // indirect
	k8s.io/controller-manager v1.33.2 // indirect
	k8s.io/cri-api v1.33.2 // indirect
	k8s.io/cri-client v1.33.2 // indirect
	k8s.io/csi-translation-lib v1.33.2 // indirect
	k8s.io/dynamic-resource-allocation v1.33.2 // indirect
	k8s.io/endpointslice v1.33.2 // indirect
	k8s.io/externaljwt v1.33.2 // indirect
	k8s.io/kms v1.33.2 // indirect
	k8s.io/kube-controller-manager v1.33.2 // indirect
	k8s.io/kube-scheduler v1.33.2 // indirect
	k8s.io/metrics v1.33.2 // indirect
	k8s.io/mount-utils v1.33.2 // indirect
	k8s.io/pod-security-admission v1.33.2 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.31.2 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/grafana/regexp v0.0.0-20240518133315-a468a5bfb3bc // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/profile v1.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.21 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	k8s.io/gengo/v2 v2.0.0-20250207200755-1244d31929d7 // indirect
	k8s.io/klog/v2 v2.130.1
	k8s.io/kubernetes v1.33.2
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
)

replace (
	github.com/onsi/ginkgo/v2 => github.com/openshift/onsi-ginkgo/v2 v2.6.1-0.20250416174521-4eb003743b54 // from kubernetes
	github.com/openshift/cluster-policy-controller => ./deps/github.com/openshift/cluster-policy-controller // deps copy
	github.com/openshift/route-controller-manager => ./deps/github.com/openshift/route-controller-manager // deps copy
	k8s.io/klog/v2 => ./deps/k8s.io/klog // deps clone github.com/kubernetes/klog from kubernetes
	k8s.io/kubernetes => ./deps/github.com/openshift/kubernetes // deps copy
	sigs.k8s.io/kube-storage-version-migrator => github.com/openshift/kubernetes-kube-storage-version-migrator v0.0.3-0.20250719041745-caf798560ad3 // release kube-storage-version-migrator via kubernetes-kube-storage-version-migrator
)

replace (
	k8s.io/api => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/api // deps kubernetes-version
	k8s.io/apiextensions-apiserver => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/apiextensions-apiserver // deps kubernetes-version
	k8s.io/apimachinery => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/apimachinery // deps kubernetes-version
	k8s.io/apiserver => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/apiserver // deps kubernetes-version
	k8s.io/cli-runtime => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/cli-runtime // deps kubernetes-version
	k8s.io/client-go => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/client-go // deps kubernetes-version
	k8s.io/cloud-provider => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/cloud-provider // deps kubernetes-version
	k8s.io/cluster-bootstrap => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/cluster-bootstrap // deps kubernetes-version
	k8s.io/code-generator => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/code-generator // deps kubernetes-version
	k8s.io/component-base => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/component-base // deps kubernetes-version
	k8s.io/component-helpers => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/component-helpers // deps kubernetes-version
	k8s.io/controller-manager => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/controller-manager // deps kubernetes-version
	k8s.io/cri-api => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/cri-api // deps kubernetes-version
	k8s.io/cri-client => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/cri-client // deps kubernetes-version
	k8s.io/csi-translation-lib => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/csi-translation-lib // deps kubernetes-version
	k8s.io/dynamic-resource-allocation => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/dynamic-resource-allocation // deps kubernetes-version
	k8s.io/endpointslice => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/endpointslice // deps kubernetes-version
	k8s.io/externaljwt => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/externaljwt // deps kubernetes-version
	k8s.io/kms => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kms // deps kubernetes-version
	k8s.io/kube-aggregator => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-aggregator // deps kubernetes-version
	k8s.io/kube-controller-manager => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-controller-manager // deps kubernetes-version
	k8s.io/kube-proxy => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-proxy // deps kubernetes-version
	k8s.io/kube-scheduler => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-scheduler // deps kubernetes-version
	k8s.io/kubectl => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kubectl // deps kubernetes-version
	k8s.io/kubelet => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/kubelet // deps kubernetes-version
	k8s.io/metrics => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/metrics // deps kubernetes-version
	k8s.io/mount-utils => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/mount-utils // deps kubernetes-version
	k8s.io/pod-security-admission => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/pod-security-admission // deps kubernetes-version
	k8s.io/sample-apiserver => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/sample-apiserver // deps kubernetes-version
	k8s.io/sample-cli-plugin => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/sample-cli-plugin // deps kubernetes-version
	k8s.io/sample-controller => ./deps/github.com/openshift/kubernetes/staging/src/k8s.io/sample-controller // deps kubernetes-version
)

replace (
	go.etcd.io/etcd/api/v3 => github.com/openshift/etcd/api/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/client/pkg/v3 => github.com/openshift/etcd/client/pkg/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/client/v3 => github.com/openshift/etcd/client/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
)
