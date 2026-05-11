*** Settings ***
Documentation       Tests for SCC admission controller enforcement

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${POD_SC_1000}      ./assets/scc/pod-security-context.yaml
${POD_SC_2000}      ./assets/scc/pod-run-as-user-2000.yaml
${POD_NO_USER}      ./assets/scc/pod-no-user.yaml


*** Test Cases ***
Pod Runs With Specified UID And Groups
    [Documentation]    Verify a pod runs with the specified runAsUser, fsGroup,
    ...    and supplementalGroups from its security context.
    Oc Create    -f ${POD_SC_1000} -n ${NAMESPACE}
    Named Pod Should Be Ready    scc-test-pod-1    ${NAMESPACE}
    ${id_output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} pod/scc-test-pod-1 -- id
    Should Match Regexp    ${id_output}    uid=1000
    Should Match Regexp    ${id_output}    groups=.*2000
    Should Match Regexp    ${id_output}    groups=.*3000

    [Teardown]    Oc Delete    -f ${POD_SC_1000} -n ${NAMESPACE} --ignore-not-found

Pod Runs With Different Specified UID
    [Documentation]    Verify a pod with runAsUser=2000 runs as that UID.
    Oc Create    -f ${POD_SC_2000} -n ${NAMESPACE}
    Named Pod Should Be Ready    scc-test-pod-2    ${NAMESPACE}
    ${id_output}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} pod/scc-test-pod-2 -- id -u
    Should Be Equal As Strings    ${id_output.strip()}    2000

    [Teardown]    Oc Delete    -f ${POD_SC_2000} -n ${NAMESPACE} --ignore-not-found

Different Namespaces Get Different SCC Ranges
    [Documentation]    Pods in different namespaces without explicit runAsUser should
    ...    receive different allocated UID ranges from the SCC admission controller.
    ${ns2}=    Create Unique Namespace
    Oc Create    -f ${POD_NO_USER} -n ${NAMESPACE}
    Oc Create    -f ${POD_NO_USER} -n ${ns2}
    Named Pod Should Be Ready    scc-test-pod    ${NAMESPACE}
    Named Pod Should Be Ready    scc-test-pod    ${ns2}
    ${id1}=    Run With Kubeconfig
    ...    oc exec -n ${NAMESPACE} pod/scc-test-pod -- id -u
    ${id2}=    Run With Kubeconfig
    ...    oc exec -n ${ns2} pod/scc-test-pod -- id -u
    Should Not Be Equal As Strings    ${id1.strip()}    ${id2.strip()}

    [Teardown]    Run Keywords
    ...    Oc Delete    -f ${POD_NO_USER} -n ${NAMESPACE} --ignore-not-found
    ...    AND
    ...    Oc Delete    -f ${POD_NO_USER} -n ${ns2} --ignore-not-found
    ...    AND
    ...    Remove Namespace    ${ns2}
