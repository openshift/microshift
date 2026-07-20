/*
Copyright 2025 The Kubernetes Authors.

SPDX-License-Identifier: Apache-2.0
*/

package ktesting_test

import (
	"k8s.io/klog/v2"
)

func callDepthHelper(logger klog.Logger, msg string) {
	helper, logger := logger.WithCallStackHelper()
	helper()
	logger.Info(msg)
}
