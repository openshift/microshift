NODE_NAME=$(kubectl get nodes -o=jsonpath='{.items[0].metadata.name}')

kubectl delete ns test-pv
kubectl create ns test-pv

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: hostpath-provisioner
spec:
  capacity:
    storage: 8Gi
  accessModes:
  - ReadWriteOnce
  hostPath:
    path: "/var/hpvolumes"
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: kube-1-hpp-claim
  namespace: test-pv
  annotations: 
    kubevirt.io/provisionOnNode: "${NODE_NAME}"
spec:
  storageClassName: "kubevirt-hostpath-provisioner"
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Mi
---
kind: Pod
apiVersion: v1
metadata:
  name: test-pod
  namespace: test-pv
spec:
  containers:
  - name: test-pod
    image: gcr.io/google_containers/busybox:1.24
    command:
      - "/bin/sh"
    args:
      - "-c"
      - "sleep 30; touch /mnt/SUCCESS && exit 0 || exit 1"
    volumeMounts:
      - name: hostpath-pvc
        mountPath: "/mnt"
  nodeSelector:
        kubernetes.io/hostname: "${NODE_NAME}"
  restartPolicy: "Never"
  volumes:
    - name: hostpath-pvc
      persistentVolumeClaim:
        claimName: kube-1-hpp-claim
EOF

kubectl wait --for=condition=ready -n test-pv pod/test-pod  --timeout=30s

STATUS=$(kubectl get pod -n test-pv test-pod -o=jsonpath='{.status.phase}')

if [ "${STATUS}" != "Running" ]; then
    echo "test failed:" ${STATUS}
    exit 1
fi

echo "test passed"
#kubectl delete ns test-pv
