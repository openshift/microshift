#!/bin/bash
set -e -o pipefail

IMGNAME=microshift
ROOTDIR=$(git rev-parse --show-toplevel)/scripts/image-builder
STARTTIME=$(date +%s)

trap ${ROOTDIR}/cleanup.sh INT
trap 'echo "Execution time: $(( ($(date +%s) - STARTTIME) / 60 )) minutes"' EXIT

title() {
    echo -e "\E[34m\n# $1\E[00m";
}

waitfor_image() {
    local uuid=$1

    local tstart=$(date +%s)
    echo "$(date +'%Y-%m-%d %H:%M:%S') STARTED"

    local status=$(sudo composer-cli compose status | grep ${uuid} | awk '{print $2}')
    while [ "${status}" = "RUNNING" ]; do
        sleep 10
        status=$(sudo composer-cli compose status | grep ${uuid} | awk '{print $2}')
        echo -en "$(date +'%Y-%m-%d %H:%M:%S') ${status}\r"
    done

    local tend=$(date +%s)
    echo "$(date +'%Y-%m-%d %H:%M:%S') ${status} - elapsed $(( (tend - tstart) / 60 )) minutes"
    
    if [ "${status}" = "FAILED" ]; then
        download_image ${uuid} 1
        echo "Blueprint build has failed. For more information, review the downloaded logs"
        exit 1
    fi
}

download_image() {
    local uuid=$1
    local logsonly=$2

    sudo composer-cli compose logs ${uuid}
    if [ -z "$logsonly" ] ; then
        sudo composer-cli compose metadata ${uuid}
        sudo composer-cli compose image ${uuid}
    fi
    sudo chown -R $(whoami). "${ROOTDIR}/_builds"
}

build_image() {
    local blueprint_file=$1
    local blueprint=$2
    local version=$3
    local image_type=$4
    local parent_blueprint=$5
    local parent_version=$6

    title "Loading ${blueprint} blueprint v${version}"
    sudo composer-cli blueprints delete ${blueprint} 2>/dev/null || true
    sudo composer-cli blueprints push "${ROOTDIR}/${blueprint_file}"
    sudo composer-cli blueprints depsolve ${blueprint} 1>/dev/null

    if [ -n "$parent_version" ]; then
        title "Serving ${parent_blueprint} v${parent_version} container locally"
        sudo podman rm -f ${parent_blueprint}-server 2>/dev/null || true
        sudo podman rmi -f localhost/${parent_blueprint}:${parent_version} 2>/dev/null || true
        imageid=$(cat ./${parent_blueprint}-${parent_version}-container.tar | sudo podman load | grep -o -P '(?<=sha256[@:])[a-z0-9]*')
        sudo podman tag ${imageid} localhost/${parent_blueprint}:${parent_version}
        sudo podman run -d --name=${parent_blueprint}-server -p 8080:8080 localhost/${parent_blueprint}:${parent_version}

        title "Building ${image_type} for ${blueprint} v${version}, parent ${parent_blueprint} v${parent_version}"
        buildid=$(sudo composer-cli compose start-ostree --ref rhel/8/$(uname -i)/edge --url http://localhost:8080/repo/ ${blueprint} ${image_type} | awk '{print $2}')
    else
        title "Building ${image_type} for ${blueprint} v${version}"
        buildid=$(sudo composer-cli compose start-ostree --ref rhel/8/$(uname -i)/edge ${blueprint} ${image_type} | awk '{print $2}')
    fi

    waitfor_image ${buildid}
    download_image ${buildid}
    rename ${buildid} ${blueprint}-${version} ${buildid}*.{tar,iso} 2>/dev/null || true
}

mkdir -p ${ROOTDIR}/_builds
pushd ${ROOTDIR}/_builds &>/dev/null

# Also enter sudo password in the beginning if necessary
title "Checking available disk space"
build_disk=$(sudo df -k --output=avail . | tail -1)
if [ ${build_disk} -lt 10485760 ] ; then
    echo "ERROR: Less then 10GB of disk space is available for the build"
    exit 1
fi

title "Downloading local OpenShift and MicroShift repositories"
# Copy microshift local RPM packages
rm -rf microshift-local 2>/dev/null || true
cp -TR "${ROOTDIR}/../../packaging/rpm/_rpmbuild/RPMS" microshift-local
createrepo microshift-local >/dev/null

# Download openshift local RPM packages
rm -rf openshift-local 2>/dev/null || true
reposync -n -a x86_64 --download-path openshift-local --repo=rhocp-4.10-for-rhel-8-x86_64-rpms >/dev/null
# Remove coreos packages to avoid conflicts
find openshift-local -name \*coreos\* -exec rm -f {} \;
createrepo openshift-local >/dev/null

title "Loading sources for OpenShift and MicroShift"
for f in openshift-local microshift-local ; do
    cat ../${f}.toml | sed "s;REPLACE_IMAGE_BUILDER_DIR;${ROOTDIR};g" > ${f}.toml
    sudo composer-cli sources delete $f 2>/dev/null || true
    sudo composer-cli sources add ${ROOTDIR}/_builds/${f}.toml
done

build_image blueprint_v0.0.1.toml "${IMGNAME}-container" 0.0.1 edge-container
build_image installer.toml        "${IMGNAME}-installer" 0.0.0 edge-installer "${IMGNAME}-container" 0.0.1

title "Embedding kickstart in the installer image"
# TODO: Replace with the real server
cat "../kickstart.ks" | sed "s;REPLACE_OSTREE_SERVER_IP;localhost:8080;g" > kickstart.ks
sudo podman run --rm --privileged -ti -v "${ROOTDIR}/_builds":/data -v /dev:/dev fedora /bin/bash -c \
    "dnf -y install lorax; cd /data; \
    mkksiso kickstart.ks ${IMGNAME}-installer-0.0.0-installer.iso ${IMGNAME}-installer.$(uname -i).iso; \
    exit"
sudo chown -R $(whoami). "${ROOTDIR}/_builds"

# Remove intermediate artifacts to free disk space
rm -f ${IMGNAME}-installer-0.0.0-installer.iso

${ROOTDIR}/cleanup.sh

title "Done"
popd &>/dev/null
