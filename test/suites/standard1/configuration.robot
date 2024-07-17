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
${AUDIT_FLAGS}          SEPARATOR=\n
...                     apiServer:
...                     \ \ auditLog:
...                     \ \ \ \ maxFileSize: 1000
...                     \ \ \ \ maxFiles: 1000
...                     \ \ \ \ maxFileAge: 1000


*** Test Cases ***
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
