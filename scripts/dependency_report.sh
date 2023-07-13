#!/bin/bash
#
# Show any go dependencies in MicroShift's go.mod that do not appear
# in other OpenShift components.

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/..")"
WORKDIR="${REPOROOT}/_output/dependency_report"

if [ $# -ne 1 ]; then
    cat - <<EOF
Usage: $0 <openshift-pull-secret-file>

  This script extracts information from the most recently rebased OCP
  payload. You may need a CI pull secret. You can get one from
  https://cloud.redhat.com/openshift/install/pull-secret

EOF
  exit 1
fi

PULL_SECRET_FILE="$1"

if [ ! -f "${PULL_SECRET_FILE}" ]; then
    >&2 echo "Error: no pull secret found at ${PULL_SECRET_FILE}"
    exit 1
fi

get_go_mod() {
    local name="$1"
    local repo_url="$2"

    raw_url=${repo_url//github.com/raw.githubusercontent.com}
    echo "${raw_url}" > "${name}/raw_url"

    if [ ! -f "${name}/go.mod" ]; then
        if ! curl --silent --fail-with-body -o "${name}/go.mod" "${raw_url}/master/go.mod"; then
            rm -f "${name}/go.mod"
            echo "No go.mod for ${name}"
        fi
    fi
}

build_dep_list() {
    local name="$1"
    local require
    local replace

    pushd "${name}" >/dev/null

    go mod edit -json > mod.json

    if [ ! -f require.txt ]; then
        require=$(jq -r '.Require' mod.json)
        if [ "${require}" = "null" ]; then
            rm -f require.txt
            touch require.txt
        else
            jq -r '.Require[].Path' mod.json > require.txt
        fi
    fi

    if [ ! -f replace.txt ]; then
        replace=$(jq -r '.Replace' mod.json)
        if [ "${replace}" = "null" ]; then
            rm -f replace.txt
            touch replace.txt
        else
            jq -r '.Replace[] | .Old.Path + " " + .New.Path' mod.json > replace.txt
        fi
    fi

    popd >/dev/null
}

mkdir -p "${WORKDIR}"
pushd "${WORKDIR}" >/dev/null

# Process the MicroShift go.mod
mkdir -p "microshift"
cp "${REPOROOT}/go.mod" "microshift/"
build_dep_list "microshift"

# Determine the most recent release image from the last_rebase script
# and get the list of repos used to build it.
RELEASE_IMAGE=$(grep "rebase.sh" "${REPOROOT}/scripts/auto-rebase/last_rebase.sh" \
                    | awk '{print $3}' \
                    | sed -e 's|"||g')
echo "Updating dependencies from release image ${RELEASE_IMAGE} ..."
oc adm release info -a "${PULL_SECRET_FILE}" "${RELEASE_IMAGE}" -o json > release.json
# shellcheck disable=SC2002
cat release.json \
    | jq -r '.references.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location")"' \
    | grep -v '^$' \
    | sort -u \
           > source-repos

# Start with a clean list of dependencies
rm -f all_require.txt

# Fetch the go.mod files for the repos
# shellcheck disable=SC2002
sort -u source-repos | while read -r name repo_url; do
    if [ -z "${repo_url}" ]; then
        continue
    fi
    mkdir -p "${name}"
    echo "${repo_url}" > "${name}/url"

    get_go_mod "${name}" "${repo_url}"

    if [ ! -f "${name}/go.mod" ]; then
        continue
    fi

    build_dep_list "${name}"

    sed -e "s|^|${repo_url} |g" "${name}/require.txt" >> all_require.txt
done

sort -u microshift/require.txt | while read -r dep; do
    found_user=false

    echo
    echo "${dep}"
    if grep -q "${dep}" -- */url | sed -e 's|^|        |g'; then
        echo "	Is an OpenShift repo"
        found_user=true
    fi
    echo "    Also imported by"
    if grep " ${dep}\$" all_require.txt | cut -f1 -d' ' | sort -u | sed -e 's|^|        |g'; then
        found_user=true
    else
        echo "        Nothing"
    fi
    if ! ${found_user}; then
        echo "    No other users"
    fi
done
