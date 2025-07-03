# **Route Controller Manager Hacking**

## Building and Deploying

This document explains how to modify and test changes to the route-controller-manager in an OpenShift cluster.

### Prerequisites

* An OpenShift cluster
* A cluster-admin scoped KUBECONFIG for the cluster

### Building the Route Controller Manager Image

To build the route-controller-manager image you can use `podman`.

```bash
$ podman build -t quay.io/username/route-controller-manager:TAG .
````

### Preparing the Cluster

To modify the route-controller-manager image, you need to scale down the operators managing it. This will allow you to make changes without the operators automatically reverting them.

### Scaling Down Operators

First, scale down the cluster-version-operator (CVO) and the openshift-controller-manager-operator. These operators will try to manage and reconcile changes, so stopping them temporarily will let you make changes to the route-controller-manager image.
```bash
$ oc -n openshift-cluster-version scale --replicas=0 deployments/cluster-version-operator
$ oc -n openshift-controller-manager-operator scale --replicas=0 deployments/openshift-controller-manager-operator
```
Confirm that the deployments have been scaled down by checking their status:

```bash
$ oc -n openshift-cluster-version cluster-version-operator get deployments
$ oc -n openshift-controller-manager-operator get deployments openshift-controller-manager-operator
```

### Modifying the Route Controller Manager Image

Once the relevant operators are scaled down, you can proceed to update the route-controller-manager deployment using one of these methods:

Using oc set image (recommended):
```bash
$ oc -n openshift-route-controller-manager set image deployment/route-controller-manager route-controller-manager=quay.io/username/route-controller-manager:TAG
```
Setting imagePullPolicy to Always (recommended for testing):
```bash
$ oc -n openshift-route-controller-manager patch deployment/route-controller-manager -p '{"spec": {"template": {"spec": {"containers": [{"name": "route-controller-manager","imagePullPolicy": "Always"}]}}}}'
```
This ensures that when pushing a new version of the image with the same tag, the node will still pull the updated image.

### Restoring the Operators

After you've tested your changes, scale the operators back up to restore normal management:
```bash
$ oc -n openshift-cluster-version scale --replicas=1 deployments/cluster-version-operator
```