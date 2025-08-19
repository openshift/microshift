# Deploying metrics-server in MicroShift
This document describes the basic workflow and changes to deploy [metrics-server](https://github.com/kubernetes-sigs/metrics-server) in MicroShift.

## Create MicroShift cluster
Use the instructions in the [Install MicroShift on RHEL for Edge](../contributor/rhel4edge_iso.md) document to configure a virtual machine running MicroShift.

Log into the virtual machine and run the following commands to configure the MicroShift access and check if the PODs are up and running.

```
mkdir ~/.kube
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
oc get pods -A
```

## Install metrics-server
The [metrics-server](https://github.com/kubernetes-sigs/metrics-server) has a [ready-to-apply yaml file](https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml), but we need to change some parts of it to use the correct certificate authorities for Kubelet.

These are shared as a ConfigMap:
```bash
$ oc get configmap -n kube-system kubelet-client-ca -o yaml
apiVersion: v1
data:
  ca.crt: |
    ***redacted***
kind: ConfigMap
metadata:
  creationTimestamp: "2025-08-19T09:53:48Z"
  name: kubelet-client-ca
  namespace: kube-system
  resourceVersion: "511"
  uid: a530b86b-de6e-41d6-ba5e-333f9eebce65
```

In order to use them, we need to mount it as a volume for the metrics-server and use the CA as an argument for the Deployment's command.
```bash
curl -sL https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml | \
yq 'select(.kind == "Deployment") |= (
.spec.template.spec.volumes += [{"name": "ca-bundle", "configMap": {"name": "kubelet-client-ca"}}] |
.spec.template.spec.containers[0].volumeMounts += [{"name": "ca-bundle", "mountPath": "/var/run/secrets/kubernetes.io/certs/ca.crt", "subPath": "ca.crt", "readOnly": true}] |
.spec.template.spec.containers[0].args += "--kubelet-certificate-authority=/var/run/secrets/kubernetes.io/certs/ca.crt"
)' | \
oc apply -f -
```

Verify that the application started successfully in the `kube-system` namespace.
```bash
$ oc get pod -n kube-system
NAME                                       READY   STATUS    RESTARTS   AGE
csi-snapshot-controller-56d8f77b99-l5plk   1/1     Running   0          20m
metrics-server-85679c99-gh8v5              1/1     Running   0          30s
```

The deployment exposes the metrics API from the apiserver, allowing `top` commands to work:
```bash
$ kubectl top node
NAME             CPU(cores)   CPU(%)   MEMORY(bytes)   MEMORY(%)
microshift-dev   187m         4%       1531Mi          41%

$ kubectl top pod -A
NAMESPACE                  NAME                                       CPU(cores)   MEMORY(bytes)
kube-system                csi-snapshot-controller-56d8f77b99-l5plk   1m           9Mi
kube-system                metrics-server-85679c99-gh8v5              5m           19Mi
openshift-dns              dns-default-l6wl2                          3m           32Mi
openshift-dns              node-resolver-p8xtp                        0m           2Mi
openshift-ingress          router-default-67fc5ddcf9-8qrmr            1m           34Mi
openshift-ovn-kubernetes   ovnkube-master-dl2ck                       10m          208Mi
openshift-ovn-kubernetes   ovnkube-node-xmhc8                         1m           6Mi
openshift-service-ca       service-ca-5dcff54cc7-cf9ht                3m           25Mi
openshift-storage          lvms-operator-cf9d8978d-l4bcc              4m           2
```

It is also possible to access the raw API:
```bash
$ oc get --raw /apis/metrics.k8s.io/v1beta1/nodes | jq
{
  "kind": "NodeMetricsList",
  "apiVersion": "metrics.k8s.io/v1beta1",
  "metadata": {},
  "items": [
    {
      "metadata": {
        "name": "microshift-dev",
        "creationTimestamp": "2025-08-19T10:17:12Z",
        "labels": {
          "beta.kubernetes.io/arch": "amd64",
          "beta.kubernetes.io/os": "linux",
          "kubernetes.io/arch": "amd64",
          "kubernetes.io/hostname": "microshift-dev",
          "kubernetes.io/os": "linux",
          "node-role.kubernetes.io/control-plane": "",
          "node-role.kubernetes.io/master": "",
          "node-role.kubernetes.io/worker": "",
          "node.kubernetes.io/instance-type": "rhde",
          "node.openshift.io/os_id": "rhel",
          "topology.topolvm.io/node": "microshift-dev"
        }
      },
      "timestamp": "2025-08-19T10:17:01Z",
      "window": "10.019s",
      "usage": {
        "cpu": "232316598n",
        "memory": "1555732Ki"
      }
    }
  ]
}
```

## Cleanup
For deleting all resources we do not need to customize the manifests, so a simple command will do:
```bash
oc delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```
