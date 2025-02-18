#!/bin/bash

set -xeuo pipefail

sudo subscription-manager repos --enable rhel-9-for-x86_64-rt-rpms
sudo dnf install kernel-rt realtime-setup realtime-tests -y
sudo grubby --set-default="/boot/vmlinuz-$(rpm -q --queryformat '%{version}-%{release}.%{arch}' kernel-rt | sort | tail -1)+rt"

sudo dnf upgrade tuned -y

# After this point, nothing new will be installed or updated,
# so let's list installed packages for debugging purposes.
sudo dnf list --installed
