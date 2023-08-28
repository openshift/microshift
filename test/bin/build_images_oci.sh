#!/usr/bin/env bash

set -euo pipefail

PURPLE='\033[0;35m'
GREEN='\033[0;32m'
NC='\033[0m'

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

time_it() {
	local cmd=$@
	echo -e "${GREEN}\$ ${cmd}${NC}"
	start=$(date +%s)
	${cmd}
	end=$(date +%s)
	elapsed=$((end - start))
	echo -e "${GREEN}${elapsed} seconds\t\t${cmd}${NC}\n\n"
}

time_it ./bin/build_images.sh -g "${TESTDIR}/image-blueprints/group1"

BUILD_ARCH=$(uname -m)
REPO="${IMAGEDIR}/repo"

if [ ! -d /mnt/tmp ]; then
    # create tmpfs because xfs doesn't handle xattrs well when importing an image
    sudo mkdir /mnt/tmp
    sudo mount -t tmpfs -o size=15g tmpfs /mnt/tmp
fi
TMP_REPO="/mnt/tmp/repo"

RHEL_DEPS_REF="rhel-9.2"
RHEL_DEPS_OCI="localhost/rhel-9.2-deps"

REPOS=("${LOCAL_REPO}" "${BASE_REPO}" "${NEXT_REPO}" "${YPLUS2_REPO}")
IMAGES=(
    rhel-9.2-microshift-source
    rhel-9.2-microshift-source-base 
    rhel-9.2-microshift-source-fake-next-minor
    rhel-9.2-microshift-source-fake-yplus2-minor
)

rm -rf "${LOCAL_REPO}/RPMS/noarch/microshift-test-agent-0.0.1-1.noarch.rpm" # should be part of base deps rhel image already

time_it sudo ostree container encapsulate --repo "${REPO}" "${RHEL_DEPS_REF}" "containers-storage:${RHEL_DEPS_OCI}"

START=$(date +%s)

sudo rm -rf "${TMP_REPO}"
sudo ostree init --repo "${TMP_REPO}" --mode bare

for ((i=0; i<${#REPOS[@]}; i++)); do
    new_oci="${IMAGES[${i}]}"
    full_new_oci="localhost/${new_oci}"
    repo="${REPOS[${i}]}"
    echo "${new_oci}:  ${repo}"

    time_it sudo podman rmi --ignore "${full_new_oci}"

    time_it sudo podman build --squash --tag "${full_new_oci}" -f- "${repo}" <<EOF
    FROM ${RHEL_DEPS_OCI}
    COPY RPMS /tmp/rpms
    RUN rpm-ostree install /tmp/rpms/noarch/* /tmp/rpms/${BUILD_ARCH}/* && \
        systemctl enable microshift crio microshift-test-agent && \
        rm -rf /tmp/* && \
        ostree container commit
EOF

    time_it sudo ostree container image pull "${TMP_REPO}" "ostree-unverified-image:containers-storage:${full_new_oci}"

    IMPORTED_OCI_REF="ostree/container/image/containers-storage_3A_${full_new_oci/./_2E_}"
    CLEAN_OCI_REF="${new_oci/:/\/}"

    time_it sudo ostree refs --repo "${TMP_REPO}" --alias --create "${CLEAN_OCI_REF}" "${IMPORTED_OCI_REF}"
    time_it sudo ostree pull-local --repo "${REPO}" "${TMP_REPO}" "${CLEAN_OCI_REF}"
done

ostree refs --repo "${REPO}" --alias --create rhel-9.2-microshift-source-aux rhel-9.2-microshift-source
ostree refs --repo "${REPO}"
ostree summary --update --repo "${REPO}" 
ostree summary --view --repo "${REPO}"

END=$(date +%s)
ELAPSED=$((END - START))
echo -e "${PURPLE}Total: ${ELAPSED} seconds${NC}"
