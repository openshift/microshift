*** Settings ***
Library    OperatingSystem
Library    SSHLibrary
Library    String
Resource   ../resources/kubeconfig.resource

*** Keywords ***
Setup Suite
    [Documentation]    Setup a namespace-less test suite by configuring Kubeconfig
    Setup Kubeconfig

Setup Suite With Namespace
    [Documentation]    Setup a test suite by creating a unique namespace and configuring Kubeconfig
    Setup Kubeconfig
    ${rand}=    Generate Random String
    ${rand}=    Convert To Lower Case    ${rand}
    ${ns}=    Catenate    SEPARATOR=-    "test"    ${rand}
    Set Suite Variable    \${NAMESPACE}    ${ns}
    Create Namespace    ${ns}

Teardown Suite
    [Documentation]    Teardown the namespace-less test suite by removing the Kubeconfig.
    Remove Kubeconfig

Teardown Suite With Namespace
    [Documentation]    Teardown the namespaced test suite by removing the namespace and removing the Kubeconfig.
    Remove Namespace    ${NAMESPACE}
    Remove Kubeconfig
