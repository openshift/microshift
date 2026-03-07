#!/usr/bin/env bash

set -o nounset
set -o pipefail

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# Build yq when it's not present and not overridden for a specific file.
if [ -z "${YQ:-}" ];then
  ${TOOLS_MAKE} yq
  YQ="${TOOLS_OUTPUT}/yq"
fi

validate_suite_files() {
  FOLDER=$1

  for crdDir in ${FOLDER}/zz_generated.featuregated-crd-manifests/*; do
    if [ ! -d ${crdDir} ]; then
      # It's likely the bash expansion didn't find any yaml files.
      continue
    fi

    for file in ${crdDir}/*.yaml; do
      if [ ! -f $file ]; then
        # It's likely the bash expansion didn't find any yaml files.
        continue
      fi

      if [ $(${YQ} eval '.apiVersion' $file) != "apiextensions.k8s.io/v1" ]; then
        continue
      fi

      if [ $(${YQ} eval '.kind' $file) != "CustomResourceDefinition" ]; then
        continue
      fi

      CRD_NAME=$(echo $file | sed s:"${FOLDER}/":: )
      GROUP=$(${YQ} eval '.spec.group' $file)
      KIND=$(${YQ} eval '.spec.names.kind' $file)
      SINGULAR=$(${YQ} eval '.spec.names.singular' $file)

      FILE_BASE=""

      FEATURESET=$(${YQ} eval '.metadata.annotations["release.openshift.io/feature-set"] // ""' $file)
      if [[ ${FEATURESET} == "Default" || ${FEATURESET} == "" ]]; then
        # Default CRDs should start with stable for their test suites.
        FILE_BASE="stable"
      elif [[ ${FEATURESET} == "TechPreviewNoUpgrade" ]]; then
        # TechPreviewNoUpgrade CRDs should start with techpreview for their test suites.
        FILE_BASE="techpreview"
      elif [[ ${FEATURESET} == "CustomNoUpgrade" ]]; then
        # CustomNoUpgrade CRDs should start with custom for their test suites.
        FILE_BASE="custom"
      else
        echo "Unknown feature set ${FEATURESET} found in CRD ${file}"
        exit 1
      fi

      crdDirName=$(basename $(dirname $file))
      testFileName=$(basename $file)
      SUITE_FILE=${FOLDER}/tests/${crdDirName}/${testFileName}

      if [ ! -f ${SUITE_FILE} ]; then
        echo "No test suite file found for CRD ${file}, expected ${SUITE_FILE}"
        exit 1
      fi
    done
  done
}

for groupVersion in ${API_GROUP_VERSION_PATHS}; do
  echo "Validating integration tests for ${groupVersion}"
  validate_suite_files ${SCRIPT_ROOT}/${groupVersion}
done
