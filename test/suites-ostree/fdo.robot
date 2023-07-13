*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}

${FAILING_REF}      ${EMPTY}
${TARGET_REF}       ${EMPTY}


*** Test Cases ***
FIDO Onboarding Device
    [Documentation]    No-MicroShift-system is staged with unhealthy deployment
    ...    which rolls back and leaves stale data and staged again with healthy deployment.
    ...
    ...    It is expected that final deployment will gracefully handle existence of:
    ...    MicroShift data, unhealthy stored in health file,
    ...    and a deployment gap (no-microshift rollback).

    Mask Grub Boot Success Timer
    System Should Not Feature MicroShift
    TestAgent.Add Action For Next Deployment    every    fail_greenboot
    Deploy Commit Expecting A Rollback    ${FAILING_REF}
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Should Not Be Empty    ${FAILING_REF}    FAILING_REF variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

System Should Not Feature MicroShift
    [Documentation]    Check if system contains MicroShift binary, service, or data
    SSHLibrary.Directory Should Not Exist    ${BACKUP_STORAGE}
    SSHLibrary.Directory Should Not Exist    ${DATA_DIR}
    SSHLibrary.File Should Not Exist    /usr/bin/microshift
