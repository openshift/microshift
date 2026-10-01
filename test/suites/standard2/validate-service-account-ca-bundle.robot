*** Settings ***
Documentation       Tests validating the service account CA bundle contents
...                 by checking the kube-root-ca.crt ConfigMap that gets
...                 automatically created in every namespace.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/kubeconfig.resource

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown Suite With Namespace

Test Tags           certificates


*** Variables ***
${USHIFT_HOST}                  ${EMPTY}
${USHIFT_USER}                  ${EMPTY}
${ROOT_CA_CONFIGMAP_NAME}       kube-root-ca.crt


*** Test Cases ***
Root CA ConfigMap Contains All Signers
    [Documentation]    Verify that the kube-root-ca.crt ConfigMap contains a certificate
    ...    from the consolidated serving CA (serving-ca).
    ${configmap}=    Oc Get    configmap    ${NAMESPACE}    ${ROOT_CA_CONFIGMAP_NAME}
    VAR    ${ca_bundle}=    ${configmap.data['ca.crt']}
    Should Not Be Empty    ${ca_bundle}

    ${subjects}=    Get Certificate Subjects From Bundle    ${ca_bundle}
    Should Contain    ${subjects}    serving-ca


*** Keywords ***
Get Certificate Subjects From Bundle
    [Documentation]    Extract all certificate subjects from a PEM-encoded CA bundle string.
    ...    For CA certificates, the Subject field contains the signer name.
    [Arguments]    ${ca_bundle}
    ${subjects}=    Run With Kubeconfig
    ...    echo "${ca_bundle}" | openssl crl2pkcs7 -nocrl -certfile /dev/stdin 2>/dev/null | openssl pkcs7 -print_certs -text -noout 2>/dev/null | grep "Subject:"
    Log    CA Bundle Subjects: ${subjects}
    RETURN    ${subjects}
