#!/usr/bin/env bash
set -e -o pipefail

# must be passed down to this script from Makefile
ENV_VARS="MICROSHIFT_VERSION RPM_RELEASE SOURCE_GIT_TAG SOURCE_GIT_TREE_STATE WITH_FLANNEL"
for env in ${ENV_VARS} ; do
  if [[ -z "${!env}" ]] ; then
    echo "Error: Mandatory environment variable '${env}' is missing"
    echo ""
    echo "Run 'make rpm' or 'make srpm' instead of this script"
    exit 1
  fi
done

# generated from other info
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

MICROSHIFT_VERSION=${MICROSHIFT_VERSION//-/_}

GIT_SHA=$(git rev-parse HEAD)
# using this instead of rev-parse --short because github's is 1 char shorter than --short
GITHUB_SHA="${GIT_SHA:0:7}"

TARBALL_FILE="microshift-${GITHUB_SHA}.tar.gz"
DEFAULT_RPMBUILD_DIR="$(git rev-parse --show-toplevel)/_output/rpmbuild/"
RPMBUILD_DIR="${RPMBUILD_DIR:-${DEFAULT_RPMBUILD_DIR}}"
RPM_INFO_DIRS=""
CHECK_RPMS="n"
CHECK_SRPMS="n"

title() {
    echo -e "\E[34m\n# $1\E[00m";
}

create_local_tarball() {
  title "Creating local tarball"
  tar -czf "${RPMBUILD_DIR}/SOURCES/${TARBALL_FILE}" \
            --exclude='.git' --exclude='.idea' --exclude='.vagrant' \
            --exclude='_output' \
            --transform="s|^|microshift-${GIT_SHA}/|"  \
            --exclude="${TARBALL_FILE}" "${SCRIPT_DIR}/../../"
}

download_commit_tarball() {
  title "Downloading commit tarball"
  GIT_SHA=${1:-${GIT_SHA}}
  spectool -g --define "_topdir ${RPMBUILD_DIR}" --define="release ${RPM_RELEASE}" --define="version ${MICROSHIFT_VERSION}" \
          --define "git_commit ${GIT_SHA}" \
          -R "${SCRIPT_DIR}/microshift.spec"
}

build_commit() {
  # using --defines works for rpm building, but not for an srpm
  cat >"${RPMBUILD_DIR}"SPECS/microshift.spec <<EOF
%global release ${RPM_RELEASE}
%global version ${MICROSHIFT_VERSION}
%global commit ${1}
%global embedded_git_tag ${SOURCE_GIT_TAG}
%global embedded_git_tree_state ${SOURCE_GIT_TREE_STATE}
EOF
  cat "${SCRIPT_DIR}/microshift.spec" >> "${RPMBUILD_DIR}SPECS/microshift.spec"

  title "Building RPM packages"
  # _binary_payload refers to the compression algorithm to be used in the final RPM.
  # To see the defaults run `rpmbuild --showrc | grep binary_payload`. In this case
  # it yields w19.zstdio, which means compression level 19 using zstd algorithm. To
  # speed this up it runs w19T8.zstdio, which is the 8 thread version of the same
  # algorithm.
  # shellcheck disable=SC2086
  # We want word splitting to happen with RPMBUILD_OPT for flags to be interpreted correctly
  rpmbuild --quiet ${RPMBUILD_OPT} \
   --define "_topdir ${RPMBUILD_DIR}" \
   --define "_binary_payload w19T8.zstdio" \
   --define "with_flannel ${WITH_FLANNEL}" \
   --define "with_topolvm ${WITH_TOPOLVM}" \
   "${RPMBUILD_DIR}"SPECS/microshift.spec
}

print_info() {
  local dirs=$1
  title "RPM info from ${dirs}"
  for dir in ${dirs}; do
    find "${RPMBUILD_DIR}${dir}" -type f -exec sh -c 'i=$1; echo "${i}" && rpm -qip --dump "${i}" && echo' shell {} \;
  done
}

check_built_rpms() {
  local dir=$1
  local rpm_list=$2
  local rpm_not_found=""
  for rpm in ${rpm_list}; do
    if [ ! "$(find "${RPMBUILD_DIR}${dir}" -name "${rpm}-${MICROSHIFT_VERSION}*.rpm")" ]; then
      rpm_not_found="${rpm}-${MICROSHIFT_VERSION} ${rpm_not_found}"
    fi
  done
  if [ -n "${rpm_not_found}" ]; then
    echo "RPMs [${rpm_not_found}] not found"
    exit 1
  fi
}

usage() {
  echo "Usage: $(basename "$0") <all | rpm | srpm> < local | commit <commit-id> >"
  exit 1
}

[ $# -lt 2 ] && usage

case $1 in
  all)
    RPMBUILD_OPT=-ba
    RPM_INFO_DIRS="RPMS SRPMS"
    CHECK_RPMS="y"
    CHECK_SRPMS="y"
    ;;
  rpm)
    RPMBUILD_OPT=-bb
    RPM_INFO_DIRS="RPMS"
    CHECK_RPMS="y"
    ;;
  srpm)
    RPMBUILD_OPT=-bs
    RPM_INFO_DIRS="SRPMS"
    CHECK_SRPMS="y"
    ;;
  *)
    usage
esac
shift

if [ -n "${TARGET_ARCH}" ]; then
  RPMBUILD_OPT="${RPMBUILD_OPT} --target=${TARGET_ARCH}"
fi

# prepare the rpmbuild env
echo "Building to ${RPMBUILD_DIR}"
mkdir -p "${RPMBUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

case $1 in
    local)
      create_local_tarball
      build_commit "${GIT_SHA}"
      ;;
    commit)
      download_commit_tarball "$2"
      build_commit "$2"
      ;;
    *)
      usage
esac

if [ "${CHECK_RPMS}" = "y" ]; then
  check_built_rpms "RPMS" "microshift microshift-networking microshift-greenboot microshift-selinux microshift-release-info microshift-low-latency"
fi

if [ "${CHECK_SRPMS}" = "y" ]; then
  check_built_rpms "SRPMS" "microshift"
fi

if [ -n "${RPM_INFO}" ]; then
  print_info "${RPM_INFO_DIRS}"
fi
