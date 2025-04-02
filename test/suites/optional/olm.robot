*** Settings ***
Documentation       Operator Lifecycle Manager on MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-rpm.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${CATALOG_SOURCE}           ./assets/olm/catalog-source.yaml
${SUBSCRIPTION}             ./assets/olm/subscription.yaml
${SUBSCRIPTION_NAME}        amq-broker
${MARKETPLACE_NAMESPACE}    openshift-marketplace
${OPERATORS_NAMESPACE}      openshift-operators


*** Test Cases ***
Deploy AmqBroker From Red Hat Operators catalog
    [Documentation]    Deploy AMQ Broker from Red Hat Operators catalog.
    [Setup]    Run Keywords
    ...    OLM Should Be Ready
    ...    Create CatalogSource
    ...    Create Subscription

    ${csv}=    Get CSV Name From Subscription    ${OPERATORS_NAMESPACE}    ${SUBSCRIPTION_NAME}
    Wait For CSV    ${OPERATORS_NAMESPACE}    ${csv}
    @{deployments}=    Get Deployments From CSV    ${OPERATORS_NAMESPACE}    ${csv}
    Wait For Deployments    ${OPERATORS_NAMESPACE}    @{deployments}

    [Teardown]    Run Keywords
    ...    Delete CatalogSource
    ...    AND
    ...    Delete Subscription
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
    Verify MicroShift RPM Install

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

OLM Should Be Ready
    [Documentation]    Verify that OLM is running.
    Named Deployment Should Be Available    catalog-operator    openshift-operator-lifecycle-manager
    Named Deployment Should Be Available    olm-operator    openshift-operator-lifecycle-manager

Create CatalogSource
    [Documentation]    Create CatalogSource resource with Red Hat Community Catalog Index.
    Oc Create    -f ${CATALOG_SOURCE}
    Wait Until Keyword Succeeds    5m    10s
    ...    CatalogSource Should Be Ready    ${MARKETPLACE_NAMESPACE}    redhat-operators

CatalogSource Should Be Ready
    [Documentation]    Checks if CatalogSource is ready.
    [Arguments]    ${namespace}    ${name}
    ${catalog}=    Oc Get    catalogsources    ${namespace}    ${name}
    TRY
        Should Be Equal As Strings    READY    ${catalog.status.connectionState.lastObservedState}
    EXCEPT
        Run With Kubeconfig    oc get events -n openshift-marketplace --sort-by='.lastTimestamp'
        Fail    Catalog Source Is Not Ready
    END

Create Subscription
    [Documentation]    Creates subscription.
    Oc Create    -f ${SUBSCRIPTION}
    Wait Until Keyword Succeeds    120s    5s
    ...    Subscription Should Be AtLatestKnown    ${OPERATORS_NAMESPACE}    ${SUBSCRIPTION_NAME}

Subscription Should Be AtLatestKnown
    [Documentation]    Checks if subscription has state "AtLeastKnown"
    [Arguments]    ${namespace}    ${name}
    ${sub}=    Oc Get    subscriptions.operators.coreos.com    ${namespace}    ${name}
    Should Be Equal As Strings    AtLatestKnown    ${sub.status.state}

Get CSV Name From Subscription
    [Documentation]    Obtains Subscription's CSV name.
    [Arguments]    ${namespace}    ${name}
    ${sub}=    Oc Get    subscriptions.operators.coreos.com    ${OPERATORS_NAMESPACE}    ${SUBSCRIPTION_NAME}
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

Delete CatalogSource
    [Documentation]    Delete CatalogSource.
    Oc Delete    -f ${CATALOG_SOURCE}

Delete Subscription
    [Documentation]    Delete Subscription.
    Oc Delete    -f ${SUBSCRIPTION}

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
