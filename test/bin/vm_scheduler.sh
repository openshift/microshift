#!/bin/bash
#
# Dynamic VM Scheduler for MicroShift Scenario Testing
#
# This scheduler manages VM resources for parallel scenario execution,
# implementing VM reuse for compatible scenarios and queuing when
# host resources are exhausted.
#
# Usage:
#   vm_scheduler.sh orchestrate <scenario_dir>  - Schedule all scenarios in directory
#   vm_scheduler.sh status                      - Show current scheduler state
#
# ============================================================================
# LOCKING MECHANISMS
# ============================================================================
#
# This scheduler uses directory-based locking with mkdir for atomic operations.
# Lock files are stored in: ${SCHEDULER_STATE_DIR}/locks/<name>.lock
#
# Lock: vm_dispatch
# -----------------
# Purpose: Protects all VM dispatch operations to prevent race conditions when
#          multiple background scenarios complete simultaneously.
#
# Critical sections protected:
#   1. dispatch_dynamic_scenarios():
#      - Iterating queued scenarios
#      - Finding compatible free VMs (find_compatible_vm)
#      - Checking if resources are available (can_allocate)
#      - Assigning VMs or creating new ones
#
#   2. run_scenario_on_vm():
#      - Releasing a VM after scenario completion
#      - Searching for the next compatible queued scenario
#      - Either reusing the VM for another scenario or destroying it
#
# Implicit Lock: VM Name Generation
# ----------------------------------
# Function: generate_dynamic_vm_name()
# Mechanism: Uses mkdir "${VM_REGISTRY}/${vm_name}" as an atomic claim operation.
# Purpose: Ensures unique VM names (dynamic-vm-001, dynamic-vm-002, etc.) when
#          multiple processes try to create VMs concurrently.
#
# ============================================================================
# SCENARIO REQUIREMENTS
# ============================================================================
#
# Scenarios opt-in to dynamic scheduling by defining dynamic_schedule_requirements().
# Supported fields:
#   - min_vcpus:    Minimum vCPUs required (default: DEFAULT_VM_VCPUS)
#   - min_memory:   Minimum memory in MB (default: DEFAULT_VM_MEMORY)
#   - min_disksize: Minimum disk size in GB
#   - networks:     Required networks (empty = network-agnostic, can run on any VM)
#   - boot_image:   Required boot image
#   - fips:         Whether FIPS mode is required
#   - slow:         Set to "true" for long-running tests (e.g., router, conformance)
#                   Slow tests get priority within their reusability tier to avoid
#                   becoming the critical path.
#
# ============================================================================

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# Scheduler state directory
SCHEDULER_STATE_DIR="${IMAGEDIR}/scheduler-state"
VM_REGISTRY="${SCHEDULER_STATE_DIR}/vms"
SCENARIO_QUEUE="${SCHEDULER_STATE_DIR}/queue"
SCENARIO_STATUS="${SCHEDULER_STATE_DIR}/scenarios"
LOCK_DIR="${SCHEDULER_STATE_DIR}/locks"
SCHEDULER_LOG="${SCHEDULER_STATE_DIR}/scheduler.log"

# VM pool settings (from common.sh)
VM_POOL_BASENAME="vm-storage"
VM_DISK_BASEDIR="${IMAGEDIR}/${VM_POOL_BASENAME}"

# Scenario info directory (also from common.sh)
SCENARIO_INFO_DIR="${SCENARIO_INFO_DIR:-${IMAGEDIR}/scenario-info}"

# Host resource limits (configurable via environment, defaults from system)
# Detect system resources if not specified
_SYSTEM_VCPUS=$(nproc 2>/dev/null || echo 8)
_SYSTEM_MEMORY_KB=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}' || echo 8388608)
_SYSTEM_MEMORY_MB=$((_SYSTEM_MEMORY_KB / 1024))

HOST_TOTAL_VCPUS="${HOST_TOTAL_VCPUS:-${_SYSTEM_VCPUS}}"
HOST_TOTAL_MEMORY="${HOST_TOTAL_MEMORY:-${_SYSTEM_MEMORY_MB}}"

# System reserved resources (for host OS, hypervisor overhead, etc.)
SYSTEM_RESERVED_VCPUS="${SYSTEM_RESERVED_VCPUS:-2}"
SYSTEM_RESERVED_MEMORY="${SYSTEM_RESERVED_MEMORY:-4096}" 

# Available resources after system reservation
HOST_AVAILABLE_VCPUS=$((HOST_TOTAL_VCPUS - SYSTEM_RESERVED_VCPUS))
HOST_AVAILABLE_MEMORY=$((HOST_TOTAL_MEMORY - SYSTEM_RESERVED_MEMORY))

# Timeout for test execution (in seconds)
# Scenarios can override via test_timeout in requirements
VM_TEST_TIMEOUT="${VM_TEST_TIMEOUT:-3600}"

# Calculated resource requirements (populated during planning phase)
declare -i STATIC_TOTAL_VCPUS=0
declare -i STATIC_TOTAL_MEMORY=0
declare -i DYNAMIC_AVAILABLE_VCPUS=0
declare -i DYNAMIC_AVAILABLE_MEMORY=0
declare -i MAX_DYNAMIC_VCPUS=0      # Largest dynamic scenario vCPU requirement
declare -i MAX_DYNAMIC_MEMORY=0     # Largest dynamic scenario memory requirement

# Current resource usage by dynamic VMs (tracked during execution)
declare -i current_vcpus=0
declare -i current_memory=0


log() {
    local timestamp
    timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    echo "[${timestamp}] $*" >> "${SCHEDULER_LOG}"
    echo "[${timestamp}] $*" >&2
}

init_scheduler() {
    mkdir -p "${SCHEDULER_STATE_DIR}"

    log "Initializing scheduler state directory"
    mkdir -p "${VM_REGISTRY}" "${SCENARIO_QUEUE}" "${SCENARIO_STATUS}" "${LOCK_DIR}"

    rm -rf "${VM_REGISTRY:?}"/* "${SCENARIO_QUEUE:?}"/* "${SCENARIO_STATUS:?}"/* "${LOCK_DIR:?}"/*

    current_vcpus=0
    current_memory=0
    STATIC_TOTAL_VCPUS=0
    STATIC_TOTAL_MEMORY=0
    DYNAMIC_AVAILABLE_VCPUS=0
    DYNAMIC_AVAILABLE_MEMORY=0
    MAX_DYNAMIC_VCPUS=0
    MAX_DYNAMIC_MEMORY=0

    log "Scheduler state initialized"
    log "  Host total: vcpus=${HOST_TOTAL_VCPUS}, memory=${HOST_TOTAL_MEMORY}MB"
    log "  System reserved: vcpus=${SYSTEM_RESERVED_VCPUS}, memory=${SYSTEM_RESERVED_MEMORY}MB"
    log "  Available for VMs: vcpus=${HOST_AVAILABLE_VCPUS}, memory=${HOST_AVAILABLE_MEMORY}MB"
}

# Default VM resources when not specified in launch_vm
# Using 4 vCPUs by default maximizes VM reuse potential:
# - All VMs can run any scenario (no vcpu mismatch)
# - More queuing = more reuse opportunities
DEFAULT_VM_VCPUS=4
DEFAULT_VM_MEMORY=4096

parse_static_scenario_resources() {
    local scenario_script="$1"

    local launch_vm_line
    launch_vm_line=$(grep -E '^\s*launch_vm' "${scenario_script}" 2>/dev/null | head -1) || true

    local vcpus="${DEFAULT_VM_VCPUS}"
    local memory="${DEFAULT_VM_MEMORY}"

    if [ -n "${launch_vm_line}" ]; then
        if [[ "${launch_vm_line}" =~ --vm_vcpus[[:space:]]+([0-9]+) ]]; then
            vcpus="${BASH_REMATCH[1]}"
        fi
        if [[ "${launch_vm_line}" =~ --vm_memory[[:space:]]+([0-9]+) ]]; then
            memory="${BASH_REMATCH[1]}"
        fi
    fi

    echo "${vcpus} ${memory}"
}

calculate_static_requirements() {
    local -a static_scenarios=("$@")

    STATIC_TOTAL_VCPUS=0
    STATIC_TOTAL_MEMORY=0

    for scenario_script in "${static_scenarios[@]}"; do
        local resources
        resources=$(parse_static_scenario_resources "${scenario_script}")
        local vcpus memory
        read -r vcpus memory <<< "${resources}"

        STATIC_TOTAL_VCPUS=$((STATIC_TOTAL_VCPUS + vcpus))
        STATIC_TOTAL_MEMORY=$((STATIC_TOTAL_MEMORY + memory))

        local scenario_name
        scenario_name=$(basename "${scenario_script}" .sh)
        log "  Static ${scenario_name}: vcpus=${vcpus}, memory=${memory}MB"
    done

    log "Static total: vcpus=${STATIC_TOTAL_VCPUS}, memory=${STATIC_TOTAL_MEMORY}MB"
}

calculate_max_dynamic_requirements() {
    local -a dynamic_scenarios=("$@")

    MAX_DYNAMIC_VCPUS=0
    MAX_DYNAMIC_MEMORY=0

    for scenario_script in "${dynamic_scenarios[@]}"; do
        local scenario_name
        scenario_name=$(basename "${scenario_script}" .sh)

        local req_file="${SCENARIO_STATUS}/${scenario_name}/requirements"
        mkdir -p "$(dirname "${req_file}")"

        if ! get_scenario_requirements "${scenario_script}" "${req_file}"; then
            log "WARNING: Failed to get requirements for ${scenario_name}"
            continue
        fi

        local vcpus memory
        vcpus=$(get_req_value "${req_file}" "min_vcpus" "${DEFAULT_VM_VCPUS}")
        memory=$(get_req_value "${req_file}" "min_memory" "${DEFAULT_VM_MEMORY}")

        log "  Dynamic ${scenario_name}: vcpus=${vcpus}, memory=${memory}MB"

        if [ "${vcpus}" -gt "${MAX_DYNAMIC_VCPUS}" ]; then
            MAX_DYNAMIC_VCPUS="${vcpus}"
        fi
        if [ "${memory}" -gt "${MAX_DYNAMIC_MEMORY}" ]; then
            MAX_DYNAMIC_MEMORY="${memory}"
        fi
    done

    log "Max dynamic scenario: vcpus=${MAX_DYNAMIC_VCPUS}, memory=${MAX_DYNAMIC_MEMORY}MB"
}

acquire_lock() {
    local lock_name="$1"
    local lock_file="${LOCK_DIR}/${lock_name}.lock"

    while ! mkdir "${lock_file}" 2>/dev/null; do
        sleep 1
    done
}

release_lock() {
    local lock_name="$1"
    local lock_file="${LOCK_DIR}/${lock_name}.lock"
    rmdir "${lock_file}" 2>/dev/null || true
}

create_static_vms() {
    local -a static_scenarios=("$@")

    if [ ${#static_scenarios[@]} -eq 0 ]; then
        log "No static VMs to create"
        return 0
    fi

    log "Creating ${#static_scenarios[@]} static VMs in parallel"

    local static_create_log="${SCHEDULER_STATE_DIR}/static_create_jobs.txt"

    local progress=""
    if [ -t 0 ]; then
        progress="--progress"
    fi

    if ! parallel \
        ${progress} \
        --results "${SCENARIO_INFO_DIR}/{/.}/boot.log" \
        --joblog "${static_create_log}" \
        --delay 5 \
        bash -x "${SCRIPTDIR}/scenario.sh" create ::: "${static_scenarios[@]}"; then
        log "ERROR: Some static VMs failed to create"
        cat "${static_create_log}"
        return 1
    fi

    cat "${static_create_log}"
    log "All static VMs created successfully"
    return 0
}

run_static_tests() {
    local -a static_scenarios=("$@")

    if [ ${#static_scenarios[@]} -eq 0 ]; then
        log "No static tests to run"
        return 0
    fi

    log "Running ${#static_scenarios[@]} static scenario tests in parallel"

    local static_run_log="${SCHEDULER_STATE_DIR}/static_run_jobs.txt"

    local progress=""
    if [ -t 0 ]; then
        progress="--progress"
    fi

    local result=0
    if ! parallel \
        ${progress} \
        --results "${SCENARIO_INFO_DIR}/{/.}/run.log" \
        --joblog "${static_run_log}" \
        --delay 2 \
        bash -x "${SCRIPTDIR}/scenario.sh" run ::: "${static_scenarios[@]}"; then
        result=1
    fi

    cat "${static_run_log}"
    return ${result}
}

generate_dynamic_vm_name() {
    local count=1
    local vm_name
    while true; do
        vm_name="dynamic-vm-$(printf '%03d' "${count}")"
        if mkdir "${VM_REGISTRY}/${vm_name}" 2>/dev/null; then
            echo "${vm_name}"
            return 0
        fi
        ((count++)) || true
        # Safety limit to prevent infinite loop
        if [ ${count} -gt 999 ]; then
            echo "ERROR: Too many VMs" >&2
            return 1
        fi
    done
}

scenario_is_dynamic() {
    local scenario_script="$1"

    if bash -c "source '${scenario_script}' 2>/dev/null && type dynamic_schedule_requirements &>/dev/null"; then
        return 0
    fi
    return 1
}

get_scenario_requirements() {
    local scenario_script="$1"
    local output_file="$2"

    # shellcheck disable=SC1090
    (
        source "${scenario_script}"
        if type dynamic_schedule_requirements &>/dev/null; then
            dynamic_schedule_requirements > "${output_file}"
        else
            # This shouldn't happen if scenario_is_dynamic was checked first
            echo "ERROR: dynamic_schedule_requirements not found" >&2
            return 1
        fi
    )
}

get_req_value() {
    local req_file="$1"
    local key="$2"
    local default="${3:-}"

    local value
    value=$(grep "^${key}=" "${req_file}" 2>/dev/null | cut -d= -f2 || true)
    if [ -z "${value}" ]; then
        echo "${default}"
    else
        echo "${value}"
    fi
}

boot_image_compatible() {
    local vm_image="$1"
    local req_image="$2"

    # Exact match always works
    [ "${vm_image}" = "${req_image}" ] && return 0

    # Special images require exact match (cannot substitute)
    # These have unique configurations that can't be approximated
    case "${req_image}" in
        *-fips|*-tuned|*-isolated|*-ai-model-serving|*-fake-next-minor)
            return 1  # Must be exact
            ;;
    esac

    # If VM has a special configuration, it can only run its exact scenarios
    case "${vm_image}" in
        *-fips|*-tuned|*-isolated|*-ai-model-serving|*-fake-next-minor)
            return 1  # Special VM can't run generic scenarios
            ;;
    esac

    # Check if VM image is a superset of required image
    # A VM with "optionals" can run scenarios that need the base image
    # This is like networks: having more capabilities is fine
    if [[ "${vm_image}" == "${req_image}-optionals" ]] || \
       [[ "${vm_image}" == "${req_image}-optional" ]] || \
       [[ "${vm_image}" == "${req_image}-with-optional" ]]; then
        return 0
    fi

    # ai-model-serving includes qemu-guest-agent, so can run isolated scenarios
    if [[ "${req_image}" == *-isolated ]] && [[ "${vm_image}" == *-ai-model-serving ]]; then
        return 0
    fi

    return 1  # Not compatible
}

vm_satisfies_requirements() {
    local vm_name="$1"
    local scenario_reqs="$2"

    local vm_state="${VM_REGISTRY}/${vm_name}/state"
    [ -f "${vm_state}" ] || return 1

    # Get VM capabilities
    local vm_vcpus vm_memory vm_disksize vm_networks vm_fips vm_boot_image
    vm_vcpus=$(get_req_value "${vm_state}" "vcpus" "2")
    vm_memory=$(get_req_value "${vm_state}" "memory" "4096")
    vm_disksize=$(get_req_value "${vm_state}" "disksize" "20")
    vm_networks=$(get_req_value "${vm_state}" "networks" "default")
    vm_fips=$(get_req_value "${vm_state}" "fips" "false")
    vm_boot_image=$(get_req_value "${vm_state}" "boot_image" "")

    # Get scenario requirements
    local req_vcpus req_memory req_disksize req_networks req_fips req_boot_image
    req_vcpus=$(get_req_value "${scenario_reqs}" "min_vcpus" "${DEFAULT_VM_VCPUS}")
    req_memory=$(get_req_value "${scenario_reqs}" "min_memory" "${DEFAULT_VM_MEMORY}")
    req_disksize=$(get_req_value "${scenario_reqs}" "min_disksize" "20")
    req_networks=$(get_req_value "${scenario_reqs}" "networks" "default")
    req_fips=$(get_req_value "${scenario_reqs}" "fips" "false")
    req_boot_image=$(get_req_value "${scenario_reqs}" "boot_image" "")

    # Check minimums
    [ "${vm_vcpus}" -ge "${req_vcpus}" ] || return 1
    [ "${vm_memory}" -ge "${req_memory}" ] || return 1
    [ "${vm_disksize}" -ge "${req_disksize}" ] || return 1

    # Check networks (VM must have all required networks)
    # Empty network requirement means "any network is fine" (network-agnostic)
    if [ -n "${req_networks}" ]; then
        for net in ${req_networks//,/ }; do
            echo ",${vm_networks}," | grep -q ",${net}," || return 1
        done
    fi

    # Check FIPS (if required, VM must have it)
    if [ "${req_fips}" = "true" ] && [ "${vm_fips}" != "true" ]; then
        return 1
    fi

    # Check boot image compatibility
    if ! boot_image_compatible "${vm_boot_image}" "${req_boot_image}"; then
        return 1
    fi

    return 0  # VM is compatible
}

find_compatible_vm() {
    local scenario_reqs="$1"

    for vm_dir in "${VM_REGISTRY}"/*; do
        [ -d "${vm_dir}" ] || continue
        local vm_name
        vm_name=$(basename "${vm_dir}")
        local vm_state="${vm_dir}/state"
        [ -f "${vm_state}" ] || continue

        local vm_status
        vm_status=$(get_req_value "${vm_state}" "status" "unknown")

        if [ "${vm_status}" = "available" ]; then
            if vm_satisfies_requirements "${vm_name}" "${scenario_reqs}"; then
                echo "${vm_name}"
                return 0
            fi
        fi
    done
    return 1
}

can_allocate() {
    local vcpus="$1"
    local memory="$2"

    # Recalculate current usage from active dynamic VMs
    recalculate_resource_usage

    # Check against dynamic available resources (after system + static reservation)
    [ $((current_vcpus + vcpus)) -le ${DYNAMIC_AVAILABLE_VCPUS} ] && \
    [ $((current_memory + memory)) -le ${DYNAMIC_AVAILABLE_MEMORY} ]
}

recalculate_resource_usage() {
    current_vcpus=0
    current_memory=0

    for vm_dir in "${VM_REGISTRY}"/*; do
        [ -d "${vm_dir}" ] || continue
        local vm_state="${vm_dir}/state"
        [ -f "${vm_state}" ] || continue

        local vm_status
        vm_status=$(get_req_value "${vm_state}" "status" "unknown")

        if [ "${vm_status}" = "in_use" ] || [ "${vm_status}" = "available" ]; then
            local vm_vcpus vm_memory
            vm_vcpus=$(get_req_value "${vm_state}" "vcpus" "0")
            vm_memory=$(get_req_value "${vm_state}" "memory" "0")
            current_vcpus=$((current_vcpus + vm_vcpus))
            current_memory=$((current_memory + vm_memory))
        fi
    done
}

register_vm() {
    local vm_name="$1"
    local scenario_reqs="$2"
    local scenario_name="$3"

    local vm_dir="${VM_REGISTRY}/${vm_name}"
    mkdir -p "${vm_dir}"

    # Copy requirements as VM state with status
    cp "${scenario_reqs}" "${vm_dir}/state"

    # Add vcpus/memory/disksize from min_* values and set status
    local vcpus memory disksize networks fips boot_image
    vcpus=$(get_req_value "${scenario_reqs}" "min_vcpus" "${DEFAULT_VM_VCPUS}")
    memory=$(get_req_value "${scenario_reqs}" "min_memory" "${DEFAULT_VM_MEMORY}")
    disksize=$(get_req_value "${scenario_reqs}" "min_disksize" "20")
    networks=$(get_req_value "${scenario_reqs}" "networks" "default")
    fips=$(get_req_value "${scenario_reqs}" "fips" "false")
    boot_image=$(get_req_value "${scenario_reqs}" "boot_image" "")

    cat > "${vm_dir}/state" <<EOF
vcpus=${vcpus}
memory=${memory}
disksize=${disksize}
networks=${networks}
fips=${fips}
boot_image=${boot_image}
status=in_use
current_scenario=${scenario_name}
EOF

    # Initialize scenario history
    echo "$(date -Iseconds) CREATED for ${scenario_name}" >> "${vm_dir}/scenario_history.log"

    log "Registered VM ${vm_name} for scenario ${scenario_name}"
}

assign_vm_to_scenario() {
    local vm_name="$1"
    local scenario_name="$2"

    local vm_dir="${VM_REGISTRY}/${vm_name}"
    local vm_state="${vm_dir}/state"

    # Update status and current scenario
    sed -i "s/^status=.*/status=in_use/" "${vm_state}"
    sed -i "s/^current_scenario=.*/current_scenario=${scenario_name}/" "${vm_state}"

    # Record in scenario history
    echo "$(date -Iseconds) START ${scenario_name}" >> "${vm_dir}/scenario_history.log"

    # Mark the scenario as running in the queue (prevents duplicate dispatch)
    local queue_file="${SCENARIO_QUEUE}/${scenario_name}"
    if [ -f "${queue_file}" ]; then
        sed -i "s/^status=.*/status=running/" "${queue_file}"
        echo "started_at=$(date -Iseconds)" >> "${queue_file}"
    fi

    # Create assignment file for scenario
    mkdir -p "${SCENARIO_STATUS}/${scenario_name}"
    echo "${vm_name}" > "${SCENARIO_STATUS}/${scenario_name}/vm_assignment"
    echo "true" > "${SCENARIO_STATUS}/${scenario_name}/vm_reused"

    log "Assigned VM ${vm_name} to scenario ${scenario_name} (reuse)"
}

release_vm() {
    local vm_name="$1"
    local scenario_name="$2"
    local result="$3"

    local vm_dir="${VM_REGISTRY}/${vm_name}"
    local vm_state="${vm_dir}/state"

    # Record in scenario history
    echo "$(date -Iseconds) END ${scenario_name} ${result}" >> "${vm_dir}/scenario_history.log"

    # Update status to available
    sed -i "s/^status=.*/status=available/" "${vm_state}"
    sed -i "s/^current_scenario=.*/current_scenario=/" "${vm_state}"

    log "Released VM ${vm_name} from scenario ${scenario_name} (${result})"
}

destroy_vm() {
    local vm_name="$1"
    local reason="${2:-no compatible scenarios}"

    local vm_dir="${VM_REGISTRY}/${vm_name}"

    # Record destruction in history
    echo "$(date -Iseconds) DESTROYED (${reason})" >> "${vm_dir}/scenario_history.log"

    # Mark as destroyed
    sed -i "s/^status=.*/status=destroyed/" "${vm_dir}/state" 2>/dev/null || \
        echo "status=destroyed" >> "${vm_dir}/state"

    log "Destroying libvirt VM ${vm_name}: ${reason}"

    # Actually destroy the libvirt VM and clean up resources
    if sudo virsh dumpxml "${vm_name}" &>/dev/null; then
        if ! sudo virsh dominfo "${vm_name}" 2>/dev/null | grep '^State' | grep -q 'shut off'; then
            sudo virsh destroy --graceful "${vm_name}" 2>/dev/null || true
        fi
        if ! sudo virsh dominfo "${vm_name}" 2>/dev/null | grep '^State' | grep -q 'shut off'; then
            sudo virsh destroy "${vm_name}" 2>/dev/null || true
        fi
        sudo virsh undefine --nvram "${vm_name}" 2>/dev/null || true
    fi

    # Clean up storage pool
    local vm_pool_name="${VM_POOL_BASENAME}-${vm_name}"
    if sudo virsh pool-info "${vm_pool_name}" &>/dev/null; then
        sudo virsh pool-destroy "${vm_pool_name}" 2>/dev/null || true
        sudo virsh pool-undefine "${vm_pool_name}" 2>/dev/null || true
    fi

    # Remove pool directory
    rm -rf "${VM_DISK_BASEDIR:?}/${vm_pool_name}" 2>/dev/null || true

    log "Destroyed VM ${vm_name}"
}

# Calculate reusability score for a scenario
# Lower score = less reusable = should run first
# Higher score = more reusable = should stay queued longer for reuse
get_reusability_score() {
    local boot_image="$1"
    local networks="$2"

    # Score 0: Completely non-reusable (special boot_image that requires exact match)
    case "${boot_image}" in
        *-fips|*-tuned|*-isolated|*-ai-model-serving|*-fake-next-minor)
            echo "0"
            return
            ;;
    esac

    # Score 1: Non-reusable due to special network (not default or empty)
    if [ -n "${networks}" ] && [ "${networks}" != "default" ]; then
        echo "1"
        return
    fi

    # Score 2: Reusable - optionals can be reused by source scenarios
    case "${boot_image}" in
        *-optionals|*-optional|*-with-optional)
            echo "2"
            return
            ;;
    esac

    # Score 3: Flexible - source/base image with default network
    if [ "${networks}" = "default" ]; then
        echo "3"
        return
    fi

    # Score 4: Most flexible - network-agnostic (empty networks)
    # Can run on ANY VM, stays queued longest for maximum reuse
    echo "4"
}

# Sort scenarios by requirements to maximize VM reuse
sort_scenarios_for_reuse() {
    local -a scenarios=("$@")

    # Sort strategy for maximum VM reuse:
    # 1. Reusability score (ascending): restrictive scenarios first
    #    - They can't be reused anyway, so get them out of the way
    #    - Their VMs are destroyed when done
    # 2. Within same reusability: slow tests first (ascending 0=slow, 1=normal)
    #    - Long-running tests start early to avoid becoming critical path
    # 3. Then boot_image descending (optionals before source)
    # 4. Then vcpus descending (larger VMs first)
    #
    # This ensures:
    # - Restrictive scenarios (special networks/boot_images) start first
    # - Slow tests start early within their reusability tier
    # - Flexible scenarios stay queued longer, maximizing reuse opportunities

    for scenario_script in "${scenarios[@]}"; do
        local scenario_name
        scenario_name=$(basename "${scenario_script}" .sh)
        local req_file="${SCENARIO_STATUS}/${scenario_name}/requirements"

        if [ -f "${req_file}" ]; then
            local boot_image networks vcpus reuse_score slow_flag slow_sort
            boot_image=$(get_req_value "${req_file}" "boot_image" "default")
            networks=$(get_req_value "${req_file}" "networks" "default")
            vcpus=$(get_req_value "${req_file}" "min_vcpus" "${DEFAULT_VM_VCPUS}")
            slow_flag=$(get_req_value "${req_file}" "slow" "false")
            reuse_score=$(get_reusability_score "${boot_image}" "${networks}")
            # slow=true -> sort value 0 (first), slow=false -> sort value 1 (after)
            slow_sort=1
            if [ "${slow_flag}" = "true" ]; then
                slow_sort=0
            fi
            # Sort key: reuse_score (asc), slow_sort (asc), boot_image (desc), vcpus (desc)
            # Use inverse vcpus (100-vcpus) so ascending sort gives descending vcpus
            printf "%d\t%d\t%s\t%02d\t%s\n" "${reuse_score}" "${slow_sort}" "${boot_image}" "$((100 - vcpus))" "${scenario_script}"
        else
            # Unknown requirements - treat as flexible, put last
            printf "9\t1\taaa\t99\t%s\n" "${scenario_script}"
        fi
    done | sort -t$'\t' -k1,1n -k2,2n -k3,3r -k4,4n | cut -f5
}

# Global sequence counter for queue ordering
QUEUE_SEQUENCE=0

queue_scenario() {
    local scenario_script="$1"
    local scenario_name="$2"
    local req_file="$3"

    QUEUE_SEQUENCE=$((QUEUE_SEQUENCE + 1))
    local queue_file="${SCENARIO_QUEUE}/${scenario_name}"
    cat > "${queue_file}" <<EOF
script=${scenario_script}
requirements=${req_file}
status=queued
queued_at=$(date -Iseconds)
sequence=${QUEUE_SEQUENCE}
EOF

    log "Queued scenario ${scenario_name} (seq=${QUEUE_SEQUENCE})"
}

get_queued_scenarios() {
    # Return queued scenarios sorted by their sequence number
    # to preserve the order they were added (which is sorted by requirements)
    for queue_file in "${SCENARIO_QUEUE}"/*; do
        [ -f "${queue_file}" ] || continue
        local status
        status=$(get_req_value "${queue_file}" "status" "")
        if [ "${status}" = "queued" ]; then
            local seq
            seq=$(get_req_value "${queue_file}" "sequence" "999")
            printf "%03d\t%s\n" "${seq}" "$(basename "${queue_file}")"
        fi
    done | sort -n | cut -f2
}

mark_scenario_running() {
    local scenario_name="$1"
    local queue_file="${SCENARIO_QUEUE}/${scenario_name}"

    sed -i "s/^status=.*/status=running/" "${queue_file}"
    echo "started_at=$(date -Iseconds)" >> "${queue_file}"
}

mark_scenario_completed() {
    local scenario_name="$1"
    local result="$2"
    local queue_file="${SCENARIO_QUEUE}/${scenario_name}"

    sed -i "s/^status=.*/status=completed/" "${queue_file}"
    echo "completed_at=$(date -Iseconds)" >> "${queue_file}"
    echo "result=${result}" >> "${queue_file}"
}

has_pending_scenarios() {
    for queue_file in "${SCENARIO_QUEUE}"/*; do
        [ -f "${queue_file}" ] || continue
        local status
        status=$(get_req_value "${queue_file}" "status" "")
        if [ "${status}" = "queued" ]; then
            return 0
        fi
    done
    return 1
}

run_scenario_on_vm() {
    local scenario_script="$1"
    local scenario_name="$2"
    local vm_name="$3"
    local is_new_vm="$4"

    log "Starting scenario ${scenario_name} on VM ${vm_name} (new_vm=${is_new_vm})"

    mark_scenario_running "${scenario_name}"

    # Set up environment for scheduler-aware scenario execution
    export SCHEDULER_ENABLED=true
    export SCHEDULER_SCENARIO_NAME="${scenario_name}"
    export SCHEDULER_STATE_DIR="${SCHEDULER_STATE_DIR}"

    # Always pass the VM name - scheduler controls naming for dynamic scenarios
    export SCHEDULER_VM_NAME="${vm_name}"
    if [ "${is_new_vm}" = "false" ]; then
        export SCHEDULER_IS_NEW_VM="false"
    else
        export SCHEDULER_IS_NEW_VM="true"
    fi

    local result="SUCCESS"
    local exit_code=0

    # Set up log directory for scenario execution
    local scenario_log_dir="${SCENARIO_INFO_DIR}/${scenario_name}"
    mkdir -p "${scenario_log_dir}"

    # Get scenario-specific timeouts if specified, otherwise use defaults
    local req_file="${SCENARIO_STATUS}/${scenario_name}/requirements"
    local test_timeout="${VM_TEST_TIMEOUT}"
    if [ -f "${req_file}" ]; then
        local custom_test_timeout
        custom_test_timeout=$(get_req_value "${req_file}" "test_timeout" "")
        if [ -n "${custom_test_timeout}" ]; then
            test_timeout="${custom_test_timeout}"
        fi

        # Export greenboot timeout for scenario.sh to use
        local custom_greenboot_timeout
        custom_greenboot_timeout=$(get_req_value "${req_file}" "greenboot_timeout" "")
        if [ -n "${custom_greenboot_timeout}" ]; then
            export VM_GREENBOOT_TIMEOUT="${custom_greenboot_timeout}"
            log "Using custom greenboot timeout: ${custom_greenboot_timeout}s"
        fi
    fi

    local boot_log="${scenario_log_dir}/boot.log"
    local run_log="${scenario_log_dir}/run.log"
    local vm_dir="${VM_REGISTRY}/${vm_name}"
    mkdir -p "${vm_dir}"
    ln -sf "${boot_log}" "${vm_dir}/creation_log"

    if [ "${is_new_vm}" = "true" ]; then
        # New VM: separate create and run phases
        # Create phase has its own retry logic with timeouts
        log "Creating VM ${vm_name} - logging to ${boot_log}"

        local create_exit=0
        bash -x "${SCRIPTDIR}/scenario.sh" create "${scenario_script}" &> "${boot_log}" || create_exit=$?

        if [ ${create_exit} -ne 0 ]; then
            result="FAILED"
            exit_code=1
            log "VM creation failed for ${scenario_name} (exit ${create_exit}) - see ${boot_log}"
        fi

        # Run tests only if creation succeeded
        if [ "${exit_code}" -eq 0 ]; then
            log "Running tests for ${scenario_name} (timeout: ${test_timeout}s) - logging to ${run_log}"

            # Track test execution time
            local test_start_time
            test_start_time=$(date +%s)

            local run_exit=0
            timeout --signal=TERM --kill-after=60 "${test_timeout}" \
                bash -x "${SCRIPTDIR}/scenario.sh" run "${scenario_script}" &> "${run_log}" || run_exit=$?

            # Record test execution time
            local test_end_time test_duration
            test_end_time=$(date +%s)
            test_duration=$((test_end_time - test_start_time))
            echo "${test_duration}" > "${scenario_log_dir}/test_time"

            if [ ${run_exit} -ne 0 ]; then
                result="FAILED"
                exit_code=1
                if [ ${run_exit} -eq 124 ]; then
                    log "Tests TIMED OUT for ${scenario_name} after ${test_timeout}s - see ${run_log}"
                else
                    log "Tests failed for ${scenario_name} (exit ${run_exit}) - see ${run_log}"
                fi
            fi
        fi
    else
        # Reused VM: use create-and-run for faster turnaround
        # The create phase is very quick (just IP lookup and SSH check)
        # Combined output goes to run.log, with setup info in boot.log
        log "Setting up reused VM ${vm_name} and running tests (timeout: ${test_timeout}s)"

        # Log reuse info to boot.log for consistency
        echo "VM ${vm_name} reused for scenario ${scenario_name}" > "${boot_log}"
        date >> "${boot_log}"

        # Track test execution time (for reused VM, this includes quick setup)
        local test_start_time
        test_start_time=$(date +%s)

        local combined_exit=0
        timeout --signal=TERM --kill-after=60 "${test_timeout}" \
            bash -x "${SCRIPTDIR}/scenario.sh" create-and-run "${scenario_script}" &> "${run_log}" || combined_exit=$?

        # Record test execution time
        local test_end_time test_duration
        test_end_time=$(date +%s)
        test_duration=$((test_end_time - test_start_time))
        echo "${test_duration}" > "${scenario_log_dir}/test_time"

        if [ ${combined_exit} -ne 0 ]; then
            result="FAILED"
            exit_code=1
            if [ ${combined_exit} -eq 124 ]; then
                log "Tests TIMED OUT for ${scenario_name} after ${test_timeout}s - see ${run_log}"
            else
                log "Tests failed for ${scenario_name} (exit ${combined_exit}) - see ${run_log}"
            fi
        fi
    fi

    mark_scenario_completed "${scenario_name}" "${result}"

    # Handle VM after scenario completion
    acquire_lock "vm_dispatch"

    release_vm "${vm_name}" "${scenario_name}" "${result}"

    # If tests failed, destroy the VM immediately as there are no guarantees to its state
    if [ "${result}" = "FAILED" ]; then
        destroy_vm "${vm_name}" "test failure"
        release_lock "vm_dispatch"
        return ${exit_code}
    fi

    # Check if scenario requested VM destruction after completion (e.g., system-level changes)
    local destroy_after
    destroy_after=$(get_req_value "${req_file}" "destroy_after" "false")
    if [ "${destroy_after}" = "true" ]; then
        destroy_vm "${vm_name}" "destroy_after flag set"
        release_lock "vm_dispatch"
        return ${exit_code}
    fi

    # Check if any queued scenario can use this VM
    local next_scenario=""
    for queued in $(get_queued_scenarios); do
        local queue_file="${SCENARIO_QUEUE}/${queued}"
        local req_file
        req_file=$(get_req_value "${queue_file}" "requirements" "")

        if [ -n "${req_file}" ] && [ -f "${req_file}" ]; then
            if vm_satisfies_requirements "${vm_name}" "${req_file}"; then
                next_scenario="${queued}"
                break
            fi
        fi
    done

    if [ -n "${next_scenario}" ]; then
        # Reuse VM for next compatible scenario
        local queue_file="${SCENARIO_QUEUE}/${next_scenario}"
        local next_script
        next_script=$(get_req_value "${queue_file}" "script" "")

        assign_vm_to_scenario "${vm_name}" "${next_scenario}"

        release_lock "vm_dispatch"

        # Run the next scenario (recursive call)
        # Capture exit code to propagate failures from reused scenarios
        local reuse_exit=0
        run_scenario_on_vm "${next_script}" "${next_scenario}" "${vm_name}" "false" || reuse_exit=$?
        if [ ${reuse_exit} -ne 0 ]; then
            exit_code=${reuse_exit}
        fi
    else
        # No compatible scenario waiting - destroy VM
        destroy_vm "${vm_name}"
        release_lock "vm_dispatch"
    fi

    return ${exit_code}
}

create_vm_for_scenario() {
    local scenario_script="$1"
    local scenario_name="$2"
    local req_file="$3"

    local vm_name
    vm_name=$(generate_dynamic_vm_name)

    register_vm "${vm_name}" "${req_file}" "${scenario_name}"

    # Create assignment file for scenario
    mkdir -p "${SCENARIO_STATUS}/${scenario_name}"
    echo "${vm_name}" > "${SCENARIO_STATUS}/${scenario_name}/vm_assignment"
    echo "false" > "${SCENARIO_STATUS}/${scenario_name}/vm_reused"

    echo "${vm_name}"
}

dispatch_dynamic_scenarios() {
    local -a pids=()
    local -a pid_scenarios=()
    local overall_result=0

    while has_pending_scenarios || [ ${#pids[@]} -gt 0 ]; do
        # Try to dispatch queued scenarios
        acquire_lock "vm_dispatch"

        for scenario_name in $(get_queued_scenarios); do
            local queue_file="${SCENARIO_QUEUE}/${scenario_name}"
            local scenario_script req_file
            scenario_script=$(get_req_value "${queue_file}" "script" "")
            req_file=$(get_req_value "${queue_file}" "requirements" "")

            if [ -z "${scenario_script}" ] || [ -z "${req_file}" ]; then
                continue
            fi

            local min_vcpus min_memory
            min_vcpus=$(get_req_value "${req_file}" "min_vcpus" "${DEFAULT_VM_VCPUS}")
            min_memory=$(get_req_value "${req_file}" "min_memory" "${DEFAULT_VM_MEMORY}")

            # Try to find compatible free VM
            local vm_name=""
            if vm_name=$(find_compatible_vm "${req_file}"); then
                # Reuse existing VM
                assign_vm_to_scenario "${vm_name}" "${scenario_name}"
                log "Reusing VM ${vm_name} for ${scenario_name}"

                # Start scenario in background
                run_scenario_on_vm "${scenario_script}" "${scenario_name}" "${vm_name}" "false" &
                local pid=$!
                pids+=("${pid}")
                pid_scenarios+=("${scenario_name}")

            elif can_allocate "${min_vcpus}" "${min_memory}"; then
                # Create new VM
                vm_name=$(create_vm_for_scenario "${scenario_script}" "${scenario_name}" "${req_file}")
                log "Creating new VM ${vm_name} for ${scenario_name}"

                # Start scenario in background
                run_scenario_on_vm "${scenario_script}" "${scenario_name}" "${vm_name}" "true" &
                local pid=$!
                pids+=("${pid}")
                pid_scenarios+=("${scenario_name}")

            else
                log "Resources exhausted, keeping ${scenario_name} queued (need: vcpus=${min_vcpus}, mem=${min_memory})"
            fi
        done

        release_lock "vm_dispatch"

        # If we have running jobs, wait for any one to finish
        if [ ${#pids[@]} -gt 0 ]; then
            # Wait for any background job to complete (bash 4.3+)
            # Returns immediately when any child exits
            local finished_pid=""
            wait -n -p finished_pid 2>/dev/null || true

            # Find which scenario finished using the captured PID
            local -a new_pids=()
            local -a new_pid_scenarios=()
            for i in "${!pids[@]}"; do
                local pid="${pids[${i}]}"
                local scenario="${pid_scenarios[${i}]}"
                if [ "${pid}" = "${finished_pid}" ]; then
                    # This is the process that finished - collect exit status
                    wait "${pid}" || overall_result=1
                    log "Scenario ${scenario} finished (pid ${pid})"
                else
                    new_pids+=("${pid}")
                    new_pid_scenarios+=("${scenario}")
                fi
            done
            pids=("${new_pids[@]+"${new_pids[@]}"}")
            pid_scenarios=("${new_pid_scenarios[@]+"${new_pid_scenarios[@]}"}")
        fi
    done

    return ${overall_result}
}

orchestrate() {
    local scenario_dir="$1"

    if [ ! -d "${scenario_dir}" ]; then
        echo "ERROR: Scenario directory not found: ${scenario_dir}" >&2
        exit 1
    fi

    init_scheduler
    log "=== PHASE 0: Preparing ==="

    log "=== PHASE 1: Classifying scenarios ==="
    local -a dynamic_scenarios=()
    local -a static_scenarios=()

    for scenario_script in "${scenario_dir}"/*.sh; do
        [ -f "${scenario_script}" ] || continue

        if scenario_is_dynamic "${scenario_script}"; then
            dynamic_scenarios+=("${scenario_script}")
            log "  Dynamic: $(basename "${scenario_script}")"
        else
            static_scenarios+=("${scenario_script}")
            log "  Static: $(basename "${scenario_script}")"
        fi
    done

    log "Found ${#dynamic_scenarios[@]} dynamic scenarios and ${#static_scenarios[@]} static scenarios"

    log "=== PHASE 2: Resource Planning and Validation ==="

    if [ ${#static_scenarios[@]} -gt 0 ]; then
        log "Calculating static scenario requirements..."
        calculate_static_requirements "${static_scenarios[@]}"
    else
        log "No static scenarios"
        STATIC_TOTAL_VCPUS=0
        STATIC_TOTAL_MEMORY=0
    fi

    if [ ${#dynamic_scenarios[@]} -gt 0 ]; then
        log "Calculating dynamic scenario requirements..."
        calculate_max_dynamic_requirements "${dynamic_scenarios[@]}"
    else
        log "No dynamic scenarios"
        MAX_DYNAMIC_VCPUS=0
        MAX_DYNAMIC_MEMORY=0
    fi

    DYNAMIC_AVAILABLE_VCPUS=$((HOST_AVAILABLE_VCPUS - STATIC_TOTAL_VCPUS))
    DYNAMIC_AVAILABLE_MEMORY=$((HOST_AVAILABLE_MEMORY - STATIC_TOTAL_MEMORY))

    log ""
    log "=== Resource Summary ==="
    log "  Host total:        vcpus=${HOST_TOTAL_VCPUS}, memory=${HOST_TOTAL_MEMORY}MB"
    log "  System reserved:   vcpus=${SYSTEM_RESERVED_VCPUS}, memory=${SYSTEM_RESERVED_MEMORY}MB"
    log "  Available for VMs: vcpus=${HOST_AVAILABLE_VCPUS}, memory=${HOST_AVAILABLE_MEMORY}MB"
    log "  Static requires:   vcpus=${STATIC_TOTAL_VCPUS}, memory=${STATIC_TOTAL_MEMORY}MB"
    log "  Dynamic available: vcpus=${DYNAMIC_AVAILABLE_VCPUS}, memory=${DYNAMIC_AVAILABLE_MEMORY}MB"
    log "  Max dynamic needs: vcpus=${MAX_DYNAMIC_VCPUS}, memory=${MAX_DYNAMIC_MEMORY}MB"
    log ""

    if [ ${STATIC_TOTAL_VCPUS} -gt ${HOST_AVAILABLE_VCPUS} ]; then
        log "ERROR: Static scenarios require more vCPUs than available"
        log "  Required: ${STATIC_TOTAL_VCPUS}, Available: ${HOST_AVAILABLE_VCPUS}"
        return 1
    fi
    if [ ${STATIC_TOTAL_MEMORY} -gt ${HOST_AVAILABLE_MEMORY} ]; then
        log "ERROR: Static scenarios require more memory than available"
        log "  Required: ${STATIC_TOTAL_MEMORY}MB, Available: ${HOST_AVAILABLE_MEMORY}MB"
        return 1
    fi

    if [ ${#dynamic_scenarios[@]} -gt 0 ]; then
        if [ ${MAX_DYNAMIC_VCPUS} -gt ${DYNAMIC_AVAILABLE_VCPUS} ]; then
            log "ERROR: Largest dynamic scenario requires more vCPUs than available after static allocation"
            log "  Required: ${MAX_DYNAMIC_VCPUS}, Available: ${DYNAMIC_AVAILABLE_VCPUS}"
            return 1
        fi
        if [ ${MAX_DYNAMIC_MEMORY} -gt ${DYNAMIC_AVAILABLE_MEMORY} ]; then
            log "ERROR: Largest dynamic scenario requires more memory than available after static allocation"
            log "  Required: ${MAX_DYNAMIC_MEMORY}MB, Available: ${DYNAMIC_AVAILABLE_MEMORY}MB"
            return 1
        fi
    fi

    log "Resource validation PASSED - all scenarios can run"
    log ""

    # Persist resource allocation for status command
    cat > "${SCHEDULER_STATE_DIR}/resource_allocation" <<EOF
STATIC_TOTAL_VCPUS=${STATIC_TOTAL_VCPUS}
STATIC_TOTAL_MEMORY=${STATIC_TOTAL_MEMORY}
DYNAMIC_AVAILABLE_VCPUS=${DYNAMIC_AVAILABLE_VCPUS}
DYNAMIC_AVAILABLE_MEMORY=${DYNAMIC_AVAILABLE_MEMORY}
MAX_DYNAMIC_VCPUS=${MAX_DYNAMIC_VCPUS}
MAX_DYNAMIC_MEMORY=${MAX_DYNAMIC_MEMORY}
STATIC_SCENARIO_COUNT=${#static_scenarios[@]}
DYNAMIC_SCENARIO_COUNT=${#dynamic_scenarios[@]}
EOF

    if [ ${#static_scenarios[@]} -gt 0 ]; then
        log "=== PHASE 3: Creating static VMs ==="
        if ! create_static_vms "${static_scenarios[@]}"; then
            log "ERROR: Failed to create static VMs, aborting"
            return 1
        fi
    fi

    # Sort scenarios by requirements to maximize VM reuse
    while IFS= read -r scenario_script; do
        [ -n "${scenario_script}" ] || continue
        local scenario_name
        scenario_name=$(basename "${scenario_script}" .sh)
        local req_file="${SCENARIO_STATUS}/${scenario_name}/requirements"
        queue_scenario "${scenario_script}" "${scenario_name}" "${req_file}"
    done < <(sort_scenarios_for_reuse "${dynamic_scenarios[@]}")

    log "=== PHASE 4: Running tests ==="

    local overall_result=0
    local static_pid=""
    local dynamic_result=0
    local static_result=0

    if [ ${#static_scenarios[@]} -gt 0 ]; then
        log "Starting static scenario tests in background"
        run_static_tests "${static_scenarios[@]}" &
        static_pid=$!
        log "Static tests running in background (pid ${static_pid})"
    fi

    if [ ${#dynamic_scenarios[@]} -gt 0 ]; then
        log "Dispatching dynamic scenarios"
        if ! dispatch_dynamic_scenarios; then
            dynamic_result=1
        fi
    fi

    if [ -n "${static_pid}" ]; then
        log "Waiting for static tests to complete (pid ${static_pid})"
        if ! wait "${static_pid}"; then
            static_result=1
        fi
        log "Static tests completed"
    fi

    if [ ${dynamic_result} -ne 0 ] || [ ${static_result} -ne 0 ]; then
        overall_result=1
    fi

    log "Orchestration complete (dynamic=${dynamic_result}, static=${static_result})"

    return ${overall_result}
}

show_status() {
    echo "============================================================================"
    echo "                         SCHEDULER STATUS"
    echo "============================================================================"
    echo ""
    echo "State directory: ${SCHEDULER_STATE_DIR}"
    echo ""

    # Load persisted resource allocation if available
    local resource_file="${SCHEDULER_STATE_DIR}/resource_allocation"
    local static_scenario_count=0
    local dynamic_scenario_count=0
    if [ -f "${resource_file}" ]; then
        # shellcheck source=/dev/null
        source "${resource_file}"
        static_scenario_count="${STATIC_SCENARIO_COUNT:-0}"
        dynamic_scenario_count="${DYNAMIC_SCENARIO_COUNT:-0}"
    fi

    # --- Scenario Results ---
    local total_scenarios=0
    local passed_scenarios=0
    local failed_scenarios=0

    for queue_file in "${SCENARIO_QUEUE}"/*; do
        [ -f "${queue_file}" ] || continue
        ((total_scenarios++)) || true

        local result
        result=$(get_req_value "${queue_file}" "result" "")
        if [ "${result}" = "SUCCESS" ]; then
            ((passed_scenarios++)) || true
        elif [ "${result}" = "FAILED" ]; then
            ((failed_scenarios++)) || true
        fi
    done

    # --- VM Metrics ---
    local vms_created=0
    local vm_reuses=0
    local vm_creation_retries=0
    local max_runs_per_vm=0

    for scenario_dir in "${SCENARIO_STATUS}"/*; do
        [ -d "${scenario_dir}" ] || continue
        local reused_file="${scenario_dir}/vm_reused"
        if [ -f "${reused_file}" ]; then
            local reused
            reused=$(cat "${reused_file}")
            if [ "${reused}" = "true" ]; then
                ((vm_reuses++)) || true
            else
                ((vms_created++)) || true
            fi
        fi
    done

    # Count VM creation retries from scenario info directories
    for scenario_dir in "${SCENARIO_INFO_DIR}"/*; do
        [ -d "${scenario_dir}" ] || continue
        local retry_file="${scenario_dir}/vm_creation_retries"
        if [ -f "${retry_file}" ]; then
            local retries
            retries=$(cat "${retry_file}")
            ((vm_creation_retries += retries)) || true
        fi
    done

    for vm_dir in "${VM_REGISTRY}"/*; do
        [ -d "${vm_dir}" ] || continue
        local history_file="${vm_dir}/scenario_history.log"
        if [ -f "${history_file}" ]; then
            local runs
            runs=$(grep -c "^.* START " "${history_file}" 2>/dev/null || echo 0)
            if [ "${runs}" -gt "${max_runs_per_vm}" ]; then
                max_runs_per_vm="${runs}"
            fi
        fi
    done

    local reuse_rate=0
    if [ ${total_scenarios} -gt 0 ]; then
        reuse_rate=$((vm_reuses * 100 / total_scenarios))
    fi

    echo "=== Scenario Results ==="
    echo "  Static scenarios:         ${static_scenario_count}"
    echo "  Dynamic scenarios:        ${dynamic_scenario_count} (passed: ${passed_scenarios}, failed: ${failed_scenarios})"
    echo ""

    echo "=== Dynamic VM Efficiency ==="
    echo "  VMs created:              ${vms_created}"
    echo "  VM creation retries:      ${vm_creation_retries}"
    echo "  VM reuses:                ${vm_reuses}"
    echo "  Reuse rate:               ${reuse_rate}%"
    echo "  Max scenarios per VM:     ${max_runs_per_vm}"
    echo ""

    echo "=== Resource Configuration ==="
    echo "  Host total:               vcpus=${HOST_TOTAL_VCPUS}, memory=${HOST_TOTAL_MEMORY}MB"
    echo "  System reserved:          vcpus=${SYSTEM_RESERVED_VCPUS}, memory=${SYSTEM_RESERVED_MEMORY}MB"
    echo "  Available for VMs:        vcpus=${HOST_AVAILABLE_VCPUS}, memory=${HOST_AVAILABLE_MEMORY}MB"
    echo ""

    echo "=== Resource Allocation ==="
    echo "  Static requires:          vcpus=${STATIC_TOTAL_VCPUS}, memory=${STATIC_TOTAL_MEMORY}MB"
    echo "  Dynamic available:        vcpus=${DYNAMIC_AVAILABLE_VCPUS}, memory=${DYNAMIC_AVAILABLE_MEMORY}MB"
    echo "  Max dynamic needs:        vcpus=${MAX_DYNAMIC_VCPUS}, memory=${MAX_DYNAMIC_MEMORY}MB"
    echo ""

    echo "=== Current Dynamic Usage ==="
    recalculate_resource_usage
    echo "  Dynamic VMs using:        vcpus=${current_vcpus}, memory=${current_memory}MB"
    echo ""

    echo "=== VMs ==="
    for vm_dir in "${VM_REGISTRY}"/*; do
        [ -d "${vm_dir}" ] || continue
        local vm_name
        vm_name=$(basename "${vm_dir}")
        local vm_state="${vm_dir}/state"
        if [ -f "${vm_state}" ]; then
            echo "${vm_name}:"
            sed 's/^/  /' "${vm_state}"
        fi
    done
    echo ""

    echo "=== Test Durations ==="
    echo "  (Tests longer than avg boot time should be 'slow', shorter should be 'fast')"
    echo ""
    # Collect test durations and check for mislabeling
    local mislabeled_count=0
    {
        for scenario_dir in "${SCENARIO_INFO_DIR}"/*; do
            [ -d "${scenario_dir}" ] || continue
            local scenario_name
            scenario_name=$(basename "${scenario_dir}")
            local test_time_file="${scenario_dir}/test_time"

            if [ -f "${test_time_file}" ]; then
                local test_time
                test_time=$(cat "${test_time_file}")
                local test_mins=$((test_time / 60))
                local test_secs=$((test_time % 60))
                local time_str
                time_str=$(printf '%2d:%02d' "${test_mins}" "${test_secs}")

                # Get slow/fast flag from requirements
                local req_file="${SCENARIO_STATUS}/${scenario_name}/requirements"
                local slow_flag="false"
                local fast_flag="false"
                local current_label="-"
                if [ -f "${req_file}" ]; then
                    slow_flag=$(get_req_value "${req_file}" "slow" "false")
                    fast_flag=$(get_req_value "${req_file}" "fast" "false")
                    if [ "${slow_flag}" = "true" ]; then
                        current_label="slow"
                    elif [ "${fast_flag}" = "true" ]; then
                        current_label="fast"
                    fi
                fi

                # Determine if mislabeled (compare against avg boot time)
                local status=""
                if [ ${avg_boot_time} -gt 0 ]; then
                    if [ ${test_time} -gt ${avg_boot_time} ]; then
                        # Test is slower than boot - should be 'slow'
                        if [ "${fast_flag}" = "true" ]; then
                            status="<-- MISLABELED (should be slow)"
                            mislabeled_count=$((mislabeled_count + 1))
                        fi
                    else
                        # Test is faster than boot - should be 'fast'
                        if [ "${slow_flag}" = "true" ]; then
                            status="<-- MISLABELED (should be fast)"
                            mislabeled_count=$((mislabeled_count + 1))
                        fi
                    fi
                fi

                # Output: time, scenario, label, status (tab-separated for sorting)
                printf "%d\t%s\t%-40s\t%-6s\t%s\n" "${test_time}" "${time_str}" "${scenario_name}" "${current_label}" "${status}"
            fi
        done
    } | sort -t$'\t' -k1,1rn | cut -f2- | while IFS=$'\t' read -r time_str scenario label status; do
        printf "  %s  %-40s  [%s] %s\n" "${time_str}" "${scenario}" "${label}" "${status}"
    done
    echo ""
    if [ ${mislabeled_count} -gt 0 ]; then
        echo "  WARNING: ${mislabeled_count} test(s) appear to be mislabeled based on avg boot time (${avg_boot_time_str})"
    fi
    echo ""

    echo "=== Scenarios ==="
    for queue_file in "${SCENARIO_QUEUE}"/*; do
        [ -f "${queue_file}" ] || continue
        local scenario_name
        scenario_name=$(basename "${queue_file}")
        echo "${scenario_name}:"
        sed 's/^/  /' "${queue_file}"
    done
    echo ""
    echo "============================================================================"
}

usage() {
    cat <<EOF
Usage: $(basename "$0") <command> [args]

Commands:
    orchestrate <scenario_dir>  Schedule and run all scenarios in directory
    status                      Show current scheduler state

Environment Variables:
    HOST_TOTAL_VCPUS       Total host vCPUs (default: system detected)
    HOST_TOTAL_MEMORY      Total host memory in MB (default: system detected)
    SYSTEM_RESERVED_VCPUS  vCPUs reserved for host OS (default: 2)
    SYSTEM_RESERVED_MEMORY Memory reserved for host OS in MB (default: 4096)
    SCHEDULER_ENABLED      Enable scheduler mode (default: false)

Execution Flow:
    1. Classify scenarios (dynamic vs static)
    2. Calculate resource requirements for ALL scenarios
       - Parse static scenarios for VM requirements
       - Parse dynamic scenarios for VM requirements
    3. Validate resources BEFORE creating any VMs:
       - System reserved + static + max_dynamic <= host_total
    4. Create static VMs (in parallel)
    5. Run static tests and dynamic scenarios concurrently
EOF
}

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

command="$1"
shift

case "${command}" in
    orchestrate)
        if [ $# -lt 1 ]; then
            echo "ERROR: orchestrate requires scenario directory argument" >&2
            usage
            exit 1
        fi
        orchestrate "$1"
        ;;
    status)
        show_status
        ;;
    *)
        echo "ERROR: Unknown command: ${command}" >&2
        usage
        exit 1
        ;;
esac
