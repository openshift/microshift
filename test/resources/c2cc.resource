*** Settings ***
Documentation       Keywords for multi-cluster C2CC test scenarios.
...                 Provides SSH and oc access to N clusters via alias-based registration.

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Library             SSHLibrary
Resource            common.resource
Resource            libvirt.resource
Resource            microshift-host.resource


*** Variables ***
${CLUSTER_A_POD_CIDR}           ${EMPTY}
${CLUSTER_A_SVC_CIDR}           ${EMPTY}
${CLUSTER_A_DOMAIN}             ${EMPTY}
${KUBECONFIG_A}                 ${EMPTY}
${HOST1_IPV6}                   ${EMPTY}
${CLUSTER_B_POD_CIDR}           ${EMPTY}
${CLUSTER_B_SVC_CIDR}           ${EMPTY}
${CLUSTER_B_DOMAIN}             ${EMPTY}
${HOST2_IP}                     ${EMPTY}
${HOST2_IPV6}                   ${EMPTY}
${HOST2_SSH_PORT}               ${EMPTY}
${HOST2_API_PORT}               ${EMPTY}
${KUBECONFIG_B}                 ${EMPTY}
${CLUSTER_C_POD_CIDR}           ${EMPTY}
${CLUSTER_C_SVC_CIDR}           ${EMPTY}
${CLUSTER_C_DOMAIN}             ${EMPTY}
${HOST3_IP}                     ${EMPTY}
${HOST3_IPV6}                   ${EMPTY}
${HOST3_SSH_PORT}               ${EMPTY}
${HOST3_API_PORT}               ${EMPTY}
${KUBECONFIG_C}                 ${EMPTY}
&{C2CC_KUBECONFIGS}             &{EMPTY}
&{C2CC_SSH_IDS}                 &{EMPTY}
@{C2CC_REMOTE_ALIASES}          @{EMPTY}
${CLUSTER_A_POD_CIDR_DUAL}      ${EMPTY}
${CLUSTER_A_SVC_CIDR_DUAL}      ${EMPTY}
${CLUSTER_B_POD_CIDR_DUAL}      ${EMPTY}
${CLUSTER_B_SVC_CIDR_DUAL}      ${EMPTY}
${CLUSTER_C_POD_CIDR_DUAL}      ${EMPTY}
${CLUSTER_C_SVC_CIDR_DUAL}      ${EMPTY}
${IP_FAMILY}                    ${EMPTY}
${IP_CMD}                       ${{'ip -6' if '${IP_FAMILY}' == 'ipv6' else 'ip -4'}}
&{NAMESPACES}                   cluster-a=${EMPTY}    cluster-b=${EMPTY}    cluster-c=${EMPTY}
&{DOMAIN_MAP}
...                             cluster-a=${CLUSTER_A_DOMAIN}
...                             cluster-b=${CLUSTER_B_DOMAIN}
...                             cluster-c=${CLUSTER_C_DOMAIN}
&{POD_CIDR_MAP}
...                             cluster-a=${CLUSTER_A_POD_CIDR}
...                             cluster-b=${CLUSTER_B_POD_CIDR}
...                             cluster-c=${CLUSTER_C_POD_CIDR}
&{SVC_CIDR_MAP}
...                             cluster-a=${CLUSTER_A_SVC_CIDR}
...                             cluster-b=${CLUSTER_B_SVC_CIDR}
...                             cluster-c=${CLUSTER_C_SVC_CIDR}
&{POD_CIDR_MAP_DUAL}
...                             cluster-a=${CLUSTER_A_POD_CIDR_DUAL}
...                             cluster-b=${CLUSTER_B_POD_CIDR_DUAL}
...                             cluster-c=${CLUSTER_C_POD_CIDR_DUAL}
&{SVC_CIDR_MAP_DUAL}
...                             cluster-a=${CLUSTER_A_SVC_CIDR_DUAL}
...                             cluster-b=${CLUSTER_B_SVC_CIDR_DUAL}
...                             cluster-c=${CLUSTER_C_SVC_CIDR_DUAL}
@{ALL_CLUSTERS}                 cluster-a    cluster-b    cluster-c


*** Keywords ***
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

Register All C2CC Clusters
    [Documentation]    Register all three C2CC clusters as remote clusters.
    Register Remote Cluster    cluster-a    ${USHIFT_HOST}    ${SSH_PORT}    ${KUBECONFIG_A}
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Register Remote Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}

Teardown All Remote Clusters
    [Documentation]    Close SSH connections for all registered remote clusters.
    FOR    ${alias}    IN    @{C2CC_REMOTE_ALIASES}
        SSHLibrary.Switch Connection    ${alias}
        SSHLibrary.Close Connection
    END
    VAR    @{C2CC_REMOTE_ALIASES}=    @{EMPTY}    scope=SUITE

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
    [Arguments]    ${alias}    ${cmd}    ${allow_fail}=${FALSE}
    ${kc}=    Get From Dictionary    ${C2CC_KUBECONFIGS}    ${alias}
    ${stdout_file}=    Create Random Temp File
    ${result}=    Process.Run Process    ${cmd}    env:KUBECONFIG=${kc}
    ...    stderr=STDOUT    shell=True
    ...    stdout=${stdout_file}
    ...    timeout=300s
    Log    ${result.stdout}
    Log    ${result.stderr}
    IF    not ${allow_fail}    Should Be Equal As Integers    ${result.rc}    0
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

Verify Cluster Is Healthy
    [Documentation]    Verify MicroShift and all core workloads are ready.
    [Arguments]    ${alias}
    Command On Cluster    ${alias}    microshift healthcheck --timeout=300s

Verify Routes In Table 200
    [Documentation]    Check that routes for the given CIDRs exist in table 200.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${ip_cmd}=${IP_CMD}
    ${stdout}=    Command On Cluster    ${alias}    ${ip_cmd} route show table 200
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify IP Rules For Table 200
    [Documentation]    Check that IP rules at priority 100 exist for the given CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${ip_cmd}=${IP_CMD}
    ${stdout}=    Command On Cluster    ${alias}    ${ip_cmd} rule show
    Should Contain    ${stdout}    to ${remote_pod_cidr} lookup 200
    Should Contain    ${stdout}    to ${remote_svc_cidr} lookup 200

Verify Service IP Rules
    [Documentation]    Check that IP rules at priority 99 exist for cross-cluster service routing.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}    ${ip_cmd}=${IP_CMD}
    ${stdout}=    Command On Cluster    ${alias}    ${ip_cmd} rule show
    Should Contain    ${stdout}    from ${remote_pod_cidr} to ${local_svc_cidr} lookup 201
    Should Contain    ${stdout}    from ${remote_svc_cidr} to ${local_svc_cidr} lookup 201

Verify All IP Rules
    [Documentation]    Check all IP rules (table 200 and 201) in a single SSH call.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}    ${ip_cmd}=${IP_CMD}
    ${stdout}=    Command On Cluster    ${alias}    ${ip_cmd} rule show
    Should Contain    ${stdout}    to ${remote_pod_cidr} lookup 200
    Should Contain    ${stdout}    to ${remote_svc_cidr} lookup 200
    Should Contain    ${stdout}    from ${remote_pod_cidr} to ${local_svc_cidr} lookup 201
    Should Contain    ${stdout}    from ${remote_svc_cidr} to ${local_svc_cidr} lookup 201

Verify Routes In Table 201
    [Documentation]    Check that service routes exist in table 201 for the local service CIDR.
    [Arguments]    ${alias}    ${local_svc_cidr}    ${ip_cmd}=${IP_CMD}
    ${stdout}=    Command On Cluster    ${alias}    ${ip_cmd} route show table 201
    Should Contain    ${stdout}    ${local_svc_cidr}

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
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc exec -n openshift-ovn-kubernetes ${pod} -- ovn-nbctl find Logical_Router_Static_Route external_ids:k8s.ovn.org/owner-controller=microshift-c2cc
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify Node SNAT Annotation
    [Documentation]    Check that the node SNAT-exclude annotation contains the remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get node -o jsonpath='{.items[0].metadata.annotations.k8s\\.ovn\\.org/node-ingress-snat-exclude-subnets}'
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify C2CC Tracking Annotation
    [Documentation]    Check that the C2CC tracking annotation exists and contains the remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get node -o jsonpath='{.items[0].metadata.annotations.microshift\\.io/c2cc-snat-subnets}'
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Get Node SNAT Annotation
    [Documentation]    Return the raw value of the SNAT-exclude annotation.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get node -o jsonpath='{.items[0].metadata.annotations.k8s\\.ovn\\.org/node-ingress-snat-exclude-subnets}'
    RETURN    ${stdout}

Inject Foreign Subnet Into SNAT Annotation
    [Documentation]    Add a foreign (non-C2CC) subnet to the SNAT-exclude annotation.
    ...    Reads the current value, appends the foreign CIDR, and patches the node.
    [Arguments]    ${alias}    ${foreign_cidr}
    ${node}=    Oc On Cluster    ${alias}
    ...    oc get nodes -o jsonpath='{.items[0].metadata.name}'
    ${current}=    Get Node SNAT Annotation    ${alias}
    ${new_value}=    Command On Cluster
    ...    ${alias}
    ...    echo '${current}' | python3 -c "import sys,json; d=json.load(sys.stdin); d.append('${foreign_cidr}'); print(json.dumps(sorted(d)))"
    ...    sudo_mode=False
    Oc On Cluster    ${alias}
    ...    oc annotate node ${node} k8s.ovn.org/node-ingress-snat-exclude-subnets='${new_value}' --overwrite

Remove C2CC CIDRs From SNAT Annotation Keeping Foreign
    [Documentation]    Remove C2CC CIDRs from the SNAT annotation, keeping only the foreign subnet.
    ...    This simulates external modification that strips C2CC entries.
    [Arguments]    ${alias}    ${foreign_cidr}
    ${node}=    Oc On Cluster    ${alias}
    ...    oc get nodes -o jsonpath='{.items[0].metadata.name}'
    Oc On Cluster    ${alias}
    ...    oc annotate node ${node} k8s.ovn.org/node-ingress-snat-exclude-subnets='["${foreign_cidr}"]' --overwrite

Remove Foreign Subnet From SNAT Annotation
    [Documentation]    Remove a foreign subnet from the SNAT annotation to restore clean state.
    ...    Triggers a reconcile by briefly corrupting the annotation.
    [Arguments]    ${alias}
    Corrupt Node SNAT Annotation On Cluster    ${alias}

Create Unique Namespace On Cluster
    [Documentation]    Creates a namespace with a unique suffix on the given cluster.
    ...    Returns the namespace name.
    [Arguments]    ${alias}
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    ${ns}=    Catenate    SEPARATOR=-    test    ${rand}
    Oc On Cluster    ${alias}    oc create namespace ${ns}
    RETURN    ${ns}

Verify Corefile Contains C2CC Server Block
    [Documentation]    Check that the CoreDNS Corefile configmap contains a server block for the given domain.
    [Arguments]    ${alias}    ${domain}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get configmap dns-default -n openshift-dns -o jsonpath='{.data.Corefile}'
    Should Contain    ${stdout}    ${domain}:5353
    Should Contain    ${stdout}    rewrite stop name suffix .${domain} .cluster.local answer auto

Verify Corefile Does Not Contain C2CC Server Block
    [Documentation]    Check that the CoreDNS Corefile does NOT contain a server block for the given domain.
    [Arguments]    ${alias}    ${domain}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get configmap dns-default -n openshift-dns -o jsonpath='{.data.Corefile}'
    Should Not Contain    ${stdout}    ${domain}:5353

Deploy Test Workloads
    [Documentation]    Create namespace and deploy hello-microshift + curl-pod on all clusters.
    VAR    ${assets}=    ${EXECDIR}/assets/c2cc
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${ns}=    Create Unique Namespace On Cluster    ${alias}
        Set To Dictionary    ${NAMESPACES}    ${alias}    ${ns}
        Oc On Cluster    ${alias}    oc apply -n ${ns} -f ${assets}/hello-microshift.yaml
        Oc On Cluster    ${alias}    oc apply -n ${ns} -f ${assets}/curl-pod.yaml
    END
    Wait For Test Pods
    Wait For Service Endpoints

Wait For Test Pods
    [Documentation]    Wait for all test pods to be Ready on all clusters.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Oc On Cluster
        ...    ${alias}
        ...    oc wait pod/hello-microshift pod/curl-pod -n ${NAMESPACES}[${alias}] --for=condition=Ready --timeout=120s
    END

Wait For Service Endpoints
    [Documentation]    Wait for hello-microshift service to have endpoints on all clusters.
    ...    Ensures OVN-K has synced the EndpointSlice and programmed the OVN load balancer.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Wait Until Keyword Succeeds    120s    5s
        ...    Verify Service Endpoints Exist    ${alias}    ${NAMESPACES}[${alias}]
    END

Verify Service Endpoints Exist
    [Documentation]    Verify the hello-microshift endpoint has at least one address.
    [Arguments]    ${alias}    ${ns}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get endpoints hello-microshift -n ${ns} -o jsonpath='{.subsets[0].addresses[0].ip}'
    Should Not Be Empty    ${stdout}

Compute 11th IP
    [Documentation]    Return the 11th host address in a CIDR (e.g. 10.43.0.0/16 -> 10.43.0.11).
    [Arguments]    ${cidr}
    VAR    ${cmd}=    import ipaddress; n=ipaddress.ip_network('${cidr}', strict=False); print(n[11])
    ${result}=    Process.Run Process    python3    -c    ${cmd}
    Should Be Equal As Integers    ${result.rc}    0
    ${ip}=    Strip String    ${result.stdout}
    RETURN    ${ip}

Cleanup Test Workloads
    [Documentation]    Delete test namespace on all clusters. Ignores errors.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Oc On Cluster    ${alias}    oc delete namespace ${NAMESPACES}[${alias}] --timeout=60s
    END

Get Hello Pod IP
    [Documentation]    Get the pod IP of hello-microshift on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIP}'
    RETURN    ${ip}

Get Hello Service IP
    [Documentation]    Get the ClusterIP of the hello-microshift service on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get svc hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.spec.clusterIP}'
    RETURN    ${ip}

Get Curl Pod IP
    [Documentation]    Get the pod IP of curl-pod on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod curl-pod -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIP}'
    RETURN    ${ip}

Get Hello Pod Dual IP
    [Documentation]    Get the secondary (dual-stack) pod IP of hello-microshift.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIPs[1].ip}'
    RETURN    ${ip}

Get Hello Service Dual IP
    [Documentation]    Get the secondary (dual-stack) ClusterIP of the hello-microshift service.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get svc hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.spec.clusterIPs[1]}'
    RETURN    ${ip}

Get Curl Pod Dual IP
    [Documentation]    Get the secondary (dual-stack) pod IP of curl-pod.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod curl-pod -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIPs[1].ip}'
    RETURN    ${ip}

Get Curl Pod IP Matching Family
    [Documentation]    Get the curl-pod IP that matches the address family of the given destination IP.
    ...    In dual-stack, OVN-K always assigns IPv4 first in podIPs regardless of clusterNetwork
    ...    order, so status.podIP is always IPv4 and podIPs[1] is IPv6. Services, however,
    ...    follow the config order — spec.clusterIP is IPv6 in IPv6-primary clusters.
    ...    In single-stack (IPv4-only or IPv6-only), podIPs has length 1 and podIPs[1] does
    ...    not exist, so we must check whether the primary IP already matches before reaching
    ...    for the secondary.
    [Arguments]    ${alias}    ${dest_ip}
    ${primary}=    Get Curl Pod IP    ${alias}
    ${dest_is_ipv6}=    Evaluate    ':' in '''${dest_ip}'''
    ${primary_is_ipv6}=    Evaluate    ':' in '''${primary}'''
    IF    ${dest_is_ipv6} == ${primary_is_ipv6}    RETURN    ${primary}
    ${ip}=    Get Curl Pod Dual IP    ${alias}
    RETURN    ${ip}

Curl From Cluster
    [Documentation]    Exec curl from curl-pod on the given cluster to the target IP and port.
    [Arguments]    ${alias}    ${ip}    ${port}
    ${is_ipv6}=    Evaluate    ':' in '''${ip}'''
    ${url}=    Set Variable If
    ...    ${is_ipv6}
    ...    http://[${ip}]:${port}/cgi-bin/hello
    ...    http://${ip}:${port}/cgi-bin/hello
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- curl -sS --max-time 10 ${url}
    RETURN    ${stdout}

Verify Connectivity Between Clusters
    [Documentation]    Verify pod on ${source} can reach ${endpoint_type} IP on ${destination}.
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END
    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    Hello from

Verify Source IP Preserved Between Clusters
    [Documentation]    Verify ${source} to ${destination} pod-to-${endpoint_type} traffic preserves the source pod IP (no SNAT).
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END
    ${curl_pod_ip}=    Get Curl Pod IP Matching Family    ${source}    ${ip_dest}
    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    source: ${curl_pod_ip}

Verify Dual Stack Connectivity Between Clusters
    [Documentation]    Verify pod on ${source} can reach ${endpoint_type} dual-stack IP on ${destination}.
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod Dual IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service Dual IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END
    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    Hello from

Verify Dual Stack Source IP Preserved Between Clusters
    [Documentation]    Verify dual-stack traffic preserves the source pod IP (no SNAT).
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod Dual IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service Dual IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END
    ${curl_pod_ip}=    Get Curl Pod IP Matching Family    ${source}    ${ip_dest}
    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    source: ${curl_pod_ip}

Verify DNS Resolution From Cluster
    [Documentation]    Resolve a DNS name from curl-pod on the given cluster. Retries for up to 200s.
    [Arguments]    ${alias}    ${fqdn}
    Wait Until Keyword Succeeds    40x    5s
    ...    Verify DNS Lookup    ${alias}    ${fqdn}

Verify DNS Lookup
    [Documentation]    Resolve a DNS name from curl-pod using getent hosts.
    [Arguments]    ${alias}    ${fqdn}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- getent hosts ${fqdn}
    Should Not Be Empty    ${stdout}

Verify All RemoteClusters Healthy
    [Documentation]    Wait for all RemoteCluster CRs on all clusters to report Healthy.
    FOR    ${alias}    IN    @{ALL_CLUSTERS}
        Wait Until Keyword Succeeds    3m    10s
        ...    Verify RemoteCluster State    ${alias}    Healthy
    END

Verify Curl DNS From Cluster
    [Documentation]    Curl a service by DNS name from curl-pod on the given cluster.
    [Arguments]    ${alias}    ${fqdn}    ${port}
    ${stdout}=    Wait Until Keyword Succeeds    40x    5s
    ...    Verify Curl DNS    ${alias}    ${fqdn}    ${port}
    RETURN    ${stdout}

Verify Curl DNS
    [Documentation]    Single attempt to curl a DNS name from curl-pod.
    [Arguments]    ${alias}    ${fqdn}    ${port}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- curl -sS --max-time 10 http://${fqdn}:${port}/cgi-bin/hello
    Should Contain    ${stdout}    Hello from
    RETURN    ${stdout}

Verify Remote Service Via DNS
    [Documentation]    Verify pod on ${source} can reach a service on ${destination} using the remote DNS name.
    [Arguments]    ${source}    ${destination}
    ${stdout}=    Verify Curl DNS From Cluster    ${source}
    ...    hello-microshift.${NAMESPACES}[${destination}].svc.${DOMAIN_MAP}[${destination}]    8080
    Should Contain    ${stdout}    Hello from

Verify RemoteCluster State
    [Documentation]    Check that all RemoteCluster CRs on this cluster have the expected state.
    [Arguments]    ${alias}    ${expected_state}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.state}'
    Should Not Be Empty    ${stdout}
    @{states}=    Split String    ${stdout}
    ${count}=    Get Length    ${states}
    ${expected_count}=    Evaluate    len($ALL_CLUSTERS) - 1
    Should Be Equal As Integers    ${count}    ${expected_count}
    ...    Expected ${expected_count} RemoteCluster states, got ${count}
    FOR    ${state}    IN    @{states}
        Should Be Equal As Strings    ${state}    ${expected_state}
    END

Verify RemoteCluster State By Name
    [Documentation]    Check that a specific RemoteCluster CR has the expected state.
    [Arguments]    ${alias}    ${cr_name}    ${expected_state}
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io ${cr_name} -o jsonpath='{.status.state}'
    Should Be Equal As Strings    ${stdout}    ${expected_state}

RemoteCluster CR Name From IP
    [Documentation]    Compute the RemoteCluster CR name from a host IP (e.g. 192.168.1.2 -> c2cc-192-168-1-2).
    [Arguments]    ${ip}
    ${dashed}=    Replace String    ${ip}    .    -
    ${dashed}=    Replace String    ${dashed}    :    -
    RETURN    c2cc-${dashed}

Primary NextHop IP For Host
    [Documentation]    Return the IP used as primary next-hop for a host. Always returns IPv4
    ...    when available (matching OVN-K's IPv4-first convention), fallback to IPv6.
    [Arguments]    ${ipv4}    ${ipv6}=${EMPTY}
    ${result}=    Set Variable If    '${ipv4}' != ''    ${ipv4}    ${ipv6}
    RETURN    ${result}

Verify RemoteCluster Unhealthy On Observers
    [Documentation]    Wait for observer clusters to report a disrupted cluster's CR as Unhealthy.
    [Arguments]    ${disrupted_ip}    @{observers}    ${disrupted_ipv6}=${EMPTY}
    ${primary}=    Primary NextHop IP For Host    ${disrupted_ip}    ${disrupted_ipv6}
    ${cr_name}=    RemoteCluster CR Name From IP    ${primary}
    FOR    ${observer}    IN    @{observers}
        Wait Until Keyword Succeeds    3m    10s
        ...    Verify RemoteCluster State By Name    ${observer}    ${cr_name}    Unhealthy
    END

Ensure All Clusters Healthy
    [Documentation]    Pre-condition: all clusters must be Healthy before fault injection.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Wait Until Keyword Succeeds    3m    10s
        ...    Verify RemoteCluster State    ${alias}    Healthy
    END

IP Command For CIDR
    [Documentation]    Return 'ip -6' if the given CIDR is IPv6, otherwise 'ip -4'.
    [Arguments]    ${cidr}
    ${is_ipv6}=    Evaluate    ':' in '''${cidr}'''
    ${cmd}=    Set Variable If    ${is_ipv6}    ip -6    ip -4
    RETURN    ${cmd}

Verify C2CC Infrastructure For Remote
    [Documentation]    Verify all C2CC infrastructure for one remote cluster's CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}    ${ip_cmd}=${IP_CMD}
    Verify Routes In Table 200    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${ip_cmd}
    Verify All IP Rules    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}    ${ip_cmd}
    Verify NFTables Bypass Rules    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify OVN Static Routes    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify Node SNAT Annotation    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify C2CC Tracking Annotation    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}

Verify C2CC Infrastructure On Cluster
    [Documentation]    Run all C2CC infrastructure checks for all remote CIDRs on the given cluster.
    [Arguments]    ${alias}
    Verify C2CC Infrastructure For CIDR Family    ${alias}    ${POD_CIDR_MAP}    ${SVC_CIDR_MAP}
    ${local_svc_cidr_dual}=    Get From Dictionary    ${SVC_CIDR_MAP_DUAL}    ${alias}
    IF    '${local_svc_cidr_dual}' != ''
        Verify C2CC Infrastructure For CIDR Family    ${alias}    ${POD_CIDR_MAP_DUAL}    ${SVC_CIDR_MAP_DUAL}
    END

Verify C2CC Infrastructure For CIDR Family
    [Documentation]    Verify C2CC infrastructure for one CIDR family (primary or dual) on a cluster.
    [Arguments]    ${alias}    ${pod_map}    ${svc_map}
    ${local_svc_cidr}=    Get From Dictionary    ${svc_map}    ${alias}
    ${ip_cmd}=    IP Command For CIDR    ${local_svc_cidr}
    FOR    ${remote}    IN    @{ALL_CLUSTERS}
        IF    '${remote}' != '${alias}'
            ${remote_pod_cidr}=    Get From Dictionary    ${pod_map}    ${remote}
            ${remote_svc_cidr}=    Get From Dictionary    ${svc_map}    ${remote}
            Verify C2CC Infrastructure For Remote
            ...    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}    ${ip_cmd}
        END
    END
    Verify Routes In Table 201    ${alias}    ${local_svc_cidr}    ${ip_cmd}

Wait For OVN-K Pods Ready On Cluster
    [Documentation]    Wait for all OVN-K pods to be Running and Ready on the given cluster.
    [Arguments]    ${alias}    ${timeout}=5m    ${retry}=15s
    Wait Until Keyword Succeeds    ${timeout}    ${retry}
    ...    Verify OVN-K Pods Ready    ${alias}

Verify OVN-K Pods Ready
    [Documentation]    Assert all pods in openshift-ovn-kubernetes are Ready.
    [Arguments]    ${alias}
    Oc On Cluster    ${alias}
    ...    oc wait pods -n openshift-ovn-kubernetes --all --for=condition=Ready --timeout=30s

Disable All NICs For VM
    [Documentation]    Disable all vnet interfaces for the given VM via virsh.
    ...    Returns the list of disabled interfaces for later re-enabling.
    [Arguments]    ${vm_name}
    ${vnet_ifaces}=    Get Vnet Devices For MicroShift Host    ${vm_name}
    Should Not Be Empty    ${vnet_ifaces}    No vnet interfaces found for VM ${vm_name}
    FOR    ${iface}    IN    @{vnet_ifaces}
        Disable NIC For MicroShift Host    ${vm_name}    ${iface}
    END
    RETURN    ${vnet_ifaces}

Enable All NICs For VM
    [Documentation]    Re-enable all specified vnet interfaces for the given VM via virsh.
    [Arguments]    ${vm_name}    ${vnet_ifaces}
    FOR    ${iface}    IN    @{vnet_ifaces}
        Enable NIC For MicroShift Host    ${vm_name}    ${iface}
    END

Deregister Remote Cluster
    [Documentation]    Close SSH connection and remove a remote cluster from all tracking structures.
    ...    Safe to call even if the alias was already deregistered.
    [Arguments]    ${alias}
    ${switch_status}    ${_}=    Run Keyword And Ignore Error    SSHLibrary.Switch Connection    ${alias}
    IF    '${switch_status}' == 'PASS'
        Run Keyword And Ignore Error    SSHLibrary.Close Connection
    END
    ${present}=    Evaluate    '${alias}' in $C2CC_REMOTE_ALIASES
    IF    ${present}
        ${idx}=    Evaluate    $C2CC_REMOTE_ALIASES.index('${alias}')
        Remove From List    ${C2CC_REMOTE_ALIASES}    ${idx}
    END
    Run Keyword And Ignore Error    Remove From Dictionary    ${C2CC_SSH_IDS}    ${alias}
    Run Keyword And Ignore Error    Remove From Dictionary    ${C2CC_KUBECONFIGS}    ${alias}

Reconnect To Cluster
    [Documentation]    Re-establish SSH connection to a remote cluster after network disruption.
    ...    Deregisters the stale connection and waits for SSH to come back.
    [Arguments]    ${alias}    ${host}    ${ssh_port}    ${kubeconfig}    ${timeout}=5m
    Deregister Remote Cluster    ${alias}
    Wait Until Keyword Succeeds    ${timeout}    15s
    ...    Register Remote Cluster    ${alias}    ${host}    ${ssh_port}    ${kubeconfig}

Reboot Clusters Simultaneously
    [Documentation]    Reboot one or more clusters and wait for all to come back
    ...    with a new boot ID and greenboot health check passed.
    ...    Issues reboot on all targets before waiting, so reboots overlap.
    [Arguments]    @{cluster_aliases}
    VAR    &{boot_ids}=    &{EMPTY}
    FOR    ${alias}    IN    @{cluster_aliases}
        ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
        SSHLibrary.Switch Connection    ${conn_id}
        ${bootid}=    Get Current Boot Id
        Set To Dictionary    ${boot_ids}    ${alias}    ${bootid}
    END
    FOR    ${alias}    IN    @{cluster_aliases}
        ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
        SSHLibrary.Switch Connection    ${conn_id}
        SSHLibrary.Start Command    reboot    sudo=True
    END
    FOR    ${alias}    IN    @{cluster_aliases}
        ${old_bootid}=    Get From Dictionary    ${boot_ids}    ${alias}
        Wait Until Keyword Succeeds    10m    15s
        ...    Verify Cluster Has Rebooted    ${alias}    ${old_bootid}
    END

Verify Cluster Has Rebooted
    [Documentation]    Assert that a cluster has rebooted and greenboot health check has passed.
    [Arguments]    ${alias}    ${old_bootid}
    Reconnect Cluster    ${alias}
    ${new_bootid}=    Get Current Boot Id
    Should Not Be Equal    ${new_bootid}    ${old_bootid}
    ${stdout}=    Command On Cluster    ${alias}
    ...    systemctl show -p SubState greenboot-healthcheck.service --value
    Should Not Be Equal As Strings    ${stdout}    failed    strip_spaces=True
    ...    msg=Greenboot healthcheck FAILED on ${alias}. Check greenboot logs.
    Should Be Equal As Strings    ${stdout}    exited    strip_spaces=True

Reconnect Cluster
    [Documentation]    Re-establish SSH connection to a cluster after reboot.
    ...    Resolves host/port from the alias and delegates to Reconnect To Cluster.
    [Arguments]    ${alias}
    IF    '${alias}' == 'cluster-a'
        VAR    ${host}=    ${USHIFT_HOST}
        VAR    ${port}=    ${SSH_PORT}
    ELSE IF    '${alias}' == 'cluster-b'
        VAR    ${host}=    ${HOST2_IP}
        VAR    ${port}=    ${HOST2_SSH_PORT}
    ELSE IF    '${alias}' == 'cluster-c'
        VAR    ${host}=    ${HOST3_IP}
        VAR    ${port}=    ${HOST3_SSH_PORT}
    ELSE
        Fail    Unknown cluster alias: ${alias}
    END
    ${kubeconfig}=    Get From Dictionary    ${C2CC_KUBECONFIGS}    ${alias}
    Reconnect To Cluster    ${alias}    ${host}    ${port}    ${kubeconfig}    timeout=10m

C2CC Suite Setup
    [Documentation]    Standard 3-cluster suite setup: registers all three clusters
    ...    as remote and optionally deploys test workloads.
    [Arguments]    ${deploy_workloads}=${FALSE}
    Check Required Env Variables
    Register All C2CC Clusters
    IF    ${deploy_workloads}    Deploy Test Workloads

C2CC Suite Teardown
    [Documentation]    Standard suite teardown: optionally cleans up workloads
    ...    and tears down remote clusters.
    [Arguments]    ${cleanup_workloads}=${FALSE}
    IF    ${cleanup_workloads}    Cleanup Test Workloads
    Teardown All Remote Clusters

Verify Full C2CC Connectivity
    [Documentation]    Verify pod-to-pod, pod-to-service connectivity and source IP preservation
    ...    across all 6 cluster pairs.
    FOR    ${src}    ${dst}    IN
    ...    cluster-a    cluster-b
    ...    cluster-a    cluster-c
    ...    cluster-b    cluster-a
    ...    cluster-b    cluster-c
    ...    cluster-c    cluster-a
    ...    cluster-c    cluster-b
        Verify Connectivity Between Clusters    ${src}    ${dst}    pod
        Verify Connectivity Between Clusters    ${src}    ${dst}    service
        Verify Source IP Preserved Between Clusters    ${src}    ${dst}    pod
        Verify Source IP Preserved Between Clusters    ${src}    ${dst}    service
    END

Verify Full C2CC DNS
    [Documentation]    Verify CoreDNS Corefile contains C2CC server blocks and
    ...    cross-cluster DNS resolution works for all pairs.
    FOR    ${alias}    IN    @{ALL_CLUSTERS}
        FOR    ${remote}    IN    @{ALL_CLUSTERS}
            IF    '${remote}' != '${alias}'
                ${domain}=    Get From Dictionary    ${DOMAIN_MAP}    ${remote}
                Verify Corefile Contains C2CC Server Block    ${alias}    ${domain}
            END
        END
    END
    FOR    ${src}    ${dst}    IN
    ...    cluster-a    cluster-b
    ...    cluster-a    cluster-c
    ...    cluster-b    cluster-a
    ...    cluster-b    cluster-c
    ...    cluster-c    cluster-a
    ...    cluster-c    cluster-b
        Verify Remote Service Via DNS    ${src}    ${dst}
    END

Verify C2CC Health Probes
    [Documentation]    Verify all RemoteCluster CRs report Healthy with populated timestamps.
    FOR    ${alias}    IN    @{ALL_CLUSTERS}
        Verify RemoteCluster State    ${alias}    Healthy
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.lastProbeTime}'
        Should Not Be Empty    ${stdout}
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.lastSuccessfulProbe}'
        Should Not Be Empty    ${stdout}
    END

Verify All Clusters Healthy
    [Documentation]    Wait for all clusters to pass the MicroShift health check.
    [Arguments]    ${timeout}=5m    ${retry}=15s
    FOR    ${cluster}    IN    @{ALL_CLUSTERS}
        Wait Until Keyword Succeeds    ${timeout}    ${retry}
        ...    Verify Cluster Is Healthy    ${cluster}
    END

Verify Full C2CC Stack
    [Documentation]    Comprehensive verification of all C2CC components across all clusters.
    ...    Wraps connectivity, infrastructure, health probes, and DNS checks
    ...    with retry logic for post-disruption convergence.
    Wait Until Keyword Succeeds    10m    10s    Verify Full C2CC Connectivity
    FOR    ${cluster}    IN    @{ALL_CLUSTERS}
        Wait Until Keyword Succeeds    10m    10s
        ...    Verify C2CC Infrastructure On Cluster    ${cluster}
    END
    Wait Until Keyword Succeeds    10m    10s    Verify C2CC Health Probes
    Wait Until Keyword Succeeds    10m    10s    Verify Full C2CC DNS
