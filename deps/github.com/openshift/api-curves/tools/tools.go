//go:build tools
// +build tools

package tools

import (
	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/sortkeys"
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "github.com/mikefarah/yq/v4"
	_ "github.com/vmware-archive/yaml-patch/cmd/yaml-patch"
	_ "k8s.io/code-generator"
	_ "k8s.io/code-generator/cmd/go-to-protobuf"
	_ "k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo"
	_ "k8s.io/code-generator/cmd/prerelease-lifecycle-gen"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "sigs.k8s.io/kube-api-linter/pkg/plugin"
)
