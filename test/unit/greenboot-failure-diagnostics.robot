*** Settings ***
Documentation       Regression tests for public Greenboot failure summaries

Library             Collections
Library             OperatingSystem
Library             greenboot_diagnostics_testlib.GreenbootDiagnosticsTestlib
...                 AS    DiagnosticSSH
Resource            ../resources/ostree-health.resource

Test Setup          Reset Test State


*** Variables ***
${PRIMARY_FAILURE}      greenboot health check did not exit
${USHIFT_HOST}          diagnostic.invalid
${USHIFT_USER}          diagnostic-user
${SSH_PORT}             ${EMPTY}
${SSH_PRIV_KEY}         ${EMPTY}


*** Test Cases ***
Privacy Sentinel Uses Synthetic Default When Environment Is Absent
    [Documentation]    Direct Robot runs have a synthetic sentinel without setup.

    ${uses_default}=    DiagnosticSSH.Privacy Sentinel Uses Synthetic Default
    Should Be True    ${uses_default}

Privacy Sentinel Uses Custom Environment Value
    [Documentation]    The verification script's exported sentinel takes precedence.

    ${uses_environment}=    DiagnosticSSH.Privacy Sentinel Uses Environment Value
    Should Be True    ${uses_environment}

Direct Healthcheck Success Does Not Capture Output Or Diagnostics
    [Documentation]    Success keeps the 600s command default and captures no raw output.

    Wait For MicroShift Healthcheck Success
    ...    execute_keyword=DiagnosticSSH.Execute Sensitive Healthcheck Command
    ...    diagnostics_keyword=DiagnosticSSH.Collect Stub Diagnostic Summary
    ${observed}=    DiagnosticSSH.Get Healthcheck Observations
    Should Be Equal As Integers    ${observed}[calls]    1
    Should Be Equal
    ...    ${observed}[command]
    ...    microshift healthcheck -v=2 --timeout="600s"
    Should Be Equal As Strings    ${observed}[options][return_stdout]    False
    Should Be Equal As Strings    ${observed}[options][return_stderr]    False
    Should Be Equal As Strings    ${observed}[options][return_rc]    True
    ${diagnostics}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal As Integers    ${diagnostics}[calls]    0
    ${artifacts}=    Find Public Summary Artifacts
    Should Be Empty    ${artifacts}

Direct Healthcheck Failure Collects Before Preserving Assertion
    [Documentation]    A command rc failure saves a summary before the original assertion fails.

    DiagnosticSSH.Set Healthcheck Return Code    17
    ${expected_status}    ${expected_error}=    Run Keyword And Ignore Error
    ...    Should Be Equal As Integers    0    17
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait For MicroShift Healthcheck Success
    ...    600s
    ...    DiagnosticSSH.Execute Sensitive Healthcheck Command
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    ${diagnostics}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal As Integers    ${diagnostics}[calls]    1
    Should Be Equal    ${diagnostics}[trigger]    direct-healthcheck
    Should Be Equal As Integers    ${diagnostics}[primary_rc]    17
    Public Summary Should Contain
    ...    trigger=direct-healthcheck
    ...    primary_exit_code=17

Direct Escaping Diagnostic Does Not Replace Command Failure
    [Documentation]    An escaping sensitive exception never becomes a Robot child failure.

    DiagnosticSSH.Set Healthcheck Return Code    18
    ${expected_status}    ${expected_error}=    Run Keyword And Ignore Error
    ...    Should Be Equal As Integers    0    18
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait For MicroShift Healthcheck Success
    ...    600s
    ...    DiagnosticSSH.Execute Sensitive Healthcheck Command
    ...    DiagnosticSSH.Raise Sensitive Diagnostic Exception
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    ${artifacts}=    Find Public Summary Artifacts
    Should Be Empty    ${artifacts}

Direct Malformed Diagnostic Does Not Replace Command Failure
    [Documentation]    A sensitive malformed return becomes only a fixed generic category.

    DiagnosticSSH.Set Healthcheck Return Code    20
    ${expected_status}    ${expected_error}=    Run Keyword And Ignore Error
    ...    Should Be Equal As Integers    0    20
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait For MicroShift Healthcheck Success
    ...    600s
    ...    DiagnosticSSH.Execute Sensitive Healthcheck Command
    ...    DiagnosticSSH.Return Sensitive Malformed Diagnostic Result
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    ${artifacts}=    Find Public Summary Artifacts
    Should Be Empty    ${artifacts}

Successful Final Wait Does Not Collect Diagnostics
    [Documentation]    A successful wait reconnects once and produces no summary.

    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Greenboot Check
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary
    Should Be Equal As Integers    ${GREENBOOT_CHECK_CALLS}    1
    ${diagnostics}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal As Integers    ${diagnostics}[calls]    0
    ${artifacts}=    Find Public Summary Artifacts
    Should Be Empty    ${artifacts}

Public Final Wait Keeps Retry Defaults
    [Documentation]    The public wait keeps its 10m timeout and 15s retry interval.

    Wait Until Greenboot Health Check Exited
    ...    Stub Greenboot Check
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary
    Should Be Equal    ${WAIT_TIMEOUT}    10m
    Should Be Equal    ${RETRY_INTERVAL}    15s
    ${diagnostics}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal As Integers    ${diagnostics}[calls]    0

Transient Wait Failure Does Not Collect Diagnostics
    [Documentation]    A retry that later passes performs no failure collection.

    VAR    ${GREENBOOT_FAILURES_REMAINING}=    ${1}    scope=TEST
    Wait Until Greenboot Health Check Exited Using Keywords
    ...    2x    0s    Stub Greenboot Check
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary
    Should Be Equal As Integers    ${GREENBOOT_CHECK_CALLS}    2
    ${diagnostics}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal As Integers    ${diagnostics}[calls]    0

Final Wait Failure Is Preserved And Summarized
    [Documentation]    The final wait failure remains primary after safe collection.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    ${expected_status}    ${expected_error}=    Get Direct Wait Failure
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Greenboot Check
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    ${diagnostics}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal    ${diagnostics}[trigger]    greenboot-final-wait
    Should Be Equal    ${diagnostics}[primary_rc]    ${None}
    Public Summary Should Contain
    ...    trigger=greenboot-final-wait
    ...    primary_exit_code=unavailable

Final Escaping Diagnostic Does Not Alter Wait Failure
    [Documentation]    An escaping sensitive exception cannot fail the user-keyword teardown.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    ${expected_status}    ${expected_error}=    Get Direct Wait Failure
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Greenboot Check
    ...    DiagnosticSSH.Raise Sensitive Diagnostic Exception
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    ${artifacts}=    Find Public Summary Artifacts
    Should Be Empty    ${artifacts}

Final Malformed Diagnostic Does Not Alter Wait Failure
    [Documentation]    A sensitive malformed return cannot fail the user-keyword teardown.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    ${expected_status}    ${expected_error}=    Get Direct Wait Failure
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Greenboot Check
    ...    DiagnosticSSH.Return Sensitive Malformed Diagnostic Result
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    ${artifacts}=    Find Public Summary Artifacts
    Should Be Empty    ${artifacts}

Collector Uses Finite Bounds And Property Query Only
    [Documentation]    Reconnect and query deadlines remain finite and no raw collector is used.

    Stub Diagnostic Collection    greenboot-final-wait
    ${observed}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal
    ...    ${observed}[reconnect_timeout]
    ...    ${GREENBOOT_DIAGNOSTIC_RECONNECT_TIMEOUT}
    Should Be Equal
    ...    ${observed}[command_timeout]
    ...    ${GREENBOOT_DIAGNOSTIC_COMMAND_TIMEOUT}
    Should Contain
    ...    ${observed}[command]
    ...    ${GREENBOOT_DIAGNOSTIC_REMOTE_TIMEOUT} systemctl show
    Should Not Contain    ${observed}[command]    systemctl status
    Should Not Contain    ${observed}[command]    journalctl
    Should Be Equal    ${observed}[return_stderr]    ${FALSE}

Unexpected Property Values Become A Generic Category
    [Documentation]    Unallowlisted fields and values are replaced, never copied.

    DiagnosticSSH.Set Diagnostic SSH Mode    invalid-response
    Stub Diagnostic Collection    greenboot-final-wait
    Public Summary Should Contain
    ...    error_category=invalid_response
    ...    service_active_state=unavailable
    ...    service_sub_state=unavailable
    ...    service_result=unavailable
    ...    exec_main_status=unavailable

Reconnect Failure Is Saved As A Generic Category
    [Documentation]    Login failure keeps the prior authenticated connection usable for cleanup.

    DiagnosticSSH.Set Diagnostic SSH Mode    reconnect-failure
    Stub Diagnostic Collection    greenboot-final-wait
    DiagnosticSSH.Simulate Clock Restore Cleanup
    Public Summary Should Contain
    ...    error_category=reconnect_failed
    ...    collection_result=failed
    ${observed}=    DiagnosticSSH.Get Diagnostic SSH Observations
    Should Be Equal As Integers    ${observed}[current_connection]    1
    Should Be True    ${observed}[current_connection_authenticated]
    Should Be Equal As Integers    ${observed}[clock_restore_calls]    1

Nonzero Query Output Is Never Published
    [Documentation]    Sensitive stdout from a failed query becomes a fixed category only.

    DiagnosticSSH.Set Diagnostic SSH Mode    query-failure
    Stub Diagnostic Collection    greenboot-final-wait
    Public Summary Should Contain
    ...    error_category=query_failed
    ...    query_exit_code=23

Elapsed Query Timeout Is Generic And Bounded
    [Documentation]    An elapsed client timeout becomes a fixed category within a short bound.

    VAR    ${GREENBOOT_DIAGNOSTIC_COMMAND_TIMEOUT}=    1s    scope=TEST
    DiagnosticSSH.Set Diagnostic SSH Mode    elapsed-timeout
    ${started}=    Evaluate    time.monotonic()    modules=time
    Stub Diagnostic Collection    greenboot-final-wait
    ${elapsed}=    Evaluate    time.monotonic() - ${started}    modules=time
    # The fake sleeps command_timeout+0.01s (~1.01s) then raises, so >=1 proves the
    # 1s override was consumed. The ceiling proves the 1s override reached the client
    # instead of the 30s default (which would elapse ~30s); 10s keeps it well clear of
    # that regression while leaving generous headroom for loaded-CI scheduling jitter.
    Should Be True    1 <= ${elapsed} < 10
    Public Summary Should Contain
    ...    error_category=query_failed
    ...    collection_result=failed

Repeated Failures Use Unique Summary Artifacts
    [Documentation]    Every failure receives its own fixed-format public summary.

    Stub Diagnostic Collection    greenboot-final-wait
    Stub Diagnostic Collection    greenboot-final-wait
    ${artifacts}=    Find Public Summary Artifacts
    Length Should Be    ${artifacts}    2
    Should Not Be Equal    ${artifacts}[0]    ${artifacts}[1]

Sensitive Mock Values Never Reach Public Summary
    [Documentation]    Leave artifacts from sensitive stdout, stderr, values, and exceptions for scanning.

    DiagnosticSSH.Set Healthcheck Return Code    19
    Run Keyword And Ignore Error
    ...    Wait For MicroShift Healthcheck Success
    ...    600s
    ...    DiagnosticSSH.Execute Sensitive Healthcheck Command
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary
    DiagnosticSSH.Set Diagnostic SSH Mode    invalid-response
    Stub Diagnostic Collection    greenboot-final-wait
    DiagnosticSSH.Set Diagnostic SSH Mode    query-exception
    Stub Diagnostic Collection    greenboot-final-wait
    DiagnosticSSH.Set Diagnostic SSH Mode    reconnect-failure
    Stub Diagnostic Collection    greenboot-final-wait
    DiagnosticSSH.Set Diagnostic SSH Mode    query-failure
    Stub Diagnostic Collection    greenboot-final-wait
    VAR    ${GREENBOOT_DIAGNOSTIC_COMMAND_TIMEOUT}=    1s    scope=TEST
    DiagnosticSSH.Set Diagnostic SSH Mode    elapsed-timeout
    Stub Diagnostic Collection    greenboot-final-wait
    ${artifacts}=    Find Public Summary Artifacts
    Length Should Be    ${artifacts}    6

Login Banner Is Suppressed By The Log Level Guard
    [Documentation]    SSHLibrary logs the server MOTD at INFO on login, so Set Log Level NONE
    ...    is the only guard keeping sensitive login output out of the published Robot log.
    ...    Exercise the collector directly (its own guard is the only one) and through the safe
    ...    wrapper, and leave the login emissions for the sentinel scan so removing the guard
    ...    makes verify-rf.sh fail instead of passing green.

    # Direct public entry point: only the collector's own Set Log Level NONE applies.
    DiagnosticSSH.Collect Stub Diagnostic Summary    greenboot-final-wait
    # Safe wrapper path: leaks only if every log level guard is removed.
    Stub Diagnostic Collection    greenboot-final-wait
    ${artifacts}=    Find Public Summary Artifacts
    Length Should Be    ${artifacts}    2


*** Keywords ***
Reset Test State
    [Documentation]    Reset local stubs and remove prior public summaries.

    DiagnosticSSH.Reset Diagnostic SSH
    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${FALSE}    scope=TEST
    VAR    ${GREENBOOT_FAILURES_REMAINING}=    ${0}    scope=TEST
    VAR    ${GREENBOOT_CHECK_CALLS}=    ${0}    scope=TEST
    VAR    ${WAIT_TIMEOUT}=    ${None}    scope=TEST
    VAR    ${RETRY_INTERVAL}=    ${None}    scope=TEST
    OperatingSystem.Remove Files
    ...    ${OUTPUTDIR}${/}greenboot-healthcheck-summary-*.txt

Wait Until Keyword Succeeds
    [Documentation]    Record wait settings and delegate to Robot's implementation.
    [Arguments]    ${timeout}    ${retry_period}    ${keyword}    @{args}

    VAR    ${WAIT_TIMEOUT}=    ${timeout}    scope=TEST
    VAR    ${RETRY_INTERVAL}=    ${retry_period}    scope=TEST
    BuiltIn.Wait Until Keyword Succeeds
    ...    ${timeout}    ${retry_period}    ${keyword}    @{args}

Stub Diagnostic Collection
    [Documentation]    Route production orchestration through the direct Python fake.
    [Arguments]    ${trigger}    ${primary_exit_code}=${None}

    Collect Greenboot Health Check Failure Diagnostics Using Keywords
    ...    ${trigger}
    ...    ${primary_exit_code}
    ...    DiagnosticSSH.Collect Stub Diagnostic Summary

Stub Greenboot Check
    [Documentation]    Fail until the configured number of retries is exhausted.

    ${next_count}=    Evaluate    ${GREENBOOT_CHECK_CALLS} + 1
    VAR    ${GREENBOOT_CHECK_CALLS}=    ${next_count}    scope=TEST
    IF    ${GREENBOOT_FAILURES_REMAINING} > 0
        ${remaining}=    Evaluate    ${GREENBOOT_FAILURES_REMAINING} - 1
        VAR    ${GREENBOOT_FAILURES_REMAINING}=    ${remaining}    scope=TEST
        Fail    ${PRIMARY_FAILURE}
    END
    IF    ${GREENBOOT_SHOULD_FAIL}    Fail    ${PRIMARY_FAILURE}

Get Direct Wait Failure
    [Documentation]    Return the unwrapped wait's status and exact failure message.

    ${status}    ${error}=    Run Keyword And Ignore Error
    ...    BuiltIn.Wait Until Keyword Succeeds    1x    0s    Stub Greenboot Check
    RETURN    ${status}    ${error}

Find Public Summary Artifacts
    [Documentation]    Return all public summary artifacts in Robot's output directory.

    ${artifacts}=    OperatingSystem.List Files In Directory
    ...    ${OUTPUTDIR}
    ...    pattern=greenboot-healthcheck-summary-*.txt
    ...    absolute=True
    RETURN    ${artifacts}

Public Summary Should Contain
    [Documentation]    Assert that the single summary artifact has expected fixed fields.
    [Arguments]    @{expected_lines}

    ${artifacts}=    Find Public Summary Artifacts
    Length Should Be    ${artifacts}    1
    ${contents}=    OperatingSystem.Get File    ${artifacts}[0]
    FOR    ${expected_line}    IN    @{expected_lines}
        Should Contain    ${contents}    ${expected_line}${\n}
    END
