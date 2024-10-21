*** Settings ***
Documentation       Test Gateway API functionality

Resource            ../../resources/microshift-network.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${NS_GATEWAY_1}                 ${EMPTY}
${GATEWAY_MANIFEST_TMPL}        ./assets/gateway-api/gateway.yaml.template
${HTTP_ROUTE_MANIFEST_TMPL}     ./assets/gateway-api/http-route.yaml.template
${GATEWAY_1_HOSTNAME}           gw1.microshift.local
${GATEWAY_1_PORT}               9000


*** Test Cases ***
Test Simple HTTP Route
    [Documentation]    Create a gateway and test it with Hello MicroShift application with HTTPRoute.
    [Setup]    Run Keywords
    ...    Setup Namespaces
    ...    Deploy Hello MicroShift
    Create Gateway    hostname=${GATEWAY_1_HOSTNAME}    port=${GATEWAY_1_PORT}    namespace=${NS_GATEWAY_1}
    Create HTTP Route    hostname=${GATEWAY_1_HOSTNAME}    namespace=${NS_GATEWAY_1}
    Access Hello MicroShift Success    ushift_port=${GATEWAY_1_PORT}    hostname=${GATEWAY_1_HOSTNAME}
    [Teardown]    Run Keywords
    ...    Delete Namespaces


*** Keywords ***
Deploy Hello MicroShift
    [Documentation]    Deploys the hello microshift application (service included)
    ...    in the given namespace.
    Create Hello MicroShift Pod    ns=${NS_GATEWAY_1}
    Expose Hello MicroShift    ${NS_GATEWAY_1}

Setup Namespaces
    [Documentation]    Configure the required namespaces for namespace ownership tests.
    Set Suite Variable    \${NS_GATEWAY_1}    ${NAMESPACE}-gw-1
    Create Namespace    ${NS_GATEWAY_1}

Delete Namespaces
    [Documentation]    Remove namespace ownership namespaces.
    Remove Namespace    ${NS_GATEWAY_1}

Create Gateway
    [Documentation]    Create a gateway using given hostname and port. Waits for readiness
    [Arguments]    ${hostname}    ${port}    ${namespace}
    ${tmp}=    Set Variable    /tmp/gateway.yaml
    Set Test Variable    ${HOSTNAME}    ${hostname}
    Set Test Variable    ${PORT}    ${port}
    Run Keyword And Ignore Error
    ...    Remove File    ${tmp}
    Generate File From Template    ${GATEWAY_MANIFEST_TMPL}    ${tmp}
    Oc Apply    -n ${namespace} -f ${tmp}
    Run With Kubeconfig    oc wait -n ${namespace} gateway/test-gateway --for="condition=Accepted" --timeout=60s
    Run With Kubeconfig
    ...    oc wait -n ${namespace} deploy test-gateway-openshift-gateway-api --for=condition=Available --timeout=60s
    Run With Kubeconfig    oc wait -n ${namespace} gateway/test-gateway --for="condition=Programmed" --timeout=60s

Create HTTP Route
    [Documentation]    Create an HTTP route using the given hostname and namespace. Waits for acceptance in a gateway.
    [Arguments]    ${hostname}    ${namespace}
    ${tmp}=    Set Variable    /tmp/route.yaml
    Set Test Variable    ${HOSTNAME}    ${hostname}
    Set Test Variable    ${NS}    ${namespace}
    Run Keyword And Ignore Error
    ...    Remove File    ${tmp}
    Generate File From Template    ${HTTP_ROUTE_MANIFEST_TMPL}    ${tmp}
    Oc Apply    -n ${namespace} -f ${tmp}
    Run With Kubeconfig
    ...    oc wait -n ${namespace} httproutes/http --for jsonpath='{.status.parents[].conditions[?(@.type=="Accepted")].status}=True' --timeout=60s
    Run With Kubeconfig
    ...    oc wait -n ${namespace} httproutes/http --for jsonpath='{.status.parents[].conditions[?(@.type=="ResolvedRefs")].status}=True' --timeout=60s

Generate File From Template
    [Documentation]    Generate file from template
    [Arguments]    ${template_file}    ${output_file}
    ${template}=    OperatingSystem.Get File    ${template_file}
    ${message}=    Replace Variables    ${template}
    OperatingSystem.Append To File    ${output_file}    ${message}
