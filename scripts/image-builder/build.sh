#!/bin/bash
set -e -o pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
SCRIPTDIR=${ROOTDIR}/scripts/image-builder
IMGNAME=microshift
IMAGE_VERSION=$(${ROOTDIR}/pkg/release/get.sh base)
BUILD_ARCH=$(uname -i)
OSTREE_SERVER_NAME=127.0.0.1:8080
LVM_SYSROOT_SIZE_MIN=8192
LVM_SYSROOT_SIZE=${LVM_SYSROOT_SIZE_MIN}
OCP_PULL_SECRET_FILE=
MICROSHIFT_RPM_SOURCE=${ROOTDIR}/_output/rpmbuild/
AUTHORIZED_KEYS_FILE=
AUTHORIZED_KEYS=
PROMETHEUS=false
EMBED_CONTAINERS=false
STARTTIME=$(date +%s)
BUILDDIR=${ROOTDIR}/_output/image-builder

trap ${SCRIPTDIR}/cleanup.sh INT

usage() {
    local error_message="$1"

    if [ -n "$error_message" ]; then
        echo "ERROR: $error_message"
        echo
    fi

    echo "Usage: $(basename $0) <-pull_secret_file path_to_file> [OPTION]..."
    echo ""
    echo "  -pull_secret_file path_to_file"
    echo "          Path to a file containing the OpenShift pull secret, which can be"
    echo "          obtained from https://console.redhat.com/openshift/downloads#tool-pull-secret"
    echo ""
    echo "Optional arguments:"
    echo "  -microshift_rpms path_or_URL"
    echo "          Path or URL to the MicroShift RPM packages to be included"
    echo "          in the image (default: _output/rpmbuild/RPMS)"
    echo "  -custom_rpms /path/to/file1.rpm,...,/path/to/fileN.rpm"
    echo "          Path to one or more comma-separated RPM packages to be"
    echo "          included in the image (default: none)"
    echo "  -embed_containers"
    echo "          Embed the MicroShift container dependencies in the image using the"
    echo "          'pkg/release/get.sh images \$(uname -i)' command to get their list"
    echo "  -ostree_server_name name_or_ip"
    echo "          Name or IP address and optionally port of the ostree"
    echo "          server (default: ${OSTREE_SERVER_NAME})"
    echo "  -lvm_sysroot_size num_in_MB"
    echo "          Size of the system root LVM partition. The remaining"
    echo "          disk space will be allocated for data (default: ${LVM_SYSROOT_SIZE})"
    echo "  -authorized_keys_file path_to_file"
    echo "          Path to an SSH authorized_keys file to allow SSH access"
    echo "          into the default 'redhat' account"
    echo "  -prometheus"
    echo "          Add Prometheus process exporter to the image. See"
    echo "          https://github.com/ncabatoff/process-exporter for more information"
    exit 1
}

title() {
    echo -e "\E[34m\n# $1\E[00m";
}

waitfor_image() {
    local uuid=$1

    local tstart=$(date +%s)
    echo "$(date +'%Y-%m-%d %H:%M:%S') STARTED"

    local status=$(sudo composer-cli compose status | grep ${uuid} | awk '{print $2}')
    while [ "${status}" = "RUNNING" -o "${status}" = "WAITING" ]; do
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
    sudo chown -R $(whoami). "${BUILDDIR}"
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
    sudo composer-cli blueprints push "${BUILDDIR}/${blueprint_file}"
    sudo composer-cli blueprints depsolve ${blueprint} 1>/dev/null

    if [ -n "$parent_version" ]; then
        title "Serving ${parent_blueprint} v${parent_version} container locally"
        sudo podman rm -f ${parent_blueprint}-server 2>/dev/null || true
        sudo podman rmi -f localhost/${parent_blueprint}:${parent_version} 2>/dev/null || true
        imageid=$(cat ./${parent_blueprint}-${parent_version}-container.tar | sudo podman load | grep -o -P '(?<=sha256[@:])[a-z0-9]*')
        sudo podman tag ${imageid} localhost/${parent_blueprint}:${parent_version}
        sudo podman run -d --name=${parent_blueprint}-server -p 8080:8080 localhost/${parent_blueprint}:${parent_version}

        title "Building ${image_type} for ${blueprint} v${version}, parent ${parent_blueprint} v${parent_version}"
        buildid=$(sudo composer-cli compose start-ostree --ref rhel/8/${BUILD_ARCH}/edge --url http://localhost:8080/repo/ ${blueprint} ${image_type} | awk '{print $2}')
    else
        title "Building ${image_type} for ${blueprint} v${version}"
        buildid=$(sudo composer-cli compose start-ostree --ref rhel/8/${BUILD_ARCH}/edge ${blueprint} ${image_type} | awk '{print $2}')
    fi

    waitfor_image ${buildid}
    download_image ${buildid}
    rename ${buildid} ${blueprint}-${version} ${buildid}*.{tar,iso} 2>/dev/null || true
}

# Parse the command line
while [ $# -gt 0 ] ; do
    case $1 in
    -pull_secret_file)
        shift
        OCP_PULL_SECRET_FILE="$1"
        [ -z "${OCP_PULL_SECRET_FILE}" ] && usage "Pull secret file not specified"
        [ ! -s "${OCP_PULL_SECRET_FILE}" ] && usage "Empty or missing pull secret file"
        shift
        ;;
    -microshift_rpms)
        shift
        MICROSHIFT_RPM_SOURCE="$1"
        [ -z "${MICROSHIFT_RPM_SOURCE}" ] && usage "MicroShift RPM path or URL not specified"
        # Verify that the specified path or URL can be accessed
        if [[ "${MICROSHIFT_RPM_SOURCE}" == http* ]] ; then
            curl -I -so /dev/null "${MICROSHIFT_RPM_SOURCE}" || usage "MicroShift RPM URL '${MICROSHIFT_RPM_SOURCE}' is not accessible"
        else
            [ ! -d "${MICROSHIFT_RPM_SOURCE}" ] && usage "MicroShift RPM path '${MICROSHIFT_RPM_SOURCE}' does not exist"
        fi
        shift
        ;;
    -custom_rpms)
        shift
        CUSTOM_RPM_FILES="$1"
        [ -z "${CUSTOM_RPM_FILES}" ] && usage "Custom RPM packages not specified"
        shift
        ;;

    -embed_containers)
        EMBED_CONTAINERS=true
        shift
        ;;
    -ostree_server_name)
        shift
        OSTREE_SERVER_NAME="$1"
        [ -z "${OSTREE_SERVER_NAME}" ] && usage "ostree server name not specified"
        shift
        ;;
    -lvm_sysroot_size)
        shift
        LVM_SYSROOT_SIZE="$1"
        [ -z "${LVM_SYSROOT_SIZE}" ] && usage "System root LVM partition size not specified"
        [ ${LVM_SYSROOT_SIZE} -lt ${LVM_SYSROOT_SIZE_MIN} ] && usage "System root LVM partition size cannot be smaller than ${LVM_SYSROOT_SIZE_MIN}MB"
        shift
        ;;
    -authorized_keys_file)
        shift
        AUTHORIZED_KEYS_FILE="$1"
        [ -z "${AUTHORIZED_KEYS_FILE}" ] && usage "Authorized keys file not specified"
        shift
        ;;
    -prometheus)
        PROMETHEUS=true
        shift
        ;;
    *)
        usage
        ;;
    esac
done

if [ -z "${OSTREE_SERVER_NAME}" ] || [ -z "${OCP_PULL_SECRET_FILE}" ] ; then
    usage
fi
if [ ! -e ${OCP_PULL_SECRET_FILE} ] ; then
    echo "ERROR: pull_secret_file file does not exist: ${OCP_PULL_SECRET_FILE}"
    exit 1
fi
if [ -n "${AUTHORIZED_KEYS_FILE}" ]; then
    if [ ! -e ${AUTHORIZED_KEYS_FILE} ]; then
        echo "ERROR: authorized_keys_file does not exist: ${AUTHORIZED_KEYS_FILE}"
        exit 1
    else
        AUTHORIZED_KEYS=$(cat ${AUTHORIZED_KEYS_FILE})
    fi
fi

# Set the elapsed time trap only if command line parsing was successful
trap 'echo "Execution time: $(( ($(date +%s) - STARTTIME) / 60 )) minutes"' EXIT

mkdir -p ${BUILDDIR}
pushd ${BUILDDIR} &>/dev/null

# Also enter sudo password in the beginning if necessary
title "Checking available disk space"
build_disk=$(sudo df -k --output=avail . | tail -1)
if [ ${build_disk} -lt 10485760 ] ; then
    echo "ERROR: Less then 10GB of disk space is available for the build"
    exit 1
fi

title "Downloading local OpenShift and MicroShift repositories"
# Copy MicroShift RPM packages
rm -rf microshift-local 2>/dev/null || true
if [[ "${MICROSHIFT_RPM_SOURCE}" == http* ]] ; then
    wget -q -nd -r -L -P microshift-local -A rpm "${MICROSHIFT_RPM_SOURCE}"
else
    [ ! -d "${MICROSHIFT_RPM_SOURCE}" ] && echo "MicroShift RPM path '${MICROSHIFT_RPM_SOURCE}' does not exist" && exit 1
    cp -TR "${MICROSHIFT_RPM_SOURCE}" microshift-local
fi
# Exit if no RPM packages were found
if [ $(find microshift-local -name '*.rpm' | wc -l) -eq 0 ] ; then
    echo "No RPM packages were found at '${MICROSHIFT_RPM_SOURCE}'. Exiting..."
    exit 1
fi
createrepo microshift-local >/dev/null

# Download openshift local RPM packages (noarch for python and selinux packages)
rm -rf openshift-local 2>/dev/null || true
reposync -n -a ${BUILD_ARCH} -a noarch --download-path openshift-local \
    --repo=rhocp-4.12-el8-beta-${BUILD_ARCH}-rpms \
    --repo=fast-datapath-for-rhel-8-${BUILD_ARCH}-rpms >/dev/null
#   --repo=rhocp-4.12-for-rhel-8-${BUILD_ARCH}-rpms \

# Remove coreos packages to avoid conflicts
find openshift-local -name \*coreos\* -exec rm -f {} \;
# Exit if no RPM packages were found
if [ $(find openshift-local -name '*.rpm' | wc -l) -eq 0 ] ; then
    echo "No RPM packages were found at the 'rhocp-4.12-for-rhel-8-${BUILD_ARCH}-rpms' repository. Exiting..."
    exit 1
fi
createrepo openshift-local >/dev/null

# Install prometheus process exporter
if ${PROMETHEUS} ; then
  owner=ncabatoff
  repo=process-exporter
  version=$(curl -s https://api.github.com/repos/${owner}/${repo}/releases/latest | jq -r '.name')
  url=https://github.com/${owner}/${repo}/releases/download/${version}/
  file=${repo}_${version#v}_linux_amd64.rpm

  title "Downloading Prometheus exporter(s)"
  wget -q ${url}${file}
  CUSTOM_RPM_FILES+="$(pwd)/${file},"
fi

# Copy user-specific RPM packages
rm -rf custom-rpms 2>/dev/null || true
if [ ! -z ${CUSTOM_RPM_FILES} ] ; then
    title "Building User-Specified RPM repository"
    mkdir custom-rpms
    for rpm in ${CUSTOM_RPM_FILES//,/ } ; do
        cp $rpm custom-rpms
    done
    createrepo custom-rpms >/dev/null
fi

title "Loading sources for OpenShift and MicroShift"
for f in openshift-local microshift-local custom-rpms ; do
    [ ! -d $f ] && continue
    cat ${SCRIPTDIR}/config/${f}.toml.template | sed "s;REPLACE_IMAGE_BUILDER_DIR;${BUILDDIR};g" > ${f}.toml
    sudo composer-cli sources delete $f 2>/dev/null || true
    sudo composer-cli sources add ${BUILDDIR}/${f}.toml
done

title "Preparing blueprints"
cp -f ${SCRIPTDIR}/config/{blueprint_v0.0.1.toml,installer.toml} .
if [ ! -z ${CUSTOM_RPM_FILES} ] ; then
    for rpm in ${CUSTOM_RPM_FILES//,/ } ; do
        rpm_name=$(rpm -qp $rpm --queryformat "%{NAME}")
        rpm_version=$(rpm -qp $rpm --queryformat "%{VERSION}")
        cat >> blueprint_v0.0.1.toml <<EOF

[[packages]]
name = "${rpm_name}"
version = "${rpm_version}"
EOF
    done
fi

# Add container images
if ${EMBED_CONTAINERS} ; then
    # TODO: This should be removed when RHEL 8.x stream gets an up-to-date package
    # Include up-to-date ostree packages in the image builder to support whiteouts    
    repo_name=ostree-copr
    cp -f ${SCRIPTDIR}/config/${repo_name}.toml .
    sudo composer-cli sources delete ${repo_name} 2>/dev/null || true
    sudo composer-cli sources add ${BUILDDIR}/${repo_name}.toml

    # Add the list of all the container images
    for img in $(${ROOTDIR}/pkg/release/get.sh images $(uname -i)) ; do
        cat >> blueprint_v0.0.1.toml <<EOF

[[containers]]
source = "${img}"
EOF
    done
fi

# Add the firewall customization required by Prometheus
if ${PROMETHEUS} ; then
    cat >> blueprint_v0.0.1.toml <<EOF

[customizations.firewall]
ports = ["9256:tcp"]
EOF
fi

build_image blueprint_v0.0.1.toml "${IMGNAME}-container" 0.0.1 edge-container
build_image installer.toml        "${IMGNAME}-installer" 0.0.0 edge-installer "${IMGNAME}-container" 0.0.1

title "Embedding kickstart in the installer image"
# Create a kickstart file from a template, compacting pull secret contents if necessary
cat "${SCRIPTDIR}/config/kickstart.ks.template" \
    | sed "s;REPLACE_LVM_SYSROOT_SIZE;${LVM_SYSROOT_SIZE};g" \
    | sed "s;REPLACE_OSTREE_SERVER_NAME;${OSTREE_SERVER_NAME};g" \
    | sed "s;REPLACE_OCP_PULL_SECRET_CONTENTS;$(cat $OCP_PULL_SECRET_FILE | jq -c);g" \
    | sed "s;REPLACE_REDHAT_AUTHORIZED_KEYS_CONTENTS;${AUTHORIZED_KEYS};g" \
    | sed "s;REPLACE_BUILD_ARCH;${BUILD_ARCH};g" \
    > kickstart.ks

# Run the ISO creation procedure
sudo mkksiso kickstart.ks ${IMGNAME}-installer-0.0.0-installer.iso ${IMGNAME}-installer-${IMAGE_VERSION}.${BUILD_ARCH}.iso
sudo chown -R $(whoami). "${BUILDDIR}"

# Remove intermediate artifacts to free disk space
rm -f ${IMGNAME}-installer-0.0.0-installer.iso

${SCRIPTDIR}/cleanup.sh

title "Done"
popd &>/dev/null
