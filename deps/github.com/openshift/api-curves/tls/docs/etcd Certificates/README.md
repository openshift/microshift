# etcd Certificates

Used to secure etcd internal communication and by apiservers to access etcd.

![PKI Graph](cert-flow.png)

- [Signing Certificate/Key Pairs](#signing-certificatekey-pairs)
    - [etcd-signer](#etcd-signer)
- [Serving Certificate/Key Pairs](#serving-certificatekey-pairs)
    - [system:etcd-peer:etcd-client](#systemetcd-peeretcd-client)
    - [system:etcd-peer:etcd-client](#systemetcd-peeretcd-client)
    - [system:etcd-peer:etcd-client](#systemetcd-peeretcd-client)
    - [system:etcd-server:etcd-client](#systemetcd-serveretcd-client)
    - [system:etcd-server:etcd-client](#systemetcd-serveretcd-client)
    - [system:etcd-server:etcd-client](#systemetcd-serveretcd-client)
- [Client Certificate/Key Pairs](#client-certificatekey-pairs)
    - [etcd](#etcd)
    - [system:etcd-peer:etcd-client](#systemetcd-peeretcd-client)
    - [system:etcd-peer:etcd-client](#systemetcd-peeretcd-client)
    - [system:etcd-peer:etcd-client](#systemetcd-peeretcd-client)
    - [system:etcd-server:etcd-client](#systemetcd-serveretcd-client)
    - [system:etcd-server:etcd-client](#systemetcd-serveretcd-client)
    - [system:etcd-server:etcd-client](#systemetcd-serveretcd-client)
- [Certificates Without Keys](#certificates-without-keys)
- [Certificate Authority Bundles](#certificate-authority-bundles)
    - [etcd-ca](#etcd-ca)

## Signing Certificate/Key Pairs


### etcd-signer
![PKI Graph](subcert-etcd-signer212144919726774987.png)

Signer for etcd to create client and serving certificates.

| Property | Value |
| ----------- | ----------- |
| Type | Signer |
| CommonName | etcd-signer |
| SerialNumber | 212144919726774987 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 10y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment<br/>- KeyUsageCertSign |
| ExtendedUsages |  |


#### etcd-signer Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-config | etcd-signer |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |



## Serving Certificate/Key Pairs


### system:etcd-peer:etcd-client
![PKI Graph](subcert-systemetcd-peeretcd-client2823097255278461792.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-peer:etcd-client |
| SerialNumber | 2823097255278461792 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-peers |
| DNS Names | - localhost<br/>- 10.0.132.68 |
| IP Addresses | - 10.0.132.68 |


#### system:etcd-peer:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-peer:etcd-client
![PKI Graph](subcert-systemetcd-peeretcd-client7040919179394354897.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-peer:etcd-client |
| SerialNumber | 7040919179394354897 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-peers |
| DNS Names | - localhost<br/>- 10.0.210.78 |
| IP Addresses | - 10.0.210.78 |


#### system:etcd-peer:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-peer:etcd-client
![PKI Graph](subcert-systemetcd-peeretcd-client5219082436276490592.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-peer:etcd-client |
| SerialNumber | 5219082436276490592 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-peers |
| DNS Names | - localhost<br/>- 10.0.183.229 |
| IP Addresses | - 10.0.183.229 |


#### system:etcd-peer:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-server:etcd-client
![PKI Graph](subcert-systemetcd-serveretcd-client1595077546154242228.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-server:etcd-client |
| SerialNumber | 1595077546154242228 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-servers |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |


#### system:etcd-server:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-server:etcd-client
![PKI Graph](subcert-systemetcd-serveretcd-client1716985330756150295.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-server:etcd-client |
| SerialNumber | 1716985330756150295 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-servers |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |


#### system:etcd-server:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-server:etcd-client
![PKI Graph](subcert-systemetcd-serveretcd-client979537366043088443.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-server:etcd-client |
| SerialNumber | 979537366043088443 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-servers |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |


#### system:etcd-server:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |


## Client Certificate/Key Pairs


### etcd
![PKI Graph](subcert-etcd3514645170127567154.png)



| Property | Value |
| ----------- | ----------- |
| Type | Client |
| CommonName | etcd |
| SerialNumber | 3514645170127567154 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 10y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth |
| Organizations (User Groups) | - etcd |


#### etcd Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-apiserver | etcd-client |
| openshift-config | etcd-client |
| openshift-etcd | etcd-client |
| openshift-etcd-operator | etcd-client |
| openshift-kube-apiserver | etcd-client |
| openshift-kube-apiserver | etcd-client-3 |
| openshift-kube-apiserver | etcd-client-4 |
| openshift-kube-apiserver | etcd-client-5 |
| openshift-kube-apiserver | etcd-client-6 |
| openshift-kube-apiserver | etcd-client-7 |
| openshift-oauth-apiserver | etcd-client |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/kube-apiserver-pod-3/secrets/etcd-client/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-apiserver-pod-3/secrets/etcd-client/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-apiserver-pod-7/secrets/etcd-client/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-apiserver-pod-7/secrets/etcd-client/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-peer:etcd-client
![PKI Graph](subcert-systemetcd-peeretcd-client2823097255278461792.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-peer:etcd-client |
| SerialNumber | 2823097255278461792 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-peers |
| DNS Names | - localhost<br/>- 10.0.132.68 |
| IP Addresses | - 10.0.132.68 |


#### system:etcd-peer:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-peer:etcd-client
![PKI Graph](subcert-systemetcd-peeretcd-client7040919179394354897.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-peer:etcd-client |
| SerialNumber | 7040919179394354897 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-peers |
| DNS Names | - localhost<br/>- 10.0.210.78 |
| IP Addresses | - 10.0.210.78 |


#### system:etcd-peer:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-peer:etcd-client
![PKI Graph](subcert-systemetcd-peeretcd-client5219082436276490592.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-peer:etcd-client |
| SerialNumber | 5219082436276490592 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-peers |
| DNS Names | - localhost<br/>- 10.0.183.229 |
| IP Addresses | - 10.0.183.229 |


#### system:etcd-peer:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-peer-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-server:etcd-client
![PKI Graph](subcert-systemetcd-serveretcd-client1595077546154242228.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-server:etcd-client |
| SerialNumber | 1595077546154242228 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-servers |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |


#### system:etcd-server:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-server:etcd-client
![PKI Graph](subcert-systemetcd-serveretcd-client1716985330756150295.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-server:etcd-client |
| SerialNumber | 1716985330756150295 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-servers |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |


#### system:etcd-server:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### system:etcd-server:etcd-client
![PKI Graph](subcert-systemetcd-serveretcd-client979537366043088443.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-server:etcd-client |
| SerialNumber | 979537366043088443 |
| Issuer CommonName | [etcd-signer](#etcd-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-servers |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |


#### system:etcd-server:etcd-client Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |


## Certificates Without Keys

These certificates are present in certificate authority bundles, but do not have keys in the cluster.
This happens when the installer bootstrap clusters with a set of certificate/key pairs that are deleted during the
installation process.

## Certificate Authority Bundles


### etcd-ca
![PKI Graph](subca-643090868.png)

CA for recognizing etcd serving, peer, and client certificates.

**Bundled Certificates**

| CommonName | Issuer CommonName | Validity | PublicKey Algorithm |
| ----------- | ----------- | ----------- | ----------- |
| [etcd-signer](#etcd-signer) | [etcd-signer](#etcd-signer) | 10y | RSA 2048 bit |

#### etcd-ca Locations
| Namespace | ConfigMap Name |
| ----------- | ----------- |
| openshift-apiserver | etcd-serving-ca |
| openshift-config | etcd-ca-bundle |
| openshift-config | etcd-serving-ca |
| openshift-etcd | etcd-ca-bundle |
| openshift-etcd | etcd-peer-client-ca |
| openshift-etcd | etcd-peer-client-ca-4 |
| openshift-etcd | etcd-peer-client-ca-5 |
| openshift-etcd | etcd-peer-client-ca-6 |
| openshift-etcd | etcd-peer-client-ca-7 |
| openshift-etcd | etcd-peer-client-ca-8 |
| openshift-etcd | etcd-serving-ca |
| openshift-etcd | etcd-serving-ca-4 |
| openshift-etcd | etcd-serving-ca-5 |
| openshift-etcd | etcd-serving-ca-6 |
| openshift-etcd | etcd-serving-ca-7 |
| openshift-etcd | etcd-serving-ca-8 |
| openshift-etcd-operator | etcd-ca-bundle |
| openshift-kube-apiserver | etcd-serving-ca |
| openshift-kube-apiserver | etcd-serving-ca-3 |
| openshift-kube-apiserver | etcd-serving-ca-4 |
| openshift-kube-apiserver | etcd-serving-ca-5 |
| openshift-kube-apiserver | etcd-serving-ca-6 |
| openshift-kube-apiserver | etcd-serving-ca-7 |
| openshift-oauth-apiserver | etcd-serving-ca |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/configmaps/etcd-peer-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/configmaps/etcd-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/configmaps/etcd-peer-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/configmaps/etcd-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/configmaps/etcd-peer-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/configmaps/etcd-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/configmaps/etcd-peer-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/configmaps/etcd-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-apiserver-pod-3/configmaps/etcd-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-apiserver-pod-7/configmaps/etcd-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |


