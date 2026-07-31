*** Settings ***
Documentation       Router access logging configuration tests (disruptive)
...                 Migrated from openshift-tests-private:
...                 OCP-81996, OCP-81997, OCP-82000, OCP-82003, OCP-82004

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/router.resource
Variables           configs.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           restart    slow


*** Variables ***
${BASE_DOMAIN}      apps.example.com


*** Test Cases ***
HTTP Capture Cookies Prefix Match
    [Documentation]    Verify httpCaptureCookies with Prefix match captures cookies matching the
    ...    prefix in router logs across HTTP, edge, and reencrypt routes.
    ...    OCP-81996

    Setup Router Config And Restart    ${CONFIG_COOKIE_PREFIX}

    Deploy Web Server Signed
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec81996.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge81996.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen81996.${BASE_DOMAIN}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Create OC Route And Admit    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${edge_host}
    Create OC Route And Admit    ${NAMESPACE}    reencrypt    route-reen    service-secure    --hostname=${reen_host}
    ${router_ip}=    Get Router Pod IP
    Curl All Cookie Routes And Verify Logs    ${routehost}    ${edge_host}    ${reen_host}    ${router_ip}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

HTTP Capture Cookies Exact Match And MaxLength
    [Documentation]    Verify httpCaptureCookies with Exact match captures only exact-named cookies,
    ...    and maxLength truncates cookie values in logs.
    ...    OCP-81997

    Setup Router Config And Restart    ${CONFIG_COOKIE_EXACT_100}
    Deploy Web Server
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec81997.${BASE_DOMAIN}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    fooor=nobar
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo=bar
    Wait For Router Logs To Contain    foo=bar
    Router Logs Should Not Contain    fooor=nobar

    Setup Router Config And Restart    ${CONFIG_COOKIE_EXACT_10}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo=bar89abdef
    Wait For Router Logs To Contain    foo=bar89a
    Router Logs Should Not Contain    foo=bar89ab
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

HTTP Capture Headers Request And Response
    [Documentation]    Verify httpCaptureHeaders captures request Host and response Server headers
    ...    in router logs, including for edge and reencrypt routes.
    ...    OCP-82000

    Setup Router Config And Restart    ${CONFIG_CAPTURE_HEADERS_120}

    Deploy Web Server Signed
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec82000.${BASE_DOMAIN}
    VAR    ${edge_host}=    route-edge82000.${BASE_DOMAIN}
    VAR    ${reen_host}=    route-reen82000.${BASE_DOMAIN}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Create OC Route And Admit    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${edge_host}
    Create OC Route And Admit    ${NAMESPACE}    reencrypt    route-reen    service-secure    --hostname=${reen_host}
    Verify Header Capture Config And Logs    ${routehost}    ${edge_host}    ${reen_host}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

HTTP Capture Headers MaxLength Adherence
    [Documentation]    Verify httpCaptureHeaders maxLength truncates captured header values in logs.
    ...    OCP-82003

    VAR    ${routehost}=    route-unsec82003.${BASE_DOMAIN}
    Setup Router Config And Restart    ${CONFIG_CAPTURE_HEADERS_MAXLEN}

    Deploy Web Server
    Deploy Test Client Pod
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    http://${routehost}/index.html    ${routehost}:80:${router_ip}
    # maxLength is 16, so hostname is truncated to exactly 16 chars: route-unsec82003
    Wait For Router Logs To Contain    route-unsec82003
    ${logs}=    Get Router Access Logs
    Should Not Contain    ${logs}    ${routehost}
    # nginx server version is 5+ chars, so only "nginx" should appear without the version
    Should Contain    ${logs}    nginx
    Should Not Contain    ${logs}    nginx/
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Cleanup Test Workloads

Custom HTTP Error Pages
    [Documentation]    Verify custom 503 and 404 error pages are served when configured via
    ...    httpErrorCodePages configmap.
    ...    OCP-82004

    Create Configmap From Files    ${ROUTER_NS}    custom-82004-error-code-pages
    ...    --from-file=./assets/router/error-page-503.http    --from-file=./assets/router/error-page-404.http
    Setup Router Config And Restart    ${CONFIG_ERROR_PAGES}

    Deploy Web Server
    Deploy Test Client Pod
    VAR    ${routehost}=    route-unsec82004.${BASE_DOMAIN}
    VAR    ${noexist_host}=    not-exist82004.${BASE_DOMAIN}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${routehost}
    Verify Custom Error Pages    ${routehost}    ${noexist_host}
    [Teardown]    Run Keywords
    ...    Remove Router Config And Restart
    ...    AND    Run With Kubeconfig    oc delete configmap custom-82004-error-code-pages -n ${ROUTER_NS} --ignore-not-found
    ...    AND    Cleanup Test Workloads


*** Keywords ***
Curl All Cookie Routes And Verify Logs
    [Documentation]    Curl all routes with cookies and verify correct log entries.
    [Arguments]    ${routehost}    ${edge_host}    ${reen_host}    ${router_ip}
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    fo=nobar
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo=bar
    Wait Until Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    http://${routehost}/index.html
    ...    ${routehost}:80:${router_ip}
    ...    foo22=bar22
    Wait Until HTTPS Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${edge_host}/index.html
    ...    ${edge_host}:443:${router_ip}
    ...    foo=barforedge
    Wait Until HTTPS Curl With Cookie Succeeds From Pod
    ...    ${CLIENT_POD_NAME}
    ...    ${NAMESPACE}
    ...    https://${reen_host}/index.html
    ...    ${reen_host}:443:${router_ip}
    ...    foo=barforreen
    Wait For Router Logs To Contain    foo=bar
    Wait For Router Logs To Contain    foo22=bar22
    Wait For Router Logs To Contain    foo=barforedge
    Wait For Router Logs To Contain    foo=barforreen
    Router Logs Should Not Contain    fo=nobar

Verify Header Capture Config And Logs
    [Documentation]    Verify haproxy header capture config and check captured headers in logs.
    [Arguments]    ${routehost}    ${edge_host}    ${reen_host}
    Verify Header Capture In Haproxy
    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    http://${routehost}/index.html    ${routehost}:80:${router_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${edge_host}/index.html    ${edge_host}:443:${router_ip}
    Wait Until HTTPS Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}    https://${reen_host}/index.html    ${reen_host}:443:${router_ip}
    Wait For Router Logs To Contain    ${routehost}
    Wait For Router Logs To Contain    ${edge_host}
    Wait For Router Logs To Contain    ${reen_host}

Verify Header Capture In Haproxy
    [Documentation]    Verify the haproxy frontend has the expected header capture configuration.
    ${router_pod}=    Get Router Pod Name
    ${haproxy}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- grep -A 20 "frontend fe_sni" haproxy.config
    Should Contain    ${haproxy}    capture request header Host len 120
    Should Contain    ${haproxy}    capture response header Server len 120

Verify Custom Error Pages
    [Documentation]    Verify custom 503 and 404 error pages are served by the router.
    [Arguments]    ${routehost}    ${noexist_host}
    ${router_pod}=    Get Router Pod Name
    ${error503}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- cat /var/lib/haproxy/errorfiles/error-page-503.http
    Should Contain    ${error503}    Custom:Application Unavailable
    ${error404}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${router_pod} -- cat /var/lib/haproxy/errorfiles/error-page-404.http
    Should Contain    ${error404}    Custom:Not Found
    ${router_ip}=    Get Router Pod IP
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} ${CLIENT_POD_NAME} -- curl http://${noexist_host} -s --resolve ${noexist_host}:80:${router_ip} --connect-timeout 10
    Should Contain    ${output}    Custom:Not Found
    Scale Deployment    ${NAMESPACE}    web-server-deploy    0
    Wait Until Keyword Succeeds    60s    2s
    ...    Custom 503 Body Should Contain    ${routehost}    ${router_ip}

Custom 503 Body Should Contain
    [Documentation]    Curl a route (GET, full response body) and assert the custom 503 error page body is returned.
    [Arguments]    ${routehost}    ${router_ip}
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} ${CLIENT_POD_NAME} -- curl http://${routehost} -s --resolve ${routehost}:80:${router_ip} --connect-timeout 10
    Should Contain    ${output}    Custom:Application Unavailable

Router Logs Should Not Contain
    [Documentation]    Verify the router access logs do NOT contain a given pattern.
    [Arguments]    ${pattern}
    ${logs}=    Get Router Access Logs
    Should Not Contain    ${logs}    ${pattern}
