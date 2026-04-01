#!/bin/bash
set -euo pipefail

# Prow Jobs for Pull Requests
# Lists open MicroShift PRs with their associated Prow test job results
# Uses GCS bucket (test-platform-results) to query job data directly

GCS_API="https://storage.googleapis.com/storage/v1/b/test-platform-results/o"
GCS_BASE="https://storage.googleapis.com/test-platform-results"
PROW_VIEW="https://prow.ci.openshift.org/view/gs/test-platform-results"
GH_REPO="openshift/microshift"
GCS_PR_PREFIX="pr-logs/pull/openshift_microshift"

# Get open PRs using GitHub CLI, optionally filtered by title substring and/or author
fetch_open_prs() {
    local filter="${1:-}"
    local author="${2:-}"
    local pr_data
    local -a gh_args=(--repo "${GH_REPO}" --state open --limit 100 --json "number,title,url")

    if [[ -n "${author}" ]]; then
        gh_args+=(--author "${author}")
    fi

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

# Get latest build result for a job
get_latest_build() {
    local pr="${1}" job="${2}"
    local build_id result

    build_id=$(curl -s --max-time 10 "${GCS_BASE}/${GCS_PR_PREFIX}/${pr}/${job}/latest-build.txt" 2>/dev/null) || return 1
    [[ -z "${build_id}" ]] && return 1

    result=$(curl -s --max-time 10 "${GCS_BASE}/${GCS_PR_PREFIX}/${pr}/${job}/${build_id}/finished.json" 2>/dev/null | \
        jq -r '.result // "PENDING"' 2>/dev/null) || result="PENDING"

    local url="${PROW_VIEW}/${GCS_PR_PREFIX}/${pr}/${job}/${build_id}"
    echo "${result}	${url}"
}

# Map result to icon
result_to_icon() {
    case "${1}" in
        SUCCESS) echo "✓" ;;
        FAILURE) echo "✗" ;;
        ABORTED) echo "⊘" ;;
        PENDING) echo "⋯" ;;
        *)       echo "?" ;;
    esac
}

# Usage
usage() {
    echo "Usage: ${0} [--mode MODE] [--filter STRING] [--author USER]"
    echo "  --mode MODE:     Operation mode (default: summary)"
    echo "    summary: Show table of open PRs with test job status summary"
    echo "    detail:  Show table of open PRs with individual test job links"
    echo "    approve: Approve PRs where ALL test jobs finished successfully"
    echo "    restart: Restart failed test jobs by commenting /test for each failure"
    echo "  --filter STRING: Only include PRs whose title contains STRING"
    echo "  --author USER:   Only include PRs authored by USER (GitHub username)"
    exit 1
}

# Fetch job results for a single PR (parallelized)
# Returns non-zero if any job fetch fails, writing a .fetch_failed marker
fetch_pr_results() {
    local pr="${1}"
    local tmpdir="${2}"
    local jobs

    jobs=$(list_pr_jobs "${pr}") || { touch "${tmpdir}/.fetch_failed"; return 1; }
    if [[ -z "${jobs}" ]]; then
        return 0
    fi

    # Fetch latest build for each job in parallel
    while IFS= read -r job; do
        (
            result_line=$(get_latest_build "${pr}" "${job}" 2>/dev/null) || { touch "${tmpdir}/.fetch_failed"; exit 1; }
            if [[ -n "${result_line}" ]]; then
                echo "${job}	${result_line}" > "${tmpdir}/${job}"
            else
                touch "${tmpdir}/.fetch_failed"
            fi
        ) &
    done <<< "${jobs}"
    wait

    # Check if any subshell signaled a failure
    if [[ -f "${tmpdir}/.fetch_failed" ]]; then
        return 1
    fi
    return 0
}

# Summary mode - show PR with pass/fail counts
mode_summary() {
    local filter="${1:-}"
    local author="${2:-}"
    local pr_data

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")

    local pr_count
    pr_count=$(echo "${pr_data}" | jq 'length')

    if [[ "${pr_count}" -eq 0 ]]; then
        echo "No open pull requests found."
        return
    fi

    echo "Fetching job results..." >&2

    {
        echo -e "PR\tTITLE\t✓\t✗\t⋯\tJOBS"
        echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv' | while IFS=$'\t' read -r pr_number pr_title pr_url; do
            local tmpdir
            tmpdir=$(mktemp -d)

            if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
                echo "PR #${pr_number}: incomplete job results, skipping" >&2
                rm -rf "${tmpdir}"
                continue
            fi

            local success=0 failure=0 pending=0 total=0
            for f in "${tmpdir}"/*; do
                [[ -f "${f}" ]] || continue
                local result
                result=$(cut -f2 "${f}")
                total=$((total + 1))
                case "${result}" in
                    SUCCESS) success=$((success + 1)) ;;
                    FAILURE) failure=$((failure + 1)) ;;
                    *)       pending=$((pending + 1)) ;;
                esac
            done
            rm -rf "${tmpdir}"

            # Truncate title to 50 chars
            if [[ ${#pr_title} -gt 50 ]]; then
                pr_title="${pr_title:0:47}..."
            fi

            echo -e "${pr_url}\t${pr_title}\t${success}\t${failure}\t${pending}\t${total}"
        done
    } | column -t -s $'\t'
}

# Detail mode - show each job for each PR
mode_detail() {
    local filter="${1:-}"
    local author="${2:-}"
    local pr_data

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")

    local pr_count
    pr_count=$(echo "${pr_data}" | jq 'length')

    if [[ "${pr_count}" -eq 0 ]]; then
        echo "No open pull requests found."
        return
    fi

    echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv' | while IFS=$'\t' read -r pr_number pr_title pr_url; do
        echo ""
        echo "=== PR #${pr_number}: ${pr_title} ==="
        echo "    ${pr_url}"
        echo ""

        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "    PR #${pr_number}: incomplete job results, skipping"
            rm -rf "${tmpdir}"
            continue
        fi

        local file_count
        file_count=$(find "${tmpdir}" -maxdepth 1 -type f | wc -l)

        if [[ "${file_count}" -eq 0 ]]; then
            echo "    No Prow jobs found."
        else
            {
                echo -e "JOB\tSTATUS\tURL"
                while IFS= read -r -d '' f; do
                    local job result url icon
                    IFS=$'\t' read -r job result url < "${f}"
                    icon=$(result_to_icon "${result}")
                    echo -e "${job}\t${icon}\t${url}"
                done < <(find "${tmpdir}" -maxdepth 1 -type f -print0 | sort -z)
            } | column -t -s $'\t' | sed 's/^/    /'
        fi

        rm -rf "${tmpdir}"
    done
}

# Approve mode - add "/lgtm" and "/verified by ci" comments to PRs with all tests passing
mode_approve() {
    local filter="${1:-}"
    local author="${2:-}"
    local pr_data

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")

    local pr_count
    pr_count=$(echo "${pr_data}" | jq 'length')

    if [[ "${pr_count}" -eq 0 ]]; then
        echo "No open pull requests found."
        return
    fi

    echo "Fetching job results..." >&2

    echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv' | while IFS=$'\t' read -r pr_number pr_title pr_url; do
        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "PR #${pr_number}: incomplete job results, skipping"
            rm -rf "${tmpdir}"
            continue
        fi

        local total=0 success=0
        for f in "${tmpdir}"/*; do
            [[ -f "${f}" ]] || continue
            local result
            result=$(cut -f2 "${f}")
            total=$((total + 1))
            case "${result}" in
                SUCCESS) success=$((success + 1)) ;;
            esac
        done
        rm -rf "${tmpdir}"

        if [[ "${total}" -eq 0 ]]; then
            echo "PR #${pr_number}: No jobs found, skipping"
            continue
        fi

        if [[ "${success}" -eq "${total}" ]]; then
            local comment=$'/lgtm\n/verified by ci\n'
            comment+=$'\n'"*Added by $(basename "${0}")* :robot:"$'\n'

            echo "PR #${pr_number}: All ${total} jobs passed, approving..."
            gh pr comment "${pr_number}" --repo "${GH_REPO}" --body "${comment}"
            echo "PR #${pr_number}: Approved"
        else
            echo "PR #${pr_number}: ${success}/${total} jobs passed, skipping"
        fi
    done
}

# Restart mode - comment /test for each failed job on PRs with failures
mode_restart() {
    local filter="${1:-}"
    local author="${2:-}"
    local pr_data

    echo "Fetching open PRs..." >&2
    pr_data=$(fetch_open_prs "${filter}" "${author}")

    local pr_count
    pr_count=$(echo "${pr_data}" | jq 'length')

    if [[ "${pr_count}" -eq 0 ]]; then
        echo "No open pull requests found."
        return
    fi

    echo "Fetching job results..." >&2

    echo "${pr_data}" | jq -r '.[] | [.number, .title, .url] | @tsv' | while IFS=$'\t' read -r pr_number pr_title pr_url; do
        local tmpdir
        tmpdir=$(mktemp -d)

        if ! fetch_pr_results "${pr_number}" "${tmpdir}"; then
            echo "PR #${pr_number}: incomplete job results, skipping"
            rm -rf "${tmpdir}"
            continue
        fi

        local failed_jobs=()
        for f in "${tmpdir}"/*; do
            [[ -f "${f}" ]] || continue
            local job result
            IFS=$'\t' read -r job result _ < "${f}"
            if [[ "${result}" == "FAILURE" ]]; then
                failed_jobs+=("${job}")
            fi
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
        comment+=$'\n'"*Added by $(basename "${0}")* :robot:"$'\n'

        echo "PR #${pr_number}: Restarting ${#failed_jobs[@]} failed job(s): ${failed_jobs[*]}"
        gh pr comment "${pr_number}" --repo "${GH_REPO}" --body "${comment}"
        echo "PR #${pr_number}: Restart comment posted"
    done
}

# Main
main() {
    local mode="summary"
    local filter=""
    local author=""

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
            --filter)
                if [[ ${#} -lt 2 ]]; then
                    echo "Error: filter requires an argument"
                    usage
                fi
                filter="${2}"
                shift 2
                ;;
            --author)
                if [[ ${#} -lt 2 ]]; then
                    echo "Error: author requires an argument"
                    usage
                fi
                author="${2}"
                shift 2
                ;;
            -*)
                echo "Unknown option: ${1}"
                usage
                ;;
            *)
                echo "Unknown argument: ${1}"
                usage
                ;;
        esac
    done

    # Execute mode
    case "${mode}" in
        summary)
            mode_summary "${filter}" "${author}"
            ;;
        detail)
            mode_detail "${filter}" "${author}"
            ;;
        approve)
            mode_approve "${filter}" "${author}"
            ;;
        restart)
            mode_restart "${filter}" "${author}"
            ;;
        *)
            echo "Error: Unknown mode '${mode}'"
            usage
            ;;
    esac
}

main "${@}"
