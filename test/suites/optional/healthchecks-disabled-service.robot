*** Settings ***
Documentation       Test if healthcheck exits quickly when MicroShift service is disabled

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Library             DateTime

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Test Cases ***
Healthchecks Should Exit Fast And Successful When MicroShift Is Disabled
    [Documentation]    When microshift.service is disabled, the healtchecks should
    ...    exit quickly and without an error. They should not attempt to query API Server.
    [Tags]    ushift-5381
    [Setup]    Disable MicroShift
    ${before}=    Get Current Date
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    /etc/greenboot/check/required.d/41_microshift_running_check_multus.sh
    ...    sudo=True
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Should Be Equal As Integers    0    ${rc}
    ${after}=    Get Current Date
    Should Contain    ${stderr}    microshift.service is not enabled
    ${diff}=    Subtract Date From Date    ${after}    ${before}
    # Verify that the command returned very quickly - it didn't waste time trying to access API Server.
    Should Be True    ${diff} < 5    Multus healthcheck script took too long to finish
    [Teardown]    Enable MicroShift
