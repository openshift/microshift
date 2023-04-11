#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
OUTPUT_DIR="${ARTIFACT_DIR:-${SCRIPT_DIR}/../_output}/microshift-e2e-$(date +'%Y%m%d-%H%M%S')/"

usage() {
    echo "Usage: $(basename "${0}") {list|run} [filter]"
    echo ""
    echo "   list      Lists tests"
    echo "   run       Runs tests"
    echo "   filter    Simple string to match against test files, e.g. 'reboot', 'boot', 'smoke'"
    echo ""
    echo " Script expects two environmental variables:"
    echo "   - USHIFT_IP"
    echo "   - USHIFT_USER"
    echo ""
    echo " Script assumes following:"
    echo "   - Passwordless SSH to \$USHIFT_USER@\$USHIFT_IP is configured"
    echo "   - Both hosts already exchanged their public keys:"
    echo "     - Test runner has MicroShift's sshd key in ~/.ssh/known_keys"
    echo "     - Remote \$USHIFT_USER has test runner's key in ~/.ssh/authorized_keys"
    echo "   - Passwordless sudo for \$USHIFT_USER"
    echo ""
    echo " Script aims to be target platform agnostic. It means that for some environments (e.g. GCP)"
    echo " it might be required to export firewall::open_port and firewall::close_port functions"
    echo " so the tests can open custom ports"

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

log() {
    echo -e "$(date +'%H:%M:%S.%N')    $*"
}

var_should_not_be_empty() {
    local var_name=${1}
    if [ -z "${!var_name+x}" ]; then
        echo >&2 "Environmental variable '${var_name}' is unset"
        return 1
    elif [ -z "${!var_name}" ]; then
        echo >&2 "Environmental variable '${var_name}' is empty"
        return 1
    fi
}

function_should_be_exported() {
    local fname=${1}
    if ! declare -F "${fname}"; then
        log "WARNING: Function '${fname}' is unexported. It is expected that function is provided for interacting with cloud provider"
        return 1
    fi
}

check_passwordless_ssh() {
    ssh -o BatchMode=yes "${USHIFT_USER}@${USHIFT_IP}" "true" || {
        echo "Failed to access ${USHIFT_IP}:"
        echo "  - Test runner should have MicroShift's sshd key in ~/.ssh/known_keys"
        echo "  - Remote \$USHIFT_USER should have test runner's key in ~/.ssh/authorized_keys"
        exit 1
    }
}

check_passwordless_sudo() {
    ssh -o BatchMode=yes "${USHIFT_USER}@${USHIFT_IP}" "sudo --non-interactive true" || {
        echo "Failed to run sudo command as ${USHIFT_USER} without password"
        exit 1
    }
}

prechecks() {
    var_should_not_be_empty USHIFT_IP && var_should_not_be_empty USHIFT_USER || exit 1
    check_passwordless_ssh
    check_passwordless_sudo

    # Just warning for now
    # Following functions needed only for runs in CI
    function_should_be_exported firewall::open_port || true
    function_should_be_exported firewall::close_port || true
}

microshift_get_konfig() {
    tmpfile=$(mktemp /tmp/microshift-e2e-konfig.XXXXXX)
    ssh "${USHIFT_USER}@${USHIFT_IP}" 'sudo cat /var/lib/microshift/resources/kubeadmin/'"${USHIFT_IP}"'/kubeconfig' >"${tmpfile}"
    echo "${tmpfile}"
}

microshift_check_readiness() {
    local test_output="${1}"
    log "Waiting for MicroShift to become ready"
    ssh "${USHIFT_USER}@${USHIFT_IP}" \
        "sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh" &>"${test_output}/0002-readiness-check.log"
}

microshift_setup() {
    local test_output="${1}"
    log "Setting up and starting MicroShift"
    ssh "${USHIFT_USER}@${USHIFT_IP}" 'cat << EOF | sudo tee /etc/microshift/config.yaml
---
apiServer:
  subjectAltNames:
  - '"${USHIFT_IP}"'
EOF' &>"${test_output}/0001-setup.log"
    ssh "${USHIFT_USER}@${USHIFT_IP}" "sudo systemctl enable --now microshift" &>>"${test_output}/0001-setup.log"
}

microshift_debug_info() {
    local test_output="${1}"
    log "Gathering debug info to ${test_output}/0020-cluster-debug-info.log"
    scp "${SCRIPT_DIR}/../validate-microshift/cluster-debug-info.sh" "${USHIFT_USER}@${USHIFT_IP}:/tmp/cluster-debug-info.sh"
    ssh "${USHIFT_USER}@${USHIFT_IP}" "sudo /tmp/cluster-debug-info.sh" &>"${test_output}/0020-cluster-debug-info.log"
}

microshift_cleanup() {
    local test_output="${1}"
    log "Cleaning MicroShift"
    ssh "${USHIFT_USER}@${USHIFT_IP}" "echo 1 | sudo microshift-cleanup-data --all" &>"${test_output}/0000-cleanup.log"
}

microshift_health_summary() {
    log "Summary of MicroShift health"

    # Because test might be "destructive" (i.e. tear down and set up again MicroShift)
    # so these commands are executed via ssh.
    # Alternative is to copy kubeconfig second time in the same test.
    ssh "${USHIFT_USER}@${USHIFT_IP}" \
        "mkdir -p ~/.kube/ && sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config ; \
            oc get pods -A ; \
            oc get nodes -o wide ; \
            oc get events -A --sort-by=.metadata.creationTimestamp | head -n 20"
}

run_test() {
    local test=$1
    echo -e "\n\n=================================================="
    log "${test} - PREPARING"

    local test_output="${OUTPUT_DIR}/${test}/"
    mkdir -p "${test_output}"

    prep_start=$(date +%s)
    microshift_cleanup "${test_output}"
    microshift_setup "${test_output}"
    microshift_check_readiness "${test_output}"
    konfig=$(microshift_get_konfig)
    trap 'rm -f "${konfig}"' RETURN
    prep_dur=$(($(date +%s) - prep_start))
    log "Cleanup, setup, and readiness took $((prep_dur / 60))m $((prep_dur % 60))s."

    log "${test} - RUNNING"
    test_start=$(date +%s)
    set +e
    KUBECONFIG="${konfig}" "${SCRIPT_DIR}/tests/${test}" &>"${test_output}/0010-test.log"
    res=$?
    set -e
    test_dur=$(($(date +%s) - test_start))

    log "${test} took $((test_dur / 60))m $((test_dur % 60))s."
    if [ ${res} -eq 0 ]; then
        log "${test} - SUCCESS"
        return 0
    fi

    log "${test} - FAILURE"
    microshift_health_summary
    microshift_debug_info "${test_output}" || true
    return 1
}

list() {
    local -r filter="*${1:-}*.sh"
    find "${SCRIPT_DIR}/tests" -maxdepth 1 -iname "${filter}" -printf "%f\n" | sort
}

run() {
    local -r to_run=$(list "${1}")

    prechecks
    log "Following tests will run:\n${to_run}"
    [ ! -d "${OUTPUT_DIR}" ] && mkdir -p "${OUTPUT_DIR}"

    all_successful=true
    for t in ${to_run}; do
        run_test "${t}" || all_successful=false
    done
    "${all_successful}"
}

[ $# -eq 0 ] && {
    usage "Expected arguments"
}

cmd="$1"
shift
"${cmd}" "${1:-}"
