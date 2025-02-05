*** Settings ***
Documentation       Tests for configuration changes

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}                           ${EMPTY}    # The journal cursor before restarting MicroShift
${BAD_LOG_LEVEL}                    SEPARATOR=\n
...                                 ---
...                                 debugging:
...                                 \ \ logLevel: unknown-value
${DEBUG_LOG_LEVEL}                  SEPARATOR=\n
...                                 ---
...                                 debugging:
...                                 \ \ logLevel: debug
${BAD_AUDIT_PROFILE}                SEPARATOR=\n
...                                 apiServer:
...                                 \ \ auditLog:
...                                 \ \ \ \ profile: BAD_PROFILE
${AUDIT_PROFILE}                    SEPARATOR=\n
...                                 apiServer:
...                                 \ \ auditLog:
...                                 \ \ \ \ profile: WriteRequestBodies
${AUDIT_FLAGS}                      SEPARATOR=\n
...                                 apiServer:
...                                 \ \ auditLog:
...                                 \ \ \ \ maxFileSize: 1000
...                                 \ \ \ \ maxFiles: 1000
...                                 \ \ \ \ maxFileAge: 1000
${LVMS_DEFAULT}                     SEPARATOR=\n
...                                 storage: {}
${LVMS_DISABLED}                    SEPARATOR=\n
...                                 storage:
...                                 \ \ driver: "none"
${CSI_SNAPSHOT_DISABLED}            SEPARATOR=\n
...                                 storage:
...                                 \ \ optionalCsiComponents: [ none ]
${LVMS_CSI_SNAPSHOT_DISABLED}       SEPARATOR=\n
...                                 storage:
...                                 \ \ driver: "none"
...                                 \ \ optionalCsiComponents: [ none ]
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
${APISERVER_ETCD_CLIENT_CERT}       /var/lib/microshift/certs/etcd-signer/apiserver-etcd-client


*** Test Cases ***
MicroShift Starts Using Default Config
    [Documentation]    Default (example) config should not fail to be parsed
    ...    and prevent MicroShift from starting.
    # Copy existing config.yaml as a drop-in because it has subjectAltNames
    # required by the `Restart MicroShift` keyword (sets up required kubeconfig).
    [Tags]    defaultcfg
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    AND
    ...    Command Should Work    mkdir -p /etc/microshift/config.d/
    ...    AND
    ...    Command Should Work    cp /etc/microshift/config.yaml /etc/microshift/config.d/00-ci.yaml
    ...    AND
    ...    Command Should Work    cp /etc/microshift/config.yaml.default /etc/microshift/config.yaml

    Restart MicroShift

    [Teardown]    Run Keywords
    ...    Restore Default MicroShift Config
    ...    AND
    ...    Command Should Work    rm -f /etc/microshift/config.d/00-ci.yaml
    ...    AND
    ...    Restart MicroShift

Unknown Log Level Produces Warning
    [Documentation]    Logs should warn that the log level setting is unknown
    Setup With Bad Log Level
    Pattern Should Appear In Log Output    ${CURSOR}    Unrecognized log level "unknown-value", defaulting to "Normal"

Debug Log Level Produces No Warning
    [Documentation]    Logs should not warn that the log level setting is unknown
    Setup With Debug Log Level
    Pattern Should Not Appear In Log Output    ${CURSOR}    Unrecognized log level "debug", defaulting to "Normal"

Known Audit Log Profile Produces No Warning
    [Documentation]    A recognized kube-apiserver audit log profile will not produce a message in logs
    Setup Known Audit Log Profile
    Pattern Should Not Appear In Log Output    ${CURSOR}    unknown audit profile \\\\"WriteRequestBodies\\\\"

Config Flags Are Logged in Audit Flags
    [Documentation]    Check that flags specified in the MicroShift audit config are passed to and logged by the
    ...    kube-apiserver. It is not essential that we test the kube-apiserver functionality as that is
    ...    already rigorously tested by upstream k8s and by OCP.
    Setup Audit Flags
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxsize=\"1000\"
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxbackup=\"1000\"
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxage=\"1000\"

Deploy MicroShift With LVMS By Default
    [Documentation]    Verify that LVMS and CSI snapshotting are deployed when config fields are null.
    [Setup]    Deploy Storage Config    ${LVMS_DEFAULT}
    LVMS Is Deployed
    CSI Snapshot Controller Is Deployed
    [Teardown]    Run Keywords
    ...    Remove Storage Drop In Config
    ...    Restart MicroShift

Deploy MicroShift Without LVMS
    [Documentation]    Verify that LVMS is not deployed when storage.driver == none, and that CSI snapshotting
    ...    components are still deployed.
    [Setup]    Deploy Storage Config    ${LVMS_DISABLED}

    CSI Snapshot Controller Is Deployed
    Run Keyword And Expect Error    1 != 0
    ...    LVMS Is Deployed
    [Teardown]    Run Keywords
    ...    Remove Storage Drop In Config
    ...    Restart MicroShift

Deploy MicroShift Without CSI Snapshotter
    [Documentation]    Verify that only LVMS is deployed when .storage.optionalCsiComponents is an empty array.
    [Setup]    Deploy Storage Config    ${CSI_SNAPSHOT_DISABLED}

    LVMS Is Deployed
    Run Keyword And Expect Error    1 != 0
    ...    CSI Snapshot Controller Is Deployed

    [Teardown]    Run Keywords
    ...    Remove Storage Drop In Config
    ...    Restart MicroShift

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


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Save Journal Cursor

Teardown
    [Documentation]    Test suite teardown
    Remove Drop In MicroShift Config    10-loglevel
    Remove Drop In MicroShift Config    10-audit
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Save Journal Cursor
    [Documentation]
    ...    Save the journal cursor then restart MicroShift so we capture the
    ...    shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}

Setup With Bad Log Level
    [Documentation]    Set log level to an unknown value and restart
    Drop In MicroShift Config    ${BAD_LOG_LEVEL}    10-loglevel
    Restart MicroShift

Setup With Debug Log Level
    [Documentation]    Set log level to debug and restart
    Drop In MicroShift Config    ${DEBUG_LOG_LEVEL}    10-loglevel
    Restart MicroShift

Setup Known Audit Log Profile
    [Documentation]    Setup audit
    Drop In MicroShift Config    ${AUDIT_PROFILE}    10-audit
    Restart MicroShift

Setup Audit Flags
    [Documentation]    Apply the audit config values set in ${AUDIT_FLAGS}
    Drop In MicroShift Config    ${AUDIT_FLAGS}    10-audit
    Restart MicroShift

Deploy Storage Config
    [Documentation]    Applies a storage ${config} to the exist MicroShift config, pushes it to the MicroShift host,
    ...    and restarts microshift.service
    [Arguments]    ${config}
    Cleanup MicroShift    opt='--keep-images'
    Drop In MicroShift Config    ${config}    10-storage
    Start MicroShift

Setup TLS Configuration
    [Documentation]    Apply the TLS configuration in the argument
    [Arguments]    ${config}
    Drop In MicroShift Config    ${config}    10-tls
    Restart MicroShift

Remove TLS Drop In Config
    [Documentation]    Remove the previously created drop-in config for storage
    Remove Drop In MicroShift Config    10-tls

Remove Storage Drop In Config
    [Documentation]    Remove the previously created drop-in config for storage
    Remove Drop In MicroShift Config    10-storage

LVMS Is Deployed
    [Documentation]    Wait for LVMS components to deploy
    Named Deployment Should Be Available    lvms-operator    openshift-storage    120s
    # Wait for vg-manager daemonset to exist before trying to "wait".
    # `oc wait` fails if the object doesn't exist.
    Wait Until Resource Exists    daemonset    vg-manager    openshift-storage    120s
    Named Daemonset Should Be Available    vg-manager    openshift-storage    120s

CSI Snapshot Controller Is Deployed
    [Documentation]    Wait for CSI snapshot controller to be deployed
    Named Deployment Should Be Available    csi-snapshot-controller    kube-system    120s

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
