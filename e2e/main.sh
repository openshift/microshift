#!/usr/bin/env bash

set -o errtrace
set -o nounset
set -o pipefail

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

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

file_should_exist() {
    local file_var="$1"

    var_should_not_be_empty "$file_var" || return 1
    local file="${!file_var}"

    if [[ ! -e "$file" ]]; then
        echo >&2 "File '$file' does not exist"
        return 1
    fi

    if [[ ! -f "$file" ]]; then
        echo >&2 "'$file' is not a file"
        return 1
    fi

    if [[ ! -r "$file" ]]; then
        echo >&2 "'$file' is missing 'read' permissions"
        return 1
    fi

}

prechecks() {
    # TODO: It'd be nice to run all prechecks each time, but exit at the end if any is unfullfiled
    file_should_exist KUBECONFIG || exit 1
    var_should_not_be_empty USHIFT_IP || exit 1
    var_should_not_be_empty USHIFT_USER || exit 1
    # TODO Check passwordless SSH
    # TODO Check passwordless sudo

    # Just warning for now
    function_should_be_exported firewall::open_port  # || exit 1
    function_should_be_exported firewall::close_port # || exit 1
}

# TODO: check_if_cluster_is_ready() {} - kuttl?

main() {
    prechecks

    export IP="$USHIFT_IP"
    export USER="$USHIFT_USER"

    local gather_debug=false
    for test in "${SCRIPT_PATH}"/tests/*.sh; do
        echo "RUNNING $(basename "${test}")"
        ${test}
        res=$?
        if [ $res -eq 0 ]; then
            echo "SUCCESS"
        else
            echo "FAILURE"
            gather_debug=true
        fi
    done

    if ${gather_debug}; then
        echo "Run cluster-debug-info.sh"
        exit 1
    fi
}

main
