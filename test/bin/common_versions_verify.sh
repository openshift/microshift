#!/usr/bin/env bash

# Following script serves as a sort of presubmit to make sure that following files stay in sync:
# - test/bin/common_versions.sh
# - test/bin/pyutils/generate_common_versions.py
# - test/assets/common_versions.sh.template

# Check if there was a change to common_versions.sh on current branch compared to base branch.
# --exit-code: 0 means no changes, 1 means changes hence the inverted condition with '!'
# Because the top commit is Merge Commit of PR branch into base branch (e.g. main), we need to check against the earlier commit (i.e. ^1).
if ! git diff --exit-code "${SCENARIO_BUILD_BRANCH}^1...HEAD" "${ROOTDIR}/test/bin/common_versions.sh"; then
    # If the file was changed, regenerate it and compare - diff means that most likely the file was updated manually.
    "${ROOTDIR}/scripts/pyutils/create-venv.sh"

    if [ "${SCENARIO_BUILD_BRANCH}" == "main" ]; then
        y=$(awk -F'[ .]' '{print $4}' < "${ROOTDIR}/Makefile.version.x86_64.var")
    else
        y=$(echo "${SCENARIO_BUILD_BRANCH}" | awk -F'[-.]' '{ print $3 }')
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
fi
