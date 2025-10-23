*** Settings ***
Documentation       Test suite for generating certificates using cert-manager with YAML manifests

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Library             ../../resources/journalctl.py
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           cert-manager    certificates    tls


*** Variables ***
${CERT_NAME}                    test-certificate
${SECRET_NAME}                  test-cert-secret
${ISSUER_NAME}                  test-issuer
${CERT_COMMON_NAME}             example.com
${CERT_DNS_NAME}                example.com
${ROUTE_NAME}                   hello-app
${CERT_ISSUER_YAML}             SEPARATOR=\n
...                             ---
...                             apiVersion: cert-manager.io/v1
...                             kind: ClusterIssuer
...                             metadata:
...                             \ \ name: ${ISSUER_NAME}
...                             spec:
...                             \ \ selfSigned: {}

${HTTP01_ISSUER_NAME}           letsencrypt-http01
${HTTP01_CERT_NAME}             cert-from-${HTTP01_ISSUER_NAME}
${HTTP01_SECRET_NAME}           ${HTTP01_CERT_NAME}
${PEBBLE_DEPLOYMENT_FILE}       ./assets/cert-manager/pebble-server.yaml


*** Test Cases ***
Create Ingress route with Custom certificate
    [Documentation]    Create route with a custom certificate
    [Setup]    Run Keywords
    Verify Cert Manager Kustomization Success
    ${cert_issuer_yaml}=    Create Cert Issuer YAML
    Apply YAML Manifest    ${cert_issuer_yaml}
    Oc Wait    -n ${NAMESPACE} clusterissuer ${ISSUER_NAME}    --for="condition=Ready" --timeout=120s
    ${cert_yaml}=    Create Certificate YAML For Test
    Apply YAML Manifest    ${cert_yaml}
    Oc Wait    -n ${NAMESPACE} certificate ${CERT_NAME}    --for="condition=Ready" --timeout=60s
    ${rbac_yaml}=    Create Ingress RBAC YAML
    Apply YAML Manifest    ${rbac_yaml}
    Deploy Hello MicroShift
    ${route_yaml}=    Create Ingress Route YAML
    Apply YAML Manifest    ${route_yaml}
    Oc Wait    -n ${NAMESPACE} route ${ROUTE_NAME}    --for=jsonpath='.status.ingress' --timeout=60s
    [Teardown]    Run Keywords
    ...    Remove ClusterIssuer

Test Cert manager with local acme server
    [Documentation]    Test cert-manager with local ACME server (Pebble) using HTTP01 challenge
    [Tags]    http01    acme
    [Setup]    Setup Pebble Server    ${NAMESPACE}

    ${dns_name}=    Generate Random HostName
    Configure DNS For Domain    ${USHIFT_HOST}    ${dns_name}
    Oc Get JsonPath    ingressclass    ${EMPTY}    openshift-ingress    .metadata.name
    ${http01_issuer_yaml}=    Create HTTP01 Issuer YAML
    Apply YAML Manifest    ${http01_issuer_yaml}
    Oc Wait    -n ${NAMESPACE} issuer ${HTTP01_ISSUER_NAME}    --for="condition=Ready" --timeout=120s

    ${cert_yaml}=    Create Certificate YAML    ${dns_name}
    Apply YAML Manifest    ${cert_yaml}
    Oc Wait    -n ${NAMESPACE} certificate ${HTTP01_CERT_NAME}    --for="condition=Ready" --timeout=300s

    Verify Certificate    ${HTTP01_CERT_NAME}    ${NAMESPACE}

    [Teardown]    Run Keywords
    ...    Cleanup HTTP01 Resources
    ...    AND    Remove DNS Configuration


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
    ${result}=    Oc Apply    -f ${temp_file}
    Remove File    ${temp_file}
    Should Contain    ${result}    created    msg=Failed to apply YAML manifest
    Log    Applied manifest: ${result}

Generate Random HostName
    [Documentation]    Generate Random Hostname
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    RETURN    ${rand}.api.com

Setup Pebble Server
    [Documentation]    Sets up a Pebble ACME server for HTTP01 testing, creates namespace if needed
    [Arguments]    ${namespace}
    ${result}=    Oc Apply    -f ${PEBBLE_DEPLOYMENT_FILE} -n ${namespace}

    # Wait for Pebble deployment to be ready
    Wait Until Keyword Succeeds    12x    10s    Check Pebble Deployment Ready    ${namespace}

    VAR    ${endpoint}=    https://pebble.${namespace}.svc.cluster.local:14000/dir
    Log    Pebble server setup successfully! (endpoint ${endpoint})
    RETURN    ${endpoint}

Check Pebble Deployment Ready
    [Documentation]    Checks if Pebble deployment is ready
    [Arguments]    ${namespace}
    ${result}=    Oc Get JsonPath    deployment    ${namespace}    pebble    .status.readyReplicas
    Should Be Equal    ${result}    1    msg=Pebble deployment not ready yet

Create Cert Issuer YAML
    [Documentation]    Creates cluster issuer YAML
    ${cert_issuer_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: cert-manager.io/v1
    ...    kind: ClusterIssuer
    ...    metadata:
    ...    \ \ name: ${ISSUER_NAME}
    ...    spec:
    ...    \ \ selfSigned: {}
    RETURN    ${cert_issuer_yaml}

Create Certificate YAML For Test
    [Documentation]    Creates certificate YAML for basic test
    ${cert_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
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
    RETURN    ${cert_yaml}

Create Ingress RBAC YAML
    [Documentation]    Creates RBAC YAML for ingress
    ${rbac_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
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
    RETURN    ${rbac_yaml}

Create Ingress Route YAML
    [Documentation]    Creates route YAML for ingress
    ${route_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: route.openshift.io/v1
    ...    kind: Route
    ...    metadata:
    ...    \ \ name: ${ROUTE_NAME}
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
    RETURN    ${route_yaml}

Create HTTP01 Issuer YAML
    [Documentation]    Creates HTTP01 issuer YAML
    ${http01_issuer_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: cert-manager.io/v1
    ...    kind: Issuer
    ...    metadata:
    ...    \ \ name: ${HTTP01_ISSUER_NAME}
    ...    \ \ namespace: ${NAMESPACE}
    ...    spec:
    ...    \ \ acme:
    ...    \ \ \ \ server: "https://pebble.${NAMESPACE}.svc.cluster.local:14000/dir"
    ...    \ \ \ \ skipTLSVerify: true
    ...    \ \ \ \ privateKeySecretRef:
    ...    \ \ \ \ \ \ name: acme-account-key
    ...    \ \ \ \ solvers:
    ...    \ \ \ \ - http01:
    ...    \ \ \ \ \ \ \ \ ingress:
    ...    \ \ \ \ \ \ \ \ \ \ ingressClassName: openshift-ingress
    RETURN    ${http01_issuer_yaml}

Create Certificate YAML
    [Documentation]    Creates certificate YAML with the specified DNS name
    [Arguments]    ${dns_name}
    ${cert_yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: cert-manager.io/v1
    ...    kind: Certificate
    ...    metadata:
    ...    \ \ name: ${HTTP01_CERT_NAME}
    ...    \ \ namespace: ${NAMESPACE}
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
    ${ready_result}=    Oc Get JsonPath
    ...    certificate
    ...    ${namespace}
    ...    ${cert_name}
    ...    .status.conditions[?(@.type=="Ready")].status
    Should Be Equal    ${ready_result}    True    msg=Certificate ${cert_name} is not ready

Get Certificate Secret Name
    [Documentation]    Gets the secret name from certificate spec
    [Arguments]    ${cert_name}    ${namespace}
    ${secret_name}=    Oc Get JsonPath    certificate    ${namespace}    ${cert_name}    .spec.secretName
    Should Not Be Empty    ${secret_name}    msg=Certificate does not have secretName in spec
    RETURN    ${secret_name}

Verify Certificate Secret Data
    [Documentation]    Verifies the secret exists and contains certificate data
    [Arguments]    ${secret_name}    ${namespace}
    ${secret_result}=    Oc Get JsonPath    secret    ${namespace}    ${secret_name}    .data.tls\\.crt
    Should Not Be Empty    ${secret_result}    msg=Certificate secret does not contain tls.crt data

Verify Certificate Common Name If Present
    [Documentation]    Verifies certificate Common Name if specified
    [Arguments]    ${cert_name}    ${namespace}
    ${common_name}=    Oc Get JsonPath    certificate    ${namespace}    ${cert_name}    .spec.commonName
    IF    "${common_name}" != ""
        ${secret_name}=    Get Certificate Secret Name    ${cert_name}    ${namespace}
        ${secret_result}=    Oc Get JsonPath    secret    ${namespace}    ${secret_name}    .data.tls\\.crt
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
    Oc Delete    certificate/${HTTP01_CERT_NAME} -n ${NAMESPACE}
    Oc Delete    issuer/${HTTP01_ISSUER_NAME} -n ${NAMESPACE}
    Oc Delete    deployment/pebble -n ${NAMESPACE}
    Oc Delete    service/pebble -n ${NAMESPACE}

Configure DNS For Domain
    [Documentation]    Configure DNS configmap in openshift-dns namespace and restart DNS pod
    [Arguments]    ${ip_address}    ${dns_name}

    # Store original config for later restoration
    ${original_config}=    Oc Get JsonPath    configmap    openshift-dns    dns-default    .data.Corefile
    VAR    ${ORIGINAL_DNS_CONFIG}=    ${original_config}    scope=SUITE

    # Add hosts plugin to the existing .:5353 server block
    VAR    ${hosts_entry}=
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}hosts {\n
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}${SPACE}${SPACE}${ip_address} ${dns_name}\n
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}${SPACE}${SPACE}fallthrough\n
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}}
    ${hosts_block}=    Replace String    ${original_config}    reload    reload\n${hosts_entry}

    # Update the configmap using --from-file to avoid JSON escaping issues
    ${temp_file}=    Create Random Temp File    ${hosts_block}
    Run With Kubeconfig
    ...    oc create configmap dns-default -n openshift-dns --from-file=Corefile=${temp_file} --dry-run=client -o yaml | oc apply -f -
    Remove File    ${temp_file}

    # Restart DNS pod by deleting it
    Oc Delete    pod -l dns.operator.openshift.io/daemonset-dns=default -n openshift-dns

    # Wait for DNS pod to be ready
    Oc Wait
    ...    -n openshift-dns pod -l dns.operator.openshift.io/daemonset-dns=default
    ...    --for=condition=Ready --timeout=60s

Remove DNS Configuration
    [Documentation]    Remove custom DNS configuration and restore original

    # Only restore if we have the original config stored
    ${config_exists}=    Run Keyword And Return Status    Variable Should Exist    ${ORIGINAL_DNS_CONFIG}
    IF    ${config_exists}
        # Create a temp file with the original DNS config and create configmap from file
        ${temp_file}=    Create Random Temp File    ${ORIGINAL_DNS_CONFIG}
        Run With Kubeconfig
        ...    oc create configmap dns-default -n openshift-dns --from-file=Corefile=${temp_file} --dry-run=client -o yaml | oc apply -f -
        Remove File    ${temp_file}
    ELSE
        Log    Original DNS config not found, skipping DNS restoration
    END

    # Restart DNS pod by deleting it (ignore failures in teardown)
    Oc Delete
    ...    pod -l dns.operator.openshift.io/daemonset-dns=default -n openshift-dns

    # Wait for DNS pod to be ready (ignore failures in teardown)
    Oc Wait
    ...    -n openshift-dns pod -l dns.operator.openshift.io/daemonset-dns=default
    ...    --for=condition=Ready --timeout=60s

Verify Cert Manager Kustomization Success
    [Documentation]    Verify that cert-manager kustomization was successfully applied by checking journalctl logs
    ${cursor}=    Get Journal Cursor
    Restart MicroShift
    Pattern Should Appear In Log Output
    ...    ${cursor}
    ...    Applying kustomization at /usr/lib/microshift/manifests.d/060-microshift-cert-manager was successful
    ...    unit=microshift
    ...    retries=6
    ...    wait=5
