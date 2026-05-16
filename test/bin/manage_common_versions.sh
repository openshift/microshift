#!/usr/bin/env bash

set -euo pipefail

ROOTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

action_verify() {
    echo "Verifying common_versions.sh is in sync..."

    # Check if there was a change to common_versions.sh on current branch compared to base branch.
    # --exit-code: 0 means no changes, 1 means changes hence the inverted condition with '!'
    # Because the top commit is Merge Commit of PR branch into base branch (e.g. main), we need to check against the earlier commit (i.e. ^1).
    if ! git diff --exit-code "${SCENARIO_BUILD_BRANCH}^1...HEAD" "${ROOTDIR}/test/bin/common_versions.sh" >/dev/null 2>&1; then
        echo "Detected changes to common_versions.sh, verifying they match generated version..."

        # If the file was changed, regenerate it and compare - diff means that most likely the file was updated manually.
        "${ROOTDIR}/scripts/pyutils/create-venv.sh"

        if [ "${SCENARIO_BUILD_BRANCH}" = "main" ]; then
            y="$(awk -F'[ .]' '{print $3 "." $4}' < "${ROOTDIR}/Makefile.version.x86_64.var")"
        else
            y="$(awk -F'[-]' '{ print $2 }' <<< "${SCENARIO_BUILD_BRANCH}")"
        fi

        "${ROOTDIR}/_output/pyutils/bin/python" \
            "${ROOTDIR}/test/bin/pyutils/generate_common_versions.py" \
            "${y}" \
            --update-file

        if ! git diff --exit-code "${ROOTDIR}/test/bin/common_versions.sh"; then
            echo "ERROR: Discovered that common_versions.sh was updated on the branch under test, but the regenerated version is different"
            git diff
            exit 1
        fi

        echo "Verification passed: common_versions.sh matches generated version"
    else
        echo "No changes detected to common_versions.sh"
    fi
}

action_generate() {
    local version="${1:-}"

    if [ -z "${version}" ]; then
        echo "ERROR: Version is required"
        echo "Usage: ${BASH_SOURCE[0]} generate <VERSION> [options]"
        exit 1
    fi

    "${ROOTDIR}/scripts/pyutils/create-venv.sh"
    shift
    "${ROOTDIR}/_output/pyutils/bin/python" \
        "${ROOTDIR}/test/bin/pyutils/generate_common_versions.py" \
        "${version}" \
        "$@"
}

usage() {
    cat <<EOF
Usage: ${BASH_SOURCE[0]} <action> [options]

ACTIONS:
  verify                  Verify that common_versions.sh is in sync with the template
                          and generation script. Checks if the file was manually edited
                          without regenerating.

  generate <VERSION>      Generate common_versions.sh for the specified version.
                          VERSION format depends on the release:
                          - For 4.22 and older: minor version only (e.g., 19, 22)
                          - For 5.0 and newer: X.Y format (e.g., 5.0, 5.1)

                          Additional options are passed to generate_common_versions.py:
                            --update-file    Update test/bin/common_versions.sh
                            --create-pr      Create a pull request with changes
                            --dry-run        Dry run mode

  -h, --help              Show this help message

EOF
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi

action="${1}"
shift

case "${action}" in
    verify|generate)
        "action_${action}" "$@"
        ;;
    -h|--help)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
