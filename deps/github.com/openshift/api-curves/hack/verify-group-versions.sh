#!/usr/bin/env bash

set -o nounset
set -o pipefail
set -e

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

error=0

validate_group_versions() {
  FOLDER=$1

  if [ ! -f ${FOLDER}/register.go ]; then
      echo "No register.go file found for ${FOLDER}"
      error=1
      return
  fi

  gv=$(cat ${FOLDER}/register.go | grep -E "\sGroupVersion += schema.GroupVersion{Group: .*, Version: .*}") || true
  if [ -z "${gv}" ]; then
      echo "No GroupVersion found for ${FOLDER}"
      error=1
  fi
}

for groupVersion in ${API_GROUP_VERSION_PATHS}; do
  echo "Validating groups version for ${groupVersion}"
  validate_group_versions ${SCRIPT_ROOT}/${groupVersion}
done

if [ $error -eq 1 ]; then
  echo "FAILURE: Validation of group versions failed!"
  exit 1
fi
