#!/bin/bash
#
# This script is a tool for determining what changes came in via rebases for a release.

set -euo pipefail

# Find the point where the release branch diverged from the main branch.
branch_start() {
    local -r branch_name="$1"

	git show --pretty=format:%H "$(git rev-list "$(git rev-list --first-parent ^"${branch_name}" origin/main | tail -n1)^^!")"
}

# Show the details of the rebase changes for one commit
show_details() {
    local -r commit="$1"
    local -r summary="$2"

    echo "$(git log -n 1 --pretty=format:%cd "${commit}") -- ${commit} ${summary}"
    echo
    git show "${commit}:scripts/auto-rebase/changelog.txt"
    echo
}

SOURCE_VERSION=$(awk '{print $3}' Makefile.version.aarch64.var | cut -f1 -d-)
MINOR_VERSION=$(echo "${SOURCE_VERSION}" | cut -f2 -d.)
PREVIOUS_MINOR_VERSION=$(( "${MINOR_VERSION}" - 1 ))
PREVIOUS_RELEASE_BRANCH="origin/release-4.${PREVIOUS_MINOR_VERSION}"
FIRST_RELEASE_COMMIT=$(branch_start "${PREVIOUS_RELEASE_BRANCH}")

mkdir -p _output

# Get the list of commits in chronological order (normally git prints
# them in reverse chronological order).
COMMIT_LIST=_output/commit_list.txt
echo "Getting the list of changes since ${PREVIOUS_RELEASE_BRANCH} was created..."
git log --date-order --reverse --pretty=format:"%H %s" "${FIRST_RELEASE_COMMIT}"..origin/main scripts/auto-rebase/changelog.txt >"${COMMIT_LIST}"

HISTORY_FILE="_output/rebase_history_4_${MINOR_VERSION}.txt"
rm -f "${HISTORY_FILE}"
# shellcheck disable=SC2162
while read commit summary; do
    show_details "${commit}" "${summary}" >>"${HISTORY_FILE}"
done <"${COMMIT_LIST}"

echo "Wrote ${HISTORY_FILE}"
