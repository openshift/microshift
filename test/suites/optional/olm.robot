*** Settings ***
Documentation       Operator Lifecycle Manager on MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${CATALOG_SOURCE}           ./assets/olm/catalog-source.yaml
${SUB_CERT_MANAGER}         ./assets/olm/subscription-cert-manager.yaml
${MARKETPLACE_NAMESPACE}    openshift-marketplace
${OPERATORS_NAMESPACE}      openshift-operators


*** Test Cases ***
Deploy CertManager From OperatorHubIO
    [Documentation]    Deploy CertManager from OperatorHub Catalog.
    [Setup]    Run Keywords
    ...    OLM Should Be Ready
    ...    Create OperatorHub CatalogSource
    ...    Create CertManager Subscription

    ${csv}=    Get CSV Name From Subscription    ${OPERATORS_NAMESPACE}    my-cert-manager
    Wait For CSV    ${OPERATORS_NAMESPACE}    ${csv}
    @{deployments}=    Get Deployments From CSV    ${OPERATORS_NAMESPACE}    ${csv}
    Wait For Deployments    ${OPERATORS_NAMESPACE}    @{deployments}

    [Teardown]    Run Keywords
    ...    Delete OperatorHub CatalogSource
    ...    AND
    ...    Delete CertManager Subscription
    ...    AND
    ...    Delete CSV    ${OPERATORS_NAMESPACE}    ${csv}
    ...    AND
    ...    Wait For Deployments Deletion    @{deployments}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

OLM Should Be Ready
    [Documentation]    Verify that OLM is running.
    Named Deployment Should Be Available    catalog-operator    openshift-operator-lifecycle-manager
    Named Deployment Should Be Available    olm-operator    openshift-operator-lifecycle-manager

Create OperatorHub CatalogSource
    [Documentation]    Create CatalogSource resource pointing to OperatorHub.io catalog.
    Oc Create    -f ${CATALOG_SOURCE}
    Wait Until Keyword Succeeds    120s    5s
    ...    CatalogSource Should Be Ready    ${MARKETPLACE_NAMESPACE}    operatorhubio-catalog

CatalogSource Should Be Ready
    [Documentation]    Checks if CatalogSource is ready.
    [Arguments]    ${namespace}    ${name}
    ${catalog}=    Oc Get    catalogsources    ${namespace}    ${name}
    Should Be Equal As Strings    READY    ${catalog.status.connectionState.lastObservedState}

Create CertManager Subscription
    [Documentation]    Creates cert-manager subscription.
    Oc Create    -f ${SUB_CERT_MANAGER}
    Wait Until Keyword Succeeds    120s    5s
    ...    Subscription Should Be AtLatestKnown    ${OPERATORS_NAMESPACE}    my-cert-manager

Subscription Should Be AtLatestKnown
    [Documentation]    Checks if subscription has state "AtLeastKnown"
    [Arguments]    ${namespace}    ${name}
    ${sub}=    Oc Get    subscriptions.operators.coreos.com    ${namespace}    ${name}
    Should Be Equal As Strings    AtLatestKnown    ${sub.status.state}

Get CSV Name From Subscription
    [Documentation]    Obtains Subscription's CSV name.
    [Arguments]    ${namespace}    ${name}
    ${sub}=    Oc Get    subscriptions.operators.coreos.com    ${OPERATORS_NAMESPACE}    my-cert-manager
    RETURN    ${sub.status.currentCSV}

Wait For CSV
    [Documentation]    Waits for ready CSV.
    [Arguments]    ${namespace}    ${name}

    Wait Until Keyword Succeeds    120s    5s
    ...    CSV Should Be Succeeded    ${namespace}    ${name}

CSV Should Be Succeeded
    [Documentation]    Verifies that phase of CSV is "Succeeded".
    [Arguments]    ${namespace}    ${name}
    ${csv_phase}=    Oc Get JsonPath    csv    ${namespace}    ${name}    .status.phase
    Should Be Equal As Strings    Succeeded    ${csv_phase}

Get Deployments From CSV
    [Documentation]    Obtains list of Deployments created by CSV.
    [Arguments]    ${namespace}    ${name}
    ${csv_dss}=    Oc Get JsonPath
    ...    csv
    ...    ${namespace}
    ...    ${name}
    ...    .spec.install.spec.deployments[*].name
    @{deployments}=    Split String    ${csv_dss}    ${SPACE}
    RETURN    @{deployments}

Wait For Deployments
    [Documentation]    Waits for availability of Deployments.
    [Arguments]    ${namespace}    @{deployments}
    FOR    ${deploy}    IN    @{deployments}
        Named Deployment Should Be Available    ${deploy}    ${namespace}    120s
    END

Delete OperatorHub CatalogSource
    [Documentation]    Delete OperatorHub's CatalogSource.
    Oc Delete    -f ${CATALOG_SOURCE}

Delete CertManager Subscription
    [Documentation]    Delete CertManager Subscription.
    Oc Delete    -f ${SUB_CERT_MANAGER}

Delete CSV
    [Documentation]    Delete CSV.
    [Arguments]    ${namespace}    ${name}
    Oc Delete    csv -n ${namespace} ${name}

Wait For Deployments Deletion
    [Documentation]    Wait for Deployments to be deleted.
    [Arguments]    ${namespace}    @{deployments}
    FOR    ${deploy}    IN    @{deployments}
        Run With Kubeconfig    kubectl wait deployment --for=delete -n ${namespace} ${deploy} --timeout=60s
    END
