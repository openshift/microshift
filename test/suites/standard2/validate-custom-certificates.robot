*** Settings ***
Documentation       Tests custom certificates functionality

Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/openssl.resource
Resource            ../../resources/hosts.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CSR_CONFIG}           ./assets/custom-certs/csr.conf.template
${CSR_NOSAN_CONFIG}     ./assets/custom-certs/csr-no-san.conf.template
${MASTER_IP}            1.1.1.1
${TMPDIR}               ${EMPTY}
${RHL_CA_PATH}          ${EMPTY}


*** Test Cases ***
Test Missing File
    [Documentation]    Missing certificate files should be ignored with a warning
    [Setup]    Setup Test
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Pattern Should Appear In Log Output    ${CURSOR}    unparsable certificates are ignored
    Setup Kubeconfig
    MicroShift Is Live

Test Expired Cert
    [Documentation]    Expired certificate files should be accepted but fail secure connection
    [Setup]    Setup Test
    # Generate CSR Config
    Create Keys
    ${hostname}=    Generate Random HostName
    Create Cert    TestCN    ${hostname}    5.5.5.5    0
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Add Entry To Local Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    ${hostname}
    OC Should Fail To Connect With Expired Cert
    [Teardown]    Remove Entry From Local Hosts    ${hostname}

Test Local Cert
    [Documentation]    localhost certs should be ignored with a warning
    [Setup]    Setup Test
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE
    Create Keys
    Create Cert    TestCN    localhost
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Pattern Should Appear In Log Output    ${CURSOR}    is not allowed - ignoring
    Setup Kubeconfig
    MicroShift Is Live

Test SAN Cert
    [Documentation]    Create regular SNI certificate
    [Setup]    Setup Test
    Create Keys
    ${hostname}=    Generate Random HostName
    Create Cert    TestCN    ${hostname}
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key    test
    Restart MicroShift
    Add Entry To Local Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    ${hostname}
    OC Should Fail To Connect With Unknown CA
    MicroShift Is Live With Custom CA    ${TMPDIR}/ca.crt
    [Teardown]    Remove Entry From Local Hosts    ${hostname}

Test Wildcard Only Cert
    [Documentation]    Create WildCard only certificate
    [Setup]    Setup Test
    Create Keys
    ${hostname}=    Generate Random HostName
    Create Cert    TestCN    *.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Add Entry To Local Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    TestCN
    Replace Server In Kubeconfig    ${hostname}
    OC Should Fail To Connect With Unknown CA
    MicroShift Is Live With Custom CA    ${TMPDIR}/ca.crt
    [Teardown]    Remove Entry From Local Hosts    ${hostname}

Test Wildcard With Names Cert
    [Documentation]    Create WildCard certificate with additional config name
    [Setup]    Setup Test
    Create Keys
    ${hostname}=    Generate Random HostName
    Create Cert    TestCN    *.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key    ${hostname}
    Restart MicroShift
    Add Entry To Local Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    ${hostname}
    OC Should Fail To Connect With Unknown CA
    MicroShift Is Live With Custom CA    ${TMPDIR}/ca.crt
    [Teardown]    Remove Entry From Local Hosts    ${hostname}


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Setup Test
    [Documentation]    Test suite setup
    ${tmp}=    Create Random Temp Directory
    VAR    ${TMPDIR}=    ${tmp}    scope=GLOBAL

Teardown
    [Documentation]    Test suite teardown
    Remove Drop In MicroShift Config    10-subjectAltNames
    Remove Kubeconfig
    Logout MicroShift Host

Create Keys
    [Documentation]    Create a certificate CA
    Openssl    genrsa -out ${TMPDIR}/ca.key 2048
    Openssl    genrsa -out ${TMPDIR}/server.key 2048
    Openssl    req -x509 -new -nodes -key ${TMPDIR}/ca.key -subj "/CN\=${MASTER_IP}"
    ...    -days 10000 -out ${TMPDIR}/ca.crt

Create Cert
    [Documentation]    Create a certificate
    [Arguments]    ${cert_cn}    ${cert_san_dns}=${EMPTY}    ${cert_san_ip}=${EMPTY}    ${expiry_days}=1000
    VAR    ${CERT_CN}=    ${cert_cn}    scope=GLOBAL
    IF    "${cert_san_dns}"!="${EMPTY}"
        VAR    ${CERT_SAN_DNS}=    DNS.1 = ${cert_san_dns}    scope=GLOBAL
    ELSE
        VAR    ${CERT_SAN_DNS}=    ${EMPTY}    scope=GLOBAL
    END

    IF    "${cert_san_ip}"!="${EMPTY}"
        VAR    ${CERT_SAN_IP}=    IP.1 = ${cert_san_ip}    scope=GLOBAL
    ELSE
        VAR    ${CERT_SAN_IP}=    ${EMPTY}    scope=GLOBAL
    END
    Generate CSR Config    ${CSR_CONFIG}    ${TMPDIR}/csr.conf
    Openssl    req -new -key ${TMPDIR}/server.key -out ${TMPDIR}/server.csr -config ${TMPDIR}/csr.conf
    Openssl    x509 -req -in ${TMPDIR}/server.csr -CA ${TMPDIR}/ca.crt -CAkey ${TMPDIR}/ca.key -CAcreateserial
    ...    -out ${TMPDIR}/server.crt -days ${expiry_days} -extensions v3_ext -extfile ${TMPDIR}/csr.conf -sha256

Upload Certificates
    [Documentation]    Upload certificates to remote host
    Put Directory    ${TMPDIR}    ${TMPDIR}

Configure Named Certificates
    [Documentation]    Replace namedCertificates entries in the configuration file.
    [Arguments]    ${cert_path}    ${key_path}    ${sni}=${EMPTY}

    ${subject_alt_names}=    CATENATE    SEPARATOR=\n
    ...    apiServer:
    ...    \ \ subjectAltNames:
    ...    \ \ -\ ${USHIFT_HOST}
    ...    \ \ namedCertificates:
    ...    \ \ - certPath:\ ${cert_path}
    ...    \ \ \ \ keyPath:\ ${key_path}

    IF    "${sni}"!="${EMPTY}"
        ${subject_alt_names}=    CATENATE    SEPARATOR=\n    ${subject_alt_names}
        ...    \ \ \ \ names:
        ...    \ \ \ \ - ${sni}
    END
    Drop In MicroShift Config    ${subject_alt_names}    10-subjectAltNames

Replace Server In Kubeconfig
    [Documentation]    replace the server part of kubeconfig
    [Arguments]    ${fqdn}
    ${result}=    Run Process    sudo sed -i "s|server:.*|server: https:\/\/${fqdn}:6443|" ${KUBECONFIG}    shell=True
    Should Be Equal As Integers    ${result.rc}    0

OC Should Fail To Connect With Unknown CA
    [Documentation]    Check the /livez endpoint
    ${stdout}=    Run With Kubeconfig    oc get --raw='/livez'    allow_fail=True
    Should Contain    ${stdout}    certificate signed by unknown authority    strip_spaces=True

OC Should Fail To Connect With Expired Cert
    [Documentation]    Check the /livez endpoint
    ${stdout}=    Run With Kubeconfig    oc get --raw='/livez'    allow_fail=True
    Should Contain    ${stdout}    certificate has expired or is not yet valid    strip_spaces=True

MicroShift Is Live With Custom CA
    [Documentation]    Check the /livez endpoint with Custom CA
    [Arguments]    ${ca_path}
    MicroShift Is Live    --certificate-authority ${ca_path}
