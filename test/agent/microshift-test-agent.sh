#!/usr/bin/bash
set -xeuo pipefail

AGENT_CFG=/var/lib/microshift-test-agent.json
SYSTEMD_NOTIFIED=false

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
    # Make sure to always notify systemd if the script exited without
    # sending an explicit notification
    _notify_systemd

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

_get_current_deployment_id() {
    local -r id="$(rpm-ostree status --booted --json | jq -r ".deployments[0].id")"
    echo "${id}"
}

_notify_systemd() {
    # Avoid double notification
    if ${SYSTEMD_NOTIFIED} ; then
        return
    fi

    if [ -n "${NOTIFY_SOCKET:-}" ] ; then
        systemd-notify --ready
    fi
    SYSTEMD_NOTIFIED=true
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
# problem with groups. Following line ensures that sshd keys are owned by
# ssh_keys group and not some other.
# Note:
# - Only change group if the keys are not already owned by root because that is
#   the new expected default upstream
# - When removing the workaround, change Before= in microshift-test-agent.service
if [ -d /etc/ssh ] ; then
    find /etc/ssh -name 'ssh_host*key' -a '!' -gid 0 -exec chown -v root:ssh_keys {} \;
else
    echo "The /etc/ssh directory does not exist, skipping file ownership update"
fi
# WORKAROUND END

_debug_info

if [ ! -f "${AGENT_CFG}" ] ; then
    exit 0
fi

current_deployment_id="$(_get_current_deployment_id)"

deploy=$(jq -c ".\"${current_deployment_id}\"" "${AGENT_CFG}")
if [[ "${deploy}" == "null" ]]; then
    exit 0
fi

current_boot_actions=$(echo "${deploy}" | jq -c "[.[]] | flatten")

# greenboot-rs takes a different approach compared to bash greenboot implementation.
# bash greenboot: when deployment is staged, boot_counter is set immediately,
#     when host boots again, the variable is present on 1st boot of new deployment.
# greenboot-rs: boot_counter is set only when the healthchecks fail for the new deployment.
#
# Therefore, test-agent cannot depend on boot_counter anymore to do
# actions on first boot of the new deployment.
# For this reason, the way how the test agent config is interpreted changed:
# the .deployment.number is no longer the "ordinal boot number" (i.e. 1st, 2nd, 3rd boot of the deployment)
# but "how many boots this action should be active".
#
# After collecting actions for the current boot, numbers are decremented, and if reach 0,
# removed from the config.
jq \
    --arg key "${current_deployment_id}" \
    '.[$key] |= (with_entries(if (.key | test("^[0-9]+$")) then .key |= (tonumber - 1 | tostring) else . end) | with_entries(select(.key != "0")))' \
    "${AGENT_CFG}" > "${AGENT_CFG}.tmp" && \
    mv "${AGENT_CFG}.tmp" "${AGENT_CFG}"

_run_actions "${current_boot_actions}"

# If running under systemd, notify systemd that the service is ready so that
# other dependent services in the startup sequence can be started
_notify_systemd

# Sleep in background and wait to not miss the signals. If sleep command is
# interrupted, wait error is ignored and the loop continues. If the script
# is interrupted, it exits and calls cleanup commands.
while true ; do
    sleep infinity &
    wait $! || true
done
