#!/bin/bash
set -exo pipefail

sudo dnf install -y git osbuild-composer composer-cli ostree rpm-ostree \
    cockpit-composer bash-completion podman genisoimage \
    createrepo yum-utils selinux-policy-devel jq wget lorax rpm-build
sudo systemctl enable osbuild-composer.socket --now
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd --add-service=cockpit --permanent

# The mock utility comes from the EPEL repository
sudo dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm
sudo dnf install -y mock 
sudo usermod -a -G mock $(whoami)

# Verify umask and home directory permissions
TEST_FILE=/tmp/configure_perm_test.$$

touch ${TEST_FILE}
mkdir ${TEST_FILE}.dir
HOME_PERM=$(stat -c 0%a ~)
FILE_PERM=$(stat -c 0%a ${TEST_FILE})
DIR_PERM=$(stat -c 0%a ${TEST_FILE}.dir)
rm -rf ${TEST_FILE}*

if [ ${HOME_PERM} -lt 0755 ] || [ ${FILE_PERM} -lt 0644 ] || [ ${DIR_PERM} -lt 0755 ] ; then
    echo "Check home directory permissions and umask. The settings must allow read to group and others"
    exit 1
fi
