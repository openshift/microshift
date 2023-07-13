*** Settings ***
Documentation       Storage version migration test suite.

Library             Process
Library             ../resources/DataFormats.py
Resource            ../resources/oc.resource
Resource            ../resources/common.resource
Resource            ../resources/kubeconfig.resource
Resource            ../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${BETA_CRD}         assets/storage-version-migration/crd.beta.yaml
${STABLE_CRD}       assets/storage-version-migration/crd.stable.yaml
${CR_RESOURCE}      assets/storage-version-migration/cr.yaml


*** Test Cases ***
Storage Version Migration Test
    [Documentation]    Verify that storage migrations get created when CRDs get updated.
    [Tags]    restart    slow    smoke

    # The migration trigger runs on a 10min cycle, we restart Microshift to speed up discovery.
    Wait Until Keyword Succeeds    5x    20s
    ...    Restart MicroShift

    Wait Until Keyword Succeeds    5x    10s
    ...    Validate Migration    v1beta1    8F5KO5MYqcM=

    Update Beta CRD To Stable

    Wait Until Keyword Succeeds    5x    20s
    ...    Restart MicroShift

    Wait Until Keyword Succeeds    5x    10s
    ...    Validate Migration    v1    O2TShlD54rQ=


*** Keywords ***
Setup
    [Documentation]    Create initial setup with CRD and resources
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

    Create Beta CRD
    Create Custom Resource

Teardown
    [Documentation]    Delete all created resources
    Delete Migration Resources

Create Beta CRD
    [Documentation]    Create beta CRD
    Run With Kubeconfig    oc apply -f ${BETA_CRD}

Create Custom Resource
    [Documentation]    Create beta version resource
    Run With Kubeconfig    oc apply -f ${CR_RESOURCE}

Update Beta CRD To Stable
    [Documentation]    Update the beta versin of CRD to stable
    Run With Kubeconfig    oc apply -f ${STABLE_CRD}

Validate Migration
    [Documentation]    Validate that a migration resource was created the CRD,
    ...    we should expect to see a StorageState and StorageVersionMigration resource created.
    ...    With in the Migration and State objects we should see the correct version and object
    ...    hash defined. Storage hash is a trimmed base64 encoded value of the APIResource, it
    ...    should be consistant per CRD structure.
    [Arguments]    ${api_version}    ${hash}
    ${yaml_data}=    Oc Get    migrates.test.resource    ""    default
    ${storage_yaml_data}=    Oc Get    storagestates.migration.k8s.io    ""    migrates.test.resource
    ${migration_json_text}=    Run With Kubeconfig
    ...    oc get storageversionmigration -o jsonpath="{.items[?(@.spec.resource.group=='test.resource')]}"
    ${migration_json_data}=    Json Parse    ${migration_json_text}
    Should Be Equal    ${yaml_data.apiVersion}    test.resource/${api_version}
    Should Be Equal    ${storage_yaml_data.status.currentStorageVersionHash}    ${hash}
    Should Be Equal    ${migration_json_data.spec.resource.version}    ${api_version}

Delete Migration Resources
    [Documentation]    Remove the CRD and Storage State and Version Migration resources
    ${query}=    Set Variable    {.items[?(@.spec.resource.group=='test.resource')].metadata.name}
    ${migration_resource_name}=    Run With Kubeconfig
    ...    oc get storageversionmigration -o jsonpath="${query}"

    Run With Kubeconfig    oc delete -f ${STABLE_CRD}    True
    Run With Kubeconfig    oc delete storagestates migrates.test.resource    True
    Run With Kubeconfig    oc delete storageversionmigration ${migration_resource_name}    True
