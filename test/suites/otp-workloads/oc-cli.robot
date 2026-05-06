*** Settings ***
Documentation       Tests for oc CLI behavior including help output, flag validation,
...                 events display, dry-run operations, version info, debug behavior,
...                 and oc explain validation.
...
...                 Ported from openshift-tests-private:
...                 Medium-28007, Critical-63850, Critical-64919, High-63855,
...                 Medium-64944, High-43030, Medium-34155, Medium-47555,
...                 Medium-49116, Medium-66724

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Library             ../../resources/DataFormats.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace


*** Variables ***
${IDMS_FILE}        ./assets/otp-workloads/idms.yaml
${ICSP_FILE}        ./assets/otp-workloads/icsp.yaml
${DEPLOY_NAME}      debug-test
${DEPLOY_IMAGE}     quay.io/microshift/busybox:1.36


*** Test Cases ***
Oc Version Shows Clean GitTreeState
    [Documentation]    Verify that oc version -o json reports
    ...    gitTreeState as clean.
    ...    OCP-28007

    ${out}=    Run With Kubeconfig    oc version -o json
    ${version_info}=    Json Parse    ${out}
    Should Be Equal As Strings    ${version_info.clientVersion.gitTreeState}    clean

Oc Image Extract Help Contains IDMS File Flag
    [Documentation]    Verify that oc image extract -h shows --idms-file
    ...    and does not show the deprecated --icsp-file flag.
    ...    OCP-63850

    ${output}=    Run With Kubeconfig    oc image extract -h
    Should Contain    ${output}    --idms-file
    Should Not Contain    ${output}    --icsp-file

Oc Adm Release Info Help Contains IDMS File Flag
    [Documentation]    Verify that oc adm release info -h shows --idms-file
    ...    and does not show the deprecated --icsp-file flag.
    ...    OCP-64919

    ${output}=    Run With Kubeconfig    oc adm release info -h
    Should Contain    ${output}    --idms-file
    Should Not Contain    ${output}    --icsp-file

Oc Image Extract Rejects Both ICSP And IDMS Flags
    [Documentation]    Verify that oc image extract errors when both
    ...    --icsp-file and --idms-file flags are used simultaneously.
    ...    OCP-63855

    ${output}    ${rc}=    Run With Kubeconfig
    ...    oc image extract quay.io/example/test:latest --idms-file\=${IDMS_FILE} --icsp-file\=${ICSP_FILE} --confirm
    ...    allow_fail=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0
    Should Contain    ${output}    icsp-file and idms-file are mutually exclusive

Oc Adm Release Info Rejects Both ICSP And IDMS Flags
    [Documentation]    Verify that oc adm release info errors when both
    ...    --icsp-file and --idms-file flags are used simultaneously.
    ...    OCP-64944

    ${output}    ${rc}=    Run With Kubeconfig
    ...    oc adm release info quay.io/example/test:latest --idms-file\=${IDMS_FILE} --icsp-file\=${ICSP_FILE}
    ...    allow_fail=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0
    Should Contain    ${output}    icsp-file and idms-file are mutually exclusive

Oc Get Events Shows Timestamps Not Unknown
    [Documentation]    Verify that oc get events shows timestamps as
    ...    LAST SEEN and none of the values are unknown.
    ...    OCP-43030

    ${namespaces}=    Run With Kubeconfig
    ...    oc get ns -o=custom-columns=NAME:.metadata.name --no-headers
    @{ns_list}=    Split String    ${namespaces}
    FOR    ${ns}    IN    @{ns_list}
        ${output}    ${rc}=    Run With Kubeconfig
        ...    oc get events -n ${ns}
        ...    allow_fail=True    return_rc=True
        IF    ${rc} != 0    CONTINUE
        ${has_events}=    Run Keyword And Return Status
        ...    Should Not Contain    ${output}    No resources found
        IF    ${has_events}
            Should Contain    ${output}    LAST SEEN
            Should Not Match Regexp    ${output}    (?m)^unknown\\s
        END
    END

Oc Get Events Sorted By LastTimestamp
    [Documentation]    Verify that oc get events --sort-by=.lastTimestamp
    ...    executes without error.
    ...    OCP-34155

    Run With Kubeconfig    oc get events -A --sort-by\=.lastTimestamp

Oc Set Data Dry Run Server Does Not Persist ConfigMap
    [Documentation]    Verify that oc set data with --dry-run=server does not
    ...    persistently update a ConfigMap or Secret.
    ...    OCP-47555

    Run With Kubeconfig    oc create configmap cm-47555 --from-literal\=name\=abc -n ${NAMESPACE}
    ${before}=    Run With Kubeconfig    oc get cm cm-47555 -o\=jsonpath\='{.data.name}' -n ${NAMESPACE}
    Run With Kubeconfig    oc set data cm cm-47555 --from-literal\=name\=def --dry-run\=server -n ${NAMESPACE}
    ${after}=    Run With Kubeconfig    oc get cm cm-47555 -o\=jsonpath\='{.data.name}' -n ${NAMESPACE}
    Should Be Equal As Strings    ${before}    ${after}

    Run With Kubeconfig    oc create secret generic secret-47555 --from-literal\=name\=abc -n ${NAMESPACE}
    ${before_secret}=    Run With Kubeconfig
    ...    oc get secret secret-47555 -o\=jsonpath\='{.data.name}' -n ${NAMESPACE}
    Run With Kubeconfig
    ...    oc set data secret secret-47555 --from-literal\=name\=def --dry-run\=server -n ${NAMESPACE}
    ${after_secret}=    Run With Kubeconfig
    ...    oc get secret secret-47555 -o\=jsonpath\='{.data.name}' -n ${NAMESPACE}
    Should Be Equal As Strings    ${before_secret}    ${after_secret}

Oc Debug Removes StartupProbe From Debug Pod
    [Documentation]    Verify that oc debug removes startupProbe when
    ...    creating a debug pod from a deployment.
    ...    OCP-49116

    Run With Kubeconfig
    ...    oc create deploy ${DEPLOY_NAME} --image\=${DEPLOY_IMAGE} -n ${NAMESPACE}

    VAR    ${patch}=
    ...    [{"op": "add", "path": "/spec/template/spec/containers/0/startupProbe", "value":{"exec": {"command": ["false"]}}}]
    Run With Kubeconfig
    ...    oc patch deploy ${DEPLOY_NAME} --type\=json -p '${patch}' -n ${NAMESPACE}

    ${out}=    Run With Kubeconfig
    ...    oc debug deploy/${DEPLOY_NAME} -o\=jsonpath\='{.spec.containers[0].startupProbe}' -n ${NAMESPACE}
    Should Be Empty    ${out}

    [Teardown]    Run With Kubeconfig
    ...    oc delete deploy ${DEPLOY_NAME} -n ${NAMESPACE}    allow_fail=True

Oc Explain Works For All API Resources
    [Documentation]    Verify that oc explain succeeds for every
    ...    resource returned by oc api-resources.
    ...    OCP-66724

    ${api_output}=    Run With Kubeconfig
    ...    oc api-resources --no-headers | awk '{print $1}' | sort -u
    @{resources}=    Split String    ${api_output}
    FOR    ${resource}    IN    @{resources}
        ${output}    ${rc}=    Run With Kubeconfig
        ...    oc explain ${resource}
        ...    allow_fail=True    return_rc=True
        IF    ${rc} != 0
            ${not_found}=    Run Keyword And Return Status
            ...    Should Match Regexp    ${output}    couldn't find resource|not found
            IF    not ${not_found}
                Fail    oc explain ${resource} failed: ${output}
            END
        END
    END
