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

ROOT="$(readlink -f $(dirname ${BASH_SOURCE[0]})/../)"
OUT_DIR="$ROOT/_output/bin"
RELEASE_BINARY_ARTIFACTS=("$OUT_DIR/linux_amd64/microshift" "$OUT_DIR/linux_arm64/microshift")
ORG="copejon"

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
  printf "$release_request"
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

git_post_artifacts(){
  local asset_dir="$1"
  curl \
    -X POST \
    -H "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/"$ORG"/microshift/releases \
    -d '{"tag_name}'
}

prepare_release_binaries(){
  pushd "$ROOT/_output/"
  local asset_dir
  asset_dir=$(mktemp -d -p ./)
  checksum_file="$asset_dir/release.sha256"

  for bin in "${RELEASE_BINARY_ARTIFACTS[@]}"; do
    arch="$(basename "$(dirname $bin)")"
    cp "$bin" "$asset_dir/microshift-$arch"
    sha256sum "$asset_dir/microshift-$arch" >> "$checksum_file"
  done

  echo "$asset_dir"
}



build_binary_artifacts() {
    pushd $ROOT
    make build-containerized-cross-build  || return 1
    verify_binary_artifacts               || return 1
    popd
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
[ $# -eq 0 ] && { help; exit 1; }

while [ $# -gt 0 ]; do
  case "$1" in
  "--target")
    TARGET="$2"
    [[ "${TOKEN:=}" =~ ^-+ ]] && {
      printf "flag $1 expects value"
      exit 1
    }
    shift 2
    ;;
  "--token")
    TOKEN="$2"
    [[ "$TOKEN" =~ ^-+ ]] && {
      printf "flag $1 expects value"
      exit 1
    }
    shift 2
    ;;
  "-d"|"--debug")
    DEBUG=0
    shift
    ;;
  "-h"|"--help")
    help && exit
    ;;
  *)
    echo "unknown input: $1" && help && exit 1
    ;;
  esac
done

[ -z ${TOKEN:-} ] && {
  printf "git auth token not defined"
  exit 1
}
# Generate data early for debugging
VERSION="$(generate_version)"
API_DATA="$(generate_api_release_request "$VERSION" "$TARGET" "$TARGET" "THIS IS THE BODY" true)"

[ ${DEBUG:=1} -eq 0 ] && {
  debug "$VERSION" "$API_DATA"
  exit 0
}

git_checkout_target "$TARGET"     || exit 1
build_binary_artifacts            || exit 1
git_push_tag "$VERSION" "$TARGET" || exit 1
git_create_release "$API_DATA"    || exit 1
git_post_artifacts                || exit 1
#build_multiarch_image
#release_image