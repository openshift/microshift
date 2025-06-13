#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Note: Avoid sourcing common.sh or common_version.sh in this script to allow
# its execution in a containerized environment with limited set of tools.

usage() {
    echo "Usage: $(basename "$0") [access | download <version> <path>]"
    echo "  download:   Download the RPM version to the path as specified"
    echo "  access:     Exit with non-zero status if brew cannot be accessed"
}

action_access() {
    local -r outfile=$(mktemp /tmp/curl-brewhub.XXXXXXXX)

    local rc=0
    if ! curl --silent --show-error --head "https://brewhub.engineering.redhat.com" &> "${outfile}" ; then
        rc=1
        # Display the error in case the site is not accessible.
        # This is useful to rule out certificate problems, etc.
        cat "${outfile}"
    fi

    rm -f "${outfile}"
    return ${rc}
}

action_download() {
    local -r version=$1
    local -r version_type=$2
    local -r dir=$3
    local -r num_versions_back=${4:-0}

    if ! action_access ; then
        echo "ERROR: Brew Hub site is not accessible"
        exit 1
    fi
    "${SCRIPTDIR}/../../scripts/fetch_tools.sh" brew

    # Attempt downloading the specified build version
    local package
    if [ "${version_type}" = "zstream" ]; then
        package_found=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${version}" | grep -v "~" | awk -F - '{print $1"-"$2}' | uniq | tail -n "$(( "${num_versions_back}" + 1))" | head -n1) || true
        package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "${package_found}-" | tail -1) || true
    else
        package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${version}.*${version_type}" | tail -1) || true
    fi

    if [ -z "${package}" ] ; then
        echo "WARNING: Cannot find MicroShift '${version}' packages in brew"
        exit 0
    fi

    package=$(awk '{print $1}' <<< "${package}")
    echo "Downloading '${package}' packages from brew"

    # Download all the supported architectures as the required architecture
    # cannot be identified easily when running in a CI job
    for arch in x86_64 aarch64 ; do
        local adir
        # shellcheck disable=SC2001
        adir="${dir}/$(echo "${package}" | sed 's/.*microshift-\([^-]*\).*/\1/')/${arch}"

        mkdir -p "${adir}"
        pushd "${adir}" &>/dev/null
        brew download-build --arch="${arch}" --arch="noarch" "${package}"
        popd &>/dev/null
    done
}

#
# Main
#
if [ $# -lt 1 ] ; then
    usage
    exit 1
fi
action="${1}"

case "${action}" in
    access)
        [ $# -ne 1 ] && usage && exit 1
        "action_${action}"
        ;;
    download)
        [ $# -gt 5 ] && usage && exit 1
        shift
        "action_${action}" "$@"
        ;;
    -h)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
