*** Settings ***
Documentation       Router end-to-end route tests
...                 Migrated from openshift-tests-private:
...                 OCP-60136, OCP-60266, OCP-60283, OCP-72802, OCP-73152, OCP-73202

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/router.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Variables ***
${BASE_DOMAIN}      apps.example.com


*** Test Cases ***
Reencrypt Route Via Ingress With Destination CA
    [Documentation]    Verify a reencrypt route created via a Kubernetes Ingress resource with
    ...    destination CA certificate is admitted and reachable.
    ...    OCP-60136
    [Setup]    Setup Reencrypt Ingress Test

    ${router_ip}=    Get Router Pod IP
    ${srv_pod}=    Get Web Server Pod Name
    Wait Until Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}
    ...    https://service-secure-test.example.com:443    service-secure-test.example.com:443:${router_ip}
    ...    200    -k
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_secure:${NAMESPACE}:ingress-ms-reen
    [Teardown]    Teardown Reencrypt Ingress Test

Edge And Passthrough Routes
    [Documentation]    Verify edge-terminated and passthrough route creation and connectivity.
    ...    OCP-60266
    [Setup]    Deploy Web Server

    ${router_ip}=    Get Router Pod IP
    ${srv_pod}=    Get Web Server Pod Name
    VAR    ${pass_host}=    route-pass-60266.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge-60266.${BASE_DOMAIN}

    Create OC Route    ${NAMESPACE}    passthrough    ms-pass    service-secure
    ...    --hostname=${pass_host}
    Route Should Be Admitted    ms-pass
    Wait Until Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}
    ...    https://${pass_host}:443    ${pass_host}:443:${router_ip}
    ...    200    -k
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_tcp:${NAMESPACE}:ms-pass

    Create OC Route    ${NAMESPACE}    edge    ms-edge    service-unsecure
    ...    --hostname=${edge_host}
    Route Should Be Admitted    ms-edge
    Wait Until Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}
    ...    https://${edge_host}:443    ${edge_host}:443:${router_ip}
    ...    200    -k
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_edge_http:${NAMESPACE}:ms-edge

    [Teardown]    Run Keywords
    ...    Oc Delete    route/ms-pass route/ms-edge -n ${NAMESPACE}
    ...    AND    Oc Delete    -f ${WEB_SERVER_DEPLOY} -n ${NAMESPACE}

HTTP And Reencrypt Routes
    [Documentation]    Verify HTTP route via oc expose and reencrypt route creation and connectivity.
    ...    OCP-60283
    [Setup]    Deploy Web Server Signed

    ${router_ip}=    Get Router Pod IP
    ${srv_pod}=    Get Web Server Pod Name
    VAR    ${http_host}=    route-http-60283.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen-60283.${BASE_DOMAIN}

    Create OC Route    ${NAMESPACE}    http    ms-http    service-unsecure
    ...    --hostname=${http_host}
    Route Should Be Admitted    ms-http
    Wait Until Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}
    ...    http://${http_host}:80    ${http_host}:80:${router_ip}
    ...    200
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_http:${NAMESPACE}:ms-http

    Create OC Route    ${NAMESPACE}    reencrypt    ms-reen    service-secure
    ...    --hostname=${reen_host}
    Route Should Be Admitted    ms-reen
    Wait Until Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}
    ...    https://${reen_host}:443    ${reen_host}:443:${router_ip}
    ...    200    -k
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_secure:${NAMESPACE}:ms-reen

    [Teardown]    Run Keywords
    ...    Oc Delete    route/ms-http route/ms-reen -n ${NAMESPACE}
    ...    AND    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE}

Namespace Ownership Default Config
    [Documentation]    Verify the default InterNamespaceAllowed config allows routes from different
    ...    namespaces to share the same hostname with different paths.
    ...    OCP-72802
    [Setup]    Setup Two Namespace Test

    ${router_ip}=    Get Router Pod IP
    VAR    ${http_host}=    service-unsecure-ocp72802.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge-ocp72802.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen-ocp72802.${BASE_DOMAIN}

    ${env}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}
    ...    .items[*].spec.containers[*].env[?(@.name=="ROUTER_DISABLE_NAMESPACE_OWNERSHIP_CHECK")].value
    Should Be Equal As Strings    ${env}    true

    Create OC Route    ${NS1}    http    service-unsecure    service-unsecure
    ...    --hostname=${http_host}    --path=/path
    Create OC Route    ${NS1}    edge    route-edge    service-unsecure
    ...    --hostname=${edge_host}    --path=/path
    Create OC Route    ${NS1}    reencrypt    route-reen    service-secure
    ...    --hostname=${reen_host}    --path=/path
    Create OC Route    ${NS2}    http    service-unsecure    service-unsecure
    ...    --hostname=${http_host}    --path=/test
    Create OC Route    ${NS2}    edge    route-edge    service-unsecure
    ...    --hostname=${edge_host}    --path=/test
    Create OC Route    ${NS2}    reencrypt    route-reen    service-secure
    ...    --hostname=${reen_host}    --path=/test

    Route Should Be Admitted    service-unsecure    ${NS1}
    Route Should Be Admitted    route-edge    ${NS1}
    Route Should Be Admitted    route-reen    ${NS1}
    Route Should Be Admitted    service-unsecure    ${NS2}
    Route Should Be Admitted    route-edge    ${NS2}
    Route Should Be Admitted    route-reen    ${NS2}

    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NS1}
    ...    http://${http_host}/path/index.html    ${http_host}:80:${router_ip}
    ...    200
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NS1}
    ...    http://${http_host}/test/index.html    ${http_host}:80:${router_ip}
    ...    200

    [Teardown]    Teardown Two Namespace Test

Router Load Balancer Service Type
    [Documentation]    Verify the router-default service is of type LoadBalancer with IPs assigned,
    ...    and that all route types are reachable through the LB IP.
    ...    OCP-73152
    [Setup]    Run Keywords
    ...    Deploy Web Server Signed
    ...    AND    Deploy Test Client Pod

    ${svc_type}=    Oc Get JsonPath    service    ${ROUTER_NS}    router-default    .spec.type
    Should Be Equal As Strings    ${svc_type}    LoadBalancer
    ${lb_ips}=    Get LB IPs
    Should Not Be Empty    ${lb_ips}

    VAR    ${http_host}=    service-unsecure-ocp73152.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge-ocp73152.${BASE_DOMAIN}
    VAR    ${pass_host}=    route-passth-ocp73152.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen-ocp73152.${BASE_DOMAIN}

    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure
    ...    --hostname=${http_host}
    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure
    ...    --hostname=${edge_host}
    Create OC Route    ${NAMESPACE}    passthrough    route-passth    service-secure
    ...    --hostname=${pass_host}
    Create OC Route    ${NAMESPACE}    reencrypt    route-reen    service-secure
    ...    --hostname=${reen_host}
    Route Should Be Admitted    route-reen

    ${lb_ip}=    Fetch From Left    ${lb_ips}    ${SPACE}

    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    http://${http_host}    ${http_host}:80:${lb_ip}
    ...    200
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    https://${edge_host}    ${edge_host}:443:${lb_ip}
    ...    200    -k
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    https://${pass_host}    ${pass_host}:443:${lb_ip}
    ...    200    -k
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    https://${reen_host}    ${reen_host}:443:${lb_ip}
    ...    200    -k

    [Teardown]    Run Keywords
    ...    Oc Delete    route/route-http route/route-edge route/route-passth route/route-reen -n ${NAMESPACE}
    ...    AND    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE}
    ...    AND    Oc Delete    -f ${TEST_CLIENT_POD} -n ${NAMESPACE}

Default Listening IPs And Ports
    [Documentation]    Verify the router-default service LB IPs match all host IPs, default ports
    ...    are 80 and 443, and all route types are reachable through each LB IP.
    ...    OCP-73202
    [Setup]    Run Keywords
    ...    Deploy Web Server Signed
    ...    AND    Deploy Test Client Pod

    ${http_port}=    Get LB Port    http
    Should Be Equal As Strings    ${http_port}    80
    ${https_port}=    Get LB Port    https
    Should Be Equal As Strings    ${https_port}    443

    ${lb_ips}=    Get LB IPs
    Should Not Be Empty    ${lb_ips}
    @{host_ips}=    Get Host IPs Via SSH
    ${sorted_host_ips}=    Evaluate    " ".join(sorted(${host_ips}))
    ${sorted_lb_ips}=    Evaluate    " ".join(sorted("${lb_ips}".split()))
    Should Be Equal As Strings    ${sorted_lb_ips}    ${sorted_host_ips}

    VAR    ${http_host}=    service-unsecure-ocp73202.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge-ocp73202.${BASE_DOMAIN}
    VAR    ${pass_host}=    route-passth-ocp73202.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen-ocp73202.${BASE_DOMAIN}

    Create OC Route    ${NAMESPACE}    http    route-http    service-unsecure
    ...    --hostname=${http_host}
    Create OC Route    ${NAMESPACE}    edge    route-edge    service-unsecure
    ...    --hostname=${edge_host}
    Create OC Route    ${NAMESPACE}    passthrough    route-passth    service-secure
    ...    --hostname=${pass_host}
    Create OC Route    ${NAMESPACE}    reencrypt    route-reen    service-secure
    ...    --hostname=${reen_host}
    Route Should Be Admitted    route-reen

    @{ips}=    Split String    ${lb_ips}
    FOR    ${lb_ip}    IN    @{ips}
        Wait Until Curl Succeeds From Pod
        ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
        ...    http://${http_host}    ${http_host}:80:${lb_ip}
        ...    200
        Wait Until Curl Succeeds From Pod
        ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
        ...    https://${edge_host}    ${edge_host}:443:${lb_ip}
        ...    200    -k
        Wait Until Curl Succeeds From Pod
        ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
        ...    https://${pass_host}    ${pass_host}:443:${lb_ip}
        ...    200    -k
        Wait Until Curl Succeeds From Pod
        ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
        ...    https://${reen_host}    ${reen_host}:443:${lb_ip}
        ...    200    -k
    END

    [Teardown]    Run Keywords
    ...    Oc Delete    route/route-http route/route-edge route/route-passth route/route-reen -n ${NAMESPACE}
    ...    AND    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE}
    ...    AND    Oc Delete    -f ${TEST_CLIENT_POD} -n ${NAMESPACE}


*** Keywords ***
Setup Reencrypt Ingress Test
    [Documentation]    Deploy web-server-signed and apply the destCA ingress.
    Deploy Web Server Signed
    Oc Create    -f ${INGRESS_DESTCA} -n ${NAMESPACE}
    Route Should Be Admitted    ingress-ms-reen

Teardown Reencrypt Ingress Test
    [Documentation]    Delete the destCA ingress and web-server deployment.
    Oc Delete    -f ${INGRESS_DESTCA} -n ${NAMESPACE}
    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE}

Setup Two Namespace Test
    [Documentation]    Create two extra namespaces and deploy workloads in each.
    VAR    ${NS1}=    ${NAMESPACE}-ocp72802-1
    VAR    ${NS2}=    ${NAMESPACE}-ocp72802-2
    VAR    ${NS1}=    ${NS1}    scope=TEST
    VAR    ${NS2}=    ${NS2}    scope=TEST
    Create Namespace    ${NS1}
    Create Namespace    ${NS2}
    Deploy Test Client Pod    ${NS1}
    Deploy Web Server    ${NS1}
    Deploy Test Client Pod    ${NS2}
    Deploy Web Server    ${NS2}

Teardown Two Namespace Test
    [Documentation]    Delete the extra namespaces.
    Remove Namespace    ${NS1}
    Remove Namespace    ${NS2}
