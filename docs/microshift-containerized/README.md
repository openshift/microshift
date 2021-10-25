---
modified: "2021-10-25T11:09:43.609+02:00"
title: Containerized
tags: container, docker, podman
layout: page
toc: true
---

## Pre-requisites

Before runnng microshift-containerized as a systemd service, ensure to update the host `crio-bridge.conf` as

```bash
{
    "cniVersion": "0.4.0",
    "name": "crio",
    "type": "bridge",
    "bridge": "cni0",
    "isGateway": true,
    "ipMasq": true,
    "hairpinMode": true,
    "ipam": {
        "type": "host-local",
        "routes": [
            { "dst": "0.0.0.0/0" }
        ],
        "ranges": [
            [{ "subnet": "10.42.0.0/24" }]
        ]
    }
}
```

## Run microshift-containerized as a systemd service

Copy microshift-containerized unit file to `/etc/systemd` and the microshift-containerized run script to `/usr/bin`

```bash
sudo cp packaging/systemd/microshift-containerized.service /etc/systemd/system/microshift-containerized.service
sudo cp packaging/systemd/microshift-containerized /usr/bin/
```

Now enable and start the service. The KUBECONFIG location will be written to `/etc/microshift-containerized/microshift-containerized.conf`.

```bash
sudo systemctl enable microshift-containerized --now
source /etc/microshift-containerized/microshift-containerized.conf
```

Verify that microshift is running.

```sh
kubectl get pods -A
```

Stop microshift-containerized service

```bash
systemctl stop microshift-containerized
```

You can check microshift-containerized via

```bash
sudo podman ps
sudo critcl ps
```

To access the cluster on the host or inside the container

### Access the cluster inside the container

Execute the following command to get into the container:

```bash
sudo podman exec -ti microshift-containerized bash
```

Inside the container, run the following to see the pods:

```bash
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
kubectl get pods -A
```

### Access the cluster on the host

#### Linux

```bash
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
kubectl get pods -A -w
```
