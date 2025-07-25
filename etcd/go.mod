module github.com/openshift/microshift/etcd

go 1.24.0

require github.com/openshift/microshift v0.0.0

replace (
	github.com/openshift/microshift => ../
	github.com/openshift/microshift/pkg/config => ../pkg/config
	github.com/openshift/microshift/pkg/util/cryptomaterial => ../pkg/util/cryptomaterial
)

require (
	github.com/openshift/api v0.0.0-20250710004639-926605d3338b
	github.com/openshift/build-machinery-go v0.0.0-20250602125535-1b6d00b8c37c
	github.com/spf13/cobra v1.8.1
	go.etcd.io/etcd/server/v3 v3.5.21
	k8s.io/apimachinery v1.33.2
	k8s.io/cli-runtime v1.33.2
	k8s.io/component-base v1.33.2
	k8s.io/klog/v2 v2.130.1
	k8s.io/kubectl v1.33.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/openshift/library-go v0.0.0-20250710130336-73c7662bc565 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/squat/generic-device-plugin v0.0.0-20250710162141-0f7fddf166f1 // indirect
	github.com/vishvananda/netlink v1.3.1-0.20250206174618-62fb240731fa // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	k8s.io/apiserver v1.33.2 // indirect
	k8s.io/kubelet v1.33.2 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/xiang90/probing v0.0.0-20221125231312-a49e3df8f510 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.etcd.io/bbolt v1.3.11 // indirect
	go.etcd.io/etcd/api/v3 v3.5.21
	go.etcd.io/etcd/client/pkg/v3 v3.5.21 // indirect
	go.etcd.io/etcd/client/v2 v2.305.21 // indirect
	go.etcd.io/etcd/client/v3 v3.5.21 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.21 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.21 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v1.33.2 // indirect
	k8s.io/client-go v1.33.2 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	k8s.io/utils v0.0.0-20241210054802-24370beab758 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/kustomize/api v0.19.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.19.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
)

replace (
	github.com/onsi/ginkgo/v2 => github.com/openshift/onsi-ginkgo/v2 v2.6.1-0.20250416174521-4eb003743b54 // from kubernetes
	go.etcd.io/etcd/api/v3 => github.com/openshift/etcd/api/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/client/pkg/v3 => github.com/openshift/etcd/client/pkg/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/client/v3 => github.com/openshift/etcd/client/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/pkg/v3 => github.com/openshift/etcd/pkg/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/raft/v3 => github.com/openshift/etcd/raft/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
	go.etcd.io/etcd/server/v3 => github.com/openshift/etcd/server/v3 v3.5.1-0.20250411172207-a5421dfe551a // from etcd
)

replace (
	k8s.io/api => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/api
	k8s.io/apiextensions-apiserver => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/apiextensions-apiserver
	k8s.io/apimachinery => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/apimachinery
	k8s.io/apiserver => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/apiserver
	k8s.io/cli-runtime => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/cli-runtime
	k8s.io/client-go => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/client-go
	k8s.io/cloud-provider => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/cloud-provider
	k8s.io/cluster-bootstrap => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/cluster-bootstrap
	k8s.io/code-generator => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/code-generator
	k8s.io/component-base => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/component-base
	k8s.io/component-helpers => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/component-helpers
	k8s.io/controller-manager => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/controller-manager
	k8s.io/cri-api => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/cri-api
	k8s.io/cri-client => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/cri-client
	k8s.io/csi-translation-lib => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/csi-translation-lib
	k8s.io/dynamic-resource-allocation => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/dynamic-resource-allocation
	k8s.io/endpointslice => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/endpointslice
	k8s.io/externaljwt => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/externaljwt
	k8s.io/kms => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kms
	k8s.io/kube-aggregator => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-aggregator
	k8s.io/kube-controller-manager => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-controller-manager
	k8s.io/kube-proxy => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-proxy
	k8s.io/kube-scheduler => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kube-scheduler
	k8s.io/kubectl => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kubectl
	k8s.io/kubelet => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/kubelet
	k8s.io/legacy-cloud-providers => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/legacy-cloud-providers
	k8s.io/metrics => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/metrics
	k8s.io/mount-utils => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/mount-utils
	k8s.io/pod-security-admission => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/pod-security-admission
	k8s.io/sample-apiserver => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/sample-apiserver
	k8s.io/sample-cli-plugin => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/sample-cli-plugin
	k8s.io/sample-controller => ../deps/github.com/openshift/kubernetes/staging/src/k8s.io/sample-controller
)
