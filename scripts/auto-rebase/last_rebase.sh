#!/bin/bash -x
./scripts/auto-rebase/rebase.sh to "registry.ci.openshift.org/ocp/release:4.14.0-0.nightly-2023-07-26-001154" "registry.ci.openshift.org/ocp-arm64/release-arm64:4.14.0-0.nightly-arm64-2023-07-25-210314" "registry.access.redhat.com/lvms4/lvms-operator-bundle:v4.13.1-5"
