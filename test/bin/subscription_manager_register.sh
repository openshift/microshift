#!/usr/bin/env bash

# Enable subscription in the VM
subscription_manager_register() {
    local -r vmname="$1"

    if [ -f /tmp/subscription-manager-org ]; then
        # CI workflow.
        # Create a subscription manager registration script which will run elevated.
        # This is a workaround to avoid sudo logging its command line containing
        # secrets in the system logs.
        local -r sub_script=$(mktemp /tmp/submgr_script.XXXXXXXX.sh)
        cat >"${sub_script}" <<'EOF'
#!/bin/bash
set -euo pipefail

if ! subscription-manager status; then
    for try in $(seq 3) ; do
        echo "Trying to register the system: attempt #${try}"
        if subscription-manager register --force \
                --org="$(cat /tmp/subscription-manager-org)" \
                --activationkey="$(cat /tmp/subscription-manager-act-key)"; then
            exit 0
        fi
        sleep 5
    done
fi
EOF
        copy_file_to_vm "${vmname}" "${sub_script}" "${sub_script}"
        copy_file_to_vm "${vmname}" /tmp/subscription-manager-org /tmp/subscription-manager-org
        copy_file_to_vm "${vmname}" /tmp/subscription-manager-act-key /tmp/subscription-manager-act-key
        run_command_on_vm "${vmname}" "chmod 600 /tmp/subscription-manager-org /tmp/subscription-manager-act-key"
        run_command_on_vm "${vmname}" "chmod 700 ${sub_script} && sudo ${sub_script}"
    else
        # Local developer workflow
        run_command_on_vm "${vmname}" "sudo subscription-manager register"
    fi
}
