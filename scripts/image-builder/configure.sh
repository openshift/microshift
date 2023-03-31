#!/bin/bash
set -exo pipefail

OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)

function osbuild_rhel9_beta() {
    local json_file=$1
    sudo mkdir -p $(dirname ${json_file})
    sudo tee ${json_file} >/dev/null <<EOF
{
  "x86_64": [
    {
      "name": "baseos",
      "baseurl": "https://cdn.redhat.com/content/beta/rhel9/9/x86_64/baseos/os",
      "rhsm": true,
      "check_gpg": false
    },
    {
      "name": "appstream",
      "baseurl": "https://cdn.redhat.com/content/beta/rhel9/9/x86_64/appstream/os",
      "rhsm": true,
      "check_gpg": false
    }
  ],
  "aarch64": [
    {
      "name": "baseos",
      "baseurl": "https://cdn.redhat.com/content/beta/rhel9/9/aarch64/baseos/os",
      "rhsm": true,
      "check_gpg": false
    },
    {
      "name": "appstream",
      "baseurl": "https://cdn.redhat.com/content/beta/rhel9/9/aarch64/appstream/os",
      "rhsm": true,
      "check_gpg": false
    }
  ]
}
EOF

    sudo systemctl stop --now osbuild-composer.socket osbuild-composer.service osbuild-worker@1.service
    sleep 5
    sudo rm -rf /var/cache/osbuild-worker/* /var/lib/osbuild-composer/*
}

sudo dnf install -y git osbuild-composer composer-cli ostree rpm-ostree \
    cockpit-composer cockpit-machines bash-completion podman genisoimage \
    createrepo yum-utils selinux-policy-devel jq wget lorax rpm-build \
    containernetworking-plugins

# Configure osbuild-composer to use RHEL 9.2 beta repositories
# This is a workaround until RHEL 9.2 becomes GA
if grep -q 'Red Hat Enterprise Linux release 9.2 Beta' /etc/redhat-release ; then
    JSON_FILE=/etc/osbuild-composer/repositories/rhel-92.json
    [ ! -e ${JSON_FILE} ] && osbuild_rhel9_beta ${JSON_FILE}
fi

sudo systemctl enable osbuild-composer.socket --now
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd --add-service=cockpit --permanent

# The mock utility comes from the EPEL repository
sudo dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-${OSVERSION}.noarch.rpm
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
