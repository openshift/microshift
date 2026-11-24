#!/bin/bash
# Run a command every 5 seconds until its output differs from a value or matches a
# shell pattern, then print that output. The caller bounds the whole poll with timeout.
# Usage: poll.sh differs <value> <command...>
#        poll.sh matches <pattern> <command...>
set -u
mode=${1:?mode}
want=${2?value}
shift 2
while :; do
    if out=$("$@"); then
        if [ "${mode}" = differs ]; then
            [ "${out}" != "${want}" ] && break
        else
            # shellcheck disable=SC2254
            case ${out} in ${want}) break ;; esac
        fi
    fi
    sleep 5
done
printf '%s\n' "${out}"
