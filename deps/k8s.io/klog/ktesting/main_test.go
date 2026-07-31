/*
Copyright 2025 The Kubernetes Authors.

SPDX-License-Identifier: Apache-2.0
*/

package ktesting_test

import (
	"flag"
	"testing"

	_ "k8s.io/klog/v2/ktesting/init"
)

func TestMain(m *testing.M) {
	flag.Parse()
	m.Run()
}
