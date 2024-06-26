#!/usr/bin/env bash

# shellcheck source=/dev/null
. /usr/lib/tuned/functions

cpu_online_offline() {
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
    cpu_online_offline 0
    # TODO /proc/sys/kernel/sched_rt_runtime_us=-1
}

stop() {
    cpu_online_offline 1
}

# shellcheck disable=SC2068
process $@
