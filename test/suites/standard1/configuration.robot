*** Settings ***
Documentation       Tests for configuration changes

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite

Test Tags           restart    slow

#FLAG: --audit-log-maxage="0"
#FLAG: --audit-log-maxbackup="0"
#FLAG: --audit-log-maxsize="0"

*** Variables ***
${CURSOR}               ${EMPTY}    # The journal cursor before restarting MicroShift
${BAD_LOG_LEVEL}        SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: unknown-value
${DEBUG_LOG_LEVEL}      SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logLevel: debug
${BAD_AUDIT_PROFILE}    SEPARATOR=\n
...                     apiServer:
...                     \ \ auditLog:
...                     \ \ \ \ profile: BAD_PROFILE
${AUDIT_PROFILE}        SEPARATOR=\n
...                     apiServer:
...                     \ \ auditLog:
...                     \ \ \ \ profile: WriteRequestBodies
${AUDIT_FLAGS}        SEPARATOR=\n
...                     apiServer:
...                     \ \ auditLog:
...                     \ \ \ \ maxFileSize: 1000
...                     \ \ \ \ maxFiles: 1000
...                     \ \ \ \ maxFileAges: 1000


*** Test Cases ***
Unknown Log Level Produces Warning
    [Documentation]    Logs should warn that the log level setting is unknown
    [Teardown]    Teardown
    Setup With Bad Log Level
    Pattern Should Appear In Log Output    ${CURSOR}    Unrecognized log level "unknown-value", defaulting to "Normal"

Debug Log Level Produces No Warning
    [Documentation]    Logs should not warn that the log level setting is unknown
    [Teardown]    Teardown
    Setup With Debug Log Level
    Pattern Should Not Appear In Log Output    ${CURSOR}    Unrecognized log level "debug", defaulting to "Normal"

Unknown Audit Log Profile Should Fail
    [Documentation]    Unrecognized kube-apiserver audit log profile should prevent kube-apiserver from starting
    [Teardown]    Teardown
    Setup With Unknown Audit Log Profile
    Pattern Should Appear In Log Output    ${CURSOR}
    ...    "SERVICE FAILED - stopping MicroShift" err="configuration failed: failed to configure kube-apiserver audit policy: unknown audit profile \"BAD_PROFILE\"" service="kube-apiserver"

Known Audit Log Profile Produces No Warning
    [Documentation]    A recognized kube-apiserver audit log profile will not produce a message in logs
    [Teardown]    Teardown
    Setup Known Audit Log Profile
    Pattern Should Not Appear In Log Output    ${CURSOR}    "failed to configure kube-apiserver audit policy: unknown audit profile"

Config Flags Are Logged in Audit Flags
    [Documentation]    Check that flags specified in the audit config are set via the kube-apiserver flags
    [Teardown]    Teardown
    Setup Audit Flags
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxsize=\"1000\"
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxbackup=\"1000\"
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxage=\"1000\"

*** Keywords ***
Setup Suite
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Save Default MicroShift Config
    Save Journal Cursor

Teardown
    [Documentation]    Test case teardown
    Restore Default MicroShift Config
    Restart MicroShift

Teardown Suite
    [Documentation]    Test suite teardown
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
    ${merged}=    Extend MicroShift Config    ${BAD_LOG_LEVEL}
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Setup With Debug Log Level
    [Documentation]    Set log level to debug and restart
    ${merged}=    Extend MicroShift Config    ${BAD_LOG_LEVEL}
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Setup With Unknown Audit Log Profile
    [Documentation]    Set audit log profile to unknown value
    ${merged}=    Extend MicroShift Config    ${BAD_AUDIT_PROFILE}
    Upload MicroShift Config    ${merged}
    Restart MicroShift
    
Setup Known Audit Log Profile
    [Documentation]    Setup audit
    ${merged}=    Extend MicroShift Config    ${AUDIT_PROFILE}
    Upload MicroShift Config    ${merged}
    Restart MicroShift

Setup Audit Flags
    [Documentation]    Apply the audit config values set in ${AUDIT_FLAGS}
    ${merged}=    Extend MicroShift Config    ${}