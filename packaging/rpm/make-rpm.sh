#!/usr/bin/env bash
set -e -o pipefail

# must be passed down to this script from Makefile
ENV_VARS="MICROSHIFT_VERSION RPM_RELEASE SOURCE_GIT_TAG SOURCE_GIT_TREE_STATE"
for env in $ENV_VARS ; do
  if [[ -z "${!env}" ]] ; then
    echo "Error: Mandatory environment variable '${env}' is missing"
    echo ""
    echo "Run 'make rpm' or 'make srpm' instead of this script"
    exit 1
  fi
done

# generated from other info
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

MICROSHIFT_VERSION=$(echo ${MICROSHIFT_VERSION} | sed s/-/_/g)

GIT_SHA=$(git rev-parse HEAD)
# using this instead of rev-parse --short because github's is 1 char shorter than --short
GITHUB_SHA="${GIT_SHA:0:7}"

TARBALL_FILE="microshift-${GITHUB_SHA}.tar.gz"
RPMBUILD_DIR="$(git rev-parse --show-toplevel)/_output/rpmbuild/"

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
  GIT_SHA=${1:-$GIT_SHA}
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
  rpmbuild --quiet ${RPMBUILD_OPT} --define "_topdir ${RPMBUILD_DIR}" "${RPMBUILD_DIR}"SPECS/microshift.spec
}

usage() {
  echo "Usage: $(basename $0) <all | rpm | srpm> < local | commit <commit-id> >"
  exit 1
}

[ $# -lt 2 ] && usage

case $1 in
  all)  RPMBUILD_OPT=-ba ;;
  rpm)  RPMBUILD_OPT=-bb ;;
  srpm) RPMBUILD_OPT=-bs ;;
  *)    usage
esac
shift

if [ -n "${TARGET_ARCH}" ]; then
  RPMBUILD_OPT="${RPMBUILD_OPT} --target=${TARGET_ARCH}"
fi

# prepare the rpmbuild env
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
