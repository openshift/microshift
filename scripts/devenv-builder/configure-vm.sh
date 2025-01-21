#!/bin/bash
set -eo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILD_AND_RUN=true
INSTALL_BUILD_DEPS=true
FORCE_FIREWALL=false
RHEL_SUBSCRIPTION=false
RHEL_BETA_VERSION=false
SET_RHEL_RELEASE=true
DNF_UPDATE=true
PULL_IMAGES=false
OPTIONAL_RPMS=false
DNF_RETRY="${SCRIPTDIR}/../dnf_retry.sh"
PULL_RETRY="${SCRIPTDIR}/../pull_retry.sh"
RHOCP_REPO="${SCRIPTDIR}/../get-latest-rhocp-repo.sh"
MAKE_VERSION="${SCRIPTDIR}/../../Makefile.version.$(uname -m).var"

start=$(date +%s)

function usage() {
    echo "Usage: $(basename "$0") [--no-build] [--no-build-deps] [--force-firewall] [--no-set-release-version] <openshift-pull-secret-file>"
    echo ""
    echo "  --no-build                Do not build, install and start MicroShift"
    echo "  --no-build-deps           Do not install dependencies for building binaries and RPMs (implies --no-build)"
    echo "  --force-firewall          Install and configure firewalld regardless of other options"
    echo "  --no-set-release-version  Do NOT set the release subscription to the current release version"
    echo "  --skip-dnf-update         Do NOT run dnf update"
    echo "  --pull-images             Force pulling of container images before starting MicroShift"
    echo "  --optional-rpms           Install optional RPMs (OLM, Multus) along with core MicroShift RPMs"

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

while [ $# -gt 1 ]; do
    case "$1" in
    --no-build)
        BUILD_AND_RUN=false
        shift
        ;;
    --no-build-deps)
        INSTALL_BUILD_DEPS=false
        BUILD_AND_RUN=false
        shift
        ;;
    --force-firewall)
        FORCE_FIREWALL=true
        shift
        ;;
    --no-set-release-version)
        SET_RHEL_RELEASE=false
        shift
        ;;
    --skip-dnf-update)
        DNF_UPDATE=false
        shift
        ;;
    --pull-images)
        PULL_IMAGES=true
        shift
        ;;
    --optional-rpms)
        OPTIONAL_RPMS=true
        shift
        ;;
    *) usage ;;
    esac
done

if [ $# -ne 1 ]; then
    usage "Wrong number of arguments"
fi

# The conditional check for presence of the used scripts and files
# is required because configure-vm.sh can be run as a standalone
# script when bootstrapping the development environment
if [ ! -x "${DNF_RETRY}" ] ; then
    DNF_RETRY=$(mktemp /tmp/dnf_retry.XXXXXXXX.sh)
    curl -s "https://raw.githubusercontent.com/openshift/microshift/main/scripts/dnf_retry.sh" -o "${DNF_RETRY}"
    chmod 755 "${DNF_RETRY}"
fi
if [ ! -x "${RHOCP_REPO}" ] ; then
    RHOCP_REPO=$(mktemp /tmp/get-latest-rhocp-repo.XXXXXXXX.sh)
    curl -s "https://raw.githubusercontent.com/openshift/microshift/main/scripts/get-latest-rhocp-repo.sh" -o "${RHOCP_REPO}"
    chmod 755 "${RHOCP_REPO}"
fi
if [ ! -f "${MAKE_VERSION}" ] ; then
    MAKE_VERSION=$(mktemp "/tmp/Makefile.version.$(uname -m).XXXXXXXX.var")
    curl -s "https://raw.githubusercontent.com/openshift/microshift/main/Makefile.version.$(uname -m).var" -o "${MAKE_VERSION}"
    chmod 644 "${MAKE_VERSION}"
fi

# Only RHEL requires a subscription
if grep -q 'Red Hat Enterprise Linux' /etc/redhat-release; then
    RHEL_SUBSCRIPTION=true
fi
# Detect RHEL Beta versions
if grep -qE 'Red Hat Enterprise Linux.*Beta' /etc/redhat-release; then
    RHEL_BETA_VERSION=true
fi

OCP_PULL_SECRET=$1
[ ! -e "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret file '${OCP_PULL_SECRET}' does not exist"
OCP_PULL_SECRET=$(realpath "${OCP_PULL_SECRET}")
[ ! -f "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret '${OCP_PULL_SECRET}' is not a regular file"

echo -e "${USER}\tALL=(ALL)\tNOPASSWD: ALL" | sudo tee "/etc/sudoers.d/${USER}"

# Check the subscription status and register if necessary
if ${RHEL_SUBSCRIPTION}; then
    if ! sudo subscription-manager status >&/dev/null; then
        sudo subscription-manager register --auto-attach
    fi
    sudo subscription-manager config --rhsm.manage_repos=1

    # Parse the OS versions and determine if EUS
    source /etc/os-release
    VERSION_ID_MAJOR="$(awk -F. '{print $1}' <<< "${VERSION_ID}")"
    VERSION_ID_MINOR="$(awk -F. '{print $2}' <<< "${VERSION_ID}")"
    VERSION_ID_EUS=""
    if (( "${VERSION_ID_MINOR}" % 2 == 0 )) ; then
        VERSION_ID_EUS="-eus"
    fi

    if ${SET_RHEL_RELEASE} && ! ${RHEL_BETA_VERSION} ; then
        # https://access.redhat.com/solutions/238533
        sudo subscription-manager release --set ${VERSION_ID}
        sudo subscription-manager release --show
        "${DNF_RETRY}" "clean" "all"
    fi

    # Enable RHEL CDN repos to avoid problems with incomplete RHUI mirrors
    if ! ${RHEL_BETA_VERSION} ; then
        sudo subscription-manager repos \
            --enable "rhel-${VERSION_ID_MAJOR}-for-$(uname -m)-baseos${VERSION_ID_EUS}-rpms" \
            --enable "rhel-${VERSION_ID_MAJOR}-for-$(uname -m)-appstream${VERSION_ID_EUS}-rpms"

        # Disable the AWS' Red Hat Update Infrastructure (RHUI) repositories.
        # Having enabled both RHUI and repos from SubMgr might cause conflicts.
        if dnf repolist --enabled | grep -q "^rhel-${VERSION_ID_MAJOR}-baseos-rhui-rpms"; then
            sudo dnf config-manager --set-disabled "rhel-${VERSION_ID_MAJOR}-baseos-rhui-rpms"
        fi
        if dnf repolist --enabled | grep -q "^rhel-${VERSION_ID_MAJOR}-appstream-rhui-rpms"; then
            sudo dnf config-manager --set-disabled "rhel-${VERSION_ID_MAJOR}-appstream-rhui-rpms"
        fi

        # Remove the rh-amazon-rhui-client as its upgrade process re-enables the RHUI repos.
        if rpm -q rh-amazon-rhui-client > /dev/null; then
            sudo dnf remove -y rh-amazon-rhui-client
        fi
    fi
fi

if ${INSTALL_BUILD_DEPS} || ${BUILD_AND_RUN}; then
    "${DNF_RETRY}" "clean" "all"
    if ${DNF_UPDATE}; then
        "${DNF_RETRY}" "update"
    fi
    "${DNF_RETRY}" "install" "gcc git golang cockpit make jq selinux-policy-devel rpm-build jq bash-completion avahi-tools createrepo"

    # run only if booted with systemd
    [[ -d /run/systemd/system ]] &&  sudo systemctl enable --now cockpit.socket
fi

GO_VER=1.22.9
GO_ARCH=$([ "$(uname -m)" == "x86_64" ] && echo "amd64" || echo "arm64")
GO_INSTALL_DIR="/usr/local/go${GO_VER}"
if ${INSTALL_BUILD_DEPS} && [ ! -d "${GO_INSTALL_DIR}" ]; then
    echo "Installing go ${GO_VER}..."
    # This is installed into different location (/usr/local/bin/go) from dnf installed Go (/usr/bin/go) so it doesn't conflict
    # /usr/local/bin is before /usr/bin in $PATH so newer one is picked up
    curl -L -o "go${GO_VER}.linux-${GO_ARCH}.tar.gz" "https://go.dev/dl/go${GO_VER}.linux-${GO_ARCH}.tar.gz"
    sudo rm -rf "/usr/local/go${GO_VER}"
    sudo mkdir -p "/usr/local/go${GO_VER}"
    sudo tar -C "/usr/local/go${GO_VER}" -xzf "go${GO_VER}.linux-${GO_ARCH}.tar.gz" --strip-components 1
    sudo rm -rfv /usr/local/bin/{go,gofmt}
    sudo ln --symbolic /usr/local/go${GO_VER}/bin/{go,gofmt} /usr/local/bin/
    rm -rfv "go${GO_VER}.linux-${GO_ARCH}.tar.gz"
fi

if ${BUILD_AND_RUN}; then
    if [ ! -e ~/microshift ]; then
        git clone https://github.com/openshift/microshift.git ~/microshift
    fi
    cd ~/microshift

    make clean
    make rpm
    make srpm
fi

if ${RHEL_SUBSCRIPTION}; then
    OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)
    sudo subscription-manager config --rhsm.manage_repos=1

    RHOCP=$("${RHOCP_REPO}")
    if [[ "${RHOCP}" =~ ^[0-9]{2} ]]; then
        sudo subscription-manager repos --enable "rhocp-4.${RHOCP}-for-rhel-9-$(uname -m)-rpms"
    elif [[ "${RHOCP}" =~ ^http ]]; then
        url=$(echo "${RHOCP}" | cut -d, -f1)
        ver=$(echo "${RHOCP}" | cut -d, -f2)
        OCP_REPO_NAME="rhocp-4.${ver}-for-rhel-9-mirrorbeta-$(uname -i)-rpms"
        sudo tee "/etc/yum.repos.d/${OCP_REPO_NAME}.repo" >/dev/null <<EOF
[${OCP_REPO_NAME}]
name=Beta rhocp-4.${ver} RPMs for RHEL 9
baseurl=${url}
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF
        PREVIOUS_RHOCP=$("${RHOCP_REPO}" $((ver-1)))
        if [[ "${PREVIOUS_RHOCP}" =~ ^[0-9]{2} ]]; then
            sudo subscription-manager repos --enable "rhocp-4.${PREVIOUS_RHOCP}-for-rhel-9-$(uname -m)-rpms"
        fi
    fi

    # Enable fast-datapath (ovn) only for non-beta. Beta RHEL will use openvswitch from RHOCP beta mirror.
    if ! ${RHEL_BETA_VERSION}; then
        sudo subscription-manager repos --enable "fast-datapath-for-rhel-${VERSION_ID_MAJOR}-$(uname -m)-rpms"
    fi
else
    "${DNF_RETRY}" "install" "centos-release-nfv-common"
    sudo dnf copr enable -y @OKD/okd "centos-stream-9-$(uname -m)"
    sudo tee "/etc/yum.repos.d/openvswitch2-$(uname -m)-rpms.repo" >/dev/null <<EOF
[sig-nfv]
name=CentOS Stream 9 - SIG NFV
baseurl=http://mirror.stream.centos.org/SIGs/9-stream/nfv/\$basearch/openvswitch-2/
gpgcheck=1
enabled=1
skip_if_unavailable=0
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-SIG-NFV
EOF
fi

if ${RHEL_SUBSCRIPTION}; then
    "${DNF_RETRY}" "install" "openshift-clients"
else
    # Assume the current development version on non-RHEL OS
    OCPVERSION="4.$(cut -d'.' -f2 "${MAKE_VERSION}")"
    OCC_SRC="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/${OCPVERSION}-el9-beta"
    OCC_RPM="$(curl -s "${OCC_SRC}/" | grep -o "openshift-clients-4[^\"']*.rpm" | sort | uniq)"
    OCC_LOC="$(mktemp /tmp/openshift-client-XXXXX.rpm)"
    OCC_REM="${OCC_SRC}/${OCC_RPM}"

    curl -s "${OCC_REM}" --output "${OCC_LOC}"
    "${DNF_RETRY}" "localinstall" "${OCC_LOC}"
    rm -f "${OCC_LOC}"
fi

if ${BUILD_AND_RUN}; then
    if ${OPTIONAL_RPMS}; then
        # Skip gateway api rpms because:
        # - Feature is still dev preview and no tests/docs are guaranteed.
        # - There is one issue with conformance (see USHIFT-4757) that needs to be addressed in the operator.
        "${DNF_RETRY}" "localinstall" "$(find ~/microshift/_output/rpmbuild/RPMS -type f -name "*.rpm" -not -name "*gateway-api*")"
    else
        createrepo "${HOME}/microshift/_output/rpmbuild"
        "${DNF_RETRY}" "install" \
            "microshift microshift-release-info \
            --repofrompath=microshift-local,${HOME}/microshift/_output/rpmbuild \
            --setopt=microshift-local.gpgcheck=0"
    fi
fi

# Configure OpenShift pull secret
if [ ! -e "/etc/crio/openshift-pull-secret" ]; then
    sudo mkdir -p /etc/crio/
    sudo cp -f "${OCP_PULL_SECRET}" /etc/crio/openshift-pull-secret
    sudo chmod 600 /etc/crio/openshift-pull-secret
fi

# Optionally configure crun runtime for crio, required on CentOS 9
if [ -e /etc/crio/crio.conf  ] ; then
    if ! grep -q '\[crio.runtime.runtimes.crun\]' /etc/crio/crio.conf ; then
        cat <<EOF | sudo tee -a /etc/crio/crio.conf &>/dev/null

[crio.runtime.runtimes.crun]
runtime_path = ""
runtime_type = "oci"
runtime_root = "/run/crun"
runtime_config_path = ""
monitor_path = ""
monitor_cgroup = "system.slice"
monitor_exec_cgroup = ""
privileged_without_host_devices = false
EOF
    fi
fi

if ${BUILD_AND_RUN} || ${FORCE_FIREWALL}; then
    "${DNF_RETRY}" "install" "firewalld"
    sudo systemctl enable firewalld --now
    sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
    sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
    sudo firewall-cmd --permanent --zone=trusted --add-source=fd01::/48
    sudo firewall-cmd --permanent --zone=public --add-port=80/tcp
    sudo firewall-cmd --permanent --zone=public --add-port=443/tcp
    sudo firewall-cmd --permanent --zone=public --add-port=5353/udp
    sudo firewall-cmd --permanent --zone=public --add-port=30000-32767/tcp
    sudo firewall-cmd --permanent --zone=public --add-port=30000-32767/udp
    sudo firewall-cmd --permanent --zone=public --add-port=6443/tcp
    sudo firewall-cmd --permanent --zone=public --add-service=mdns
    sudo firewall-cmd --reload
fi

if ${BUILD_AND_RUN}; then
    sudo systemctl enable --now crio
    if ${PULL_IMAGES}; then
        # shellcheck disable=SC2046
        "${PULL_RETRY}" $(rpm -qa | grep -e  "microshift.*-release-info" | xargs rpm -ql | grep $(uname -m).json | xargs jq -r '.images | values[]')
    fi
    sudo systemctl start microshift

    echo ""
    echo "The configuration phase completed. Run the following commands to:"
    echo " - Wait until all MicroShift pods are running"
    echo "      watch sudo \$(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A"
    echo ""
    echo " - Get MicroShift logs"
    echo "      sudo journalctl -u microshift"
    echo ""
    echo " - Get microshift-etcd logs"
    echo "      sudo journalctl -u microshift-etcd.scope"
    echo ""
    echo " - Clean up MicroShift service configuration"
    echo "      echo 1 | sudo microshift-cleanup-data --all"
fi

end="$(date +%s)"
duration_total_seconds=$((end - start))
duration_minutes=$((duration_total_seconds / 60))
duration_seconds=$((duration_total_seconds % 60))

echo ""
echo "Done in ${duration_minutes}m ${duration_seconds}s"
