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

set -euo pipefail
shopt -s expand_aliases

#trap 'echo "# $BASH_COMMAND"' DEBUG
#set -x

########
# INIT #
########
ROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"

#### Debugging Vars.  For testing e2e release without pushing to upstream, set to your own git/quay accounts.
ORG="redhat-et"
IMAGE_OWNER="microshift"
####

IMAGE_REPO="quay.io/$IMAGE_OWNER/microshift"
STAGING_DIR="$ROOT/_output/staging"
IMAGE_ARCH_DIGESTS="$(cat "$ROOT/scripts/release_config/base_digests")"

mkdir -p "$STAGING_DIR"

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

Inputs:
    --target      (Required) The commit-ish (hash, tag, or branch) to build a release from. Abbreviated commits are NOT permitted.
    --token       (Required) The github application auth token, use to create a github release.
    --debug, -d   Print generated script values for debugging.
    --help, -h    Print this help text.
Outputs:
- A version, formatted as 4.7.0-0.microshift-YYYY-MM-DD-HHMMSS, is applied as a git tag and pushed to the repo
- Cross-compiled binaries
- Multi-architecture container image tagged and pushed as quay.io/microshift/microshift:$VERSION
- A sha256 checksum file, containing the checksums for all binary artifacts
- A github release, containing the binary artifacts and checksum file.
'
}

generate_api_release_request() {
  local is_prerelease="${1:=true}" # (copejon) assume for now that all releases are prerelease, unless otherwise specified
  printf '{"tag_name": "%s","target_commitish": "%s","name": "%s","prerelease": %s}' "$VERSION" "$TARGET" "$VERSION" "$is_prerelease"
}

git_checkout_target() {
  target="$1"
  [ -z "$(git status --porcelain)" ] || {
    printf "The working tree is dirty - commit or stash changes before cutting a release!" >&2
    return 1
  }
  git checkout "$target"
}

git_create_release() {
  local data="$1"
  local response
  response="$(
    curl -X POST \
      -H "Accept: application/vnd.github.v3+json" \
      -H "Authorization: token $TOKEN" \
      "https://api.github.com/repos/$ORG/microshift/releases" \
      -d "${data[@]}"
  )"
  local raw_upload_url
  raw_upload_url="$(echo "$response" | grep "upload_url")"
  local upload_url
  upload_url=$(echo "$raw_upload_url" | sed -n 's,.*\(https://uploads.github.com/repos/'$ORG'/microshift/releases/[0-9a-zA-Z]*/assets\).*,\1,p')
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
  arch_ver=${tag#*:}
  podman cp "$(podman create "$tag")":/usr/bin/microshift "$dest"/microshift-"$arch_ver"
}

stage_release_image_binaries() {
  source_tags="$1"
  dest="$(prep_stage_area)"
  for t in $source_tags; do
    extract_release_image_binary "$t" "$dest"
  done
  echo $dest
}

build_release_image() {
  local arch="$1"
  local image_digest="$2"
  local tag="$IMAGE_REPO:$VERSION-$arch"
  podman build \
    -t "$tag" \
    -f "$ROOT"/images/build/Dockerfile \
    --build-arg ARCH="$arch" \
    --build-arg MAKE_TARGET="cross-build-linux-$arch" \
    --build-arg DIGEST="$image_digest" \
    --build-arg SOURCE_GIT_TAG="$VERSION" \
    . >&2
  echo "${tag}"
}

build_container_images_artifacts() {
  declare -a BUILT_RELEASE_IMAGE_TAGS
  while read ad; do
    local arch=${ad%=*}
    local digest=${ad##*=}
    BUILT_RELEASE_IMAGE_TAGS+=("$(build_release_image "$arch" "@$digest")")
  done <<<"$IMAGE_ARCH_DIGESTS"
  echo "${BUILT_RELEASE_IMAGE_TAGS[@]}"
}

push_container_image_artifacts() {
  local image_tags="${1}"
  for t in $image_tags; do
    podman push "$t"
  done
}

generate_container_manifest() {
  local source_tags="$1"
  local manifest_tag_options=()
  for t in $source_tags; do
    manifest_tag_options+=("--amend ${t}")
  done

  podman manifest create "$IMAGE_REPO:$VERSION" ${manifest_tag_options[*]} >&2
  podman manifest push "$IMAGE_REPO:$VERSION"
  echo "$IMAGE_REPO:$VERSION"
}

debug() {
  local version="$1"
  local api_request="$2"
  printf "generate_version: %s\n" "$version"
  printf "compose_release_request: %s\n" "$api_request"
}

########
# MAIN #
########
[ $# -eq 0 ] && {
  help
  exit 1
}

while [ $# -gt 0 ]; do
  case "$1" in
  "--target")
    TARGET="${2:-}"
    [[ "${TOKEN:=}" =~ ^-+ ]] || [[ -z "$TARGET" ]] && {
      printf "flag $1 requires git commit-ish (branch, tag, hash) value"
      exit 1
    }
    shift 2
    ;;
  "--token")
    TOKEN="${2:-}"
    [[ "$TOKEN" =~ ^-+ ]] || [[ -z "$TOKEN" ]] && {
      printf "flag $1 git release API calls require robot token"
      exit 1
    }
    shift 2
    ;;
  "--version")
    VERSION="${2:-}"
    [[ "$VERSION" =~ ^-+ ]] || [[ -z "$VERSION" ]] && {
      printf "flag $1"
      exit 1
    }
    shift 2
    ;;
  "-d" | "--debug")
    DEBUG=0
    shift
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

[ -z ${TOKEN:-} ] && {
  printf ""
  exit 1
}
[ -z ${TARGET:-} ] && {
  printf ""
  exit 1
}

# Generate data early for debugging
API_DATA="$(generate_api_release_request "true")" # leave body empty for now

[ ${DEBUG:=1} -eq 0 ] && {
  debug "$VERSION" "$API_DATA"
  exit 0
}

git_checkout_target "$TARGET" || exit 1
RELEASE_IMAGE_TAGS="$(build_container_images_artifacts)"  || exit 1
STAGE_DIR=$(stage_release_image_binaries "$RELEASE_IMAGE_TAGS")    || exit 1
push_container_image_artifacts "$RELEASE_IMAGE_TAGS"      || exit 1
RELEASE_MANIFEST="$(generate_container_manifest "$RELEASE_IMAGE_TAGS" "$VERSION")"  || exit 1
UPLOAD_URL="$(git_create_release "$API_DATA" "$TOKEN")" || exit 1
git_post_artifacts "$STAGE_DIR" "$UPLOAD_URL" "$TOKEN" || exit 1
git switch -