#! /usr/bin/env bash

set -euo pipefail

CLUSTER_NAME=${1:? "\$1 should be a cluster name in the standard FQDN format"}

WORK_DIR=$HOME/.acm/"$CLUSTER_NAME"
SPOKE_DIR="$WORK_DIR"/spoke
rm -rf "$WORK_DIR"
mkdir -p "$WORK_DIR" "$SPOKE_DIR"

cat <<EOF >"$WORK_DIR"/managed-cluster.yaml
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
metadata:
  name: "$CLUSTER_NAME"
spec:
  hubAcceptsClient: true
EOF

cat <<EOF >"$WORK_DIR"/klusterlet-addon-config.yaml
apiVersion: agent.open-cluster-management.io/v1
kind: KlusterletAddonConfig
metadata:
  name: "$CLUSTER_NAME"
  namespace: "$CLUSTER_NAME"
spec:
  clusterName: "$CLUSTER_NAME"
  clusterNamespace: "$CLUSTER_NAME"
  applicationManager:
    enabled: true
  certPolicyController:
    enabled: true
  clusterLabels:
    cloud: auto-detect
    vendor: auto-detect
  iamPolicyController:
    enabled: true
  policyController:
    enabled: true
  searchCollector:
    enabled: true
  version: 2.2.0
EOF

oc new-project "$CLUSTER_NAME" 1>/dev/null
oc label namespace "$CLUSTER_NAME" cluster.open-cluster-management.io/managedCluster="$CLUSTER_NAME"
oc apply -f "$WORK_DIR"/managed-cluster.yaml
oc apply -f "$WORK_DIR"/klusterlet-addon-config.yaml
sleep 3
echo "Creating kluster-crd.yaml and import.yaml.  These files need to be accessible to the client script"
oc get secret "$CLUSTER_NAME"-import -n "$CLUSTER_NAME" -o jsonpath={.data.crds\\.yaml} | base64 --decode >"$SPOKE_DIR"/klusterlet-crd.yaml
oc get secret "$CLUSTER_NAME"-import -n "$CLUSTER_NAME" -o jsonpath={.data.import\\.yaml} | base64 --decode >"$SPOKE_DIR"/import.yaml
