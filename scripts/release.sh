#! /usr/bin/env bash
#   Copyright 2021 The Microshift authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
# release.sh
# This helper script generates and publishes Microshift releases for a given git ref. This is done by checking out the
# git ref and cross-compiling to architecture specific image digests. Images are composed from the multistage container
# file stored in ./images/build/Dockerfile, with the release images being layered on top of
# registry.access.redhat.com/ubi8/ubi-minimal:8.4. Images are wrapped with a container manifest and pushed to
# quay.io/microshift/microshift.  A github release and a tag are created and identified with the version generated
# by the Makefile. Cross-compiled binaries are copied from the container images and published in the git release.

set -euo pipefail
shopt -s expand_aliases

# debugging options
#trap 'echo "# $BASH_COMMAND"' DEBUG
#set -x

########
# INIT #
########
ROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"

# Check for a container manager cli (podman || docker), and alias it to "podman", since
# they implement the same cli interface.
__ctr_mgr_alias=$({ which podman &>/dev/null && echo "podman"; } || { which docker &>/dev/null && echo "docker"; } || echo "")
alias podman=${__ctr_mgr_alias:?"a container manager (podman || docker) is required as part of the release automation; none found"}

#########
# FUNCS #
#########

help() {
  printf 'Microshift: release.sh
This script provides some simple automation for cutting new releases of Microshift.

Use:
    ./release.sh --token $(cat /token/path)
    Note: do not use "=" with flag values
Inputs:
    --token       (Required) The github application auth token, use to create a github release.
    --debug, -d   Print generated script values for debugging.
    --help, -h    Print this help text.
Outputs:
- A version, formatted as 4.7.0-0.microshift-YYYY-MM-DD-HHMMSS, is applied as a git tag and pushed to the repo
- Multi-architecture container manifest, tagged as `quay.io/microshift/microshift:$VERSION` and `:latest`
- Cross-compiled binaries
- A sha256 checksum file, containing the checksums for all binary artifacts
- A github release, containing the binary artifacts and checksum file.

DEBUG
To test releases against a downstream/fork repository, override GIT_OWNER to forked git org/owner and QUAY_OWNER to your
quay.io owner or org.

  e.g.  GIT_OWNER=my_repo QUAY_OWNER=my_quay_repo ./release.sh --token $(cat /token/path
'
}

generate_api_release_request() {
  local is_prerelease="${1:=true}" # (copejon) assume for now that all releases are prerelease, unless otherwise specified
  printf '{"tag_name": "%s","name": "%s","prerelease": %s}' "$VERSION" "$VERSION" "$is_prerelease"
}

git_create_release() {
  local data="$1"
  local response
  response="$(
    curl -X POST \
      -H "Accept: application/vnd.github.v3+json" \
      -H "Authorization: token $TOKEN" \
      "https://api.github.com/repos/$GIT_OWNER/microshift/releases" \
      -d "${data[@]}"
  )"
  local raw_upload_url
  raw_upload_url="$(echo "$response" | grep "upload_url")"
  local upload_url
  upload_url=$(echo "$raw_upload_url" | sed -n 's,.*\(https://uploads.github.com/repos/'$GIT_OWNER'/microshift/releases/[0-9a-zA-Z]*/assets\).*,\1,p')
  # curl will return 0 even on 4xx http errors, so verify that the actually got an up_load url
  [ -z "$upload_url" ] && return 1
  echo "$upload_url"
}

git_post() {
  local bin_file="$1"
  local upload_url="$2"
  local mime_type
  mime_type="$(file -b --mime-type "$bin_file")"
  curl --fail-early \
    -X POST \
    -H "Accept: application/vnd.github.v3" \
    -H "Authorization: token $TOKEN" \
    -H "Content-Type: $mime_type" \
    --data-binary @"$bin_file" \
    "$upload_url"?name="$(basename $bin_file)"
}

git_post_artifacts() {
  local asset_dir="$1"
  local upload_url="$2"
  local files
  files="$(ls "$asset_dir")"
  for f in $files; do
    git_post "$asset_dir/$f" "$upload_url"
  done
}

prep_stage_area() {
  local asset_dir
  asset_dir=$(mktemp -d -p "$STAGING_DIR/")
  echo "$asset_dir"
}

extract_release_image_binary() {
  local tag="$1"
  local dest="$2"
  local out_bin="$dest"/microshift-"${tag#*"$VERSION-"}"
  podman cp "$(podman create "$tag")":/usr/bin/microshift "$out_bin" >&2
  echo "$out_bin"
}

stage_release_image_binaries() {
  local dest
  dest="$(prep_stage_area)"
  for t in "${RELEASE_IMAGE_TAGS[@]}"; do
    local out_bin
    out_bin=$(extract_release_image_binary "$t" "$dest") || return 1
    (
      cd "$dest"
      sha256sum "$(basename "$out_bin")" >>"$dest"/release.sha256
    ) || return 1
  done
  echo "$dest"
}

build_container_images_artifacts() {
  (
    cd "$ROOT"
    make build-containerized-cross-build SOURCE_GIT_TAG="$VERSION" IMAGE_REPO="$IMAGE_REPO"
  ) || return 1
}

push_container_image_artifacts() {
  for t in "${RELEASE_IMAGE_TAGS[@]}"; do
    podman push "$t"
  done
}

podman_create_manifest(){
  podman manifest create "$IMAGE_REPO:$VERSION" >&2
  for ref in "${RELEASE_IMAGE_TAGS[@]}"; do
    podman manifest add "$IMAGE_REPO:$VERSION" "docker://$ref"
  done
    podman manifest push "$IMAGE_REPO:$VERSION" "$IMAGE_REPO:$VERSION"
    podman manifest push "$IMAGE_REPO:$VERSION" "$IMAGE_REPO:latest"
}

docker_create_manifest(){
  local amend_images_options
  for image in "${RELEASE_IMAGE_TAGS[@]}"; do
    amend_images_options+="--amend $image"
  done
  # use docker cli directly for clarity, as this is a docker-only func
  docker manifest create "$IMAGE_REPO:$VERSION" "${RELEASE_IMAGE_TAGS[@]}" >&2
  docker tag "$IMAGE_REPO:$VERSION" "$IMAGE_REPO:latest"
  docker manifest push "$IMAGE_REPO:$VERSION"
  docker manifest push "$IMAGE_REPO:latest"
}

push_container_manifest() {
  local cli="$(alias podman)"
  if [[ "${cli#*=}" =~ docker ]]; then
    docker_create_manifest
  else
    podman_create_manifest
  fi

}
debug() {
  local version="$1"
  local api_request="$2"
  printf "Git Target: %s\n" "$TARGET"
  printf "Image Artifact: %s\n" "$IMAGE_REPO:$VERSION"
  printf "generate_version: %s\n" "$version"
  printf "compose_release_request: %s\n" "$api_request"
}

########
# MAIN #
########
while [ $# -gt 0 ]; do
  case "$1" in
    "--token")
      TOKEN="${2:-}"
      [[ "$TOKEN" =~ ^-.* ]] || [[ -z "$TOKEN" ]] && {
        printf "flag $1 git release API calls require robot token"
        exit 1
      }
      shift 2
      ;;
    "--version")
      VERSION="${2:-}"
      [[ "$VERSION" =~ ^-.* ]] || [[ -z "$VERSION" ]] && {
        printf "flag $1 expects a version input value"
        exit 1
      }
      shift 2
      ;;
    "-h" | "--help")
      help && exit
      ;;
    *)
      echo "unknown input: $1" && help && exit 1
      ;;
  esac
done

printf "Using container manager: %s\n" "$(podman --version)"

#### Debugging Vars.  For generating a full release to a fork, set to your own git/quay owner.
GIT_OWNER=${GIT_OWNER:="redhat-et"}
QUAY_OWNER=${QUAY_OWNER:="microshift"}
####

# Generate data early for debugging
API_DATA="$(generate_api_release_request "true")" # leave body empty for now

IMAGE_REPO="quay.io/$QUAY_OWNER/microshift"
RELEASE_IMAGE_TAGS=("$IMAGE_REPO:$VERSION-linux-amd64" "$IMAGE_REPO:$VERSION-linux-arm64" )

STAGING_DIR="$ROOT/_output/staging"
mkdir -p "$STAGING_DIR"

build_container_images_artifacts                                      || exit 1
STAGE_DIR=$(stage_release_image_binaries)                             || exit 1
push_container_image_artifacts                                        || exit 1
push_container_manifest                                               || exit 1
UPLOAD_URL="$(git_create_release "$API_DATA" "$TOKEN")"               || exit 1
git_post_artifacts "$STAGE_DIR" "$UPLOAD_URL" "$TOKEN"                || exit 1
