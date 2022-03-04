#!/usr/bin/env bash
set -e -o pipefail

# generated from other info
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
BASE_VERSION="$(${SCRIPT_DIR}/../../pkg/release/get.sh base)"
TARBALL_FILE="microshift-pkg-release-${BASE_VERSION}.tar.gz"
RPMBUILD_DIR="${SCRIPT_DIR}/_rpmbuild/"
BUILD=${BUILD:-$2}
BUILD=${BUILD:-all}
TARGET=${TARGET:-$3}
TARGET=${TARGET:-x86_64}


case $BUILD in
  all) RPMBUILD_OPT=-ba ;;
  rpm) RPMBUILD_OPT=-bb ;;
  srpm) RPMBUILD_OPT=-bs ;;
esac

ARCHITECTURES=${ARCHITECTURES:-"x86_64 arm64 arm ppc64le riscv64"}

build() {
  cat >"${RPMBUILD_DIR}"SPECS/microshift-images.spec <<EOF
%global baseVersion ${BASE_VERSION}

EOF
  for arch in $ARCHITECTURES; do
    echo "%define images_${arch} \"$(${SCRIPT_DIR}/../../pkg/release/get.sh images $arch |  tr '\n' ' ')\"" >> "${RPMBUILD_DIR}"SPECS/microshift-images.spec
    echo "" >> "${RPMBUILD_DIR}"SPECS/microshift-images.spec
  done

  cat "${SCRIPT_DIR}/microshift-images.spec" >> "${RPMBUILD_DIR}SPECS/microshift-images.spec"

  sudo rpmbuild "${RPMBUILD_OPT}" --target $TARGET --define "_topdir ${RPMBUILD_DIR}" "${RPMBUILD_DIR}SPECS/microshift-images.spec"
}

# prepare the rpmbuild env
mkdir -p "${RPMBUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

case $1 in
    local)
           build
           ;;
    *)
      echo "Usage: $0 local [all|rpm|srpm]"
      exit 1
esac
