#!/usr/bin/env bash

# shellcheck source=/dev/null
source /usr/lib/tuned/functions

toggle_cpu_state() {
    # 0 - set CPUs offline
    # 1 - set CPUs online
    local -r op=$1
    # shellcheck disable=SC2154
    # shellcheck disable=SC2086
    for CPU in ${TUNED_offline_cpu_set_expanded//,/ }; do
        echo "${op}" | sudo tee /sys/devices/system/cpu/cpu${CPU}/online
    done
}

start() {
    toggle_cpu_state 0
}

stop() {
    toggle_cpu_state 1
}

# shellcheck disable=SC2068
process $@
