// Package tools imports things required by build scripts, to force `go mod` to see them as dependencies
//go:build tools
// +build tools

package tools

import (
	_ "k8s.io/kube-openapi/cmd/openapi-gen"
)
