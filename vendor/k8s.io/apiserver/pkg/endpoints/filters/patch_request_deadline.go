/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filters

import (
	"context"
	"net/http"
	"time"
)

// RequestContextWithUpperBoundOrWorkAroundOurBrokenCaseWhereTimeoutWasNotAppliedYet returns
// a new deadline bound context for the request.
//
// If the request context is already setup with a deadline then use the
// requestTimeoutUpperBound as the upper bound.
// If the request context is not setup with a deadline then use:
//  - user specified timeout in the request URI, otherwise
//  - use the default value in requestTimeoutUpperBound
func RequestContextWithUpperBoundOrWorkAroundOurBrokenCaseWhereTimeoutWasNotAppliedYet(req *http.Request, requestTimeoutUpperBound time.Duration) (context.Context, context.CancelFunc) {
	ctx := req.Context()
	if _, ok := ctx.Deadline(); ok {
		// the request already has a deadline set, use the parent
		// context to setup an upper bound deadline.
		return context.WithTimeout(ctx, requestTimeoutUpperBound)
	}

	// the request context does not have any deadline set yet, it could be
	// a long running request that WithRequestDeadline did not apply to.
	// set an upper bound deadline using the user specified timeout in the
	// request URI if available, otherwise use the default value.
	timeout := parseTimeoutWithDefault(req, requestTimeoutUpperBound)
	return context.WithTimeout(ctx, timeout)
}

// parseTimeoutWithDefault parses the given HTTP request URL and extracts
// the timeout query parameter value if specified by the user.
// If a timeout is not specified it returns the default value specified.
func parseTimeoutWithDefault(req *http.Request, defaultTimeout time.Duration) time.Duration {
	userSpecifiedTimeout, ok, _ := parseTimeout(req)
	if ok && userSpecifiedTimeout > 0 {
		return userSpecifiedTimeout
	}
	return defaultTimeout
}
