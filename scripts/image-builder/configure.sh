#!/bin/bash
set -exo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DNF_RETRY="${SCRIPTDIR}/../dnf_retry.sh"

# Necessary for embedding container images
if [ ! -e /etc/osbuild-worker/pull-secret.json ] ; then
    sudo mkdir -p /etc/osbuild-worker
    sudo ln -s /etc/crio/openshift-pull-secret /etc/osbuild-worker/pull-secret.json
    sudo tee /etc/osbuild-worker/osbuild-worker.toml &>/dev/null <<EOF
[containers]
auth_file_path = "/etc/osbuild-worker/pull-secret.json"
EOF
fi

"${DNF_RETRY}" "install" \
     "osbuild osbuild-composer \
     git composer-cli ostree rpm-ostree \
     cockpit-composer bash-completion podman runc genisoimage \
     createrepo yum-utils selinux-policy-devel jq wget lorax rpm-build \
     containernetworking-plugins expect"

# Parse the OS versions and determine if EUS
source /etc/os-release
VERSION_ID_MAJOR="$(awk -F. '{print $1}' <<< "${VERSION_ID}")"
VERSION_ID_MINOR="$(awk -F. '{print $2}' <<< "${VERSION_ID}")"
VERSION_ID_EUS="dist"
if (( "${VERSION_ID_MINOR}" % 2 == 0 )) ; then
    VERSION_ID_EUS="eus"
fi

# Edit composer configuration file for the current operating system
COMPOSER_CONFIG="/etc/osbuild-composer/repositories/rhel-${VERSION_ID}.json"

# Enable RT repository by duplicating the 'baseos' repository, changing its name,
# and replacing 'baseos' with 'rt'.
# Note that kernel-rt is only available for x86_64.
"${SCRIPTDIR}/../fetch_tools.sh" yq
sudo mkdir -p /etc/osbuild-composer/repositories/
"${SCRIPTDIR}/../../_output/bin/yq" \
    '.["x86_64"] += (.["x86_64"][0] | .name = "kernel-rt" | .baseurl |= sub("baseos", "rt"))' \
    "/usr/share/osbuild-composer/repositories/rhel-${VERSION_ID}.json" | jq | sudo tee "${COMPOSER_CONFIG}" >/dev/null

# Enable beta or EUS repositories.
# The configuration will remain unchanged for non-beta and non-EUS operating systems.
if grep -qE "Red Hat Enterprise Linux.*Beta" /etc/redhat-release; then
    sudo sed -i "s,dist/rhel${VERSION_ID_MAJOR}/${VERSION_ID},beta/rhel${VERSION_ID_MAJOR}/${VERSION_ID_MAJOR},g" "${COMPOSER_CONFIG}"
else
    sudo sed -i "s,dist/rhel${VERSION_ID_MAJOR}/${VERSION_ID},${VERSION_ID_EUS}/rhel${VERSION_ID_MAJOR}/${VERSION_ID},g" "${COMPOSER_CONFIG}"
fi

composer_active=$(sudo systemctl is-active osbuild-composer.service || true)
sudo systemctl enable osbuild-composer.socket --now
if [[ "${composer_active}" == "active" ]]; then
    # If composer was active before, restart it to make kernel-rt repository configuration active.
    sudo systemctl restart osbuild-composer.service
fi
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd --add-service=cockpit --permanent

# The mock utility comes from the EPEL repository
"${DNF_RETRY}" "install" "https://dl.fedoraproject.org/pub/epel/epel-release-latest-${VERSION_ID_MAJOR}.noarch.rpm"
"${DNF_RETRY}" "install" "mock nginx tomcli parallel aria2"
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
