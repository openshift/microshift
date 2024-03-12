#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

# If a glob pattern does not match anything, return a null value
# instead of the pattern. This ensures for loops over files do not
# produce errors when a group directory does not include any matching
# files.
shopt -s nullglob

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"
# shellcheck source=test/bin/get_rel_version_repo.sh
source "${SCRIPTDIR}/get_rel_version_repo.sh"

SKIP_LOG_COLLECTION=${SKIP_LOG_COLLECTION:-false}

osbuild_logs() {
    if ${SKIP_LOG_COLLECTION}; then
        return
    fi
    workers_services=$(sudo systemctl list-units | awk '/osbuild-worker@/ {print $1} /osbuild-composer\.service/ {print $1}')
    for service in ${workers_services}; do
        # shellcheck disable=SC2024  # redirect and sudo
        sudo journalctl -u "${service}" &> "${LOGDIR}/${service}.log"
    done
}

extract_container_images() {
    local -r version=$1    # full version
    local -r repo_spec=$2  # repo name, path, or URL
    local -r outfile=$3    # destination file

    echo "Extracting images from ${version}"
    mkdir -p "${IMAGEDIR}/release-info-rpms"
    pushd "${IMAGEDIR}/release-info-rpms"
    dnf_options=""
    local -r repo_name="$(basename "${repo_spec}")"
    if [[ "${repo_spec}" =~ ^https://.* ]]; then
        # If the spec is a URL, set up the arguments to point to that location.
        dnf_options="--repofrompath ${repo_name},${repo_spec} --repo ${repo_name}"
    elif [[ "${repo_spec}" =~ ^/.* ]]; then
        # If the spec is a path, set up the arguments to point to that path.
        dnf_options="--repofrompath ${repo_name},${repo_spec} --repo ${repo_name}"
    elif [[ -n ${repo_spec} ]]; then
        # If the spec is a name, assume it is already known to the
        # system through normal configuration. The repo does not need
        # to be enabled in order for dnf to download a package from
        # it.
        dnf_options="--repo ${repo_spec}"
    fi
    # shellcheck disable=SC2086  # double quotes
    sudo dnf download ${dnf_options} microshift-release-info-"${version}"
    get_container_images "${version}" "${IMAGEDIR}/release-info-rpms" | sed 's/,/\n/g' >> "${outfile}"
    sudo rm -f microshift-release-info-*.rpm
    popd
}

configure_package_sources() {
    ## TEMPLATE VARIABLES
    export UNAME_M                 # defined in common.sh
    export LOCAL_REPO              # defined in common.sh
    export NEXT_REPO               # defined in common.sh
    export BASE_REPO               # defined in common.sh
    export CURRENT_RELEASE_REPO
    export PREVIOUS_RELEASE_REPO

    export SOURCE_VERSION
    export FAKE_NEXT_MINOR_VERSION
    export MINOR_VERSION
    export PREVIOUS_MINOR_VERSION
    export YMINUS2_MINOR_VERSION
    export SOURCE_VERSION_BASE
    export CURRENT_RELEASE_VERSION
    export PREVIOUS_RELEASE_VERSION
    export YMINUS2_RELEASE_VERSION
    export RHOCP_MINOR_Y
    export RHOCP_MINOR_Y1
    export RHOCP_MINOR_Y2

    # Add our sources. It is OK to run these steps repeatedly, if the
    # details change they are updated in the service.
    title "Expanding package source templates to ${IMAGEDIR}/package-sources"
    mkdir -p "${IMAGEDIR}/package-sources"
    for template in "${TESTDIR}"/package-sources/*.toml; do
        name=$(basename "${template}" .toml)
        outfile="${IMAGEDIR}/package-sources/${name}.toml"

        echo "Rendering ${template} to ${outfile}"
        ${GOMPLATE} --file "${template}" >"${outfile}"
        if [[ "$(wc -l "${outfile}" | cut -d ' ' -f1)" -eq 0 ]]; then
            echo "WARNING: Templating '${template}' resulted in empty file! - SKIPPING"
            continue
        fi

        echo "Adding package source from ${outfile}"
        if sudo composer-cli sources list | grep "^${name}\$"; then
            sudo composer-cli sources delete "${name}"
        fi
        sudo composer-cli sources add "${outfile}"
    done

    # Show details about the available sources to make debugging easier.
    for name in $(sudo composer-cli sources list); do
        echo
        echo "Package source: ${name}"
        sudo composer-cli sources info "${name}" | sed -e 's/gpgkeys.*/gpgkeys = .../g'
    done
}

# Reads release-info RPM for provided version to obtain images
# and returns them as comma-separated list.
get_container_images() {
    local -r version="${1}"
    local -r path="${2}"

    # Find the microshift-release-info RPM with the specified version
    local -r release_info_rpm=$(find "${path}" -name "microshift-release-info-${version}*.rpm" | sort | tail -1)
    if [ -z "${release_info_rpm}" ] ; then
        echo "Error: missing microshift-release-info RPM for the '${version}' version"
        exit 1
    fi

    # Extract list of image URIs and join them with a comma
    rpm2cpio "${release_info_rpm}" | cpio  -i --to-stdout "*release-$(uname -m).json" 2> /dev/null | jq -r '[ .images[] ] | join(",")'
}

# Given a blueprint filename, extract the name value. It does not have
# to match the filename, but some commands take the file and others
# take the name, so we need to be able to have both.
get_blueprint_name() {
    local filename="${1}"
    tomcli-get "${filename}" name
}

# Given a blueprint filename, extract the parent blue filename from
# the prefix and use that to find the actual blueprint name that
# composer knows.
#
# rhel92-microshift-source -> rhel-9.2
#
# FIXME: We may need to change the prefix separator in the future if
# we need a multi-level hierarchy.
get_image_parent() {
    local blueprint_filename="$1"

    local base
    base=$(basename "${blueprint_filename}" .toml)
    if [[ "${base}" =~ '-' ]]; then
        base="${base//-*/}"

        local name
        name=$(find "${TESTDIR}/image-blueprints" -name "${base}.toml")
        if [ -n "${name}" ] ; then
            get_blueprint_name "${name}"
        else
            echo ""
        fi
    else
        echo ""
    fi
}

start_junit() {
    local groupdir="$1"

    local group
    group="$(basename "${groupdir}")"

    mkdir -p "${LOGDIR}/${group}"
    local outputfile="${LOGDIR}/${group}/junit.xml"

    echo "Creating ${outputfile}"

    cat - >"${outputfile}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="build-images:${group}" timestamp="$(date --iso-8601=ns)">
EOF
}

close_junit() {
    local groupdir="$1"

    local group
    group="$(basename "${groupdir}")"
    local outputfile="${LOGDIR}/${group}/junit.xml"

    echo '</testsuite>' >>"${outputfile}"
}

record_junit() {
    local groupdir="$1"
    local image_reference="$2"
    local step="$3"
    local results="$4"

    local group
    group="$(basename "${groupdir}")"
    local outputfile="${LOGDIR}/${group}/junit.xml"

    cat - >>"${outputfile}" <<EOF
<testcase classname="${group} ${image_reference}" name="${step}">
EOF

    case "${results}" in
        OK)
        ;;
        SKIP*)
        cat - >>"${outputfile}" <<EOF
<skipped message="${results}" type="${step}-skipped" />
EOF
        ;;
        FAIL*)
        cat - >>"${outputfile}" <<EOF
<failure message="${results}" type="${step}-failure" />
EOF
    esac

    cat - >>"${outputfile}" <<EOF
</testcase>
EOF

}

should_skip() {
    local -r blueprint="${1}"

    if "${FORCE_REBUILD}"; then
        echo "Forcing all rebuilds"
        return 1
    fi
    if [[ "${blueprint}" =~ source ]] && "${FORCE_SOURCE}"; then
        echo "Forcing source rebuild"
        return 1
    fi
    return 0
}

# Process a set of blueprint templates to create edge commit images
# for them.
do_group() {
    local groupdir="$1"
    local -r template_arg="$2"

    local -r ip_addr_default=$(hostname -I | awk '{print $1}')

    title "Building ${groupdir}"

    start_junit "${groupdir}"
    trap 'close_junit ${groupdir}' RETURN

    local blueprint
    local blueprint_file
    local build_name
    local buildid
    local buildid_list=()
    local download_opts=()
    local builds_to_get=""
    local parent
    local parent_args
    local template
    local template_list

    SOURCE_IMAGES="$(get_container_images "${SOURCE_VERSION}" "${IMAGEDIR}/rpm-repos")"
    export SOURCE_IMAGES

    echo "Existing images:"
    ls "${VM_DISK_BASEDIR}"/*.iso || echo "No ISO files in ${VM_DISK_BASEDIR}"
    ostree summary --view --repo="${IMAGEDIR}/repo" || echo "Could not get image list from ${IMAGEDIR}/repo"

    # Upload the blueprint definitions
    if [ -n "${template_arg}" ]; then
        template_list="${template_arg}"
    else
        echo "Expanding blueprint templates to ${IMAGEDIR}/blueprints and starting edge-commit builds"
        if ! ${ONLY_SOURCE}; then
            template_list=$(echo "${groupdir}"/*.toml)
        else
            template_list=$(echo "${groupdir}"/*source*.toml)
        fi
    fi
    for template in ${template_list}; do
        echo
        echo "Blueprint ${template}"

        blueprint_file="${IMAGEDIR}/blueprints/$(basename "${template}")"

        # Check for the file to exist, in case the user passed a
        # template on the command line.
        if [ ! -f "${template}" ]; then
            echo "ERROR: Template ${template} does not exist"
            record_junit "${groupdir}" "${template}" "render" "FAILED"
            return 1
        fi

        echo "Rendering ${template} to ${blueprint_file}"
        ${GOMPLATE} --file "${template}" >"${blueprint_file}"
        if [[ "$(wc -l "${blueprint_file}" | cut -d ' ' -f1)" -eq 0 ]]; then
            echo "WARNING: Templating '${template}' resulted in empty file! - SKIPPING"
            record_junit "${groupdir}" "${template}" "compose" "SKIPPED"
            continue
        fi
        record_junit "${groupdir}" "${template}" "render" "OK"

        blueprint=$(get_blueprint_name "${blueprint_file}")

        # Check if the image for this blueprint already exists, in
        # case it was downloaded from the cache.
        if ostree summary --view --repo="${IMAGEDIR}/repo" | grep -q " ${blueprint}\$"; then
            echo "Found ${blueprint} in existing images"
            if should_skip "${blueprint}"; then
                record_junit "${groupdir}" "${template}" "compose" "SKIPPED"
                continue
            fi
        fi

        if sudo composer-cli blueprints list | grep -q "^${blueprint}$"; then
            echo "Removing existing definition of ${blueprint}"
            sudo composer-cli blueprints delete "${blueprint}"
        fi

        echo "Loading new definition of ${blueprint}"
        sudo composer-cli blueprints push "${blueprint_file}"

        echo "Resolving dependencies for ${blueprint}"
        # shellcheck disable=SC2024  # redirect and sudo
        if sudo composer-cli blueprints depsolve "${blueprint}" \
                >"${LOGDIR}/${blueprint}-depsolve.log" 2>&1; then
            record_junit "${groupdir}" "${blueprint}" "depsolve" "OK"
        else
            record_junit "${groupdir}" "${blueprint}" "depsolve" "FAILED"
        fi

        if ${COMPOSER_DRY_RUN} ; then
            echo "Skipping the composer start operation"
            continue
        fi

        parent_args=""
        parent=$(get_image_parent "${template}")
        if [ -n "${parent}" ]; then
            parent_args="--parent ${parent} --url http://${ip_addr_default}:${WEB_SERVER_PORT}/repo"
        fi
        echo "Building edge-commit from ${blueprint} ${parent_args}"
        # shellcheck disable=SC2086  # quote to avoid glob expansion
        build_cmd="sudo composer-cli compose start-ostree ${parent_args} --ref ${blueprint} ${blueprint} edge-commit"
        for _ in $(seq 3); do
            set +e
            build_cmd_output=$(${build_cmd})
            rc=$?
            set -e
            if [[ "${rc}" -eq 0 ]]; then
                buildid=$(echo "${build_cmd_output}" | awk '{print $2}')
                break
            fi
            sleep 15
        done

        if [[ "${rc}" -ne 0 ]]; then
            echo "Command failed consistently: ${build_cmd}"
            exit 1
        fi

        echo "Build ID ${buildid}"
        # Record a "build name" to be used as part of the unique
        # filename for the log we download next.
        echo "${blueprint}-edge-commit" > "${IMAGEDIR}/builds/${buildid}.build"
        buildid_list+=("${buildid},${build_cmd}")
    done

    if ${BUILD_INSTALLER} && ! ${COMPOSER_DRY_RUN}; then
        for image_installer in "${groupdir}"/*.image-installer; do
            # If a template arg was given, only build the image installer for the
            # matching template.
            if [ -n "${template_arg}" ]; then
                installer_file=$(basename "${image_installer}")
                template_file=$(basename "${template_arg}")
                if [ "${installer_file%.image-installer}" != "${template_file%.toml}" ]; then
                    continue
                fi
            fi
            blueprint=$("${GOMPLATE}" --file "${image_installer}")
            local expected_iso_file="${VM_DISK_BASEDIR}/${blueprint}.iso"
            if [ -f "${expected_iso_file}" ]; then
                echo "${expected_iso_file} already exists"
                if should_skip "${blueprint}"; then
                    record_junit "${groupdir}" "${image_installer}" "compose" "SKIPPED"
                    continue
                fi
            fi
            echo "Building image-installer from ${blueprint}"
            build_cmd="sudo composer-cli compose start ${blueprint} image-installer"
            buildid=$(${build_cmd} | awk '{print $2}')
            echo "Build ID ${buildid}"
            # Record a "build name" to be used as part of the unique
            # filename for the log we download next.
            echo "${blueprint}-image-installer" > "${IMAGEDIR}/builds/${buildid}.build"
            buildid_list+=("${buildid},${build_cmd}")
        done
    fi

    if ${BUILD_INSTALLER} && ! ${COMPOSER_DRY_RUN}; then
        for download_file in "${groupdir}"/*.image-fetcher; do
            local download_url
            download_url=$("${GOMPLATE}" --file "${download_file}")
            blueprint="$(basename -s .image-fetcher "${download_file}")"
            local expected_iso_file="${VM_DISK_BASEDIR}/${blueprint}.iso"
            if [ -f "${expected_iso_file}" ]; then
                echo "${expected_iso_file} already exists"
                if should_skip "${blueprint}"; then
                    record_junit "${groupdir}" "${download_file}" "download" "SKIPPED"
                    continue
                fi
            fi

            echo "Adding image-fetcher for ${download_file}"
            download_opts+=("${expected_iso_file} ${download_url}")
        done
    fi

    # Run image-fetcher while osbuilder is running in background
    if [ ${#download_opts[@]} -ne 0 ]; then
        local -r wget_tmp="wget.part"
        local -r wget_res="${IMAGEDIR}/image_fetcher_result.json"
        local -r wget_job="${IMAGEDIR}/image_fetcher_jobs.txt"

        local fetch_ok=true
        local progress=""
        if [ -t 0 ]; then
            progress="--progress"
        fi
        # Download the files under temporary names
        echo "Waiting for image-fetcher to complete..."
        if parallel \
            ${progress} \
            --colsep ' ' \
            --results "${wget_res}" \
            --joblog "${wget_job}" \
            --jobs $(( ${#download_opts[@]} / 2 )) \
            wget -c -nv -O "{1}.${wget_tmp}" "{2}" ::: "${download_opts[@]}" ; then
            # On successful download, rename the files to their original names
            for fwget in "${VM_DISK_BASEDIR}"/*."${wget_tmp}" ; do
                local forig
                forig="$(basename -s ".${wget_tmp}" "${fwget}")"
                mv "${fwget}" "${VM_DISK_BASEDIR}/${forig}"
            done
        else
            fetch_ok=false
        fi

        # Show the summary of the output of the parallel jobs.
        cat "${wget_job}"
        if [ -f "${wget_res}" ] ; then
            jq < "${wget_res}"
        else
            echo "The image-fetcher results file does not exist"
            fetch_ok=false
        fi

        if ! ${fetch_ok} ; then
            echo "ERROR: The image-fetcher failed to complete successfully"
            exit 1
        fi
    fi
    if [ ${#buildid_list[@]} -ne 0 ]; then
        echo "Waiting for builds to complete..."
        # wait_images.py returns possibly updated list of builds that must be handled
        # "update" means replacing initial build ID with retry build ID
        builds_to_get=$(time "${SCRIPTDIR}/wait_images.py" "${buildid_list[@]}")
    fi

    echo "Downloading build logs, metadata, and image"
    cd "${IMAGEDIR}/builds"

    failed_builds=()
    # shellcheck disable=SC2231  # allow glob expansion without quotes in for loop
    for buildid in ${builds_to_get}; do
        # shellcheck disable=SC2086  # pass glob args without quotes
        rm -f ${buildid}-*.tar

        sudo composer-cli compose logs "${buildid}"
        # shellcheck disable=SC2086  # pass glob args without quotes
        sudo chown "$(whoami)." ${buildid}-*

        # The log tar file contains 1 log file. Extract that file and
        # move it to the log directory with a unique name.
        tar xf "${buildid}-logs.tar"
        build_name=$(cat "${buildid}.build")
        mv logs/osbuild.log "${LOGDIR}/osbuild-${build_name}-${buildid}.log"

        # Skip the remaining steps for anything that has a status that
        # is not finished (failed, canceled, etc.).
        status=$(sudo composer-cli compose status | grep "${buildid}" | awk '{print $2}')
        if [ "${status}" != "FINISHED" ]; then
            failed_builds+=("${buildid}")
            record_junit "${groupdir}" "${build_name}" "compose" "${status}"
            sudo composer-cli compose info --json "${buildid}"
            sudo composer-cli compose log "${buildid}"
            continue
        fi

        sudo composer-cli compose metadata "${buildid}"
        sudo composer-cli compose image "${buildid}"
        # shellcheck disable=SC2086  # pass glob args without quotes
        sudo chown "$(whoami)." ${buildid}-*

        if [[ "${build_name}" =~ edge-commit ]]; then
            commit_file="${buildid}-commit.tar"
            echo "Unpacking ${commit_file} ${build_name}"
            tar -C "${IMAGEDIR}" -xf "${commit_file}"
        elif [[ "${build_name}" =~ image-installer ]]; then
            blueprint=${build_name//-image-installer/}
            iso_file="${buildid}-installer.iso"
            echo "Moving ${iso_file} to ${VM_DISK_BASEDIR}/${blueprint}.iso"
            mv -f "${iso_file}" "${VM_DISK_BASEDIR}/${blueprint}.iso"
        else
            echo "Do not know how to handle build ${build_name}"
        fi

        record_junit "${groupdir}" "${build_name}" "compose" "OK"
    done

    # Exit the function on build errors
    if [ ${#failed_builds[@]} -ne 0 ] ; then
        echo "Error: check the failed build jobs"
        echo "${failed_builds[@]}"
        return 1
    fi

    cd "${IMAGEDIR}"
    echo "Updating ostree references in ${IMAGEDIR}/repo before adding aliases"
    ostree summary --update --repo=repo
    ostree summary --view --repo=repo

    for alias_file in "${groupdir}"/*.alias; do
        alias_name=$(basename "${alias_file}" .alias)
        point_to=$(cat "${alias_file}")
        echo "Creating image reference alias ${alias_name} -> ${point_to}"
        if (cd "${IMAGEDIR}" &&
                ostree refs --repo=repo --force \
                       --create "${alias_name}" "${point_to}"); then
            record_junit "${groupdir}" "${alias_name}" "alias" "OK"
        else
            record_junit "${groupdir}" "${alias_name}" "alias" "FAILED"
        fi
    done

    cd "${IMAGEDIR}"
    echo "Updating ostree references in ${IMAGEDIR}/repo"
    ostree summary --update --repo=repo
    ostree summary --view --repo=repo
}

is_rhocp_available() {
    local -r ver="${1}"
    repository="rhocp-4.${ver}-for-rhel-9-$(uname -m)-rpms"
    if sudo dnf -v repository-packages "${repository}" info cri-o 1>&2; then
        return 0
    fi
    return 1
}

usage() {
    if [ $# -gt 0 ] ; then
        echo "ERROR: $*"
        echo
    fi

    cat - <<EOF
build_images.sh [-iIsdf] [-l layer-dir | -g group-dir] [-t template]

  -d      Dry run by skipping the composer start commands.

  -E      Do not extract container images.

  -f      Force rebuilding images that already exist.

  -g DIR  Build only one group (cannot be used with -l or -t).
          The DIR should be the path to the group to build.
          Implies -l based on the path.

  -h      Show this help

  -i      Build the installer image(s).

  -I      Do not build the installer image(s).

  -l DIR  Build only one layer (cannot be used with -g or -t).
          The DIR should be the path to the layer to build.

  -s      Only build source images (implies -I). Ignores cached images.

  -S      Skip collecting builder logs when there is a failure. Speeds
          up local development cycle.

  -t FILE Build only one template (cannot be used with -l or -g).
          The FILE should be the path to the template to build.
          Implies -f along with -l and -g based on the filename.

EOF
}

BUILD_INSTALLER=true
ONLY_SOURCE=false
COMPOSER_DRY_RUN=false
LAYER=""
GROUP=""
TEMPLATE=""
FORCE_REBUILD=false
FORCE_SOURCE=false
EXTRACT_CONTAINER_IMAGES=true

selCount=0
while getopts "dEfg:hiIl:sSt:" opt; do
    case "${opt}" in
        d)
            COMPOSER_DRY_RUN=true
            ;;
        E)  EXTRACT_CONTAINER_IMAGES=false
            ;;
        f)
            FORCE_REBUILD=true
            ;;
        g)
            GROUP="$(realpath "${OPTARG}")"
            selCount=$((selCount+1))
            ;;
        h)
            usage
            exit 0
            ;;
        i)
            BUILD_INSTALLER=true
            ;;
        I)
            BUILD_INSTALLER=false
            ;;
        l)
            LAYER="$(realpath "${OPTARG}")"
            selCount=$((selCount+1))
            ;;
        s)
            BUILD_INSTALLER=false
            ONLY_SOURCE=true
            FORCE_SOURCE=true
            ;;
        S)
            SKIP_LOG_COLLECTION=true
            ;;
        t)
            TEMPLATE="${OPTARG}"
            GROUP="$(dirname "$(realpath "${OPTARG}")")"
            selCount=$((selCount+1))
            FORCE_REBUILD=true
            ;;
        *)
            usage "ERROR: Unknown option ${opt}"
            exit 1
            ;;
    esac
done

if [ ${selCount} -gt 1 ] ; then
    usage "The layer, group and template options are mutually exclusive"
    exit 1
fi

if [ ! -f "${GOMPLATE}" ]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" gomplate
fi

# Determine the version of the RPM in the local repo so we can use it
# in the blueprint templates.
if [ ! -d "${LOCAL_REPO}" ]; then
    error "Run ${SCRIPTDIR}/create_local_repo.sh before building images."
    exit 1
fi
release_info_rpm=$(find "${LOCAL_REPO}" -name 'microshift-release-info-*.rpm' | sort | tail -n 1)
if [ -z "${release_info_rpm}" ]; then
    error "Failed to find microshift-release-info RPM in ${LOCAL_REPO}"
    exit 1
fi
release_info_rpm_base=$(find "${BASE_REPO}" -name 'microshift-release-info-*.rpm' | sort | tail -n 1)
if [ -z "${release_info_rpm_base}" ]; then
    error "Failed to find microshift-release-info RPM in ${BASE_REPO}"
    exit 1
fi
SOURCE_VERSION=$(rpm -q --queryformat '%{version}' "${release_info_rpm}")
MINOR_VERSION=$(echo "${SOURCE_VERSION}" | cut -f2 -d.)
PREVIOUS_MINOR_VERSION=$(( "${MINOR_VERSION}" - 1 ))
YMINUS2_MINOR_VERSION=$(( "${MINOR_VERSION}" - 2 ))
FAKE_NEXT_MINOR_VERSION=$(( "${MINOR_VERSION}" + 1 ))
SOURCE_VERSION_BASE=$(rpm -q --queryformat '%{version}' "${release_info_rpm_base}")

current_version_repo=$(get_rel_version_repo "${MINOR_VERSION}")
CURRENT_RELEASE_VERSION=$(echo "${current_version_repo}" | cut -d, -f1)
CURRENT_RELEASE_REPO=$(echo "${current_version_repo}" | cut -d, -f2)

previous_version_repo=$(get_rel_version_repo "${PREVIOUS_MINOR_VERSION}")
PREVIOUS_RELEASE_VERSION=$(echo "${previous_version_repo}" | cut -d, -f1)
PREVIOUS_RELEASE_REPO=$(echo "${previous_version_repo}" | cut -d, -f2)

RHOCP_MINOR_Y=""
RHOCP_MINOR_Y1=""
if is_rhocp_available "${MINOR_VERSION}"; then
    RHOCP_MINOR_Y="${MINOR_VERSION}"
fi
if is_rhocp_available "${PREVIOUS_MINOR_VERSION}"; then
    RHOCP_MINOR_Y1="${PREVIOUS_MINOR_VERSION}"
fi

# For Y-2, there will always be a real repository, so we can always
# set the template variable for enabling that package source and use
# the well-known name of that repo instead of figuring out the URL.
yminus2_version_repo=$(get_rel_version_repo "${YMINUS2_MINOR_VERSION}")
YMINUS2_RELEASE_VERSION=$(echo "${yminus2_version_repo}" | cut -d, -f1)
YMINUS2_RELEASE_REPO="$(get_ocp_repo_name_for_version ${YMINUS2_MINOR_VERSION})"
RHOCP_MINOR_Y2="${YMINUS2_MINOR_VERSION}"

mkdir -p "${IMAGEDIR}"
LOGDIR="${IMAGEDIR}/build-logs"
mkdir -p "${LOGDIR}"
mkdir -p "${IMAGEDIR}/blueprints"
mkdir -p "${IMAGEDIR}/builds"
mkdir -p "${VM_DISK_BASEDIR}"

configure_package_sources

# Prepare container lists for mirroring registries.
rm -f "${CONTAINER_LIST}"
if ${EXTRACT_CONTAINER_IMAGES}; then
    extract_container_images "${SOURCE_VERSION}" "${LOCAL_REPO}" "${CONTAINER_LIST}"
    # The following images are specific to layers that use fake rpms built from source.
    extract_container_images "4.${FAKE_NEXT_MINOR_VERSION}.*" "${NEXT_REPO}" "${CONTAINER_LIST}"
    extract_container_images "${PREVIOUS_RELEASE_VERSION}" "${PREVIOUS_RELEASE_REPO}" "${CONTAINER_LIST}"
    extract_container_images "${YMINUS2_RELEASE_VERSION}" "${YMINUS2_RELEASE_REPO}" "${CONTAINER_LIST}"
fi

trap 'osbuild_logs' EXIT

if [ -n "${LAYER}" ]; then
    for group in "${LAYER}"/group*; do
       do_group "${group}" ""
    done
elif [ -n "${GROUP}" ]; then
    do_group "${GROUP}" "${TEMPLATE}"
else
    for group in "${TESTDIR}"/image-blueprints/layer*/group*; do
        do_group "${group}" ""
    done
fi
