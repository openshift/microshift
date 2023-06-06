#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
PS4='+ $(date "+%T.%N")\011 '
set -x

function check_greenboot_exit_status() {
    local expectedRC=$1
    local cleanup=$2

    if [ "${cleanup}" -ne 0 ]; then
        # Skip image deletion to speed up the test and make it more reliable
        echo 1 | microshift-cleanup-data --all --keep-images
        systemctl enable --now microshift || true
    fi

    for check_script in $(find /etc/greenboot/check/ -name \*.sh | sort); do
        echo Running "${check_script}"...
        local currentRC=1
        if ${check_script}; then
            currentRC=0
        fi

        if [ ${currentRC} != "${expectedRC}" ]; then
            exit 1
        fi
    done
}

#
# Initial check must succeed (set timeout of 180s to speed up the process)
#
tee /etc/greenboot/greenboot.conf &>/dev/null <<EOF
MICROSHIFT_WAIT_TIMEOUT_SEC=180
EOF
check_greenboot_exit_status 0 1

#
# User workload health
# See https://github.com/openshift/microshift/blob/main/docs/greenboot_dev.md#user-workload-health
#
MANIFEST_DIR=/etc/microshift/manifests
mkdir -p ${MANIFEST_DIR}

tee ${MANIFEST_DIR}/busybox.yaml &>/dev/null <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-deployment
spec:
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: busybox
        image: BUSYBOX_IMAGE
        command:
          - sleep
          - "3600"
EOF

tee ${MANIFEST_DIR}/kustomization.yaml &>/dev/null <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: busybox
resources:
  - busybox.yaml
images:
  - name: BUSYBOX_IMAGE
    newName: busybox:1.35
EOF

SCRIPT_FILE=/etc/greenboot/check/required.d/50_busybox_running_check.sh
# The file comes from payload.tar
ln -s /tmp/busybox_running_check.sh ${SCRIPT_FILE}
trap 'rm -f ${SCRIPT_FILE}; rm -f /etc/microshift/manifests/*' EXIT # in case next check_greenboot_exit_status fails

check_greenboot_exit_status 0 1
rm -f ${SCRIPT_FILE}
rm -f /etc/microshift/manifests/*

#
# Service failure
# See https://github.com/openshift/microshift/blob/main/docs/greenboot_dev.md#microshift-service-failure
#
dnf remove -y hostname
check_greenboot_exit_status 1 1
dnf install -y hostname

#
# Pod failure
# https://github.com/openshift/microshift/blob/main/docs/greenboot_dev.md#microshift-pod-failure
#
YAML_FILE=/etc/microshift/config.yaml
tee ${YAML_FILE} &>/dev/null <<EOF
network:
  serviceNetwork:
  - 10.66.0.0/16
EOF

check_greenboot_exit_status 1 0
rm -f ${YAML_FILE}

#
# Last check must succeed
#
check_greenboot_exit_status 0 1
