*** Settings ***
Documentation       Keywords for using the oc command line.
...
...                 Requires that the openshift-clients package be installed and
...                 'oc' is in the $PATH.

Resource            common.resource
Library             DataFormats.py


*** Keywords ***
Oc Get
    [Documentation]    Run 'oc get' for a specific instance of a type in a namespace.
    ...    Returns the YAML output parsed to a DottedDict for use in
    ...    other keywords.
    [Arguments]    ${type}    ${namespace}    ${resource}

    ${yaml_text}=    Run With Kubeconfig    oc get -n ${namespace} -o yaml ${type} ${resource}
    ${yaml_data}=    Yaml Parse    ${yaml_text}

    RETURN    ${yaml_data}

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
    [Arguments]    ${pod}    ${cmd}
    ${output}=    Run With Kubeconfig    oc exec -n ${NAMESPACE} pod/${pod} -- /bin/bash -c '${cmd}'
    RETURN    ${output}

Named Pod Should Be Ready
    [Documentation]    Wait for pod with name \${name} to become "Ready"
    ...    ${name}    Name of pod to wait for
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pod to reach ready status.    Default 30s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=30s

    Run With Kubeconfig    oc wait -n ${NAMESPACE} pod/${name} --for="condition=Ready" --timeout=${timeout}

Named Pod Should Be Deleted
    [Documentation]    Wait for pod with ${name} to be deleted
    ...    ${name}    Name of pod to wait for
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pod to reach ready status.    Default 30s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=30s
    Run With Kubeconfig    oc wait -n ${NAMESPACE} pod/${name} --for=delete --timeout=${timeout}

Labeled Pod Should Be Ready
    [Documentation]    Wait for pod(s) ready by ${label} to become "Ready"
    ...    ${label}    A label selector value to match by (e.g. "app\=my_app"). Note that '=' must be escaped with '\'.
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pod to reach ready status.    Default 30s.
    [Arguments]    ${label}    ${ns}=${NAMESPACE}    ${timeout}=30s
    Run With Kubeconfig    oc wait -n ${ns} pod --selector="${label}" --for=condition=Ready --timeout=${timeout}

Named Deployment Should Be Available
    [Documentation]    Wait for a given deployment's Available condition to be true
    ...    ${name}    Name of deployment to wait for
    ...    ${ns}    Namespace of named deployment. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for deployment to reach ready status.    Default 30s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=30s
    Run With Kubeconfig    oc wait -n ${ns} deploy ${name} --for=condition=Available --timeout=${timeout}

Named PVC Should Be Resized
    [Documentation]    Wait for pvc with ${name} to resize to ${to_size}
    ...    ${name}    Name of pvc to wait for
    ...    ${ns}    Namespace of named pvc. Defaults to NAMESPACE suite variable
    ...    ${to_size}    Size pvc is expected to be updated to
    ...    ${timeout}    Period of time to wait for pvc to resize to new size.    Default 30s.
    [Arguments]    ${name}    ${to_size}    ${ns}=${NAMESPACE}    ${timeout}=30s
    ${cmd}=    Catenate
    ...    oc wait -n ${ns} pvc/${name} --for=jsonpath="{.spec.resources.requests.storage}"=${to_size}
    ...    --timeout=${timeout}
    Run With Kubeconfig    ${cmd}

Named PVC Should Be Deleted
    [Documentation]    Wait for pvc with ${name} to be deleted
    ...    ${name}    Name of pvc to wait for
    ...    ${ns}    Namespace of named pvc. Defaults to NAMESPACE suite variable
    ...    ${timeout}    Period of time to wait for pvc to be deleted. Default 30s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=30s
    ${timeout_opt}=    Set Variable    --timeout\=${timeout}
    ${output}=    Run With Kubeconfig    oc wait -n ${ns} pvc/${name} --for=Delete ${timeout_opt}
    RETURN    ${output}

Named VolumeSnapshot Should Be Ready
    [Documentation]    Wait for a volumesnap ${name} to become "readyToUse"
    ...    ${name}    Name of volumesnapshot to wait for
    ...    ${timeout}    Period of time to wait for volumeSnapshot ready "readyToUse" state.    Default 30s.
    [Arguments]    ${name}    ${timeout}=30s
    Run With Kubeconfig    oc wait volumesnapshot/${name} -n ${NAMESPACE} --for=jsonpath='{.status.readyToUse}=true'

Named VolumeSnapshot Should Be Deleted
    [Documentation]    Wait for volumesnapshot with ${name} to be deleted
    ...    ${name}    Name of volumesnapshot to wait for
    ...    ${ns}    Namespace of named pod. Defaults to NAMESPACE suite variable, specified by
    ...    ${timeout}    Period of time to wait for volumesnapshot to reach ready status.    Default 30s.
    [Arguments]    ${name}    ${ns}=${NAMESPACE}    ${timeout}=30s
    Run With Kubeconfig    oc wait -n ${ns} volumesnapshot/${name} --for=Delete --timeout=${timeout}

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
