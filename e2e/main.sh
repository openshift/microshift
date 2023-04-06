#!/usr/bin/env bash

set -o errtrace
set -o nounset
set -o pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
OUTPUT_DIR="${ARTIFACT_DIR:-_output}/microshift-e2e-$(date +'%Y%m%d-%H%M%S')/"
[ ! -d "${OUTPUT_DIR}" ] && mkdir -p "${OUTPUT_DIR}"

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
    if ! declare -F "$fname"; then
        echo >&2 "Function '$fname' is unexported. It is expected that function is provided for interacting with cloud provider"
        return 1
    fi
}

prechecks() {
    var_should_not_be_empty USHIFT_IP || exit 1
    var_should_not_be_empty USHIFT_USER || exit 1
    # TODO Check passwordless SSH
    # TODO Check passwordless sudo

    # Just warning for now
    # Following functions needed only for runs in CI
    function_should_be_exported firewall::open_port
    function_should_be_exported firewall::close_port
}

microshift_get_konfig() {
    tmpfile=$(mktemp /tmp/microshift-e2e-konfig.XXXXXX)
    ssh "$USHIFT_USER@$USHIFT_IP" 'sudo cat /var/lib/microshift/resources/kubeadmin/'"$USHIFT_IP"'/kubeconfig' >"$tmpfile"
    echo "$tmpfile"
}

microshift_check_readiness() {
    local test_output="${1}"
    log "Waiting for MicroShift to become ready"
    ssh "$USHIFT_USER@$USHIFT_IP" \
        "sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh | \
        while IFS= read -r line; do printf '%s %s\\n' \"\$(date +'%H:%M:%S.%N')\" \"\$line\"; done" &>"${test_output}/0002-readiness-check.log"
}

microshift_setup() {
    local test_output="${1}"
    log "Setting up and starting MicroShift"
    ssh "$USHIFT_USER@$USHIFT_IP" 'cat << EOF | sudo tee /etc/microshift/config.yaml
---
apiServer:
  subjectAltNames:
  - '"$USHIFT_IP"'
EOF' &>"${test_output}/0001-setup.log"
    ssh "$USHIFT_USER@$USHIFT_IP" "sudo systemctl enable --now microshift" &>>"${test_output}/0001-setup.log"
}

microshift_debug_info() {
    local test_output="${1}"
    log "Gathering debug info to ${test_output}/0020-cluster-debug-info.log"
    scp "$SCRIPT_DIR/../validate-microshift/cluster-debug-info.sh" "$USHIFT_USER@$USHIFT_IP:/tmp/cluster-debug-info.sh"
    ssh "$USHIFT_USER@$USHIFT_IP" "sudo /tmp/cluster-debug-info.sh" &>"${test_output}/0020-cluster-debug-info.log"
}

microshift_cleanup() {
    local test_output="${1}"
    log "Cleaning MicroShift"
    ssh "$USHIFT_USER@$USHIFT_IP" "echo 1 | sudo microshift-cleanup-data --all" &>"${test_output}/0000-cleanup.log"
}

microshift_health_summary() {
    log "Summary of MicroShift health"

    # Because test might be "destructive" (i.e. tear down and set up again MicroShift)
    # so these commands are executed via ssh.
    # Alternative is to copy kubeconfig second time in the same time.
    ssh "$USHIFT_USER@$USHIFT_IP" \
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
    KUBECONFIG="$konfig" "${SCRIPT_DIR}/tests/${test}" &>"${test_output}/0010-test.log"
    res=$?
    set -e
    test_dur=$(($(date +%s) - test_start))

    log "${test} took $((test_dur / 60))m $((test_dur % 60))s."
    if [ $res -eq 0 ]; then
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
    find "${SCRIPT_DIR}/tests" -maxdepth 1 -iname "$filter" -printf "%f\n"
}

run() {
    local -r to_run=$(list "${1}")
    log "Following tests will run:\n$to_run"

    prechecks
    all_successful=true
    for t in $to_run; do
        run_test "${t}" || all_successful=false
    done
    "${all_successful}"
}

[ $# -eq 0 ] && {
    echo "usage"
    exit 1
}

cmd="$1"
shift
"${cmd}" "${1:-}"
