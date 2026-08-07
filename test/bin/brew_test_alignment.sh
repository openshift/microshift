#!/bin/bash
#
# Aligns test code (test/suites/, test/resources/) with the brew RPM
# source commit so that release CI jobs do not run HEAD tests against
# older brew-built binaries.
#
# Must be sourced, not executed directly.

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script must be sourced, not executed."
    exit 1
fi

MAX_BREW_SKEW="${MAX_BREW_SKEW:-100}"

BREW_SOURCE_COMMIT="${BREW_SOURCE_COMMIT:-}"
BREW_SOURCE_SHORTCOMMIT="${BREW_SOURCE_SHORTCOMMIT:-}"
BREW_TEST_ALIGNED="${BREW_TEST_ALIGNED:-false}"
BREW_COMMIT_DELTA="${BREW_COMMIT_DELTA:-0}"

# Extract the source commit SHA from the brew LREL RPM's release field.
# The NVR release field format is:
#   YYYYMMDDHHMMSS.p0.g<shortcommit>.assembly.<stream>.el<N>
# Sets BREW_SOURCE_COMMIT (full SHA) and BREW_SOURCE_SHORTCOMMIT.
get_brew_source_commit() {
    local -r vrel="${BREW_LREL_RELEASE_VERSION:-}"
    if [[ -z "${vrel}" ]]; then
        return 0
    fi

    local -r release_field="${vrel#*-}"

    local shortcommit=""
    if [[ "${release_field}" =~ \.g([0-9a-f]{7,})\. ]]; then
        shortcommit="${BASH_REMATCH[1]}"
    fi

    if [[ -z "${shortcommit}" ]]; then
        echo "ERROR: Cannot parse git shortcommit from brew release field: ${release_field}"
        return 1
    fi

    BREW_SOURCE_SHORTCOMMIT="${shortcommit}"
    export BREW_SOURCE_SHORTCOMMIT

    local full_sha=""
    if full_sha=$(git rev-parse "${shortcommit}^{commit}" 2>/dev/null); then
        BREW_SOURCE_COMMIT="${full_sha}"
    else
        BREW_SOURCE_COMMIT="${shortcommit}"
    fi
    export BREW_SOURCE_COMMIT
}

# Fetch the brew source commit if it is not available in the local
# git history. CI environments may use shallow clones.
fetch_brew_commit_if_needed() {
    local -r commit="$1"

    if git cat-file -e "${commit}^{commit}" 2>/dev/null; then
        return 0
    fi

    echo "Brew source commit ${commit} not found locally, fetching..."

    if git fetch origin "${commit}" 2>/dev/null; then
        if git cat-file -e "${commit}^{commit}" 2>/dev/null; then
            return 0
        fi
    fi

    if git rev-parse --is-shallow-repository 2>/dev/null | grep -q true; then
        echo "Shallow repository detected, unshallowing..."
        if git fetch --unshallow origin 2>/dev/null; then
            if git cat-file -e "${commit}^{commit}" 2>/dev/null; then
                return 0
            fi
        fi
    fi

    echo "ERROR: Cannot fetch brew source commit ${commit}"
    return 1
}

# Check out test/suites/ and test/resources/ at the brew source commit.
# Uses git checkout <sha> -- <paths>, which modifies the working tree
# without changing HEAD.
#
# Globals read:
#   BREW_SOURCE_COMMIT  - full SHA from get_brew_source_commit()
#   ROOTDIR             - repository root (from common.sh)
#   MAX_BREW_SKEW       - maximum allowed commit delta
#
# Globals set:
#   BREW_TEST_ALIGNED   - "true" if checkout was performed
#   BREW_COMMIT_DELTA   - integer commit count between brew SHA and HEAD
checkout_brew_aligned_tests() {
    BREW_TEST_ALIGNED="false"
    BREW_COMMIT_DELTA=0

    local -r brew_sha="${BREW_SOURCE_COMMIT:-}"
    if [[ -z "${brew_sha}" ]]; then
        echo "BREW_SOURCE_COMMIT is not set, skipping test alignment"
        return 0
    fi

    local -r head_sha="$(git -C "${ROOTDIR}" rev-parse HEAD)"
    if [[ "${brew_sha}" == "${head_sha}" ]]; then
        echo "Brew source commit matches HEAD, no alignment needed"
        BREW_TEST_ALIGNED="true"
        BREW_COMMIT_DELTA=0
        export BREW_TEST_ALIGNED BREW_COMMIT_DELTA
        return 0
    fi

    if ! fetch_brew_commit_if_needed "${brew_sha}"; then
        echo "ERROR: Cannot proceed with test alignment"
        return 1
    fi

    local delta
    if delta=$(git -C "${ROOTDIR}" rev-list --count "${brew_sha}..${head_sha}" 2>/dev/null); then
        BREW_COMMIT_DELTA="${delta}"
    else
        echo "WARNING: Cannot compute commit delta (commits not on same lineage)"
        BREW_COMMIT_DELTA=-1
    fi
    export BREW_COMMIT_DELTA

    echo "Brew source commit: ${brew_sha}"
    echo "HEAD commit:        ${head_sha}"
    echo "Commit delta:       ${BREW_COMMIT_DELTA}"

    if [[ "${BREW_COMMIT_DELTA}" -ne -1 ]] && [[ "${BREW_COMMIT_DELTA}" -gt "${MAX_BREW_SKEW}" ]]; then
        echo "ERROR: Commit delta ${BREW_COMMIT_DELTA} exceeds MAX_BREW_SKEW=${MAX_BREW_SKEW}"
        return 1
    fi

    echo "Checking out test/suites/ and test/resources/ at brew commit ${brew_sha}"
    if ! git -C "${ROOTDIR}" checkout "${brew_sha}" -- test/suites/ test/resources/ 2>&1; then
        echo "ERROR: Failed to checkout test directories at ${brew_sha}"
        return 1
    fi

    BREW_TEST_ALIGNED="true"
    export BREW_TEST_ALIGNED

    echo "Test code aligned to brew source commit ${brew_sha}"
    return 0
}

# Restore test directories to HEAD after tests complete.
restore_head_tests() {
    if [[ "${BREW_TEST_ALIGNED:-false}" == "true" ]]; then
        echo "Restoring test/suites/ and test/resources/ to HEAD"
        git -C "${ROOTDIR}" checkout HEAD -- test/suites/ test/resources/ 2>/dev/null || true
    fi
}

# Write a JSON sidecar artifact with brew alignment metadata.
# Each scenario invocation writes its own copy for per-scenario audit.
emit_brew_alignment_artifact() {
    local -r artifact_dir="${SCENARIO_INFO_DIR}/${SCENARIO}"
    mkdir -p "${artifact_dir}"

    local -r brew_sha="${BREW_SOURCE_COMMIT:-unset}"
    local -r head_sha="$(git -C "${ROOTDIR}" rev-parse HEAD 2>/dev/null || echo 'unknown')"
    local -r delta="${BREW_COMMIT_DELTA:-0}"
    local -r aligned="${BREW_TEST_ALIGNED:-false}"

    local coverage_delta="[]"
    if [[ "${brew_sha}" != "unset" ]] && [[ "${brew_sha}" != "${head_sha}" ]]; then
        local added_files
        added_files=$(git -C "${ROOTDIR}" diff --name-only --diff-filter=A \
            "${brew_sha}..HEAD" -- test/suites/ test/resources/ 2>/dev/null || true)
        if [[ -n "${added_files}" ]]; then
            coverage_delta=$(echo "${added_files}" | jq -R -s 'split("\n") | map(select(length > 0))')
        fi
    fi

    cat > "${artifact_dir}/brew-alignment.json" <<EOF
{
  "brew_source_commit": "${brew_sha}",
  "head_commit": "${head_sha}",
  "commit_delta": ${delta},
  "test_aligned": ${aligned},
  "max_brew_skew": ${MAX_BREW_SKEW},
  "coverage_delta": ${coverage_delta}
}
EOF

    echo "Wrote brew alignment artifact to ${artifact_dir}/brew-alignment.json"
}
