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

echo "Removing the existing 'osbuild' packages..."
LIST2REMOVE=$(rpm -qa | egrep 'osbuild|cockpit-composer|rpm-ostree' || true)
[ ! -z "${LIST2REMOVE}" ] && sudo rpm -e ${LIST2REMOVE}

echo "Configuring the 'copr' repositories..."
sudo dnf copr -y $REPO_MODE @osbuild/osbuild
sudo dnf copr -y $REPO_MODE @osbuild/osbuild-composer

# sudo dnf copr -y $REPO_MODE walters/ostreerhel8
sudo rm -f /etc/yum.repos.d/walters-ostreerhel8-centos-stream-8.repo
if [ "${REPO_MODE}" = enable ] ; then
    sudo wget -P /etc/yum.repos.d https://copr.fedorainfracloud.org/coprs/walters/ostreerhel8/repo/centos-stream-8/walters-ostreerhel8-centos-stream-8.repo
fi

echo "Installing new 'osbuild' packages..."
$ROOTDIR/../scripts/image-builder/configure.sh
$ROOTDIR/../scripts/image-builder/cleanup.sh -full

echo "Querying the installed 'osbuild' packages..."
rpm -qa | egrep '^osbuild|ostree'
