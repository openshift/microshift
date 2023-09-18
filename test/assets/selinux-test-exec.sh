#!/bin/bash
set -euo pipefail

process_context_switch() {
    echo "Forked process running under context $(id -Z)"
    "$@"
}

# We fork a process here to try and escape
echo "Executing Command: $*"
process_context_switch "$@" &
wait $!
