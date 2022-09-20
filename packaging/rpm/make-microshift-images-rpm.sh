#!/usr/bin/env bash
set -e -o pipefail

# generated from other info
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
BASE_VERSION="$(${SCRIPT_DIR}/../../pkg/release/get.sh base)"
RPMBUILD_DIR="${SCRIPT_DIR}/../../_output/image-rpmbuild/"
RELEASE=${RELEASE:-1}

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
    url: https://github.com/openshift/microshift
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

  ${SCRIPT_DIR}/paack.py ${BUILD} "${RPMBUILD_DIR}"microshift-images.yaml -r "${RPMBUILD_DIR}" $BUILD_OPT
}

function usage() {
  echo "Usage:"
  echo "   $(basename $0) rpm  <pull_secret> <architectures> <rpm_mock_target>"
  echo "   $(basename $0) srpm <pull_secret> <architectures>"
  echo "   $(basename $0) copr <pull_secret> <architectures> <copr_repo>"
  echo ""
  echo "pull_secret:     Path to a file containing the OpenShift pull secret"
  echo "architectures:   One or more RPM architectures"
  echo "rpm_mock_target: Target for building RPMs inside a chroot (e.g. 'rhel-8-x86_64')"
  echo "copr_repo:       Target Fedora Copr repository name (e.g. '@redhat-et/microshift-containers')"
  echo ""
  echo "Notes:"
  echo " - The OpenShift pull secret can be downloaded from https://console.redhat.com/openshift/downloads#tool-pull-secret"
  echo " - Use 'x86_64:amd64' or 'aarch64:arm64' as an architecture value"
  echo " - See /etc/mock/*.cfg for possible RPM mock target values"
  exit 1
}

# parse command line
BUILD=$1
PULL_SECRET=$2
ARCHITECTURES=$3
case $BUILD in
rpm)
  [ $# -ne 4 ] && usage
  TARGET=$4
  ;;
srpm)
  [ $# -ne 3 ] && usage
  ;;
copr)
  [ $# -ne 4 ] && usage
  COPR_REPO=$4
  ;;
*)
  usage
esac

# prepare the rpmbuild env
mkdir -p "${RPMBUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# pass pull secret as an environment variable
export REGISTRY_AUTH_FILE=$PULL_SECRET

# run the build
case $BUILD in
rpm)
  BUILD_OPT=$TARGET build
  ;;
srpm)
  build
  ;;
copr)
  BUILD_OPT=$COPR_REPO build
  ;;
esac
