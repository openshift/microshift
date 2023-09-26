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

Oc Wait For
    [Documentation]    Runs 'oc wait'.
    ...    ${type_name}    expects a "type/name", e.g. pod/foo
    ...    ${wait_for}    expects a key=value. See 'oc wait --help' for examples
    ...    ${timeout}    Takes an integer and a single character time-unit identifier, e.g.: 30s
    ...    If no timeout value is provided, defaults to 30s.
    [Arguments]    ${type_name}    ${wait_for}    ${timeout}=30s    ${ns}=${NAMESPACE}

    ${timeout_opt}=    Set Variable    --timeout\=${timeout}
    ${output}=    Run With Kubeconfig    oc wait -n ${ns} ${type_name} --for=${wait_for} ${timeout_opt}
    RETURN    ${output}

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
