#!/bin/bash

set -o nounset
set -o pipefail

if ! command -v yq > /dev/null; then
  echo "Can't find 'yq' tool in PATH, please install from https://github.com/mikefarah/yq"
  exit 1
fi

if [[ "$#" -lt 2 ]]; then
  echo "${0} <folder> <version>: generate minimal tests"
  echo
  echo "  Examples:"
  echo "    $0 operator/v1 v1   # Generates minimal tests for operator/v1 folder, and version v1 specifically."
  echo
  exit 2
fi

FOLDER=$1
VERSION=$2

for crdDir in ${FOLDER}/zz_generated.featuregated-crd-manifests/*; do
  if [ ! -d ${crdDir} ]; then
    # It's likely the bash expansion didn't find any yaml files.
    continue
  fi

  for file in ${crdDir}/*.yaml; do
    if [ $(yq eval '.apiVersion' $file) != "apiextensions.k8s.io/v1" ]; then
      continue
    fi

    if [ $(yq eval '.kind' $file) != "CustomResourceDefinition" ]; then
      continue
    fi

    CRD_NAME=$(echo $file | sed s:"${FOLDER}/":: )
    GROUP=$(yq eval '.spec.group' $file)
    KIND=$(yq eval '.spec.names.kind' $file)
    PLURAL=$(yq eval '.spec.names.plural' $file)

    crdDirName=$(basename $(dirname $file))
    testFileName=$(basename $file)
    SUITE_FILE=${FOLDER}/tests/${crdDirName}/${testFileName}

    if [ -f ${SUITE_FILE} ]; then
      continue
    fi

    mkdir -p $(dirname ${SUITE_FILE})

    featureGateName="${testFileName%.*}"
    if [ ${featureGateName} == "AAA_ungated" ]; then
      cat > ${SUITE_FILE} <<EOF
apiVersion: apiextensions.k8s.io/v1 # Hack because controller-gen complains if we don't have this
name: "${KIND}"
crdName: ${PLURAL}.${GROUP}
tests:
  onCreate:
  - name: Should be able to create a minimal ${KIND}
    initial: |
      apiVersion: ${GROUP}/${VERSION}
      kind: ${KIND}
      spec: {} # No spec is required for a ${KIND}
    expected: |
      apiVersion: ${GROUP}/${VERSION}
      kind: ${KIND}
      spec: {}
EOF
    else
      cat > ${SUITE_FILE} <<EOF
apiVersion: apiextensions.k8s.io/v1 # Hack because controller-gen complains if we don't have this
name: "${KIND}"
crdName: ${PLURAL}.${GROUP}
featureGate: ${featureGateName}
tests:
  onCreate:
  - name: Should be able to create a minimal ${KIND}
    initial: |
      apiVersion: ${GROUP}/${VERSION}
      kind: ${KIND}
      spec: {} # No spec is required for a ${KIND}
    expected: |
      apiVersion: ${GROUP}/${VERSION}
      kind: ${KIND}
      spec: {}
EOF

    fi



  done


  MAKEFILE=${FOLDER}/Makefile
  if [ ! -f ${MAKEFILE} ]; then
    cat > ${MAKEFILE} <<EOF
.PHONY: test
test:
  make -C ../../tests test GINKGO_EXTRA_ARGS=--focus="${GROUP}/${VERSION}"
EOF
  fi
done
