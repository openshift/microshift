#!/usr/bin/env bash

set -o errtrace
set -o nounset
set -o pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

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
    log "Waiting for MicroShift to become ready"
    ssh "$USHIFT_USER@$USHIFT_IP" \
        "sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh | \
        while IFS= read -r line; do printf '%s %s\\n' \"\$(date +'%H:%M:%S.%N')\" \"\$line\"; done"
}

microshift_setup() {
    log "Setting up and starting MicroShift"
    ssh "$USHIFT_USER@$USHIFT_IP" 'cat << EOF | sudo tee /etc/microshift/config.yaml
---
apiServer:
  subjectAltNames:
  - '"$USHIFT_IP"'
EOF'
    ssh "$USHIFT_USER@$USHIFT_IP" "sudo systemctl enable --now microshift"
}

microshift_debug_info() {
    log "Gathering debug info"
    scp "$SCRIPT_DIR/cluster-debug-info.sh" "$USHIFT_USER@$USHIFT_IP:/tmp/cluster-debug-info.sh"
    ssh "$USHIFT_USER@$USHIFT_IP" "sudo KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig /tmp/cluster-debug-info.sh"
}

microshift_cleanup() {
    log "Cleaning MicroShift"
    ssh "$USHIFT_USER@$USHIFT_IP" "echo 1 | sudo microshift-cleanup-data --all"
}

run_test() {
    local test=$1
    echo -e "\n\n=================================================="
    log "${test} - RUNNING"

    microshift_cleanup
    microshift_setup
    microshift_check_readiness
    konfig=$(microshift_get_konfig)

    start=$(date +%s)
    set +e
    KUBECONFIG="$konfig" "${SCRIPT_DIR}/tests/${test}"
    res=$?
    set -e
    rm -f "${konfig}"

    end="$(date +%s)"
    duration_total_seconds=$((end - start))
    log "${test} - took ${duration_total_seconds}s"
    if [ $res -eq 0 ]; then
        log "${test} - SUCCESS"
        return 0
    fi

    log "${test} - FAILURE"
    microshift_debug_info || true # TODO: > artifacts/file
    return 1
}

list() {
    find "${SCRIPT_DIR}/tests" -maxdepth 1 -iname "*.sh" -printf "%f\n"
}

run() {
    local -r filter="*${1:-}*.sh"

    local -r to_run=$(find "${SCRIPT_DIR}/tests" -maxdepth 1 -iname "$filter" -printf "%f\n")
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
