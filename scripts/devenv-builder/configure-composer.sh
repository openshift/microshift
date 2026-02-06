#!/bin/bash
set -euxo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DNF_RETRY="${SCRIPTDIR}/../dnf_retry.sh"

install_and_configure_composer() {
    local -r version_id=$1
    local -r version_id_major="$(awk -F. '{print $1}' <<< "${version_id}")"

    "${DNF_RETRY}" "install" \
        "osbuild osbuild-composer \
        git composer-cli ostree rpm-ostree \
        cockpit-composer bash-completion podman runc genisoimage \
        createrepo yum-utils selinux-policy-devel jq wget lorax rpm-build \
        containernetworking-plugins expect httpd-tools vim-common"

    # The mock utility comes from the EPEL repository
    "${DNF_RETRY}" "install" "https://dl.fedoraproject.org/pub/epel/epel-release-latest-${version_id_major}.noarch.rpm"
    "${DNF_RETRY}" "install" "mock nginx tomcli parallel aria2"
    sudo usermod -a -G mock "$(whoami)"

    # Necessary for embedding container images
    if [ ! -e /etc/osbuild-worker/pull-secret.json ] ; then
        sudo mkdir -p /etc/osbuild-worker
        sudo ln -sf /etc/crio/openshift-pull-secret /etc/osbuild-worker/pull-secret.json
        sudo tee /etc/osbuild-worker/osbuild-worker.toml &>/dev/null <<EOF
[containers]
auth_file_path = "/etc/osbuild-worker/pull-secret.json"
EOF
    fi
}

enable_or_restart_composer_services() {
    local -r composer_active=$(sudo systemctl is-active osbuild-composer.service || true)

    sudo systemctl enable osbuild-composer.socket --now
    if [[ "${composer_active}" == "active" ]]; then
        # If composer was active before, restart it to make kernel-rt repository configuration active
        sudo systemctl restart osbuild-composer.service
    fi
    sudo systemctl enable cockpit.socket --now
    sudo firewall-cmd --add-service=cockpit --permanent
}

check_umask_and_permissions() {
    # Verify umask and home directory permissions
    local -r test_file=$(mktemp /tmp/configure-perm-test.XXXXX)

    touch "${test_file}.file"
    mkdir "${test_file}.dir"
    local -r home_perm=$(stat -c 0%a ~)
    local -r file_perm=$(stat -c 0%a "${test_file}.file")
    local -r dir_perm=$(stat -c 0%a "${test_file}.dir")

    # Set the correct permissions for osbuild-composer
    [ "${home_perm}" -lt 0711 ]  && chmod go+x ~

    if [ "${file_perm}" -lt 0644 ] || [ "${dir_perm}" -lt 0711 ] ; then
        echo "Check ${test_file}.dir permissions. The umask setting must allow execute to group/others"
        echo "Check ${test_file}.file permissions. The umask setting must allow read to group/others"
        exit 1
    fi

    # Cleanup
    rm -rf "${test_file}"*
}

enable_rt_repositories() {
    local -r version_id=$1
    local -r composer_config=$2

    # Enable RT repository by duplicating the 'baseos' repository, changing its name,
    # and replacing 'baseos' with 'rt'.
    # Note that kernel-rt is only available for x86_64.
    "${SCRIPTDIR}/../fetch_tools.sh" yq
    sudo mkdir -p "$(dirname "${composer_config}")"
    "${SCRIPTDIR}/../../_output/bin/yq" \
        '.["x86_64"] += (.["x86_64"][0] | .name = "kernel-rt" | .baseurl |= sub("baseos", "rt"))' \
        "/usr/share/osbuild-composer/repositories/rhel-${version_id}.json" | jq | sudo tee "${composer_config}" >/dev/null
}

enable_beta_or_eus_repositories() {
    local -r version_id=$1
    local -r composer_config=$2

    local -r version_id_major="$(awk -F. '{print $1}' <<< "${version_id}")"
    local -r version_id_minor="$(awk -F. '{print $2}' <<< "${version_id}")"

    local version_id_eus="dist"
    if (( "${version_id_minor}" % 2 == 0 )) ; then
        version_id_eus="eus"
    fi

    # The configuration will remain unchanged for non-beta and non-EUS operating systems.
    if grep -qE "Red Hat Enterprise Linux.*Beta" /etc/redhat-release; then
        sudo sed -i "s,dist/rhel${version_id_major}/${version_id},beta/rhel${version_id_major}/${version_id_major},g" "${composer_config}"
    else
        sudo sed -i "s,dist/rhel${version_id_major}/${version_id}/$(uname -m)/baseos/,${version_id_eus}/rhel${version_id_major}/${version_id}/$(uname -m)/baseos/,g" "${composer_config}"
        sudo sed -i "s,dist/rhel${version_id_major}/${version_id}/$(uname -m)/appstream/,${version_id_eus}/rhel${version_id_major}/${version_id}/$(uname -m)/appstream/,g" "${composer_config}"
    fi
    # If the host OS is configured to use the internal repo, overwrite the composer configuration to match
    if dnf repolist | grep -q download.eng.brq.redhat.com; then
        # The gpgkey from /usr/share/osbuild-composer/repositories is valid and common for all repos
        local -r gpgkey=$(ARCH=$(uname -m) jq '.[env.ARCH][] | select(.name=="baseos") | .gpgkey' /usr/share/osbuild-composer/repositories/rhel-"${version_id}".json)
        sudo tee "${composer_config}" &>/dev/null <<EOF
{
  "$(uname -m)": [
    {
      "name": "baseos",
      "baseurl": "http://download.eng.brq.redhat.com/rhel-${version_id_major}/nightly/RHEL-${version_id_major}/latest-RHEL-${version_id}/compose/BaseOS/$(uname -m)/os",
      "gpgkey": ${gpgkey},
      "rhsm": false,
      "check_gpg": true
    },
    {
      "name": "appstream",
      "baseurl": "http://download.eng.brq.redhat.com/rhel-${version_id_major}/nightly/RHEL-${version_id_major}/latest-RHEL-${version_id}/compose/AppStream/$(uname -m)/os",
      "gpgkey": ${gpgkey},
      "rhsm": false,
      "check_gpg": true
    },
    {
      "name": "rt",
      "baseurl": "http://download.eng.brq.redhat.com/rhel-${version_id_major}/nightly/RHEL-${version_id_major}/latest-RHEL-${version_id}/compose/RT/$(uname -m)/os",
      "gpgkey": ${gpgkey},
      "rhsm": false,
      "check_gpg": true
    }
  ]
}
EOF
    fi
}

enable_ocp_mirror_repositories() {
    local -r version_id=$1
    local -r composer_config=$3

    local version_id_ocp=$2
    if [ "$(uname -m)" = "aarch64" ]; then
        version_id_ocp="${version_id_ocp}_aarch64"
    fi

    # Check if a released version of the configuration file exists
    local -r config_file="/usr/share/osbuild-composer/repositories/rhel-${version_id}.json"
    if [ -f "${config_file}" ]; then
        echo "WARNING: Skipping pre-release RHEL repository configuration for version '${version_id}'"
        echo "WARNING: Using '${config_file}' configuration file"
        return
    fi

    # Check if OCP mirror credentials are present
    local -r ocp_mirror_ufile="${HOME}/.ocp_mirror_username"
    local -r ocp_mirror_pfile="${HOME}/.ocp_mirror_password"
    if ! [ -f "${ocp_mirror_ufile}" ] || ! [ -f "${ocp_mirror_pfile}" ]; then
        echo "WARNING: OCP mirror credentials are not present"
        return
    fi

    # Read the OCP mirror credentials from the files
    local -r ocp_mirror_username=$(cat "${ocp_mirror_ufile}")
    local -r ocp_mirror_password=$(cat "${ocp_mirror_pfile}")
    local -r version_id_short="$(tr -d '.' <<< "${version_id}")"

    # Create the configuration file in the composer configuration directory
    sudo mkdir -p "$(dirname "${composer_config}")"
    sudo tee "${composer_config}" &>/dev/null <<EOF
{
  "$(uname -m)": [
    {
      "name": "baseos",
      "baseurl": "https://${ocp_mirror_username}:${ocp_mirror_password}@mirror2.openshift.com/enterprise/reposync/${version_id_ocp}/rhel-${version_id_short}-baseos/",
      "rhsm": false,
      "check_gpg": false
    },
    {
      "name": "appstream",
      "baseurl": "https://${ocp_mirror_username}:${ocp_mirror_password}@mirror2.openshift.com/enterprise/reposync/${version_id_ocp}/rhel-${version_id_short}-appstream/",
      "rhsm": false,
      "check_gpg": false
    }
  ]
}
EOF
    sudo chmod 0640 "${composer_config}"
    sudo chgrp _osbuild-composer "${composer_config}"
}

#
# Main
#

# Read the current OS version (i.e. VERSION_ID)
source /etc/os-release

# shellcheck disable=SC2153
install_and_configure_composer "${VERSION_ID}"
check_umask_and_permissions

# Configure repositories for the current OS
enable_rt_repositories          "${VERSION_ID}" "/etc/osbuild-composer/repositories/rhel-${VERSION_ID}.json"
enable_beta_or_eus_repositories "${VERSION_ID}" "/etc/osbuild-composer/repositories/rhel-${VERSION_ID}.json"

# Configure OCP mirror repositories for pre-release versions
enable_ocp_mirror_repositories "9.8" "4.22" "/etc/osbuild-composer/repositories/rhel-9.8.json"

# This step must come in the end to make sure all the potential configuration
# changes are picked up by the service
enable_or_restart_composer_services
