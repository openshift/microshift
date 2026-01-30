#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Note: Avoid sourcing common.sh or common_version.sh in this script to allow
# its execution in a containerized environment with limited set of tools.

UNAME_M="${UNAME_M:-$(uname -m)}"

usage() {
    echo "Usage: $(basename "$0") [access | download <version> <path> [version_type]]"
    echo "  download:   Download the latest RPM packages for the given version, Z-1, Y-2 and Y-1 to the path as specified"
    echo "    - version: the X.Y version. Example: 4.19"
    echo "    - path: the output directory. Example: /_output/test-images/brew-rpms"
    echo "    - version_type: Optional. Valid values: rc, ec, zstream and nightly. Default: nightly"
    echo "  access:     Exit with non-zero status if brew cannot be accessed"
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

find_package() {
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
    set +x
    case ${ver_type} in
        zstream)
            package_list=$(sudo dnf repoquery --quiet --repo "rhocp-${ver_x}.${ver_y}-for-rhel-9-${UNAME_M}-rpms" 2>/dev/null) || true
            package_filtered=$(echo "${package_list}" | grep "microshift-0:" | sed 's/0://' | sed "s/.${UNAME_M}$//" | sort -V | uniq ) || true
            if [ -z "${package_filtered}" ] ; then
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
    set -x
    echo "${package}"
}

action_download() {
    local -r ver=$1
    local -r main_dir=$2
    local -r ver_type=${3:-nightly}

    local -a package_list=()

    # Find the packages to download
    # - package_latest:       Latest z-stream of current minor (for current release testing)
    # - package_prev:         Previous z-stream of current minor (for Z-1 upgrade testing)
    # - package_yminus1:      Latest z-stream of previous minor (for Y-1 upgrade testing)
    # - package_yminus2:      Latest z-stream of Y-2 minor (for Y-2 upgrade testing)
    case ${ver_type} in
        zstream)
            package_latest="$(find_package "${ver}" "${ver_type}" "0" "0")"
            package_prev="$(find_package "${ver}" "${ver_type}" "0" "1")"
            package_yminus1="$(find_package "${ver}" "${ver_type}" "1" "0")"
            package_yminus2="$(find_package "${ver}" "${ver_type}" "2" "0")"
            package_list=("${package_latest}" "${package_yminus1}" "${package_yminus2}" "${package_prev}")
            ;;
        nightly|rc|ec)
            package_latest="$(find_package "${ver}" "${ver_type}")"
            package_list=("${package_latest}")
            ;;
        *)
            echo "ERROR: Invalid version_type '${ver_type}'. Valid values are: rc, ec, zstream and nightly"
            exit 1
            ;;
    esac

    # Check at least one package is found
    local all_empty=true
    for package in "${package_list[@]}" ; do
        if [ -n "${package}" ] ; then
            all_empty=false
            break
        fi
    done
    if ${all_empty} ; then
        echo "ERROR: No packages found for ${ver} ${ver_type}"
        return 1
    fi

    # Download the packages
    for package in "${package_list[@]}" ; do
        if [ -z "${package}" ] ; then
            continue
        fi
        sub_dir=$(echo "${package}" | cut -d'-' -f2)
        if ! brew_cli_download "${package}" "${main_dir}" "${sub_dir}" "${ver_type}" ; then
            echo "ERROR: Failed to download package: ${package}"
            return 1
        fi
    done
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
        final_sub_dir=$(echo "${sub_dir}" | sed -E 's/(.*)(~)(.*)(rc|ec|nightly)(.*)/\1-\4/g')
    elif [ "${ver_type}" = "zstream" ] ; then
        final_sub_dir="${sub_dir}"
    fi

    # Download all the supported architectures as the required architecture
    # cannot be identified easily when running in a CI job
    for arch in x86_64 aarch64 ; do
        local adir
        adir="${main_dir}/${final_sub_dir}/${arch}"

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
