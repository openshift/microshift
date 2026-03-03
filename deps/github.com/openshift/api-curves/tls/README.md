This content was generated from a live cluster.
It may change over time and is not guaranteed to be stable.
This is a useful starting point for understanding the cert chains in openshift used to secure kubernetes.

1. Build an image to collect the certs, keys, and ca bundles from the host.
   1. Something like `docker build pkg/cmd/locateinclustercerts/ -t docker.io/$USER/locateinclustercerts:latest -f Dockerfile`
   2. Push to dockerhub
2. Gather data.
   1. `oc adm inspect clusteroperators` -- this will gather all the in-cluster certificates and ca bundles
   2. run pods on the masters.  Something like:

```bash
NODE=$(kubectl get nodes | grep master | head -n1 | awk '{ print $1 }') \
    IMAGE=$(podman image list --sort created | grep locateinclustercerts | awk '{ print $1 ":" $2 }') \
    oc debug --image=$IMAGE node/$NODE
```

   3. in those pods, run `master-cert-collection.sh` to collect the data from the host.  Leave the pod running after completion.
   4. pull the on-disk data locally. Something like:

```bash
POD=$(kubectl get pods -A | rg debug | awk '{ print $2 }') \
   oc rsync $POD:/must-gather .
```


    5. Be sure dot is installed locally
    6. To produce the doc, use something like: 


```bash
INSPECT_DIR=$(find . -name 'inspect.local.*') \
    kubectl-dev_tool certs locate-incluster-certs --local -f $INSPECT_DIR --additional-input-dir ./must-gather -odoc
```

