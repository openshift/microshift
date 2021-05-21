set -e

kubectl delete ns test-route || true
kubectl create ns test-route

cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: test-route
spec:
  selector:
    matchLabels:
      run: nginx
  replicas: 1
  template:
    metadata:
      labels:
        run: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
EOF

kubectl wait --for=condition=available -n test-route deployment/nginx  --timeout=30s

kubectl expose -n test-route deployment/nginx
oc expose  -n test-route service nginx

curl --connect-timeout 5 http://$(kubectl get svc nginx -n test-route -o=jsonpath='{.spec.clusterIP}')

ROUTE=$(oc get routes -n test-route -o=jsonpath='{.items[0].status.ingress[0].host}')

if [ ${ROUTE} == "" ]; then
  echo "test failed"
  exit 1
fi
#curl --connect-timeout 5 http://${ROUTE}

echo "test passed"