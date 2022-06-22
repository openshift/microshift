# Deploying a TCP Load Balancer for User Workloads
MicroShift does not currently offer an implementation of network load balancers ([Services of type LoadBalancer](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer)). If load balancer facilities are required by user workloads, it is possible to deploy 3rd party load balancer services. 

This document demonstates how to deploy the [MetalLB](https://metallb.universe.tf) service, which is a load balancer implementation for bare metal clusters, using standard routing protocols.

## Create MicroShift Server
Use the instructions in the [Install MicroShift for Edge](./devenv_rhel8.md#install-microshift-for-edge) section to configure a virtual machine running MicroShift. 

Log into the virtual machine and run the following commands to configure the MicroShift access and check if the PODs are up and running.

```
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
oc get pods -A
```

## Install Load Balancer
Log into the virtual machine and run the following commands to create `MetalLB` namespace and deployment.

```
oc apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11.0/manifests/namespace.yaml
oc apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11.0/manifests/metallb.yaml
```

Once the components are available, create a `ConfigMap` required to define the address pool for the load balancer to use.

```bash
oc create -f - <<EOF
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
EOF
```

## Install User Workload
Log into the virtual machine and create the namespace to be used for deploying the test application.

```bash
NAMESPACE=nginx-lb-test
oc create ns $NAMESPACE
```

Run the following command to deploy **3 replicas** of a test `nginx` application in the specified namespace.

```bash
oc apply -n $NAMESPACE -f https://raw.githubusercontent.com/openshift/microshift/main/docs/config/nginx-IP-header.spec
```
The application is configured to return `X-Server-IP` header containing the container IP address.

Verify that all the **3 replicas** of the application started successfully.

```bash
oc get pods -n $NAMESPACE
```

## Create Load Balancer Service
Log into the virtual machine and run the following commands to create the `LoadBalancer` service for `nginx` application using the `MetalLB` address pool.

```bash
oc create -n $NAMESPACE -f - <<EOF
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
EOF
```

Verify that the service exists and an external IP address has been assigned to it.

```bash
oc get svc -n $NAMESPACE
NAME    TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
nginx   LoadBalancer   10.43.183.104   192.168.1.241   80:32434/TCP   2m
```

## Test Load Balancer
Log into the virtual machine and run the following commands to verify that the load balancer routes requests to all the running application instances.

> Set the `EXTERNAL_IP` environment variable to the external IP of the `LoadBalancer` service.

```bash
EXTERNAL_IP=192.168.1.241
seq 5 | xargs -Iz curl -s -I http://$EXTERNAL_IP | grep X-Server-IP
```

The above command attempts to perform five connections to the `nginx` application using the `LoadBalancer` external IP address and print the value of the `X-Server-IP` header. 

The output of this command should return different IP addresses, demonstrating that the load balancer service works by distributing the traffic among the instances of the application.

```
X-Server-IP: 10.42.0.41
X-Server-IP: 10.42.0.41
X-Server-IP: 10.42.0.43
X-Server-IP: 10.42.0.41
X-Server-IP: 10.42.0.43
```
