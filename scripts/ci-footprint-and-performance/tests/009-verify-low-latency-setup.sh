#!/usr/bin/bash

set -xeuo pipefail

VARIABLES_FILE=/etc/tuned/microshift-baseline-variables.conf

fail() {
    local -r message=$1

    printf 'LOW-LATENCY SETUP FAILURE: %s\n' "${message}" >&2
    exit 1
}

get_tuned_variable() {
    local -r name=$1
    local value

    value=$(awk -F= -v name="${name}" '$1 == name { print substr($0, index($0, "=") + 1); exit }' "${VARIABLES_FILE}")
    if [[ -z "${value}" ]]; then
        fail "${name} is missing from ${VARIABLES_FILE}; rerun setup/001-configure-wp-lowlat.sh."
    fi

    printf '%s\n' "${value}"
}

require_kernel_argument() {
    local -r argument=$1
    local -r cmdline=$2

    if [[ " ${cmdline} " != *" ${argument} "* ]]; then
        fail "missing or invalid kernel argument '${argument}' in /proc/cmdline; reapply microshift-baseline and reboot."
    fi
}

check_cpu_state() {
    local -r cpu=$1
    local -r expected_state=$2
    local -r state_path="/sys/devices/system/cpu/cpu${cpu}/online"
    local state

    if [[ "${cpu}" -eq 0 ]]; then
        if [[ "${expected_state}" != 1 ]]; then
            fail "CPU 0 cannot be offlined, but the configured offline CPU set includes it."
        fi
        return
    fi

    if [[ ! -r "${state_path}" ]]; then
        fail "CPU ${cpu} is not present at ${state_path}; check the configured CPU sets."
    fi

    state=$(<"${state_path}")
    if [[ "${state}" != "${expected_state}" ]]; then
        fail "CPU ${cpu} is ${state}, expected ${expected_state}; reapply microshift-baseline and reboot."
    fi
}

check_cpu_list_state() {
    local -r cpu_list=$1
    local -r expected_state=$2
    local range start end cpu
    local IFS=,
    local -a ranges

    read -r -a ranges <<< "${cpu_list}"
    for range in "${ranges[@]}"; do
        range=${range//[[:space:]]/}
        if [[ "${range}" == *-* ]]; then
            start=${range%-*}
            end=${range#*-}
        else
            start=${range}
            end=${range}
        fi

        if ! [[ "${start}" =~ ^[0-9]+$ && "${end}" =~ ^[0-9]+$ && "${start}" -le "${end}" ]]; then
            fail "invalid CPU range '${range}' in ${VARIABLES_FILE}."
        fi

        for ((cpu = start; cpu <= end; cpu++)); do
            check_cpu_state "${cpu}" "${expected_state}"
        done
    done
}

if [[ ! -r "${VARIABLES_FILE}" ]]; then
    fail "${VARIABLES_FILE} is missing; setup/001-configure-wp-lowlat.sh did not complete."
fi

isolated_cores=$(get_tuned_variable isolated_cores)
hugepages_size=$(get_tuned_variable hugepages_size)
hugepages=$(get_tuned_variable hugepages)
offline_cpu_set=$(get_tuned_variable offline_cpu_set)

if ! kernel_rt_versions=$(rpm -q --queryformat '%{version}-%{release}.%{arch}\n' kernel-rt); then
    fail "kernel-rt is not installed; rerun setup/021-install-kernel-rt.sh."
fi
expected_kernel="$(printf '%s\n' "${kernel_rt_versions}" | sort -V | tail -n 1)+rt"
running_kernel=$(uname -r)
if [[ "${running_kernel}" != "${expected_kernel}" ]]; then
    fail "running kernel '${running_kernel}' does not match selected kernel-rt '${expected_kernel}'; reboot into the RT kernel."
fi

if ! default_kernel=$(sudo grubby --default-kernel); then
    fail "could not read the default kernel with grubby; verify the RPM bootloader configuration."
fi
if [[ "${default_kernel}" != "/boot/vmlinuz-${expected_kernel}" ]]; then
    fail "default kernel '${default_kernel}' does not match '${expected_kernel}'; rerun setup/021-install-kernel-rt.sh and reboot."
fi

if ! active_profile=$(sudo tuned-adm active); then
    fail "could not determine the active TuneD profile; ensure tuned is running and reapply microshift-baseline."
fi
if [[ "${active_profile}" != "Current active profile: microshift-baseline" ]]; then
    fail "active TuneD profile is '${active_profile}', expected microshift-baseline; rerun setup/022-enable-profile.sh and reboot."
fi

set +x
cmdline=$(</proc/cmdline)
require_kernel_argument nohz=on "${cmdline}"
require_kernel_argument "nohz_full=${isolated_cores}" "${cmdline}"
require_kernel_argument "rcu_nocbs=${isolated_cores}" "${cmdline}"
require_kernel_argument "hugepagesz=${hugepages_size}" "${cmdline}"
require_kernel_argument "hugepages=${hugepages}" "${cmdline}"
unset cmdline
set -x

check_cpu_list_state "${isolated_cores}" 1
check_cpu_list_state "${offline_cpu_set}" 0
