*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/selinux.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${TARGET_REF}           ${EMPTY}

${LVMS_TOPOLVM_DIFF}    ./assets/topolvm-to-lvms-diff.yaml


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade to given reference
    ...    and verifies if it was successful, with SELinux validation

    Wait Until Greenboot Health Check Exited

    ${future_backup}=    Get Future Backup Name For Current Boot
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}
    Backup Should Exist    ${future_backup}

    Validate SELinux With Backup    ${future_backup}

    # Skip topolvm deletion check if upgrade is to "crel" because (at the time of writing)
    # it doesn't yet contain the topolvm -> lvms changes.
    IF    'microshift-crel' not in '${TARGET_REF}'
        # verifies that old TopoLVM resources are cleaned up after migration
        Oc Wait    -f ${LVMS_TOPOLVM_DIFF} --for=Delete --timeout=${DEFAULT_WAIT_TIMEOUT}
    END


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
