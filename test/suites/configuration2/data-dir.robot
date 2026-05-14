*** Settings ***
Documentation       Tests that MicroShift only uses /var/lib/microshift for data storage

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart


*** Variables ***
${HOME_DATA_DIR}        /root/.microshift/data
${GLOBAL_RESOURCES}     /var/lib/microshift/resources


*** Test Cases ***
MicroShift Does Not Use Home Directory For Data
    [Documentation]    MicroShift should only use /var/lib/microshift/ for data.
    ...    The legacy search path ~/.microshift/data/ must not be populated.
    Command Should Work    mkdir -p ${HOME_DATA_DIR}
    Command Should Work    find ${HOME_DATA_DIR} -mindepth 1 -delete

    Command Should Work    find ${GLOBAL_RESOURCES} -mindepth 1 -delete
    Directory Should Be Empty    ${GLOBAL_RESOURCES}

    Restart MicroShift

    Directory Should Be Empty    ${HOME_DATA_DIR}
    Wait Until Keyword Succeeds    40x    5s
    ...    Directory Should Not Be Empty    ${GLOBAL_RESOURCES}

    [Teardown]    Run Keywords
    ...    Command Should Work    rm -rf /root/.microshift
    ...    AND
    ...    Restart MicroShift


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

Directory Should Be Empty
    [Documentation]    Verify that a directory contains no files or subdirectories
    [Arguments]    ${path}
    ${stdout}=    Command Should Work    find ${path} -mindepth 1 -maxdepth 1 | wc -l
    Should Be Equal As Strings    ${stdout.strip()}    0    Directory ${path} is not empty

Directory Should Not Be Empty
    [Documentation]    Verify that a directory contains at least one file or subdirectory
    [Arguments]    ${path}
    ${stdout}=    Command Should Work    find ${path} -mindepth 1 -maxdepth 1 | wc -l
    Should Not Be Equal As Strings    ${stdout.strip()}    0    Directory ${path} is empty
