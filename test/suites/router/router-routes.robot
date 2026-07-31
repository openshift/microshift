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
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}
    ...    https://service-secure-test.example.com:443    service-secure-test.example.com:443:${router_ip}
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

    Create OC Route And Admit    ${NAMESPACE}    passthrough    ms-pass    service-secure    --hostname=${pass_host}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}    https://${pass_host}:443    ${pass_host}:443:${router_ip}
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_tcp:${NAMESPACE}:ms-pass

    Create OC Route And Admit    ${NAMESPACE}    edge    ms-edge    service-unsecure    --hostname=${edge_host}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}    https://${edge_host}:443    ${edge_host}:443:${router_ip}
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_edge_http:${NAMESPACE}:ms-edge
    [Teardown]    Run Keywords
    ...    Oc Delete    route/ms-pass route/ms-edge -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${WEB_SERVER_DEPLOY} -n ${NAMESPACE} --ignore-not-found

HTTP And Reencrypt Routes
    [Documentation]    Verify HTTP route via oc expose and reencrypt route creation and connectivity.
    ...    OCP-60283
    [Setup]    Deploy Web Server Signed

    ${router_ip}=    Get Router Pod IP
    ${srv_pod}=    Get Web Server Pod Name
    VAR    ${http_host}=    route-http-60283.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen-60283.${BASE_DOMAIN}

    Create OC Route And Admit    ${NAMESPACE}    http    ms-http    service-unsecure    --hostname=${http_host}
    Wait Until Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}    http://${http_host}:80    ${http_host}:80:${router_ip}
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_http:${NAMESPACE}:ms-http

    Create OC Route And Admit    ${NAMESPACE}    reencrypt    ms-reen    service-secure    --hostname=${reen_host}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${srv_pod}    ${NAMESPACE}    https://${reen_host}:443    ${reen_host}:443:${router_ip}
    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    backend be_secure:${NAMESPACE}:ms-reen
    [Teardown]    Run Keywords
    ...    Oc Delete    route/ms-http route/ms-reen -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE} --ignore-not-found

Namespace Ownership Default Config
    [Documentation]    Verify the default InterNamespaceAllowed config allows routes from different
    ...    namespaces to share the same hostname with different paths.
    ...    OCP-72802
    [Setup]    Setup Two Namespace Test

    ${router_ip}=    Get Router Pod IP
    VAR    ${HTTP_HOST}=    service-unsecure-ocp72802.${BASE_DOMAIN}    scope=TEST
    VAR    ${EDGE_HOST}=    route-edge-ocp72802.${BASE_DOMAIN}    scope=TEST
    VAR    ${REEN_HOST}=    route-reen-ocp72802.${BASE_DOMAIN}    scope=TEST

    ${env}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}
    ...    .items[*].spec.containers[*].env[?(@.name=="ROUTER_DISABLE_NAMESPACE_OWNERSHIP_CHECK")].value
    Should Be Equal As Strings    ${env}    true

    Create NS Ownership Routes
    All NS Ownership Routes Should Be Admitted

    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NS1}    http://${HTTP_HOST}/path/index.html    ${HTTP_HOST}:80:${router_ip}
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NS1}    http://${HTTP_HOST}/test/index.html    ${HTTP_HOST}:80:${router_ip}
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

    VAR    ${HTTP_HOST}=    service-unsecure-ocp73152.${BASE_DOMAIN}    scope=TEST
    VAR    ${EDGE_HOST}=    route-edge-ocp73152.${BASE_DOMAIN}    scope=TEST
    VAR    ${PASS_HOST}=    route-passth-ocp73152.${BASE_DOMAIN}    scope=TEST
    VAR    ${REEN_HOST}=    route-reen-ocp73152.${BASE_DOMAIN}    scope=TEST

    Create And Admit Four Route Types    ${HTTP_HOST}    ${EDGE_HOST}    ${PASS_HOST}    ${REEN_HOST}

    ${lb_ip}=    Fetch From Left    ${lb_ips}    ${SPACE}
    Curl All LB Route Types Via IP    ${lb_ip}
    [Teardown]    Run Keywords
    ...    Oc Delete    route/route-http route/route-edge route/route-passth route/route-reen -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${TEST_CLIENT_POD} -n ${NAMESPACE} --ignore-not-found

Default Listening IPs And Ports
    [Documentation]    Verify the router-default service LB IPs match all host IPs, default ports
    ...    are 80 and 443, and all route types are reachable through each LB IP.
    ...    OCP-73202
    [Setup]    Run Keywords
    ...    Deploy Web Server Signed
    ...    AND    Deploy Test Client Pod

    Verify Default LB IPs And Ports

    VAR    ${HTTP_HOST}=    service-unsecure-ocp73202.${BASE_DOMAIN}    scope=TEST
    VAR    ${EDGE_HOST}=    route-edge-ocp73202.${BASE_DOMAIN}    scope=TEST
    VAR    ${PASS_HOST}=    route-passth-ocp73202.${BASE_DOMAIN}    scope=TEST
    VAR    ${REEN_HOST}=    route-reen-ocp73202.${BASE_DOMAIN}    scope=TEST

    Create And Admit Four Route Types    ${HTTP_HOST}    ${EDGE_HOST}    ${PASS_HOST}    ${REEN_HOST}

    ${lb_ips}=    Get LB IPs
    @{ips}=    Split String    ${lb_ips}
    FOR    ${lb_ip}    IN    @{ips}
        Curl All LB Route Types Via IP    ${lb_ip}
    END
    [Teardown]    Run Keywords
    ...    Oc Delete    route/route-http route/route-edge route/route-passth route/route-reen -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${TEST_CLIENT_POD} -n ${NAMESPACE} --ignore-not-found


*** Keywords ***
Setup Reencrypt Ingress Test
    [Documentation]    Deploy web-server-signed and apply the destCA ingress.
    ...    The ingress-to-route controller generates Route names with a random suffix,
    ...    so we wait for a Route owned by the Ingress to appear and be admitted.
    Deploy Web Server Signed
    Oc Create    -f ${INGRESS_DESTCA} -n ${NAMESPACE}
    Ingress Route Should Be Admitted    ingress-ms-reen    timeout=300s

Teardown Reencrypt Ingress Test
    [Documentation]    Delete the destCA ingress and web-server deployment.
    Oc Delete    -f ${INGRESS_DESTCA} -n ${NAMESPACE} --ignore-not-found
    Oc Delete    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE} --ignore-not-found

Setup Two Namespace Test
    [Documentation]    Create two extra namespaces and deploy workloads in each.
    VAR    ${NS1}=    ${NAMESPACE}-ocp72802-1    scope=TEST
    VAR    ${NS2}=    ${NAMESPACE}-ocp72802-2    scope=TEST
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

Create NS Ownership Routes
    [Documentation]    Create HTTP, edge, and reencrypt routes in both test namespaces.
    Create OC Route    ${NS1}    http    service-unsecure    service-unsecure
    ...    --hostname=${HTTP_HOST}    --path=/path
    Create OC Route    ${NS1}    edge    route-edge    service-unsecure
    ...    --hostname=${EDGE_HOST}    --path=/path
    Create OC Route    ${NS1}    reencrypt    route-reen    service-secure
    ...    --hostname=${REEN_HOST}    --path=/path
    Create OC Route    ${NS2}    http    service-unsecure    service-unsecure
    ...    --hostname=${HTTP_HOST}    --path=/test
    Create OC Route    ${NS2}    edge    route-edge    service-unsecure
    ...    --hostname=${EDGE_HOST}    --path=/test
    Create OC Route    ${NS2}    reencrypt    route-reen    service-secure
    ...    --hostname=${REEN_HOST}    --path=/test

All NS Ownership Routes Should Be Admitted
    [Documentation]    Verify all six routes across both namespaces are admitted.
    Route Should Be Admitted    service-unsecure    ${NS1}
    Route Should Be Admitted    route-edge    ${NS1}
    Route Should Be Admitted    route-reen    ${NS1}
    Route Should Be Admitted    service-unsecure    ${NS2}
    Route Should Be Admitted    route-edge    ${NS2}
    Route Should Be Admitted    route-reen    ${NS2}

Curl All LB Route Types Via IP
    [Documentation]    Curl all four route types through a single LB IP.
    [Arguments]    ${ip}
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${HTTP_HOST}
    ...    ${HTTP_HOST}:80:${ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${EDGE_HOST}    ${EDGE_HOST}:443:${ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${PASS_HOST}    ${PASS_HOST}:443:${ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${REEN_HOST}    ${REEN_HOST}:443:${ip}

Verify Default LB IPs And Ports
    [Documentation]    Verify the router-default service LB IPs match all host IPs and ports are 80/443.
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
