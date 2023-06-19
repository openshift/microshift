#!/bin/bash -x
./scripts/auto-rebase/rebase.sh to "registry.ci.openshift.org/ocp/release:4.14.0-0.nightly-2023-06-18-111801" "registry.ci.openshift.org/ocp-arm64/release-arm64:4.14.0-0.nightly-arm64-2023-06-18-233314" "registry.access.redhat.com/lvms4/lvms-operator-bundle:v4.12"
