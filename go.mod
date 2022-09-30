module github.com/openshift/microshift

go 1.17

require (
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // openshift-controller-manager
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/kelseyhightower/envconfig v1.4.0 // microshift
	github.com/miekg/dns v1.1.35 // microshift
	github.com/mitchellh/go-homedir v1.1.0 // microshift
	github.com/openshift/api v0.0.0-20220616165336-689617d54300
	github.com/openshift/build-machinery-go v0.0.0-20220429084610-baff9f8d23b3
	github.com/openshift/client-go v0.0.0-20220603133046-984ee5ebedcf
	github.com/openshift/cluster-policy-controller v0.0.0-20220610112634-c7201edacb6e
	github.com/openshift/library-go v0.0.0-20220615161831-8b2df431789c
	github.com/openshift/openshift-controller-manager v0.0.0-alpha.0.0.20220825163933-911da5750170
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.etcd.io/etcd/server/v3 v3.5.3
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.24.0
	k8s.io/apiextensions-apiserver v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/apiserver v0.24.0
	k8s.io/cli-runtime v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/component-base v0.24.0
	k8s.io/controller-manager v0.24.0 // indirect
	k8s.io/klog/v2 v2.60.1
	k8s.io/kube-aggregator v0.24.0
	k8s.io/kube-openapi v0.0.0-20220803164354-a70c9af30aea
	k8s.io/kubectl v0.24.0
	k8s.io/kubernetes v1.24.0
	sigs.k8s.io/yaml v1.2.0
)

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/Azure/azure-sdk-for-go v55.0.0+incompatible // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/mocks v0.4.1 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.1.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v1.16.1-0.20210702024009-ea6160c1d0e3 // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/Microsoft/hcsshim v0.8.22 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20210826220005-b48c857c3a0e // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/aws/aws-sdk-go v1.38.49 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chai2010/gettext-go v0.0.0-20160711120539-c6fed771bfd5 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/cilium/ebpf v0.7.0 // indirect
	github.com/clusterhq/flocker-go v0.0.0-20160920122132-2b8b7259d313 // indirect
	github.com/container-storage-interface/spec v1.5.0 // indirect
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/containerd v1.5.0-beta.1 // indirect
	github.com/containerd/ttrpc v1.0.2 // indirect
	github.com/containers/image v3.0.2+incompatible // indirect
	github.com/containers/storage v1.28.1 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.12+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/felixge/httpsnoop v1.0.1 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fvbommel/sortorder v1.0.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-ozzo/ozzo-validation v3.5.0+incompatible // indirect
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/cadvisor v0.44.1 // indirect
	github.com/google/cel-go v0.10.1 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/heketi/heketi v10.3.0+incompatible // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mindprince/gonvml v0.0.0-20190828220739-9ebdce4bb989 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/mrunalp/fileutils v0.5.0 // indirect
	github.com/mtrmac/gpgme v0.1.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417 // indirect
	github.com/opencontainers/selinux v1.10.0 // indirect
	github.com/openshift/runtime-utils v0.0.0-20220513161558-c736ec4e99ce // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/profile v1.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20171018203845-0dec1b30a021 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/quobyte/api v0.1.8 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rubiojr/go-vhd v0.0.0-20200706105327-02e210299021 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/seccomp/libseccomp-golang v0.9.2-0.20210429002308-3879420cc921 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/storageos/go-api v2.2.0+incompatible // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae // indirect
	github.com/vmware/govmomi v0.20.3 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd/api/v3 v3.5.3 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.3 // indirect
	go.etcd.io/etcd/client/v2 v2.305.3 // indirect
	go.etcd.io/etcd/client/v3 v3.5.3 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.3 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.3 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib v0.20.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0 // indirect
	go.opentelemetry.io/otel v0.20.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp v0.20.0 // indirect
	go.opentelemetry.io/otel/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/trace v0.20.0 // indirect
	go.opentelemetry.io/proto/otlp v0.7.0 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gonum.org/v1/gonum v0.6.2 // indirect
	google.golang.org/api v0.46.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220107163113-42d7afdf6368 // indirect
	google.golang.org/grpc v1.40.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/gcfg.v1 v1.2.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/warnings.v0 v0.1.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/cloud-provider v0.0.0 // indirect
	k8s.io/cluster-bootstrap v0.0.0 // indirect
	k8s.io/component-helpers v0.24.0 // indirect
	k8s.io/cri-api v0.20.1 // indirect
	k8s.io/csi-translation-lib v0.0.0 // indirect
	k8s.io/gengo v0.0.0-20211129171323-c02415ce4185 // indirect
	k8s.io/kube-controller-manager v0.0.0 // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/kubelet v0.0.0 // indirect
	k8s.io/legacy-cloud-providers v0.0.0 // indirect
	k8s.io/metrics v0.0.0 // indirect
	k8s.io/mount-utils v0.0.0 // indirect
	k8s.io/pod-security-admission v0.0.0 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.30 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/kube-storage-version-migrator v0.0.4 // indirect
	sigs.k8s.io/kustomize/api v0.11.4 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

require (
	github.com/auth0/go-jwt-middleware v0.0.0-00010101000000-000000000000 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/boltdb/bolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/containers/image/v5 v5.11.0 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/heketi/tests v0.0.0-00010101000000-000000000000 // indirect
	github.com/lpabon/godbc v0.0.0-00010101000000-000000000000 // indirect
	github.com/openshift/apiserver-library-go v0.0.0-20220617080758-f441877bb41d // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/urfave/negroni v1.0.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/ldap.v2 v2.5.1 // indirect
)

replace (
	github.com/MakeNowJust/heredoc => github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // from kubernetes
	github.com/auth0/go-jwt-middleware => github.com/auth0/go-jwt-middleware v1.0.1 // from kubernetes
	github.com/boltdb/bolt => github.com/boltdb/bolt v1.3.1 // from kubernetes
	github.com/go-bindata/go-bindata => github.com/go-bindata/go-bindata v3.1.2+incompatible // from kubernetes
	github.com/go-logr/logr => github.com/go-logr/logr v1.2.0 // from kubernetes
	github.com/go-ozzo/ozzo-validation => github.com/go-ozzo/ozzo-validation v3.5.0+incompatible // from kubernetes
	github.com/google/go-cmp => github.com/google/go-cmp v0.5.5 // from kubernetes
	github.com/google/gofuzz => github.com/google/gofuzz v1.1.0 // from kubernetes
	github.com/hashicorp/golang-lru => github.com/hashicorp/golang-lru v0.5.1 // from kubernetes
	github.com/heketi/tests => github.com/heketi/tests v0.0.0-20151005000721-f3775cbcefd6 // from kubernetes
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5 // from kubernetes
	github.com/lpabon/godbc => github.com/lpabon/godbc v0.1.1 // from kubernetes
	github.com/moby/sys/mountinfo => github.com/moby/sys/mountinfo v0.6.0 // from kubernetes
	github.com/mohae/deepcopy => github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb // from kubernetes
	github.com/onsi/ginkgo => github.com/openshift/ginkgo v4.7.0-origin.0+incompatible // from kubernetes
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2 // from kubernetes
	github.com/openshift/api => github.com/openshift/api v0.0.0-20220525145417-ee5b62754c68 // from kubernetes
	github.com/openshift/apiserver-library-go => github.com/openshift/apiserver-library-go v0.0.0-20220617080758-f441877bb41d // from kubernetes
	github.com/openshift/build-machinery-go => github.com/openshift/build-machinery-go v0.0.0-20211213093930-7e33a7eb4ce3 // from kubernetes
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20220525160904-9e1acff93e4a // from kubernetes
	github.com/openshift/library-go => github.com/openshift/library-go v0.0.0-20220525173854-9b950a41acdc // from kubernetes
	github.com/pkg/errors => github.com/pkg/errors v0.9.1 // from kubernetes
	github.com/pquerna/cachecontrol => github.com/pquerna/cachecontrol v0.0.0-20171018203845-0dec1b30a021 // from kubernetes
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.12.1 // from kubernetes
	github.com/spf13/cobra => github.com/spf13/cobra v1.4.0 // from kubernetes
	github.com/spf13/pflag => github.com/spf13/pflag v1.0.5 // from kubernetes
	github.com/stretchr/testify => github.com/stretchr/testify v1.7.0 // from kubernetes
	github.com/urfave/negroni => github.com/urfave/negroni v1.0.0 // from kubernetes
	github.com/vishvananda/netns => github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae // from kubernetes
	go.etcd.io/etcd/api/v3 => github.com/openshift/etcd/api/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // from etcd
	go.etcd.io/etcd/client/pkg/v3 => github.com/openshift/etcd/client/pkg/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // from etcd
	go.etcd.io/etcd/client/v3 => github.com/openshift/etcd/client/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // from etcd
	go.etcd.io/etcd/pkg/v3 => github.com/openshift/etcd/pkg/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // from etcd
	go.etcd.io/etcd/raft/v3 => github.com/openshift/etcd/raft/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // from etcd
	go.etcd.io/etcd/server/v3 => github.com/openshift/etcd/server/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // from etcd
	go.etcd.io/etcd/v3 => github.com/openshift/etcd/v3 v3.5.1-0.20220530161601-80cc14ea9ec9 // release etcd
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // from kubernetes
	golang.org/x/exp => golang.org/x/exp v0.0.0-20210220032938-85be41e4509f // from kubernetes
	golang.org/x/net => golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // from kubernetes
	gonum.org/v1/netlib => gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // from kubernetes
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.2.2 // from kubernetes
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0 // from kubernetes
	k8s.io/api => github.com/openshift/kubernetes/staging/src/k8s.io/api v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/apiextensions-apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/apimachinery => github.com/openshift/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/cli-runtime => github.com/openshift/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/client-go => github.com/openshift/kubernetes/staging/src/k8s.io/client-go v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/cloud-provider => github.com/openshift/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/cluster-bootstrap => github.com/openshift/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/code-generator => github.com/openshift/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/component-base => github.com/openshift/kubernetes/staging/src/k8s.io/component-base v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/component-helpers => github.com/openshift/kubernetes/staging/src/k8s.io/component-helpers v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/cri-api => github.com/openshift/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/csi-translation-lib => github.com/openshift/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.60.1 // from kubernetes
	k8s.io/kube-aggregator => github.com/openshift/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20220929164501-dc5a2fd8877b // staging kubernetes
	k8s.io/kube-controller-manager => github.com/openshift/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42 // from kubernetes
	k8s.io/kube-proxy => github.com/openshift/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/kube-scheduler => github.com/openshift/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/kubectl => github.com/openshift/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/kubelet => github.com/openshift/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/kubernetes => github.com/openshift/kubernetes v0.0.0-20220929164501-dc5a2fd8877b // release kubernetes
	k8s.io/legacy-cloud-providers => github.com/openshift/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/metrics => github.com/openshift/kubernetes/staging/src/k8s.io/metrics v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/mount-utils => github.com/openshift/kubernetes/staging/src/k8s.io/mount-utils v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/pod-security-admission => github.com/openshift/kubernetes/staging/src/k8s.io/pod-security-admission v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/sample-apiserver => github.com/openshift/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/sample-cli-plugin => github.com/openshift/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	k8s.io/sample-controller => github.com/openshift/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20220929164501-dc5a2fd8877b // from kubernetes
	sigs.k8s.io/json => sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // from kubernetes
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.2.0 // from kubernetes
)
