*** Settings ***
Documentation       Tests for feature gate configuration

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}                       ${EMPTY}    # The journal cursor before restarting MicroShift
${CUSTOM_FEATURE_GATES}         SEPARATOR=\n
...                             apiServer:
...                             \ \ featureGates:
...                             \ \ \ \ featureSet: CustomNoUpgrade
...                             \ \ \ \ customNoUpgrade:
...                             \ \ \ \ \ \ enabled:
...                             \ \ \ \ \ \ \ \ - TestFeatureEnabled
...                             \ \ \ \ \ \ disabled:
...                             \ \ \ \ \ \ \ \ - TestFeatureDisabled
${DIFFERENT_FEATURE_GATES}      SEPARATOR=\n
...                             apiServer:
...                             \ \ featureGates:
...                             \ \ \ \ featureSet: CustomNoUpgrade
...                             \ \ \ \ customNoUpgrade:
...                             \ \ \ \ \ \ enabled:
...                             \ \ \ \ \ \ \ \ - DifferentTestFeature
${FEATURE_GATE_LOCK_FILE}       /var/lib/microshift/no-upgrade


*** Test Cases ***
Custom Feature Gates Are Passed To Kube APIServer
    [Documentation]    Check that custom feature gates specified in the MicroShift config are passed to and logged by the
    ...    kube-apiserver. This test verifies that arbitrary feature gate values are correctly propagated from the
    ...    MicroShift configuration to the kube-apiserver, regardless of whether the feature gates are valid or have any effect.
    ...    Also verify that feature gate lock file is created when custom feature gates are configured.
    ...    The lock file prevents upgrades and configuration changes when CustomNoUpgrade feature set is used.
    [Setup]    Setup Custom Feature Gates Test
    Wait Until Keyword Succeeds    2 min    5 sec
    ...    Pattern Should Appear In Log Output    ${CURSOR}    kube:feature-gates=.*TestFeatureEnabled=true
    Wait Until Keyword Succeeds    2 min    5 sec
    ...    Pattern Should Appear In Log Output    ${CURSOR}    kube:feature-gates=.*TestFeatureDisabled=false
    Wait Until Keyword Succeeds    2 min    5 sec
    ...    Feature Gate Lock File Should Exist
    Feature Gate Lock File Should Contain Feature Gates    CustomNoUpgrade    TestFeatureEnabled
    [Teardown]    Teardown Custom Feature Gates Test

Feature Gate Config Change Blocked After Lock Created
    [Documentation]    Verify that changing feature gate config is blocked after lock file exists.
    ...    MicroShift must refuse to start if feature gates change after CustomNoUpgrade is set.
    [Setup]    Setup Custom Feature Gates Test
    Stop MicroShift
    Drop In MicroShift Config    ${DIFFERENT_FEATURE_GATES}    10-featuregates
    Save Journal Cursor
    MicroShift Should Fail To Start
    Pattern Should Appear In Log Output    ${CURSOR}    feature gate configuration has changed
    # Restore original config and verify that MicroShift starts
    Drop In MicroShift Config    ${CUSTOM_FEATURE_GATES}    10-featuregates
    Start MicroShift
    [Teardown]    Teardown Custom Feature Gates Test

Feature Gate Lock File Persists Across Restarts With Same Config
    [Documentation]    Verify that feature gate lock file persists and validation succeeds across restarts
    ...    when the same feature gate configuration is maintained.
    [Setup]    Setup Custom Feature Gates Test
    Wait Until Keyword Succeeds    2 min    5 sec
    ...    Feature Gate Lock File Should Exist
    Restart MicroShift
    Feature Gate Lock File Should Exist
    [Teardown]    Teardown Custom Feature Gates Test


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Save Journal Cursor
    [Documentation]
    ...    Save the journal cursor then restart MicroShift so we capture the
    ...    shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE

Setup Custom Feature Gates Test
    [Documentation]    Drop in custom feature gates config and restart MicroShift
    Stop MicroShift
    Drop In MicroShift Config    ${CUSTOM_FEATURE_GATES}    10-featuregates
    Save Journal Cursor
    Start MicroShift
    Feature Gate Lock File Should Exist

Teardown Custom Feature Gates Test
    [Documentation]    Remove custom feature gates config and restart MicroShift
    Stop MicroShift
    Remove Drop In MicroShift Config    10-featuregates
    Remove Feature Gate Lock File If Exists
    Start MicroShift

Remove Feature Gate Lock File If Exists
    [Documentation]    Remove the feature gate lock file if it exists, for test cleanup
    Command Should Work    rm -f ${FEATURE_GATE_LOCK_FILE}

Feature Gate Lock File Should Exist
    [Documentation]    Verify that the feature gate lock file exists
    Command Should Work    test -f ${FEATURE_GATE_LOCK_FILE}

Feature Gate Lock File Should Contain Feature Gates
    [Documentation]    Verify that feature gate lock file contains the expected feature gate configuration
    [Arguments]    ${feature_set}    ${feature_name}
    ${contents}=    Command Should Work    cat ${FEATURE_GATE_LOCK_FILE}
    Should Contain    ${contents}    ${feature_set}
    Should Contain    ${contents}    ${feature_name}

MicroShift Should Fail To Start
    [Documentation]    Verify that MicroShift fails to start and returns a non-zero exit code.
    ...    This keyword is unique and differs from a composite keyword like
    ...    Run Keyword And Expect Error    1 != 0    Start MicroShift
    ...    because there is no need to poll the service for an "active" state, which Start MicroShift does.
    ${stdout}    ${stderr}    ${rc}=    Execute Command    sudo systemctl start microshift.service
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    1    ${rc}
