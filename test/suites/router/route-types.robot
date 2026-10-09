*** Settings ***
Documentation       HTTP route created from a Kubernetes Ingress resource.
...                 Migrated from openshift-tests-private: OCP-60149.
...                 The edge/passthrough/reencrypt route types (OCP-60266, 60283, 60136, 73152)
...                 are already covered by suites/router/router-routes.robot.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/router.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Test Cases ***
HTTP Route Via Ingress
    [Documentation]    Verify that a Kubernetes Ingress resource is reconciled into an admitted
    ...    HTTP route that forwards traffic to the backend service.
    ...    OCP-60149
    [Setup]    Run Keywords
    ...    Deploy Web Server
    ...    AND    Deploy Test Client Pod

    Oc Create    -f ${INGRESS_HTTP} -n ${NAMESPACE}
    Ingress Route Should Be Admitted    ingress-on-microshift    timeout=300s

    ${router_ip}=    Get Router Pod IP
    Wait Until Curl Succeeds From Pod
    ...    ${CLIENT_POD_NAME}    ${NAMESPACE}
    ...    http://service-unsecure-test.example.com:80
    ...    service-unsecure-test.example.com:80:${router_ip}

    ${haproxy}=    Read Haproxy Config
    Should Contain    ${haproxy}    be_http:${NAMESPACE}:

    [Teardown]    Run Keywords
    ...    Oc Delete    -f ${INGRESS_HTTP} -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${WEB_SERVER_DEPLOY} -n ${NAMESPACE} --ignore-not-found
    ...    AND    Oc Delete    -f ${TEST_CLIENT_POD} -n ${NAMESPACE} --ignore-not-found
