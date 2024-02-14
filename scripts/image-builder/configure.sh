#!/bin/bash
set -exo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DNF_RETRY="${SCRIPTDIR}/../dnf_retry.sh"

OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)

# Necessary for embedding container images
if [ ! -e /etc/osbuild-worker/pull-secret.json ] ; then
    sudo mkdir -p /etc/osbuild-worker
    sudo ln -s /etc/crio/openshift-pull-secret /etc/osbuild-worker/pull-secret.json
    sudo tee /etc/osbuild-worker/osbuild-worker.toml &>/dev/null <<EOF
[containers]
auth_file_path = "/etc/osbuild-worker/pull-secret.json"
EOF
fi

# Temporary workaround for broken selinux-policy
# dependencies of osbuild package on RHEL 9.3
OSBUILD_VER=""
source /etc/os-release
if [ "${ID}" == "rhel" ] && [ "${VERSION_ID}" == "9.3" ]; then
    OSBUILD_VER="-93-1.el9"
fi

"${DNF_RETRY}" "install" \
     "osbuild${OSBUILD_VER} osbuild-composer \
     git composer-cli ostree rpm-ostree \
     cockpit-composer bash-completion podman runc genisoimage \
     createrepo yum-utils selinux-policy-devel jq wget lorax rpm-build \
     containernetworking-plugins expect"

sudo systemctl enable osbuild-composer.socket --now
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd --add-service=cockpit --permanent

# The mock utility comes from the EPEL repository
"${DNF_RETRY}" "install" "https://dl.fedoraproject.org/pub/epel/epel-release-latest-${OSVERSION}.noarch.rpm"
"${DNF_RETRY}" "install" "mock nginx tomcli parallel"
sudo usermod -a -G mock "$(whoami)"

# Verify umask and home directory permissions
TEST_FILE=$(mktemp /tmp/configure-perm-test.XXXXX)

touch "${TEST_FILE}.file"
mkdir "${TEST_FILE}.dir"
HOME_PERM=$(stat -c 0%a ~)
FILE_PERM=$(stat -c 0%a "${TEST_FILE}.file")
DIR_PERM=$(stat -c 0%a "${TEST_FILE}.dir")

# Set the Correct Permissions for osbuild-composer
[ "${HOME_PERM}" -lt 0711 ]  && chmod go+x ~

if [ "${FILE_PERM}" -lt 0644 ] || [ "${DIR_PERM}" -lt 0711 ] ; then
    echo "Check ${TEST_FILE}.dir permissions. The umask setting must allow execute to group/others"
    echo "Check ${TEST_FILE}.file permissions. The umask setting must allow read to group/others"
    exit 1
fi

rm -rf "${TEST_FILE}"*
