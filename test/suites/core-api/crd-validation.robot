*** Settings ***
Documentation       Tests for CRD schema validation

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${CRD_FILE}     ./assets/crd-validation/crd.yaml
${CR_VALID}     ./assets/crd-validation/cr-valid.yaml


*** Test Cases ***
CRD Is Established And Custom Resource Can Be Created
    [Documentation]    Verify a CRD reaches Established condition and a valid
    ...    custom resource can be created and retrieved.
    Wait Until Keyword Succeeds    5x    10s
    ...    Oc Apply    -f ${CR_VALID} -n ${NAMESPACE}
    ${cr}=    Oc Get    crontab    ${NAMESPACE}    test-crontab
    Should Be Equal    ${cr.spec.cronSpec}    * * * * */5
    Should Be Equal    ${cr.spec.image}    test-image

    [Teardown]    Oc Delete    -f ${CR_VALID} -n ${NAMESPACE} --ignore-not-found


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Setup Suite With Namespace
    Oc Apply    -f ${CRD_FILE}

Teardown
    [Documentation]    Test suite teardown
    Oc Delete    -f ${CRD_FILE} --ignore-not-found
    Teardown Suite With Namespace
