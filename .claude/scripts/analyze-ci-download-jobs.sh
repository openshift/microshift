#!/bin/bash
set -euo pipefail

# Download Prow job artifacts for analysis.
#
# Accepts JSON on stdin — either flat job array (from microshift-prow-jobs-for-release.sh)
# or nested PR array (from microshift-prow-jobs-for-pull-requests.sh --mode detail).
# Downloads artifacts into ${WORKDIR}/artifacts/${BUILD_ID}/ with parallel workers.
# Skips already-downloaded jobs. Outputs JSON job list with local paths on stdout.
#
# Usage:
#   microshift-prow-jobs-for-release.sh 4.22 | analyze-ci-download-jobs.sh
#   microshift-prow-jobs-for-release.sh 4.22 | analyze-ci-download-jobs.sh --parallel 4
#   microshift-prow-jobs-for-pull-requests.sh --mode detail | analyze-ci-download-jobs.sh
#
# Output (stdout): JSON array of job objects with "artifacts_dir" added:
#   [{"job":"...","url":"...","build_id":"...","artifacts_dir":"/tmp/.../artifacts/BUILD_ID"}, ...]
#
# Progress/errors: stderr

WORKDIR="${WORKDIR:-/tmp/analyze-ci-claude-workdir.$(date +%y%m%d)}"

# Convert a Prow view URL to a GCS path
url_to_gcs() {
    echo "$1" | sed \
        -e 's|https://prow.ci.openshift.org/view/gs/|gs://|' \
        -e 's|https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/|gs://|'
}

# Download a single job's artifacts
# Args: build_id url
# Returns: 0 on success (cached or downloaded), 1 on failure
#
# gcloud storage cp -r gs://bucket/.../BUILD_ID/ dest/ creates dest/BUILD_ID/...
# so the final layout is: ${WORKDIR}/artifacts/${BUILD_ID}/finished.json etc.
download_job() {
    local build_id="$1"
    local url="$2"
    local dest="${WORKDIR}/artifacts/${build_id}"

    if [[ -d "${dest}" ]] && [[ -f "${dest}/finished.json" ]]; then
        echo "  cached: ${build_id}" >&2
        return 0
    fi

    local gcs_path
    gcs_path=$(url_to_gcs "${url}")

    # gcloud cp -r .../BUILD_ID/ parent/ → parent/BUILD_ID/...
    # so we download into the parent and let gcloud create the BUILD_ID dir
    local parent="${WORKDIR}/artifacts"
    mkdir -p "${parent}"
    if gcloud storage cp -r "${gcs_path}/" "${parent}/" >/dev/null 2>&1; then
        echo "  downloaded: ${build_id}" >&2
        return 0
    else
        echo "  FAILED: ${build_id}" >&2
        return 1
    fi
}

usage() {
    echo "Usage: <jobs-json> | ${0} [--parallel N]" >&2
    echo "  --parallel N: number of parallel downloads (default: 6)" >&2
    echo "" >&2
    echo "Accepts JSON on stdin from:" >&2
    echo "  microshift-prow-jobs-for-release.sh (flat job array)" >&2
    echo "  microshift-prow-jobs-for-pull-requests.sh --mode detail (nested PR array)" >&2
    exit 1
}

main() {
    local parallel=6

    while [[ ${#} -gt 0 ]]; do
        case "${1}" in
            --parallel)
                [[ ${#} -lt 2 ]] && { echo "Error: --parallel requires a number" >&2; usage; }
                parallel="${2}"; shift 2 ;;
            -h|--help) usage ;;
            -*) echo "Unknown option: ${1}" >&2; usage ;;
            *) echo "Unknown argument: ${1}" >&2; usage ;;
        esac
    done

    mkdir -p "${WORKDIR}/artifacts"

    # Read all stdin into a variable
    local input
    input=$(cat)

    # Normalize input: detect format and extract flat job list
    # Release format: [{"job":...,"url":...,"build_id":...}, ...]
    # PR format: [{"pr_number":...,"jobs":[{"job":...,"url":...,"build_id":...}, ...]}, ...]
    local jobs_json
    if echo "${input}" | jq -e '.[0].jobs' >/dev/null 2>&1; then
        # PR format — flatten nested jobs, carry pr_number into each job
        jobs_json=$(echo "${input}" | jq '[.[] | .pr_number as $pr | .jobs[] | . + {pr_number: $pr}]')
    else
        # Release format — use as-is
        jobs_json="${input}"
    fi

    local total
    total=$(echo "${jobs_json}" | jq 'length')

    if [[ "${total}" -eq 0 ]]; then
        echo "No jobs to download." >&2
        echo "[]"
        return 0
    fi

    echo "Downloading artifacts for ${total} jobs (${parallel} parallel)..." >&2

    # Export functions and vars for subshells
    export WORKDIR
    export -f download_job url_to_gcs

    # Download all jobs in parallel
    local status_file
    status_file=$(mktemp)

    while IFS=$'\t' read -r build_id url; do
        (
            if download_job "${build_id}" "${url}"; then
                echo "${build_id}:ok" >> "${status_file}"
            else
                echo "${build_id}:fail" >> "${status_file}"
            fi
        ) &

        # Limit parallelism
        while [[ $(jobs -rp | wc -l) -ge ${parallel} ]]; do
            wait -n 2>/dev/null || true
        done
    done < <(echo "${jobs_json}" | jq -r '.[] | [.build_id, .url] | @tsv')
    wait

    # Count results
    local ok=0 fail=0
    if [[ -f "${status_file}" ]]; then
        ok=$(grep -c ':ok$' "${status_file}" 2>/dev/null || true)
        fail=$(grep -c ':fail$' "${status_file}" 2>/dev/null || true)
    fi
    rm -f "${status_file}"

    echo "Done: ${ok} downloaded/cached, ${fail} failed." >&2

    # Output enriched JSON with artifacts_dir added
    echo "${jobs_json}" | jq --arg workdir "${WORKDIR}" '[.[] | . + {artifacts_dir: ($workdir + "/artifacts/" + .build_id)}]'
}

main "${@}"
