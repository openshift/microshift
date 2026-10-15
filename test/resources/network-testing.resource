*** Settings ***
Documentation       Reusable keywords for networking and service tests.

Library             SSHLibrary
Resource            oc.resource
Resource            common.resource
Resource            microshift-host.resource
Resource            microshift-network.resource


*** Keywords ***
Create Labeled Pod
    [Documentation]    Create a hello-microshift-style pod with the given name and labels.
    ...    Uses 'oc run' to create the pod inline with a simple HTTP server on port 8080.
    [Arguments]    ${name}    ${ns}    ${labels}=${EMPTY}    ${port}=8080
    ${label_arg}=    Set Variable If    '${labels}' != '${EMPTY}'    --labels=${labels}    ${EMPTY}
    # robocop: off=line-too-long
    Run With Kubeconfig
    ...    oc run ${name} -n ${ns} ${label_arg} --image=quay.io/microshift/busybox:1.36 --command -- /bin/sh -c "while true; do echo -ne 'HTTP/1.0 200 OK\\r\\nContent-Length: 16\\r\\n\\r\\nHello MicroShift' | nc -l -p ${port}; done"
    # robocop: on=line-too-long
    # Wait by pod name: the pod's label (${labels}) differs from its name.
    Oc Wait    pod/${name} -n ${ns}    --for=condition=Ready --timeout=120s

Get Pod IP
    [Documentation]    Return the IP address of the named pod.
    [Arguments]    ${name}    ${ns}
    ${ip}=    Oc Get JsonPath    pod    ${ns}    ${name}    .status.podIP
    RETURN    ${ip}

Curl From Pod Should Succeed
    [Documentation]    Exec wget from inside a (busybox) pod and assert HTTP success.
    ...    Pods use the busybox image, which has wget but not curl. 'oc exec' runs on the
    ...    test runner, so use the runner's KUBECONFIG via Run With Kubeconfig.
    [Arguments]    ${pod}    ${ns}    ${url}    ${timeout}=5    ${expected}=Hello
    ${stdout}=    Run With Kubeconfig
    ...    oc exec ${pod} -n ${ns} -- wget -q -T ${timeout} -O- ${url}
    Should Contain    ${stdout}    ${expected}

Curl From Pod Should Be Blocked
    [Documentation]    Exec wget from inside a (busybox) pod and assert the connection is blocked.
    ...    busybox wget returns non-zero on a dropped/blocked/refused connection (no distinct
    ...    timeout code). Pair every call with a positive Curl From Pod Should Succeed on the
    ...    same source pod so a non-zero result means "blocked", not "pod/exec broken".
    [Arguments]    ${pod}    ${ns}    ${url}    ${timeout}=5
    ${stdout}    ${rc}=    Run With Kubeconfig
    ...    oc exec ${pod} -n ${ns} -- wget -q -T ${timeout} -O- ${url}
    ...    allow_fail=True    return_rc=True
    Log    ${stdout}
    Should Not Be Equal As Integers    ${rc}    0
    ...    msg=Expected blocked connection (non-zero wget rc) but got rc=${rc}. stdout: ${stdout}

Get Service ClusterIP
    [Documentation]    Return the ClusterIP of the named service.
    [Arguments]    ${name}    ${ns}
    ${ip}=    Oc Get JsonPath    service    ${ns}    ${name}    .spec.clusterIP
    RETURN    ${ip}

Get Service NodePort
    [Documentation]    Return the first NodePort of the named service.
    [Arguments]    ${name}    ${ns}
    ${port}=    Oc Get JsonPath    service    ${ns}    ${name}    .spec.ports[0].nodePort
    RETURN    ${port}

Wait For LoadBalancer IP
    [Documentation]    Poll until the service has a load balancer ingress IP. Return the IP.
    [Arguments]    ${name}    ${ns}    ${retries}=30    ${interval}=10s
    Wait Until Keyword Succeeds    ${retries}x    ${interval}
    ...    Service Should Have LB IP    ${name}    ${ns}
    ${ip}=    Oc Get JsonPath    service    ${ns}    ${name}    .status.loadBalancer.ingress[0].ip
    RETURN    ${ip}

Service Should Have LB IP
    [Documentation]    Assert that the service has a load balancer ingress IP.
    [Arguments]    ${name}    ${ns}
    ${ip}=    Oc Get JsonPath    service    ${ns}    ${name}    .status.loadBalancer.ingress[0].ip
    Should Not Be Empty    ${ip}

Service Should Not Have LB IP
    [Documentation]    Assert that the service has no load balancer ingress IP. The jsonpath read
    ...    returns empty (rc 0) when the field is absent; a real API error fails loudly.
    [Arguments]    ${name}    ${ns}
    ${ip}=    Run With Kubeconfig
    ...    oc get service ${name} -n ${ns} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    Should Be Empty    ${ip}

Service Should Not Get LB IP
    [Documentation]    Assert the service does NOT acquire a LoadBalancer IP over a sustained
    ...    window, not merely at one instant (which could just be pre-reconciliation).
    [Arguments]    ${name}    ${ns}    ${attempts}=5    ${interval}=3s
    FOR    ${idx}    IN RANGE    ${attempts}
        Service Should Not Have LB IP    ${name}    ${ns}
        Sleep    ${interval}
    END

Set Namespace Privileged
    [Documentation]    Relax Pod Security on a namespace so test pods created directly (oc run),
    ...    pods using hostPort, or pods with a fixed runAsUser are admitted.
    [Arguments]    ${ns}=${NAMESPACE}
    # robocop: off=line-too-long
    Run With Kubeconfig
    ...    oc label namespace ${ns} --overwrite pod-security.kubernetes.io/enforce=privileged pod-security.kubernetes.io/warn=privileged pod-security.kubernetes.io/audit=privileged
    # robocop: on=line-too-long

Label Namespace
    [Documentation]    Add a label to a namespace.
    [Arguments]    ${ns}    ${label}
    Run With Kubeconfig    oc label namespace ${ns} ${label} --overwrite

Conntrack Should Contain
    [Documentation]    Assert the conntrack table has a UDP entry referencing the pattern (a pod IP).
    ...    Matches on "<ip> " (trailing space) so e.g. 10.42.0.5 does not match 10.42.0.50.
    ...    No --dport filter: the pod-side (DNAT'd) entry is tracked on the pod port, not the NodePort.
    [Arguments]    ${proto}    ${pattern}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    conntrack -L -p ${proto}
    IF    ${rc} != 0
        Fail    conntrack command failed with rc=${rc}: ${stderr}
    END
    Should Contain    ${stdout}    ${pattern}${SPACE}

Conntrack Should Not Contain
    [Documentation]    Assert the conntrack table has no UDP entry referencing the pattern (a pod IP).
    ...    Matches on "<ip> " (trailing space) to avoid substring false matches.
    ...    conntrack returns rc=1 when no entries are found, which is acceptable.
    [Arguments]    ${proto}    ${pattern}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    conntrack -L -p ${proto}
    IF    ${rc} != 0 and ${rc} != 1
        Fail    conntrack command failed with rc=${rc}: ${stderr}
    END
    Should Not Contain    ${stdout}    ${pattern}${SPACE}
