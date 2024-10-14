#!/usr/bin/bash
set -xeuo pipefail

GREENBOOT_CONFIGURATION_FILE=/etc/greenboot/greenboot.conf
AGENT_CFG=/var/lib/microshift-test-agent.json

# Example config
# {
#     "deploy-id": {
#         "every": [ "prevent_backup" ],
#         "1": [ "fail_greenboot" ],
#         "2": [ "..." ],
#         "3": [ "..." ]
#     }
# }

CLEANUP_CMDS=()
_cleanup() {
    for cmd in "${CLEANUP_CMDS[@]}"; do
        ${cmd}
    done
    exit 0
}
trap "_cleanup" EXIT

_run_actions() {
    local -r actions="${1}"
    if [[ "${actions}" == "null" ]]; then
        return
    fi

    num=$(echo "${actions}" | jq -c ". | length")
    for ((i=0; i <num; i++)); do
        action=$(echo "${actions}" | jq -c -r ".[${i}]")

        if ! declare -F "${action}"; then
            echo "Unknown action: ${action}"
        else
            "${action}"
        fi
    done
}

_debug_info() {
    grub2-editenv - list || true
    ostree admin status -v || true
    rpm-ostree status -v || true
    bootc status || true
    journalctl --list-boots --reverse | head -n6 || true
    ls -lah /var/lib/ || true
    ls -lah /var/lib/microshift || true
    ls -lah /var/lib/microshift-backups || true
    cat "${AGENT_CFG}" || true
}

_get_current_boot_number() {
    if ! /usr/bin/grub2-editenv list | grep -q boot_counter; then
        echo "boot_counter is missing - script only for newly staged deployments"
        exit 0
    fi

    local -r boot_counter=$(/usr/bin/grub2-editenv list | grep boot_counter | sed 's,boot_counter=,,')
    local max_boot_attempts

    if test -f "${GREENBOOT_CONFIGURATION_FILE}"; then
        # shellcheck source=/dev/null
        source "${GREENBOOT_CONFIGURATION_FILE}"
    fi

    if [ -v GREENBOOT_MAX_BOOT_ATTEMPTS ]; then
        max_boot_attempts="${GREENBOOT_MAX_BOOT_ATTEMPTS}"
    else
        max_boot_attempts=3
    fi

    # When deployment is staged, greenboot sets boot_counter to 3
    # and this variable gets decremented on each boot.
    # First boot of new deployment will have it set to 2.
    echo "$((max_boot_attempts - boot_counter))"
}

_get_current_deployment_id() {
    local -r id="$(rpm-ostree status --booted --json | jq -r ".deployments[0].id")"
    echo "${id}"
}

prevent_backup() {
    local -r path="/var/lib/microshift-backups"
    if [[ ! -e "${path}" ]]; then
        mkdir -vp "${path}"

        # because of immutable attr, if the file does not exist, it can't be created
        touch "${path}/health.json"
    fi
    # prevents from creating a new backup directory, but doesn't prevent from updating health.json
    chattr -V +i "${path}"
    CLEANUP_CMDS+=("chattr -V -i ${path}")
}

fail_greenboot() {
    local -r path="/etc/greenboot/check/required.d/99_microshift_test_failure.sh"
    cat >"${path}" <<EOF
#!/usr/bin/bash
echo 'Forced greenboot failure by MicroShift Failure Agent'
sleep 5
exit 1
EOF
    chmod +x "${path}"
    CLEANUP_CMDS+=("rm -v ${path}")
}

# WORKAROUND START
# When going from "just RHEL" to RHEL+MicroShift+dependencies there might be a
# problem with groups. Following line ensures that sshd's keys are owned by
# ssh_keys group and not some other.
# When removing, change Before= in microshift-test-agent.service
if [ -d /etc/ssh ] ; then
    find /etc/ssh -name 'ssh_host*key' -exec chown -v root:ssh_keys {} \;
else
    echo "The /etc/ssh directory does not exist, skipping file ownership update"
fi
# WORKAROUND END

_debug_info

if [ ! -f "${AGENT_CFG}" ] ; then
    exit 0
fi

current_boot="$(_get_current_boot_number)"
current_deployment_id="$(_get_current_deployment_id)"

deploy=$(jq -c ".\"${current_deployment_id}\"" "${AGENT_CFG}")
if [[ "${deploy}" == "null" ]]; then
    exit 0
fi

every_boot_actions=$(echo "${deploy}" | jq -c ".\"every\"")
current_boot_actions=$(echo "${deploy}" | jq -c ".\"${current_boot}\"")

_run_actions "${every_boot_actions}"
_run_actions "${current_boot_actions}"

# If running under systemd, notify systemd that the service is ready so that
# other dependent services in the startup sequence can be started
if [ -n "${NOTIFY_SOCKET:-}" ] ; then
    systemd-notify --ready
fi

# Sleep in background and wait to not miss the signals. If sleep command is
# interrupted, wait error is ignored and the loop continues. If the script
# is interrupted, it exits and calls cleanup commands.
while true ; do
    sleep infinity &
    wait $! || true
done
