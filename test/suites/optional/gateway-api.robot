*** Settings ***
Documentation       Test Gateway API functionality

Resource            ../../resources/microshift-network.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           optional    gateway-api


*** Variables ***
${NS_GATEWAY}                   ${EMPTY}
${GATEWAY_MANIFEST_TMPL}        ./assets/gateway-api/gateway.yaml.template
${HTTP_ROUTE_MANIFEST_TMPL}     ./assets/gateway-api/http-route.yaml.template
${GATEWAY_HOSTNAME}             gw1.microshift.local
${GATEWAY_PORT}                 9000


*** Test Cases ***
Test Simple HTTP Route
    [Documentation]    Create a gateway and test it with Hello MicroShift application with HTTPRoute.
    [Setup]    Run Keywords
    ...    Setup Namespace
    ...    Deploy Hello MicroShift
    Create Gateway    ${GATEWAY_HOSTNAME}    ${GATEWAY_PORT}    ${NS_GATEWAY}
    ${gateway_ip}    Create HTTP Route    ${GATEWAY_HOSTNAME}    ${NS_GATEWAY}
    Wait Until Keyword Succeeds
    ...    20x
    ...    6s
    ...    Access Hello MicroShift Success
    ...    ushift_port=${GATEWAY_PORT}
    ...    hostname=${GATEWAY_HOSTNAME}
    ...    ushift_ip=${gateway_ip}
    [Teardown]    Run Keywords
    ...    Delete Namespace


*** Keywords ***
Deploy Hello MicroShift
    [Documentation]    Deploys the hello microshift application (service included)
    ...    in the given namespace.
    Create Hello MicroShift Pod    ${NS_GATEWAY}
    Expose Hello MicroShift    ${NS_GATEWAY}

Setup Namespace
    [Documentation]    Configure a namespace where to create all resources for later cleanup.
    VAR    ${NS_GATEWAY}    ${NAMESPACE}-gw-1    scope=SUITE
    Create Namespace    ${NS_GATEWAY}

Delete Namespace
    [Documentation]    Remove gateway api specific namespace.
    Remove Namespace    ${NS_GATEWAY}

Create Gateway
    [Documentation]    Create a gateway using given hostname and port. Waits for readiness
    [Arguments]    ${hostname}    ${port}    ${namespace}
    VAR    ${tmp}    /tmp/gateway.yaml
    VAR    ${HOSTNAME}    ${hostname}    scope=TEST
    VAR    ${PORT}    ${port}    scope=TEST
    Run Keyword And Ignore Error
    ...    Remove File    ${tmp}
    Generate File From Template    ${GATEWAY_MANIFEST_TMPL}    ${tmp}
    Oc Apply    -n ${namespace} -f ${tmp}
    Oc Wait    -n ${namespace} gateway/test-gateway    --for="condition=Accepted" --timeout=${DEFAULT_WAIT_TIMEOUT}
    # Run healthcheck command to avoid premature wait returns if the object does not yet exist
    Command Should Work
    ...    sudo microshift healthcheck --namespace ${namespace} --deployments test-gateway-openshift-gateway-api --timeout 120s

Create HTTP Route
    [Documentation]    Create an HTTP route using the given hostname and namespace. Waits for acceptance in a gateway.
    ...    Returns the gateway service external IP.
    [Arguments]    ${hostname}    ${namespace}
    VAR    ${tmp}    /tmp/route.yaml
    VAR    ${HOSTNAME}    ${hostname}    scope=TEST
    VAR    ${NS}    ${namespace}    scope=TEST
    Run Keyword And Ignore Error
    ...    Remove File    ${tmp}
    Generate File From Template    ${HTTP_ROUTE_MANIFEST_TMPL}    ${tmp}
    Oc Apply    -n ${namespace} -f ${tmp}
    ${gateway_ip}    Wait For HTTPRoute And Gateway Ready    ${namespace}
    RETURN    ${gateway_ip}

Wait For HTTPRoute And Gateway Ready
    [Documentation]    Wait for HTTPRoute acceptance and gateway readiness. Returns gateway service IP.
    [Arguments]    ${namespace}
    Wait Until Keyword Succeeds    20x    6s
    ...    Verify HTTPRoute Parent Accepted    ${namespace}
    Wait Until Keyword Succeeds    20x    6s
    ...    Verify HTTPRoute References Resolved    ${namespace}
    Wait Until Keyword Succeeds    20x    6s
    ...    Verify Gateway Programmed    ${namespace}
    ${gateway_ip}    Wait Until Keyword Succeeds    20x    6s
    ...    Verify Gateway Service Has External IP    ${namespace}
    RETURN    ${gateway_ip}

Verify HTTPRoute Parent Accepted
    [Documentation]    Verify that the HTTPRoute is accepted by its parent gateway
    [Arguments]    ${namespace}
    ${result}    Run With Kubeconfig
    ...    oc get httproutes/http -n ${namespace} -o jsonpath='{.status.parents[*].conditions[?(@.type=="Accepted")].status}'
    Should Not Be Empty    ${result}    HTTPRoute parent conditions not found
    Should Not Contain    ${result}    False    HTTPRoute not accepted by parent gateway
    Should Not Contain    ${result}    Unknown    HTTPRoute acceptance status is unknown

Verify HTTPRoute References Resolved
    [Documentation]    Verify that all references in the HTTPRoute are resolved
    [Arguments]    ${namespace}
    ${result}    Run With Kubeconfig
    ...    oc get httproutes/http -n ${namespace} -o jsonpath='{.status.parents[*].conditions[?(@.type=="ResolvedRefs")].status}'
    Should Not Be Empty    ${result}    HTTPRoute reference conditions not found
    Should Not Contain    ${result}    False    HTTPRoute references not resolved
    Should Not Contain    ${result}    Unknown    HTTPRoute reference resolution status is unknown

Verify Gateway Programmed
    [Documentation]    Verify that the gateway data plane is programmed and ready to receive traffic
    [Arguments]    ${namespace}
    ${result}    Run With Kubeconfig
    ...    oc get gateway/test-gateway -n ${namespace} -o jsonpath='{.status.conditions[?(@.type=="Programmed")].status}'
    Should Be Equal As Strings    ${result}    True    Gateway is not yet programmed

Verify Gateway Service Has External IP
    [Documentation]    Verify that the gateway's LoadBalancer service has an external IP assigned and return it
    [Arguments]    ${namespace}
    ${result}    Run With Kubeconfig
    ...    oc get svc test-gateway-openshift-gateway-api -n ${namespace} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    Should Not Be Empty    ${result}    Gateway service does not have an external IP assigned
    RETURN    ${result}

Generate File From Template
    [Documentation]    Generate file from template
    [Arguments]    ${template_file}    ${out_file}
    ${template}    OperatingSystem.Get File    ${template_file}
    ${message}    Replace Variables    ${template}
    OperatingSystem.Append To File    ${out_file}    ${message}
