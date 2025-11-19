#!/bin/bash

set -eux

# Set KUBECONFIG path - use dynamic IP-based path
KUBECONFIG=${KUBECONFIG:-/var/lib/microshift/resources/kubeadmin/kubeconfig}

ns="test-lvms"
appLabel="app-lvms"

echo "INFO:" "Create Namespace, PVC and Deployment resources......"
sudo oc --kubeconfig "${KUBECONFIG}" create ns ${ns}

cat <<EOF | sudo oc --kubeconfig "${KUBECONFIG}" -n ${ns} apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: mypvc-lvms
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: topolvm-provisioner
  volumeMode: Filesystem
  resources:
    requests:
      storage: 1Gi
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: mydep-lvms
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${appLabel}
  template:
    metadata:
      labels:
        app: ${appLabel}
    spec:
      containers:
      - name: http-server
        image: quay.io/openshifttest/hello-openshift@sha256:b1aabe8c8272f750ce757b6c4263a2712796297511e0c6df79144ee188933623
        ports:
          - name: httpd
            containerPort: 80
        volumeMounts:
        - name: local
          mountPath: /mnt/storage
      volumes:
      - name: local
        persistentVolumeClaim:
            claimName: mypvc-lvms
EOF

echo "INFO:" "Check if deployment Pod is 'Running'......."
iter=48
period=5
podName=""
set +e
while [[ "${podName=}" == "" && ${iter} -gt 0 ]]; do
  sleep ${period}
  podName=$(sudo oc --kubeconfig "${KUBECONFIG}" get pod -n ${ns} -l app=${appLabel} --no-headers | awk '{print $1}')
  (( iter -- ))
done
if [ "${podName}" == "" ]; then
  echo "ERROR:" "Deployment Pod not found."
  exit 1
fi
echo "INFO:" "Waiting for Pod to become ready(max 4-minutes)....."
iter=48
result=""
while [[ "${result}" != "Running" && ${iter} -gt 0 ]]; do
  #shellcheck disable=SC1083
  result=$(sudo oc --kubeconfig "${KUBECONFIG}" get pod "${podName}" -n ${ns} -o=jsonpath={.status.phase})
  (( iter -- ))
  sleep ${period}
done
set -e
if [ "${result}" == "Running" ]; then
  echo "INFO:" "Deployment Pod is Running."
else
  echo "ERROR:" "Deployment Pod creation Failed."
  sudo oc --kubeconfig "${KUBECONFIG}" -n ${ns} describe pod "${podName}"
  exit 1
fi

echo "INFO:" "Check if data can be read/write into Pod mounted volume......."
#shellcheck disable=SC2016
sudo oc --kubeconfig "${KUBECONFIG}" exec -n ${ns} "${podName}" -- /bin/sh -c 'echo Storage_Test $(date) > /mnt/storage/testfile'
data=$(sudo oc --kubeconfig "${KUBECONFIG}" exec -n ${ns} "${podName}" -- /bin/sh -c 'cat /mnt/storage/testfile')
if [[ ${data} =~ "Storage_Test" ]]; then
  echo "SUCCESS:" "Data successfully written into Pod"
else
  echo "ERROR:" "Failed to write data into the Pod"
  exit 1
fi
