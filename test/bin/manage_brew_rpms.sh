#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Note: Avoid sourcing common.sh or common_version.sh in this script to allow
# its execution in a containerized environment with limited set of tools.

usage() {
    echo "Usage: $(basename "$0") [access | download <version> <path> [version_type]]"
    echo "  download:   Download the RPM version to the path as specified"
    echo "    - version: the X.Y version. Example: 4.19"
    echo "    - path: the output directory. Example: /_output/test-images/brew-rpms"
    echo "    - version_type: optional version type. Valid values: rc, ec, zstream and nightly."
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
    local -r ver_type=${3:-}

    if ! action_access ; then
        echo "ERROR: Brew Hub site is not accessible"
        exit 1
    fi
    "${SCRIPTDIR}/../../scripts/fetch_tools.sh" brew

    # Attempt downloading the specified build version
    local package
    case ${ver_type} in 
        zstream)
            package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver}" | grep -v "~" | uniq | tail -n1) || true
            ;;
        rc|ec)
            package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver}.0~${ver_type}." | tail -1) || true
            ;;
        *)
            package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver}.*${ver_type}" | tail -1) || true
            ;;
    esac

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
        # shellcheck disable=SC2001
        if [ -z "${ver_type}" ] ; then
            adir="${dir}/${arch}"
        else
            # remove date and commit id from package name and use it in dir name
            version=$(echo "${package}" | sed 's/.*microshift-\([^-]*\).*/\1/')
            adir="${dir}/${version}/${arch}"
        fi

        mkdir -p "${adir}"
        pushd "${adir}" &>/dev/null
        if ! brew download-build --arch="${arch}" --arch="noarch" "${package}" ; then
            echo "ERROR: Failed to download '${package}' packages from brew"
            exit 1
        fi
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
        [ $# -ne 3 ] && [ $# -ne 4 ] && usage && exit 1
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
