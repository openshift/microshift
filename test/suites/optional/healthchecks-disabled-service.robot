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

    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    find /etc/greenboot/check/required.d/ -iname '*_microshift_running_check_*'
    ...    sudo=True
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Should Be Equal As Integers    0    ${rc}

    @{checks}=    Split String    ${stdout}
    FOR    ${check}    IN    @{checks}
        ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
        ...    ${check}
        ...    sudo=True
        ...    return_rc=True
        ...    return_stderr=True
        ...    return_stdout=True
        Should Be Equal As Integers    0    ${rc}
        Should Contain    ${stderr}    microshift.service is not enabled
    END
    [Teardown]    Enable MicroShift
