#!/bin/bash
set -eo pipefail

ROOTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../../" && pwd )"
SCRIPTDIR=${ROOTDIR}/scripts/image-builder
IMGNAME=microshift
IMAGE_VERSION=undefined
BUILD_ARCH=$(uname -m)
OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)
OSTREE_SERVER_URL=file:///var/lib/ostree-local/repo
LVM_SYSROOT_SIZE_MIN=10240
LVM_SYSROOT_SIZE=${LVM_SYSROOT_SIZE_MIN}
OCP_PULL_SECRET_FILE=
MICROSHIFT_RPM_SOURCE=${ROOTDIR}/_output/rpmbuild/
AUTHORIZED_KEYS_FILE=
AUTHORIZED_KEYS=
CA_TRUST_FILES=
MIRROR_REGISTRY_HOST=
MIRROR_REGISTRY_INSECURE=true
PROMETHEUS=false
EMBED_CONTAINERS=false
BUILD_IMAGE_TYPE=container
# shellcheck disable=SC2034
STARTTIME="$(date +%s)"
BUILDDIR=${ROOTDIR}/_output/image-builder

usage() {
    local error_message="$1"

    if [ -n "${error_message}" ]; then
        echo "ERROR: ${error_message}"
        echo
    fi

    echo "Usage: $(basename "$0") <-pull_secret_file path_to_file> [OPTION]..."
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
    echo "          Embed the MicroShift container dependencies in the image"
    echo "  -ostree_server_url URL"
    echo "          URL of the ostree server (default: ${OSTREE_SERVER_URL})"
    echo "  -build_edge_commit"
    echo "          Build edge commit archive instead of an ISO image. The"
    echo "          archive contents can be used for serving ostree updates."
    echo "  -lvm_sysroot_size num_in_MB"
    echo "          Size of the system root LVM partition. The remaining"
    echo "          disk space will be allocated for data (default: ${LVM_SYSROOT_SIZE})"
    echo "  -authorized_keys_file path_to_file"
    echo "          Path to an SSH authorized_keys file to allow SSH access"
    echo "          into the default 'redhat' account"
    echo "  -open_firewall_ports port1[:protocol1],...,portN[:protocolN]"
    echo "          One or more comma-separated ports (optionally with protocol)"
    echo "          to be allowed by firewall (default: none)"
    echo "  -mirror_registry_host host[:port]"
    echo "          Host and optionally port of the mirror container registry to"
    echo "          be used by the container runtime when pulling images. The connection"
    echo "          to the mirror is configured as unsecure unless a CA trust certificate"
    echo "          is specified using -ca_trust_files parameter"
    echo "  -ca_trust_files /path/to/file1.pem,...,/path/to/fileN.pem"
    echo "          Path to one or more comma-separated public certificate files"
    echo "          to be included in the image at the /etc/pki/ca-trust/source/anchors"
    echo "          directory and installed using the update-ca-trust utility"
    echo "  -prometheus"
    echo "          Add Prometheus process exporter to the image. See"
    echo "          https://github.com/ncabatoff/process-exporter for more information"
    exit 1
}

title() {
    echo -e "\E[34m\n# $1\E[00m"
}

waitfor_image() {
    local uuid=$1

    local -r tstart=$(date +%s)
    echo "$(date +'%Y-%m-%d %H:%M:%S') STARTED"

    local status
    status=$(sudo composer-cli compose status | grep "${uuid}" | awk '{print $2}')
    while [[ "${status}" = "RUNNING" ]] || [[ "${status}" = "WAITING" ]]; do
        sleep 10
        status=$(sudo composer-cli compose status | grep "${uuid}" | awk '{print $2}')
        echo -en "$(date +'%Y-%m-%d %H:%M:%S') ${status}\r"
    done

    local -r tend=$(date +%s)
    echo "$(date +'%Y-%m-%d %H:%M:%S') ${status} - elapsed $(( (tend - tstart) / 60 )) minutes"

    if [ "${status}" = "FAILED" ]; then
        download_image "${uuid}" 1
        echo "Blueprint build has failed. For more information, review the downloaded logs"
        exit 1
    fi
}

download_image() {
    local uuid=$1
    local logsonly=$2

    sudo composer-cli compose logs "${uuid}"
    if [ -z "${logsonly}" ] ; then
        sudo composer-cli compose metadata "${uuid}"
        sudo composer-cli compose image "${uuid}"
    fi
    sudo chown -R "$(whoami)". "${BUILDDIR}"
}

build_image() {
    local blueprint_file=$1
    local blueprint=$2
    local version=$3
    local image_type=$4
    local parent_blueprint=$5
    local parent_version=$6
    local buildid

    title "Loading ${blueprint} blueprint v${version}"
    sudo composer-cli blueprints delete "${blueprint}" 2>/dev/null || true
    sudo composer-cli blueprints push "${BUILDDIR}/${blueprint_file}"
    sudo composer-cli blueprints depsolve "${blueprint}" 1>/dev/null

    if [ -n "${parent_version}" ]; then
        title "Serving ${parent_blueprint} v${parent_version} container locally"
        sudo podman rm -f "${parent_blueprint}-server" 2>/dev/null || true
        sudo podman rmi -f "localhost/${parent_blueprint}:${parent_version}" 2>/dev/null || true
        imageid=$(cat < "./${parent_blueprint}-${parent_version}-container.tar" | sudo podman load | grep -o -P '(?<=sha256[@:])[a-z0-9]*')
        sudo podman tag "${imageid}" "localhost/${parent_blueprint}:${parent_version}"
        sudo podman run -d --name="${parent_blueprint}-server" -p 8085:8080 "localhost/${parent_blueprint}:${parent_version}"

        title "Building ${image_type} for ${blueprint} v${version}, parent ${parent_blueprint} v${parent_version}"
        buildid=$(sudo composer-cli compose start-ostree --ref "rhel/${OSVERSION}/${BUILD_ARCH}/edge" --url http://localhost:8085/repo/ "${blueprint}" "${image_type}" | awk '/Compose/{print $2}')
    else
        title "Building ${image_type} for ${blueprint} v${version}"
        buildid=$(sudo composer-cli compose start-ostree --ref "rhel/${OSVERSION}/${BUILD_ARCH}/edge" "${blueprint}" "${image_type}" | awk '/Compose/{print $2}')
    fi

    waitfor_image "${buildid}"
    download_image "${buildid}"
    rename "${buildid}" "${blueprint}-${version}" "${buildid}"*.{tar,iso} 2>/dev/null || true
}

install_prometheus_rpm() {
    # Install prometheus process exporter
    if ${PROMETHEUS} ; then
        local -r owner=ncabatoff
        local -r repo=process-exporter
        local -r version="$(curl -s https://api.github.com/repos/${owner}/${repo}/releases/latest | jq -r '.name')"
        local -r url="https://github.com/${owner}/${repo}/releases/download/${version}/"
        local -r file="${repo}_${version#v}_linux_amd64.rpm"

        title "Downloading Prometheus exporter(s)"
        wget -q "${url}${file}"
        CUSTOM_RPM_FILES+="$(pwd)/${file},"
    fi
}

open_repo_permissions() {
    find "$1" -type f -exec chmod a+r  {} \;
    find "$1" -type d -exec chmod a+rx {} \;
}

comma_separated_files_readable() {
    local file_list="$1"
    for file in ${file_list//,/ } ; do
        if [ ! -r "${file}" ] ; then
            echo "The '${file}' input file is not readable or it does not exist"
            exit 1
        fi
    done
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
        comma_separated_files_readable "${CUSTOM_RPM_FILES}"
        shift
        ;;
    -embed_containers)
        EMBED_CONTAINERS=true
        shift
        ;;
    -ostree_server_url)
        shift
        OSTREE_SERVER_URL="$1"
        [ -z "${OSTREE_SERVER_URL}" ] && usage "ostree server URL not specified"
        shift
        ;;
    -build_edge_commit)
        BUILD_IMAGE_TYPE=commit
        shift
        ;;
    -lvm_sysroot_size)
        shift
        LVM_SYSROOT_SIZE="$1"
        [ -z "${LVM_SYSROOT_SIZE}" ] && usage "System root LVM partition size not specified"
        [ "${LVM_SYSROOT_SIZE}" -lt ${LVM_SYSROOT_SIZE_MIN} ] && usage "System root LVM partition size cannot be smaller than ${LVM_SYSROOT_SIZE_MIN}MB"
        shift
        ;;
    -authorized_keys_file)
        shift
        AUTHORIZED_KEYS_FILE="$1"
        [ -z "${AUTHORIZED_KEYS_FILE}" ] && usage "Authorized keys file not specified"
        shift
        ;;
     -open_firewall_ports)
        shift
        OPEN_FIREWALL_PORTS="$1"
        [ -z "${OPEN_FIREWALL_PORTS}" ] && usage "Firewall ports not specified"
        shift
        ;;
    -mirror_registry_host)
        shift
        MIRROR_REGISTRY_HOST="$1"
        [ -z "${MIRROR_REGISTRY_HOST}" ] && usage "Mirror registry host not specified"
        shift
        ;;
    -ca_trust_files)
        shift
        MIRROR_REGISTRY_INSECURE=false
        CA_TRUST_FILES="$1"
        [ -z "${CA_TRUST_FILES}" ] && usage "CA trust certificate files not specified"
        comma_separated_files_readable "${CA_TRUST_FILES}"
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

if [ -z "${OSTREE_SERVER_URL}" ] || [ -z "${OCP_PULL_SECRET_FILE}" ] ; then
    usage
fi
if [ ! -r "${OCP_PULL_SECRET_FILE}" ] ; then
    echo "ERROR: pull_secret_file file does not exist or not readable: ${OCP_PULL_SECRET_FILE}"
    exit 1
fi
if [ -n "${AUTHORIZED_KEYS_FILE}" ]; then
    if [ ! -e "${AUTHORIZED_KEYS_FILE}" ]; then
        echo "ERROR: authorized_keys_file does not exist: ${AUTHORIZED_KEYS_FILE}"
        exit 1
    else
        AUTHORIZED_KEYS=$(cat "${AUTHORIZED_KEYS_FILE}")
    fi
fi

mkdir -p "${BUILDDIR}"
pushd "${BUILDDIR}" &>/dev/null

# Also enter sudo password in the beginning if necessary
title "Checking available disk space"
build_disk=$(sudo df -k --output=avail . | tail -1)
if [ "${build_disk}" -lt 20971520 ] ; then
    echo "ERROR: Less then 20GB of disk space is available for the build"
    exit 1
fi

# Set the cleanup and elapsed time traps only if command line parsing was successful
trap '${SCRIPTDIR}/cleanup.sh; echo "Execution time: $(( ($(date +%s) - STARTTIME) / 60 )) minutes"' EXIT

title "Setting up local MicroShift repository"
# Copy MicroShift RPM packages
rm -rf microshift-local 2>/dev/null || true
if [[ "${MICROSHIFT_RPM_SOURCE}" == http* ]] ; then
    wget -q -nd -r -L -P microshift-local -A rpm "${MICROSHIFT_RPM_SOURCE}"
else
    [ ! -d "${MICROSHIFT_RPM_SOURCE}" ] && echo "MicroShift RPM path '${MICROSHIFT_RPM_SOURCE}' does not exist" && exit 1
    cp -TR "${MICROSHIFT_RPM_SOURCE}" microshift-local
fi

# Exit if no RPM packages were found
if [ "$(find microshift-local -name '*.rpm' | wc -l)" -eq 0 ] ; then
    echo "No RPM packages were found at '${MICROSHIFT_RPM_SOURCE}'. Exiting..."
    exit 1
fi
createrepo microshift-local >/dev/null
open_repo_permissions microshift-local

# Determine the version of microshift we have, for use in blueprint templates later.
MICROSHIFT_RELEASE_RPM=$(find microshift-local -name 'microshift-release-info*.rpm' | tail -n 1)
MICROSHIFT_VERSION=$(rpm -q --queryformat '%{version}' "${MICROSHIFT_RELEASE_RPM}")

# Determine the image version from the RPM contents
RELEASE_INFO_FILE=$(find . -name 'microshift-release-info-*.rpm' | tail -1)
if [ -z "${RELEASE_INFO_FILE}" ] ; then
    echo "Cannot find microshift-release-info RPM package at '${MICROSHIFT_RPM_SOURCE}'. Exiting..."
    exit 1
fi
rpm2cpio "${RELEASE_INFO_FILE}" | cpio --quiet -idm "*release-$(uname -m).json"
IMAGE_VERSION=$(jq -r '.release.base' "./usr/share/microshift/release/release-$(uname -m).json")
if [ -z "${IMAGE_VERSION}" ] ; then
    echo "Cannot determine image version from microshift-release-info RPM package at '${MICROSHIFT_RPM_SOURCE}'. Exiting..."
    exit 1
fi

# Install prometheus process exporter
install_prometheus_rpm

# Copy user-specific RPM packages
rm -rf custom-rpms 2>/dev/null || true
if [ -n "${CUSTOM_RPM_FILES}" ] ; then
    title "Building User-Specified RPM repository"
    mkdir custom-rpms
    for rpm in ${CUSTOM_RPM_FILES//,/ } ; do
        cp "${rpm}" custom-rpms
    done
    createrepo custom-rpms >/dev/null
    open_repo_permissions custom-rpms
fi

title "Loading package sources"

RHOCP=$("${ROOTDIR}/scripts/get-latest-rhocp-repo.sh")
if [[ "${RHOCP}" =~ ^[0-9]{2} ]]; then
    OCP_MINOR="${RHOCP}"
    RHOCP_URL="https://cdn.redhat.com/content/dist/layered/rhel9/${BUILD_ARCH}/rhocp/4.${OCP_MINOR}/os"
    RHOCP_GPG_RHSM="true"
elif [[ "${RHOCP}" =~ ^http ]]; then
    RHOCP_URL=$(echo "${RHOCP}" | cut -d, -f1)
    OCP_MINOR=$(echo "${RHOCP}" | cut -d, -f2)
    RHOCP_GPG_RHSM="false"
fi

PACKAGE_SOURCES="microshift-local rhocp fast-datapath"
if [ -n "${CUSTOM_RPM_FILES}" ] ; then
    PACKAGE_SOURCES+=" custom-rpms"
fi
# shellcheck disable=SC2086
for f in ${PACKAGE_SOURCES} ; do
    sed -e "s;REPLACE_IMAGE_BUILDER_DIR;${BUILDDIR};g" \
        -e "s;REPLACE_BUILD_ARCH;${BUILD_ARCH};g" \
        -e "s;REPLACE_OCP_MINOR;${OCP_MINOR};g" \
        -e "s;REPLACE_RHOCP_URL;${RHOCP_URL};g" \
        -e "s;REPLACE_RHOCP_GPG_RHSM;${RHOCP_GPG_RHSM};g" \
        "${SCRIPTDIR}/config/${f}.toml.template" \
        > ${f}.toml
    sudo composer-cli sources delete ${f} 2>/dev/null || true
    sudo composer-cli sources add "${BUILDDIR}/${f}.toml"
done

title "Preparing blueprints"
cp -f "${SCRIPTDIR}"/config/installer.toml .
sed -e "s;REPLACE_MICROSHIFT_VERSION;${MICROSHIFT_VERSION};g" \
    "${SCRIPTDIR}"/config/blueprint_v0.0.1.toml \
    >blueprint_v0.0.1.toml
if [ -n "${CUSTOM_RPM_FILES}" ] ; then
    for rpm in ${CUSTOM_RPM_FILES//,/ } ; do
        rpm_name=$(rpm -qp "${rpm}" --queryformat "%{NAME}")
        rpm_version=$(rpm -qp "${rpm}" --queryformat "%{VERSION}")
        cat >> blueprint_v0.0.1.toml <<EOF

[[packages]]
name = "${rpm_name}"
version = "${rpm_version}"
EOF
    done
fi

# Add container images
if ${EMBED_CONTAINERS} ; then
    # Add the list of all the container images
    jq -r '.images | .[] | ("\n[[containers]]\nsource = \"" + . + "\"\n")' \
        "./usr/share/microshift/release/release-$(uname -m).json" \
        >> blueprint_v0.0.1.toml
fi

# Configure registry mirror
if [ -n "${MIRROR_REGISTRY_HOST}" ] ; then
    cat < "${SCRIPTDIR}"/config/registries.conf.template | \
        sed "s;REPLACE_MIRROR_REGISTRY_HOST;${MIRROR_REGISTRY_HOST};g" | \
        sed "s;REPLACE_MIRROR_REGISTRY_INSECURE;${MIRROR_REGISTRY_INSECURE};g" | \
        sed 's;";\\";g' \
        > registries.conf

        cat >> blueprint_v0.0.1.toml <<EOF

[[customizations.files]]
path = "/etc/containers/registries.conf.d/999-microshift-mirror.conf"
data = "$(sed 's/$/\\n/g' < registries.conf | tr -d '\n')"
EOF
fi

# Add CA trust certificate file names and contents with new lines replaced by '\n'
if [ -n "${CA_TRUST_FILES}" ] ; then
    for cafile in ${CA_TRUST_FILES//,/ } ; do
        cat >> blueprint_v0.0.1.toml <<EOF

[[customizations.files]]
path = "/etc/pki/ca-trust/source/anchors/$(basename "${cafile}")"
data = "$(sed 's/$/\\n/g' < "${cafile}" | tr -d '\n')"
EOF
    done
fi

# Open the firewall ports required by Prometheus
if ${PROMETHEUS} ; then
    if [ -z "${OPEN_FIREWALL_PORTS}" ] ; then
        OPEN_FIREWALL_PORTS="9256:tcp"
    else
        OPEN_FIREWALL_PORTS+=",9256:tcp"
    fi
fi

# Add open firewall ports to the blueprint
if [ -n "${OPEN_FIREWALL_PORTS}" ] ; then
    cat >> blueprint_v0.0.1.toml <<EOF

[customizations.firewall]
ports = ["${OPEN_FIREWALL_PORTS//,/\", \"}"]
EOF
fi

title "Preparing kickstart"
# Create a kickstart file from a template, compacting pull secret contents if necessary
cat < "${SCRIPTDIR}/config/kickstart.ks.template" \
    | sed "s;REPLACE_LVM_SYSROOT_SIZE;${LVM_SYSROOT_SIZE};g" \
    | sed "s;REPLACE_OSTREE_SERVER_URL;${OSTREE_SERVER_URL};g" \
    | sed "s;REPLACE_OCP_PULL_SECRET_CONTENTS;$(cat < "${OCP_PULL_SECRET_FILE}" | jq -c);g" \
    | sed "s^REPLACE_REDHAT_AUTHORIZED_KEYS_CONTENTS^${AUTHORIZED_KEYS}^g" \
    | sed "s;REPLACE_OSVERSION;${OSVERSION};g" \
    | sed "s;REPLACE_BUILD_ARCH;${BUILD_ARCH};g" \
    > kickstart.ks

# Build the edge container/commit image using the blueprint
build_image blueprint_v0.0.1.toml "${IMGNAME}" 0.0.1 edge-${BUILD_IMAGE_TYPE}

if [ "${BUILD_IMAGE_TYPE}" = "commit" ] ; then
    # Nothing else to be done for edge commits
    title "Edge commit created"
    echo "The contents of the archive can be used for serving ostree updates:"
    ls -1 "${BUILDDIR}/${IMGNAME}-0.0.1-${BUILD_IMAGE_TYPE}.tar"
else
    # Create an ISO installer for edge containers
    build_image installer.toml "${IMGNAME}-installer" 0.0.0 edge-installer "${IMGNAME}" 0.0.1

    title "Embedding kickstart in the installer image"
    sudo mkksiso kickstart.ks "${IMGNAME}-installer-0.0.0-installer.iso" "${IMGNAME}-installer-${IMAGE_VERSION}.${BUILD_ARCH}.iso"

    # Remove intermediate artifacts to free disk space
    rm -f "${IMGNAME}-installer-0.0.0-installer.iso"
fi

# Update the build output files to be owned by the current user
sudo chown -R "$(whoami)". "${BUILDDIR}"

title "Done"
popd &>/dev/null
