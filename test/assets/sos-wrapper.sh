#!/usr/bin/env bash
if ! hash sos ; then
    echo "WARNING: The sos command does not exist"
elif [ -f /usr/bin/microshift-sos-report ]; then
    /usr/bin/microshift-sos-report || {
        echo "WARNING: The /usr/bin/microshift-sos-report script failed"
    }
else
    chmod +x /tmp/microshift-sos-report.sh
    PROFILES=network,security /tmp/microshift-sos-report.sh || {
        echo "WARNING: The /tmp/microshift-sos-report.sh script failed"
    }
fi
chmod +r /tmp/sosreport-* || echo "WARNING: The sos report files do not exist in /tmp"