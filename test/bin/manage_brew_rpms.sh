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
    echo "    - version_type: Optional. Valid values: rc, ec and zstream. Default: ec"
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
    local -r ver_type=${3:-ec}

    if [ -z "${ver}" ] || [ -z "${dir}" ] ; then
        echo "ERROR: At least two parameters (version and path) are required"
        exit 1
    fi

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
            package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver}.0~${ver_type}." | tail -n1) || true
            ;;
        *)
            echo "ERROR: Invalid version_type '${ver_type}'. Valid values are: rc, ec and zstream"
            exit 1
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
        adir="${dir}/${ver}-${ver_type}/${arch}"

        mkdir -p "${adir}"
        pushd "${adir}" &>/dev/null
        if ! brew download-build --arch="${arch}" --arch="noarch" "${package}" ; then
            echo "WARNING: Failed to download '${package}' packages using brew download-build command, using curl as a fallback mechanism"
            if ! brew_curl_download "${package}" "${arch}" ; then
                echo "ERROR: Failed to download '${package}' packages using curl command"
                popd &>/dev/null
                exit 1
            fi
        fi
        popd &>/dev/null
    done
}

brew_curl_download() {
    local package=$1
    local arch=$2

    # Parse package to extract version and build release
    local version_and_release="${package#microshift-}"
    local pkg_version="${version_and_release%%-*}"
    local pkg_release="${version_and_release#*-}"

    for current_arch in ${arch} noarch; do
        local base_url="http://download.devel.redhat.com/rhel-9/brew/packages/microshift/${pkg_version}/${pkg_release}/${current_arch}/"

        local rpm_files
        rpm_files=$(curl -k -s "${base_url}" | sed -n 's/.*href="\([^"]*\.rpm\)".*/\1/p') || true
        if [ -z "${rpm_files}" ]; then
            echo "ERROR: No RPM files found at ${base_url}"
            return 1
        fi

        echo "Downloading from: ${base_url}"
        for rpm_file in ${rpm_files}; do
            echo "Downloading: ${rpm_file}"
            if ! curl -s -O "${base_url}${rpm_file}"; then
                echo "ERROR: Failed to download ${rpm_file}"
                return 1
            fi
        done
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
