*** Settings ***
Documentation       Keywords for accessing a kubeconfig file for the MicroShift host

Library             OperatingSystem
Library             String
Library             DataFormats.py
Library             libipv6.py
Resource            ../resources/microshift-host.resource
Resource            ../resources/common.resource


*** Variables ***
${API_PORT}     ${EMPTY}    # overridden by scenario.sh in CI


*** Keywords ***
Get Kubeconfig
    [Documentation]    Get the kubeconfig file from the host argument and return contents
    [Arguments]    ${host}
    ${kubeconfig}    ${rc}=    Execute Command
    ...    cat /var/lib/microshift/resources/kubeadmin/${host}/kubeconfig
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${kubeconfig}
    RETURN    ${kubeconfig}

Setup Kubeconfig
    [Documentation]    Get the kubeconfig file from the configured $USHIFT_HOST,
    ...    update the server URL based on the API_PORT, create a temporary file,
    ...    and export it as $KUBECONFIG variable.
    ${kubeconfig}=    Get Kubeconfig    ${USHIFT_HOST}
    # If we have been given an explicit API port to use, make sure that
    # port appears in the kubeconfig file.
    IF    "${API_PORT}"!="${EMPTY}"
        ${ip}=    Add Brackets If Ipv6    ${USHIFT_HOST}
        ${kubeconfig}=    Update Kubeconfig Server Url
        ...    ${kubeconfig}
        ...    https://${ip}:${API_PORT}
    END
    ${path}=    Create Random Temp File    ${kubeconfig}
    Set Suite Variable    \${KUBECONFIG}    ${path}

Setup Custom Kubeconfig
    [Documentation]    Get the kubeconfig file from the path,
    ...    and export it as $KUBECONFIG variable.
    [Arguments]    ${custom_path}
    ${kubeconfig}=    Get Kubeconfig    ${custom_path}
    ${path}=    Create Random Temp File    ${kubeconfig}
    Set Suite Variable    \${KUBECONFIG}    ${path}

Remove Kubeconfig
    [Documentation]    Remove kubeconfig ${KUBECONFIG} file.
    Remove File    ${KUBECONFIG}

Run With Kubeconfig
    [Documentation]    Run a command using KUBECONFIG from the test suite.
    [Arguments]    ${cmd}    ${allow_fail}=False    ${return_rc}=False
    ${stdout_file}=    Create Random Temp File
    ${result}=    Run Process    ${cmd}    env:KUBECONFIG=${KUBECONFIG}
    ...    stderr=STDOUT    shell=True
    ...    stdout=${stdout_file}

    Log    ${result.stdout}
    Log    ${result.stderr}
    IF    ${allow_fail} == False
        Should Be Equal As Integers    ${result.rc}    0
    END
    ${stdout}=    OperatingSystem.Get File    ${stdout_file}
    ${stdout}=    Strip String    ${stdout}    # strip \n character at the end
    IF    ${return_rc}    RETURN    ${stdout}    ${result.rc}
    RETURN    ${stdout}

Create Namespace
    [Documentation]    Creates a namespace with the given name.
    [Arguments]    ${ns}
    Run With Kubeconfig    oc create namespace ${ns}

Create Random Namespace
    [Documentation]    Creates a namespace with a random name and return the name.
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lowercase    ${rand}
    Run With Kubeconfig    oc create namespace test-${rand}
    RETURN    test-${rand}

Remove Namespace
    [Documentation]    Removes the given namespace.
    [Arguments]    ${ns}
    Run With Kubeconfig    oc delete namespace ${ns}
