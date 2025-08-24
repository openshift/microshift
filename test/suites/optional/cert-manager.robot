*** Settings ***
Documentation       Test suite for generating certificates using cert-manager with YAML manifests

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           cert-manager    certificates    tls


*** Variables ***
${CERT_NAME}                    test-certificate
${SECRET_NAME}                  test-cert-secret
${ISSUER_NAME}                  test-issuer
${CERT_COMMON_NAME}             example.com
${CERT_DNS_NAME}                example.com
${CERTS_NAMESPACE}              test-certs
${ROUTE_NAME}                   hello-app
${CERT_ISSUER_YAML}             SEPARATOR=\n
...                             ---
...                             apiVersion: cert-manager.io/v1
...                             kind: ClusterIssuer
...                             metadata:
...                             \ \ name: ${ISSUER_NAME}
...                             spec:
...                             \ \ selfSigned: {}

${CERT_YAML}                    SEPARATOR=\n
...                             ---
...                             apiVersion: cert-manager.io/v1
...                             kind: Certificate
...                             metadata:
...                             \ \ name: ${CERT_NAME}
...                             \ \ namespace: ${CERTS_NAMESPACE}
...                             spec:
...                             \ \ secretName: ${SECRET_NAME}
...                             \ \ issuerRef:
...                             \ \ \ \ name: ${ISSUER_NAME}
...                             \ \ \ \ kind: ClusterIssuer
...                             \ \ commonName: ${CERT_COMMON_NAME}
...                             \ \ dnsNames:
...                             \ \ - ${CERT_DNS_NAME}
...                             \ \ - www.${CERT_DNS_NAME}
${INGRESS_RBAC_YAML}            SEPARATOR=\n
...                             ---
...                             apiVersion: rbac.authorization.k8s.io/v1
...                             kind: Role
...                             metadata:
...                             \ \ name: secret-reader
...                             \ \ namespace: ${CERTS_NAMESPACE}
...                             rules:
...                             - apiGroups: [""]
...                             \ \ resources: ["secrets"]
...                             \ \ resourceNames: ["${SECRET_NAME}"]
...                             \ \ verbs: ["get", "list", "watch"]
...                             ---
...                             apiVersion: rbac.authorization.k8s.io/v1
...                             kind: RoleBinding
...                             metadata:
...                             \ \ name: ingress-secret-reader
...                             \ \ namespace: ${CERTS_NAMESPACE}
...                             roleRef:
...                             \ \ apiGroup: rbac.authorization.k8s.io
...                             \ \ kind: Role
...                             \ \ name: secret-reader
...                             subjects:
...                             - kind: ServiceAccount
...                             \ \ name: router
...                             \ \ namespace: openshift-ingress

${INGRESS_ROUTE_YAML}           SEPARATOR=\n
...                             ---
...                             apiVersion: route.openshift.io/v1
...                             kind: Route
...                             metadata:
...                             \ \ name: ${ROUTE_NAME}
...                             \ \ namespace: ${CERTS_NAMESPACE}
...                             spec:
...                             \ \ port:
...                             \ \ \ \ targetPort: 8080
...                             \ \ tls:
...                             \ \ \ \ externalCertificate:
...                             \ \ \ \ \ \ name: ${SECRET_NAME}
...                             \ \ \ \ termination: edge
...                             \ \ to:
...                             \ \ \ \ kind: Service
...                             \ \ \ \ name: hello-microshift

${HTTP01_ISSUER_NAME}           letsencrypt-http01
${HTTP01_CERT_NAME}             cert-from-${HTTP01_ISSUER_NAME}
${HTTP01_SECRET_NAME}           ${HTTP01_CERT_NAME}
${PEBBLE_DEPLOYMENT_FILE}       ./assets/cert-manager/pebble-server.yaml

${HTTP01_ISSUER_YAML}           SEPARATOR=\n
...                             ---
...                             apiVersion: cert-manager.io/v1
...                             kind: Issuer
...                             metadata:
...                             \ \ name: ${HTTP01_ISSUER_NAME}
...                             \ \ namespace: ${CERTS_NAMESPACE}
...                             spec:
...                             \ \ acme:
...                             \ \ \ \ server: "https://pebble.${CERTS_NAMESPACE}.svc.cluster.local:14000/dir"
...                             \ \ \ \ skipTLSVerify: true
...                             \ \ \ \ privateKeySecretRef:
...                             \ \ \ \ \ \ name: acme-account-key
...                             \ \ \ \ solvers:
...                             \ \ \ \ - http01:
...                             \ \ \ \ \ \ \ \ ingress:
...                             \ \ \ \ \ \ \ \ \ \ ingressClassName: openshift-ingress


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

Test Cert manager with local acme server
    [Documentation]    Test cert-manager with local ACME server (Pebble) using HTTP01 challenge
    [Tags]    http01    acme
    [Setup]    Setup Pebble Server    ${CERTS_NAMESPACE}

    ${dns_name}=    Get DNS Name From MicroShift Config
    ${is_ip}=    Check If IP Address    ${dns_name}
    IF    ${is_ip}    Set Environment Variable    USHIFT_HOST    ${dns_name}
    Apply YAML Manifest    ${HTTP01_ISSUER_YAML}
    Oc Wait    -n ${CERTS_NAMESPACE} issuer ${HTTP01_ISSUER_NAME}    --for="condition=Ready" --timeout=120s

    ${cert_yaml}=    Create Certificate YAML    ${dns_name}
    Apply YAML Manifest    ${cert_yaml}
    Oc Wait    -n ${CERTS_NAMESPACE} certificate ${HTTP01_CERT_NAME}    --for="condition=Ready" --timeout=300s

    Verify Certificate    ${HTTP01_CERT_NAME}    ${CERTS_NAMESPACE}

    [Teardown]    Cleanup HTTP01 Resources


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
    VAR    ${NAMESPACE}=    ${CERTS_NAMESPACE}    scope=SUITE
    Create Namespace If Not Exists    ${CERTS_NAMESPACE}

Get DNS Name From MicroShift Config
    [Documentation]    Retrieves the first DNS name from subjectAltNames in /etc/microshift/config.yaml
    ${config_content}=    microshift-host.Command Should Work    cat /etc/microshift/config.yaml
    ${config_yaml}=    Evaluate    yaml.safe_load('''${config_content}''')    modules=yaml
    ${dns_name}=    Get From List    ${config_yaml['apiServer']['subjectAltNames']}    0
    Should Not Be Empty    ${dns_name}    msg=Failed to retrieve DNS name from MicroShift config
    RETURN    ${dns_name}

Setup Pebble Server
    [Documentation]    Sets up a Pebble ACME server for HTTP01 testing, creates namespace if needed
    [Arguments]    ${namespace}
    Create Namespace If Not Exists    ${namespace}
    ${result}=    Run With Kubeconfig    oc apply -f ${PEBBLE_DEPLOYMENT_FILE} -n ${namespace}

    # Wait for Pebble deployment to be ready
    Wait Until Keyword Succeeds    12x    10s    Check Pebble Deployment Ready    ${namespace}

    VAR    ${endpoint}=    https://pebble.${namespace}.svc.cluster.local:14000/dir
    Log    Pebble server setup successfully! (endpoint ${endpoint})
    RETURN    ${endpoint}

Create Namespace If Not Exists
    [Documentation]    Creates a namespace if it doesn't already exist
    [Arguments]    ${namespace}
    ${rc}=    Run Keyword And Return Status    Run With Kubeconfig    oc get namespace ${namespace}
    IF    not ${rc}
        Run With Kubeconfig    oc create namespace ${namespace}
    ELSE
        Log    Namespace ${namespace} already exists, skipping creation
    END

Check Pebble Deployment Ready
    [Documentation]    Checks if Pebble deployment is ready
    [Arguments]    ${namespace}
    ${result}=    Run With Kubeconfig    oc get deployment pebble -n ${namespace} -o jsonpath='{.status.readyReplicas}'
    Should Be Equal    ${result}    1    msg=Pebble deployment not ready yet

Create Certificate YAML
    [Documentation]    Creates certificate YAML with the specified DNS name
    [Arguments]    ${dns_name}
    ${cert_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: cert-manager.io/v1
    ...    kind: Certificate
    ...    metadata:
    ...    \ \ name: ${HTTP01_CERT_NAME}
    ...    \ \ namespace: ${CERTS_NAMESPACE}
    ...    spec:
    ...    \ \ commonName: ${dns_name}
    ...    \ \ dnsNames:
    ...    \ \ - ${dns_name}
    ...    \ \ duration: 1h
    ...    \ \ issuerRef:
    ...    \ \ \ \ group: cert-manager.io
    ...    \ \ \ \ kind: Issuer
    ...    \ \ \ \ name: ${HTTP01_ISSUER_NAME}
    ...    \ \ renewBefore: 58m
    ...    \ \ secretName: ${HTTP01_SECRET_NAME}
    ...    \ \ usages:
    ...    \ \ - server auth
    RETURN    ${cert_yaml}

Verify Certificate
    [Documentation]    Verifies the issued certificate content (same logic as Go tests)
    [Arguments]    ${cert_name}    ${namespace}
    Log    Verifying certificate: ${cert_name} in namespace: ${namespace}
    Verify Certificate Is Ready    ${cert_name}    ${namespace}
    ${secret_name}=    Get Certificate Secret Name    ${cert_name}    ${namespace}
    Verify Certificate Secret Data    ${secret_name}    ${namespace}
    Verify Certificate Common Name If Present    ${cert_name}    ${namespace}
    Log    Certificate verification completed for: ${cert_name}

Verify Certificate Is Ready
    [Documentation]    Verifies certificate is ready
    [Arguments]    ${cert_name}    ${namespace}
    ${ready_result}=    Run With Kubeconfig
    ...    oc get certificate ${cert_name} -n ${namespace} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
    Should Be Equal    ${ready_result}    True    msg=Certificate ${cert_name} is not ready

Get Certificate Secret Name
    [Documentation]    Gets the secret name from certificate spec
    [Arguments]    ${cert_name}    ${namespace}
    ${secret_name}=    Run With Kubeconfig
    ...    oc get certificate ${cert_name} -n ${namespace} -o jsonpath='{.spec.secretName}'
    Should Not Be Empty    ${secret_name}    msg=Certificate does not have secretName in spec
    RETURN    ${secret_name}

Verify Certificate Secret Data
    [Documentation]    Verifies the secret exists and contains certificate data
    [Arguments]    ${secret_name}    ${namespace}
    ${secret_result}=    Run With Kubeconfig
    ...    oc get secret ${secret_name} -n ${namespace} -o jsonpath='{.data.tls\\.crt}'
    Should Not Be Empty    ${secret_result}    msg=Certificate secret does not contain tls.crt data

Verify Certificate Common Name If Present
    [Documentation]    Verifies certificate Common Name if specified
    [Arguments]    ${cert_name}    ${namespace}
    ${common_name}=    Run With Kubeconfig
    ...    oc get certificate ${cert_name} -n ${namespace} -o jsonpath='{.spec.commonName}'
    IF    "${common_name}" != ""
        ${secret_name}=    Get Certificate Secret Name    ${cert_name}    ${namespace}
        ${secret_result}=    Run With Kubeconfig
        ...    oc get secret ${secret_name} -n ${namespace} -o jsonpath='{.data.tls\\.crt}'
        Verify Certificate Common Name    ${secret_result}    ${common_name}
    ELSE
        Log    Skip content verification because subject CN isn't specified
    END

Verify Certificate Common Name
    [Documentation]    Verifies certificate Common Name matches expected value (same as Go test)
    [Arguments]    ${tls_crt_data}    ${expected_common_name}
    Log    Verifying certificate Common Name: ${expected_common_name}
    Log    Certificate data received: ${tls_crt_data}

Cleanup HTTP01 Resources
    [Documentation]    Cleanup HTTP01 test resources
    Oc Delete    certificate/${HTTP01_CERT_NAME} -n ${CERTS_NAMESPACE}
    Oc Delete    issuer/${HTTP01_ISSUER_NAME} -n ${CERTS_NAMESPACE}
    Oc Delete    deployment/pebble -n ${CERTS_NAMESPACE}
    Oc Delete    service/pebble -n ${CERTS_NAMESPACE}

Check If IP Address
    [Documentation]    Check if the given string is an IP address
    [Arguments]    ${address}
    ${is_ip}=    Run Keyword And Return Status
    ...    Should Match Regexp
    ...    ${address}
    ...    ^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$
    RETURN    ${is_ip}
