*** Settings ***
Documentation       Test suite for generating certificates using cert-manager with YAML manifests

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Library             ../../resources/journalctl.py
Resource            ../../resources/common.resource
Resource            ../../resources/hosts.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/optional-config.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           cert-manager    certificates    tls


*** Variables ***
${CERT_NAME}                        test-certificate
${SECRET_NAME}                      test-cert-secret
${ISSUER_NAME}                      test-issuer
${CERT_COMMON_NAME}                 example.com
${CERT_DNS_NAME}                    example.com
${ROUTE_NAME}                       hello-app
${CERT_ISSUER_YAML}                 SEPARATOR=\n
...                                 ---
...                                 apiVersion: cert-manager.io/v1
...                                 kind: ClusterIssuer
...                                 metadata:
...                                 \ \ name: ${ISSUER_NAME}
...                                 spec:
...                                 \ \ selfSigned: {}

${HTTP01_ISSUER_NAME}               letsencrypt-http01
${HTTP01_CERT_NAME}                 cert-from-${HTTP01_ISSUER_NAME}
${HTTP01_SECRET_NAME}               ${HTTP01_CERT_NAME}
${PEBBLE_DEPLOYMENT_FILE}           ./assets/cert-manager/pebble-server.yaml
${HOSTSFILE_ENABLED}                SEPARATOR=\n
...                                 ---
...                                 dns:
...                                 \ \ hosts:
...                                 \ \ \ \ status: Enabled

${TRUST_MANAGER_BUNDLE_NAME}        test-trust-bundle
${TRUST_MANAGER_OPERATOR_NS}        cert-manager-operator
${TRUST_MANAGER_NS}                 cert-manager
${TRUST_MANAGER_DEPLOYMENT}         cert-manager-operator-controller-manager
${TRUST_MANAGER_MANIFESTS_DIR}      /etc/microshift/manifests.d/trust-manager


*** Test Cases ***
Create Ingress route with Custom certificate
    [Documentation]    Create route with a custom certificate
    [Setup]    Run Keywords
    Verify Cert Manager Kustomization Success
    ${cert_issuer_yaml}=    Create Cert Issuer YAML
    Apply YAML Manifest    ${cert_issuer_yaml}
    Oc Wait    -n ${NAMESPACE} clusterissuer ${ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}
    ${cert_yaml}=    Create Certificate YAML For Test
    Apply YAML Manifest    ${cert_yaml}
    Oc Wait    -n ${NAMESPACE} certificate ${CERT_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}
    ${rbac_yaml}=    Create Ingress RBAC YAML
    Apply YAML Manifest    ${rbac_yaml}
    Deploy Hello MicroShift
    ${route_yaml}=    Create Ingress Route YAML
    Apply YAML Manifest    ${route_yaml}
    Oc Wait    -n ${NAMESPACE} route ${ROUTE_NAME}
    ...    --for=jsonpath='.status.ingress' --timeout=${DEFAULT_WAIT_TIMEOUT}
    [Teardown]    Run Keywords
    ...    Remove ClusterIssuer

Test Cert manager with local acme server
    [Documentation]    Test cert-manager with local ACME server (Pebble) using HTTP01 challenge
    [Tags]    http01    acme
    [Setup]    Setup Pebble Server    ${NAMESPACE}

    ${dns_name}=    Generate Random HostName
    Setup DNS For Test    ${USHIFT_HOST}    ${dns_name}
    Oc Get JsonPath    ingressclass    ${EMPTY}    openshift-ingress    .metadata.name
    ${http01_issuer_yaml}=    Create HTTP01 Issuer YAML
    Apply YAML Manifest    ${http01_issuer_yaml}
    Oc Wait    -n ${NAMESPACE} issuer ${HTTP01_ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${cert_yaml}=    Create Certificate YAML    ${dns_name}
    Apply YAML Manifest    ${cert_yaml}
    Oc Wait    -n ${NAMESPACE} certificate ${HTTP01_CERT_NAME}    --for="condition=Ready" --timeout=300s

    Verify Certificate    ${HTTP01_CERT_NAME}    ${NAMESPACE}

    [Teardown]    Run Keywords
    ...    Cleanup HTTP01 Resources
    ...    AND    Cleanup DNS For Test    ${dns_name}

Trust Manager Deployment
    [Documentation]    Verify trust-manager can be enabled and deploys successfully
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager
    Labeled Pod Should Be Ready    app.kubernetes.io/name=cert-manager-trust-manager    ns=${TRUST_MANAGER_NS}
    ${status}=    Oc Get JsonPath    trustmanager    ${EMPTY}    cluster
    ...    .status.conditions[?(@.type=="Ready")].status
    Should Be Equal    ${status}    True    msg=TrustManager CR is not ready
    [Teardown]    Disable Trust Manager

Trust Manager Bundle Creates ConfigMap
    [Documentation]    Verify trust-manager Bundle CR syncs a CA cert into a ConfigMap
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager

    Create CA Secret For Trust Manager
    ${bundle_yaml}=    Create Trust Bundle From Source Secret YAML
    Apply Trust Manager YAML    ${bundle_yaml}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[0].reason}'=Synced --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    BEGIN CERTIFICATE    msg=ConfigMap does not contain CA certificate data

    [Teardown]    Run Keywords
    ...    Cleanup Trust Bundle
    ...    AND    Run With Kubeconfig    oc delete secret ca-source-secret -n ${TRUST_MANAGER_NS} --ignore-not-found
    ...    AND    Disable Trust Manager

Trust Manager Bundle With Cert Manager CA
    [Documentation]    Verify trust-manager Bundle can use a cert-manager CA secret as a source
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager

    ${issuer_yaml}=    Create Cert Issuer YAML
    Apply Trust Manager YAML    ${issuer_yaml}
    Oc Wait    -n ${NAMESPACE} clusterissuer ${ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${ca_cert_yaml}=    Create CA Certificate YAML
    Apply Trust Manager YAML    ${ca_cert_yaml}
    Oc Wait    -n ${TRUST_MANAGER_NS} certificate ca-certificate
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${bundle_yaml}=    Create Trust Bundle From Secret YAML
    Apply Trust Manager YAML    ${bundle_yaml}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[0].reason}'=Synced --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    BEGIN CERTIFICATE    msg=ConfigMap does not contain CA certificate data

    [Teardown]    Run Keywords
    ...    Cleanup Trust Bundle
    ...    AND    Oc Delete    certificate/ca-certificate -n ${TRUST_MANAGER_NS}
    ...    AND    Remove ClusterIssuer
    ...    AND    Disable Trust Manager


*** Keywords ***
Setup
    [Documentation]    Setup cert-manager suite with only its required optionals
    Setup Suite
    Setup MicroShift With Optionals    060-microshift-cert-manager
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=SUITE

Teardown
    [Documentation]    Restore config and teardown suite
    Teardown MicroShift With Optionals
    Teardown Suite With Namespace

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
    ...    --for=condition=Ready --timeout=${DEFAULT_WAIT_TIMEOUT}

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
    ...    --for=condition=Ready --timeout=${DEFAULT_WAIT_TIMEOUT}

Verify Cert Manager Kustomization Success
    [Documentation]    Verify that cert-manager kustomization was successfully applied by checking journalctl logs
    ${cursor}=    Get Journal Cursor
    Restart MicroShift
    Pattern Should Appear In Log Output
    ...    ${cursor}
    ...    Applying kustomization at /usr/lib/microshift/manifests.d/060-microshift-cert-manager was successful
    ...    unit=microshift
    ...    retries=30
    ...    wait=10

Resolve Host From Pod
    [Documentation]    Resolve host from pod
    [Arguments]    ${hostname}
    Wait Until Keyword Succeeds    40x    2s
    ...    Router Should Resolve Hostname    ${hostname}

Router Should Resolve Hostname
    [Documentation]    Check if the router pod resolves the given hostname
    [Arguments]    ${hostname}
    ${fuse_device}=    Oc Exec    router-default    nslookup ${hostname}    openshift-ingress    deployment
    Should Contain    ${fuse_device}    Name:    ${hostname}

Setup DNS For Test
    [Documentation]    Setup DNS using CoreDNS hosts feature
    [Arguments]    ${ip_address}    ${dns_name}

    Add Entry To Hosts    ${ip_address}    ${dns_name}    /etc/hosts
    Drop In MicroShift Config    ${HOSTSFILE_ENABLED}    20-dns
    Restart MicroShift
    Wait For MicroShift Healthcheck Success

Cleanup DNS For Test
    [Documentation]    Cleanup DNS configuration
    [Arguments]    ${dns_name}

    Remove Entry From Hosts    ${dns_name}
    Remove Drop In MicroShift Config    20-dns
    Restart MicroShift

Enable Trust Manager
    [Documentation]    Deploy trust-manager by creating a TrustManager CR via manifests.d
    ...    and restarting MicroShift. The UNSUPPORTED_ADDON_FEATURES=TrustManager=true
    ...    feature gate is already set in the system cert-manager kustomization.
    Create Trust Manager CR Manifests
    Restart MicroShift
    Wait Until Keyword Succeeds    30x    10s
    ...    Labeled Pod Should Be Ready    app.kubernetes.io/name=cert-manager-trust-manager    ns=${TRUST_MANAGER_NS}

Disable Trust Manager
    [Documentation]    Remove the TrustManager CR manifests.d and restart MicroShift.
    Run With Kubeconfig    oc delete trustmanager cluster --ignore-not-found
    Run With Kubeconfig    oc delete bundle ${TRUST_MANAGER_BUNDLE_NAME} --ignore-not-found
    Run With Kubeconfig    oc delete deployment trust-manager -n ${TRUST_MANAGER_NS} --ignore-not-found
    Remove Trust Manager CR Manifests
    Restart MicroShift
    Wait Until Keyword Succeeds    12x    10s
    ...    Trust Manager Pod Should Not Exist

Trust Manager Pod Should Not Exist
    [Documentation]    Verify trust-manager pod no longer exists in cert-manager namespace
    ${output}=    Run With Kubeconfig
    ...    oc get pods -n ${TRUST_MANAGER_NS} -l app.kubernetes.io/name\=cert-manager-trust-manager --no-headers
    ...    allow_fail=True
    Should Be Empty    ${output}    msg=trust-manager pod still exists

Create Trust Manager CR Manifests
    [Documentation]    Create the manifests.d kustomization with the TrustManager CR
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    mkdir -p ${TRUST_MANAGER_MANIFESTS_DIR}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    ${rc}    0
    ${kustomization}=    CATENATE    SEPARATOR=\n
    ...    apiVersion: kustomize.config.k8s.io/v1beta1
    ...    kind: Kustomization
    ...    resources:
    ...    \ \ - trust-manager-cr.yaml
    Upload String To File    ${kustomization}    ${TRUST_MANAGER_MANIFESTS_DIR}/kustomization.yaml
    ${tm_cr}=    CATENATE    SEPARATOR=\n
    ...    apiVersion: operator.openshift.io/v1alpha1
    ...    kind: TrustManager
    ...    metadata:
    ...    \ \ name: cluster
    ...    spec:
    ...    \ \ trustManagerConfig: {}
    Upload String To File    ${tm_cr}    ${TRUST_MANAGER_MANIFESTS_DIR}/trust-manager-cr.yaml

Remove Trust Manager CR Manifests
    [Documentation]    Remove the trust-manager manifests.d directory
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rm -rf ${TRUST_MANAGER_MANIFESTS_DIR}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    ${rc}    0

Create CA Secret For Trust Manager
    [Documentation]    Generate a self-signed CA cert locally and create a secret in the trust namespace
    ${cert_file}=    Create Random Temp File
    ${result}=    Process.Run Process
    ...    openssl    req    -x509    -newkey    ec    -pkeyopt    ec_paramgen_curve:prime256v1
    ...    -nodes    -keyout    /dev/null    -out    ${cert_file}    -days    365
    ...    -subj    /CN\=test-ca.example.com
    ...    stderr=STDOUT
    Should Be Equal As Integers    ${result.rc}    0
    Run With Kubeconfig
    ...    oc create secret generic ca-source-secret -n ${TRUST_MANAGER_NS} --from-file=tls.crt=${cert_file}
    Remove File    ${cert_file}

Create Trust Bundle From Source Secret YAML
    [Documentation]    Creates a Bundle CR YAML sourced from a manually created secret in the trust namespace
    ${yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: trust.cert-manager.io/v1alpha1
    ...    kind: Bundle
    ...    metadata:
    ...    \ \ name: ${TRUST_MANAGER_BUNDLE_NAME}
    ...    spec:
    ...    \ \ sources:
    ...    \ \ \ \ - secret:
    ...    \ \ \ \ \ \ \ \ name: ca-source-secret
    ...    \ \ \ \ \ \ \ \ key: tls.crt
    ...    \ \ target:
    ...    \ \ \ \ configMap:
    ...    \ \ \ \ \ \ key: ca-bundle.crt
    ...    \ \ \ \ namespaceSelector:
    ...    \ \ \ \ \ \ matchLabels:
    ...    \ \ \ \ \ \ \ \ kubernetes.io/metadata.name: ${NAMESPACE}
    RETURN    ${yaml}

Create CA Certificate YAML
    [Documentation]    Creates a cert-manager CA Certificate in the trust namespace
    ${yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: cert-manager.io/v1
    ...    kind: Certificate
    ...    metadata:
    ...    \ \ name: ca-certificate
    ...    \ \ namespace: ${TRUST_MANAGER_NS}
    ...    spec:
    ...    \ \ isCA: true
    ...    \ \ commonName: test-ca.example.com
    ...    \ \ secretName: ca-certificate-secret
    ...    \ \ issuerRef:
    ...    \ \ \ \ name: ${ISSUER_NAME}
    ...    \ \ \ \ kind: ClusterIssuer
    RETURN    ${yaml}

Create Trust Bundle From Secret YAML
    [Documentation]    Creates a Bundle CR YAML sourced from a cert-manager CA secret in the trust namespace
    ${yaml}=    CATENATE    SEPARATOR=\n
    ...    ---
    ...    apiVersion: trust.cert-manager.io/v1alpha1
    ...    kind: Bundle
    ...    metadata:
    ...    \ \ name: ${TRUST_MANAGER_BUNDLE_NAME}
    ...    spec:
    ...    \ \ sources:
    ...    \ \ \ \ - secret:
    ...    \ \ \ \ \ \ \ \ name: ca-certificate-secret
    ...    \ \ \ \ \ \ \ \ key: tls.crt
    ...    \ \ target:
    ...    \ \ \ \ configMap:
    ...    \ \ \ \ \ \ key: ca-bundle.crt
    ...    \ \ \ \ namespaceSelector:
    ...    \ \ \ \ \ \ matchLabels:
    ...    \ \ \ \ \ \ \ \ kubernetes.io/metadata.name: ${NAMESPACE}
    RETURN    ${yaml}

Apply Trust Manager YAML
    [Documentation]    Apply YAML manifest, allowing both created and configured/unchanged results
    [Arguments]    ${yaml_content}
    ${temp_file}=    Create Random Temp File    ${yaml_content}
    ${result}=    Oc Apply    -f ${temp_file}
    Remove File    ${temp_file}
    Log    Applied manifest: ${result}

Cleanup Trust Bundle
    [Documentation]    Remove the test trust-manager Bundle CR and its target ConfigMap
    Run With Kubeconfig    oc delete bundle ${TRUST_MANAGER_BUNDLE_NAME} --ignore-not-found
    Run With Kubeconfig    oc delete configmap ${TRUST_MANAGER_BUNDLE_NAME} -n ${NAMESPACE} --ignore-not-found
