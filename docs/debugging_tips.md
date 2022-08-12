# Debugging Tips

## Checking the MicroShift Version

From the command line, use `microshift version` to check the version
information.

```bash
$ microshift version
MicroShift Version: 4.10.0-0.microshift-e6980e25
Base OCP Version: 4.10.18
```

Through the API, access the `kube-public/microshift-version` ConfigMap
to retrieve the same information.

```bash
$ oc get configmap -n kube-public microshift-version -o yaml
apiVersion: v1
data:
  major: "4"
  minor: "10"
  version: 4.10.0-0.microshift-e6980e25
kind: ConfigMap
metadata:
  creationTimestamp: "2022-08-08T21:06:11Z"
  name: microshift-version
  namespace: kube-public
```
