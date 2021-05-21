kubectl delete ns tff
kubectl create ns tff
kubectl create deployment -n tff tff-workers --image=gcr.io/tensorflow-federated/remote-executor-service:latest
kubectl scale deployment -n tff tff-workers --replicas=1
kubectl expose deployment -n tff tff-workers --type=NodePort --port 8000
NODE_PORT=$(kubectl get svc  -n tff tff-workers -o=jsonpath='{.spec.ports[0].nodePort}')
echo "start training on port" ${NODE_PORT}
NODE_IP="127.0.0.1" NODE_PORT=${NODE_PORT} python3 ./training.py
