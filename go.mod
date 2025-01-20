module github.com/openshift/microshift

go 1.23.0

toolchain go1.23.2

require (
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // openshift-controller-manager
	github.com/google/go-cmp v0.6.0
	github.com/miekg/dns v1.1.35 // microshift
	github.com/openshift/api v0.0.0-20241004095111-b1f700bdd8d2
	github.com/openshift/build-machinery-go v0.0.0-20240910153727-5725581bdf8f
	github.com/openshift/client-go v0.0.0-20241001162912-da6d55e4611f
	github.com/openshift/cluster-policy-controller v0.0.0-20241210163452-327d3cbd82fd
	github.com/openshift/library-go v0.0.0-20241107160307-0064ad7bd060
	github.com/openshift/route-controller-manager v0.0.0-20241218160919-bc97534a12a7
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/vishvananda/netlink v1.1.0
	go.etcd.io/etcd/client/pkg/v3 v3.5.17
	go.etcd.io/etcd/client/v3 v3.5.14
	golang.org/x/sys v0.28.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.31.1
	k8s.io/apiextensions-apiserver v0.31.1
	k8s.io/apimachinery v0.31.1
	k8s.io/apiserver v0.31.1
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v0.31.1
	k8s.io/cloud-provider v0.30.1
	k8s.io/component-base v0.31.1
	k8s.io/klog/v2 v2.130.1
	k8s.io/kube-aggregator v0.31.1
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.31.1
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/fsnotify/fsnotify v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/utils v0.0.0-20240921022957-49e7df575cb6
	sigs.k8s.io/kube-storage-version-migrator v0.0.6-0.20230721195810-5c8923c5ff96
	sigs.k8s.io/kustomize/api v0.17.2
	sigs.k8s.io/kustomize/kyaml v0.17.1
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/felixge/fgprof v0.9.4 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/google/pprof v0.0.0-20240827171923-fa2c70bbbfe5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20211209120228-48547f28849e // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Microsoft/hcsshim v0.8.26 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/cilium/ebpf v0.9.1 // indirect
	github.com/container-storage-interface/spec v1.9.0 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/ttrpc v1.2.2 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.4 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-ldap/ldap/v3 v3.4.3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/cadvisor v0.49.0 // indirect
	github.com/google/cel-go v0.20.1 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.4.0 // indirect
	github.com/moby/sys/mountinfo v0.7.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/mrunalp/fileutils v0.5.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.1.13 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20220909204839-494a5a6aca78 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/openshift/apiserver-library-go v0.0.0-20241001175710-6064b62894a6 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/profile v1.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/seccomp/libseccomp-golang v0.10.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.17 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful v0.42.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/sdk v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.starlark.net v0.0.0-20230525235612-a134d8f9ddca // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240930140551-af27646dc61f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240924160255-9d4c2d233b61 // indirect
	google.golang.org/grpc v1.67.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	k8s.io/cluster-bootstrap v0.0.0 // indirect
	k8s.io/component-helpers v0.31.1 // indirect
	k8s.io/controller-manager v0.31.1 // indirect
	k8s.io/cri-api v0.0.0 // indirect
	k8s.io/cri-client v0.0.0 // indirect
	k8s.io/csi-translation-lib v0.0.0 // indirect
	k8s.io/dynamic-resource-allocation v0.0.0 // indirect
	k8s.io/endpointslice v0.0.0 // indirect
	k8s.io/gengo/v2 v2.0.0-20240228010128-51d4e06bde70 // indirect
	k8s.io/kms v0.31.1 // indirect
	k8s.io/kube-controller-manager v0.0.0 // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/kubelet v0.30.1 // indirect
	k8s.io/metrics v0.0.0 // indirect
	k8s.io/mount-utils v0.0.0 // indirect
	k8s.io/pod-security-admission v0.31.1 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.30.3 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace (
	github.com/onsi/ginkgo/v2 => github.com/openshift/onsi-ginkgo/v2 v2.6.1-0.20241008152707-25bf9f14db44 // from kubernetes
	k8s.io/api => github.com/openshift/kubernetes/staging/src/k8s.io/api v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/apiextensions-apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/apimachinery => github.com/openshift/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/cli-runtime => github.com/openshift/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/client-go => github.com/openshift/kubernetes/staging/src/k8s.io/client-go v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/cloud-provider => github.com/openshift/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/cluster-bootstrap => github.com/openshift/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/code-generator => github.com/openshift/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/component-base => github.com/openshift/kubernetes/staging/src/k8s.io/component-base v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/component-helpers => github.com/openshift/kubernetes/staging/src/k8s.io/component-helpers v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/cri-api => github.com/openshift/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/csi-translation-lib => github.com/openshift/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/dynamic-resource-allocation => github.com/openshift/kubernetes/staging/src/k8s.io/dynamic-resource-allocation v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/kms => github.com/openshift/kubernetes/staging/src/k8s.io/kms v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/kube-aggregator => github.com/openshift/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
	k8s.io/kube-controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/kube-proxy => github.com/openshift/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/kube-scheduler => github.com/openshift/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/kubectl => github.com/openshift/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/kubelet => github.com/openshift/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/kubernetes => github.com/openshift/kubernetes v0.0.0-20250115192950-4b2db1ec33fa // release kubernetes
	k8s.io/metrics => github.com/openshift/kubernetes/staging/src/k8s.io/metrics v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/mount-utils => github.com/openshift/kubernetes/staging/src/k8s.io/mount-utils v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/pod-security-admission => github.com/openshift/kubernetes/staging/src/k8s.io/pod-security-admission v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/sample-apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/sample-cli-plugin => github.com/openshift/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
	k8s.io/sample-controller => github.com/openshift/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20250115192950-4b2db1ec33fa // from kubernetes
)

replace (
	go.etcd.io/etcd/api/v3 => github.com/openshift/etcd/api/v3 v3.5.1-0.20241118170732-914783cae63a // from etcd
	go.etcd.io/etcd/client/pkg/v3 => github.com/openshift/etcd/client/pkg/v3 v3.5.1-0.20241118170732-914783cae63a // from etcd
	go.etcd.io/etcd/client/v3 => github.com/openshift/etcd/client/v3 v3.5.1-0.20241118170732-914783cae63a // from etcd
)

replace sigs.k8s.io/kube-storage-version-migrator => github.com/openshift/kubernetes-kube-storage-version-migrator v0.0.3-0.20250108152203-4c521b09f4dc // release kube-storage-version-migrator via kubernetes-kube-storage-version-migrator

replace k8s.io/endpointslice => github.com/openshift/kubernetes/staging/src/k8s.io/endpointslice v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes

replace k8s.io/cri-client => github.com/openshift/kubernetes/staging/src/k8s.io/cri-client v0.0.0-20250115192950-4b2db1ec33fa // staging kubernetes
