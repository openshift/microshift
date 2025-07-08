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
${CERT_NAME}            test-certificate
${SECRET_NAME}          test-cert-secret
${ISSUER_NAME}          test-issuer
${CERT_COMMON_NAME}     example.com
${CERT_DNS_NAME}        example.com
${CERT_TIMEOUT}         300s


*** Test Cases ***
Create Ingress route with Custom certificate
    [Documentation]    Create route with a custom certificate
    ${issuer_yaml}=    Create Self-Signed Issuer YAML
    Apply YAML Manifest    ${issuer_yaml}
    Wait For ClusterIssuer Ready    ${ISSUER_NAME}
    ${cert_yaml}=    Create Certificate YAML
    Apply YAML Manifest    ${cert_yaml}
    Wait For Certificate Ready    ${CERT_NAME}
    Wait For Secret Exists    ${SECRET_NAME}
    ${ingress_rbac_yaml}=    Create Ingress RBAC YAML
    Apply YAML Manifest    ${ingress_rbac_yaml}
    Deploy Hello MicroShift
    ${route_yaml}=    Create Route YAML
    Apply YAML Manifest    ${route_yaml}
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

Create Self-Signed Issuer YAML
    [Documentation]    Generate YAML for a self-signed ClusterIssuer
    ${yaml_content}=    Catenate    SEPARATOR=\n
    ...    apiVersion: cert-manager.io/v1
    ...    kind: ClusterIssuer
    ...    metadata:
    ...    \ \ name: ${ISSUER_NAME}
    ...    spec:
    ...    \ \ selfSigned: {}
    RETURN    ${yaml_content}

Create Ingress RBAC YAML
    [Documentation]    Generate YAML for a role and role binding
    ${yaml_content}=    Catenate    SEPARATOR=\n
    ...    apiVersion: rbac.authorization.k8s.io/v1
    ...    kind: Role
    ...    metadata:
    ...    \ \ name: secret-reader
    ...    \ \ namespace: ${NAMESPACE}
    ...    rules:
    ...    - apiGroups: [""]
    ...    \ \ resources: ["secrets"]
    ...    \ \ resourceNames: ["${SECRET_NAME}"]
    ...    \ \ verbs: ["get", "list", "watch"]
    ...    ---
    ...    apiVersion: rbac.authorization.k8s.io/v1
    ...    kind: RoleBinding
    ...    metadata:
    ...    \ \ name: ingress-secret-reader
    ...    \ \ namespace: ${NAMESPACE}
    ...    roleRef:
    ...    \ \ apiGroup: rbac.authorization.k8s.io
    ...    \ \ kind: Role
    ...    \ \ name: secret-reader
    ...    subjects:
    ...    - kind: ServiceAccount
    ...    \ \ name: router
    ...    \ \ namespace: openshift-ingress
    RETURN    ${yaml_content}

Create Route YAML
    [Documentation]    Generate YAML for a route
    ${yaml_content}=    Catenate    SEPARATOR=\n
    ...    apiVersion: route.openshift.io/v1
    ...    kind: Route
    ...    metadata:
    ...    \ \ name: hello-app
    ...    \ \ namespace: ${NAMESPACE}
    ...    spec:
    ...    \ \ port:
    ...    \ \ \ \ targetPort: 8080
    ...    \ \ tls:
    ...    \ \ \ \ externalCertificate:
    ...    \ \ \ \ \ \ name: ${SECRET_NAME}
    ...    \ \ \ \ termination: edge
    ...    \ \ to:
    ...    \ \ \ \ kind: Service
    ...    \ \ \ \ name: hello-microshift
    RETURN    ${yaml_content}

Create Certificate YAML
    [Documentation]    Generate YAML for a basic certificate
    ${yaml_content}=    Catenate    SEPARATOR=\n
    ...    apiVersion: cert-manager.io/v1
    ...    kind: Certificate
    ...    metadata:
    ...    \ \ name: ${CERT_NAME}
    ...    \ \ namespace: ${NAMESPACE}
    ...    spec:
    ...    \ \ secretName: ${SECRET_NAME}
    ...    \ \ issuerRef:
    ...    \ \ \ \ name: ${ISSUER_NAME}
    ...    \ \ \ \ kind: ClusterIssuer
    ...    \ \ commonName: ${CERT_COMMON_NAME}
    ...    \ \ dnsNames:
    ...    \ \ - ${CERT_DNS_NAME}
    ...    \ \ - www.${CERT_DNS_NAME}
    RETURN    ${yaml_content}

Apply YAML Manifest
    [Documentation]    Apply a YAML manifest to the cluster
    [Arguments]    ${yaml_content}
    ${temp_file}=    Create Random Temp File    ${yaml_content}
    ${result}=    Run With Kubeconfig    oc apply -f ${temp_file}
    Remove File    ${temp_file}
    Should Contain    ${result}    created    msg=Failed to apply YAML manifest
    Log    Applied manifest: ${result}

Wait For ClusterIssuer Ready
    [Documentation]    Wait for ClusterIssuer to be ready
    [Arguments]    ${issuer_name}
    FOR    ${i}    IN RANGE    60
        ${status}=    Run With Kubeconfig
        ...    oc get clusterissuer ${issuer_name} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
        ...    allow_fail=True
        IF    "${status}" == "True"
            Log    ClusterIssuer ${issuer_name} is ready
            RETURN
        END
        Sleep    5s
    END
    Fail    ClusterIssuer ${issuer_name} did not become ready within timeout

Wait For Certificate Ready
    [Documentation]    Wait for Certificate to be ready
    [Arguments]    ${cert_name}
    FOR    ${i}    IN RANGE    60
        ${status}=    Run With Kubeconfig
        ...    oc get certificate ${cert_name} -n ${NAMESPACE} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
        ...    allow_fail=True
        IF    "${status}" == "True"
            Log    Certificate ${cert_name} is ready
            RETURN
        END
        Sleep    5s
    END
    Fail    Certificate ${cert_name} did not become ready within timeout

Wait For Secret Exists
    [Documentation]    Wait for secret to be created
    [Arguments]    ${secret_name}
    FOR    ${i}    IN RANGE    60
        ${exists}=    Run With Kubeconfig
        ...    oc get secret ${secret_name} -n ${NAMESPACE}
        ...    allow_fail=True
        ...    return_rc=True
        IF    ${exists[1]} == 0
            Log    Secret ${secret_name} exists
            RETURN
        END
        Sleep    5s
    END
    Fail    Secret ${secret_name} was not created within timeout

Verify Secret Contains TLS Data
    [Documentation]    Verify that the secret contains TLS certificate and key
    [Arguments]    ${secret_name}
    ${secret_data}=    Run With Kubeconfig    oc get secret ${secret_name} -n ${NAMESPACE} -o jsonpath='{.data}'
    Should Contain    ${secret_data}    tls.crt
    Should Contain    ${secret_data}    tls.key
    Log    Secret ${secret_name} contains required TLS data
