*** Settings ***
Documentation       Operator Lifecycle Manager on MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-rpm.resource
Resource            ../../resources/oc.resource
Library             DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${CATALOG_SOURCE}           ./assets/olm/catalog-source.yaml
${SUBSCRIPTION}             ./assets/olm/subscription.yaml
${SUBSCRIPTION_NAME}        amq-broker
${MARKETPLACE_NAMESPACE}    openshift-marketplace
${OPERATORS_NAMESPACE}      openshift-operators
${OLM_NAMESPACE}            openshift-operator-lifecycle-manager

# Single namespace install mode
${SINGLE_NS}                olm-microshift-single
${SINGLE_OG}                ./assets/olm/og-single.yaml
${SINGLE_CATALOG}           ./assets/olm/nginx-ok-catalog-source-single.yaml
${SINGLE_CATALOG_NAME}      nginx-ok-catalog
${SINGLE_SUB}               ./assets/olm/nginx-ok1-subscription.yaml
${SINGLE_SUB_NAME}          nginx-ok1-1399
${SINGLE_PKG}               nginx-ok1-1399

# All namespaces install mode with OperatorGroup conflict
${ALL_OG}                   ./assets/olm/og-all.yaml
${ALL_OG_NAME}              og-all
${ALL_CATALOG}              ./assets/olm/nginx-ok-catalog-source-all.yaml
${ALL_CATALOG_NAME}         catalog-all
${ALL_SUB}                  ./assets/olm/nginx-ok2-subscription.yaml
${ALL_SUB_NAME}             nginx-ok2-1399


*** Test Cases ***
Deploy AmqBroker From Red Hat Operators catalog
    [Documentation]    Deploy AMQ Broker from Red Hat Operators catalog.
    [Setup]    Setup Test

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
    ...    Wait For Deployments Deletion    ${OPERATORS_NAMESPACE}    @{deployments}

Install Operator In Single Namespace Mode
    [Documentation]    Creates a dedicated namespace with a SingleNamespace OperatorGroup
    ...    targeting the dedicated namespace (${SINGLE_NS}), installs nginx-ok1-1399 operator,
    ...    verifying successful CSV installation and expected operator resources.
    ...    Migrated from openshift-tests-private 69867.
    [Setup]    OLM Should Be Ready

    Create Namespace    ${SINGLE_NS}
    Oc Create    -f ${SINGLE_OG}
    Oc Create    -f ${SINGLE_CATALOG}
    Wait Until Keyword Succeeds    10m    15s
    ...    CatalogSource Should Be Ready    ${SINGLE_NS}    ${SINGLE_CATALOG_NAME}
    Oc Create    -f ${SINGLE_SUB}
    Wait Until Keyword Succeeds    10m    15s
    ...    Subscription Should Be AtLatestKnown    ${SINGLE_NS}    ${SINGLE_SUB_NAME}
    ${csv}=    Get CSV Name From Subscription    ${SINGLE_NS}    ${SINGLE_SUB_NAME}
    Wait For CSV    ${SINGLE_NS}    ${csv}
    Operator Should Have Expected Resources    ${SINGLE_PKG}    ${SINGLE_NS}

    [Teardown]    Single Namespace Test Teardown

Install Operator In All Namespaces Mode With OperatorGroup Conflict
    [Documentation]    Creates a second AllNamespaces OperatorGroup in openshift-operators
    ...    alongside the existing global-operators OG, installs nginx-ok2-1399 and verifies
    ...    the MultipleOperatorGroupsFound error blocks installation. Resolves the conflict by
    ...    deleting the extra OG and verifies the CSV installs successfully and is copied into
    ...    the default namespace, confirming AllNamespaces mode propagation.
    ...    Migrated from openshift-tests-private 69868.
    [Setup]    OLM Should Be Ready

    VAR    ${csv}=    ${EMPTY}
    Oc Get    operatorgroup    ${OPERATORS_NAMESPACE}    global-operators
    Oc Create    -f ${ALL_OG}
    Oc Create    -f ${ALL_CATALOG}
    Wait Until Keyword Succeeds    10m    15s
    ...    CatalogSource Should Be Ready    ${MARKETPLACE_NAMESPACE}    ${ALL_CATALOG_NAME}
    Oc Create    -f ${ALL_SUB}
    Wait Until Keyword Succeeds    10m    15s
    ...    OperatorGroup Should Have MultipleOperatorGroupsFound    ${OPERATORS_NAMESPACE}    ${ALL_OG_NAME}
    Wait Until Keyword Succeeds    2m    10s
    ...    Subscription Should Have Empty Installed CSV    ${OPERATORS_NAMESPACE}    ${ALL_SUB_NAME}
    Oc Delete    operatorgroup ${ALL_OG_NAME} -n ${OPERATORS_NAMESPACE}
    ${csv}=    Wait For Installed CSV    ${OPERATORS_NAMESPACE}    ${ALL_SUB_NAME}
    Wait Until Keyword Succeeds    2m    10s
    ...    CSV Should Exist In Namespace    ${csv}    default

    [Teardown]    All Namespaces Test Teardown    ${csv}

OLM Network Policies Are Correctly Configured
    [Documentation]    Verifies that OLM-managed NetworkPolicies exist with correct pod
    ...    selectors, policy types, and key ingress/egress port rules.
    ...    Migrated from openshift-tests-private 83581.
    [Setup]    OLM Should Be Ready

    # catalog-operator: metrics ingress; API server, DNS, and gRPC (50051) egress; pod-scoped
    Verify NetworkPolicy Pod Selector Label    catalog-operator    ${OLM_NAMESPACE}    app    catalog-operator
    Verify NetworkPolicy Policy Types    catalog-operator    ${OLM_NAMESPACE}
    Verify NetworkPolicy Spec Field    catalog-operator    ${OLM_NAMESPACE}    ingress    metrics
    Verify NetworkPolicy Spec Field    catalog-operator    ${OLM_NAMESPACE}    egress    50051

    # default-deny-all-traffic: no ingress/egress rules, applies to all pods in OLM namespace
    Verify NetworkPolicy Has Empty Pod Selector    default-deny-all-traffic    ${OLM_NAMESPACE}
    Verify NetworkPolicy Policy Types    default-deny-all-traffic    ${OLM_NAMESPACE}
    Verify NetworkPolicy Spec Field    default-deny-all-traffic    ${OLM_NAMESPACE}    ingress    ${EMPTY}
    Verify NetworkPolicy Spec Field    default-deny-all-traffic    ${OLM_NAMESPACE}    egress    ${EMPTY}

    # olm-operator: metrics ingress; API server and DNS egress; pod-scoped
    Verify NetworkPolicy Pod Selector Label    olm-operator    ${OLM_NAMESPACE}    app    olm-operator
    Verify NetworkPolicy Policy Types    olm-operator    ${OLM_NAMESPACE}
    Verify NetworkPolicy Spec Field    olm-operator    ${OLM_NAMESPACE}    ingress    metrics
    Verify NetworkPolicy Spec Field    olm-operator    ${OLM_NAMESPACE}    egress    53

    # default-deny-all: no ingress/egress rules, applies to all pods in marketplace namespace
    Verify NetworkPolicy Has Empty Pod Selector    default-deny-all    ${MARKETPLACE_NAMESPACE}
    Verify NetworkPolicy Policy Types    default-deny-all    ${MARKETPLACE_NAMESPACE}
    Verify NetworkPolicy Spec Field    default-deny-all    ${MARKETPLACE_NAMESPACE}    ingress    ${EMPTY}
    Verify NetworkPolicy Spec Field    default-deny-all    ${MARKETPLACE_NAMESPACE}    egress    ${EMPTY}

    # default-allow-all: both Ingress and Egress defined with no port restrictions in openshift-operators
    Verify NetworkPolicy Has Empty Pod Selector    default-allow-all    ${OPERATORS_NAMESPACE}
    Verify NetworkPolicy Policy Types    default-allow-all    ${OPERATORS_NAMESPACE}
    Verify NetworkPolicy Spec Field    default-allow-all    ${OPERATORS_NAMESPACE}    ingress
    Verify NetworkPolicy Spec Field    default-allow-all    ${OPERATORS_NAMESPACE}    egress


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Verify MicroShift RPM Install

Setup Test
    [Documentation]    Test setup
    TRY
        OLM Should Be Ready
        Download CatalogSource Image
        Create CatalogSource
        Create Subscription
    EXCEPT
        Oc Logs    deploy/catalog-operator    openshift-operator-lifecycle-manager
        Oc Logs    deploy/olm-operator    openshift-operator-lifecycle-manager
        Fail    Setup failed
    END

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig

OLM Should Be Ready
    [Documentation]    Verify that OLM is running.
    Named Deployment Should Be Available    catalog-operator    openshift-operator-lifecycle-manager
    Named Deployment Should Be Available    olm-operator    openshift-operator-lifecycle-manager

Download CatalogSource Image
    [Documentation]    CatalogSource container image contains a few gigabytes of data.
    ...    Preload it to avoid timeouts during the CatalogSource resource creation.
    ${yaml_content}=    OperatingSystem.Get File    ${CATALOG_SOURCE}
    ${yaml_data}=    Yaml Parse    ${yaml_content}
    VAR    ${image}=    ${yaml_data.spec.image}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    podman pull ${image}
    ...    return_stdout=True    return_stderr=True    return_rc=True    sudo=True
    Should Be Equal As Integers    0    ${rc}

Create CatalogSource
    [Documentation]    Create CatalogSource resource with Red Hat Community Catalog Index.
    Oc Create    -f ${CATALOG_SOURCE}
    Wait Until Keyword Succeeds    10m    15s
    ...    CatalogSource Should Be Ready    ${MARKETPLACE_NAMESPACE}    redhat-operators

CatalogSource Should Be Ready
    [Documentation]    Checks if CatalogSource is ready.
    [Arguments]    ${namespace}    ${name}
    ${catalog}=    Oc Get    catalogsources    ${namespace}    ${name}
    TRY
        Should Be Equal As Strings    READY    ${catalog.status.connectionState.lastObservedState}
    EXCEPT
        Run With Kubeconfig    oc get events -n ${namespace} --sort-by='.lastTimestamp'
        Fail    Catalog Source Is Not Ready
    END

Create Subscription
    [Documentation]    Creates subscription.
    Oc Create    -f ${SUBSCRIPTION}
    Wait Until Keyword Succeeds    10m    15s
    ...    Subscription Should Be AtLatestKnown    ${OPERATORS_NAMESPACE}    ${SUBSCRIPTION_NAME}

Subscription Should Be AtLatestKnown
    [Documentation]    Checks if subscription has state "AtLatestKnown"
    [Arguments]    ${namespace}    ${name}
    ${sub}=    Oc Get    subscriptions.operators.coreos.com    ${namespace}    ${name}
    Should Be Equal As Strings    AtLatestKnown    ${sub.status.state}

Get CSV Name From Subscription
    [Documentation]    Obtains Subscription's CSV name.
    [Arguments]    ${namespace}    ${name}
    ${sub}=    Oc Get    subscriptions.operators.coreos.com    ${namespace}    ${name}
    Should Not Be Empty    ${sub.status.currentCSV}
    ...    msg=Subscription ${name} in ${namespace} has no currentCSV set yet
    RETURN    ${sub.status.currentCSV}

Wait For CSV
    [Documentation]    Waits for ready CSV.
    [Arguments]    ${namespace}    ${name}

    Wait Until Keyword Succeeds    10m    15s
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
        Oc Wait    -n ${namespace} deployment/${deploy}    --for=delete --timeout=${DEFAULT_WAIT_TIMEOUT}
    END

Operator Should Have Expected Resources
    [Documentation]    Verifies that the operators.operators.coreos.com resource for a given
    ...    package and namespace contains expected resource type references in its status.
    ...    The Operator resource name follows the convention <package>.<namespace>.
    [Arguments]    ${package}    ${namespace}
    ${status}=    Oc Get JsonPath
    ...    operators.operators.coreos.com
    ...    ${EMPTY}
    ...    ${package}.${namespace}
    ...    .status
    Should Contain    ${status}    ClusterRole
    Should Contain    ${status}    ClusterRoleBinding
    Should Contain    ${status}    ClusterServiceVersion
    Should Contain    ${status}    CustomResourceDefinition
    Should Contain    ${status}    Deployment
    Should Contain    ${status}    OperatorCondition
    Should Contain    ${status}    Subscription

Single Namespace Test Teardown
    [Documentation]    Cleanup resources created by the single namespace install test.
    ...    Uses Run Keyword And Continue On Failure so all steps run even if one errors.
    Run Keyword And Continue On Failure
    ...    Oc Delete    subscription ${SINGLE_SUB_NAME} -n ${SINGLE_NS} --ignore-not-found
    Run Keyword And Continue On Failure
    ...    Oc Delete    csv --all -n ${SINGLE_NS} --ignore-not-found
    Run Keyword And Continue On Failure
    ...    Oc Delete    catalogsource ${SINGLE_CATALOG_NAME} -n ${SINGLE_NS} --ignore-not-found
    Run Keyword And Continue On Failure
    ...    Oc Delete    namespace ${SINGLE_NS} --ignore-not-found

OperatorGroup Should Have MultipleOperatorGroupsFound
    [Documentation]    Checks that the OperatorGroup status contains the MultipleOperatorGroupsFound condition.
    [Arguments]    ${namespace}    ${name}
    ${status}=    Oc Get JsonPath    operatorgroup    ${namespace}    ${name}    .status
    Should Contain    ${status}    MultipleOperatorGroupsFound

Subscription Should Have Empty Installed CSV
    [Documentation]    Verifies that the subscription's installedCSV is empty, indicating no CSV
    ...    has been installed (e.g. due to OperatorGroup conflict). Call this only after confirming
    ...    the OperatorGroup has MultipleOperatorGroupsFound, which guarantees OLM has reconciled
    ...    the conflict before the subscription state is checked.
    [Arguments]    ${namespace}    ${name}
    ${installed_csv}=    Oc Get JsonPath
    ...    subscriptions.operators.coreos.com
    ...    ${namespace}
    ...    ${name}
    ...    .status.installedCSV
    Should Be Empty    ${installed_csv}

Wait For Installed CSV
    [Documentation]    Polls the subscription until installedCSV is set, then returns it.
    [Arguments]    ${namespace}    ${name}    ${timeout}=10m
    ${csv}=    Wait Until Keyword Succeeds    ${timeout}    15s
    ...    Subscription Should Have Installed CSV    ${namespace}    ${name}
    RETURN    ${csv}

Subscription Should Have Installed CSV
    [Documentation]    Fails if the subscription's installedCSV field is empty, returns the CSV name.
    [Arguments]    ${namespace}    ${name}
    ${csv}=    Oc Get JsonPath
    ...    subscriptions.operators.coreos.com
    ...    ${namespace}
    ...    ${name}
    ...    .status.installedCSV
    Should Not Be Empty    ${csv}
    RETURN    ${csv}

CSV Should Exist In Namespace
    [Documentation]    Verifies that a CSV with the given name exists in the namespace.
    [Arguments]    ${csv}    ${namespace}
    Oc Get    clusterserviceversion.operators.coreos.com    ${namespace}    ${csv}

All Namespaces Test Teardown
    [Documentation]    Cleanup resources created by the all namespaces install test.
    ...    When ${csv} is empty (test failed before CSV was recorded), falls back to
    ...    bulk CSV cleanup to avoid orphaned resources. Uses Run Keyword And Continue On Failure
    ...    so all steps run even if one errors.
    [Arguments]    ${csv}=${EMPTY}
    Run Keyword And Continue On Failure
    ...    Oc Delete    operatorgroup ${ALL_OG_NAME} -n ${OPERATORS_NAMESPACE} --ignore-not-found
    Run Keyword And Continue On Failure
    ...    Oc Delete    subscription ${ALL_SUB_NAME} -n ${OPERATORS_NAMESPACE} --ignore-not-found
    IF    "${csv}" != "${EMPTY}"
        Run Keyword And Continue On Failure
        ...    Oc Delete    csv ${csv} -n ${OPERATORS_NAMESPACE} --ignore-not-found
    ELSE
        Log    csv not recorded; attempting bulk CSV cleanup to avoid orphaned resources    WARN
        Run Keyword And Continue On Failure
        ...    Oc Delete    csv --all -n ${OPERATORS_NAMESPACE} --ignore-not-found
    END
    Run Keyword And Continue On Failure
    ...    Oc Delete    catalogsource ${ALL_CATALOG_NAME} -n ${MARKETPLACE_NAMESPACE} --ignore-not-found

Verify NetworkPolicy Pod Selector Label
    [Documentation]    Verifies that a NetworkPolicy's podSelector has the expected label key=value.
    [Arguments]    ${name}    ${namespace}    ${label_key}    ${expected_value}
    ${actual}=    Oc Get JsonPath
    ...    networkpolicy
    ...    ${namespace}
    ...    ${name}
    ...    .spec.podSelector.matchLabels.${label_key}
    Should Be Equal    ${actual}    ${expected_value}

Verify NetworkPolicy Has Empty Pod Selector
    [Documentation]    Verifies that a NetworkPolicy's podSelector has no matchLabels
    ...    (i.e. applies to all pods in the namespace).
    [Arguments]    ${name}    ${namespace}
    ${labels}=    Oc Get JsonPath
    ...    networkpolicy
    ...    ${namespace}
    ...    ${name}
    ...    .spec.podSelector.matchLabels
    Should Be Empty    ${labels}

Verify NetworkPolicy Policy Types
    [Documentation]    Verifies that a NetworkPolicy has both Ingress and Egress policy types.
    [Arguments]    ${name}    ${namespace}
    Verify NetworkPolicy Spec Field    ${name}    ${namespace}    policyTypes    Ingress
    Verify NetworkPolicy Spec Field    ${name}    ${namespace}    policyTypes    Egress

Verify NetworkPolicy Spec Field
    [Documentation]    Gets .spec.${field} from a NetworkPolicy and asserts based on ${expected}:
    ...    - not provided (defaults to ${NONE}): asserts the field value is not empty
    ...    - ${EMPTY}: asserts the field value is empty (deny-all / no rules check)
    ...    - any other string: asserts the field value contains that string (port or type check)
    [Arguments]    ${name}    ${namespace}    ${field}    ${expected}=${NONE}
    ${value}=    Oc Get JsonPath    networkpolicy    ${namespace}    ${name}    .spec.${field}
    IF    $expected is None
        Should Not Be Empty    ${value}
    ELSE IF    $expected == ''
        Should Be Empty    ${value}
    ELSE
        Should Contain    ${value}    ${expected}
    END
