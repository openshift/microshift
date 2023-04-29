*** Settings ***
Library    OperatingSystem
Library    String
Resource   ../resources/microshift-host.resource

*** Keywords ***
Get Kubeconfig
    [Documentation]    Get the kubeconfig file from the host argument and return contents
    [Arguments]    ${host}
    Login MicroShift Host
    ${kubeconfig}    ${rc}=    Execute Command
    ...    cat /var/lib/microshift/resources/kubeadmin/${host}/kubeconfig
    ...    sudo=True    return_rc=True
    Logout MicroShift Host
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty            ${kubeconfig}
    [Return]    ${kubeconfig}

Setup Kubeconfig
    [Documentation]    Get the kubeconfig file from the configured $USHIFT_HOST, create a temporary file
    ...                and export it as $KUBECONFIG variable.
    ${kubeconfig}=    Get Kubeconfig    ${USHIFT_HOST}
    ${rand}=    Generate Random String
    ${path}=    Join Path    /tmp    ${rand}
    Create File    ${path}    ${kubeconfig}
    Set Suite Variable    \${KUBECONFIG}    ${path}

Remove Kubeconfig
    [Documentation]    Remove kubeconfig ${KUBECONFIG} file.
    Remove File    ${KUBECONFIG}

Run With Kubeconfig
    [Documentation]    Run a command using KUBECONFIG from the test suite.
    [Arguments]    ${cmd}    ${allow_fail}=False
    ${result}=    Run Process    ${cmd}    env:KUBECONFIG=${KUBECONFIG}    stderr=STDOUT    shell=True
    Log    ${result.stdout}
    IF    ${allow_fail} == False
        Should Be Equal As Integers    ${result.rc}    0
    END
    RETURN  ${result.stdout}

Create Namespace
    [Documentation]    Creates a namespace with the given name.
    [Arguments]    ${ns}
    Run With Kubeconfig    oc create namespace ${ns}

Remove Namespace
    [Documentation]    Removes the given namespace.
    [Arguments]    ${ns}
    Run With Kubeconfig    oc delete namespace ${ns}
