#! /usr/bin/env bash

set -eu

CLUSTER_NAME=${1:?"expected the cluster name, required to find import data"}
WORK_DIR="$HOME/.acm/$CLUSTER_NAME/spoke"

oc apply -f "$WORK_DIR"/klusterlet-crd.yaml
sleep 2
oc apply -f "$WORK_DIR"/import.yaml
