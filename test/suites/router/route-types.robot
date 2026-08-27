*** Settings ***
Documentation       Route type tests: HTTP/edge/passthrough/reencrypt routes via both
...                 Route and Ingress resources, and router-as-LoadBalancer validation.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/network-testing.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow


*** Variables ***
${ROUTE_ASSETS}                 ./assets/route-types
${WEB_SERVER_DEPLOY}            ./assets/route-types/web-server-deploy.yaml
${WEB_SERVER_SIGNED_DEPLOY}     ./assets/route-types/web-server-signed-deploy.yaml
${INGRESS_HTTP}                 ./assets/route-types/ingress-http.yaml
${INGRESS_DESTCA}               ./assets/route-types/ingress-destca.yaml


*** Test Cases ***
HTTP Route Via Ingress
    [Documentation]    Verify that an Ingress resource creates an HTTP route that
    ...    correctly forwards traffic to the backend service.
    ...    Covers QE test 60149.
    [Setup]    Deploy Web Server Unsecure

    Oc Create    -f ${INGRESS_HTTP} -n ${NAMESPACE}
    ${route_name}=    Wait For Ingress Route    ingress-http    ${NAMESPACE}

    Verify Route Is Admitted    ${route_name}    ${NAMESPACE}

    ${router_ip}=    Get Router Pod IP
    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    service-unsecure-test.example.com    ${router_ip}    80    scheme=http

    Verify HAProxy Backend Exists    be_http:${NAMESPACE}:

    [Teardown]    Run Keywords
    ...    Run With Kubeconfig    oc delete -f ${INGRESS_HTTP} -n ${NAMESPACE}    allow_fail=True
    ...    AND
    ...    Teardown Web Server Unsecure

Edge And Passthrough Routes
    [Documentation]    Verify creation of edge and passthrough route types. Edge routes
    ...    terminate TLS at the router; passthrough routes forward TLS to the backend.
    ...    Covers QE test 60266.
    [Setup]    Deploy Web Server Unsecure And Signed

    ${router_ip}=    Get Router Pod IP

    # Passthrough route -> secure service
    Run With Kubeconfig
    ...    oc create route passthrough ms-pass --service=service-secure --port=27443 --hostname=route-pass.example.com -n ${NAMESPACE}
    Wait Until Keyword Succeeds    10x    2s
    ...    Verify Route Is Admitted    ms-pass    ${NAMESPACE}
    ${termination}=    Oc Get JsonPath    route    ${NAMESPACE}    ms-pass    .spec.tls.termination
    Should Be Equal As Strings    ${termination}    passthrough

    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    route-pass.example.com    ${router_ip}    443
    Verify HAProxy Backend Exists    be_tcp:${NAMESPACE}:ms-pass

    # Edge route -> unsecure service
    Run With Kubeconfig
    ...    oc create route edge ms-edge --service=service-unsecure --port=27017 --hostname=route-edge.example.com -n ${NAMESPACE}
    Wait Until Keyword Succeeds    10x    2s
    ...    Verify Route Is Admitted    ms-edge    ${NAMESPACE}
    ${termination}=    Oc Get JsonPath    route    ${NAMESPACE}    ms-edge    .spec.tls.termination
    Should Be Equal As Strings    ${termination}    edge

    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    route-edge.example.com    ${router_ip}    443
    Verify HAProxy Backend Exists    be_edge_http:${NAMESPACE}:ms-edge

    [Teardown]    Run Keywords
    ...    Run With Kubeconfig    oc delete route ms-pass ms-edge -n ${NAMESPACE}    allow_fail=True
    ...    AND
    ...    Teardown Web Server Signed

HTTP And Reencrypt Routes
    [Documentation]    Verify creation of HTTP and reencrypt route types. Reencrypt routes
    ...    terminate client TLS at the router and re-encrypt to the backend.
    ...    Covers QE test 60283.
    [Setup]    Deploy Web Server Signed

    ${router_ip}=    Get Router Pod IP

    # HTTP route
    Run With Kubeconfig
    ...    oc expose svc service-unsecure --name=ms-http --hostname=route-http.example.com -n ${NAMESPACE}
    Wait Until Keyword Succeeds    10x    2s
    ...    Verify Route Is Admitted    ms-http    ${NAMESPACE}

    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    route-http.example.com    ${router_ip}    80    scheme=http
    Verify HAProxy Backend Exists    be_http:${NAMESPACE}:ms-http

    # Reencrypt route
    Run With Kubeconfig
    ...    oc create route reencrypt ms-reen --service=service-secure --port=27443 --hostname=route-reen.example.com -n ${NAMESPACE}
    Wait Until Keyword Succeeds    10x    2s
    ...    Verify Route Is Admitted    ms-reen    ${NAMESPACE}
    ${termination}=    Oc Get JsonPath    route    ${NAMESPACE}    ms-reen    .spec.tls.termination
    Should Be Equal As Strings    ${termination}    reencrypt

    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    route-reen.example.com    ${router_ip}    443
    Verify HAProxy Backend Exists    be_secure:${NAMESPACE}:ms-reen

    [Teardown]    Run Keywords
    ...    Run With Kubeconfig    oc delete route ms-http ms-reen -n ${NAMESPACE}    allow_fail=True
    ...    AND
    ...    Teardown Web Server Signed

Reencrypt Route Via Ingress With Destination CA
    [Documentation]    Verify that an Ingress with reencrypt annotation and destination CA
    ...    certificate secret creates a working reencrypt route.
    ...    Covers QE test 60136.
    [Setup]    Deploy Web Server Signed

    Oc Create    -f ${INGRESS_DESTCA} -n ${NAMESPACE}
    ${route_name}=    Wait For Ingress Route    ingress-reencrypt    ${NAMESPACE}

    Verify Route Is Admitted    ${route_name}    ${NAMESPACE}
    ${termination}=    Oc Get JsonPath    route    ${NAMESPACE}    ${route_name}    .spec.tls.termination
    Should Be Equal As Strings    ${termination}    reencrypt

    ${router_ip}=    Get Router Pod IP
    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    service-secure-test.example.com    ${router_ip}    443
    Verify HAProxy Backend Exists    be_secure:${NAMESPACE}:

    [Teardown]    Run Keywords
    ...    Run With Kubeconfig    oc delete -f ${INGRESS_DESTCA} -n ${NAMESPACE}    allow_fail=True
    ...    AND
    ...    Teardown Web Server Signed

Router As LoadBalancer Service
    [Documentation]    Verify that the router is exposed as a LoadBalancer service and that
    ...    all 4 route types (HTTP, edge, passthrough, reencrypt) are accessible via
    ...    the LoadBalancer IP.
    ...    Covers QE test 73152.
    [Setup]    Deploy Web Server Signed

    # Verify router service is type LoadBalancer
    ${svc_type}=    Oc Get JsonPath    service    openshift-ingress    router-default    .spec.type
    Should Be Equal As Strings    ${svc_type}    LoadBalancer

    ${lb_ip}=    Get Router LB IP

    # Create all 4 route types
    Run With Kubeconfig
    ...    oc expose svc service-unsecure --name=rt-http --hostname=rt-http.example.com -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc create route edge rt-edge --service=service-unsecure --port=27017 --hostname=rt-edge.example.com -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc create route passthrough rt-pass --service=service-secure --port=27443 --hostname=rt-pass.example.com -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc create route reencrypt rt-reen --service=service-secure --port=27443 --hostname=rt-reen.example.com -n ${NAMESPACE}

    # Wait for all routes to be admitted
    FOR    ${route}    IN    rt-http    rt-edge    rt-pass    rt-reen
        Wait Until Keyword Succeeds    10x    2s
        ...    Verify Route Is Admitted    ${route}    ${NAMESPACE}
    END

    # Curl all routes via LB IP
    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    rt-http.example.com    ${lb_ip}    80    scheme=http
    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    rt-edge.example.com    ${lb_ip}    443
    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    rt-pass.example.com    ${lb_ip}    443
    Wait Until Keyword Succeeds    30x    2s
    ...    Curl Route Should Succeed    rt-reen.example.com    ${lb_ip}    443

    [Teardown]    Run Keywords
    ...    Run With Kubeconfig    oc delete route rt-http rt-edge rt-pass rt-reen -n ${NAMESPACE}    allow_fail=True
    ...    AND
    ...    Teardown Web Server Signed


*** Keywords ***
Setup
    [Documentation]    Suite-level setup.
    Setup Suite With Namespace

Teardown
    [Documentation]    Suite-level teardown.
    Teardown Suite With Namespace

Deploy Web Server Unsecure
    [Documentation]    Deploy the nginx web server without TLS.
    Oc Create    -f ${WEB_SERVER_DEPLOY} -n ${NAMESPACE}
    Named Deployment Should Be Available    web-server-deploy    ${NAMESPACE}    5m

Deploy Web Server Signed
    [Documentation]    Deploy the nginx web server with TLS via service-ca annotation.
    Oc Create    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE}
    Named Deployment Should Be Available    web-server-deploy    ${NAMESPACE}    5m

Deploy Web Server Unsecure And Signed
    [Documentation]    Deploy the TLS web server (which also serves HTTP on 8080).
    Deploy Web Server Signed

Teardown Web Server Unsecure
    [Documentation]    Delete the unsecure web server deployment.
    Run With Kubeconfig    oc delete -f ${WEB_SERVER_DEPLOY} -n ${NAMESPACE}    allow_fail=True

Teardown Web Server Signed
    [Documentation]    Delete the signed web server deployment and associated resources.
    Run With Kubeconfig    oc delete -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${NAMESPACE}    allow_fail=True

Wait For Ingress Route
    [Documentation]    Wait for the auto-generated route from an Ingress to appear.
    ...    Returns the route name.
    [Arguments]    ${ingress_name}    ${ns}
    Wait Until Keyword Succeeds    30x    2s
    ...    Ingress Should Have Route    ${ingress_name}    ${ns}
    ${route_name}=    Run With Kubeconfig
    ...    oc get route -n ${ns} -o jsonpath='{.items[?(@.metadata.ownerReferences[0].name=="${ingress_name}")].metadata.name}'
    Should Not Be Empty    ${route_name}
    RETURN    ${route_name}

Ingress Should Have Route
    [Documentation]    Assert that the ingress has an associated route.
    [Arguments]    ${ingress_name}    ${ns}
    ${route_name}=    Run With Kubeconfig
    ...    oc get route -n ${ns} -o jsonpath='{.items[?(@.metadata.ownerReferences[0].name=="${ingress_name}")].metadata.name}'
    Should Not Be Empty    ${route_name}
