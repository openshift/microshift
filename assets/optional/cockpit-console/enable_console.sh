#!/bin/bash

set -euo pipefail


export KUBECONFIG=/home/microshift/kubeconfig
#oc get pods
export hostname=$(hostname -f)
export ipAddress=$(hostname -i)
export token=$(oc create token -n kube-system openshift-console)


cat<<__EOF__ >/tmp/run.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: openshift-console-cluster-role-binding
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: openshift-console
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openshift-console-deployment
  namespace: kube-system
  labels:
    app: openshift-console
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openshift-console
  template:
    metadata:
      labels:
        app: openshift-console
    spec:
      securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
      containers:
      - name: openshift-console-app
        image: quay.io/openshift/origin-console:latest
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
        env:
        - name: BRIDGE_USER_AUTH
          value: disabled
        - name: BRIDGE_K8S_MODE
          value: off-cluster
        - name: BRIDGE_K8S_MODE_OFF_CLUSTER_ENDPOINT
          value: https://${hostname}:6443
        - name: BRIDGE_K8S_MODE_OFF_CLUSTER_SKIP_VERIFY_TLS
          value: "true"
        - name: BRIDGE_K8S_AUTH
          value: bearer-token
        - name: BRIDGE_K8S_AUTH_BEARER_TOKEN
          value: "${token}"
---
apiVersion: v1
kind: Service
metadata:
  name: openshift-console-service
  namespace: kube-system
spec:
  selector:
    app: openshift-console
  ports:
  - port: 9000
    targetPort: 9000
  type: LoadBalancer
__EOF__

oc apply -f /tmp/run.yaml >/dev/null

echo "http://${ipAddress}:9000"