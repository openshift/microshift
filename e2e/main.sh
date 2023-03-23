#!/usr/bin/env bash

set -o errtrace
set -o nounset
set -o pipefail

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

var_should_not_be_empty() {
    local var=${1}
    if [ -z "${!var}" ]; then
        echo >&2 "Environmental variable '${var}' is empty"
        exit 1
    fi
}

verify_gcloud() {
    var_should_not_be_empty INSTANCE_PREFIX
    var_should_not_be_empty GOOGLE_PROJECT_ID
    var_should_not_be_empty GOOGLE_COMPUTE_REGION
    var_should_not_be_empty GOOGLE_COMPUTE_ZONE
}

prepare_kubeconfig() {
    export IP_ADDRESS="$(gcloud compute instances describe ${INSTANCE_PREFIX} --format='get(networkInterfaces[0].accessConfigs[0].natIP)')"
    gssh "sudo cat /var/lib/microshift/resources/kubeadmin/${IP_ADDRESS}/kubeconfig" >/tmp/kubeconfig
    export KUBECONFIG=/tmp/kubeconfig
}

main() {
    verify_gcloud
    prepare_kubeconfig

    local gather_debug=false
    for test in ${SCRIPT_PATH}/tests/*; do
        echo "RUNNING $(basename ${test})"
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
