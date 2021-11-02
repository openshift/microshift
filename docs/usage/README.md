# Using MicroShift
MicroShift operates similar to many other Kubernetes providers. This means that you can use the same tools to deploy and manage your applications.

All of the standard Kubernetes management tools can be used to maintain and modify your MicroShift applications. Below we will show some examples using kubectl, kustomize, and helm to deploy and maintain applications.

## Example Applications

### Metal LB
Metal LB is a load balancer that can be used to route traffic to a number of backends.

Creating the Metal LB namespace and deployment.
```
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11.0/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11.0/manifests/metallb.yaml
```

Once the components are available a configMap is required to define the address pool for the load balancer to use.

Create the configmap.yaml file
```
vi configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - 192.168.1.240-192.168.1.250
```

Load the configMap into the cluster.
```
kubectl create -f /tmp/configmap.yaml
```

Now we are able to deploy a test application to verify thing are working as expected.

```
kubectl create ns test
kubectl create deployment nginx -n test --image nginx
```

Create the following file to be used for the service.
```
vi service.yaml

apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    metallb.universe.tf/address-pool: default
spec:
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer
```

Use kubectl to create the service.
```
kubectl create -f service.yaml -n test
```

Verify the service exists and that an IP address has been assigned.
```
kubectl get svc -n test
NAME    TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
nginx   LoadBalancer   10.43.183.104   192.168.1.241   80:32434/TCP   29m
```

Using your browser you can now access the NGINX application by the `EXTERNAL-IP` provided by the service.
