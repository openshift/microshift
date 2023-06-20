*** Settings ***
Documentation       Keywords common to many test suites

Library             OperatingSystem
Library             String
Resource            ../resources/kubeconfig.resource


*** Keywords ***
Setup Suite
    [Documentation]    Setup a namespace-less test suite by configuring Kubeconfig
    Check Required Env Variables
    Setup Kubeconfig

Setup Suite With Namespace
    [Documentation]    Setup a test suite by creating a unique namespace and configuring Kubeconfig
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    ${ns}=    Catenate    SEPARATOR=-    "test"    ${rand}
    Set Suite Variable    \${NAMESPACE}    ${ns}
    Create Namespace    ${ns}

Teardown Suite
    [Documentation]    Teardown the namespace-less test suite by removing the Kubeconfig.
    Remove Kubeconfig
    Login MicroShift Host

Teardown Suite With Namespace
    [Documentation]    Teardown the namespaced test suite by removing the namespace and removing the Kubeconfig.
    Remove Namespace    ${NAMESPACE}
    Remove Kubeconfig

Check Required Env Variables
    [Documentation]    Fail if any of the required environment variables is missing.
    IF    "${USHIFT_HOST}"=="${EMPTY}"
        Fatal Error    USHIFT_HOST variable is required
    END
    IF    "${USHIFT_USER}"=="${EMPTY}"
        Fatal Error    USHIFT_USER variable is required
    END

Upload String To File    # robocop: disable=too-many-calls-in-keyword
    [Documentation]    Write the string to a remote file
    [Arguments]    ${content}    ${remote_filename}
    ${rand}=    Generate Random String
    ${local_tmp}=    Join Path    /tmp    ${rand}
    Create File    ${local_tmp}    ${content}
    ${rand}=    Generate Random String
    ${remote_tmp}=    Join Path    /tmp    ${rand}
    Put File    ${local_tmp}    ${remote_tmp}    mode=0644
    Remove File    ${local_tmp}
    ${stdout}    ${rc}=    Execute Command
    ...    mv ${remote_tmp} ${remote_filename}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
    ${stdout}    ${rc}=    Execute Command
    ...    chown root:root ${remote_filename}
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
