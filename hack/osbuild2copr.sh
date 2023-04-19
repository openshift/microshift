#!/bin/bash
set -e

ROOTDIR=$(dirname "$0")

COPR_MODE=
case "$1" in
copr)
    COPR_MODE=enable
    ;;
appstream)
    COPR_MODE=disable
    ;;
*)
    echo "Usage: $(basename "$0") <copr | appstream>"
    exit 1
    ;;
esac

echo "Removing the existing 'osbuild' packages..."
LIST2REMOVE=$(rpm -qa | grep -E '^osbuild' || true)
# shellcheck disable=SC2086
# The list need not be quoted to allow multiple package removal
[ -n "${LIST2REMOVE}" ] && sudo dnf remove -y ${LIST2REMOVE}

# Clean-up the old osbuild jobs, state and copr packages to avoid incompatibilities between versions
sudo rm -rf /var/lib/osbuild-composer || true
sudo rm -rf /var/cache/{osbuild-composer,osbuild-worker} || true
sudo rm -f  /etc/yum.repos.d/_copr:copr.fedorainfracloud.org:*osbuild* || true

sudo dnf copr -y "${COPR_MODE}" @osbuild/osbuild          "epel-9-$(uname -i)"
sudo dnf copr -y "${COPR_MODE}" @osbuild/osbuild-composer "epel-9-$(uname -i)"

# Uncomment the following to use the packages from PRs before their merge
# sudo dnf copr -y "${COPR_MODE}" packit/osbuild-osbuild-1252          "epel-9-$(uname -i)"
# sudo dnf copr -y "${COPR_MODE}" packit/osbuild-osbuild-composer-3398 "epel-9-$(uname -i)"

echo "Installing new 'osbuild' packages..."
"${ROOTDIR}/../scripts/image-builder/configure.sh"
"${ROOTDIR}/../scripts/image-builder/cleanup.sh" -full

echo "Querying the installed 'osbuild' packages..."
rpm -qa | grep -E '^osbuild'
