#!/usr/bin/env bash

set -xeuo pipefail

retries=20
while [ ${retries} -gt 0 ] ; do
    ((retries-=1))
    if sudo systemctl status greenboot-healthcheck | grep -q 'active (exited)'; then
      exit 0
    fi
    echo "Not ready yet. Waiting 30 seconds... (${retries} retries remaining)"
    sleep 30
done
exit 1
