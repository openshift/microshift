*** Settings ***
Documentation       Keywords for using the oc command line.
...
...                 Requires that the openshift-clients package be installed and
...                 'oc' is in the $PATH.

Resource            common.resource
Library             DataFormats.py


*** Variables ***
${DEFAULT_WAIT_TIMEOUT}     60s


*** Keywords ***
Oc Get
    [Documentation]    Run 'oc get' for a specific instance of a type in a namespace.
    ...    Returns the YAML output parsed to a DottedDict for use in
    ...    other keywords.
    [Arguments]    ${type}    ${namespace}    ${resource}

    ${yaml_text}=    Run With Kubeconfig    oc get -n ${namespace} -o yaml ${type} ${resource}
    ${yaml_data}=    Yaml Parse    ${yaml_text}

    RETURN    ${yaml_data}

Oc Get JsonPath
    [Documentation]    Run 'oc get' for a specific instance of a type in a namespace
    ...    with jsonpath provided. Returns output of the command without any parsing.
    [Arguments]    ${type}    ${namespace}    ${resource}    ${jsonpath}

    IF    "${namespace}"=="${EMPTY}"
        ${namespace_arg}=    Set Variable    -A
    ELSE
        ${namespace_arg}=    Set Variable    -n ${namespace}
    END

    ${text}=    Run With Kubeconfig    oc get ${namespace_arg} -o=jsonpath='{ ${jsonpath} }' ${type} ${resource}

    RETURN    ${text}

Oc Apply
    [Documentation]    Run 'oc apply' on a specific pod in the curret test namespace
    ...    Returns the command's combined STDOUT/STDER
    [Arguments]    ${opts}
    ${output}=    Run With Kubeconfig    oc apply ${opts}
    RETURN    ${output}

Oc Delete
    [Documentation]    Run 'oc delete' on a specific pod in the curret test namespace
    ...    Returns the command's combined STDOUT/STDER
    [Arguments]    ${opts}
    ${output}=    Run With Kubeconfig    oc delete ${opts}
    RETURN    ${output}

Oc Exec
    [Documentation]    Run 'oc exec' on a specific pod in the curret test namespace
    ...    Returns the command's combined STDOUT/STDER
    [Arguments]    ${pod}    ${cmd}    ${ns}=${NAMESPACE}
    ${output}=    Run With Kubeconfig    oc exec -n ${ns} pod/${pod} -- /bin/bash -c '${cmd}'
    RETURN    ${output}

Oc Wait
    [Documentation]    Run 'oc wait' on a specific resource.
    ...    'resource' should contain: "-n namespace" (if applicable), name, and kind.
    ...    For example: "-n test-ns pod test-pod".
    ...    'opts' should contain a condition and additional parameters like timeout.
    ...    Returns the combined STDOUT/STDERR output of the command.
    [Arguments]    ${resource}    ${opts}
    ${output}    ${rc}=    Run With Kubeconfig
    ...    oc wait ${resource} ${opts} --timeout=${timeout}
    ...    allow_fail=${TRUE}
    ...    return_rc=${TRUE}
    IF    ${rc} != 0
        Run With Kubeconfig    oc describe ${resource}
        ...    allow_fail=${TRUE}
        ...    timeout=${timeout}
        Log    'oc wait ${resource} ${opts}' failed
        Should Be Equal As Integers    ${rc}    0
    END
    RETURN    ${output}

Named Pod Should Be Ready
    [Documentation]    Wait for pod with name \${name} to become "Ready"
    ...    ${name}    Name of pod to wait for
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pod to reach ready status. Default 60s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}

    Oc Wait    -n ${ns} pod/${name}    --for="condition=Ready" --timeout=${timeout}

Named Pod Should Be Deleted
    [Documentation]    Wait for pod with ${name} to be deleted
    ...    ${name}    Name of pod to wait for
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pod to reach ready status. Default 60s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} pod/${name}    --for=delete --timeout=${timeout}

Labeled Pod Should Be Ready
    [Documentation]    Wait for pod(s) ready by ${label} to become "Ready"
    ...    ${label}    A label selector value to match by (e.g. "app\=my_app"). Note that '=' must be escaped with '\'.
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pod to reach ready status. Default 60s.
    [Arguments]    ${label}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} pod --selector="${label}"    --for=condition=Ready --timeout=${timeout}

Named Deployment Should Be Available
    [Documentation]    Wait for a given deployment's Available condition to be true
    ...    ${name}    Name of deployment to wait for
    ...    ${ns}    Namespace of named deployment. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for deployment to reach ready status. Default 60s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} deploy ${name}    --for=condition=Available --timeout=${timeout}

Named Daemonset Should Be Available
    [Documentation]    Wait for a given daemonset's Available condition to be true
    ...    ${name}    Name of daemonset to wait for
    ...    ${ns}    Namespace of named daemonset. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for daemonset to reach ready status. Default 60s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} daemonset ${name}    --for=jsonpath='{.status.numberAvailable}'=1 --timeout=${timeout}

Wait Until Resource Exists
    [Documentation]    Waits for a resource to exist.
    [Arguments]    ${type}    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Wait Until Keyword Succeeds    ${timeout}    5s
    ...    Oc Get    ${type}    ${ns}    ${name}

Named PVC Should Be Resized
    [Documentation]    Wait for pvc with ${name} to resize to ${to_size}
    ...    ${name}    Name of pvc to wait for
    ...    ${ns}    Namespace of named pvc. Defaults to NAMESPACE suite variable
    ...    ${to_size}    Size pvc is expected to be updated to
    ...    ${timeout}    Period of time to wait for pvc to resize to new size. Default 60s.
    [Arguments]    ${name}    ${to_size}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} pvc/${name}
    ...    --for=jsonpath="{.spec.resources.requests.storage}"=${to_size} --timeout=${timeout}

Named PVC Should Be Deleted
    [Documentation]    Wait for pvc with ${name} to be deleted
    ...    ${name}    Name of pvc to wait for
    ...    ${ns}    Namespace of named pvc. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pvc to be deleted. Default 60s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    ${output}=    Oc Wait    -n ${ns} pvc/${name}    --for=Delete --timeout=${timeout}
    RETURN    ${output}

Named VolumeSnapshot Should Be Ready
    [Documentation]    Wait for a volumesnap ${name} to become "readyToUse"
    ...    ${name}    Name of volumesnapshot to wait for
    ...    ${timeout}    Period of time to wait for volumeSnapshot ready "readyToUse" state. Default 60s.
    [Arguments]    ${name}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    volumesnapshot/${name} -n ${NAMESPACE}    --for=jsonpath='{.status.readyToUse}=true'

Named VolumeSnapshot Should Be Deleted
    [Documentation]    Wait for volumesnapshot with ${name} to be deleted
    ...    ${name}    Name of volumesnapshot to wait for
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable, specified by
    ...    ${timeout}    Period of time to wait for volumesnapshot to reach ready status. Default 60s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    Oc Wait    -n ${ns} volumesnapshot/${name}    --for=Delete --timeout=${timeout}

Oc Create
    [Documentation]    Run 'oc create' on a specific resource in the current test namespace
    ...    Returns the combined STDOUT/STDERR output of the command.
    [Arguments]    ${opts}
    ${output}=    Run With Kubeconfig    oc create ${opts}
    RETURN    ${output}

Oc Expose
    [Documentation]    Run 'oc expose' on a specific resource in the current test namespace
    ...    Returns the combined STDOUT/STDERR output of the command.
    [Arguments]    ${opts}
    ${output}=    Run With Kubeconfig    oc expose ${opts}
    RETURN    ${output}

Oc Patch
    [Documentation]    Run 'oc patch' on a specific resource in the current test namespace
    ...    Returns the combined STDOUT/STDER output of the command.
    [Arguments]    ${type_name}    ${opts}
    ${output}=    Run With Kubeconfig    oc patch -n ${NAMESPACE} ${type_name} -p ${opts}
    RETURN    ${output}

Oc Logs
    [Documentation]    Run 'oc logs' on a specific resource in the given namespace.
    ...    Returns the combined STDOUT/STDER output of the command.
    [Arguments]    ${opts}    ${namespace}
    ${output}=    Run With Kubeconfig    oc logs -n ${namespace} ${opts}
    RETURN    ${output}

All Pods Should Be Running
    [Documentation]    Check if all pods are running in the given namespace
    [Arguments]    ${ns}=-A    ${timeout}=${DEFAULT_WAIT_TIMEOUT}
    IF    "${ns}"=="-A"
        Oc Wait    pods --all -A    --for="condition=Ready" --timeout=${timeout}
    ELSE
        Oc Wait    pods --all -n ${ns}    --for="condition=Ready" --timeout=${timeout}
    END
