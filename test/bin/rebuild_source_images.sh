#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

# Rebuild the RPM from source
cd "${ROOTDIR}"
rm -rf _output/rpmbuild
make rpm

cd "${TESTDIR}"

# Update the local repository
"./bin/create_local_repo.sh"

# Given a blueprint filename, extract the name value. It does not have
# to match the filename, but some commands take the file and others
# take the name, so we need to be able to have both.
get_blueprint_name() {
    local filename="${1}"
    tomcli-get "${filename}" name
}

TO_BUILD=""

# shellcheck disable=SC2231  # allow glob expansion without quotes in for loop
for template in ${TESTDIR}/image-blueprints/*.toml; do
    name=$(get_blueprint_name "${template}")
    if [[ "${name}" =~ source ]]; then
        TO_BUILD="${TO_BUILD} ${template}"
    fi
done

# shellcheck disable=SC2086  # pass command arguments quotes to allow word splitting
./bin/build_images.sh ${TO_BUILD}

# Downloading the images again assumes that all of them are still in
# the composer cache. The logic for building the ostree repo does not
# currently cope with replacing content or updating individual images.
./bin/download_images.sh
