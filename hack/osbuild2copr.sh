#!/bin/bash
set -e

ROOTDIR=$(dirname $0)

REPO_MODE=
case "$1" in
copr)
    REPO_MODE=enable
    ;;
appstream)
    REPO_MODE=disable
    ;;
*)
    echo "Usage: $(basename $0) <copr | appstream>"
    exit 1
    ;;
esac

echo "Removing the existing 'rpm-ostree' packages..."
LIST2REMOVE=$(rpm -qa | egrep '^rpm-ostree' || true)
[ ! -z "${LIST2REMOVE}" ] && sudo dnf remove -y ${LIST2REMOVE}

# Clean-up the old osbuild jobs and state to avoid incompatibilities between versions
sudo rm -rf /var/lib/osbuild-composer || true
sudo rm -rf /var/cache/{osbuild-composer,osbuild-worker} || true

# sudo dnf copr -y $REPO_MODE walters/ostreerhel8
sudo rm -f /etc/yum.repos.d/walters-ostreerhel8-centos-stream-8.repo
if [ "${REPO_MODE}" = enable ] ; then
    sudo wget -P /etc/yum.repos.d https://copr.fedorainfracloud.org/coprs/walters/ostreerhel8/repo/centos-stream-8/walters-ostreerhel8-centos-stream-8.repo
fi

echo "Installing new 'rpm-ostree' packages..."
$ROOTDIR/../scripts/image-builder/configure.sh
$ROOTDIR/../scripts/image-builder/cleanup.sh -full

echo "Querying the installed 'rpm-ostree' packages..."
rpm -qa | egrep 'rpm-ostree'
