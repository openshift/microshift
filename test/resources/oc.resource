*** Settings ***
Resource    common.resource
Library     YAML.py

*** Keywords ***

Oc Get
    [Arguments]  ${type}  ${namespace}  ${resource}

    ${yaml_text}=  Run With Kubeconfig  oc get -n ${namespace} -o yaml ${type} ${resource}
    ${yaml_data}=  Yaml Parse  ${yaml_text}

    RETURN  ${yaml_data}
