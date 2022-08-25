# Firewall Configuration
MicroShift does not require a firewall for normal operation. However, it is recommended enabling a firewall to avoid undesired access to the MicroShift API. If a firewall is enabled, it should be configured according to the requirements in this document.

It is mandatory to allow MicroShift pods the access to the internal CoreDNS and API servers.
> Users may choose to change the pod IP range to be different from the default `10.42.0.0/16` setting. This must be reflected in the firewall configuration.

|IP Range      |Description|
|:-------------|:----------|
|10.42.0.0/16  |Host network pod access to CoreDNS and MicroShift API |
|169.254.169.1 |Host network pod access to MicroShift API Server      |

The following ports are optional and they should be considered for MicroShift if a firewall is enabled.

|Port(s)    |Protocol(s)|Description|
|:----------|:----------|:----------|
|80         |TCP        |HTTP port used to serve applications through the OpenShift router |
|443        |TCP        |HTTPS port used to serve applications through the OpenShift router |
|5353       |UDP        |mDNS service to respond for OpenShift route mDNS hosts |
|30000-32767|TCP/UDP    |Port range reserved for NodePort type of services, can be used to expose applications on the LAN |
|6443       |TCP        |HTTPS API port for the MicroShift API |

## Firewalld
The following commands can be used for enabling `firewalld` and opening all the above mentioned source IP addresses and ports.
> Use the appropriate pod IP range if it is different from the default `10.42.0.0/16` setting.

```bash
sudo dnf install -y firewalld
sudo systemctl enable firewalld --now
# Mandatory settings
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16 
sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
sudo firewall-cmd --reload
# Optional settings
sudo firewall-cmd --permanent --zone=public --add-port=80/tcp
sudo firewall-cmd --permanent --zone=public --add-port=443/tcp
sudo firewall-cmd --permanent --zone=public --add-port=5353/udp
sudo firewall-cmd --permanent --zone=public --add-port=30000-32767/tcp
sudo firewall-cmd --permanent --zone=public --add-port=30000-32767/udp
sudo firewall-cmd --permanent --zone=public --add-port=6443/tcp
sudo firewall-cmd --reload
```
