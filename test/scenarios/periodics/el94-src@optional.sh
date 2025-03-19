#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.4-microshift-source-optionals
    # Two nics - one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm  --network "${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK}"
    local tempfile
    tempfile=$(mktemp)

    copy_file_from_vm "host1" /etc/microshift/opentelemetry-collector.yaml "${tempfile}"
    # add the debug exporter and set all pipelines to use only the debug exporter. This is necessary, otherwise
    # the opentelemetry-collector will fill logs with errors about being unable to reach the backend endpoint. Handling
    # the configuration here should keep the microshift-observability.service stable throughout the test.
    yq -i '.service.pipelines |= with_entries(.value.exporters |= [ "debug" ]) | .exporters["debug"] = {}' "${tempfile}"
    copy_file_to_vm host1 "${tempfile}" /etc/microshift/opentelemetry-collector.yaml
    run_command_on_vm host1 "sudo systemctl restart microshift-observability.service"
    run_command_on_vm host1 "sudo systemctl is-active microshift-observability.service"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/optional/
}
