*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${OLDER_MICROSHIFT_REF}     ${EMPTY}
${BOOTC_REGISTRY}           ${EMPTY}


*** Test Cases ***
Downgrade Is Blocked
    [Documentation]    Verifies that staging a new deployment featuring
    ...    MicroShift "older" than existing data is blocked
    ...    and results in system rolling back.

    # Capture the deployment ID (not deployment_boot) because the downgrade is
    # blocked: the older MicroShift never rewrites the data version file, so the
    # data keeps belonging to this deployment and every backup taken of it is
    # keyed to this deployment ID. Only the boot ID suffix changes across the
    # rollback reboots, and MicroShift prunes to a single backup per deployment,
    # so we must assert on the deployment rather than on a specific boot.
    ${initial_deploy_id}=    Get Booted Deployment Id
    Remove Existing Backup For Current Deployment
    Backup For Deployment Should Not Exist    ${initial_deploy_id}

    Deploy Commit Expecting A Rollback
    ...    ${OLDER_MICROSHIFT_REF}
    ...    False
    ...    ${BOOTC_REGISTRY}

    Wait Until Greenboot Health Check Exited
    Backup For Deployment Should Exist    ${initial_deploy_id}
    Journal Should Have Information About Failed Version Comparison


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${OLDER_MICROSHIFT_REF}    OLDER_MICROSHIFT_REF variable is required
    Login MicroShift Host
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
