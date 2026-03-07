# Service Serving Certificates

Used to secure inter-service communication on the local cluster.

![PKI Graph](cert-flow.png)

- [Signing Certificate/Key Pairs](#signing-certificatekey-pairs)
    - [service-serving-signer](#service-serving-signer)
- [Serving Certificate/Key Pairs](#serving-certificatekey-pairs)
    - [*.cluster-monitoring-operator.openshift-monitoring.svc](#*.cluster-monitoring-operator.openshift-monitoring.svc)
    - [*.image-registry-operator.openshift-image-registry.svc](#*.image-registry-operator.openshift-image-registry.svc)
    - [*.kube-state-metrics.openshift-monitoring.svc](#*.kube-state-metrics.openshift-monitoring.svc)
    - [*.machine-approver.openshift-cluster-machine-approver.svc](#*.machine-approver.openshift-cluster-machine-approver.svc)
    - [*.metrics.openshift-cluster-samples-operator.svc](#*.metrics.openshift-cluster-samples-operator.svc)
    - [*.metrics.openshift-network-operator.svc](#*.metrics.openshift-network-operator.svc)
    - [*.network-metrics-service.openshift-multus.svc](#*.network-metrics-service.openshift-multus.svc)
    - [*.node-exporter.openshift-monitoring.svc](#*.node-exporter.openshift-monitoring.svc)
    - [*.node-tuning-operator.openshift-cluster-node-tuning-operator.svc](#*.node-tuning-operator.openshift-cluster-node-tuning-operator.svc)
    - [*.openshift-state-metrics.openshift-monitoring.svc](#*.openshift-state-metrics.openshift-monitoring.svc)
    - [*.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc](#*.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc)
    - [*.prometheus-operator.openshift-monitoring.svc](#*.prometheus-operator.openshift-monitoring.svc)
    - [*.sdn-controller.openshift-sdn.svc](#*.sdn-controller.openshift-sdn.svc)
    - [*.sdn.openshift-sdn.svc](#*.sdn.openshift-sdn.svc)
    - [*.telemeter-client.openshift-monitoring.svc](#*.telemeter-client.openshift-monitoring.svc)
    - [alertmanager-main.openshift-monitoring.svc](#alertmanager-main.openshift-monitoring.svc)
    - [api.openshift-apiserver.svc](#api.openshift-apiserver.svc)
    - [api.openshift-oauth-apiserver.svc](#api.openshift-oauth-apiserver.svc)
    - [aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc](#aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc)
    - [catalog-operator-metrics.openshift-operator-lifecycle-manager.svc](#catalog-operator-metrics.openshift-operator-lifecycle-manager.svc)
    - [cco-metrics.openshift-cloud-credential-operator.svc](#cco-metrics.openshift-cloud-credential-operator.svc)
    - [cluster-autoscaler-operator.openshift-machine-api.svc](#cluster-autoscaler-operator.openshift-machine-api.svc)
    - [cluster-baremetal-operator-service.openshift-machine-api.svc](#cluster-baremetal-operator-service.openshift-machine-api.svc)
    - [cluster-baremetal-webhook-service.openshift-machine-api.svc](#cluster-baremetal-webhook-service.openshift-machine-api.svc)
    - [cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc](#cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc)
    - [console.openshift-console.svc](#console.openshift-console.svc)
    - [controller-manager.openshift-controller-manager.svc](#controller-manager.openshift-controller-manager.svc)
    - [csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc](#csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc)
    - [csi-snapshot-webhook.openshift-cluster-storage-operator.svc](#csi-snapshot-webhook.openshift-cluster-storage-operator.svc)
    - [dns-default.openshift-dns.svc](#dns-default.openshift-dns.svc)
    - [etcd.openshift-etcd.svc](#etcd.openshift-etcd.svc)
    - [image-registry.openshift-image-registry.svc](#image-registry.openshift-image-registry.svc)
    - [kube-controller-manager.openshift-kube-controller-manager.svc](#kube-controller-manager.openshift-kube-controller-manager.svc)
    - [machine-api-controllers.openshift-machine-api.svc](#machine-api-controllers.openshift-machine-api.svc)
    - [machine-api-operator-webhook.openshift-machine-api.svc](#machine-api-operator-webhook.openshift-machine-api.svc)
    - [machine-api-operator.openshift-machine-api.svc](#machine-api-operator.openshift-machine-api.svc)
    - [machine-config-controller.openshift-machine-config-operator.svc](#machine-config-controller.openshift-machine-config-operator.svc)
    - [machine-config-daemon.openshift-machine-config-operator.svc](#machine-config-daemon.openshift-machine-config-operator.svc)
    - [marketplace-operator-metrics.openshift-marketplace.svc](#marketplace-operator-metrics.openshift-marketplace.svc)
    - [metrics.openshift-apiserver-operator.svc](#metrics.openshift-apiserver-operator.svc)
    - [metrics.openshift-authentication-operator.svc](#metrics.openshift-authentication-operator.svc)
    - [metrics.openshift-config-operator.svc](#metrics.openshift-config-operator.svc)
    - [metrics.openshift-console-operator.svc](#metrics.openshift-console-operator.svc)
    - [metrics.openshift-controller-manager-operator.svc](#metrics.openshift-controller-manager-operator.svc)
    - [metrics.openshift-dns-operator.svc](#metrics.openshift-dns-operator.svc)
    - [metrics.openshift-etcd-operator.svc](#metrics.openshift-etcd-operator.svc)
    - [metrics.openshift-ingress-operator.svc](#metrics.openshift-ingress-operator.svc)
    - [metrics.openshift-insights.svc](#metrics.openshift-insights.svc)
    - [metrics.openshift-kube-apiserver-operator.svc](#metrics.openshift-kube-apiserver-operator.svc)
    - [metrics.openshift-kube-controller-manager-operator.svc](#metrics.openshift-kube-controller-manager-operator.svc)
    - [metrics.openshift-kube-scheduler-operator.svc](#metrics.openshift-kube-scheduler-operator.svc)
    - [metrics.openshift-kube-storage-version-migrator-operator.svc](#metrics.openshift-kube-storage-version-migrator-operator.svc)
    - [metrics.openshift-service-ca-operator.svc](#metrics.openshift-service-ca-operator.svc)
    - [multus-admission-controller.openshift-multus.svc](#multus-admission-controller.openshift-multus.svc)
    - [oauth-openshift.openshift-authentication.svc](#oauth-openshift.openshift-authentication.svc)
    - [olm-operator-metrics.openshift-operator-lifecycle-manager.svc](#olm-operator-metrics.openshift-operator-lifecycle-manager.svc)
    - [performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc](#performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc)
    - [pod-identity-webhook.openshift-cloud-credential-operator.svc](#pod-identity-webhook.openshift-cloud-credential-operator.svc)
    - [prometheus-adapter.openshift-monitoring.svc](#prometheus-adapter.openshift-monitoring.svc)
    - [prometheus-k8s.openshift-monitoring.svc](#prometheus-k8s.openshift-monitoring.svc)
    - [prometheus-operator-admission-webhook.openshift-monitoring.svc](#prometheus-operator-admission-webhook.openshift-monitoring.svc)
    - [router-internal-default.openshift-ingress.svc](#router-internal-default.openshift-ingress.svc)
    - [scheduler.openshift-kube-scheduler.svc](#scheduler.openshift-kube-scheduler.svc)
    - [thanos-querier.openshift-monitoring.svc](#thanos-querier.openshift-monitoring.svc)
- [Client Certificate/Key Pairs](#client-certificatekey-pairs)
- [Certificates Without Keys](#certificates-without-keys)
- [Certificate Authority Bundles](#certificate-authority-bundles)
    - [service-ca](#service-ca)

## Signing Certificate/Key Pairs


### service-serving-signer
![PKI Graph](subcert-openshift-service-serving-signer1660235194316321027491433813.png)

Signer used by service-ca to sign serving certificates for internal service DNS names.

| Property | Value |
| ----------- | ----------- |
| Type | Signer |
| CommonName | openshift-service-serving-signer@1660235194 |
| SerialNumber | 316321027491433813 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y60d |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment<br/>- KeyUsageCertSign |
| ExtendedUsages |  |


#### service-serving-signer Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-service-ca | signing-key |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |



## Serving Certificate/Key Pairs


### *.cluster-monitoring-operator.openshift-monitoring.svc
![PKI Graph](subcert-*.cluster-monitoring-operator.openshift-monitoring.svc5408701497193776530.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.cluster-monitoring-operator.openshift-monitoring.svc |
| SerialNumber | 5408701497193776530 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.cluster-monitoring-operator.openshift-monitoring.svc<br/>- *.cluster-monitoring-operator.openshift-monitoring.svc.cluster.local<br/>- cluster-monitoring-operator.openshift-monitoring.svc<br/>- cluster-monitoring-operator.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.cluster-monitoring-operator.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | cluster-monitoring-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.image-registry-operator.openshift-image-registry.svc
![PKI Graph](subcert-*.image-registry-operator.openshift-image-registry.svc4675784410318447775.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.image-registry-operator.openshift-image-registry.svc |
| SerialNumber | 4675784410318447775 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.image-registry-operator.openshift-image-registry.svc<br/>- *.image-registry-operator.openshift-image-registry.svc.cluster.local<br/>- image-registry-operator.openshift-image-registry.svc<br/>- image-registry-operator.openshift-image-registry.svc.cluster.local |
| IP Addresses |  |


#### *.image-registry-operator.openshift-image-registry.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-image-registry | image-registry-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.kube-state-metrics.openshift-monitoring.svc
![PKI Graph](subcert-*.kube-state-metrics.openshift-monitoring.svc149350220636519763.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.kube-state-metrics.openshift-monitoring.svc |
| SerialNumber | 149350220636519763 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.kube-state-metrics.openshift-monitoring.svc<br/>- *.kube-state-metrics.openshift-monitoring.svc.cluster.local<br/>- kube-state-metrics.openshift-monitoring.svc<br/>- kube-state-metrics.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.kube-state-metrics.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | kube-state-metrics-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.machine-approver.openshift-cluster-machine-approver.svc
![PKI Graph](subcert-*.machine-approver.openshift-cluster-machine-approver.svc6308567208770339707.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.machine-approver.openshift-cluster-machine-approver.svc |
| SerialNumber | 6308567208770339707 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.machine-approver.openshift-cluster-machine-approver.svc<br/>- *.machine-approver.openshift-cluster-machine-approver.svc.cluster.local<br/>- machine-approver.openshift-cluster-machine-approver.svc<br/>- machine-approver.openshift-cluster-machine-approver.svc.cluster.local |
| IP Addresses |  |


#### *.machine-approver.openshift-cluster-machine-approver.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-machine-approver | machine-approver-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.metrics.openshift-cluster-samples-operator.svc
![PKI Graph](subcert-*.metrics.openshift-cluster-samples-operator.svc376205288372847070.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.metrics.openshift-cluster-samples-operator.svc |
| SerialNumber | 376205288372847070 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.metrics.openshift-cluster-samples-operator.svc<br/>- *.metrics.openshift-cluster-samples-operator.svc.cluster.local<br/>- metrics.openshift-cluster-samples-operator.svc<br/>- metrics.openshift-cluster-samples-operator.svc.cluster.local |
| IP Addresses |  |


#### *.metrics.openshift-cluster-samples-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-samples-operator | samples-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.metrics.openshift-network-operator.svc
![PKI Graph](subcert-*.metrics.openshift-network-operator.svc6421516406206552641.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.metrics.openshift-network-operator.svc |
| SerialNumber | 6421516406206552641 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.metrics.openshift-network-operator.svc<br/>- *.metrics.openshift-network-operator.svc.cluster.local<br/>- metrics.openshift-network-operator.svc<br/>- metrics.openshift-network-operator.svc.cluster.local |
| IP Addresses |  |


#### *.metrics.openshift-network-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-network-operator | metrics-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.network-metrics-service.openshift-multus.svc
![PKI Graph](subcert-*.network-metrics-service.openshift-multus.svc1150444052307414109.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.network-metrics-service.openshift-multus.svc |
| SerialNumber | 1150444052307414109 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.network-metrics-service.openshift-multus.svc<br/>- *.network-metrics-service.openshift-multus.svc.cluster.local<br/>- network-metrics-service.openshift-multus.svc<br/>- network-metrics-service.openshift-multus.svc.cluster.local |
| IP Addresses |  |


#### *.network-metrics-service.openshift-multus.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-multus | metrics-daemon-secret |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.node-exporter.openshift-monitoring.svc
![PKI Graph](subcert-*.node-exporter.openshift-monitoring.svc5820670729257377840.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.node-exporter.openshift-monitoring.svc |
| SerialNumber | 5820670729257377840 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.node-exporter.openshift-monitoring.svc<br/>- *.node-exporter.openshift-monitoring.svc.cluster.local<br/>- node-exporter.openshift-monitoring.svc<br/>- node-exporter.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.node-exporter.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | node-exporter-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.node-tuning-operator.openshift-cluster-node-tuning-operator.svc
![PKI Graph](subcert-*.node-tuning-operator.openshift-cluster-node-tuning-operator.svc1966650171226976553.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.node-tuning-operator.openshift-cluster-node-tuning-operator.svc |
| SerialNumber | 1966650171226976553 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.node-tuning-operator.openshift-cluster-node-tuning-operator.svc<br/>- *.node-tuning-operator.openshift-cluster-node-tuning-operator.svc.cluster.local<br/>- node-tuning-operator.openshift-cluster-node-tuning-operator.svc<br/>- node-tuning-operator.openshift-cluster-node-tuning-operator.svc.cluster.local |
| IP Addresses |  |


#### *.node-tuning-operator.openshift-cluster-node-tuning-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-node-tuning-operator | node-tuning-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.openshift-state-metrics.openshift-monitoring.svc
![PKI Graph](subcert-*.openshift-state-metrics.openshift-monitoring.svc4684927540193099053.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.openshift-state-metrics.openshift-monitoring.svc |
| SerialNumber | 4684927540193099053 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.openshift-state-metrics.openshift-monitoring.svc<br/>- *.openshift-state-metrics.openshift-monitoring.svc.cluster.local<br/>- openshift-state-metrics.openshift-monitoring.svc<br/>- openshift-state-metrics.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.openshift-state-metrics.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | openshift-state-metrics-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc
![PKI Graph](subcert-*.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc9043370268645417073.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc |
| SerialNumber | 9043370268645417073 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc<br/>- *.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc.cluster.local<br/>- prometheus-k8s-thanos-sidecar.openshift-monitoring.svc<br/>- prometheus-k8s-thanos-sidecar.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.prometheus-k8s-thanos-sidecar.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | prometheus-k8s-thanos-sidecar-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.prometheus-operator.openshift-monitoring.svc
![PKI Graph](subcert-*.prometheus-operator.openshift-monitoring.svc8610502219646032799.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.prometheus-operator.openshift-monitoring.svc |
| SerialNumber | 8610502219646032799 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.prometheus-operator.openshift-monitoring.svc<br/>- *.prometheus-operator.openshift-monitoring.svc.cluster.local<br/>- prometheus-operator.openshift-monitoring.svc<br/>- prometheus-operator.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.prometheus-operator.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | prometheus-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.sdn-controller.openshift-sdn.svc
![PKI Graph](subcert-*.sdn-controller.openshift-sdn.svc7537155398406497139.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.sdn-controller.openshift-sdn.svc |
| SerialNumber | 7537155398406497139 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.sdn-controller.openshift-sdn.svc<br/>- *.sdn-controller.openshift-sdn.svc.cluster.local<br/>- sdn-controller.openshift-sdn.svc<br/>- sdn-controller.openshift-sdn.svc.cluster.local |
| IP Addresses |  |


#### *.sdn-controller.openshift-sdn.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-sdn | sdn-controller-metrics-certs |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.sdn.openshift-sdn.svc
![PKI Graph](subcert-*.sdn.openshift-sdn.svc2028218742012863699.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.sdn.openshift-sdn.svc |
| SerialNumber | 2028218742012863699 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.sdn.openshift-sdn.svc<br/>- *.sdn.openshift-sdn.svc.cluster.local<br/>- sdn.openshift-sdn.svc<br/>- sdn.openshift-sdn.svc.cluster.local |
| IP Addresses |  |


#### *.sdn.openshift-sdn.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-sdn | sdn-metrics-certs |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### *.telemeter-client.openshift-monitoring.svc
![PKI Graph](subcert-*.telemeter-client.openshift-monitoring.svc6316470005412686670.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | *.telemeter-client.openshift-monitoring.svc |
| SerialNumber | 6316470005412686670 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - *.telemeter-client.openshift-monitoring.svc<br/>- *.telemeter-client.openshift-monitoring.svc.cluster.local<br/>- telemeter-client.openshift-monitoring.svc<br/>- telemeter-client.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### *.telemeter-client.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | telemeter-client-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### alertmanager-main.openshift-monitoring.svc
![PKI Graph](subcert-alertmanager-main.openshift-monitoring.svc6223384734159182299.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | alertmanager-main.openshift-monitoring.svc |
| SerialNumber | 6223384734159182299 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - alertmanager-main.openshift-monitoring.svc<br/>- alertmanager-main.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### alertmanager-main.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | alertmanager-main-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### api.openshift-apiserver.svc
![PKI Graph](subcert-api.openshift-apiserver.svc2457272755999329987.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | api.openshift-apiserver.svc |
| SerialNumber | 2457272755999329987 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - api.openshift-apiserver.svc<br/>- api.openshift-apiserver.svc.cluster.local |
| IP Addresses |  |


#### api.openshift-apiserver.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-apiserver | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### api.openshift-oauth-apiserver.svc
![PKI Graph](subcert-api.openshift-oauth-apiserver.svc2801162004244502483.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | api.openshift-oauth-apiserver.svc |
| SerialNumber | 2801162004244502483 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - api.openshift-oauth-apiserver.svc<br/>- api.openshift-oauth-apiserver.svc.cluster.local |
| IP Addresses |  |


#### api.openshift-oauth-apiserver.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-oauth-apiserver | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc
![PKI Graph](subcert-aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc3785796387182627089.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc |
| SerialNumber | 3785796387182627089 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc<br/>- aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc.cluster.local |
| IP Addresses |  |


#### aws-ebs-csi-driver-controller-metrics.openshift-cluster-csi-drivers.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-csi-drivers | aws-ebs-csi-driver-controller-metrics-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### catalog-operator-metrics.openshift-operator-lifecycle-manager.svc
![PKI Graph](subcert-catalog-operator-metrics.openshift-operator-lifecycle-manager.svc8354235402282530330.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | catalog-operator-metrics.openshift-operator-lifecycle-manager.svc |
| SerialNumber | 8354235402282530330 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - catalog-operator-metrics.openshift-operator-lifecycle-manager.svc<br/>- catalog-operator-metrics.openshift-operator-lifecycle-manager.svc.cluster.local |
| IP Addresses |  |


#### catalog-operator-metrics.openshift-operator-lifecycle-manager.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-operator-lifecycle-manager | catalog-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### cco-metrics.openshift-cloud-credential-operator.svc
![PKI Graph](subcert-cco-metrics.openshift-cloud-credential-operator.svc761473066828534832.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | cco-metrics.openshift-cloud-credential-operator.svc |
| SerialNumber | 761473066828534832 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - cco-metrics.openshift-cloud-credential-operator.svc<br/>- cco-metrics.openshift-cloud-credential-operator.svc.cluster.local |
| IP Addresses |  |


#### cco-metrics.openshift-cloud-credential-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cloud-credential-operator | cloud-credential-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### cluster-autoscaler-operator.openshift-machine-api.svc
![PKI Graph](subcert-cluster-autoscaler-operator.openshift-machine-api.svc4413431980141135024.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | cluster-autoscaler-operator.openshift-machine-api.svc |
| SerialNumber | 4413431980141135024 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - cluster-autoscaler-operator.openshift-machine-api.svc<br/>- cluster-autoscaler-operator.openshift-machine-api.svc.cluster.local |
| IP Addresses |  |


#### cluster-autoscaler-operator.openshift-machine-api.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-api | cluster-autoscaler-operator-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### cluster-baremetal-operator-service.openshift-machine-api.svc
![PKI Graph](subcert-cluster-baremetal-operator-service.openshift-machine-api.svc6219679491641346115.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | cluster-baremetal-operator-service.openshift-machine-api.svc |
| SerialNumber | 6219679491641346115 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - cluster-baremetal-operator-service.openshift-machine-api.svc<br/>- cluster-baremetal-operator-service.openshift-machine-api.svc.cluster.local |
| IP Addresses |  |


#### cluster-baremetal-operator-service.openshift-machine-api.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-api | cluster-baremetal-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### cluster-baremetal-webhook-service.openshift-machine-api.svc
![PKI Graph](subcert-cluster-baremetal-webhook-service.openshift-machine-api.svc7893619671345938179.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | cluster-baremetal-webhook-service.openshift-machine-api.svc |
| SerialNumber | 7893619671345938179 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - cluster-baremetal-webhook-service.openshift-machine-api.svc<br/>- cluster-baremetal-webhook-service.openshift-machine-api.svc.cluster.local |
| IP Addresses |  |


#### cluster-baremetal-webhook-service.openshift-machine-api.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-api | cluster-baremetal-webhook-server-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc
![PKI Graph](subcert-cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc2369046118565646928.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc |
| SerialNumber | 2369046118565646928 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc<br/>- cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc.cluster.local |
| IP Addresses |  |


#### cluster-storage-operator-metrics.openshift-cluster-storage-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-storage-operator | cluster-storage-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### console.openshift-console.svc
![PKI Graph](subcert-console.openshift-console.svc8197011233209597696.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | console.openshift-console.svc |
| SerialNumber | 8197011233209597696 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - console.openshift-console.svc<br/>- console.openshift-console.svc.cluster.local |
| IP Addresses |  |


#### console.openshift-console.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-console | console-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### controller-manager.openshift-controller-manager.svc
![PKI Graph](subcert-controller-manager.openshift-controller-manager.svc7579546110906496244.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | controller-manager.openshift-controller-manager.svc |
| SerialNumber | 7579546110906496244 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - controller-manager.openshift-controller-manager.svc<br/>- controller-manager.openshift-controller-manager.svc.cluster.local |
| IP Addresses |  |


#### controller-manager.openshift-controller-manager.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-controller-manager | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc
![PKI Graph](subcert-csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc5515853684854288522.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc |
| SerialNumber | 5515853684854288522 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc<br/>- csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc.cluster.local |
| IP Addresses |  |


#### csi-snapshot-controller-operator-metrics.openshift-cluster-storage-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-storage-operator | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### csi-snapshot-webhook.openshift-cluster-storage-operator.svc
![PKI Graph](subcert-csi-snapshot-webhook.openshift-cluster-storage-operator.svc4841878348055970704.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | csi-snapshot-webhook.openshift-cluster-storage-operator.svc |
| SerialNumber | 4841878348055970704 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - csi-snapshot-webhook.openshift-cluster-storage-operator.svc<br/>- csi-snapshot-webhook.openshift-cluster-storage-operator.svc.cluster.local |
| IP Addresses |  |


#### csi-snapshot-webhook.openshift-cluster-storage-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-storage-operator | csi-snapshot-webhook-secret |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### dns-default.openshift-dns.svc
![PKI Graph](subcert-dns-default.openshift-dns.svc1698320698676275660.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | dns-default.openshift-dns.svc |
| SerialNumber | 1698320698676275660 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - dns-default.openshift-dns.svc<br/>- dns-default.openshift-dns.svc.cluster.local |
| IP Addresses |  |


#### dns-default.openshift-dns.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-dns | dns-default-metrics-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### etcd.openshift-etcd.svc
![PKI Graph](subcert-etcd.openshift-etcd.svc3848846062663214511.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | etcd.openshift-etcd.svc |
| SerialNumber | 3848846062663214511 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - etcd.openshift-etcd.svc<br/>- etcd.openshift-etcd.svc.cluster.local |
| IP Addresses |  |


#### etcd.openshift-etcd.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### image-registry.openshift-image-registry.svc
![PKI Graph](subcert-image-registry.openshift-image-registry.svc2533061878705206203.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | image-registry.openshift-image-registry.svc |
| SerialNumber | 2533061878705206203 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - image-registry.openshift-image-registry.svc<br/>- image-registry.openshift-image-registry.svc.cluster.local |
| IP Addresses |  |


#### image-registry.openshift-image-registry.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-image-registry | image-registry-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### kube-controller-manager.openshift-kube-controller-manager.svc
![PKI Graph](subcert-kube-controller-manager.openshift-kube-controller-manager.svc1896250909081897532.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | kube-controller-manager.openshift-kube-controller-manager.svc |
| SerialNumber | 1896250909081897532 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - kube-controller-manager.openshift-kube-controller-manager.svc<br/>- kube-controller-manager.openshift-kube-controller-manager.svc.cluster.local |
| IP Addresses |  |


#### kube-controller-manager.openshift-kube-controller-manager.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-kube-controller-manager | serving-cert |
| openshift-kube-controller-manager | serving-cert-2 |
| openshift-kube-controller-manager | serving-cert-3 |
| openshift-kube-controller-manager | serving-cert-4 |
| openshift-kube-controller-manager | serving-cert-5 |
| openshift-kube-controller-manager | serving-cert-6 |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-4/secrets/serving-cert/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-4/secrets/serving-cert/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-5/secrets/serving-cert/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-5/secrets/serving-cert/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-6/secrets/serving-cert/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-6/secrets/serving-cert/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### machine-api-controllers.openshift-machine-api.svc
![PKI Graph](subcert-machine-api-controllers.openshift-machine-api.svc4805346021808337200.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | machine-api-controllers.openshift-machine-api.svc |
| SerialNumber | 4805346021808337200 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - machine-api-controllers.openshift-machine-api.svc<br/>- machine-api-controllers.openshift-machine-api.svc.cluster.local |
| IP Addresses |  |


#### machine-api-controllers.openshift-machine-api.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-api | machine-api-controllers-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### machine-api-operator-webhook.openshift-machine-api.svc
![PKI Graph](subcert-machine-api-operator-webhook.openshift-machine-api.svc821173380931708658.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | machine-api-operator-webhook.openshift-machine-api.svc |
| SerialNumber | 821173380931708658 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - machine-api-operator-webhook.openshift-machine-api.svc<br/>- machine-api-operator-webhook.openshift-machine-api.svc.cluster.local |
| IP Addresses |  |


#### machine-api-operator-webhook.openshift-machine-api.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-api | machine-api-operator-webhook-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### machine-api-operator.openshift-machine-api.svc
![PKI Graph](subcert-machine-api-operator.openshift-machine-api.svc7061257560361596120.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | machine-api-operator.openshift-machine-api.svc |
| SerialNumber | 7061257560361596120 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - machine-api-operator.openshift-machine-api.svc<br/>- machine-api-operator.openshift-machine-api.svc.cluster.local |
| IP Addresses |  |


#### machine-api-operator.openshift-machine-api.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-api | machine-api-operator-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### machine-config-controller.openshift-machine-config-operator.svc
![PKI Graph](subcert-machine-config-controller.openshift-machine-config-operator.svc1669568144903591080.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | machine-config-controller.openshift-machine-config-operator.svc |
| SerialNumber | 1669568144903591080 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - machine-config-controller.openshift-machine-config-operator.svc<br/>- machine-config-controller.openshift-machine-config-operator.svc.cluster.local |
| IP Addresses |  |


#### machine-config-controller.openshift-machine-config-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-config-operator | mcc-proxy-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### machine-config-daemon.openshift-machine-config-operator.svc
![PKI Graph](subcert-machine-config-daemon.openshift-machine-config-operator.svc6197080748599099553.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | machine-config-daemon.openshift-machine-config-operator.svc |
| SerialNumber | 6197080748599099553 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - machine-config-daemon.openshift-machine-config-operator.svc<br/>- machine-config-daemon.openshift-machine-config-operator.svc.cluster.local |
| IP Addresses |  |


#### machine-config-daemon.openshift-machine-config-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-machine-config-operator | proxy-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### marketplace-operator-metrics.openshift-marketplace.svc
![PKI Graph](subcert-marketplace-operator-metrics.openshift-marketplace.svc3229932011111333191.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | marketplace-operator-metrics.openshift-marketplace.svc |
| SerialNumber | 3229932011111333191 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - marketplace-operator-metrics.openshift-marketplace.svc<br/>- marketplace-operator-metrics.openshift-marketplace.svc.cluster.local |
| IP Addresses |  |


#### marketplace-operator-metrics.openshift-marketplace.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-marketplace | marketplace-operator-metrics |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-apiserver-operator.svc
![PKI Graph](subcert-metrics.openshift-apiserver-operator.svc3364654077885766892.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-apiserver-operator.svc |
| SerialNumber | 3364654077885766892 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-apiserver-operator.svc<br/>- metrics.openshift-apiserver-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-apiserver-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-apiserver-operator | openshift-apiserver-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-authentication-operator.svc
![PKI Graph](subcert-metrics.openshift-authentication-operator.svc1159011675408915412.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-authentication-operator.svc |
| SerialNumber | 1159011675408915412 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-authentication-operator.svc<br/>- metrics.openshift-authentication-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-authentication-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-authentication-operator | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-config-operator.svc
![PKI Graph](subcert-metrics.openshift-config-operator.svc2875861632403042198.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-config-operator.svc |
| SerialNumber | 2875861632403042198 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-config-operator.svc<br/>- metrics.openshift-config-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-config-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-config-operator | config-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-console-operator.svc
![PKI Graph](subcert-metrics.openshift-console-operator.svc2594480697393708491.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-console-operator.svc |
| SerialNumber | 2594480697393708491 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-console-operator.svc<br/>- metrics.openshift-console-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-console-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-console-operator | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-controller-manager-operator.svc
![PKI Graph](subcert-metrics.openshift-controller-manager-operator.svc6876654977757942073.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-controller-manager-operator.svc |
| SerialNumber | 6876654977757942073 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-controller-manager-operator.svc<br/>- metrics.openshift-controller-manager-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-controller-manager-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-controller-manager-operator | openshift-controller-manager-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-dns-operator.svc
![PKI Graph](subcert-metrics.openshift-dns-operator.svc3456920330567479390.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-dns-operator.svc |
| SerialNumber | 3456920330567479390 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-dns-operator.svc<br/>- metrics.openshift-dns-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-dns-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-dns-operator | metrics-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-etcd-operator.svc
![PKI Graph](subcert-metrics.openshift-etcd-operator.svc6677616133175219939.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-etcd-operator.svc |
| SerialNumber | 6677616133175219939 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-etcd-operator.svc<br/>- metrics.openshift-etcd-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-etcd-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-etcd-operator | etcd-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-ingress-operator.svc
![PKI Graph](subcert-metrics.openshift-ingress-operator.svc848448456271825148.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-ingress-operator.svc |
| SerialNumber | 848448456271825148 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-ingress-operator.svc<br/>- metrics.openshift-ingress-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-ingress-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-ingress-operator | metrics-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-insights.svc
![PKI Graph](subcert-metrics.openshift-insights.svc7854905682415005669.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-insights.svc |
| SerialNumber | 7854905682415005669 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-insights.svc<br/>- metrics.openshift-insights.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-insights.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-insights | openshift-insights-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-kube-apiserver-operator.svc
![PKI Graph](subcert-metrics.openshift-kube-apiserver-operator.svc6235465719572846931.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-kube-apiserver-operator.svc |
| SerialNumber | 6235465719572846931 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-kube-apiserver-operator.svc<br/>- metrics.openshift-kube-apiserver-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-kube-apiserver-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-kube-apiserver-operator | kube-apiserver-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-kube-controller-manager-operator.svc
![PKI Graph](subcert-metrics.openshift-kube-controller-manager-operator.svc5438971602805831466.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-kube-controller-manager-operator.svc |
| SerialNumber | 5438971602805831466 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-kube-controller-manager-operator.svc<br/>- metrics.openshift-kube-controller-manager-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-kube-controller-manager-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-kube-controller-manager-operator | kube-controller-manager-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-kube-scheduler-operator.svc
![PKI Graph](subcert-metrics.openshift-kube-scheduler-operator.svc4366531022562557239.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-kube-scheduler-operator.svc |
| SerialNumber | 4366531022562557239 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-kube-scheduler-operator.svc<br/>- metrics.openshift-kube-scheduler-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-kube-scheduler-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-kube-scheduler-operator | kube-scheduler-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-kube-storage-version-migrator-operator.svc
![PKI Graph](subcert-metrics.openshift-kube-storage-version-migrator-operator.svc1904991570617038281.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-kube-storage-version-migrator-operator.svc |
| SerialNumber | 1904991570617038281 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-kube-storage-version-migrator-operator.svc<br/>- metrics.openshift-kube-storage-version-migrator-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-kube-storage-version-migrator-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-kube-storage-version-migrator-operator | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### metrics.openshift-service-ca-operator.svc
![PKI Graph](subcert-metrics.openshift-service-ca-operator.svc7400587186269430937.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | metrics.openshift-service-ca-operator.svc |
| SerialNumber | 7400587186269430937 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - metrics.openshift-service-ca-operator.svc<br/>- metrics.openshift-service-ca-operator.svc.cluster.local |
| IP Addresses |  |


#### metrics.openshift-service-ca-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-service-ca-operator | serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### multus-admission-controller.openshift-multus.svc
![PKI Graph](subcert-multus-admission-controller.openshift-multus.svc1734585295770866263.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | multus-admission-controller.openshift-multus.svc |
| SerialNumber | 1734585295770866263 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - multus-admission-controller.openshift-multus.svc<br/>- multus-admission-controller.openshift-multus.svc.cluster.local |
| IP Addresses |  |


#### multus-admission-controller.openshift-multus.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-multus | multus-admission-controller-secret |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### oauth-openshift.openshift-authentication.svc
![PKI Graph](subcert-oauth-openshift.openshift-authentication.svc4046784375362903289.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | oauth-openshift.openshift-authentication.svc |
| SerialNumber | 4046784375362903289 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - oauth-openshift.openshift-authentication.svc<br/>- oauth-openshift.openshift-authentication.svc.cluster.local |
| IP Addresses |  |


#### oauth-openshift.openshift-authentication.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-authentication | v4-0-config-system-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### olm-operator-metrics.openshift-operator-lifecycle-manager.svc
![PKI Graph](subcert-olm-operator-metrics.openshift-operator-lifecycle-manager.svc5116318574914027760.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | olm-operator-metrics.openshift-operator-lifecycle-manager.svc |
| SerialNumber | 5116318574914027760 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - olm-operator-metrics.openshift-operator-lifecycle-manager.svc<br/>- olm-operator-metrics.openshift-operator-lifecycle-manager.svc.cluster.local |
| IP Addresses |  |


#### olm-operator-metrics.openshift-operator-lifecycle-manager.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-operator-lifecycle-manager | olm-operator-serving-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc
![PKI Graph](subcert-performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc4586704378265138575.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc |
| SerialNumber | 4586704378265138575 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc<br/>- performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc.cluster.local |
| IP Addresses |  |


#### performance-addon-operator-service.openshift-cluster-node-tuning-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cluster-node-tuning-operator | performance-addon-operator-webhook-cert |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### pod-identity-webhook.openshift-cloud-credential-operator.svc
![PKI Graph](subcert-pod-identity-webhook.openshift-cloud-credential-operator.svc2242223452909203393.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | pod-identity-webhook.openshift-cloud-credential-operator.svc |
| SerialNumber | 2242223452909203393 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - pod-identity-webhook.openshift-cloud-credential-operator.svc<br/>- pod-identity-webhook.openshift-cloud-credential-operator.svc.cluster.local |
| IP Addresses |  |


#### pod-identity-webhook.openshift-cloud-credential-operator.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-cloud-credential-operator | pod-identity-webhook |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### prometheus-adapter.openshift-monitoring.svc
![PKI Graph](subcert-prometheus-adapter.openshift-monitoring.svc1732525563472031104.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | prometheus-adapter.openshift-monitoring.svc |
| SerialNumber | 1732525563472031104 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - prometheus-adapter.openshift-monitoring.svc<br/>- prometheus-adapter.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### prometheus-adapter.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | prometheus-adapter-5f9puh20l637u |
| openshift-monitoring | prometheus-adapter-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### prometheus-k8s.openshift-monitoring.svc
![PKI Graph](subcert-prometheus-k8s.openshift-monitoring.svc7607235645352172219.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | prometheus-k8s.openshift-monitoring.svc |
| SerialNumber | 7607235645352172219 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - prometheus-k8s.openshift-monitoring.svc<br/>- prometheus-k8s.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### prometheus-k8s.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | prometheus-k8s-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### prometheus-operator-admission-webhook.openshift-monitoring.svc
![PKI Graph](subcert-prometheus-operator-admission-webhook.openshift-monitoring.svc8444555083473456618.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | prometheus-operator-admission-webhook.openshift-monitoring.svc |
| SerialNumber | 8444555083473456618 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - prometheus-operator-admission-webhook.openshift-monitoring.svc<br/>- prometheus-operator-admission-webhook.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### prometheus-operator-admission-webhook.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | prometheus-operator-admission-webhook-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### router-internal-default.openshift-ingress.svc
![PKI Graph](subcert-router-internal-default.openshift-ingress.svc5196656952303372444.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | router-internal-default.openshift-ingress.svc |
| SerialNumber | 5196656952303372444 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - router-internal-default.openshift-ingress.svc<br/>- router-internal-default.openshift-ingress.svc.cluster.local |
| IP Addresses |  |


#### router-internal-default.openshift-ingress.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-ingress | router-metrics-certs-default |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |




### scheduler.openshift-kube-scheduler.svc
![PKI Graph](subcert-scheduler.openshift-kube-scheduler.svc449821371379777563.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | scheduler.openshift-kube-scheduler.svc |
| SerialNumber | 449821371379777563 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - scheduler.openshift-kube-scheduler.svc<br/>- scheduler.openshift-kube-scheduler.svc.cluster.local |
| IP Addresses |  |


#### scheduler.openshift-kube-scheduler.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-kube-scheduler | serving-cert |
| openshift-kube-scheduler | serving-cert-3 |
| openshift-kube-scheduler | serving-cert-4 |
| openshift-kube-scheduler | serving-cert-5 |
| openshift-kube-scheduler | serving-cert-6 |
| openshift-kube-scheduler | serving-cert-7 |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/kube-scheduler-pod-6/secrets/serving-cert/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-scheduler-pod-6/secrets/serving-cert/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-scheduler-pod-7/secrets/serving-cert/tls.crt/tls.crt | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-scheduler-pod-7/secrets/serving-cert/tls.crt/tls.key | -rw-------. | root | root | system_u:object_r:kubernetes_file_t:s0 |



### thanos-querier.openshift-monitoring.svc
![PKI Graph](subcert-thanos-querier.openshift-monitoring.svc2215041012551543033.png)



| Property | Value |
| ----------- | ----------- |
| Type | Serving |
| CommonName | thanos-querier.openshift-monitoring.svc |
| SerialNumber | 2215041012551543033 |
| Issuer CommonName | [service-serving-signer](#service-serving-signer) |
| Validity | 2y |
| Signature Algorithm | SHA256-RSA |
| PublicKey Algorithm | RSA 2048 bit |
| Usages | - KeyUsageDigitalSignature<br/>- KeyUsageKeyEncipherment |
| ExtendedUsages | - ExtKeyUsageServerAuth |
| DNS Names | - thanos-querier.openshift-monitoring.svc<br/>- thanos-querier.openshift-monitoring.svc.cluster.local |
| IP Addresses |  |


#### thanos-querier.openshift-monitoring.svc Locations
| Namespace | Secret Name |
| ----------- | ----------- |
| openshift-monitoring | thanos-querier-tls |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |



## Client Certificate/Key Pairs

## Certificates Without Keys

These certificates are present in certificate authority bundles, but do not have keys in the cluster.
This happens when the installer bootstrap clusters with a set of certificate/key pairs that are deleted during the
installation process.

## Certificate Authority Bundles


### service-ca
![PKI Graph](subca-2412928394.png)

CA for recognizing serving certificates for services that were signed by our service-ca controller.

**Bundled Certificates**

| CommonName | Issuer CommonName | Validity | PublicKey Algorithm |
| ----------- | ----------- | ----------- | ----------- |
| [service-serving-signer](#service-serving-signer) | [service-serving-signer](#service-serving-signer) | 2y60d | RSA 2048 bit |

#### service-ca Locations
| Namespace | ConfigMap Name |
| ----------- | ----------- |
| openshift-config-managed | service-ca |
| openshift-kube-controller-manager | service-ca |
| openshift-kube-controller-manager | service-ca-2 |
| openshift-kube-controller-manager | service-ca-3 |
| openshift-kube-controller-manager | service-ca-4 |
| openshift-kube-controller-manager | service-ca-5 |
| openshift-kube-controller-manager | service-ca-6 |
| openshift-service-ca | signing-cabundle |

| File | Permissions | User | Group | SE Linux |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-4/configmaps/service-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-5/configmaps/service-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |
| /etc/kubernetes/static-pod-resources/kube-controller-manager-pod-6/configmaps/service-ca/ca-bundle.crt/ca-bundle.crt | -rw-r--r--. | root | root | system_u:object_r:kubernetes_file_t:s0 |


