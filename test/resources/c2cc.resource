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
    ${conn}=    SSHLibrary.Get Connection
    Set To Dictionary    ${C2CC_KUBECONFIGS}    ${alias}    ${KUBECONFIG}
    Set To Dictionary    ${C2CC_SSH_IDS}    ${alias}    ${conn.index}

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
    ...    and switch back to the local cluster connection so that
    ...    Logout MicroShift Host can close it.
    FOR    ${alias}    IN    @{C2CC_REMOTE_ALIASES}
        SSHLibrary.Switch Connection    ${alias}
        SSHLibrary.Close Connection
    END
    @{C2CC_REMOTE_ALIASES}=    Create List
    Set Suite Variable    @{C2CC_REMOTE_ALIASES}
    ${local_conn}=    Get From Dictionary    ${C2CC_SSH_IDS}    cluster-a
    SSHLibrary.Switch Connection    ${local_conn}

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

Disruptive Command On Cluster
    [Documentation]    Run a shell command on the specified cluster via SSH.
    ...    Does NOT assert rc=0 — intended for fault-injection commands
    ...    where the command may race with the controller's reconcile loop.
    [Arguments]    ${alias}    ${command}    ${sudo_mode}=True
    ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
    SSHLibrary.Switch Connection    ${conn_id}
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ${command}
    ...    sudo=${sudo_mode}
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Log    Disruptive command rc=${rc} stdout=${stdout} stderr=${stderr}
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

Oc Exec On Cluster
    [Documentation]    Run 'oc exec' on a pod in the specified cluster.
    [Arguments]    ${alias}    ${pod}    ${cmd}    ${ns}=${EMPTY}
    ${ns_arg}=    Set Variable If    '${ns}'    -n ${ns}    ${EMPTY}
    ${stdout}=    Oc On Cluster    ${alias}    oc exec ${ns_arg} ${pod} -- ${cmd}
    RETURN    ${stdout}

Oc Apply On Cluster
    [Documentation]    Run 'oc apply -f' on the specified cluster.
    [Arguments]    ${alias}    ${file}    ${ns}=${EMPTY}
    ${ns_arg}=    Set Variable If    '${ns}'    -n ${ns}    ${EMPTY}
    ${stdout}=    Oc On Cluster    ${alias}    oc apply ${ns_arg} -f ${file}
    RETURN    ${stdout}

Verify Routes In Table 200
    [Documentation]    Check that routes for the given CIDRs exist in table 200.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip route show table 200
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify IP Rules For Table 200
    [Documentation]    Check that IP rules at priority 100 exist for the given CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip rule show
    Should Contain    ${stdout}    to ${remote_pod_cidr} lookup 200
    Should Contain    ${stdout}    to ${remote_svc_cidr} lookup 200

Verify Routes In Table 201
    [Documentation]    Check that service routes exist in table 201 for the local service CIDR.
    [Arguments]    ${alias}    ${local_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip route show table 201
    Should Contain    ${stdout}    ${local_svc_cidr}

Verify Service IP Rules
    [Documentation]    Check that IP rules at priority 99 exist for cross-cluster service routing.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip rule show
    Should Contain    ${stdout}    from ${remote_pod_cidr} to ${local_svc_cidr} lookup 201
    Should Contain    ${stdout}    from ${remote_svc_cidr} to ${local_svc_cidr} lookup 201

Verify NFTables Bypass Rules
    [Documentation]    Check that nftables masquerade bypass rules exist for remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}
    ...    nft list chain inet ovn-kubernetes ovn-kube-pod-subnet-masq
    Should Contain    ${stdout}    c2cc-no-masq:${remote_pod_cidr}
    Should Contain    ${stdout}    c2cc-no-masq:${remote_svc_cidr}

Verify OVN Static Routes
    [Documentation]    Check that OVN NB static routes tagged with microshift-c2cc exist for remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${pod}=    Oc On Cluster    ${alias}
    ...    oc get pod -n openshift-ovn-kubernetes -l app=ovnkube-master -o jsonpath='{.items[0].metadata.name}'
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec -n openshift-ovn-kubernetes ${pod} -- ovn-nbctl find Logical_Router_Static_Route external_ids:k8s.ovn.org/owner-controller=microshift-c2cc
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify Node SNAT Annotation
    [Documentation]    Check that the node SNAT-exclude annotation contains the remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get node -o json | python3 -c "import json,sys; print(json.load(sys.stdin)['items'][0]['metadata']['annotations'].get('k8s.ovn.org/node-ingress-snat-exclude-subnets',''))"
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify C2CC Network Policy
    [Documentation]    Check that the C2CC network policy exists in the default namespace.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get networkpolicy c2cc-allow-remote-pods -n default -o jsonpath='{.metadata.labels.app\\.kubernetes\\.io/managed-by}'
    Should Be Equal As Strings    ${stdout}    microshift-c2cc
