*** Settings ***
Documentation       Regression tests for final greenboot wait failure diagnostics

Library             Collections
Resource            ../resources/ostree-health.resource

Test Setup          Reset Test State
Test Teardown       Remove Diagnostic Artifacts


*** Variables ***
${PRIMARY_FAILURE}          greenboot health check did not exit
${FULL_DIAGNOSTIC}          full service status${\n}complete journal output${\n}
${STATUS_DIAGNOSTIC}        bounded service status
${JOURNAL_DIAGNOSTIC}       bounded journal output
${USHIFT_HOST}              diagnostic-test-host
${USHIFT_USER}              diagnostic-test-user
${SSH_PORT}                 ${EMPTY}
${SSH_PRIV_KEY}             ${EMPTY}
${IPV6_HOST}                2001:db8::10


*** Test Cases ***
Successful Wait Does Not Collect Diagnostics
    [Documentation]    A successful final wait must not invoke any diagnostic collection.

    ${status}    ${error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics
    Should Be Equal    ${status}    PASS
    Should Be Equal    ${error}    ${None}
    Should Be Equal As Integers    ${RECONNECT_CALLS}    1
    Should Be Equal As Integers    ${STATUS_CALLS}    0
    Should Be Equal As Integers    ${JOURNAL_CALLS}    0
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    0
    Should Be Equal As Integers    ${DOWNLOAD_CALLS}    0
    ${artifacts}=    Find Diagnostic Artifacts
    Should Be Empty    ${artifacts}

Public Wait Keeps Ten Minute Retry Contract
    [Documentation]    The public keyword must keep its 10m wait, 15s retry, and reconnect behavior.

    Wait Until Greenboot Health Check Exited
    Should Be Equal    ${WAIT_TIMEOUT}    10m
    Should Be Equal    ${RETRY_INTERVAL}    15s
    Should Be Equal As Integers    ${RECONNECT_CALLS}    1
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    0

Transient Failure Reconnects On Every Retry Without Diagnostics
    [Documentation]    A retry that later passes reconnects each time and does not collect diagnostics.

    VAR    ${GREENBOOT_FAILURES_REMAINING}=    ${1}    scope=TEST
    ${status}    ${error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    2x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics
    Should Be Equal    ${status}    PASS
    Should Be Equal    ${error}    ${None}
    Should Be Equal As Integers    ${RECONNECT_CALLS}    2
    Should Be Equal As Integers    ${GREENBOOT_CHECK_CALLS}    2
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    0

Diagnostic Reconnect Has Finite Timeout
    [Documentation]    Failure diagnostics must use a bounded SSH reconnect.

    Make New SSH Connection For Greenboot Diagnostics
    Should Be Equal
    ...    ${RECONNECT_TIMEOUT}
    ...    ${GREENBOOT_DIAGNOSTIC_RECONNECT_TIMEOUT}

Original Wait Failure Is Preserved And Full Diagnostics Are Saved    # robocop: off=too-many-calls-in-test-case
    [Documentation]    The wrapper must preserve the wait failure and save full output under OUTPUTDIR.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    ${expected_status}    ${expected_error}=    Run Keyword And Ignore Error
    ...    Wait Until Keyword Succeeds    1x    0s
    ...    Stub Reconnect And Check Greenboot Health Check Exited
    Reset Test State
    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST

    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    Should Be Equal As Integers    ${STATUS_CALLS}    1
    Should Be Equal As Integers    ${JOURNAL_CALLS}    1
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    1
    Should Be Equal As Integers    ${DOWNLOAD_CALLS}    1
    Should Contain    ${JOURNAL_COMMAND}    tail -c 32000
    Should Contain    ${FULL_CAPTURE_COMMAND}    journalctl --no-pager
    Should Not Contain    ${FULL_CAPTURE_COMMAND}    tail -c 32000
    Should Contain    ${FULL_CAPTURE_COMMAND}    install -m 600
    Should Contain    ${FULL_CAPTURE_COMMAND}    owner="$SUDO_USER"
    Should Not Contain    ${FULL_CAPTURE_COMMAND}    chmod a+r
    Should Be Equal    ${STATUS_TIMEOUT}    ${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}
    Should Be Equal    ${FULL_CAPTURE_TIMEOUT}    ${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}
    Should Be Equal    ${CLEANUP_TIMEOUT}    ${SYSTEMD_DIAGNOSTIC_CLEANUP_TIMEOUT}
    Length Should Be    ${DIAGNOSTIC_PATHS}    1
    VAR    ${diagnostic_path}=    ${DIAGNOSTIC_PATHS}[0]
    Should Start With
    ...    ${diagnostic_path}
    ...    ${OUTPUTDIR}${/}greenboot-healthcheck-diagnostics-
    Should End With    ${diagnostic_path}    .log
    ${contents}=    OperatingSystem.Get File    ${diagnostic_path}
    Should Be Equal    ${contents}    ${FULL_DIAGNOSTIC}

Diagnostic Failures Do Not Mask Original Wait Failure
    [Documentation]    Reconnect, status, journal, and local-write errors must remain secondary.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    ${expected_status}    ${expected_error}=    Run Keyword And Ignore Error
    ...    Wait Until Keyword Succeeds    1x    0s
    ...    Stub Reconnect And Check Greenboot Health Check Exited
    Reset Test State
    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    VAR    ${FAIL_DIAGNOSTIC_RECONNECT}=    ${TRUE}    scope=TEST
    VAR    ${FAIL_STATUS}=    ${TRUE}    scope=TEST
    VAR    ${FAIL_JOURNAL}=    ${TRUE}    scope=TEST
    VAR    ${FAIL_DOWNLOAD}=    ${TRUE}    scope=TEST

    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    Should Be Equal As Integers    ${RECONNECT_CALLS}    2
    Should Be Equal As Integers    ${STATUS_CALLS}    1
    Should Be Equal As Integers    ${JOURNAL_CALLS}    1
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    1
    Should Be Equal As Integers    ${DOWNLOAD_CALLS}    1
    ${artifacts}=    Find Diagnostic Artifacts
    Should Be Empty    ${artifacts}

Full Capture Timeout And Cleanup Failure Preserve Original Wait Failure
    [Documentation]    Capture and cleanup timeouts must return the exact primary wait failure promptly.

    ${expected_status}    ${expected_error}=    Get Direct Wait Failure
    Reset Test State
    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    VAR    ${FAIL_FULL_CAPTURE}=    ${TRUE}    scope=TEST
    VAR    ${FAIL_CLEANUP}=    ${TRUE}    scope=TEST

    ${started}=    Evaluate    time.monotonic()    modules=time
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics
    ${elapsed}=    Evaluate    time.monotonic() - ${started}    modules=time
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    Should Be True    ${elapsed} < 2
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    1
    Should Be Equal As Integers    ${DOWNLOAD_CALLS}    0
    Should Be Equal As Integers    ${CLEANUP_CALLS}    1
    Should Be Equal    ${FULL_CAPTURE_TIMEOUT}    ${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}
    Should Be Equal    ${CLEANUP_TIMEOUT}    ${SYSTEMD_DIAGNOSTIC_CLEANUP_TIMEOUT}

Transfer Timeout Preserves Original Wait Failure
    [Documentation]    A transfer timeout removes partial data without replacing the wait failure.

    ${expected_status}    ${expected_error}=    Get Direct Wait Failure
    Reset Test State
    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    VAR    ${TIMEOUT_DOWNLOAD}=    ${TRUE}    scope=TEST
    VAR    ${SYSTEMD_DIAGNOSTIC_TRANSFER_TIMEOUT}=    100ms    scope=TEST
    VAR    ${USHIFT_HOST}=    ${IPV6_HOST}    scope=TEST

    ${started}=    Evaluate    time.monotonic()    modules=time
    ${actual_status}    ${actual_error}=    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics
    ${elapsed}=    Evaluate    time.monotonic() - ${started}    modules=time
    Should Be Equal    ${actual_status}    ${expected_status}
    Should Be Equal    ${actual_error}    ${expected_error}
    Should Be True    ${elapsed} < 2
    Should Be Equal As Integers    ${DOWNLOAD_CALLS}    1
    Should Be Equal As Integers    ${CLEANUP_CALLS}    1
    Should Not Be Empty    ${TIMED_PARTIAL_PATH}
    OperatingSystem.File Should Not Exist    ${TIMED_PARTIAL_PATH}
    ${artifacts}=    Find Diagnostic Artifacts
    Should Be Empty    ${artifacts}

Journal Pipeline Bounds Combined Output And Reports Journal Failure
    [Documentation]    Pipefail must expose a journal failure while all displayed output stays bounded.

    Create Failing Journalctl
    ${journal_text}=    Systemctl Get Service Journal Tail Using Keyword
    ...    greenboot-healthcheck.service    Run Local Journal Command
    ${displayed_bytes}=    Evaluate    len($journal_text.encode("utf-8"))
    Should Contain    ${journal_text}    journal pipeline exit code: 23
    Should Contain    ${journal_text}    0000000000
    Should Be Equal As Integers    ${displayed_bytes}    32000
    Should Contain    ${JOURNAL_COMMAND}    bash -o pipefail
    Should Contain    ${JOURNAL_COMMAND}    2>&1 | tail -c 32000
    Should Contain    ${JOURNAL_COMMAND}    tail -c 32000
    Should Be Equal    ${JOURNAL_TIMEOUT}    ${SYSTEMD_DIAGNOSTIC_COMMAND_TIMEOUT}

Unsafe Shell Arguments Are Rejected Before Execution
    [Documentation]    Unit names and diagnostic IDs must be validated before shell interpolation.

    ${unit_status}    ${unit_error}=    Run Keyword And Ignore Error
    ...    Systemctl Get Service Status    greenboot.service; touch unsafe
    Should Be Equal    ${unit_status}    FAIL
    Should Contain    ${unit_error}    does not match
    Should Be Equal As Integers    ${STATUS_CALLS}    0
    ${option_status}    ${option_error}=    Run Keyword And Ignore Error
    ...    Systemctl Get Service Status    -help.service
    Should Be Equal    ${option_status}    FAIL
    Should Contain    ${option_error}    does not match
    Should Be Equal As Integers    ${STATUS_CALLS}    0
    ${id_status}    ${id_error}=    Run Keyword And Ignore Error
    ...    Systemctl Save Service Status And Logs Using Keyword
    ...    greenboot-healthcheck.service    ${OUTPUTDIR}${/}unused.log
    ...    invalid/id    Stub Download Systemd Diagnostic File
    Should Be Equal    ${id_status}    FAIL
    Should Contain    ${id_error}    does not match
    Should Be Equal As Integers    ${FULL_CAPTURE_CALLS}    0

Multiple Failures Use Distinct Diagnostic Files
    [Documentation]    Repeated failures must create separate, complete diagnostic artifacts.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    Run Stubbed Greenboot Wait
    Run Stubbed Greenboot Wait
    Length Should Be    ${DIAGNOSTIC_PATHS}    2
    Should Not Be Equal    ${DIAGNOSTIC_PATHS}[0]    ${DIAGNOSTIC_PATHS}[1]
    FOR    ${diagnostic_path}    IN    @{DIAGNOSTIC_PATHS}
        OperatingSystem.File Should Exist    ${diagnostic_path}
        ${contents}=    OperatingSystem.Get File    ${diagnostic_path}
        Should Be Equal    ${contents}    ${FULL_DIAGNOSTIC}
    END


*** Keywords ***
Reset Test State    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Reset failure controls, counters, and diagnostic files for one test.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${FALSE}    scope=TEST
    VAR    ${GREENBOOT_FAILURES_REMAINING}=    ${0}    scope=TEST
    VAR    ${FAIL_DIAGNOSTIC_RECONNECT}=    ${FALSE}    scope=TEST
    VAR    ${FAIL_STATUS}=    ${FALSE}    scope=TEST
    VAR    ${FAIL_JOURNAL}=    ${FALSE}    scope=TEST
    VAR    ${FAIL_DOWNLOAD}=    ${FALSE}    scope=TEST
    VAR    ${FAIL_FULL_CAPTURE}=    ${FALSE}    scope=TEST
    VAR    ${FAIL_CLEANUP}=    ${FALSE}    scope=TEST
    VAR    ${TIMEOUT_DOWNLOAD}=    ${FALSE}    scope=TEST
    VAR    ${FAKE_JOURNAL_BIN}=    ${EMPTY}    scope=TEST
    VAR    ${EXPECTED_REMOTE_OPERAND}=    ${EMPTY}    scope=TEST
    VAR    ${TIMED_PARTIAL_PATH}=    ${EMPTY}    scope=TEST
    VAR    ${RECONNECT_CALLS}=    ${0}    scope=TEST
    VAR    ${GREENBOOT_CHECK_CALLS}=    ${0}    scope=TEST
    VAR    ${STATUS_CALLS}=    ${0}    scope=TEST
    VAR    ${JOURNAL_CALLS}=    ${0}    scope=TEST
    VAR    ${FULL_CAPTURE_CALLS}=    ${0}    scope=TEST
    VAR    ${DOWNLOAD_CALLS}=    ${0}    scope=TEST
    VAR    ${CLEANUP_CALLS}=    ${0}    scope=TEST
    VAR    ${JOURNAL_RC}=    ${0}    scope=TEST
    VAR    ${JOURNAL_COMMAND}=    ${EMPTY}    scope=TEST
    VAR    ${FULL_CAPTURE_COMMAND}=    ${EMPTY}    scope=TEST
    VAR    ${FULL_CAPTURE_TIMEOUT}=    ${EMPTY}    scope=TEST
    VAR    ${STATUS_TIMEOUT}=    ${EMPTY}    scope=TEST
    VAR    ${JOURNAL_TIMEOUT}=    ${EMPTY}    scope=TEST
    VAR    ${CLEANUP_TIMEOUT}=    ${EMPTY}    scope=TEST
    VAR    ${WAIT_TIMEOUT}=    ${EMPTY}    scope=TEST
    VAR    ${RETRY_INTERVAL}=    ${EMPTY}    scope=TEST
    VAR    ${RECONNECT_TIMEOUT}=    ${EMPTY}    scope=TEST
    VAR    @{DIAGNOSTIC_PATHS}=    @{EMPTY}    scope=TEST
    Remove Diagnostic Artifacts

Remove Diagnostic Artifacts
    [Documentation]    Remove diagnostic files created by the current test.

    OperatingSystem.Remove Files
    ...    ${OUTPUTDIR}${/}greenboot-healthcheck-diagnostics-*.log
    Run Keyword And Ignore Error
    ...    OperatingSystem.Remove Directory    ${OUTPUTDIR}${/}fake-journal-bin    recursive=True

Create Failing Journalctl
    [Documentation]    Create a local journalctl fake that emits more than 32 KB on stderr and fails.

    VAR    ${FAKE_JOURNAL_BIN}=    ${OUTPUTDIR}${/}fake-journal-bin    scope=TEST
    OperatingSystem.Create Directory    ${FAKE_JOURNAL_BIN}
    VAR    ${fake_journalctl}=    ${FAKE_JOURNAL_BIN}${/}journalctl
    OperatingSystem.Create File
    ...    ${fake_journalctl}
    ...    \#!/bin/bash${\n}printf '%040000d' 0 >&2${\n}exit 23${\n}
    ${result}=    Process.Run Process    chmod    +x    ${fake_journalctl}
    Should Be Equal As Integers    ${result.rc}    0

Run Local Journal Command
    [Documentation]    Execute the constructed journal command against the local failing fake.
    [Arguments]    ${command}    &{options}

    VAR    ${JOURNAL_COMMAND}=    ${command}    scope=TEST
    VAR    ${JOURNAL_TIMEOUT}=    ${options}[timeout]    scope=TEST
    ${result}=    Process.Run Process
    ...    bash    -c    ${command}
    ...    env:PATH=${FAKE_JOURNAL_BIN}:%{PATH}
    RETURN    ${result.stdout}    ${result.stderr}    ${result.rc}

Find Diagnostic Artifacts
    [Documentation]    Return greenboot diagnostic artifacts from Robot's output directory.

    ${artifacts}=    OperatingSystem.List Files In Directory
    ...    ${OUTPUTDIR}
    ...    pattern=greenboot-healthcheck-diagnostics-*.log
    ...    absolute=True
    RETURN    ${artifacts}

Run Stubbed Greenboot Wait
    [Documentation]    Run the greenboot wait orchestration with short retries and test stubs.

    Run Keyword And Ignore Error
    ...    Wait Until Greenboot Health Check Exited Using Keywords
    ...    1x    0s    Stub Reconnect And Check Greenboot Health Check Exited
    ...    Stub Collect Greenboot Health Check Failure Diagnostics

Get Direct Wait Failure
    [Documentation]    Return the baseline status and message from the unwrapped failing wait.

    VAR    ${GREENBOOT_SHOULD_FAIL}=    ${TRUE}    scope=TEST
    ${status}    ${error}=    Run Keyword And Ignore Error
    ...    BuiltIn.Wait Until Keyword Succeeds    1x    0s
    ...    Stub Reconnect And Check Greenboot Health Check Exited
    RETURN    ${status}    ${error}

Wait Until Keyword Succeeds
    [Documentation]    Record wait settings, then delegate to the BuiltIn implementation.
    [Arguments]    ${timeout}    ${retry_period}    ${keyword}    @{args}

    VAR    ${WAIT_TIMEOUT}=    ${timeout}    scope=TEST
    VAR    ${RETRY_INTERVAL}=    ${retry_period}    scope=TEST
    BuiltIn.Wait Until Keyword Succeeds
    ...    ${timeout}    ${retry_period}    ${keyword}    @{args}

Reconnect And Check Greenboot Health Check Exited
    [Documentation]    Replace the public keyword's host operation with local unit-test stubs.

    Stub Reconnect And Check Greenboot Health Check Exited

Collect Greenboot Health Check Failure Diagnostics
    [Documentation]    Replace the public keyword's diagnostics with the local test stub.

    Stub Collect Greenboot Health Check Failure Diagnostics

Make New SSH Connection
    [Documentation]    Record the timeout supplied by the diagnostic reconnect wrapper.
    [Arguments]    ${timeout}=${None}

    VAR    ${RECONNECT_TIMEOUT}=    ${timeout}    scope=TEST

Stub Reconnect And Check Greenboot Health Check Exited
    [Documentation]    Reconnect and check greenboot using test stubs.

    Stub Make New SSH Connection
    Stub Greenboot Health Check Exited

Stub Make New SSH Connection
    [Documentation]    Stub reconnection and optionally fail during diagnostic collection.

    ${next_count}=    Evaluate    ${RECONNECT_CALLS} + 1
    VAR    ${RECONNECT_CALLS}=    ${next_count}    scope=TEST
    IF    ${FAIL_DIAGNOSTIC_RECONNECT} and ${GREENBOOT_CHECK_CALLS} > 0
        Fail    simulated diagnostic reconnect failure
    END

Stub Greenboot Health Check Exited
    [Documentation]    Stub the final greenboot state check.

    ${next_count}=    Evaluate    ${GREENBOOT_CHECK_CALLS} + 1
    VAR    ${GREENBOOT_CHECK_CALLS}=    ${next_count}    scope=TEST
    IF    ${GREENBOOT_FAILURES_REMAINING} > 0
        ${remaining}=    Evaluate    ${GREENBOOT_FAILURES_REMAINING} - 1
        VAR    ${GREENBOOT_FAILURES_REMAINING}=    ${remaining}    scope=TEST
        Fail    ${PRIMARY_FAILURE}
    END
    IF    ${GREENBOOT_SHOULD_FAIL}    Fail    ${PRIMARY_FAILURE}

Execute Command    # robocop: off=too-many-calls-in-keyword
    [Documentation]    Stub bounded status/journal and full remote diagnostic commands.
    [Arguments]    ${command}    &{options}

    IF    $command.startswith('rm -f --')
        ${next_count}=    Evaluate    ${CLEANUP_CALLS} + 1
        VAR    ${CLEANUP_CALLS}=    ${next_count}    scope=TEST
        VAR    ${CLEANUP_TIMEOUT}=    ${options}[timeout]    scope=TEST
        IF    ${FAIL_CLEANUP}    Fail    simulated cleanup command timeout
    ELSE IF    'systemd-diagnostics-' in $command and 'journalctl' in $command
        ${next_count}=    Evaluate    ${FULL_CAPTURE_CALLS} + 1
        VAR    ${FULL_CAPTURE_CALLS}=    ${next_count}    scope=TEST
        VAR    ${FULL_CAPTURE_COMMAND}=    ${command}    scope=TEST
        VAR    ${FULL_CAPTURE_TIMEOUT}=    ${options}[timeout]    scope=TEST
        IF    ${FAIL_FULL_CAPTURE}
            Fail    simulated full capture command timeout
        END
    ELSE IF    'systemctl status --no-pager --full' in $command
        ${next_count}=    Evaluate    ${STATUS_CALLS} + 1
        VAR    ${STATUS_CALLS}=    ${next_count}    scope=TEST
        VAR    ${STATUS_TIMEOUT}=    ${options}[timeout]    scope=TEST
        IF    ${FAIL_STATUS}    Fail    simulated status command failure
        RETURN    ${STATUS_DIAGNOSTIC}    ${EMPTY}    ${0}
    ELSE IF    'tail -c 32000' in $command
        ${next_count}=    Evaluate    ${JOURNAL_CALLS} + 1
        VAR    ${JOURNAL_CALLS}=    ${next_count}    scope=TEST
        VAR    ${JOURNAL_COMMAND}=    ${command}    scope=TEST
        VAR    ${JOURNAL_TIMEOUT}=    ${options}[timeout]    scope=TEST
        IF    ${FAIL_JOURNAL}    Fail    simulated journal command failure
        RETURN    ${JOURNAL_DIAGNOSTIC}    ${EMPTY}    ${JOURNAL_RC}
    END
    RETURN    ${EMPTY}    ${EMPTY}    ${0}

Stub Collect Greenboot Health Check Failure Diagnostics
    [Documentation]    Collect diagnostics through the real orchestration and test stubs.

    Collect Greenboot Health Check Failure Diagnostics Using Keywords
    ...    Stub Make New SSH Connection
    ...    Systemctl Print Service Status And Logs
    ...    Stub Systemctl Save Service Status And Logs

Stub Systemctl Save Service Status And Logs
    [Documentation]    Save through the real systemd helper and a stubbed local download.
    [Arguments]    ${unit_name}    ${local_path}    ${diagnostic_id}

    IF    ${TIMEOUT_DOWNLOAD}
        VAR    ${download_keyword}=    Stub Timed Out Download
    ELSE
        VAR    ${download_keyword}=    Stub Download Systemd Diagnostic File
    END
    Systemctl Save Service Status And Logs Using Keyword
    ...    ${unit_name}    ${local_path}    ${diagnostic_id}    ${download_keyword}

Stub Download Systemd Diagnostic File
    [Documentation]    Stub a download by writing full diagnostics to the requested local path.
    [Arguments]    ${remote_path}    ${local_path}

    ${next_count}=    Evaluate    ${DOWNLOAD_CALLS} + 1
    VAR    ${DOWNLOAD_CALLS}=    ${next_count}    scope=TEST
    Log    Simulating download from ${remote_path}    DEBUG
    Append To List    ${DIAGNOSTIC_PATHS}    ${local_path}
    IF    ${FAIL_DOWNLOAD}    Fail    simulated local diagnostic write failure
    OperatingSystem.Create File    ${local_path}    ${FULL_DIAGNOSTIC}

Stub Timed Out Download
    [Documentation]    Exercise the production transfer deadline with a local sleeping process.
    [Arguments]    ${remote_path}    ${local_path}

    ${next_count}=    Evaluate    ${DOWNLOAD_CALLS} + 1
    VAR    ${DOWNLOAD_CALLS}=    ${next_count}    scope=TEST
    VAR    ${EXPECTED_REMOTE_OPERAND}=    ${USHIFT_USER}@[${USHIFT_HOST}]:${remote_path}    scope=TEST
    VAR    ${TIMED_PARTIAL_PATH}=    ${local_path}.partial    scope=TEST
    Download Systemd Diagnostic File Using Keyword
    ...    ${remote_path}    ${local_path}    Stub Timed Transfer Process

Stub Timed Transfer Process
    [Documentation]    Replace scp with a process that must be killed by the configured timeout.
    [Arguments]    ${command}    @{args}    &{configuration}

    Should Be Equal    ${command}    scp
    Should Be Equal    ${configuration}[timeout]    ${SYSTEMD_DIAGNOSTIC_TRANSFER_TIMEOUT}
    Should Be Equal    ${configuration}[on_timeout]    kill
    List Should Contain Value    ${args}    StrictHostKeyChecking=accept-new
    List Should Not Contain Value    ${args}    StrictHostKeyChecking=no
    Should Be Equal    ${args}[-2]    ${EXPECTED_REMOTE_OPERAND}
    Should Be Equal    ${args}[-1]    ${TIMED_PARTIAL_PATH}
    ${result}=    Process.Run Process
    ...    bash    -c    printf partial > "$1"; sleep 5
    ...    timed-transfer    ${TIMED_PARTIAL_PATH}
    ...    timeout=${configuration}[timeout]    on_timeout=${configuration}[on_timeout]
    ...    stdout=DEVNULL    stderr=DEVNULL
    RETURN    ${result}
