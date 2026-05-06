*** Settings ***
Documentation       Keywords for deploying and cleaning up kustomize test manifests

Resource            microshift-host.resource
Resource            kubeconfig.resource


*** Keywords ***
Deploy Test Manifests
    [Documentation]    Create a kustomization that deploys a configmap in the given namespace
    [Arguments]    ${manifest_dir}    ${namespace}    ${configmap_name}=test-configmap
    Command Should Work    mkdir -p ${manifest_dir}
    ${kustomization}=    Catenate    SEPARATOR=\n
    ...    resources:
    ...    - configmap.yaml
    ...    namespace: ${namespace}
    Upload String To File    ${kustomization}    ${manifest_dir}/kustomization.yaml
    ${configmap}=    Catenate    SEPARATOR=\n
    ...    apiVersion: v1
    ...    kind: ConfigMap
    ...    metadata:
    ...    \ \ name: ${configmap_name}
    ...    data:
    ...    \ \ path: ${manifest_dir}
    Upload String To File    ${configmap}    ${manifest_dir}/configmap.yaml

Remove Manifest Directory
    [Documentation]    Completely remove a manifest directory under /etc/microshift
    [Arguments]    ${manifest_dir}
    Should Not Be Empty    ${manifest_dir}
    Should Start With    ${manifest_dir}    /etc/microshift/
    Command Should Work    rm -rf ${manifest_dir}

ConfigMap Should Not Exist
    [Documentation]    Verify the configmap was not created in the namespace
    [Arguments]    ${namespace}    ${configmap_name}=test-configmap
    ${stdout}    ${rc}=    Run With Kubeconfig
    ...    oc get configmap ${configmap_name} -n ${namespace}
    ...    allow_fail=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0
    Should Contain    ${stdout}    NotFound
