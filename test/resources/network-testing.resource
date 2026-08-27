*** Settings ***
Documentation       Reusable keywords for networking, service, route, and
...                 infrastructure tests.

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
    Run With Kubeconfig
    ...    oc run ${name} -n ${ns} ${label_arg} --image=quay.io/microshift/busybox:1.36 --command -- /bin/sh -c "while true; do echo -ne 'HTTP/1.0 200 OK\\r\\nContent-Length: 16\\r\\n\\r\\nHello MicroShift' | nc -l -p ${port}; done"
    Labeled Pod Should Be Ready    name\=${name}    ns=${ns}

Get Pod IP
    [Documentation]    Return the IP address of the named pod.
    [Arguments]    ${name}    ${ns}
    ${ip}=    Oc Get JsonPath    pod    ${ns}    ${name}    .status.podIP
    RETURN    ${ip}

Curl From Pod Should Succeed
    [Documentation]    Exec a curl from inside a pod and assert HTTP success.
    [Arguments]    ${pod}    ${ns}    ${url}    ${timeout}=5    ${expected}=Hello
    ${stdout}=    Run With Kubeconfig
    ...    oc exec ${pod} -n ${ns} -- curl -s --max-time ${timeout} ${url}
    Should Contain    ${stdout}    ${expected}

Curl From Pod Should Fail
    [Documentation]    Exec a curl from inside a pod and assert it fails (timeout or connection refused).
    [Arguments]    ${pod}    ${ns}    ${url}    ${timeout}=5
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    KUBECONFIG\=${KUBECONFIG} oc exec ${pod} -n ${ns} -- curl -s --max-time ${timeout} ${url}
    ...    sudo=False    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Not Be Equal As Integers    ${rc}    0

Curl From Pod Should Timeout
    [Documentation]    Exec a curl from inside a pod and assert it returns exit code 28 (timeout).
    [Arguments]    ${pod}    ${ns}    ${url}    ${timeout}=5
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    KUBECONFIG\=${KUBECONFIG} oc exec ${pod} -n ${ns} -- curl -s --max-time ${timeout} ${url}
    ...    sudo=False    return_rc=True    return_stderr=True    return_stdout=True
    Log Many    ${stdout}    ${stderr}    ${rc}
    Should Be Equal As Integers    ${rc}    28
    ...    msg=Expected curl timeout (rc=28) but got rc=${rc}. stdout: ${stdout}, stderr: ${stderr}

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
    [Documentation]    Assert that the service exists but does NOT have a load balancer ingress IP.
    [Arguments]    ${name}    ${ns}
    Run With Kubeconfig    oc get service ${name} -n ${ns}
    ${ip}=    Run With Kubeconfig
    ...    oc get service ${name} -n ${ns} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    ...    allow_fail=True
    Should Be Empty    ${ip}

Get Router Pod IP
    [Documentation]    Return the IP of the router-default pod in openshift-ingress.
    ${ip}=    Run With Kubeconfig
    ...    oc get pod -n openshift-ingress -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller\=default -o jsonpath='{.items[0].status.podIP}'
    Should Not Be Empty    ${ip}
    RETURN    ${ip}

Get Router LB IP
    [Documentation]    Return the LoadBalancer IP of the router-default service.
    ${ip}=    Oc Get JsonPath    service    openshift-ingress    router-default
    ...    .status.loadBalancer.ingress[0].ip
    Should Not Be Empty    ${ip}
    RETURN    ${ip}

Verify Route Is Admitted
    [Documentation]    Assert the named route has been admitted by the router.
    [Arguments]    ${name}    ${ns}
    ${status}=    Run With Kubeconfig
    ...    oc get route ${name} -n ${ns} -o jsonpath='{.status.ingress[0].conditions[?(@.type=="Admitted")].status}'
    Should Be Equal As Strings    ${status}    True

Curl Route Should Succeed
    [Documentation]    Curl a route via --resolve and assert HTTP 200.
    [Arguments]    ${hostname}    ${ip}    ${port}    ${scheme}=https
    ${result}=    Run Process
    ...    curl -k -s -o /dev/null -w "%{http_code}" --max-time 10 --resolve ${hostname}:${port}:${ip} ${scheme}://${hostname}:${port}/
    ...    shell=True    timeout=15s
    Log Many    stdout: ${result.stdout}    stderr: ${result.stderr}    rc: ${result.rc}
    Should Be Equal As Integers    ${result.rc}    0
    ...    msg=curl failed with rc=${result.rc}: ${result.stderr}
    Should Be Equal As Strings    ${result.stdout}    200
    ...    msg=Expected HTTP 200 but got ${result.stdout} for ${scheme}://${hostname}:${port}/

Verify HAProxy Backend Exists
    [Documentation]    Check that the haproxy config in the router contains the expected backend.
    [Arguments]    ${backend_pattern}
    ${config}=    Run With Kubeconfig
    ...    oc rsh -n openshift-ingress deployment/router-default cat /var/lib/haproxy/conf/haproxy.config
    Should Contain    ${config}    ${backend_pattern}

Label Namespace
    [Documentation]    Add a label to a namespace.
    [Arguments]    ${ns}    ${label}
    Run With Kubeconfig    oc label namespace ${ns} ${label} --overwrite

Conntrack Should Contain
    [Documentation]    Assert that conntrack table contains entries matching the pattern.
    [Arguments]    ${proto}    ${port}    ${pattern}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    conntrack -L -p ${proto} --dport ${port}
    IF    ${rc} != 0
        Fail    conntrack command failed with rc=${rc}: ${stderr}
    END
    Should Contain    ${stdout}    ${pattern}

Conntrack Should Not Contain
    [Documentation]    Assert that conntrack table does NOT contain entries matching the pattern.
    ...    conntrack returns rc=1 when no entries are found, which is acceptable.
    [Arguments]    ${proto}    ${port}    ${pattern}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    conntrack -L -p ${proto} --dport ${port}
    IF    ${rc} != 0 and ${rc} != 1
        Fail    conntrack command failed with rc=${rc}: ${stderr}
    END
    Should Not Contain    ${stdout}    ${pattern}

Wait For Router Ready
    [Documentation]    Wait for the default router to be ready.
    Oc Wait    namespace/openshift-ingress    --for jsonpath='{.status.phase}=Active' --timeout=5m
    Named Deployment Should Be Available    router-default    openshift-ingress    5m

Restart Router
    [Documentation]    Restart the router and wait for readiness.
    Run With Kubeconfig    oc rollout restart deployment router-default -n openshift-ingress
    Wait For Router Ready
