#!/usr/bin/env bash

# reprovision_after_test=false

IFS=$'\n\t'
PS4='+ $(date "+%T.%N")\011 '

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

cleanup() {
    ssh -q "${USHIFT_USER}@${USHIFT_IP}" "rm -f /tmp/selinux-verify.sh /tmp/audit-results.txt"
}
trap 'cleanup' EXIT

scp "${SCRIPT_PATH}/assets/selinux-verify.sh" "${USHIFT_USER}@${USHIFT_IP}":/tmp/
ssh -q "${USHIFT_USER}@${USHIFT_IP}" "chmod +x /tmp/selinux-verify.sh && /tmp/selinux-verify.sh"

# If audit-results.txt does not exist, then there were no errors discovered
if results=$(ssh "${USHIFT_USER}@${USHIFT_IP}" "cat /tmp/audit-results.txt");
then
    echo "Failed to Valiate SELinux"
    echo "${results}"
    exit 1
else
    echo " Successfully validate SELinux"
fi
