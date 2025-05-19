#!/usr/bin/env bash

# Copyright 2014 The Kubernetes Authors.
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

# This script runs e2e tests on Google Cloud Platform.
# Usage: `hack/ginkgo-e2e.sh`.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${KUBE_ROOT}/cluster/common.sh"
source "${KUBE_ROOT}/hack/lib/init.sh"

# Find the ginkgo binary build as part of the release.
ginkgo=$(kube::util::find-binary "ginkgo")
e2e_test=$(kube::util::find-binary "e2e.test")

# --- Setup some env vars.

GINKGO_PARALLEL=${GINKGO_PARALLEL:-n} # set to 'y' to run tests in parallel
GINKGO_SILENCE_SKIPS=${GINKGO_SILENCE_SKIPS:-y} # set to 'n' to see S character for each skipped test
GINKGO_FORCE_NEWLINES=${GINKGO_FORCE_NEWLINES:-$( if [ "${CI:-false}" = "true" ]; then echo "y"; else echo "n"; fi )} # set to 'y' to print a newline after each S or o character
CLOUD_CONFIG=${CLOUD_CONFIG:-""}


# If 'y', Ginkgo's reporter will not use escape sequence to color output.
#
# Since Kubernetes 1.25, the default is to use colors only when connected to
# a terminal. That is the right choice for all Prow jobs (Spyglass doesn't
# render them properly).
GINKGO_NO_COLOR=${GINKGO_NO_COLOR:-$(if [ -t 2 ]; then echo n; else echo y; fi)}

# If 'y', will rerun failed tests once to give them a second chance.
GINKGO_TOLERATE_FLAKES=${GINKGO_TOLERATE_FLAKES:-n}

# If set, the command executed will be:
# - `dlv exec` if set to "delve"
# - `gdb` if set to "gdb"
# NOTE: for this to work the e2e.test binary has to be compiled with
# make DBG=1 WHAT=test/e2e/e2e.test
E2E_TEST_DEBUG_TOOL=${E2E_TEST_DEBUG_TOOL:-}

: "${KUBECTL:="${KUBE_ROOT}/cluster/kubectl.sh"}"
: "${KUBE_CONFIG_FILE:="config-test.sh"}"

export KUBECTL KUBE_CONFIG_FILE

source "${KUBE_ROOT}/cluster/kube-util.sh"

function detect-master-from-kubeconfig() {
    export KUBECONFIG=${KUBECONFIG:-$DEFAULT_KUBECONFIG}

    local cc
    cc=$("${KUBE_ROOT}/cluster/kubectl.sh" config view -o jsonpath="{.current-context}")
    if [[ -n "${KUBE_CONTEXT:-}" ]]; then
      cc="${KUBE_CONTEXT}"
    fi
    local cluster
    cluster=$("${KUBE_ROOT}/cluster/kubectl.sh" config view -o jsonpath="{.contexts[?(@.name == \"${cc}\")].context.cluster}")
    KUBE_MASTER_URL=$("${KUBE_ROOT}/cluster/kubectl.sh" config view -o jsonpath="{.clusters[?(@.name == \"${cluster}\")].cluster.server}")
}

# ---- Do cloud-provider-specific setup
if [[ -n "${KUBERNETES_CONFORMANCE_TEST:-}" ]]; then
    echo "Conformance test: not doing test setup."
    KUBERNETES_PROVIDER=${KUBERNETES_CONFORMANCE_PROVIDER:-"skeleton"}

    detect-master-from-kubeconfig

    auth_config=(
      "--kubeconfig=${KUBECONFIG}"
    )
else
    echo "Setting up for KUBERNETES_PROVIDER=\"${KUBERNETES_PROVIDER}\"."

    prepare-e2e

    detect-master >/dev/null

    KUBE_MASTER_URL="${KUBE_MASTER_URL:-}"
    if [[ -z "${KUBE_MASTER_URL:-}" && -n "${KUBE_MASTER_IP:-}" ]]; then
      KUBE_MASTER_URL="https://${KUBE_MASTER_IP}"
    fi

    auth_config=(
      "--kubeconfig=${KUBECONFIG:-$DEFAULT_KUBECONFIG}"
    )
fi

if [[ -n "${NODE_INSTANCE_PREFIX:-}" ]]; then
  NODE_INSTANCE_GROUP="${NODE_INSTANCE_PREFIX}-group"
fi

if [[ "${KUBERNETES_PROVIDER}" == "gce" ]]; then
  set_num_migs
  NODE_INSTANCE_GROUP=""
  for ((i=1; i<=NUM_MIGS; i++)); do
    if [[ ${i} == "${NUM_MIGS}" ]]; then
      # We are assigning the same mig names as create-nodes function from cluster/gce/util.sh.
      NODE_INSTANCE_GROUP="${NODE_INSTANCE_GROUP}${NODE_INSTANCE_PREFIX}-group"
    else
      NODE_INSTANCE_GROUP="${NODE_INSTANCE_GROUP}${NODE_INSTANCE_PREFIX}-group-${i},"
    fi
  done
fi

# TODO(kubernetes/test-infra#3330): Allow NODE_INSTANCE_GROUP to be
# set before we get here, which eliminates any cluster/gke use if
# KUBERNETES_CONFORMANCE_PROVIDER is set to "gke".
if [[ -z "${NODE_INSTANCE_GROUP:-}" ]] && [[ "${KUBERNETES_PROVIDER}" == "gke" ]]; then
  detect-node-instance-groups
  NODE_INSTANCE_GROUP=$(kube::util::join , "${NODE_INSTANCE_GROUP[@]}")
fi

if [[ "${KUBERNETES_PROVIDER}" == "azure" ]]; then
    if [[ ${CLOUD_CONFIG} == "" ]]; then
        echo "Missing azure cloud config"
        exit 1
    fi
fi

# These arguments are understood by both Ginkgo test suite and CLI.
# Some arguments (like --nodes) are only supported when using the CLI.
# Those get set below when choosing the program.
ginkgo_args=(
  "--poll-progress-after=${GINKGO_POLL_PROGRESS_AFTER:-60m}"
  "--poll-progress-interval=${GINKGO_POLL_PROGRESS_INTERVAL:-5m}"
  "--source-root=${KUBE_ROOT}"
)

# NOTE: Ginkgo's default timeout has been reduced from 24h to 1h in V2, set it manually here as "24h"
# for backward compatibility purpose.
ginkgo_args+=("--timeout=${GINKGO_TIMEOUT:-24h}")

if [[ -n "${CONFORMANCE_TEST_SKIP_REGEX:-}" ]]; then
  ginkgo_args+=("--skip=${CONFORMANCE_TEST_SKIP_REGEX}")
  ginkgo_args+=("--seed=1436380640")
fi

if [[ "${GINKGO_UNTIL_IT_FAILS:-}" == true ]]; then
  ginkgo_args+=("--until-it-fails=true")
fi

FLAKE_ATTEMPTS=1
if [[ "${GINKGO_TOLERATE_FLAKES}" == "y" ]]; then
  FLAKE_ATTEMPTS=2
fi
ginkgo_args+=("--flake-attempts=${FLAKE_ATTEMPTS}")

if [[ "${GINKGO_SILENCE_SKIPS}" == "y" ]]; then
  ginkgo_args+=("--silence-skips")
fi

if [[ "${GINKGO_FORCE_NEWLINES}" == "y" ]]; then
  ginkgo_args+=("--force-newlines")
fi

if [[ "${GINKGO_NO_COLOR}" == "y" ]]; then
  ginkgo_args+=("--no-color")
fi

# The --host setting is used only when providing --auth_config
# If --kubeconfig is used, the host to use is retrieved from the .kubeconfig
# file and the one provided with --host is ignored.
# Add path for things like running kubectl binary.
PATH=$(dirname "${e2e_test}"):"${PATH}"
export PATH

# Choose the program to execute and additional arguments for it.
#
# Note that the e2e.test binary must have been built with "make DBG=1"
# when using debuggers, otherwise debug information gets stripped.
case "${E2E_TEST_DEBUG_TOOL:-ginkgo}" in
  ginkgo)
    program=("${ginkgo}")
    if [[ -n "${GINKGO_PARALLEL_NODES:-}" ]]; then
      program+=("--nodes=${GINKGO_PARALLEL_NODES}")
    elif [[ ${GINKGO_PARALLEL} =~ ^[yY]$ ]]; then
      program+=("--nodes=25")
    fi
    program+=("${ginkgo_args[@]:+${ginkgo_args[@]}}")
    ;;
  delve) program=("dlv" "exec") ;;
  gdb) program=("gdb") ;;
  *) kube::log::error_exit "Unsupported E2E_TEST_DEBUG_TOOL=${E2E_TEST_DEBUG_TOOL}" ;;
esac

# Move Ginkgo arguments that are understood by the suite when not using
# the CLI.
suite_args=()
if [ "${E2E_TEST_DEBUG_TOOL:-ginkgo}" != "ginkgo" ]; then
  for arg in "${ginkgo_args[@]}"; do
    suite_args+=("--ginkgo.${arg#--}")
  done
fi

# Generate full dumps of the test result and progress in <report-dir>/ginkgo/,
# using the Ginkgo-specific JSON format and JUnit XML. Ignored if --report-dir
# is not used.
suite_args+=(--report-complete-ginkgo --report-complete-junit)

# The following invocation is fairly complex. Let's dump it to simplify
# determining what the final options are. Enabled by default in CI
# environments like Prow.
case "${GINKGO_SHOW_COMMAND:-${CI:-no}}" in y|yes|true) set -x ;; esac

"${program[@]}" "${e2e_test}" -- \
  "${auth_config[@]:+${auth_config[@]}}" \
  --host="${KUBE_MASTER_URL}" \
  --provider="${KUBERNETES_PROVIDER}" \
  --gce-project="${PROJECT:-}" \
  --gce-zone="${ZONE:-}" \
  --gce-region="${REGION:-}" \
  --gce-multizone="${MULTIZONE:-false}" \
  --gke-cluster="${CLUSTER_NAME:-}" \
  --kube-master="${KUBE_MASTER:-}" \
  --cluster-tag="${CLUSTER_ID:-}" \
  --cloud-config-file="${CLOUD_CONFIG:-}" \
  --repo-root="${KUBE_ROOT}" \
  --node-instance-group="${NODE_INSTANCE_GROUP:-}" \
  --prefix="${KUBE_GCE_INSTANCE_PREFIX:-e2e}" \
  --network="${KUBE_GCE_NETWORK:-${KUBE_GKE_NETWORK:-e2e}}" \
  --node-tag="${NODE_TAG:-}" \
  --master-tag="${MASTER_TAG:-}" \
  --docker-config-file="${DOCKER_CONFIG_FILE:-}" \
  --dns-domain="${KUBE_DNS_DOMAIN:-cluster.local}" \
  --prepull-images="${PREPULL_IMAGES:-false}" \
  ${MASTER_OS_DISTRIBUTION:+"--master-os-distro=${MASTER_OS_DISTRIBUTION}"} \
  ${NODE_OS_DISTRIBUTION:+"--node-os-distro=${NODE_OS_DISTRIBUTION}"} \
  ${NUM_NODES:+"--num-nodes=${NUM_NODES}"} \
  ${E2E_REPORT_DIR:+"--report-dir=${E2E_REPORT_DIR}"} \
  ${E2E_REPORT_PREFIX:+"--report-prefix=${E2E_REPORT_PREFIX}"} \
  "${suite_args[@]:+${suite_args[@]}}" \
  "${@}"
