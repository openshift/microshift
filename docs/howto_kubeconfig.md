# Kubeconfig management
Every time MicroShift starts it checks if the set of kubeconfig files for local and remote management of the API server exist and include the latest CAs. These are tied to the certificates for the API server and are also impacted by MicroShift's configuration.

## Configuration
MicroShift allows configuring additional names and IP addresses on top of the defaults, which are `<node.hostnameOverride>`, `<node.nodeIP>` and `api.<dns.baseDomain>`. A configuration example:
```yaml
dns:
  baseDomain: example.com
node:
  hostnameOverride: "microshift-rhel9"
  nodeIP: 10.0.0.1
apiServer:
  subjectAltNames:
  - alt-name-1
  - 1.2.3.4
```
All these parameters are included as common names (CN) and subject alternative names (SAN) in the external serving certificates for API server.

## Kubeconfig files
Upon starting, MicroShift generates a set of kubeconfig files under `/var/lib/microshift/resources/kubeadmin` for different network access types. From the example above we would have:
```bash
/var/lib/microshift/resources/kubeadmin/
├── kubeconfig
├── alt-name-1
│   └── kubeconfig
├── 1.2.3.4
│   └── kubeconfig
└── microshift-rhel9
    └── kubeconfig
```
Sections below will assume this configuration is loaded into the system.

### Local access
When connecting to the API server locally, a certificate like this one, with `localhost` as the only SAN, is served for validation:
```bash
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 2 (0x2)
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN = kube-apiserver-localhost-signer
        Validity
            Not Before: Mar 17 08:44:07 2023 GMT
            Not After : Mar 16 08:44:08 2024 GMT
        Subject: CN = localhost
...
            X509v3 Subject Alternative Name: 
                DNS:localhost
```
This means we need to use `localhost` to connect to it, and have our kubeconfig use the same CA that signed the certificate above.

The local access kubeconfig is always generated at `/var/lib/microshift/resources/kubeadmin/kubeconfig`. This one is not impacted/driven by configuration parameters. It allows access to API server using `localhost` and uses the `localhost` internal CA:
```yaml
clusters:
- cluster:
    certificate-authority-data: <base64 CA>
    server: https://localhost:6443
```
This file cannot be used outside of the MicroShift host as its CA only validates certificates with `localhost` network access type.

### External access
When connecting to the API server from an external source, a certificate with all of the alternative names in the SAN field is served for validation, as in the following example:
```bash
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 2 (0x2)
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN = kube-apiserver-external-signer
        Validity
            Not Before: Mar 17 08:44:07 2023 GMT
            Not After : Mar 16 08:44:08 2024 GMT
        Subject: CN = 10.0.0.1
...
            X509v3 Subject Alternative Name: 
                DNS:api.example.com, DNS:microshift-rhel9, DNS:alt-name-1, DNS:10.0.0.1, DNS:1.2.3.4, IP Address:10.0.0.1, IP Address:1.2.3.4
```
As with the local access use case, we need the kubeconfig to include one of those names as `server` and also use the same CA for validation.

MicroShift will generate a default kubeconfig for external access using the hostname, and an additional one per entry in `apiServer.subjectAltNames`.

`/var/lib/microshift/resources/kubeadmin/<hostname>/kubeconfig` uses the machine's hostname (or `node.hostnameOverride` if that option is set) to reach API server. Its CA is able to validate certificates when accessed externally.
```yaml
clusters:
- cluster:
    certificate-authority-data: <base64 CA>
    server: https://microshift-rhel9:6443
```
> As we have seen in the served certificate we can swap the hostname in `server` stanza to use the node IP instead.

`/var/lib/microshift/resources/kubeadmin/alt-name-1/kubeconfig` and `/var/lib/microshift/resources/kubeadmin/1.2.3.4/kubeconfig` come from the `apiServer.subjectAltNames` configuration values. These are useful to refer to the MicroShift's host using alternative host names and/or IPs (i.e. using a different DNS name or an external IP).
```yaml
clusters:
- cluster:
    certificate-authority-data: <base64 CA>
    server: https://alt-name-1:6443
---
clusters:
- cluster:
    certificate-authority-data: <base64 CA>
    server: https://1.2.3.4:6443
```

All external access kubeconfig files can be extracted from the MicroShift's host to be used from elsewhere, provided there is IP connectivity when in use.
