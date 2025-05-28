module k8s.io/klog/examples

go 1.13

require (
	github.com/go-logr/logr v1.4.1
	github.com/go-logr/zapr v1.2.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	go.uber.org/goleak v1.1.12
	go.uber.org/zap v1.19.0
	golang.org/x/tools v0.1.5
	k8s.io/klog/v2 v2.30.0
)

replace k8s.io/klog/v2 => ../
