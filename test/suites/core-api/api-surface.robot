*** Settings ***
Documentation       Tests verifying MicroShift's minimal API surface: the
...                 recommended OpenShift APIs are served and the APIs not
...                 required for the MVP are disabled.
...
...                 Ported from openshift-tests-private:
...                 OCP-53693, OCP-54816

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Library             Collections

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           slow


*** Test Cases ***
MicroShift Serves Only The Expected API Surface
    [Documentation]    Verify the recommended OpenShift APIs are served and the
    ...    APIs not required for the MVP (including RoleBindingRestrictions and
    ...    apirequestcount) are disabled.
    ...    OCP-53693, OCP-54816

    ${api_resources}=    Run With Kubeconfig    oc api-resources

    Recommended APIs Are Served    ${api_resources}
    Disabled APIs Are Not Served    ${api_resources}

    # RoleBindingRestrictions API (authorization.openshift.io) is removed (OCP-54816)
    Should Not Contain    ${api_resources}    rolebindingrestriction
    Oc Get Should Fail    rolebindingrestriction -A

    # apirequestcount API is disabled (OCP-53693)
    Oc Get Should Fail    apirequestcount


*** Keywords ***
Recommended APIs Are Served
    [Documentation]    The MVP/recommended OpenShift APIs must be present.
    [Arguments]    ${api_resources}
    Should Contain    ${api_resources}    security.openshift.io/v1
    Should Contain    ${api_resources}    route.openshift.io/v1

Disabled APIs Are Not Served
    [Documentation]    OpenShift API groups not required for the MVP must be
    ...    absent. operator.openshift.io is expected only when cert-manager is
    ...    installed.
    [Arguments]    ${api_resources}
    VAR    @{disabled_apis}=
    ...    build    apps    image    imageregistry.operator    config    user    template
    IF    "certmanager" not in """${api_resources}"""
        Append To List    ${disabled_apis}    operator
    END
    FOR    ${api}    IN    @{disabled_apis}
        Should Not Contain    ${api_resources}    ${api}.openshift
    END

Oc Get Should Fail
    [Documentation]    'oc get' for the given resource must fail (API disabled).
    [Arguments]    ${resource}
    ${output}    ${rc}=    Run With Kubeconfig    oc get ${resource}
    ...    allow_fail=${TRUE}    return_rc=${TRUE}
    Should Not Be Equal As Integers    ${rc}    0
