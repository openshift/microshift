#!/bin/bash

set -xeuo pipefail

sudo subscription-manager repos --enable rhel-9-for-x86_64-rt-rpms
sudo dnf install kernel-rt realtime-setup realtime-tests -y
sudo grubby --set-default="/boot/vmlinuz-$(rpm -q --queryformat '%{version}-%{release}.%{arch}' kernel-rt | sort | tail -1)+rt"
