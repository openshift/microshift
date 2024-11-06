#!/bin/bash

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script must be sourced, not executed."
    exit 1
fi

# process_kickstart_into_script transforms kickstart file with its includes into a shell script.
# The script is executed on start of the bootc container and is VM's installation counterpart.
process_kickstart_into_script() {
    local -r container_name="$1"
    local -r file=${2:-kickstart.ks}

    local -r output_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${container_name}"

    while IFS= read -r line; do
        if [[ ${line} =~ ^%include[[:space:]]+(.+) ]]; then
            process_kickstart_into_script "${container_name}" "${BASH_REMATCH[1]}"
        else
            echo "${line}" >> "${output_dir}/init.sh"
        fi
    done < "${output_dir}/${file}"
}

# create_container_entrypoint prepares a script that will become a new entrypoint of a container,
# i.e.: it will be executed before original entrypoint.
create_container_entrypoint() {
    local -r name="$1"
    local -r output_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${name}/"

    cat << EOF > "${output_dir}/init.sh"
#!/bin/bash
set -xeuo pipefail

# Write new boot ID to simulate new boot when container (re)starts.
cat /proc/sys/kernel/random/uuid > /proc/sys/kernel/random/boot_id

# Check if marker exists to only run init one time - on first start of the container.
if [ -e /var/adm/microshift-container-init-ok ]; then
    echo "Container already initialized"
    exit 0
fi

EOF
    chmod +x "${output_dir}/init.sh"
    process_kickstart_into_script "${name}"

    cat << EOF >> "${output_dir}/init.sh"
mkdir -p /var/adm/
touch /var/adm/microshift-container-init-ok

EOF

    cat << 'EOF' > "${output_dir}/entrypoint.sh"
#!/bin/bash
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
bash "${SCRIPTDIR}/init.sh" &> /var/log/microshift_container_entrypoint.log
exec /sbin/init
EOF
    chmod +x "${output_dir}/entrypoint.sh"
}

# launch_container creates a new container with provided image reference.
#
# Usage: launch_container
#           --image <image>
#           [--name <name>]
#           [--vcpus <vcpus>]
#           [--memory <memory>]
#
# Arguments:
#   --image <image>: Bootc container image for the container.
#   [--name <name>]: The short name of the container in the scenario (e.g., "host1").
#   [--vcpus <vcpus>]: Number of vCPUs for the container (default: 2).
#   [--memory <memory>]: Size of RAM in MB for the container (default: 4096).
launch_container() {
    local name="host1"
    local image=""
    local memory=4096
    local vcpus=2

    while [ $# -gt 0 ]; do
        case "$1" in
            --name|--image|--vcpus|--memory)
                var="${1/--/}"
                if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
                    declare "${var}=$2"
                    shift 2
                else
                    error "Failed parsing arguments: ${var} value not set"
                    record_junit "${name}" "ctr-launch-args" "FAILED"
                    exit 1
                fi
                ;;
            *)
                error "Invalid argument: ${1}"
                record_junit "${name}" "ctr-launch-args" "FAILED"
                exit 1
                ;;
        esac
    done

    if [ -z "${image}" ]; then
        echo "--image is empty"
        exit 1
    fi

    local -r full_image="${MIRROR_REGISTRY_URL}/${image}"

    record_junit "${name}" "ctr-launch-args" "OK"

    local -r full_ctr_name=$(full_vm_name "${name}")

    if sudo podman inspect -f "{{.Id}}" "${full_ctr_name}" &> /dev/null; then
        echo "Container '${full_ctr_name}' already exists."
        record_junit "${name}" "launch_ctr" "SKIP"
        exit 0
    fi

    sudo modprobe openvswitch

    create_container_entrypoint "${name}"
    local -r output_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${name}/"
    touch "${output_dir}/boot_id"

    if ! sudo podman run \
        --detach \
        --privileged \
        --restart always \
        --hostname "${full_ctr_name}" \
        --volume /dev:/dev \
        --volume "${output_dir}/boot_id":/proc/sys/kernel/random/boot_id \
        --volume /var/lib/containers/storage:/var/lib/containers/storage \
        --volume "${output_dir}":/tmp/ctr_script \
        --entrypoint '/tmp/ctr_script/entrypoint.sh' \
        --memory "${memory}m" \
        --cpus "${vcpus}" \
        --name "${full_ctr_name}" \
        --tls-verify=false \
        "${full_image}"; then
        record_junit "${name}" "launch_ctr" "FAILED"
        exit 1
    fi

    # TODO: Poll
    local -r ip=$(sudo podman inspect "${full_ctr_name}" | jq -r '.[0].NetworkSettings.IPAddress')
    echo "Container ${full_ctr_name} has IP ${ip}"

    # Remove any previous key info for the host
    if [ -f "${HOME}/.ssh/known_hosts" ]; then
        echo "Clearing known_hosts entry for ${ip}"
        ssh-keygen -R "${ip}"
    fi

    set_vm_property "${name}" "ip" "${ip}"
    set_vm_property "${name}" "ssh_port" "22"
    set_vm_property "${name}" "api_port" "6443"
    set_vm_property "${name}" "lb_port" "5678"

    if wait_for_ssh "${ip}"; then
        record_junit "${name}" "ssh-access" "OK"
    else
        record_junit "${name}" "ssh-access" "FAILED"
        return 1
    fi

    if ! run_command_on_vm "${name}" "file /var/container-init-ok" > /dev/null; then
        echo "Container's init script didn't run to completion"
        record_junit "${name}" "launch_ctr" "FAILED"
        exit 1
    fi

    record_junit "${name}" "launch_ctr" "OK"
    echo "Container ${full_ctr_name} is up and ready"
}

remove_container() {
    local name="${1:-host1}"
    local full_ctr_name
    full_ctr_name=$(full_vm_name "${name}")
    sudo podman stop "${full_ctr_name}" > /dev/null || true
    sudo podman rm --force "${full_ctr_name}" > /dev/null || true
}
