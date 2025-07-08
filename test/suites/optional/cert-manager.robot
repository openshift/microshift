*** Settings ***
Documentation       Test suite for generating certificates using cert-manager with YAML manifests

Library             Collections
Library             OperatingSystem
Library             String
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           cert-manager    certificates    tls


*** Variables ***
${CERT_NAME}                test-certificate
${SECRET_NAME}              test-cert-secret
${ISSUER_NAME}              test-issuer
${CERT_COMMON_NAME}         example.com
${CERT_DNS_NAME}            example.com
${CERTS_NAMESPACE}          test-certs
${ROUTE_NAME}               hello-app
${CERT_ISSUER_YAML}         SEPARATOR=\n
...                         ---
...                         apiVersion: cert-manager.io/v1
...                         kind: ClusterIssuer
...                         metadata:
...                         \ \ name: ${ISSUER_NAME}
...                         spec:
...                         \ \ selfSigned: {}

${CERT_YAML}                SEPARATOR=\n
...                         ---
...                         apiVersion: cert-manager.io/v1
...                         kind: Certificate
...                         metadata:
...                         \ \ name: ${CERT_NAME}
...                         \ \ namespace: ${CERTS_NAMESPACE}
...                         spec:
...                         \ \ secretName: ${SECRET_NAME}
...                         \ \ issuerRef:
...                         \ \ \ \ name: ${ISSUER_NAME}
...                         \ \ \ \ kind: ClusterIssuer
...                         \ \ commonName: ${CERT_COMMON_NAME}
...                         \ \ dnsNames:
...                         \ \ - ${CERT_DNS_NAME}
...                         \ \ - www.${CERT_DNS_NAME}
${INGRESS_RBAC_YAML}        SEPARATOR=\n
...                         ---
...                         apiVersion: rbac.authorization.k8s.io/v1
...                         kind: Role
...                         metadata:
...                         \ \ name: secret-reader
...                         \ \ namespace: ${CERTS_NAMESPACE}
...                         rules:
...                         - apiGroups: [""]
...                         \ \ resources: ["secrets"]
...                         \ \ resourceNames: ["${SECRET_NAME}"]
...                         \ \ verbs: ["get", "list", "watch"]
...                         ---
...                         apiVersion: rbac.authorization.k8s.io/v1
...                         kind: RoleBinding
...                         metadata:
...                         \ \ name: ingress-secret-reader
...                         \ \ namespace: ${CERTS_NAMESPACE}
...                         roleRef:
...                         \ \ apiGroup: rbac.authorization.k8s.io
...                         \ \ kind: Role
...                         \ \ name: secret-reader
...                         subjects:
...                         - kind: ServiceAccount
...                         \ \ name: router
...                         \ \ namespace: openshift-ingress

${INGRESS_ROUTE_YAML}       SEPARATOR=\n
...                         ---
...                         apiVersion: route.openshift.io/v1
...                         kind: Route
...                         metadata:
...                         \ \ name: ${ROUTE_NAME}
...                         \ \ namespace: ${CERTS_NAMESPACE}
...                         spec:
...                         \ \ port:
...                         \ \ \ \ targetPort: 8080
...                         \ \ tls:
...                         \ \ \ \ externalCertificate:
...                         \ \ \ \ \ \ name: ${SECRET_NAME}
...                         \ \ \ \ termination: edge
...                         \ \ to:
...                         \ \ \ \ kind: Service
...                         \ \ \ \ name: hello-microshift


*** Test Cases ***
Create Ingress route with Custom certificate
    [Documentation]    Create route with a custom certificate
    [Setup]    Run Keywords
    ...    Setup Namespace
    Apply YAML Manifest    ${CERT_ISSUER_YAML}
    Oc Wait    -n ${CERTS_NAMESPACE} clusterissuer ${ISSUER_NAME}    --for="condition=Ready" --timeout=120s
    Apply YAML Manifest    ${CERT_YAML}
    Oc Wait    -n ${CERTS_NAMESPACE} certificate ${CERT_NAME}    --for="condition=Ready" --timeout=60s
    Apply YAML Manifest    ${INGRESS_RBAC_YAML}
    Deploy Hello MicroShift
    Apply YAML Manifest    ${INGRESS_ROUTE_YAML}
    Oc Wait    -n ${CERTS_NAMESPACE} route ${ROUTE_NAME}    --for=jsonpath='.status.ingress' --timeout=60s
    [Teardown]    Run Keywords
    ...    Remove ClusterIssuer


*** Keywords ***
Deploy Hello MicroShift
    [Documentation]    Deploys the hello microshift application (service included)
    ...    in the given namespace.
    Create Hello MicroShift Pod
    Expose Hello MicroShift

Remove ClusterIssuer
    [Documentation]    Remove the cluster issuer
    Oc Delete    clusterissuer/${ISSUER_NAME}

Apply YAML Manifest
    [Documentation]    Apply YAML manifest to the cluster
    [Arguments]    ${yaml_content}
    ${temp_file}=    Create Random Temp File    ${yaml_content}
    ${result}=    Run With Kubeconfig    oc apply -f ${temp_file}
    Remove File    ${temp_file}
    Should Contain    ${result}    created    msg=Failed to apply YAML manifest
    Log    Applied manifest: ${result}

Setup Namespace
    [Documentation]    Setup namespace for cert-manager tests
    Set Suite Variable    \${NAMESPACE}    ${CERTS_NAMESPACE}
    Create Namespace    ${CERTS_NAMESPACE}
