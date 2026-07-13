*** Settings ***
Documentation       Router TLS and certificate configuration tests (disruptive)
...                 Migrated from openshift-tests-private:
...                 OCP-80508, OCP-80510, OCP-80514, OCP-80517

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/router.resource
Variables           configs.py
Library             configs.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${BASE_DOMAIN}      apps.example.com


*** Test Cases ***
Custom Default Certificate
    [Documentation]    Verify a custom TLS certificate secret can be configured as the default
    ...    ingress certificate, and that routes use the custom cert for TLS.
    ...    OCP-80508
    [Setup]    Prepare Custom Cert For Test    80508    route-edge80508.${BASE_DOMAIN}

    Setup Router Config And Restart    ${CONFIG_CUSTOM_CERT}

    Create OC Route And Admit    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${CERT_EDGE_HOST}
    Verify Custom Cert Is Active
    Deploy Test Client Pod
    Copy Files To Pod    ${NAMESPACE}    ${CLIENT_POD_NAME}    ${CERT_TMPDIR}    /data/certs
    ${router_ip}=    Get Router Pod IP
    VAR    ${resolve}=    ${CERT_EDGE_HOST}:443:${router_ip}
    Wait Until Curl With Client Cert Succeeds Insecure
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${CERT_EDGE_HOST}    ${resolve}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${CERT_EDGE_HOST}    ${resolve}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete secret router-test-cert -n ${ROUTER_NS} --ignore-not-found
    ...    AND    Cleanup Test Workloads

Old And Intermediate TLS Profiles
    [Documentation]    Verify the default Intermediate TLS profile cipher settings, then apply Old
    ...    profile and verify updated cipher settings, then restore Intermediate.
    ...    OCP-80510

    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.2
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERSUITES")].value
    Should Contain    ${ciphers}    TLS_AES_128_GCM_SHA256

    Setup Router Config And Restart    ${CONFIG_OLD_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.1
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERS")].value
    Should Contain    ${ciphers}    DES-CBC3-SHA

    Setup Router Config And Restart    ${CONFIG_INTERMEDIATE_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.2
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERSUITES")].value
    Should Contain    ${ciphers}    TLS_AES_128_GCM_SHA256
    [Teardown]    Remove Router Config And Restart

Modern And Custom TLS Profiles
    [Documentation]    Verify Modern TLS profile enforces TLSv1.3 in env vars and haproxy config,
    ...    then apply a Custom profile with specific ciphers.
    ...    OCP-80514

    Setup Router Config And Restart    ${CONFIG_MODERN_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.3
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERSUITES")].value
    Should Contain    ${ciphers}    TLS_AES_128_GCM_SHA256
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    ssl-default-bind-options ssl-min-ver TLSv1.3

    Setup Router Config And Restart    ${CONFIG_CUSTOM_TLS}
    Router Pod Env Should Have Value    SSL_MIN_VERSION    TLSv1.2
    ${ciphers}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}    .items[*].spec.containers[*].env[?(@.name=="ROUTER_CIPHERS")].value
    Should Contain    ${ciphers}    DHE-RSA-AES256-GCM-SHA384
    [Teardown]    Remove Router Config And Restart

MTLS Optional And Required Policy
    [Documentation]    Verify mTLS with clientCertificatePolicy Required rejects connections without
    ...    client cert, and Optional policy allows them.
    ...    OCP-80517
    [Setup]    Prepare MTLS Cert For Test    80517    route-edge80517.${BASE_DOMAIN}

    Run With Kubeconfig    oc create configmap ocp80517 --from-file=ca-bundle.pem=${MTLS_TMPDIR}/ca.crt -n ${ROUTER_NS}
    Setup Router Config And Restart    ${CONFIG_MTLS17_REQUIRED}
    Router Pod Env Should Have Value    ROUTER_MUTUAL_TLS_AUTH    required
    Deploy MTLS Test Workloads
    ${router_ip}=    Get Router Pod IP
    VAR    ${resolve}=    ${MTLS_EDGE_HOST}:443:${router_ip}
    Wait Until Curl With Client Cert Succeeds
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${MTLS_EDGE_HOST}    ${resolve}
    Curl Without Cert Should Return SSL Error
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${MTLS_EDGE_HOST}    ${resolve}

    Setup Router Config And Restart    ${CONFIG_MTLS17_OPTIONAL}
    Router Pod Env Should Have Value    ROUTER_MUTUAL_TLS_AUTH    optional
    ${router_ip}=    Get Router Pod IP
    VAR    ${resolve}=    ${MTLS_EDGE_HOST}:443:${router_ip}
    Wait Until Curl With Client Cert Succeeds
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${MTLS_EDGE_HOST}    ${resolve}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${MTLS_EDGE_HOST}    ${resolve}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete configmap ocp80517 -n ${ROUTER_NS} --ignore-not-found
    ...    AND    Cleanup Test Workloads


*** Keywords ***
Prepare Custom Cert For Test
    [Documentation]    Generate CA and user cert for custom certificate test, create TLS secret.
    [Arguments]    ${case_id}    ${edge_host}
    VAR    ${tmpdir}=    /tmp/ocp-${case_id}
    Create Directory    ${tmpdir}
    VAR    ${CERT_TMPDIR}=    ${tmpdir}    scope=TEST
    VAR    ${CERT_EDGE_HOST}=    ${edge_host}    scope=TEST
    Generate CA Certificate    ${tmpdir}/ca.key    ${tmpdir}/ca.crt    /CN=MS-default-CA
    ${san}=    San Extension    ${BASE_DOMAIN}
    Generate CSR And Key    ${tmpdir}/usr.key    ${tmpdir}/usr.csr    /CN=example-ne.com
    Sign CSR With CA    ${tmpdir}/usr.csr    ${tmpdir}/ca.crt    ${tmpdir}/ca.key    ${tmpdir}/usr.crt    ${san}
    Run With Kubeconfig
    ...    oc create secret tls router-test-cert --cert=${tmpdir}/ca.crt --key=${tmpdir}/ca.key -n ${ROUTER_NS}
    Deploy Web Server

Verify Custom Cert Is Active
    [Documentation]    Verify the custom cert secret is mounted and the cert issuer is correct.
    ${vol}=    Oc Get JsonPath
    ...    deployment
    ...    ${ROUTER_NS}
    ...    router-default
    ...    ..volumes[?(@.name=="default-certificate")].secret.secretName
    Should Contain    ${vol}    router-test-cert
    ${router_pod}=    Get Router Pod Name
    ${cert_info}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- openssl x509 -noout -in /etc/pki/tls/private/tls.crt -text
    Should Contain    ${cert_info}    CN=MS-default-CA

Prepare MTLS Cert For Test
    [Documentation]    Generate CA and client cert for mTLS tests.
    [Arguments]    ${case_id}    ${edge_host}
    VAR    ${tmpdir}=    /tmp/ocp-${case_id}-ca
    Create Directory    ${tmpdir}
    VAR    ${MTLS_TMPDIR}=    ${tmpdir}    scope=TEST
    VAR    ${MTLS_EDGE_HOST}=    ${edge_host}    scope=TEST
    Generate MTLS Client Cert    ${tmpdir}    ${edge_host}

Deploy MTLS Test Workloads
    [Documentation]    Deploy workloads and create the mTLS edge route.
    Deploy Web Server
    Deploy Test Client Pod
    Copy Files To Pod    ${NAMESPACE}    ${CLIENT_POD_NAME}    ${MTLS_TMPDIR}    /data/certs
    Create OC Route And Admit    ${NAMESPACE}    edge    route-edge    service-unsecure
    ...    --hostname=${MTLS_EDGE_HOST}    --cert=${MTLS_TMPDIR}/usr.crt    --key=${MTLS_TMPDIR}/usr.key
