#!/bin/bash
set -euo pipefail

# Prow Jobs Analyzer
# Analyzes status of Prow jobs for MicroShift

PROW_URL="https://prow.ci.openshift.org/data.js"
PROW_JOB="microshift"

# Base query - fetch all MicroShift jobs for a release
fetch_base_data() {
    local release="${1}"
    curl -s --max-time 60 "${PROW_URL}" | jq -c --arg release "${release}" --arg job "${PROW_JOB}" '
        .[] | select((.job | contains($job)) and (.job | contains($release))) |
        {job: .job, status: .state, started: .started, finished: .finished, duration: .duration, url: .url}
    '
}

# Map status to icon
status_to_icon() {
    case "${1}" in
        success) echo "✓" ;;
        failure) echo "✗" ;;
        pending) echo "⋯" ;;
        *) echo "?" ;;
    esac
}

# Query for status mode - show latest run for each job
query_status() {
    {
        echo -e "JOB\tSTATUS\tFINISHED\tDURATION\tURL"
        jq -sr '
            group_by(.job) |
            map(sort_by(.started | tonumber) | reverse | first) |
            .[] |
            .status = (if .status == "success" then "✓"
                      elif .status == "failure" then "✗"
                      elif .status == "pending" then "⋯"
                      else "?" end) |
            [.job, .status, .finished, .duration, .url] |
            @tsv
        '
    } | column -t -s $'\t'
}

# Query for failed mode - show only latest jobs with failure status
query_failed() {
    {
        echo -e "JOB\tSTATUS\tFINISHED\tDURATION\tURL"
        jq -sr '
            group_by(.job) |
            map(sort_by(.started | tonumber) | reverse | first) |
            .[] |
            select(.status == "failure") |
            .status = "✗" |
            [.job, .status, .finished, .duration, .url] |
            @tsv
        '
    } | column -t -s $'\t'
}

# Usage
usage() {
    echo "Usage: ${0} [--mode MODE] <release>"
    echo "  --mode MODE: Operation mode (default: failed)"
    echo "    status: Show status of latest run for each job"
    echo "    failed: Show only latest jobs with failure status"
    echo "  release: OpenShift release version (e.g., 4.17, 4.16)"
    exit 1
}

# Status mode - show latest run for each job
mode_status() {
    local release="${1}"
    fetch_base_data "${release}" | query_status
}

# Failed mode - show only failed jobs
mode_failed() {
    local release="${1}"
    fetch_base_data "${release}" | query_failed
}

# Main
main() {
    local mode="failed"
    local release=""

    # Parse arguments
    while [[ ${#} -gt 0 ]]; do
        case "${1}" in
            --mode)
                if [[ ${#} -lt 2 ]]; then
                    echo "Error: mode requires an argument"
                    usage
                fi
                mode="${2}"
                shift 2
                ;;
            -*)
                echo "Unknown option: ${1}"
                usage
                ;;
            *)
                release="${1}"
                shift
                ;;
        esac
    done

    # Validate arguments
    if [[ -z "${release}" ]]; then
        echo "Error: release argument is required"
        usage
    fi

    # Execute mode
    case "${mode}" in
        status)
            mode_status "${release}"
            ;;
        failed)
            mode_failed "${release}"
            ;;
        *)
            echo "Error: Unknown mode '${mode}'"
            usage
            ;;
    esac
}

main "${@}"
