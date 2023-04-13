*** Settings ***
Library     SSHLibrary
Library     String


*** Keywords ***
Get Kubeconfig
    [Documentation]    X
    Open Connection    ${USHIFT_IP}
    Login    ${USHIFT_USER}    allow_agent=True
    ${konfig}    ${rc}=    Execute Command
    ...    cat /var/lib/microshift/resources/kubeadmin/${USHIFT_IP}/kubeconfig
    ...    sudo=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Not Be Empty    ${konfig}
    ${rand}=    Generate Random String
    ${path}=    Join Path    /tmp    ${rand}
    Create File    ${path}    ${konfig}
    Close Connection
    Set Suite Variable    \${KUBECONFIG}    ${path}

Remove Kubeconfig
    Remove File    ${KUBECONFIG}

Run With Kubeconfig
    [Arguments]    ${cmd}    ${allow_fail}=False
    ${result}=    Run Process    ${cmd}    env:KUBECONFIG=${KUBECONFIG}    stderr=STDOUT    shell=True
    Log    ${result.stdout}
    IF    ${allow_fail} == False
        Should Be Equal As Integers    ${result.rc}    0
    END
