*** Settings ***
Documentation       Keywords for multi-cluster C2CC test scenarios.
...                 Provides SSH and oc access to N clusters via alias-based registration.

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Library             SSHLibrary
Resource            common.resource
Resource            microshift-host.resource


*** Variables ***
${CLUSTER_A_POD_CIDR}       ${EMPTY}
${CLUSTER_A_SVC_CIDR}       ${EMPTY}
${CLUSTER_A_DOMAIN}         ${EMPTY}
${CLUSTER_B_POD_CIDR}       ${EMPTY}
${CLUSTER_B_SVC_CIDR}       ${EMPTY}
${CLUSTER_B_DOMAIN}         ${EMPTY}
${HOST2_IP}                 ${EMPTY}
${HOST2_SSH_PORT}           ${EMPTY}
${HOST2_API_PORT}           ${EMPTY}
${KUBECONFIG_B}             ${EMPTY}
&{C2CC_KUBECONFIGS}
&{C2CC_SSH_IDS}
@{C2CC_REMOTE_ALIASES}


*** Keywords ***
Register Local Cluster
    [Documentation]    Register the primary cluster (host1) for use with generic cluster keywords.
    ...    Must be called after Login MicroShift Host and Setup Kubeconfig.
    [Arguments]    ${alias}
    Set To Dictionary    ${C2CC_KUBECONFIGS}    ${alias}    ${KUBECONFIG}
    Set To Dictionary    ${C2CC_SSH_IDS}    ${alias}    1

Register Remote Cluster
    [Documentation]    Open an SSH connection to a remote cluster and register it by alias.
    [Arguments]    ${alias}    ${host}    ${ssh_port}    ${kubeconfig}
    IF    '${ssh_port}'
        SSHLibrary.Open Connection    ${host}    alias=${alias}    port=${ssh_port}
    ELSE
        SSHLibrary.Open Connection    ${host}    alias=${alias}
    END
    IF    '${SSH_PRIV_KEY}'
        SSHLibrary.Login With Public Key    ${USHIFT_USER}    ${SSH_PRIV_KEY}    keep_alive_interval=30
    ELSE
        SSHLibrary.Login    ${USHIFT_USER}    allow_agent=True    keep_alive_interval=30
    END
    Set To Dictionary    ${C2CC_KUBECONFIGS}    ${alias}    ${kubeconfig}
    Set To Dictionary    ${C2CC_SSH_IDS}    ${alias}    ${alias}
    Append To List    ${C2CC_REMOTE_ALIASES}    ${alias}

Teardown All Remote Clusters
    [Documentation]    Close SSH connections for all registered remote clusters
    ...    and switch back to the primary (host1) connection.
    FOR    ${alias}    IN    @{C2CC_REMOTE_ALIASES}
        SSHLibrary.Switch Connection    ${alias}
        SSHLibrary.Close Connection
    END
    SSHLibrary.Switch Connection    1

Command On Cluster
    [Documentation]    Run a shell command on the specified cluster via SSH.
    [Arguments]    ${alias}    ${command}    ${sudo_mode}=True
    ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
    SSHLibrary.Switch Connection    ${conn_id}
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ${command}
    ...    sudo=${sudo_mode}
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Should Be Equal As Integers    ${rc}    0
    RETURN    ${stdout}

Oc On Cluster
    [Documentation]    Run an oc command against the specified cluster using its kubeconfig.
    [Arguments]    ${alias}    ${cmd}
    ${kc}=    Get From Dictionary    ${C2CC_KUBECONFIGS}    ${alias}
    ${stdout_file}=    Create Random Temp File
    ${result}=    Process.Run Process    ${cmd}    env:KUBECONFIG=${kc}
    ...    stderr=STDOUT    shell=True
    ...    stdout=${stdout_file}
    ...    timeout=300s
    Log    ${result.stdout}
    Log    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0
    ${stdout}=    OperatingSystem.Get File    ${stdout_file}
    ${stdout}=    Strip String    ${stdout}
    RETURN    ${stdout}
