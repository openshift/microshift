*** Settings ***
Library    OperatingSystem
Library    String
Resource   ../resources/kubeconfig.resource

*** Keywords ***
Setup Suite
    [Documentation]    Setup a namespace-less test suite by configuring Kubeconfig
    Check Required Env Variables
    Setup Kubeconfig

Setup Suite With Namespace
    [Documentation]    Setup a test suite by creating a unique namespace and configuring Kubeconfig
    Check Required Env Variables
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

Check Required Env Variables
    [Documentation]    Fail if any of the required environment variables is missing.
    Run Keyword If  "${USHIFT_HOST}"=="${EMPTY}"    Fatal Error    USHIFT_HOST variable is required
    Run Keyword If  "${USHIFT_USER}"=="${EMPTY}"    Fatal Error    USHIFT_USER variable is required
