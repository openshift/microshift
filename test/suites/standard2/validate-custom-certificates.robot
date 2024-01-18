*** Settings ***
Documentation       Tests verifying hostname resolution

Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/openssl.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${NEW_HOSTNAME}     microshift.local
${OLD_HOSTNAME}     ${EMPTY}
${CHECK_PAYLOAD_IMAGE}      registry.ci.openshift.org/ci/check-payload:latest
${PULL_SECRET_PATH}         /etc/crio/openshift-pull-secret
${CSR_CONFIG}               ./assets/custom-certs/csr.conf.template
${CSR_NOSAN_CONFIG}         ./assets/custom-certs/csr-no-san.conf.template
${MASTER_IP}                1.1.1.1
${TMPDIR}                   ${EMPTY}
${KUBECONFIG_CUSTOM_CERTS}  custom-certs


*** Test Cases ***
Test Missing File
    [Documentation]    Missing certificate files should be ignored with a warning
    [Setup]    Setup Test
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}    
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Pattern Should Appear In Log Output    ${CURSOR}    unparseable certificates are ignored
    Setup Kubeconfig
    MicroShift Is Live

Test Expired Cert
    [Documentation]    Create regular SNI certificate
    [Setup]    Setup Test
    # Generate CSR Config
    Create Keys
    Create Cert    TestCN    test-expired.api.com    5.5.5.5    0
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Add Entry To Hosts    ${USHIFT_HOST}    test-expired.api.com
    Setup Custom Kubeconfig    ${KUBECONFIG_CUSTOM_CERTS}/test-expired.api.com
    OC Should Fail To Connect With Expired Cert
    [Teardown]    RemoveRootCAFromRhel

Test Local Cert
    [Documentation]    localhost certs should be ignored
    [Setup]    Setup Test    
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Create Keys
    Create Cert    TestCN    localhost
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Pattern Should Appear In Log Output    ${CURSOR}    not allowed and ignored
    Setup Kubeconfig
    MicroShift Is Live
    [Teardown]    RemoveRootCAFromRhel

Test SAN Cert
    [Documentation]    Create regular SNI certificate
    [Setup]    Setup Test
    # Generate CSR Config
    Create Keys
    Create Cert    TestCN    test.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key    test
    Restart MicroShift
    Add Entry To Hosts    ${USHIFT_HOST}    test.api.com
    Setup Custom Kubeconfig    ${KUBECONFIG_CUSTOM_CERTS}/test.api.com
    OC Should Fail To Connect With Unknown CA
    AddRootCAToRHEL
    MicroShift Is Live
    [Teardown]    RemoveRootCAFromRhel

Test Wildcard Only Cert
    [Documentation]    Create WildCard only certificate
    [Setup]    Setup Test
    Create Keys
    Create Cert    TestCN    *.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Add Entry To Hosts    ${USHIFT_HOST}    test.api.com
    Setup Custom Kubeconfig    ${KUBECONFIG_CUSTOM_CERTS}/TestCN
    Replace Server In Kubeconfig    test.api.com
    OC Should Fail To Connect With Unknown CA
    AddRootCAToRHEL
    MicroShift Is Live
    [Teardown]    RemoveRootCAFromRhel

Test Wildcard With Names Cert
    [Documentation]    Create WildCard only certificate
    [Setup]    Setup Test
    Create Keys
    Create Cert    TestCN    *.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key    test.api.com
    Restart MicroShift
    Add Entry To Hosts    ${USHIFT_HOST}    test.api.com
    Setup Custom Kubeconfig    ${KUBECONFIG_CUSTOM_CERTS}/test.api.com
    OC Should Fail To Connect With Unknown CA
    AddRootCAToRHEL
    MicroShift Is Live
    [Teardown]    RemoveRootCAFromRhel

Test Wildcard Cert CN Only
    [Documentation]    Documentation
    [Setup]    Setup Test
    Create Keys
    Create Cert No San    *.testing.com
    Upload Certificates


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Setup Test
    [Documentation]    sdas
    ${tmp}=    Create Random Temp Directory
    Set Global Variable    ${TMPDIR}    ${tmp}
    Save Default MicroShift Config

Teardown
    [Documentation]    Test suite teardown
    Cleanup Hosts
    Remove Kubeconfig
    Logout MicroShift Host

Create Keys
    [Documentation]    Create a certificate CA
    Log To Console    ${TMPDIR}
    Openssl    genrsa -out ${TMPDIR}/ca.key 2048
    Openssl    genrsa -out ${TMPDIR}/server.key 2048
    Openssl    req -x509 -new -nodes -key ${TMPDIR}/ca.key -subj "/CN\=${MASTER_IP}"
    ...    -days 10000 -out ${TMPDIR}/ca.crt

Create Cert No San
    [Documentation]    Create a certificate
    [Arguments]    ${CERT_CN}
    Set Global Variable    ${CERT_CN}
    Generate CSR Config    ${CSR_NOSAN_CONFIG}    ${TMPDIR}/csr.conf
    Openssl    req -new -key ${TMPDIR}/server.key -out ${TMPDIR}/server.csr -config ${TMPDIR}/csr.conf
    Openssl    x509 -req -in ${TMPDIR}/server.csr -CA ${TMPDIR}/ca.crt -CAkey ${TMPDIR}/ca.key -CAcreateserial
    ...    -out ${TMPDIR}/server.crt -days 10000 -extensions v3_ext -extfile ${TMPDIR}/csr.conf -sha256

Create Cert
    [Documentation]    Create a certificate
    [Arguments]    ${CERT_CN}    ${CERT_SAN_DNS}=${EMPTY}    ${CERT_SAN_IP}=${EMPTY}    ${EXPIRY_DAYS}=1000
    Set Global Variable    ${CERT_CN}
    IF    "${CERT_SAN_DNS}"!="${EMPTY}"
        Set Global Variable    ${CERT_SAN_DNS}    DNS.1 = ${CERT_SAN_DNS}
    ELSE
        Set Global Variable    ${CERT_SAN_DNS}
    END

    IF    "${CERT_SAN_IP}"!="${EMPTY}"
        Set Global Variable    ${CERT_SAN_IP}    IP.1 = ${CERT_SAN_IP}
    ELSE
        Set Global Variable    ${CERT_SAN_IP}
    END
    Generate CSR Config    ${CSR_CONFIG}    ${TMPDIR}/csr.conf
    Openssl    req -new -key ${TMPDIR}/server.key -out ${TMPDIR}/server.csr -config ${TMPDIR}/csr.conf
    Openssl    x509 -req -in ${TMPDIR}/server.csr -CA ${TMPDIR}/ca.crt -CAkey ${TMPDIR}/ca.key -CAcreateserial
    ...    -out ${TMPDIR}/server.crt -days ${EXPIRY_DAYS} -extensions v3_ext -extfile ${TMPDIR}/csr.conf -sha256

Upload Certificates
    [Documentation]    Upload certificates to remote host
    Put Directory    ${TMPDIR}    ${TMPDIR}

Configure Named Certificates
    [Documentation]    Replace namedCertificates entries in the configuration file.
    [Arguments]    ${certPath}    ${keyPath}    ${sni}=${EMPTY}

    ${subject_alt_names}=    CATENATE    SEPARATOR=\n
    ...    apiServer:
    ...    \ \ subjectAltNames:
    ...    \ \ -\ ${USHIFT_HOST}
    ...    \ \ namedCertificates:
    ...    \ \ - certPath:\ ${certPath}
    ...    \ \ \ \ keyPath:\ ${keyPath}

    IF    "${sni}"!="${EMPTY}"
        ${subject_alt_names}=    CATENATE    SEPARATOR=\n    ${subject_alt_names}
        ...    \ \ \ \ names:
        ...    \ \ \ \ - ${sni}
    END
    Upload MicroShift Config    ${subject_alt_names}

Add Entry To Hosts
    [Documentation]    Add new entry to local /etc/hosts
    [Arguments]    ${ip}    ${host}
    ${ttt}=    Set Variable    ${ip}\t${host} # RF test marker
    ${result}=    Run Process    echo -e "${ttt}" | sudo tee -a /etc/hosts    shell=True
    Should Be Equal As Integers    ${result.rc}    0

Cleanup Hosts
    [Documentation]    remove entries from  local /etc/hosts
    ${result}=    Run Process    sudo sed -i "/# RF test marker/d" /etc/hosts    shell=True
    Should Be Equal As Integers    ${result.rc}    0

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

AddRootCAToRHEL
    [Documentation]    add certificate to the rhel trust store
    [Arguments]    ${src_crt}=ca.crt    ${dst_crt}=oc_ca.pem
    ${rhl_ca_path}=    Set Variable    /etc/pki/ca-trust/source/anchors/${dst_crt}
    ${result}=    Run Process    sudo cp ${TMPDIR}/${src_crt} ${rhl_ca_path} && sudo update-ca-trust    shell=True
    Should Be Equal As Integers    ${result.rc}    0

RemoveRootCAFromRhel
    [Documentation]    remove certificate from the rhel trust store
    [Arguments]    ${dst_crt}=oc_ca.pem
    ${rhl_ca_path}=    Set Variable    /etc/pki/ca-trust/source/anchors/${dst_crt}
    ${result}=    Run Process    sudo rm -rf ${rhl_ca_path} && sudo update-ca-trust    shell=True
    Should Be Equal As Integers    ${result.rc}    0
