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
${CLUSTER_ISSUER_TMPL}              ./assets/cert-manager/cluster-issuer.yaml.template
${CERTIFICATE_TMPL}                 ./assets/cert-manager/certificate.yaml.template
${INGRESS_RBAC_TMPL}                ./assets/cert-manager/ingress-rbac.yaml.template
${INGRESS_ROUTE_TMPL}               ./assets/cert-manager/ingress-route.yaml.template
${HTTP01_ISSUER_TMPL}               ./assets/cert-manager/http01-issuer.yaml.template
${HTTP01_CERTIFICATE_TMPL}          ./assets/cert-manager/http01-certificate.yaml.template
${TRUST_BUNDLE_SRC_SECRET_TMPL}     ./assets/cert-manager/trust-bundle-source-secret.yaml.template
${CA_CERTIFICATE_TMPL}              ./assets/cert-manager/ca-certificate.yaml.template
${TRUST_BUNDLE_SECRET_TMPL}         ./assets/cert-manager/trust-bundle-secret.yaml.template

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
${TRUST_MANAGER_NS}                 cert-manager
${TRUST_MANAGER_CR_FILE}            ./assets/cert-manager/trust-manager-cr.yaml
${TRUST_MANAGER_DEFAULT_CA_CR}      ./assets/cert-manager/trust-manager-cr-default-ca.yaml
${CM_SOURCE_BUNDLE_TMPL}            ./assets/cert-manager/trust-bundle-configmap-source.yaml.template
${DEFAULT_CAS_BUNDLE_TMPL}          ./assets/cert-manager/trust-bundle-default-cas.yaml.template
${CA_SOURCE_CONFIGMAP_NAME}         ca-configmap-source
${CERT_MANAGER_OPERATOR_NS}         cert-manager-operator
${OPERATOR_TRUSTED_CA_CM_NAME}      cert-manager-operator-trusted-ca-bundle


*** Test Cases ***
Create Ingress route with Custom certificate
    [Documentation]    Create route with a custom certificate
    [Setup]    Run Keywords
    Verify Cert Manager Kustomization Success
    Apply Template    ${CLUSTER_ISSUER_TMPL}
    Oc Wait    -n ${NAMESPACE} clusterissuer ${ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}
    Apply Template    ${CERTIFICATE_TMPL}
    Oc Wait    -n ${NAMESPACE} certificate ${CERT_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}
    Apply Template    ${INGRESS_RBAC_TMPL}
    Deploy Hello MicroShift
    Apply Template    ${INGRESS_ROUTE_TMPL}
    Oc Wait    -n ${NAMESPACE} route ${ROUTE_NAME}
    ...    --for=jsonpath='.status.ingress' --timeout=${DEFAULT_WAIT_TIMEOUT}
    [Teardown]    Run Keywords
    ...    Remove ClusterIssuer

Test Cert manager with local acme server
    [Documentation]    Test cert-manager with local ACME server (Pebble) using HTTP01 challenge
    [Tags]    http01    acme
    [Setup]    Setup Pebble Server    ${NAMESPACE}

    ${DNS_NAME}=    Generate Random HostName
    VAR    ${DNS_NAME}=    ${DNS_NAME}    scope=TEST
    Setup DNS For Test    ${USHIFT_HOST}    ${DNS_NAME}
    Oc Get JsonPath    ingressclass    ${EMPTY}    openshift-ingress    .metadata.name
    Apply Template    ${HTTP01_ISSUER_TMPL}
    Oc Wait    -n ${NAMESPACE} issuer ${HTTP01_ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}

    Apply Template    ${HTTP01_CERTIFICATE_TMPL}
    Oc Wait    -n ${NAMESPACE} certificate ${HTTP01_CERT_NAME}    --for="condition=Ready" --timeout=300s

    Verify Certificate    ${HTTP01_CERT_NAME}    ${NAMESPACE}

    [Teardown]    Run Keywords
    ...    Cleanup HTTP01 Resources
    ...    AND    Cleanup DNS For Test    ${DNS_NAME}

Trust Manager Deployment
    [Documentation]    Verify trust-manager can be enabled and deploys successfully
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager
    Wait Until Keyword Succeeds    30x    10s
    ...    TrustManager CR Should Be Ready
    [Teardown]    Disable Trust Manager

Trust Manager Bundle Creates ConfigMap
    [Documentation]    Verify trust-manager Bundle CR syncs a CA cert into a ConfigMap
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager

    Create CA Secret For Trust Manager
    Apply Template    ${TRUST_BUNDLE_SRC_SECRET_TMPL}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[?(@.type=="Synced")].status}'=True --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    BEGIN CERTIFICATE    msg=ConfigMap does not contain CA certificate data

    [Teardown]    Run Keywords
    ...    Cleanup Trust Bundle
    ...    AND    Oc Delete    secret ca-source-secret -n ${TRUST_MANAGER_NS} --ignore-not-found
    ...    AND    Disable Trust Manager

Trust Manager Bundle With Cert Manager CA
    [Documentation]    Verify trust-manager Bundle can use a cert-manager CA secret as a source
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager

    Setup Trust Bundle With Cert Manager CA

    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    BEGIN CERTIFICATE    msg=ConfigMap does not contain CA certificate data

    [Teardown]    Run Keywords
    ...    Cleanup Trust Bundle
    ...    AND    Oc Delete    certificate/ca-certificate -n ${TRUST_MANAGER_NS} --ignore-not-found
    ...    AND    Remove ClusterIssuer
    ...    AND    Disable Trust Manager

Trust Manager Bundle With Pebble ACME CA
    [Documentation]    Verify trust-manager can distribute a Pebble ACME root CA via Bundle CR
    [Tags]    trust-manager    acme
    [Setup]    Run Keywords
    ...    Enable Trust Manager
    ...    AND    Setup Pebble Server    ${NAMESPACE}

    ${DNS_NAME}=    Generate Random HostName
    VAR    ${DNS_NAME}=    ${DNS_NAME}    scope=TEST
    Setup DNS For Test    ${USHIFT_HOST}    ${DNS_NAME}

    Apply Template    ${HTTP01_ISSUER_TMPL}
    Oc Wait    -n ${NAMESPACE} issuer ${HTTP01_ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}

    Apply Template    ${HTTP01_CERTIFICATE_TMPL}
    Oc Wait    -n ${NAMESPACE} certificate ${HTTP01_CERT_NAME}    --for="condition=Ready" --timeout=300s

    ${pebble_ca}=    Fetch Pebble Root CA    ${NAMESPACE}
    ${cert_file}=    Create Random Temp File    ${pebble_ca}
    Oc Create    secret generic ca-source-secret -n ${TRUST_MANAGER_NS} --from-file=tls.crt=${cert_file}
    Remove File    ${cert_file}

    Apply Template    ${TRUST_BUNDLE_SRC_SECRET_TMPL}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[?(@.type=="Synced")].status}'=True --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    ${pebble_ca}    msg=ConfigMap does not contain Pebble CA certificate data

    [Teardown]    Run Keywords
    ...    Cleanup Trust Bundle
    ...    AND    Oc Delete    secret ca-source-secret -n ${TRUST_MANAGER_NS} --ignore-not-found
    ...    AND    Cleanup HTTP01 Resources
    ...    AND    Cleanup DNS For Test    ${DNS_NAME}
    ...    AND    Disable Trust Manager

Trust Manager Survives MicroShift Restart
    [Documentation]    Verify trust-manager Bundle CR and ConfigMap persist after MicroShift restart
    [Tags]    trust-manager    restart
    [Setup]    Enable Trust Manager

    Setup Trust Bundle With Cert Manager CA
    ${cm_data_before}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data_before}    BEGIN CERTIFICATE

    Restart MicroShift
    Wait For MicroShift Healthcheck Success
    Wait For Trust Manager After Restart

    ${cm_data_after}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Be Equal    ${cm_data_before}    ${cm_data_after}    msg=ConfigMap data changed after restart

    [Teardown]    Run Keywords
    ...    Cleanup Trust Bundle
    ...    AND    Oc Delete    certificate/ca-certificate -n ${TRUST_MANAGER_NS} --ignore-not-found
    ...    AND    Remove ClusterIssuer
    ...    AND    Disable Trust Manager

Trust Manager Bundle Injects Into Labeled Namespace
    [Documentation]    Verify trust-manager injects a Bundle (sourced from a ConfigMap) into a
    ...    namespace labeled with trust.cert-manager.io/inject=true, following the OpenShift
    ...    best-practice namespaceSelector pattern, and does not inject into unlabeled namespaces.
    [Tags]    trust-manager
    [Setup]    Enable Trust Manager

    ${INJECT_NS}=    Create Unique Namespace
    VAR    ${INJECT_NS}=    ${INJECT_NS}    scope=TEST
    Run With Kubeconfig    oc label namespace ${INJECT_NS} trust.cert-manager.io/inject=true --overwrite

    Create CA ConfigMap For Trust Manager
    Apply Template    ${CM_SOURCE_BUNDLE_TMPL}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[?(@.type=="Synced")].status}'=True --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${source_ca}=    Oc Get JsonPath
    ...    configmap
    ...    ${TRUST_MANAGER_NS}
    ...    ${CA_SOURCE_CONFIGMAP_NAME}
    ...    .data.ca\\.crt
    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${INJECT_NS}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    ${source_ca}    msg=Bundle was not injected into the labeled namespace

    ConfigMap Should Not Exist    ${TRUST_MANAGER_BUNDLE_NAME}    ${NAMESPACE}

    [Teardown]    Run Keywords
    ...    Cleanup Inject Label Resources    ${INJECT_NS}
    ...    AND    Disable Trust Manager

Trust Manager Bundle With Default CAs
    [Documentation]    Verify trust-manager can distribute CA certificates via the useDefaultCAs
    ...    source. On MicroShift there is no cluster-network-operator to populate the operator's
    ...    trusted CA bundle, so the cert-manager-operator-trusted-ca-bundle ConfigMap must be
    ...    created manually before enabling the default CA package.
    [Tags]    trust-manager
    [Setup]    Run Keywords
    ...    Create Operator Trusted CA Bundle ConfigMap
    ...    AND    Enable Trust Manager With Default CA Package

    Apply Template    ${DEFAULT_CAS_BUNDLE_TMPL}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[?(@.type=="Synced")].status}'=True --timeout=${DEFAULT_WAIT_TIMEOUT}

    ${source_ca}=    Oc Get JsonPath
    ...    configmap
    ...    ${CERT_MANAGER_OPERATOR_NS}
    ...    ${OPERATOR_TRUSTED_CA_CM_NAME}
    ...    .data.ca-bundle\\.crt
    ${cm_data}=    Oc Get JsonPath
    ...    configmap
    ...    ${NAMESPACE}
    ...    ${TRUST_MANAGER_BUNDLE_NAME}
    ...    .data.ca-bundle\\.crt
    Should Contain    ${cm_data}    ${source_ca}    msg=ConfigMap does not contain default CA certificate data

    [Teardown]    Teardown Default CA Package Test


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
    Oc Delete    clusterissuer/${ISSUER_NAME} --ignore-not-found

Apply Template
    [Documentation]    Generate YAML from template and apply to cluster
    [Arguments]    ${template_file}
    ${tmp}=    Create Random Temp File
    Generate File From Template    ${template_file}    ${tmp}
    TRY
        Oc Apply    -f ${tmp}
    FINALLY
        Remove File    ${tmp}
    END

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
    Oc Delete    certificate/${HTTP01_CERT_NAME} -n ${NAMESPACE} --ignore-not-found
    Oc Delete    issuer/${HTTP01_ISSUER_NAME} -n ${NAMESPACE} --ignore-not-found
    Oc Delete    deployment/pebble -n ${NAMESPACE} --ignore-not-found
    Oc Delete    service/pebble -n ${NAMESPACE} --ignore-not-found
    Oc Delete    configmap/pebble -n ${NAMESPACE} --ignore-not-found

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
    [Documentation]    Deploy trust-manager by applying the TrustManager CR directly.
    ...    The UNSUPPORTED_ADDON_FEATURES=TrustManager=true feature gate is already
    ...    set in the system cert-manager kustomization.
    Oc Apply    -f ${TRUST_MANAGER_CR_FILE}
    Wait Until Keyword Succeeds    30x    10s
    ...    Labeled Pod Should Be Ready    app.kubernetes.io/name=cert-manager-trust-manager    ns=${TRUST_MANAGER_NS}

Enable Trust Manager With Default CA Package
    [Documentation]    Deploy trust-manager with the default CA package enabled so that Bundles
    ...    can use the "useDefaultCAs: true" source. Requires the operator trusted CA bundle
    ...    ConfigMap to already exist. The default policy is Disabled.
    Oc Apply    -f ${TRUST_MANAGER_DEFAULT_CA_CR}
    Wait Until Keyword Succeeds    30x    10s
    ...    Labeled Pod Should Be Ready    app.kubernetes.io/name=cert-manager-trust-manager    ns=${TRUST_MANAGER_NS}
    Wait Until Keyword Succeeds    30x    10s
    ...    TrustManager CR Should Be Ready

Disable Trust Manager
    [Documentation]    Remove the TrustManager CR and leftover Bundle.
    ...    The operator keeps the trust-manager deployment running while the
    ...    TrustManager feature gate is enabled, so we do not assert pod absence.
    Oc Delete    trustmanager cluster --ignore-not-found
    Oc Delete    bundle ${TRUST_MANAGER_BUNDLE_NAME} --ignore-not-found

Teardown Default CA Package Test
    [Documentation]    Restore the default CA package policy to Disabled and wait for the
    ...    trust-manager deployment rollout to finish before removing the CR and test
    ...    ConfigMaps, so the operand is left in a clean state for subsequent tests.
    Oc Apply    -f ${TRUST_MANAGER_CR_FILE}
    Wait For Trust Manager Rollout
    Cleanup Trust Bundle
    Oc Delete    configmap ${OPERATOR_TRUSTED_CA_CM_NAME} -n ${CERT_MANAGER_OPERATOR_NS} --ignore-not-found
    Disable Trust Manager

Wait For Trust Manager Rollout
    [Documentation]    Wait for the trust-manager deployment rollout to complete so the operand
    ...    reaches a steady state after a default CA package policy change. Tolerant of failure
    ...    so a transient issue does not block the remaining teardown cleanup.
    ${name}=    Run With Kubeconfig
    ...    oc get deployment -n ${TRUST_MANAGER_NS} -l app.kubernetes.io/name=cert-manager-trust-manager -o jsonpath={.items[0].metadata.name}
    ...    allow_fail=True
    IF    "${name}" != "${EMPTY}"
        Run With Kubeconfig
        ...    oc rollout status deployment/${name} -n ${TRUST_MANAGER_NS} --timeout=${DEFAULT_WAIT_TIMEOUT}
        ...    allow_fail=True
    END

Create CA Secret For Trust Manager
    [Documentation]    Generate a self-signed CA cert locally and create a secret in the trust namespace
    ${cert_file}=    Create Random Temp File
    ${result}=    Process.Run Process
    ...    openssl    req    -x509    -newkey    ec    -pkeyopt    ec_paramgen_curve:prime256v1
    ...    -nodes    -keyout    /dev/null    -out    ${cert_file}    -days    365
    ...    -subj    /CN\=test-ca.example.com
    ...    stderr=STDOUT
    Should Be Equal As Integers    ${result.rc}    0
    Oc Create    secret generic ca-source-secret -n ${TRUST_MANAGER_NS} --from-file=tls.crt=${cert_file}
    Remove File    ${cert_file}

Create CA ConfigMap For Trust Manager
    [Documentation]    Generate a self-signed CA cert locally and create a ConfigMap in the trust namespace
    ${cert_file}=    Create Random Temp File
    ${result}=    Process.Run Process
    ...    openssl    req    -x509    -newkey    ec    -pkeyopt    ec_paramgen_curve:prime256v1
    ...    -nodes    -keyout    /dev/null    -out    ${cert_file}    -days    365
    ...    -subj    /CN\=test-configmap-ca.example.com
    ...    stderr=STDOUT
    Should Be Equal As Integers    ${result.rc}    0
    Oc Create    configmap ${CA_SOURCE_CONFIGMAP_NAME} -n ${TRUST_MANAGER_NS} --from-file=ca.crt=${cert_file}
    Remove File    ${cert_file}

Create Operator Trusted CA Bundle ConfigMap
    [Documentation]    Create the cert-manager-operator-trusted-ca-bundle ConfigMap that the
    ...    operator injects into trust-manager as the default CA package. On MicroShift there is
    ...    no cluster-network-operator to populate it, so it is created manually here.
    ${cert_file}=    Create Random Temp File
    ${result}=    Process.Run Process
    ...    openssl    req    -x509    -newkey    ec    -pkeyopt    ec_paramgen_curve:prime256v1
    ...    -nodes    -keyout    /dev/null    -out    ${cert_file}    -days    365
    ...    -subj    /CN\=test-default-ca.example.com
    ...    stderr=STDOUT
    Should Be Equal As Integers    ${result.rc}    0
    Oc Create
    ...    configmap ${OPERATOR_TRUSTED_CA_CM_NAME} -n ${CERT_MANAGER_OPERATOR_NS} --from-file=ca-bundle.crt=${cert_file}
    Remove File    ${cert_file}

TrustManager CR Should Be Ready
    [Documentation]    Check TrustManager CR has Ready=True status condition
    ${status}=    Oc Get JsonPath    trustmanager    ${EMPTY}    cluster
    ...    .status.conditions[?(@.type=="Ready")].status
    Should Be Equal    ${status}    True    msg=TrustManager CR is not ready

ConfigMap Should Not Exist
    [Documentation]    Fails unless the ConfigMap is reported NotFound, so that a transient
    ...    API error (which also exits non-zero) is not mistaken for proof of absence.
    ...    Mirrors the error handling of Namespace Should Not Exist.
    [Arguments]    ${name}    ${ns}
    ${stdout}    ${rc}=    Run With Kubeconfig    oc get configmap ${name} -n ${ns}
    ...    allow_fail=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0    msg=ConfigMap ${name} unexpectedly exists in namespace ${ns}
    Should Contain    ${stdout}    NotFound
    ...    msg=oc get configmap ${name} -n ${ns} failed for a reason other than NotFound: ${stdout}

Cleanup Trust Bundle
    [Documentation]    Remove the test trust-manager Bundle CR and its target ConfigMap
    Oc Delete    bundle ${TRUST_MANAGER_BUNDLE_NAME} --ignore-not-found
    Oc Delete    configmap ${TRUST_MANAGER_BUNDLE_NAME} -n ${NAMESPACE} --ignore-not-found

Cleanup Inject Label Resources
    [Documentation]    Remove the inject-label test Bundle, source ConfigMap, and labeled namespace.
    ...    Deleting the Bundle lets trust-manager garbage-collect the injected target ConfigMap.
    [Arguments]    ${inject_ns}
    Oc Delete    bundle ${TRUST_MANAGER_BUNDLE_NAME} --ignore-not-found
    Oc Delete    configmap ${CA_SOURCE_CONFIGMAP_NAME} -n ${TRUST_MANAGER_NS} --ignore-not-found
    Remove Namespace    ${inject_ns}

Setup Trust Bundle With Cert Manager CA
    [Documentation]    Create a self-signed CA via cert-manager and configure a trust Bundle
    Apply Template    ${CLUSTER_ISSUER_TMPL}
    Oc Wait    -n ${NAMESPACE} clusterissuer ${ISSUER_NAME}
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}
    Apply Template    ${CA_CERTIFICATE_TMPL}
    Oc Wait    -n ${TRUST_MANAGER_NS} certificate ca-certificate
    ...    --for="condition=Ready" --timeout=${DEFAULT_WAIT_TIMEOUT}
    Apply Template    ${TRUST_BUNDLE_SECRET_TMPL}
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[?(@.type=="Synced")].status}'=True --timeout=${DEFAULT_WAIT_TIMEOUT}

Wait For Trust Manager After Restart
    [Documentation]    Wait for trust-manager components to be ready after a MicroShift restart
    Wait Until Keyword Succeeds    30x    10s
    ...    Labeled Pod Should Be Ready    app.kubernetes.io/name=cert-manager-trust-manager    ns=${TRUST_MANAGER_NS}
    Wait Until Keyword Succeeds    30x    10s
    ...    TrustManager CR Should Be Ready
    Oc Wait    bundle ${TRUST_MANAGER_BUNDLE_NAME}
    ...    --for=jsonpath='{.status.conditions[?(@.type=="Synced")].status}'=True --timeout=${DEFAULT_WAIT_TIMEOUT}

Fetch Pebble Root CA
    [Documentation]    Fetch the root CA certificate from Pebble management API via port-forward
    [Arguments]    ${namespace}
    ${port_fwd}=    Process.Start Process
    ...    oc    port-forward    deployment/pebble    15000:15000
    ...    -n    ${namespace}    --kubeconfig    ${KUBECONFIG}
    ...    stderr=STDOUT
    Sleep    5s
    ${running}=    Process.Is Process Running    ${port_fwd}
    IF    not ${running}
        ${fwd_result}=    Process.Wait For Process    ${port_fwd}
        Fail    Port-forward failed to start: ${fwd_result.stdout}
    END
    TRY
        ${pem}=    Wait Until Keyword Succeeds    10x    3s
        ...    Pebble Root CA Should Be Available
        RETURN    ${pem}
    FINALLY
        Process.Terminate Process    ${port_fwd}
    END

Pebble Root CA Should Be Available
    [Documentation]    Check that the Pebble management API returns a valid CA certificate
    ${result}=    Process.Run Process    curl    -skf    https://localhost:15000/roots/0
    ...    stderr=STDOUT
    Should Be Equal As Integers    ${result.rc}    0    msg=Failed to fetch Pebble root CA
    Should Contain    ${result.stdout}    BEGIN CERTIFICATE    msg=Pebble root CA response missing certificate data
    RETURN    ${result.stdout}
