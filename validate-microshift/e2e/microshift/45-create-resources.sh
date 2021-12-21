#!/bin/bash
kubectl create ns uploader
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    app: uploader
  name: uploader
  namespace: uploader
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: uploader
status:
  loadBalancer: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: uploader
  name: uploader
  namespace: uploader
spec:
  replicas: 1
  selector:
    matchLabels:
      app: uploader
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: uploader
    spec:
      securityContext:
        runAsUser: 1000580000
        fsGroup: 1000580000
      containers:
      - image: quay.io/rcook/tools:php-demo
        name: tools
        ports:
        - containerPort: 8080
        volumeMounts:
          - name: uploader-persistent
            mountPath: /opt/app-root/src/uploaded
      volumes:
        - name: uploader-persistent
          persistentVolumeClaim:
            claimName: uploader-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: uploader-pvc
  namespace: uploader
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: kubevirt-hostpath-provisioner
  resources:
    requests:
      storage: 1Gi
EOF
