// +k8s:conversion-gen=github.com/openshift/openshift-apiserver/pkg/image/apis/image
// +k8s:conversion-gen-external-types=github.com/openshift/api/image/docker10
// +k8s:defaulter-gen=TypeMeta
// +k8s:defaulter-gen-input=../../../../../../../../github.com/openshift/api/image/docker10

// +groupName=image.openshift.io
// Package docker10 provides types used by docker/distribution and moby/moby.
// This package takes no dependency on external types.
package docker10
