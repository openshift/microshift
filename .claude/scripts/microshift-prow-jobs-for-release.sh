#!/bin/bash
set -euo pipefail

# Prow Jobs Analyzer for MicroShift
# Output: JSON array of job objects on stdout
# Progress/errors: stderr

PROW_URL="https://prow.ci.openshift.org/data.js"

# Fetch all MicroShift jobs for a release, return latest run per job as JSON
fetch_latest_per_job() {
    local release="${1}"
    curl -s --max-time 60 "${PROW_URL}" | jq --arg release "${release}" '
        [.[] | select((.job | contains("microshift")) and (.job | contains($release))
                       and .finished and .finished != "")] |
        group_by(.job) |
        map(sort_by(.started | tonumber) | reverse | first) |
        [.[] | {
            job: .job,
            type: .type,
            status: .state,
            finished: .finished,
            duration: .duration,
            url: .url,
            build_id: .build_id
        }]
    '
}

usage() {
    echo "Usage: ${0} [--mode MODE] <release>" >&2
    echo "  --mode MODE: Operation mode (default: failed)" >&2
    echo "    status: Latest run status for each job" >&2
    echo "    failed: Only jobs with failure status" >&2
    echo "  release: OpenShift release version (e.g., 4.22, main)" >&2
    exit 1
}

main() {
    local mode="failed"
    local release=""

    while [[ ${#} -gt 0 ]]; do
        case "${1}" in
            --mode)
                [[ ${#} -lt 2 ]] && { echo "Error: mode requires an argument" >&2; usage; }
                mode="${2}"; shift 2 ;;
            -*) echo "Unknown option: ${1}" >&2; usage ;;
            *) release="${1}"; shift ;;
        esac
    done

    [[ -z "${release}" ]] && { echo "Error: release argument is required" >&2; usage; }

    case "${mode}" in
        status) fetch_latest_per_job "${release}" ;;
        failed) fetch_latest_per_job "${release}" | jq '[.[] | select(.status == "failure")]' ;;
        *) echo "Error: Unknown mode '${mode}'" >&2; usage ;;
    esac
}

main "${@}"
