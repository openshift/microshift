package klog

import "k8s.io/klog/v2/internal/buffer"

func WithMicroshiftLoggerComponent(c string, f func()) {
	buffer.WithMicroshiftLoggerComponent(c, f)
}
