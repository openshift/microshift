*** Settings ***
Documentation       Tests for configuration changes

Resource            ../../resources/fault-tests.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}                           ${EMPTY}    # The journal cursor before restarting MicroShift
${TLS_12_CUSTOM_CIPHER}             SEPARATOR=\n
...                                 apiServer:
...                                 \ \ tls:
...                                 \ \ \ \ cipherSuites:
...                                 \ \ \ \ - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
...                                 \ \ \ \ minVersion: VersionTLS12
${TLS_13_MIN_VERSION}               SEPARATOR=\n
...                                 apiServer:
...                                 \ \ tls:
...                                 \ \ \ \ minVersion: VersionTLS13
${TLS_INVALID_CIPHER}               SEPARATOR=\n
...                                 apiServer:
...                                 \ \ tls:
...                                 \ \ \ \ cipherSuites:
...                                 \ \ \ \ - TLS_INVALID_CIPHER
...                                 \ \ \ \ minVersion: VersionTLS12
${TLS_INVALID_VERSION}              SEPARATOR=\n
...                                 apiServer:
...                                 \ \ tls:
...                                 \ \ \ \ cipherSuites:
...                                 \ \ \ \ - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
...                                 \ \ \ \ minVersion: VersionTLSInvalid
${APISERVER_ETCD_CLIENT_CERT}       /var/lib/microshift/certs/etcd-signer/apiserver-etcd-client


*** Test Cases ***
Custom TLS 1_2 configuration
    [Documentation]    Configure a custom cipher suite using TLSv1.2 as min version and verify it is used
    [Setup]    Setup TLS Configuration    ${TLS_12_CUSTOM_CIPHER}

    ${config}=    Show Config    effective
    Should Be Equal    ${config.apiServer.tls.minVersion}    VersionTLS12
    Length Should Be    ${config.apiServer.tls.cipherSuites}    2
    Should Contain    ${config.apiServer.tls.cipherSuites}    TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
    Should Contain    ${config.apiServer.tls.cipherSuites}    TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256

    # on TLSv1.2, openssl ciphers string codes (defined by IANA) does not excatly match openshift ones
    # custom cipher defined for this test
    Check TLS Endpoints    0    TLSv1.2    ECDHE-RSA-CHACHA20-POLY1305    # TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
    Check TLS Endpoints    1    TLSv1.3    ECDHE-RSA-CHACHA20-POLY1305    # TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256

    # mandatory cipher needed for internal enpoints (i.e. etcd), set if not defined by the user
    Check TLS Endpoints    0    TLSv1.2    ECDHE-RSA-AES128-GCM-SHA256    # TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    Check TLS Endpoints    1    TLSv1.3    ECDHE-RSA-AES128-GCM-SHA256    # TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256

    # when TLSv1.2 is set as min version, TLSv1.3 must also work
    Check TLS Endpoints    1    TLSv1.2    TLS_AES_128_GCM_SHA256
    Check TLS Endpoints    0    TLSv1.3    TLS_AES_128_GCM_SHA256

    [Teardown]    Run Keywords
    ...    Remove TLS Drop In Config
    ...    Restart MicroShift

Custom TLS 1_3 configuration
    [Documentation]    Configure API server to use TLSv1.3 as min version and verify only that version works
    ...    TLSv1.2 must fail and cipher suites for TLSv1.3 can not be config by the user, always 3 are enabled.
    [Setup]    Setup TLS Configuration    ${TLS_13_MIN_VERSION}

    ${config}=    Show Config    effective
    Should Be Equal    ${config.apiServer.tls.minVersion}    VersionTLS13
    Length Should Be    ${config.apiServer.tls.cipherSuites}    3
    Should Contain    ${config.apiServer.tls.cipherSuites}    TLS_AES_128_GCM_SHA256
    Should Contain    ${config.apiServer.tls.cipherSuites}    TLS_AES_256_GCM_SHA384
    Should Contain    ${config.apiServer.tls.cipherSuites}    TLS_CHACHA20_POLY1305_SHA256

    # checking the 3 ciphers available for TLSv1.3 on openshift
    Check TLS Endpoints    1    TLSv1.2    TLS_AES_128_GCM_SHA256
    Check TLS Endpoints    0    TLSv1.3    TLS_AES_128_GCM_SHA256
    Check TLS Endpoints    1    TLSv1.2    TLS_AES_256_GCM_SHA384
    Check TLS Endpoints    0    TLSv1.3    TLS_AES_256_GCM_SHA384
    Check TLS Endpoints    1    TLSv1.2    TLS_CHACHA20_POLY1305_SHA256
    Check TLS Endpoints    0    TLSv1.3    TLS_CHACHA20_POLY1305_SHA256

    [Teardown]    Run Keywords
    ...    Remove TLS Drop In Config
    ...    Restart MicroShift

TLS Config: Unsupported Cipher Suite
    [Documentation]    Test behavior when TLS version is set to an unsupported cipher suite
    [Setup]    Setup Invalid TLS Configuration    ${TLS_INVALID_CIPHER}
    Check Journal Logs    config    tls_invalid_cipher
    [Teardown]    Restore Valid TLS Configuration

TLS Config: Invalid TLS Version
    [Documentation]    Test behavior when TLS version is set to an invalid value
    [Setup]    Setup Invalid TLS Configuration    ${TLS_INVALID_VERSION}
    Check Journal Logs    config    tls_invalid_version
    [Teardown]    Restore Valid TLS Configuration


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks

Teardown
    [Documentation]    Test suite teardown
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Save Journal Cursor
    [Documentation]
    ...    Save the journal cursor then restart MicroShift so we capture the
    ...    shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    Set Test Variable    \${CURSOR}    ${cursor}

Setup TLS Configuration
    [Documentation]    Apply the TLS configuration in the argument
    [Arguments]    ${config}
    Drop In MicroShift Config    ${config}    10-tls
    Restart MicroShift

Setup Invalid TLS Configuration
    [Documentation]    Apply the TLS configuration in the argument
    [Arguments]    ${config}
    Drop In MicroShift Config    ${config}    10-tls
    Save Journal Cursor
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl restart microshift.service
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Numbers    1    ${rc}

Remove TLS Drop In Config
    [Documentation]    Remove the previously created drop-in config for storage
    Remove Drop In MicroShift Config    10-tls

Restore Valid TLS Configuration
    [Documentation]    Restore the TLS configuration
    Remove TLS Drop In Config
    Sleep    10s    # To avoid systemctl start error: Start request repeated too quickly
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    systemctl restart microshift.service
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Numbers    0    ${rc}

Openssl Connect Command
    [Documentation]    Run Openssl Connect Command in the remote server
    [Arguments]    ${host_and_port}    ${args}
    ${stdout}    ${rc}=    Execute Command
    ...    openssl s_client -connect ${host_and_port} ${args} <<< "Q"
    ...    sudo=True    return_stdout=True    return_stderr=False    return_rc=True
    RETURN    ${stdout}    ${rc}

Check TLS Endpoints
    [Documentation]    Run Openssl Connect Command to check k8s internal endpoints
    [Arguments]    ${return_code}    ${tls_version}    ${cipher}
    IF    "${tls_version}" == "TLSv1.2"
        Set Test Variable    ${TLS_AND_CIPHER_ARGS}    -tls1_2 -cipher ${cipher}
    ELSE IF    "${tls_version}" == "TLSv1.3"
        Set Test Variable    ${TLS_AND_CIPHER_ARGS}    -tls1_3 -ciphersuites ${cipher}
    END

    # api server, kubelet, kube controller manager and kube scheduler endpoint ports
    FOR    ${port}    IN    6443    10250    10257    10259
        ${stdout}    ${rc}=    Openssl Connect Command    ${USHIFT_HOST}:${port}    ${TLS_AND_CIPHER_ARGS}
        Should Be Equal As Integers    ${return_code}    ${rc}
        IF    "${rc}" == "0"
            Should Contain    ${stdout}    ${tls_version}, Cipher is ${cipher}
        END
    END

    # etcd endpoint, need to use cert and key because etcd requires mTLS
    Set Test Variable    ${CERT_ARG}    -cert ${APISERVER_ETCD_CLIENT_CERT}/client.crt
    Set Test Variable    ${KEY_ARG}    -key ${APISERVER_ETCD_CLIENT_CERT}/client.key
    ${stdout}    ${rc}=    Openssl Connect Command    localhost:2379    ${TLS_AND_CIPHER_ARGS} ${CERT_ARG} ${KEY_ARG}
    Should Be Equal As Integers    ${return_code}    ${rc}

Check Journal Logs
    [Documentation]    Verify system logs contain expected error messages for configuration errors
    [Arguments]    ${error_type}    ${error_case}
    ${output}    ${rc}=    Get Log Output With Pattern    ${CURSOR}    apiServer.tls
    @{expected_lines}=    Get Expected Fault Messages    ${error_type}    ${error_case}
    Compare Output Logs    ${output}    @{expected_lines}
