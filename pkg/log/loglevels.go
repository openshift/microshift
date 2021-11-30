/*
© 2021 Red Hat, Inc.

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
package log

// Log levels defined for use with:
//   klog.V(__).InfoS

// based on:
//   https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md

const (
	INFO              = 0
	WARNING           = 1
	SYSTEM_STATE      = 2
	REQUESTS          = 2
	EXTENDED          = 3
	DEBUG             = 4
	TRACE             = 5
)

