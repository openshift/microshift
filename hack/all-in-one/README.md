# Containerized Microshift 

## Build Container Image
First copy microshift binary to this directory, then build the container image:
```bash
podman build -t ushift .
```

## Run the Image

First, enable the following selinux rule:
```bash
setsebool -P container_manage_cgroup true
```
Next, create a container volume:
```bash
podman volume create ushift-vol
```
The following example binds localhost the container volume to `/var/lib`

```bash
 podman run -d --rm --name ushift --privileged -v /lib/modules:/lib/modules -v ushift-vol:/var/lib --hostname ushift -p 6443:6443 ushift  
```

Then you can access the cluster either on the host or inside the container

### Access the cluster inside the container
Execute the following command to get into the container:
```bash
podman exec -ti ushift bash
```
Inside the container, run the following to see the pods:
```bash
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
kubectl get pods -A
```

### Access the cluster on the host
#### Linux
```bash
export KUBECONFIG=$(podman volume inspect ushift-vol --format "{{.Mountpoint}}")/microshift/resources/kubeadmin/kubeconfig
kubectl get pods -A -w
```
#### MacOS
```bash
docker cp ushift:/var/lib/microshift/resources/kubeadmin/kubeconfig ./kubeconfig
kubectl get pods -A -w --kubeconfig ./kubeconfig
```
#### Windows
```bash
docker.exe cp ushift:/var/lib/microshift/resources/kubeadmin/kubeconfig .\kubeconfig
kubectl.exe get pods -A -w --kubeconfig .\kubeconfig
```
## Limitation

These instructions are tested on Linux, Mac, and Windows. 
On MacOS, running containerized Microshift as non-root is not supported on MacOS. 
