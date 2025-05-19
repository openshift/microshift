#!/usr/bin/env bash

# Copyright 2024 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script fixes the vanity imports programmatically.
# Usage: `hack/update-vanity-imports.sh`.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

kube::golang::verify_go_version

GOTOOLCHAIN="$(kube::golang::hack_tools_gotoolchain)" go -C "${KUBE_ROOT}/hack/tools" install github.com/jcchavezs/porto/cmd/porto

porto --restrict-to-dirs="staging" --restrict-to-files="doc\\.go$" -w "${KUBE_ROOT}"

