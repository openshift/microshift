#!/bin/bash

set -xeuo pipefail

sudo subscription-manager repos --enable rhel-9-for-x86_64-rt-rpms
sudo dnf install kernel-rt realtime-setup realtime-tests -y
sudo grubby --set-default="/boot/vmlinuz-$(rpm -q --queryformat '%{version}-%{release}.%{arch}\n' kernel-rt | sort -V | tail -n 1)+rt"

# TEMPORARY CI MITIGATION: tuned 2.28.0 does not persist the microshift-baseline
# kernel arguments to the RHEL 9.8 BLS entry. Keep all TuneD packages in this
# transaction synchronized; the profile packages require the matching tuned NVR.
readonly TUNED_NVR="2.27.0-2.1.20270724git0eb28ac3.el9fdp"
readonly -a TUNED_PACKAGES=(
    "tuned-${TUNED_NVR}.noarch"
    "tuned-profiles-cpu-partitioning-${TUNED_NVR}.noarch"
    "tuned-profiles-realtime-${TUNED_NVR}.noarch"
)

# Check that every exact NVR is retained in the currently enabled repositories.
for package in "${TUNED_PACKAGES[@]}"; do
    if ! sudo dnf -q repoquery --available --qf '%{name}-%{version}-%{release}.%{arch}' "${package}" | grep -Fxq "${package}"; then
        echo "ERROR: required TuneD package ${package} is unavailable in the enabled repositories." >&2
        exit 1
    fi
done

# Resolve the exact downgrade before changing the host. This transaction
# deliberately does not authorize replacing installed packages by erasing them.
preflight_output=$(mktemp)
trap 'rm -f "${preflight_output}"' EXIT

set +e
sudo dnf --assumeno downgrade "${TUNED_PACKAGES[@]}" 2>&1 | tee "${preflight_output}"
preflight_status=${PIPESTATUS[0]}
set -e

if grep -Eqi '(^|[[:space:]])(Error|Problem):|No match for argument|No matches found|No matching [Pp]ackages|protected package|read-only|read only|permission denied|conflicting requests|dependency conflict|nothing provides|cannot install the best candidate|unable to resolve transaction' "${preflight_output}"; then
    echo "ERROR: TuneD downgrade preflight reported a blocking DNF diagnostic; refusing to continue." >&2
    exit 1
fi
if grep -Eq '^[[:space:]]*Nothing to do\.?[[:space:]]*$' "${preflight_output}"; then
    :
# DNF --assumeno refuses a resolved transaction with this exact summary and
# terminal message; other uses of "Operation aborted." are blocking failures.
elif [[ "${preflight_status}" -ne 0 ]] && awk '
    /^[[:space:]]*Transaction Summary[[:space:]]*$/ { transaction_summary = 1 }
    transaction_summary && /^[[:space:]]*Downgrade[[:space:]]+/ { downgrade = 1 }
    downgrade && /^[[:space:]]*Operation aborted\.[[:space:]]*$/ { refusal_line = NR }
    END { exit !(transaction_summary && downgrade && refusal_line == NR) }
' "${preflight_output}"; then
    :
else
    echo "ERROR: TuneD downgrade preflight failed (dnf exit status ${preflight_status})." >&2
    exit 1
fi
if grep -Eq '^[[:space:]]*(Removing|Obsoleting):' "${preflight_output}"; then
    echo "ERROR: TuneD downgrade preflight would remove or obsolete packages; refusing to continue." >&2
    exit 1
fi

sudo dnf downgrade -y "${TUNED_PACKAGES[@]}"

# After this point, nothing new will be installed or updated,
# so let's list installed packages for debugging purposes.
sudo dnf list --installed
