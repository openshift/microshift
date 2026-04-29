#!/bin/bash
# Copy the system CA bundle into the cert-manager manifests directory
# so kustomize can use it to create the trusted-ca-bundle ConfigMap.
#
# This script runs as ExecStartPre before MicroShift starts.

set -euo pipefail

SRC="/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"
DST="/usr/lib/microshift/manifests.d/060-microshift-cert-manager/manager/tls-ca-bundle.pem"

# Only copy if the cert-manager manifests directory exists (package installed)
if [ -d "$(dirname "${DST}")" ] && [ -f "${SRC}" ]; then
    cp -f "${SRC}" "${DST}"
fi
