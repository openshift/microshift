#!/bin/sh

export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig

oc apply -f - <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
  name: scc-check
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: default-role 
  namespace: scc-check 
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: default-rolebinding
  namespace: scc-check 
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: default-role # Should match name of Role
subjects:
- namespace: scc-check # Should match namespace where SA lives
  kind: ServiceAccount
  name: default 
EOF

SERVICE_ACCOUNT_NAME=default
NAMESPACE=scc-check
SERVER=https://127.0.0.1:6443
SECRET_NAME=$(kubectl get serviceaccount ${SERVICE_ACCOUNT_NAME} \
  --namespace ${NAMESPACE} \
  -o jsonpath='{.secrets[0].name}')

ca=$(kubectl get -n $NAMESPACE secret/$SECRET_NAME -o jsonpath='{.data.ca\.crt}')
token=$(kubectl get -n $NAMESPACE secret/$SECRET_NAME -o jsonpath='{.data.token}' | base64 --decode)
namespace=$(kubectl get -n $NAMESPACE secret/$SECRET_NAME -o jsonpath='{.data.namespace}' | base64 --decode)

echo "
apiVersion: v1
kind: Config
clusters:
- name: default-cluster
  cluster:
    certificate-authority-data: ${ca}
    server: ${SERVER}
contexts:
- name: default-context
  context:
    cluster: default-cluster
    namespace: ${NAMESPACE} 
    user: default-user
current-context: default-context
users:
- name: default-user
  user:
    token: ${token}
" > sa.kubeconfig


