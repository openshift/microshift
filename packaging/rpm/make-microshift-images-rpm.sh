#!/usr/bin/env bash
set -e -o pipefail

# generated from other info
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
BASE_VERSION="$(${SCRIPT_DIR}/../../pkg/release/get.sh base)"

RPMBUILD_DIR="${SCRIPT_DIR}/_rpmbuild/"
BUILD=${BUILD:-$1}
BUILD=${BUILD:-rpm}
TARGET=${TARGET:-$2}
TARGET=${TARGET:-fedora-35-x86_64}
RELEASE=${RELEASE:-1}
COPR_REPO=${COPR_REPO:-@redhat-et/microshift-containers}

ARCHITECTURES=${ARCHITECTURES:-"x86_64:amd64 aarch64:arm64"}

build() {
  cat >"${RPMBUILD_DIR}"microshift-images.yaml <<EOF
packages:
  - name: microshift-containers
    version: $(echo $BASE_VERSION | tr \- _ )
    release: $RELEASE
    license: Apache License 2.0
    summary: MicroShift related container images
    description: |
      MicroShift container images packaged for use as a read-only cri-o container storage.
    url: https://github.com/redhat-et/microshift
    path: /usr/lib/microshift/images/
    arch:
EOF

  for arch in $ARCHITECTURES; do
    IFS=":" read -r rpm_arch container_arch <<< "${arch}"
    cat >>"${RPMBUILD_DIR}"microshift-images.yaml <<EOF
      - name: ${rpm_arch}
        image_arch: ${container_arch}
        images:
EOF
    ${SCRIPT_DIR}/../../pkg/release/get.sh images $container_arch | \
      while read image; do echo "          - $image"; done >> "${RPMBUILD_DIR}"microshift-images.yaml
  done

  ${SCRIPT_DIR}/paack.py ${BUILD}  "${RPMBUILD_DIR}"microshift-images.yaml -r "${RPMBUILD_DIR}" $BUILD_OPT

}

# prepare the rpmbuild env
mkdir -p "${RPMBUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

case $BUILD in
  copr) BUILD_OPT=$COPR_REPO build;;
  rpm)  BUILD_OPT=$TARGET build;;
  srpm) build;;
      *)
      echo "Usage: $0 [copr|rpm|srpm]"
      exit 1
esac
