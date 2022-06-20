# Deploying MicroShift Behind HTTP(S) Proxy
When deploying MicroShift behind a proxy, it is necessary to configure the host OS to use this proxy for the components initiating HTTP(S) requests.

## CRI-O Container Engine
To use an HTTP(S) proxy in `CRI-O`, you need to set the `HTTP_PROXY` and `HTTPS_PROXY` environment variables.
> Optionally set the `NO_PROXY` variable to exclude a list of hosts from being proxied

Add the following settings to the `/etc/systemd/system/crio.service.d/00-proxy.conf` file.
```
[Service]
Environment=NO_PROXY="localhost,127.0.0.1"
Environment=HTTP_PROXY="http://$PROXY_USER:$PROXY_PASSWORD@$PROXY_SERVER:$PROXY_PORT/"
Environment=HTTPS_PROXY="http://$PROXY_USER:$PROXY_PASSWORD@$PROXY_SERVER:$PROXY_PORT/"
```

Reload configuration and restart the service for the changes to take effect.
```
sudo systemctl daemon-reload
sudo systemctl restart crio
```

## rpm-ostree Image/Package System
To use the HTTP(S) proxy in `rpm-ostree`, you need to set the `http_proxy` environment variable for the `rpm-ostreed` service.

Add the following setting to the `/etc/systemd/system/rpm-ostreed.service.d/00-proxy.conf` file.
```
[Service]
Environment="http_proxy=http://$PROXY_USER:$PROXY_PASSWORD@$PROXY_SERVER:$PROXY_PORT/"
```

Reload configuration and restart the service for the changes to take effect.
```
sudo systemctl daemon-reload
sudo systemctl restart rpm-ostreed.service
```

## Testing Configuration
Use the instructions in the [Install MicroShift for Edge](./devenv_rhel8.md#install-microshift-for-edge) section to configure a virtual machine running MicroShift.

### Hypervisor Settings
Log into the hypervisor host and set up an `tinyproxy` server to be used as a forward proxy.
```
podman build -t tinyproxy -f https://raw.githubusercontent.com/openshift/microshift/http_proxy/docs/podman/Containerfile.tinyproxy
podman run --rm -d --name tinyproxy -p 8443:8888 tinyproxy 
```

### Virtual Machine Settings
Log into the virtual machine and run the following commands to configure `CRI-O` and `rpm-ostree` to use a proxy.
> The hypervisor host IP address should be used to denote a proxy server and its port in the `PROXY_IP` variable set below.

```
PROXY_IP=192.168.50.103:8443

sudo mkdir -p /etc/systemd/system/crio.service.d/
sudo mkdir -p /etc/systemd/system/rpm-ostreed.service.d/

sudo cat > /etc/systemd/system/crio.service.d/00-proxy.conf <<EOF
[Service]
Environment=NO_PROXY="localhost,127.0.0.1"
Environment=HTTP_PROXY="http://${PROXY_IP}"
Environment=HTTPS_PROXY="http://${PROXY_IP}"
EOF

sudo cat > /etc/systemd/system/rpm-ostreed.service.d/00-proxy.conf <<EOF
[Service]
Environment="http_proxy=http://${PROXY_IP}"
EOF
```

Restart the services for the settings to take effect.
```
sudo systemctl daemon-reload
sudo systemctl restart crio
sudo systemctl restart rpm-ostreed.service
```
