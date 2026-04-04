#!/bin/bash
set -euo pipefail

# Prow Jobs for Pull Requests
# Data modes (summary, detail): JSON on stdout
# Action modes (approve, restart): text on stdout
# Progress/errors: stderr

GCS_API="https://storage.googleapis.com/storage/v1/b/test-platform-results/o"
GCS_BASE="https://storage.googleapis.com/test-platform-results"
PROW_VIEW="https://prow.ci.openshift.org/view/gs/test-platform-results"
GH_REPO="openshift/microshift"
GCS_PR_PREFIX="pr-logs/pull/openshift_microshift"
SIGNATURE=$'\n'"*Added by $(basename "${0}")* :robot:"$'\n'

# Get open PRs as JSON array
fetch_open_prs() {
    local filter="${1:-}"
    local author="${2:-}"
    local -a gh_args=(--repo "${GH_REPO}" --state open --limit 100 --json "number,title,url")

    [[ -n "${author}" ]] && gh_args+=(--author "${author}")

    local pr_data
    pr_data=$(gh pr list "${gh_args[@]}")

    if [[ -n "${filter}" ]]; then
        echo "${pr_data}" | jq -c --arg f "${filter}" '[.[] | select(.title | contains($f))]'
    else
        echo "${pr_data}"
    fi
}

# List job names for a PR from GCS
list_pr_jobs() {
    local pr="${1}"
    curl -s --max-time 30 "${GCS_API}?prefix=${GCS_PR_PREFIX}/${pr}/&delimiter=/" | \
        jq -r '.prefixes[]? // empty' | \
        sed "s|${GCS_PR_PREFIX}/${pr}/||; s|/$||"
}

# Get latest build result as a JSON object
get_latest_build() {
    local pr="${1}" job="${2}"
    local build_id finished_json

    build_id=$(curl -s --max-time 10 "${GCS_BASE}/${GCS_PR_PREFIX}/${pr}/${job}/latest-build.txt" 2>/dev/null) || return 1
    [[ -z "${build_id}" ]] && return 1

    finished_json=$(curl -s --max-time 10 "${GCS_BASE}/${GCS_PR_PREFIX}/${pr}/${job}/${build_id}/finished.json" 2>/dev/null) || return 1

    local url="${PROW_VIEW}/pr-logs/pull/openshift_microshift/${pr}/${job}/${build_id}"
    echo "${finished_json}" | jq -c \
        --arg job "${job}" --arg url "${url}" --arg build_id "${build_id}" \
        '{
            job: $job,
            status: (.result // "PENDING"),
            url: $url,
            build_id: $build_id,
            finished: (if (.timestamp // 0) > 0 then .timestamp | todate else null end)
        }'
}

# Fetch job results for a single PR into temp dir (parallelized)
fetch_pr_results() {
    local pr="${1}"
    local tmpdir="${2}"
    local jobs

    jobs=$(list_pr_jobs "${pr}") || { touch "${tmpdir}/.fetch_failed"; return 1; }
    [[ -z "${jobs}" ]] && return 0

    while IFS= read -r job; do
        (
            result=$(get_latest_build "${pr}" "${job}" 2>/dev/null) || { touch "${tmpdir}/.fetch_failed"; exit 1; }
            if [[ -n "${result}" ]]; then
                echo "${result}" > "${tmpdir}/${job}.json"
            else
                touch "${tmpdir}/.fetch_failed"
            fi
        ) &
    done <<< "${jobs}"
    wait

    [[ -f "${tmpdir}/.fetch_failed" ]] && return 1
    return 0
}

# Collect per-job JSON files into a single JSON array
collect_jobs_json() {
    local tmpdir="${1}"
    local files=("${tmpdir}"/*.json)
    if [[ ! -f "${files[0]}" ]]; then
        echo "[]"
        return
    fi
    cat "${files[@]}" | jq -s '.'
}

# Summary mode: JSON array of PRs with pass/fail counts
mode_summary() {
    local filter="${1:-}" author="${2:-}"
    local pr_data output_tmp

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")
    [[ "$(echo "${pr_data}" | jq 'length')" -eq 0 ]] && { echo "[]"; return; }

    echo "Fetching job results..." >&2
    output_tmp=$(mktemp)

    while IFS=$'\t' read -r pr_number pr_title pr_url; do
        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "PR #${pr_number}: incomplete job results, skipping" >&2
            rm -rf "${tmpdir}"
            continue
        fi

        local passed=0 failed=0 other=0 total=0
        for f in "${tmpdir}"/*.json; do
            [[ -f "${f}" ]] || continue
            local status
            status=$(jq -r '.status' "${f}")
            total=$((total + 1))
            case "${status}" in
                SUCCESS) passed=$((passed + 1)) ;;
                FAILURE) failed=$((failed + 1)) ;;
                *)       other=$((other + 1)) ;;
            esac
        done
        rm -rf "${tmpdir}"

        jq -nc --argjson n "${pr_number}" --arg t "${pr_title}" --arg u "${pr_url}" \
            --argjson p "${passed}" --argjson f "${failed}" \
            --argjson o "${other}" --argjson to "${total}" \
            '{pr_number: $n, title: $t, url: $u, passed: $p, failed: $f, other: $o, total: $to}' \
            >> "${output_tmp}"
    done < <(echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv')

    jq -s '.' "${output_tmp}"
    rm -f "${output_tmp}"
}

# Detail mode: JSON array of PRs with full job lists
mode_detail() {
    local filter="${1:-}" author="${2:-}"
    local pr_data output_tmp

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")
    [[ "$(echo "${pr_data}" | jq 'length')" -eq 0 ]] && { echo "[]"; return; }

    echo "Fetching job results..." >&2
    output_tmp=$(mktemp)

    while IFS=$'\t' read -r pr_number pr_title pr_url; do
        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "PR #${pr_number}: incomplete job results, skipping" >&2
            rm -rf "${tmpdir}"
            continue
        fi

        local jobs_json
        jobs_json=$(collect_jobs_json "${tmpdir}")
        rm -rf "${tmpdir}"

        jq -nc --argjson n "${pr_number}" --arg t "${pr_title}" --arg u "${pr_url}" \
            --argjson jobs "${jobs_json}" \
            '{pr_number: $n, title: $t, url: $u, jobs: $jobs}' >> "${output_tmp}"
    done < <(echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv')

    jq -s '.' "${output_tmp}"
    rm -f "${output_tmp}"
}

# Approve mode: add /lgtm to PRs where all jobs pass
mode_approve() {
    local filter="${1:-}" author="${2:-}"
    local pr_data

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")
    [[ "$(echo "${pr_data}" | jq 'length')" -eq 0 ]] && { echo "No open pull requests found."; return; }

    echo "Fetching job results..." >&2

    while IFS=$'\t' read -r pr_number pr_title pr_url; do
        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "PR #${pr_number}: incomplete job results, skipping"
            rm -rf "${tmpdir}"
            continue
        fi

        local total=0 success=0
        for f in "${tmpdir}"/*.json; do
            [[ -f "${f}" ]] || continue
            local status
            status=$(jq -r '.status' "${f}")
            total=$((total + 1))
            [[ "${status}" == "SUCCESS" ]] && success=$((success + 1))
        done
        rm -rf "${tmpdir}"

        if [[ "${total}" -eq 0 ]]; then
            echo "PR #${pr_number}: No jobs found, skipping"
            continue
        fi

        if [[ "${success}" -eq "${total}" ]]; then
            local comment=$'/lgtm\n/verified by ci\n'
            comment+="${SIGNATURE}"

            echo "PR #${pr_number}: All ${total} jobs passed, approving..."
            gh pr comment "${pr_number}" --repo "${GH_REPO}" --body "${comment}"
            echo "PR #${pr_number}: Approved"
        else
            echo "PR #${pr_number}: ${success}/${total} jobs passed, skipping"
        fi
    done < <(echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv')
}

# Restart mode: comment /test for each failed job
mode_restart() {
    local filter="${1:-}" author="${2:-}"
    local pr_data

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")
    [[ "$(echo "${pr_data}" | jq 'length')" -eq 0 ]] && { echo "No open pull requests found."; return; }

    echo "Fetching job results..." >&2

    while IFS=$'\t' read -r pr_number pr_title pr_url; do
        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "PR #${pr_number}: incomplete job results, skipping"
            rm -rf "${tmpdir}"
            continue
        fi

        local failed_jobs=()
        for f in "${tmpdir}"/*.json; do
            [[ -f "${f}" ]] || continue
            local job status
            job=$(jq -r '.job' "${f}")
            status=$(jq -r '.status' "${f}")
            [[ "${status}" == "FAILURE" ]] && failed_jobs+=("${job}")
        done

        if [[ ${#failed_jobs[@]} -eq 0 ]]; then
            rm -rf "${tmpdir}"
            echo "PR #${pr_number}: No failed jobs, skipping"
            continue
        fi

        rm -rf "${tmpdir}"

        # Fetch short /test names from prowjob.json for each failed job
        local comment=""
        for job in "${failed_jobs[@]}"; do
            local build_id short_name
            build_id=$(curl -s --max-time 10 "${GCS_BASE}/${GCS_PR_PREFIX}/${pr_number}/${job}/latest-build.txt" 2>/dev/null) || continue
            short_name=$(curl -s --max-time 10 "${GCS_BASE}/${GCS_PR_PREFIX}/${pr_number}/${job}/${build_id}/prowjob.json" 2>/dev/null | \
                jq -r '.spec.rerun_command // empty' 2>/dev/null | sed 's|^/test ||') || short_name=""
            short_name=$(echo "${short_name}" | xargs)
            [[ -z "${short_name}" ]] && continue
            comment+="/test ${short_name}"$'\n'
        done

        if [[ -z "${comment}" ]]; then
            echo "PR #${pr_number}: Could not resolve rerun commands for failed job(s), skipping"
            continue
        fi
        comment+="${SIGNATURE}"

        echo "PR #${pr_number}: Restarting ${#failed_jobs[@]} failed job(s): ${failed_jobs[*]}"
        gh pr comment "${pr_number}" --repo "${GH_REPO}" --body "${comment}"
        echo "PR #${pr_number}: Restart comment posted"
    done < <(echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv')
}

usage() {
    echo "Usage: ${0} [--mode MODE] [--filter STRING] [--author USER]" >&2
    echo "  --mode MODE:     Operation mode (default: summary)" >&2
    echo "    summary: JSON array of PRs with pass/fail counts" >&2
    echo "    detail:  JSON array of PRs with full job lists" >&2
    echo "    approve: Approve PRs where ALL test jobs passed" >&2
    echo "    restart: Restart failed test jobs by commenting /test" >&2
    echo "  --filter STRING: Only include PRs whose title contains STRING" >&2
    echo "  --author USER:   Only include PRs authored by USER" >&2
    exit 1
}

main() {
    local mode="summary"
    local filter=""
    local author=""

    while [[ ${#} -gt 0 ]]; do
        case "${1}" in
            --mode)
                [[ ${#} -lt 2 ]] && { echo "Error: mode requires an argument" >&2; usage; }
                mode="${2}"; shift 2 ;;
            --filter)
                [[ ${#} -lt 2 ]] && { echo "Error: filter requires an argument" >&2; usage; }
                filter="${2}"; shift 2 ;;
            --author)
                [[ ${#} -lt 2 ]] && { echo "Error: author requires an argument" >&2; usage; }
                author="${2}"; shift 2 ;;
            -*) echo "Unknown option: ${1}" >&2; usage ;;
            *) echo "Unknown argument: ${1}" >&2; usage ;;
        esac
    done

    case "${mode}" in
        summary) mode_summary "${filter}" "${author}" ;;
        detail)  mode_detail "${filter}" "${author}" ;;
        approve) mode_approve "${filter}" "${author}" ;;
        restart) mode_restart "${filter}" "${author}" ;;
        *) echo "Error: Unknown mode '${mode}'" >&2; usage ;;
    esac
}

main "${@}"
