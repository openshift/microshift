#!/bin/bash

create_and_wait_for_pod() {
    local -r yaml="${1}"
    local -r pod_name="${2}"

    oc create -f "${yaml}"
    if ! oc wait pod "${pod_name}" --for=condition=Ready --timeout 5m; then
        oc describe pod "${pod_name}"
        oc logs "${pod_name}"
        exit 1
    fi
    oc get pods -A
}

wait_for_test_to_finish() {
    local -r pod="${1}"
    local -r test_duration="${2}"
    local -r finish_indicator="${3}"

    : Waiting for a duration of a test
    sleep "${test_duration}"

    : Giving another minute to finish the test for robustment
    start_time=$(date +%s)
    while true ; do
        if oc logs "${pod}" | grep "${finish_indicator}"; then
            break
        fi
        if [ $(( $(date +%s) - start_time )) -gt 60 ]; then
            echo "ERROR: Test didn't finish in time"
            exit 1
        fi
        sleep 5
    done
}
