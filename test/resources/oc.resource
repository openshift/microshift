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
