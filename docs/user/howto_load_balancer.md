# Deploying a TCP Load Balancer type of Service for User Workloads
MicroShift offers an built-in implementation of network load balancers ([Services of type LoadBalancer](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer)). It will use the node IP as the ingress IP of the loadbalancer type of service.

## Create MicroShift Server
Use the instructions in the [Install MicroShift on RHEL for Edge](../contributor/rhel4edge_iso.md) document to configure a virtual machine running MicroShift.

Log into the virtual machine and run the following commands to configure the MicroShift access and check if the PODs are up and running.

```
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
oc get pods -A
```

## Install User Workload
Log into the virtual machine and create the namespace to be used for deploying the test application.

```bash
NAMESPACE=nginx-lb-test
oc create ns $NAMESPACE
```

Run the following command to deploy **3 replicas** of a test `nginx` application in the specified namespace.

```bash
oc apply -n $NAMESPACE -f https://raw.githubusercontent.com/openshift/microshift/main/docs/config/nginx-IP-header.yaml
```

The application is configured to return the `X-Server-IP` header with the container IP address.

Verify that all the **3 replicas** of the application started successfully.

```bash
oc get pods -n $NAMESPACE
```

## Create Load Balancer Service
Log into the virtual machine and run the following command to create the `LoadBalancer` service for `nginx` application.

```bash
oc create -n $NAMESPACE -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  ports:
  - port: 81
    targetPort: 8080
  selector:
    app: nginx
  type: LoadBalancer
EOF
```

> The `port` setting should be a host port that is not occupied by other
> `LoadBalancer` services, host processes or MicroShift components.

Verify that the service exists and an external IP address has been assigned to it. And the external IP is the same as the node IP.

```bash
oc get svc -n $NAMESPACE
NAME    TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
nginx   LoadBalancer   10.43.183.104   192.168.1.241   81:32434/TCP   2m
```

## Test Load Balancer
Log into the virtual machine and run the following commands to verify that the load balancer distributes requests among all the running application instances.

> Set the `EXTERNAL_IP` environment variable to the external IP of the `LoadBalancer` service.

```bash
EXTERNAL_IP=192.168.1.241
seq 5 | xargs -Iz curl -s -I http://$EXTERNAL_IP:81 | grep X-Server-IP
```

The above command attempts to perform five connections to the `nginx` application using the external IP address of the `LoadBalancer` service. Only the value of the `X-Server-IP` header is filtered from the application response.

The output of this command should contain different IP addresses, demonstrating that the load balancer service works by distributing the traffic among the instances of the application.

```
X-Server-IP: 10.42.0.41
X-Server-IP: 10.42.0.41
X-Server-IP: 10.42.0.43
X-Server-IP: 10.42.0.41
X-Server-IP: 10.42.0.43
```
