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
    local -r ver=$1
    local -r dir=$2

    if ! action_access ; then
        echo "ERROR: Brew Hub site is not accessible"
        exit 1
    fi
    "${SCRIPTDIR}/../../scripts/fetch_tools.sh" brew

    # Attempt downloading the specified build version
    local package
    package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver}" | sort | tail -1) || true
    if [ -z "${package}" ] ; then
        echo "ERROR: Cannot find MicroShift '${ver}' packages in brew"
        exit 1
    fi

    package=$(awk '{print $1}' <<< "${package}")
    echo "Downloading '${package}' packages from brew"

    # Download all the supported architectures as the required architecture
    # cannot be identified easily when running in a CI job
    for arch in x86_64 aarch64 ; do
        local adir
        adir="${dir}/${arch}"

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
        [ $# -ne 3 ] && usage && exit 1
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
