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

SOURCE_VERSION=$(grep '^OCP_VERSION' Makefile.version.aarch64.var | cut -d'=' -f2 | tr -d ' ' | cut -d'-' -f1)
MAJOR_VERSION=$(echo "${SOURCE_VERSION}" | cut -d'.' -f1)
MINOR_VERSION=$(echo "${SOURCE_VERSION}" | cut -d'.' -f2)

# Calculate previous version, handling cross-major boundaries (e.g., 5.0 -> 4.22)
if (( MINOR_VERSION > 0 )); then
    PREVIOUS_MAJOR_VERSION="${MAJOR_VERSION}"
    PREVIOUS_MINOR_VERSION=$(( MINOR_VERSION - 1 ))
else
    # Cross-major boundary: map of last minor version for each major
    declare -A LAST_MINOR_FOR_MAJOR=([4]=22)
    PREVIOUS_MAJOR_VERSION=$(( MAJOR_VERSION - 1 ))
    PREVIOUS_MINOR_VERSION="${LAST_MINOR_FOR_MAJOR[${PREVIOUS_MAJOR_VERSION}]:-}"
    if [[ -z "${PREVIOUS_MINOR_VERSION}" ]]; then
        echo "ERROR: No last minor version defined for major ${PREVIOUS_MAJOR_VERSION}" >&2
        exit 1
    fi
fi

PREVIOUS_RELEASE_BRANCH="origin/release-${PREVIOUS_MAJOR_VERSION}.${PREVIOUS_MINOR_VERSION}"
FIRST_RELEASE_COMMIT=$(branch_start "${PREVIOUS_RELEASE_BRANCH}")

mkdir -p _output

# Get the list of commits in chronological order (normally git prints
# them in reverse chronological order).
COMMIT_LIST=_output/commit_list.txt
echo "Getting the list of changes since ${PREVIOUS_RELEASE_BRANCH} was created..."
git log --date-order --reverse --pretty=format:"%H %s" "${FIRST_RELEASE_COMMIT}"..origin/main scripts/auto-rebase/changelog.txt >"${COMMIT_LIST}"

HISTORY_FILE="_output/rebase_history_${MAJOR_VERSION}_${MINOR_VERSION}.txt"
rm -f "${HISTORY_FILE}"
# shellcheck disable=SC2162
while read commit summary; do
    show_details "${commit}" "${summary}" >>"${HISTORY_FILE}"
done <"${COMMIT_LIST}"

echo "Wrote ${HISTORY_FILE}"
