#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Note: Avoid sourcing common.sh or common_version.sh in this script to allow
# its execution in a containerized environment with limited set of tools.

usage() {
    echo "Usage: $(basename "$0") [access | download <version> <path> [version_type] [ver_prev_y] [ver_prev_z]]"
    echo "  download:   Download the RPM version to the path as specified"
    echo "    - version: the X.Y version. Example: 4.19"
    echo "    - path: the output directory. Example: /_output/test-images/brew-rpms"
    echo "    - version_type: Optional. Valid values: rc, ec and zstream. Default: ec"
    echo "    - ver_prev_y: Optional. How far back from the current Y version to look for the previous Y version. Example: 2 (for 4.20 is 4.18)"
    echo "    - ver_prev_z: Optional. How far back from the current Z version to look for the previous Z version. Example: 2 (for 4.20.3 is 4.20.1)"
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

action_find_package() {
    local -r ver=$1
    local -r main_dir=$2
    local -r ver_type=${3:-ec}
    local -r ver_prev_y=${4:-0}
    local -r ver_prev_z=${5:-0}

    # Validate version format
    if ! [[ "${ver}" =~ ^4\.[0-9]+(\.[0-9]+)?$ ]]; then
        echo "ERROR: Version '${ver}' does not match required format 4.Y or 4.Y.Z (e.g., 4.20 or 4.20.1)"
        return 1
    fi

    # Validate optional numeric parameters
    if [ "${ver_prev_y}" -gt 0 ] && ! [[ "${ver_prev_y}" =~ ^[0-9]$ ]]; then
        echo "ERROR: ver_prev_y '${ver_prev_y}' must be a non-negative integer and less than 10"
        return 1
    fi
    if [ "${ver_prev_z}" -gt 0 ] && ! [[ "${ver_prev_z}" =~ ^[0-9]$ ]]; then
        echo "ERROR: ver_prev_z '${ver_prev_z}' must be a non-negative integer and less than 10"
        return 1
    fi

    # Extract version components
    ver_x=$(echo "${ver}" | cut -d'.' -f1)
    ver_y=$(echo "${ver}" | cut -d'.' -f2)
    ver_z=$(echo "${ver}" | cut -d'.' -f3)

    # Calculate previous X.Y version
    ver_y=$((ver_y - ver_prev_y))

    local package
    case ${ver_type} in
        zstream)
            package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver_x}.${ver_y}" | grep -v "~" | uniq) || true
            ;;
        rc|ec)
            package=$(brew list-builds --quiet --package=microshift --state=COMPLETE | grep "^microshift-${ver_x}.${ver_y}.0~${ver_type}." | uniq) || true
            ;;
        *)
            echo "ERROR: Invalid version_type '${ver_type}'. Valid values are: rc, ec and zstream"
            return 1
            ;;
    esac

    # If the previous version is Z-1 or Z-2, we need to find the previous package
    package=$(echo "${package}" | tail -n$((1 + ver_prev_z)) | head -n1 | awk '{print $1}')

    if [ -z "${package}" ] ; then
        echo "ERROR: Cannot find MicroShift '${ver_x}.${ver_y}' packages in brew"
        return 1
    fi

    echo "${package}"
}

action_download() {
    local package
    if ! package=$(action_find_package "$@" 2>&1); then
        echo "${package}"
        return 1
    fi

    local -r main_dir="${2}"
    local -r sub_dir=$(echo "${package}" | cut -d'-' -f2)
    if ! brew_cli_download "${package}" "${main_dir}" "${sub_dir}" 2>&1; then
        return 1
    fi
}

brew_cli_download() {
    local -r package=$1
    local -r main_dir=$2
    local -r sub_dir=$3

    # Validate parameters
    if [ -z "${package}" ] ; then
        echo "ERROR: Package is required"
        exit 1
    fi

    if [ -z "${main_dir}" ] ; then
        echo "ERROR: Main directory is required"
        exit 1
    fi

    if [ -z "${sub_dir}" ] ; then
        echo "ERROR: Sub directory is required"
        exit 1
    fi
    # Check if brew is accessible
    if ! action_access ; then
        echo "ERROR: Brew Hub site is not accessible"
        exit 1
    fi
    "${SCRIPTDIR}/../../scripts/fetch_tools.sh" brew

    echo "Downloading '${package}' packages from brew"

    # Download all the supported architectures as the required architecture
    # cannot be identified easily when running in a CI job
    for arch in x86_64 aarch64 ; do
        local adir
        adir="${main_dir}/${sub_dir}/${arch}"

        if ! mkdir -p "${adir}" ; then
            echo "ERROR: Failed to create directory '${adir}'"
            exit 1
        fi
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
        local base_url="https://download-01.beak-001.prod.iad2.dc.redhat.com/rhel-9/brew/packages/microshift/${pkg_version}/${pkg_release}/${current_arch}/"

        local rpm_files
        rpm_files=$(curl -k -s "${base_url}" | sed -n 's/.*href="\([^"]*\.rpm\)".*/\1/p') || true
        if [ -z "${rpm_files}" ]; then
            echo "ERROR: No RPM files found at ${base_url}"
            return 1
        fi

        echo "Downloading from: ${base_url}"
        for rpm_file in ${rpm_files}; do
            echo "Downloading: ${rpm_file}"
            if ! curl -k -s -O "${base_url}${rpm_file}"; then
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
    find_package)
        [ $# -ne 3 ] && [ $# -ne 4 ] && [ $# -ne 5 ] && [ $# -ne 6 ] && usage && exit 1
        shift
        args=("$1" "/dev/null" "$2" "$3" "$4")
        action_"${action}" "${args[@]}"
        ;;
    download)
        [ $# -ne 3 ] && [ $# -ne 4 ] && [ $# -ne 5 ] && [ $# -ne 6 ] && usage && exit 1
        shift
        action_"${action}" "$@"
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
