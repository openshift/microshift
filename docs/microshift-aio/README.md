# Containerized Microshift 

## Run Microshift All-In-One as a Systemd Service

Copy microshift-aio unit file to /etc/systemd and the aio run script to /usr/bin

```bash
cp packaging/systemd/microshift-aio.service /etc/systemd/system/microshift-aio.service
cp packaging/systemd/microshift-aio /usr/bin/
```

Now enable and start the service. The KUBECONFIG location will be written to /etc/microshift-aio/microshift-aio.conf.    
If the `microshift-data` podman volume does not exist, the systemd service will create one.

```bash
systemctl enable microshift-aio --now
source /etc/microshift-aio/microshift-aio.conf
```

Verify that microshift is running.

```
kubectl get pods -A
```

Stop microshift-aio service

```bash
systemctl stop microshift-aio
```

**NOTE** Stopping microshift-aio service _does not_ remove the podman volume `microshift-data`.
A restart will use the same volume.

## Run the Image Without Systemd

First, enable the following selinux rule:

```bash
setsebool -P container_manage_cgroup true
```

Next, create a container volume:

```bash
sudo podman volume create microshift-data
```

The following example binds localhost the container volume to `/var/lib`

```bash
sudo podman run -d --rm --name microshift-aio --privileged -v /lib/modules:/lib/modules -v microshift-data:/var/lib  -p 6443:6443 microshift-aio  
```

You can access the cluster either on the host or inside the container

### Access the Cluster Inside the Container

Execute the following command to get into the container:

```bash
sudo podman exec -ti microshift-aio bash
```

Inside the container, install kubectl:

```bash
export ARCH=$(uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/")
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/${ARCH}/kubectl" && \
chmod +x ./kubectl && \
mv ./kubectl /usr/local/bin/kubectl
```

Inside the container, run the following to see the pods:

```bash
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
kubectl get pods -A
```

### Access the Cluster From the Host
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

## Build All-In-One Container Image

### Build With Locally Built Binary

```bash
make microshift-aio FROM_SOURCE="true"
```

### Build With Latest Released Binary Download

```bash
make microshfit-aio
```

## Limitation

These instructions are tested on Linux, Mac, and Windows. 
On MacOS, running containerized Microshift as non-root is not supported on MacOS.
