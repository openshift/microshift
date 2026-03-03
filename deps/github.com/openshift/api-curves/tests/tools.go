//go:build tools

package tests

import (
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "sigs.k8s.io/controller-runtime/tools/setup-envtest"
)
