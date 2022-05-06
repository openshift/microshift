#!/usr/bin/env bash
set -e -o pipefail

# must be passed down to this script from Makefile
RELEASE_BASE=${RELEASE_BASE:-4.10.0}
RELEASE_PRE=${RELEASE_PRE:-${RELEASE_BASE}-0.microshift}
BUILD=${BUILD:-all}

# generated from other info
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
RPM_REL=$(git describe --tags | sed s/"${RELEASE_PRE}-"//g | sed s/-/_/g )

# add the git commit timestamp for nightlies, so updates will always work on devices old pkg < new pkg
RPM_REL=$(echo "${RPM_REL}" | sed s/nightly_/nightly_$(git show -s --format=%ct)_/g)

GIT_SHA=$(git rev-parse HEAD)
# using this instead of rev-parse --short because github's is 1 char shorter than --short
GIT_SHORTHASH="${GIT_SHA:0:7}"
TARBALL_FILE="microshift-${GIT_SHORTHASH}.tar.gz"
RPMBUILD_DIR="${SCRIPT_DIR}/_rpmbuild/"

SOURCE_GIT_TAG="$(git describe --tags | sed s/nightly-/nightly-$(git show -s --format=%ct)_/g )"


create_local_tarball() {
  tar -czf "${RPMBUILD_DIR}/SOURCES/${TARBALL_FILE}" \
            --exclude='.git' --exclude='.idea' --exclude='.vagrant' \
            --exclude='_output' --exclude='rpm/_rpmbuild' \
            --transform="s|^|microshift-${GIT_SHA}/|"  \
            --exclude="${TARBALL_FILE}" "${SCRIPT_DIR}/../../"
}

download_commit_tarball() {
  GIT_SHA=${1:-$GIT_SHA}
  spectool -g --define "_topdir ${RPMBUILD_DIR}" --define="release ${RPM_REL}" --define="version ${RELEASE_BASE}" \
          --define "git_commit ${GIT_SHA}" \
          -R "${SCRIPT_DIR}/microshift.spec"
}

download_tag_tarball() {
  spectool -g --define "_topdir ${RPMBUILD_DIR}" --define="release ${RPM_REL}" --define="version ${RELEASE_BASE}" \
          --define "github_tag ${1}" \
          -R "${SCRIPT_DIR}/microshift.spec"
}

case $BUILD in
  all) RPMBUILD_OPT=-ba ;;
  rpm) RPMBUILD_OPT=-bb ;;
  srpm) RPMBUILD_OPT=-bs ;;
esac

build_commit() {
  # using --defines worka for rpm building, but not for an srpm
  cat >"${RPMBUILD_DIR}"SPECS/microshift.spec <<EOF
%global release ${RPM_REL}
%global version ${RELEASE_BASE}
%global git_commit ${1}
%global embedded_git_commit ${SOURCE_GIT_COMMIT}
%global embedded_git_tag ${SOURCE_GIT_TAG}
%global embedded_git_tree_state ${SOURCE_GIT_TREE_STATE}
EOF
  cat "${SCRIPT_DIR}/microshift.spec" >> "${RPMBUILD_DIR}SPECS/microshift.spec"

  rpmbuild "${RPMBUILD_OPT}" --define "_topdir ${RPMBUILD_DIR}" "${RPMBUILD_DIR}"SPECS/microshift.spec
}

build_tag_commit() {
    cat >"${RPMBUILD_DIR}"SPECS/microshift.spec <<EOF
%global release ${RPM_REL}
%global version ${RELEASE_BASE}
%global github_tag ${1}
%global embedded_git_commit ${SOURCE_GIT_COMMIT}
%global embedded_git_tag ${SOURCE_GIT_TAG}
%global embedded_git_tree_state ${SOURCE_GIT_TREE_STATE}
EOF
  cat "${SCRIPT_DIR}/microshift.spec" >> "${RPMBUILD_DIR}SPECS/microshift.spec"

  rpmbuild "${RPMBUILD_OPT}" --define "_topdir ${RPMBUILD_DIR}" "${RPMBUILD_DIR}"SPECS/microshift.spec
}

# prepare the rpmbuild env
mkdir -p "${RPMBUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

case $1 in
    local) create_local_tarball
           build_commit "${GIT_SHA}"
           ;;
    commit) download_commit_tarball "$2"
            build_commit "$2"
            ;;
    tag) download_tag_tarball "$2"
         build_tag_commit "$2"
         ;;

    *)
      echo "Usage: $0 local|commit <commit-id>|tag <tag-name>"
      exit 1
esac
