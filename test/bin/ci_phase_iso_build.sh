#!/bin/bash
#
# This script runs on the build host to create all test artifacts.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

# Cannot use common.sh yet because some dependencies may be missing,
# but we only need ROOTDIR at this time.
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="$(cd "${SCRIPTDIR}/../.." && pwd)"

# Log output automatically
LOGDIR="${ROOTDIR}/_output/ci-logs"
LOGFILE="${LOGDIR}/$(basename "$0" .sh).log"
if [ ! -d "${LOGDIR}" ]; then
    mkdir -p "${LOGDIR}"
fi
echo "Logging to ${LOGFILE}"
# Set fd 1 and 2 to write to the log file
exec &> >(tee >(awk '{ print strftime("%Y-%m-%d %H:%M:%S"), $0; fflush() }' >"${LOGFILE}"))

PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}

# Detect a bootc build mode based on the job name
COMPOSER_CLI_BUILDS=true
if [ -v CI_JOB_NAME ] && [[ "${CI_JOB_NAME}" =~ .*bootc.* ]]; then
    COMPOSER_CLI_BUILDS=false
fi

# Allow for a dry-run option to save on testing time
BUILD_DRY_RUN=${BUILD_DRY_RUN:-false}
dry_run() {
    ${BUILD_DRY_RUN} && echo "echo"
}

# Try downloading the 'last' build cache.
# Return 0 on success or 1 otherwise.
download_build_cache() {
    local -r cache_last="$(\
        ./bin/manage_build_cache.sh getlast \
            -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}" | \
            awk '/LAST:/ {print $NF}' \
        )"

    if ./bin/manage_build_cache.sh verify -b "${SCENARIO_BUILD_BRANCH}" -t "${cache_last}" ; then
        # Download the cached images
        ./bin/manage_build_cache.sh download -b "${SCENARIO_BUILD_BRANCH}" -t "${cache_last}"
        return 0
    fi
    return 1
}

# Run image build for the 'base' layers and update the cache:
# - Upload build artifacts
# - Update 'last' to point to the current build tag
# - Clean up older images, preserving the 'last' and the previous build tag
# Note that the build and upload are skipped if valid cached data already exists.
update_build_cache() {
    if ./bin/manage_build_cache.sh verify -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}" ; then
        echo "Valid build cache already exists for the '${SCENARIO_BUILD_BRANCH}' branch and '${SCENARIO_BUILD_TAG}' tag"
        echo "WARNING: Skipping cache build, update and cleanup procedures"
        return
    fi

    # Build the composer-cli base layer to be cached
    $(dry_run) bash -x ./bin/build_images.sh -l ./image-blueprints/layer1-base
    # Build the bootc base layer to be cached
    $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer1-base

    # Prepare for the cache upload by stopping composer services and cleaning
    # temporary artifacts
    ./bin/manage_composer_config.sh cleanup

    # Upload the images and update the 'last' setting
    ./bin/manage_build_cache.sh upload  -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}"
    ./bin/manage_build_cache.sh setlast -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}"

    # Cleanup older images in the cache, preserving the previous cache if any
    # The 'last' cache is preserved by default
    ./bin/manage_build_cache.sh keep -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG_PREV}"
}

# Run image build, potentially skipping the 'periodic' layer in CI builds.
# Full builds are run if the 'CI_JOB_NAME' environment variable is not set.
#
# When the 'CI_JOB_NAME' environment variable is set:
# - Always build the 'base' and 'presubmit' layers.
# - Only build the 'periodic' layer when 'CI_JOB_NAME' contains 'periodic' token.
run_image_build() {
    if [ -v CI_JOB_NAME ] ; then
        # Conditional per-layer builds when running in CI.
        # The build_images.sh script skips any images that have been downloaded from the cache.
        $(dry_run) bash -x ./bin/build_images.sh -l ./image-blueprints/layer1-base
        $(dry_run) bash -x ./bin/build_images.sh -l ./image-blueprints/layer2-presubmit

        if [[ "${CI_JOB_NAME}" =~ .*periodic.* ]]; then
            $(dry_run) bash -x ./bin/build_images.sh -l ./image-blueprints/layer3-periodic
        fi
    else
        # Fall back to full build when not running in CI
        $(dry_run) bash -x ./bin/build_images.sh
    fi
}

# Run container file verification and bootc image build
run_bootc_image_build() {
    make -C "${ROOTDIR}" verify-containers

    if [ -v CI_JOB_NAME ] ; then
        $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer1-base
        $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer2-presubmit
    
        if [[ "${CI_JOB_NAME}" =~ .*periodic.* ]]; then
            $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer3-periodic
        fi
    else
        $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer1-base
        $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer2-presubmit
        $(dry_run) bash -x ./bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer3-periodic
    fi
}

cat /etc/os-release

# Show what other dnf commands have been run to try to debug why we
# sometimes see cache collisons.
$(dry_run) sudo dnf history --reverse

cd "${ROOTDIR}"

# Get firewalld and repos in place. Use scripts to get the right repos
# for each branch.
$(dry_run) bash -x ./scripts/devenv-builder/configure-vm.sh --no-build --force-firewall "${PULL_SECRET}"
$(dry_run) bash -x ./test/bin/manage_composer_config.sh create

cd "${ROOTDIR}/test/"

# Source common.sh only after all dependencies are installed.
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

#

if ${COMPOSER_CLI_BUILDS} ; then
    # Determine and create the ideal number of workers
    $(dry_run) bash -x ./bin/manage_composer_config.sh create-workers
fi

# Check if cache can be used for builds
# This may fail when AWS S3 connection is not configured, or there is no cache bucket
HAS_CACHE_ACCESS=false
if ./bin/manage_build_cache.sh getlast -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}" ; then
    HAS_CACHE_ACCESS=true
fi

# Check the build mode: "try using cache" (default) or "update cache"
if [ $# -gt 0 ] && [ "$1" = "-update_cache" ] ; then
    if ${HAS_CACHE_ACCESS} ; then
        # Re-build from source before updating the cache because some
        # build artifacts may be cached
        $(dry_run) bash -x ./bin/build_rpms.sh

        update_build_cache
    else
        echo "ERROR: Access to the build cache is not available"
        exit 1
    fi
else
    GOT_CACHED_DATA=false
    if ${HAS_CACHE_ACCESS} ; then
        if download_build_cache ; then
            GOT_CACHED_DATA=true
        fi
    fi
    if ! ${GOT_CACHED_DATA} ; then
        echo "WARNING: Build cache is not available, rebuilding all the artifacts"
    fi

    # Re-build from source after downloading the cache because
    # the build may depend on some cached artifacts
    $(dry_run) bash -x ./bin/build_rpms.sh

    # Optionally run bootc image builds
    if ${COMPOSER_CLI_BUILDS} ; then
        run_image_build
    else
        run_bootc_image_build
    fi
fi

echo "Build phase complete"
exit 0
