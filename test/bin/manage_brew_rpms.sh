#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Note: Avoid sourcing common.sh or common_version.sh in this script to allow
# its execution in a containerized environment with limited set of tools.

UNAME_M="${UNAME_M:-$(uname -m)}"

usage() {
    echo "Usage: $(basename "$0") [access | download | find_package <version> <path> [version_type] [ver_prev_y] [ver_prev_z]]"
    echo "  access:     Exit with non-zero status if brew cannot be accessed"
    echo "  download:   Download the RPM version to the path as specified"
    echo "    - version: the X.Y version. Example: 4.19"
    echo "    - path: the output directory. Example: /_output/test-images/brew-rpms"
    echo "    - version_type: Optional. Valid values: rc, ec, zstream and nightly. Default: nightly"
    echo "    - ver_prev_y: Optional. How far back from the current Y version to look for the previous Y version. Example: 2 (for 4.20 is 4.18). Default: 0"
    echo "    - ver_prev_z: Optional. How far back from the current Z version to look for the previous Z version. Example: 2 (for 4.20.3 is 4.20.1). Default: 0"
    echo "  find_package: Find the package version for the given version and version type"
    echo "    - version: the X.Y version. Example: 4.19"
    echo "    - version_type: Optional. Valid values: rc, ec, zstream and nightly. Default: nightly"
    echo "    - ver_prev_y: Optional. How far back from the current Y version to look for the previous Y version. Example: 2 (for 4.20 is 4.18). Default: 0"
    echo "    - ver_prev_z: Optional. How far back from the current Z version to look for the previous Z version. Example: 2 (for 4.20.3 is 4.20.1). Default: 0"
}

action_access() {
    local -r outfile=$(mktemp /tmp/curl-brewhub.XXXXXXXX)

    local rc=0
    if ! curl --retry 10 --retry-delay 5 --retry-max-time 120 --silent --show-error --head "https://brewhub.engineering.redhat.com" &> "${outfile}" ; then
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
    local -r ver_type=${2:-nightly}
    local -r ver_prev_y=${3:-0}
    local -r ver_prev_z=${4:-0}

    # Validate version format
    if ! [[ "${ver}" =~ ^4\.[0-9]+(\.[0-9]+)?$ ]]; then
        echo "ERROR: Version '${ver}' does not match required format 4.Y or 4.Y.Z (e.g., 4.20 or 4.20.1)"
        return 1
    fi

    # Validate optional numeric parameters
    if ! [[ "${ver_prev_y}" =~ ^[0-9]$ ]] || ! [[ "${ver_prev_z}" =~ ^[0-9]$ ]]; then
        echo "ERROR: ver_prev_y '${ver_prev_y}' and ver_prev_z '${ver_prev_z}' must be a non-negative integer and less than 10"
        return 1
    fi

    # Extract version components
    ver_x=$(echo "${ver}" | cut -d'.' -f1)
    ver_y=$(echo "${ver}" | cut -d'.' -f2)

    # Calculate previous X.Y version
    ver_y=$((ver_y - ver_prev_y))

    local package=""
    case ${ver_type} in
        zstream)
            package_list=$(sudo dnf repoquery --quiet --repo "rhocp-${ver_x}.${ver_y}-for-rhel-9-${UNAME_M}-rpms" 2>/dev/null) || true
            package_filtered=$(echo "${package_list}" | grep "microshift-0:" | sed 's/0://' | sed "s/.${UNAME_M}$//" | sort -V | uniq ) || true
            if [ -z "${package}" ] ; then
                package_list=$(brew list-builds --quiet --package=microshift --state=COMPLETE 2>/dev/null) || true
                package_filtered=$(echo "${package_list}" | grep "^microshift-${ver_x}.${ver_y}" | grep -v "~" | sort -V | uniq ) || true
            fi
            package=$(echo "${package_filtered}" | tail -n$((1 + ver_prev_z)) | head -n1 | awk '{print $1}') || true
            ;;
        nightly)
            package_list=$(brew list-builds --quiet --package=microshift --state=COMPLETE 2>/dev/null  ) || true
            package_filtered=$(echo "${package_list}" | grep "^microshift-${ver_x}.${ver_y}" | grep "nightly" | sort -V | uniq ) || true
            package=$(echo "${package_filtered}" | tail -n1 | awk '{print $1}') || true
            ;;
        rc|ec)
            package_list=$(brew list-builds --quiet --package=microshift --state=COMPLETE 2>/dev/null ) || true
            package_filtered=$(echo "${package_list}" | grep "^microshift-${ver_x}.${ver_y}.0~${ver_type}." | sort -V | uniq ) || true
            package=$(echo "${package_filtered}" | tail -n1 | awk '{print $1}') || true
            ;;
        *)
            echo "ERROR: Invalid version_type '${ver_type}'. Valid values are: rc, ec, zstream and nightly"
            exit 1
            ;;
    esac

    echo "${package}"
}

action_download() {
    local -r ver=$1
    local -r main_dir=$2
    local -r ver_type=$3
    local -r ver_prev_y=$4
    local -r ver_prev_z=$5

    local package
    package="$(action_find_package "${ver}" "${ver_type}" "${ver_prev_y}" "${ver_prev_z}")"
    if [ -z "${package}" ] ; then
        echo "ERROR: Package not found: ${ver} ${ver_type} ${ver_prev_y} ${ver_prev_z}"
        return 1
    fi

    local -r sub_dir=$(echo "${package}" | cut -d'-' -f2)
    if ! brew_cli_download "${package}" "${main_dir}" "${sub_dir}" "${ver_type}" ; then
        echo "ERROR: Failed to download package: ${package}"
        return 1
    fi
}

brew_cli_download() {
    local -r package=$1
    local -r main_dir=$2
    local -r sub_dir=$3
    local -r ver_type=$4

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

    # Format sub_dir for EC, RC and nightly
    if [ "${ver_type}" = "ec" ] || [ "${ver_type}" = "rc" ] || [ "${ver_type}" = "nightly" ] ; then
        sub_dir=$(echo "${sub_dir}" | sed -E 's/(.*)(\.0~)(rc|ec|nightly)(.*)/\1-\3/g')
    fi

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
    find_package)
        [ $# -ne 3 ] && [ $# -ne 4 ] && [ $# -ne 5 ] && usage && exit 1
        shift
        action_"${action}" "$@"
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
