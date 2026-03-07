#!/bin/bash

set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE}")/..

OPENSHIFT_CI=${OPENSHIFT_CI:-""}
ARTIFACT_DIR=${ARTIFACT_DIR:-""}
GINKGO=${GINKGO:-"go run ${REPO_ROOT}/vendor/github.com/onsi/ginkgo/v2/ginkgo"}
GINKGO_ARGS=${GINKGO_ARGS:-"-r -v --randomize-all --randomize-suites --keep-going --timeout=60m -p"}
GINKGO_EXTRA_ARGS=${GINKGO_EXTRA_ARGS:-""}

# Ensure that some home var is set and that it's not the root.
# This is required for the kubebuilder cache.
export HOME=${HOME:=/tmp/kubebuilder-testing}
if [ $HOME == "/" ]; then
  export HOME=/tmp/kubebuilder-testing
fi

if [ "$OPENSHIFT_CI" == "true" ] && [ -n "$ARTIFACT_DIR" ] && [ -d "$ARTIFACT_DIR" ]; then # detect ci environment there
  GINKGO_ARGS="${GINKGO_ARGS} --junit-report=junit_api_example_tests.xml --cover --output-dir=${ARTIFACT_DIR} --no-color"
fi

# Print the command we are going to run as Make would.
echo ${GINKGO} ${GINKGO_ARGS} ${GINKGO_EXTRA_ARGS} ./
${GINKGO} ${GINKGO_ARGS} ${GINKGO_EXTRA_ARGS} ./
# Capture the test result to exit on error.
TEST_RESULT=$?

# Ensure we exit based on the test result
exit ${TEST_RESULT}
