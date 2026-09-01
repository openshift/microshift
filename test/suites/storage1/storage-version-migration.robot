*** Settings ***
Documentation       Storage version migration test suite.

Library             Process
Library             ../../resources/DataFormats.py
Resource            ../../resources/oc.resource
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${BETA_CRD}                     assets/storage-version-migration/crd.beta.yaml
${STABLE_CRD}                   assets/storage-version-migration/crd.stable.yaml
${CR_RESOURCE}                  assets/storage-version-migration/cr.yaml
${BETA_MIGRATION_REQUEST}       assets/storage-version-migration/migration.beta.yaml
${STABLE_MIGRATION_REQUEST}     assets/storage-version-migration/migration.stable.yaml


*** Test Cases ***
Storage Version Migration Test
    [Documentation]    Verify that storage migrations get created when CRDs get updated.
    [Tags]    restart    slow    smoke

    Create Beta Migration
    Wait Until Keyword Succeeds    5x    10s
    ...    Validate Migration    v1beta1

    Update Beta CRD To Stable

    Create Stable Migration
    Wait Until Keyword Succeeds    5x    10s
    ...    Validate Migration    v1


*** Keywords ***
Setup
    [Documentation]    Create initial setup with CRD and resources
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

    Create Beta CRD
    # There may be a lag between creating the CRD and being able to
    # use it, so retry uploading the first copy of the resource a few
    # times.
    Wait Until Keyword Succeeds    5x    10s
    ...    Create Custom Resource

Teardown
    [Documentation]    Delete all created resources
    Delete Migration Resources

Create Beta CRD
    [Documentation]    Create beta CRD
    Run With Kubeconfig    oc apply -f ${BETA_CRD}

Create Beta Migration
    [Documentation]    Create beta migration request
    Run With Kubeconfig    oc apply -f ${BETA_MIGRATION_REQUEST}

Create Stable Migration
    [Documentation]    Create stable migration request
    Run With Kubeconfig    oc apply -f ${STABLE_MIGRATION_REQUEST}

Create Custom Resource
    [Documentation]    Create beta version resource
    Run With Kubeconfig    oc apply -f ${CR_RESOURCE}

Update Beta CRD To Stable
    [Documentation]    Update the beta versin of CRD to stable
    Run With Kubeconfig    oc apply -f ${STABLE_CRD}

Validate Migration
    [Documentation]    Validate that a migration resource has succeeded in migrating the CR,
    ...    we should expect to see a StorageVersionMigration resource has succeeded and that the resource
    ...    has the correct version updated.
    [Arguments]    ${api_version}
    ${yaml_data}=    Oc Get    migrates.test.resource    ""    default
    ${storage_yaml_data}=    Oc Get    storageversionmigrations.migration.k8s.io    ""    test.resource-${api_version}
    Should Be Equal    ${yaml_data.apiVersion}    test.resource/${api_version}
    Should Be Equal    ${storage_yaml_data.status.conditions[0].type}    Succeeded

Delete Migration Resources
    [Documentation]    Remove the CRD and Storage State and Version Migration resources
    VAR    ${query}=    {.items[?(@.spec.resource.group=='test.resource')].metadata.name}
    ${migration_resource_name}=    Run With Kubeconfig
    ...    oc get storageversionmigration -o jsonpath="${query}"

    Run With Kubeconfig    oc delete -f ${STABLE_CRD}    True
    Run With Kubeconfig    oc delete storageversionmigration ${migration_resource_name}    True
