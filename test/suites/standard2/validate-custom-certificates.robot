*** Settings ***
Documentation       Tests custom certificates functionality

Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/openssl.resource
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
    Set Suite Variable    \${CURSOR}    ${cursor}
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
    Add Entry To Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    ${hostname}
    OC Should Fail To Connect With Expired Cert
    [Teardown]    Remove Entry From Hosts    ${hostname}

Test Local Cert
    [Documentation]    localhost certs should be ignored with a warning
    [Setup]    Setup Test
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
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
    Add Entry To Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    ${hostname}
    OC Should Fail To Connect With Unknown CA
    MicroShift Is Live With Custom CA    ${TMPDIR}/ca.crt
    [Teardown]    Remove Entry From Hosts    ${hostname}

Test Wildcard Only Cert
    [Documentation]    Create WildCard only certificate
    [Setup]    Setup Test
    Create Keys
    ${hostname}=    Generate Random HostName
    Create Cert    TestCN    *.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key
    Restart MicroShift
    Add Entry To Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    TestCN
    Replace Server In Kubeconfig    ${hostname}
    OC Should Fail To Connect With Unknown CA
    MicroShift Is Live With Custom CA    ${TMPDIR}/ca.crt
    [Teardown]    Remove Entry From Hosts    ${hostname}

Test Wildcard With Names Cert
    [Documentation]    Create WildCard certificate with additional config name
    [Setup]    Setup Test
    Create Keys
    ${hostname}=    Generate Random HostName
    Create Cert    TestCN    *.api.com
    Upload Certificates
    Configure Named Certificates    ${TMPDIR}/server.crt    ${TMPDIR}/server.key    ${hostname}
    Restart MicroShift
    Add Entry To Hosts    ${USHIFT_HOST}    ${hostname}
    Setup Custom Kubeconfig    ${hostname}
    OC Should Fail To Connect With Unknown CA
    MicroShift Is Live With Custom CA    ${TMPDIR}/ca.crt
    [Teardown]    Remove Entry From Hosts    ${hostname}


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Save Default MicroShift Config

Setup Test
    [Documentation]    Test suite setup
    ${tmp}=    Create Random Temp Directory
    Set Global Variable    ${TMPDIR}    ${tmp}

Teardown
    [Documentation]    Test suite teardown
    Remove Kubeconfig
    Restore Default MicroShift Config
    Restart MicroShift
    Logout MicroShift Host

Create Keys
    [Documentation]    Create a certificate CA
    Openssl    genrsa -out ${TMPDIR}/ca.key 2048
    Openssl    genrsa -out ${TMPDIR}/server.key 2048
    Openssl    req -x509 -new -nodes -key ${TMPDIR}/ca.key -subj "/CN\=${MASTER_IP}"
    ...    -days 10000 -out ${TMPDIR}/ca.crt

Create Cert No San
    [Documentation]    Create a certificate
    [Arguments]    ${cert_cn}
    Set Global Variable    ${CERT_CN}
    Generate CSR Config    ${CSR_NOSAN_CONFIG}    ${TMPDIR}/csr.conf
    Openssl    req -new -key ${TMPDIR}/server.key -out ${TMPDIR}/server.csr -config ${TMPDIR}/csr.conf
    Openssl    x509 -req -in ${TMPDIR}/server.csr -CA ${TMPDIR}/ca.crt -CAkey ${TMPDIR}/ca.key -CAcreateserial
    ...    -out ${TMPDIR}/server.crt -days 10000 -extensions v3_ext -extfile ${TMPDIR}/csr.conf -sha256

Create Cert
    [Documentation]    Create a certificate
    [Arguments]    ${cert_cn}    ${cert_san_dns}=${EMPTY}    ${cert_san_ip}=${EMPTY}    ${expiry_days}=1000
    Set Global Variable    ${CERT_CN}
    IF    "${cert_san_dns}"!="${EMPTY}"
        Set Global Variable    ${CERT_SAN_DNS}    DNS.1 = ${cert_san_dns}
    ELSE
        Set Global Variable    ${CERT_SAN_DNS}
    END

    IF    "${cert_san_ip}"!="${EMPTY}"
        Set Global Variable    ${CERT_SAN_IP}    IP.1 = ${cert_san_ip}
    ELSE
        Set Global Variable    ${CERT_SAN_IP}
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
    ${replaced}=    Replace MicroShift Config    ${subject_alt_names}
    Upload MicroShift Config    ${replaced}

Generate Random HostName
    [Documentation]    Generate Random Hostname
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    RETURN    ${rand}.api.com

Add Entry To Hosts
    [Documentation]    Add new entry to local /etc/hosts
    [Arguments]    ${ip}    ${host}
    ${ttt}=    Set Variable    ${ip}\t${host} # RF test marker\n
    ${result}=    Run Process    sudo tee -a /etc/hosts    shell=True    stdin=${ttt}
    Should Be Equal As Integers    ${result.rc}    0

Remove Entry From Hosts
    [Documentation]    Removes entry from local /etc/hosts
    [Arguments]    ${host}
    ${result}=    Run Process    sudo sed -i "/${host} # RF test marker/d" /etc/hosts    shell=True
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

MicroShift Is Live With Custom CA
    [Documentation]    Check the /livez endpoint with Custom CA
    [Arguments]    ${ca_path}
    MicroShift Is Live    --certificate-authority ${ca_path}
