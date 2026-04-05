#!/bin/bash
set -euo pipefail

# Deterministic orchestration for analyze-ci:doctor.
#
# Two phases called by the doctor skill with LLM steps in between:
#
#   analyze-ci-doctor.sh prepare --workdir DIR <releases> [--rebase]
#     - Collects failed jobs for each release and rebase PRs
#     - Downloads all artifacts in parallel
#     - Writes per-release and PR jobs JSON files
#
#   analyze-ci-doctor.sh finalize --workdir DIR <releases>
#     - Runs analyze-ci-aggregate.py for each release and PRs
#     - Runs analyze-ci-create-report.py to generate HTML
#
# Usage from doctor skill:
#   1. analyze-ci-doctor.sh prepare --workdir $WORKDIR 4.18,4.19,4.20,main --rebase
#   2. (LLM launches prow-job agents for all jobs)
#   3. (LLM launches create-bugs agents for Jira search)
#   4. analyze-ci-doctor.sh finalize --workdir $WORKDIR 4.18,4.19,4.20,main

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WORKDIR=""

# ---------------------------------------------------------------------------
# prepare
# ---------------------------------------------------------------------------

cmd_prepare() {
    local releases_arg=""
    local do_rebase=false

    while [[ ${#} -gt 0 ]]; do
        case "${1}" in
            --workdir) WORKDIR="${2}"; shift 2 ;;
            --rebase) do_rebase=true; shift ;;
            -*) echo "Unknown option: ${1}" >&2; return 1 ;;
            *) releases_arg="${1}"; shift ;;
        esac
    done

    WORKDIR="${WORKDIR:-/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)}"

    if [[ -z "${releases_arg}" ]]; then
        echo "Error: releases argument required" >&2
        echo "Usage: $(basename "$0") prepare [--workdir DIR] <release1,release2,...> [--rebase]" >&2
        return 1
    fi

    mkdir -p "${WORKDIR}"

    IFS=',' read -ra RELEASES <<< "${releases_arg}"
    local total_jobs=0

    # Collect and download for each release
    for release in "${RELEASES[@]}"; do
        release=$(echo "${release}" | xargs)  # trim whitespace
        echo "=== Release ${release} ===" >&2

        local jobs_file="${WORKDIR}/analyze-ci-release-${release}-jobs.json"

        echo "  Collecting failed periodic jobs..." >&2
        local raw_json
        raw_json=$(bash "${SCRIPT_DIR}/microshift-prow-jobs-for-release.sh" "${release}" 2>/dev/null) || raw_json="[]"

        local filtered_json
        filtered_json=$(echo "${raw_json}" | jq '[.[] | select(.type == "periodic")]')

        local count
        count=$(echo "${filtered_json}" | jq 'length')

        if [[ "${count}" -eq 0 ]]; then
            echo "  No failed periodic jobs found" >&2
            echo "[]" > "${jobs_file}"
            continue
        fi

        echo "  Found ${count} failed periodic jobs, downloading artifacts..." >&2
        echo "${filtered_json}" | \
            bash "${SCRIPT_DIR}/analyze-ci-download-jobs.sh" --workdir "${WORKDIR}" 2>/dev/null \
            > "${jobs_file}"

        total_jobs=$((total_jobs + count))
        echo "  Done: ${jobs_file}" >&2
    done

    # Collect and download for rebase PRs
    if ${do_rebase}; then
        echo "=== Rebase Pull Requests ===" >&2

        local prs_file="${WORKDIR}/analyze-ci-prs-jobs.json"
        local prs_status_file="${WORKDIR}/analyze-ci-prs-status.json"

        echo "  Collecting rebase PRs..." >&2
        local pr_json
        pr_json=$(bash "${SCRIPT_DIR}/microshift-prow-jobs-for-pull-requests.sh" \
            --mode detail --author "microshift-rebase-script[bot]" 2>/dev/null) || pr_json="[]"

        local pr_count
        pr_count=$(echo "${pr_json}" | jq 'length')

        if [[ "${pr_count}" -eq 0 ]]; then
            echo "  No rebase PRs found" >&2
            echo "[]" > "${prs_file}"
            echo "[]" > "${prs_status_file}"
        else
            # Save job status snapshot for all PRs (used by HTML report)
            echo "${pr_json}" | jq '[.[] | {
                pr_number, title, url,
                passed:  [.jobs[] | select(.status == "SUCCESS")] | length,
                failed:  [.jobs[] | select(.status == "FAILURE")] | length,
                pending: [.jobs[] | select(.status != "SUCCESS" and .status != "FAILURE")] | length,
                total:   (.jobs | length)
            }]' > "${prs_status_file}"
            echo "  Saved status for ${pr_count} rebase PRs" >&2

            # Filter to PRs with failed jobs for artifact download
            local failed_prs
            failed_prs=$(echo "${pr_json}" | \
                jq '[.[] | select(.jobs | map(select(.status == "FAILURE")) | length > 0)]')

            local failed_pr_count
            failed_pr_count=$(echo "${failed_prs}" | jq 'length')

            if [[ "${failed_pr_count}" -eq 0 ]]; then
                echo "  No PRs with failures to investigate" >&2
                echo "[]" > "${prs_file}"
            else
                local job_count
                job_count=$(echo "${failed_prs}" | jq '[.[].jobs[] | select(.status == "FAILURE")] | length')

                echo "  Downloading artifacts for ${job_count} failed jobs across ${failed_pr_count} PRs..." >&2
                echo "${failed_prs}" | \
                    bash "${SCRIPT_DIR}/analyze-ci-download-jobs.sh" --workdir "${WORKDIR}" 2>/dev/null \
                    > "${prs_file}"

                total_jobs=$((total_jobs + job_count))
                echo "  Done: ${prs_file}" >&2
            fi
        fi
    fi

    echo "" >&2
    echo "Prepare complete: ${total_jobs} total jobs ready for analysis in ${WORKDIR}" >&2

    # Output a JSON summary for the LLM to consume
    local result="{\"workdir\":\"${WORKDIR}\",\"releases\":["
    local first=true
    for release in "${RELEASES[@]}"; do
        release=$(echo "${release}" | xargs)
        local jobs_file="${WORKDIR}/analyze-ci-release-${release}-jobs.json"
        local count=0
        if [[ -f "${jobs_file}" ]]; then
            count=$(jq 'length' "${jobs_file}")
        fi
        ${first} || result+=","
        first=false
        result+="{\"release\":\"${release}\",\"jobs\":${count},\"jobs_file\":\"${jobs_file}\"}"
    done
    result+="]"

    if ${do_rebase}; then
        local prs_file="${WORKDIR}/analyze-ci-prs-jobs.json"
        local pr_job_count=0
        if [[ -f "${prs_file}" ]]; then
            pr_job_count=$(jq 'length' "${prs_file}")
        fi
        result+=",\"prs\":{\"jobs\":${pr_job_count},\"jobs_file\":\"${prs_file}\"}"
    fi

    result+="}"
    echo "${result}" | jq .
}

# ---------------------------------------------------------------------------
# finalize
# ---------------------------------------------------------------------------

cmd_finalize() {
    local releases_arg=""

    while [[ ${#} -gt 0 ]]; do
        case "${1}" in
            --workdir) WORKDIR="${2}"; shift 2 ;;
            -*) echo "Unknown option: ${1}" >&2; return 1 ;;
            *) releases_arg="${1}"; shift ;;
        esac
    done

    WORKDIR="${WORKDIR:-/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)}"

    if [[ -z "${releases_arg}" ]]; then
        echo "Error: releases argument required" >&2
        echo "Usage: $(basename "$0") finalize [--workdir DIR] <release1,release2,...>" >&2
        return 1
    fi

    IFS=',' read -ra RELEASES <<< "${releases_arg}"

    # Aggregate each release
    for release in "${RELEASES[@]}"; do
        release=$(echo "${release}" | xargs)
        echo "=== Aggregating release ${release} ===" >&2
        python3 "${SCRIPT_DIR}/analyze-ci-aggregate.py" \
            --release "${release}" --workdir "${WORKDIR}" >/dev/null 2>&1 || \
            echo "  WARNING: aggregation failed for ${release}" >&2
    done

    # Aggregate PRs (if job files exist)
    local pr_files
    pr_files=$(find "${WORKDIR}" -name 'analyze-ci-prs-job-*.txt' 2>/dev/null | head -1)
    if [[ -n "${pr_files}" ]]; then
        echo "=== Aggregating PRs ===" >&2
        python3 "${SCRIPT_DIR}/analyze-ci-aggregate.py" \
            --prs --workdir "${WORKDIR}" >/dev/null 2>&1 || \
            echo "  WARNING: PR aggregation failed" >&2
    fi

    # Generate HTML report
    echo "=== Generating HTML report ===" >&2
    python3 "${SCRIPT_DIR}/analyze-ci-create-report.py" \
        --workdir "${WORKDIR}" "${releases_arg}"
}

# ---------------------------------------------------------------------------
# main
# ---------------------------------------------------------------------------

usage() {
    echo "Usage: $(basename "$0") <command> [--workdir DIR] <releases> [options]" >&2
    echo "" >&2
    echo "Commands:" >&2
    echo "  prepare [--workdir DIR] <releases> [--rebase]  Collect jobs and download artifacts" >&2
    echo "  finalize [--workdir DIR] <releases>            Aggregate results and generate HTML" >&2
    echo "" >&2
    echo "  <releases>: comma-separated release versions (e.g., 4.18,4.19,4.20,main)" >&2
    echo "  --workdir DIR: work directory (default: /tmp/analyze-ci-claude-workdir.YYMMDD)" >&2
    exit 1
}

main() {
    if [[ ${#} -lt 1 ]]; then
        usage
    fi

    local cmd="${1}"
    shift

    case "${cmd}" in
        prepare)  cmd_prepare "${@}" ;;
        finalize) cmd_finalize "${@}" ;;
        *) echo "Unknown command: ${cmd}" >&2; usage ;;
    esac
}

main "${@}"
