#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# Use PULL_BASE_REF for CI, otherwise use master unless overriden.
COMPARISON_BASE=${COMPARISON_BASE:-${PULL_BASE_SHA:-"master"}}

# Use a trap so this gets printed at the end of the log even when we exit early due to an error/failure
trap 'echo This verifier checks all files that have changed. In some cases you may have changed or renamed a file that \
already contained api violations, but you are not introducing a new violation.  In such cases it is appropriate to /override the \
failing CI job.' EXIT

GENERATOR=crdify EXTRA_ARGS=--crdify:comparison-base=${COMPARISON_BASE} ${SCRIPT_ROOT}/hack/update-codegen.sh
