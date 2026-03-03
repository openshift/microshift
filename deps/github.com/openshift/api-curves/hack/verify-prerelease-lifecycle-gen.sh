#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

${SCRIPT_ROOT}/hack/update-prerelease-lifecycle-gen.sh

ret=0
git diff --exit-code --quiet || ret=$?
if [[ $ret -ne 0 ]]; then
  echo "Prerelease lifecycle generation is out of date. Please run hack/update-prerelease-lifecycle-gen.sh"
  exit 1
fi
echo "Prerelease lifecycle generation up to date."
