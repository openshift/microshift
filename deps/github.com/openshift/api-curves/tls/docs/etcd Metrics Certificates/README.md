# etcd Metrics Certificates

Used to access etcd metrics using mTLS.

![PKI Graph](cert-flow.png)

- [Signing Certificate/Key Pairs](#signing-certificatekey-pairs)
    - [etcd-metric-signer](#etcd-metric-signer)
- [Serving Certificate/Key Pairs](#serving-certificatekey-pairs)
    - [etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal](#etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal)
    - [etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal](#etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal)
    - [etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal](#etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal)
- [Client Certificate/Key Pairs](#client-certificatekey-pairs)
    - [etcd-metric](#etcd-metric)
    - [etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal](#etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal)
    - [etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal](#etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal)
    - [etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal](#etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal)
- [Certificates Without Keys](#certificates-without-keys)
- [Certificate Authority Bundles](#certificate-authority-bundles)
    - [etcd-metrics-ca](#etcd-metrics-ca)

## Signing Certificate/Key Pairs


### etcd-metric-signer
![PKI Graph](subcert-etcd-metric-signer4218002264880163000.png)



| Property | Value |
| ----------- | ----------- |
| Type | Signer |
| CommonName | etcd-metric-signer |
| SerialNumber | 4218002264880163000 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 10y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment<br/>- KeyUsageCertSign |
| ExtendedUsages |  |


#### etcd-metric-signer Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-config | etcd-metric-signer |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |



## Serving Certificate/Key Pairs


### etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal
![PKI Graph](subcert-systemetcd-metricetcd-client1768016573087407501.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-metric:etcd-client |
| SerialNumber | 1768016573087407501 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-metrics |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |


#### etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal
![PKI Graph](subcert-systemetcd-metricetcd-client45102667796783726.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-metric:etcd-client |
| SerialNumber | 45102667796783726 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-metrics |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |


#### etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal
![PKI Graph](subcert-systemetcd-metricetcd-client564640857130170731.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-metric:etcd-client |
| SerialNumber | 564640857130170731 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-metrics |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |


#### etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |


## Client Certificate/Key Pairs


### etcd-metric
![PKI Graph](subcert-etcd-metric2163297920510713389.png)



| Property | Value |
| ----------- | ----------- |
| Type | Client |
| CommonName | etcd-metric |
| SerialNumber | 2163297920510713389 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 10y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth |
| Organizations (User Groups) | - etcd-metric |


#### etcd-metric Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-config | etcd-metric-client |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal
![PKI Graph](subcert-systemetcd-metricetcd-client1768016573087407501.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-metric:etcd-client |
| SerialNumber | 1768016573087407501 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-metrics |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.183.229<br/>- 127.0.0.1<br/>- ::1 |


#### etcd-metrics-for-master-ip-10-0-183-229.eu-central-1.compute.internal Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-183-229.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal
![PKI Graph](subcert-systemetcd-metricetcd-client45102667796783726.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-metric:etcd-client |
| SerialNumber | 45102667796783726 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-metrics |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.132.68<br/>- 127.0.0.1<br/>- ::1 |


#### etcd-metrics-for-master-ip-10-0-132-68.eu-central-1.compute.internal Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-132-68.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal
![PKI Graph](subcert-systemetcd-metricetcd-client564640857130170731.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving,Client |
| CommonName | system:etcd-metric:etcd-client |
| SerialNumber | 564640857130170731 |
| Issuer CommonName | [etcd-metric-signer](#etcd-metric-signer) |
| Validity | 3y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageClientAuth<br/>- ExtKeyUsageServerAuth |
| Organizations (User Groups) | - system:etcd-metrics |
| DNS Names | - etcd.kube-system.svc<br/>- etcd.kube-system.svc.cluster.local<br/>- etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local<br/>- localhost<br/>- ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |
| IP Addresses | - ::1<br/>- 10.0.210.78<br/>- 127.0.0.1<br/>- ::1 |


#### etcd-metrics-for-master-ip-10-0-210-78.eu-central-1.compute.internal Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/secrets/etcd-all-certs/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.crt/etcd-serving-metrics-ip-10-0-210-78.eu-central-1.compute.internal.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |


## Certificates Without Keys

These certificates are present in certificate authority bundles, but do not have keys in the cluster.
This happens when the installer bootstrap clusters with a set of certificate/key pairs that are deleted during the
installation process.

## Certificate Authority Bundles


### etcd-metrics-ca
![PKI Graph](subca-2806240455.png)

CA used to recognize etcd metrics serving and client certificates.

**Bundled Certificates**

| CommonName | Issuer CommonName | Validity | PublicKey Algorithm |
| ----------- | ----------- | ----------- | ----------- |
| [etcd-metric-signer](#etcd-metric-signer) | [etcd-metric-signer](#etcd-metric-signer) | 10y | RSA 2048 bit |

#### etcd-metrics-ca Locations
| Namespace | ConfigMap Name |
| ----------- | ----------- |
| openshift-config | etcd-metric-serving-ca |
| openshift-etcd | etcd-metrics-proxy-client-ca |
| openshift-etcd | etcd-metrics-proxy-client-ca-4 |
| openshift-etcd | etcd-metrics-proxy-client-ca-5 |
| openshift-etcd | etcd-metrics-proxy-client-ca-6 |
| openshift-etcd | etcd-metrics-proxy-client-ca-7 |
| openshift-etcd | etcd-metrics-proxy-client-ca-8 |
| openshift-etcd | etcd-metrics-proxy-serving-ca |
| openshift-etcd | etcd-metrics-proxy-serving-ca-4 |
| openshift-etcd | etcd-metrics-proxy-serving-ca-5 |
| openshift-etcd | etcd-metrics-proxy-serving-ca-6 |
| openshift-etcd | etcd-metrics-proxy-serving-ca-7 |
| openshift-etcd | etcd-metrics-proxy-serving-ca-8 |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/etcd-certs/configmaps/etcd-metrics-proxy-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-certs/configmaps/etcd-metrics-proxy-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/configmaps/etcd-metrics-proxy-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-4/configmaps/etcd-metrics-proxy-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/configmaps/etcd-metrics-proxy-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-7/configmaps/etcd-metrics-proxy-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/configmaps/etcd-metrics-proxy-client-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/etcd-pod-8/configmaps/etcd-metrics-proxy-serving-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |


