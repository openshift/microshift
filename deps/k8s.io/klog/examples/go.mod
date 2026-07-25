module k8s.io/klog/examples

go 1.22.0

toolchain go1.23.0

require (
	github.com/go-logr/logr v1.4.1
	github.com/go-logr/zapr v1.2.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	go.uber.org/goleak v1.1.12
	go.uber.org/zap v1.19.0
	golang.org/x/tools v0.27.0
	k8s.io/klog/v2 v2.30.0
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
)

replace k8s.io/klog/v2 => ../
