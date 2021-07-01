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

########
# INIT #
########
ROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"

IMAGE_REPO="quay.io/copejon/microshift"
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
    --target      (Required) The commit-ish (hash, tag, or branch) to build a release from.
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

generate_version() {
  local datetime=$(date '+%F-%H%M%S')
  local version
  version=$(printf "0.4.7-microshift-%s" "$datetime")
  printf "$version"
}

generate_api_release_request() {
  local version="$1"
  local target="$2"
  local body="$3"
  local is_prerelease="${4:=true}" # (copejon) assume for now that all releases are prerelease, unless otherwise specified

  [ -z "$version" ] && {
    printf "version not set" >&2
    return 1
  }
  [ -z "$target" ] && {
    printf "target not set" >&2
    return 1
  }
  [ -z "$body" ] && {
    printf "body not set" >&2
    return 1
  }

  local release_request=$(printf '{"tag_name": "%s","target_commitish": "%s","name": "%s","body": "%s","draft": false,"prerelease": %s}' "$version" "$target" "$target" "$body" "$is_prerelease")
  echo "$release_request"
}

git_checkout_target() {
  target="$1"
  [ -z "$(git status --porcelain)" ] || {
    printf "The working tree is dirty - commit or stash changes before cutting a release!" >&2
    return 1
  }
  git checkout "$target" || return 1
}

git_push_tag() {
  local version="$1"
  local target="$2"
  git tag -a "$version" -m "$version" "$target" || return 1
  git push "$version" --dry-run
}

git_create_release() {
  local data="$1"
  local access_token="$2"
  local reponse
  response=$(curl --data "$data" https://api.github.com/repos/"$ORG"/microshift/releases?access_token=:"$access_token")

}

git_post_artifacts() {
  local asset_dir="$1"
  curl \
    -X POST \
    -H "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/"$ORG"/microshift/releases \
    -d '{"tag_name}'
}

prep_stage_area() {
  local asset_dir
  asset_dir=$(mktemp -d -p "$STAGING_DIR/")
  echo "$asset_dir"
}

build_release_image() {
  local arch="${1:-''}"
  local image_digest="${2:-''}"
  local tag="$IMAGE_REPO:$VERSION-$arch"
  printf "BUILDING CONTAINER IMAGE: %s\n" "$tag"
  podman build \
    -t "$tag" \
    -f "$ROOT"/images/build/Dockerfile \
    --build-arg ARCH="$arch" \
    --build-arg MAKE_TARGET="cross-build-linux-$arch" \
    --build-arg DIGEST="$image_digest" \
        . >&2
  echo "${tag}"
}

extract_release_image_binary() {
  local tag="$1"
  local dest="$2"
  arch_ver=${tag#*:}
  podman cp "$(podman create -d --rm --entrypoint="bash" "$tag")":/usr/bin/local/microshift "$dest"/microshift-"$arch_ver"
}

stage_release_image_binaries() {
  source_tags="$1"
  dest="$(prep_stage_area)"
  for t in $source_tags; do
    extract_release_image_binary "$t" "$dest"
  done
}

build_container_images_artifacts() {
  declare -a BUILT_RELEASE_IMAGE_TAGS
  while read ad; do
    printf "BUILDING TO DIGEST %s\n" "$ad"
    local arch=${ad%=*}
    local digest=${ad##*=}
    BUILT_RELEASE_IMAGE_TAGS+=("$(build_release_image "$arch" "$digest")")
  done <<< "$IMAGE_ARCH_DIGESTS"
  echo "${BUILT_RELEASE_IMAGE_TAGS[@]}"
}

generate_container_manifest(){
  local source_tags="$1"
  local ver="$2"
  local manifest_tag_options=()
  for t in $source_tags; do
    maniftest_tag_options+=("--amend ${t}")
  done

  podman manifest create "$IMAGE_REPO:$ver" "${manifest_tag_options[@]}"
}

push_image_manifest(){
  local tag="$1"

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
API_DATA="$(generate_api_release_request "$VERSION" "$TARGET" "$TARGET" " " true)" # leave body empty for now

[ ${DEBUG:=1} -eq 0 ] && {
  debug "$VERSION" "$API_DATA"
  exit 0
}

#git_checkout_target "$TARGET" || exit 1
release_image_tags="$(build_container_images_artifacts)"  || exit 1
stage_release_image_binaries "$release_image_tags"        || exit 1
release_image "$release_image_tags"                       || exit 1
#git_push_tag "$VERSION" "$TARGET" || exit 1
#git_create_release "$API_DATA" || exit 1
#git_post_artifacts || exit 1