*** Settings ***
Documentation       Tests verifying kube-controller-manager runs with
...                 OpenShift flavor flags.
...
...                 Ported from openshift-tests-private:
...                 OCP-56673

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/oc.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${KCM_LOG_VLEVEL}       SEPARATOR=\n
...                     ---
...                     debugging:
...                     \ \ logVLevel: 2


*** Test Cases ***
Verify KCM Runs With OpenShift Flavor Flags
    [Documentation]    Verify that kube-controller-manager runs with the
    ...    OpenShift flavor flags logged in MicroShift journal.
    ...    OCP-56673
    [Setup]    Setup KCM Log VLevel

    ${kcm_flags}=    Command Should Work
    ...    journalctl -u microshift --no-pager | grep kube-controller-manager | grep FLAG

    FOR    ${flag}    IN
    ...    --enable-dynamic-provisioning\="true"
    ...    --allocate-node-cidrs\="true"
    ...    --use-service-account-credentials\="true"
    ...    --leader-elect\="false"
    ...    --leader-elect-retry-period\="3s"
    ...    --leader-elect-resource-lock\="leases"
    ...    --cluster-signing-duration\="720h0m0s"
    ...    --secure-port\="10257"
    ...    --cert-dir\="/var/run/kubernetes"
    ...    --root-ca-file\="/var/lib/microshift/certs/ca-bundle/service-account-token-ca.crt"
    ...    --service-account-private-key-file\="/var/lib/microshift/resources/kube-apiserver/secrets/service-account-key/service-account.key"
    ...    --cluster-signing-cert-file\="/var/lib/microshift/certs/kubelet-csr-signer-signer/csr-signer/ca.crt"
    ...    --cluster-signing-key-file\="/var/lib/microshift/certs/kubelet-csr-signer-signer/csr-signer/ca.key"
    ...    --kube-api-qps\="150"
    ...    --kube-api-burst\="300"
        Should Contain    ${kcm_flags}    ${flag}
    END

    Should Match Regexp    ${kcm_flags}    --controllers\="\\[.*-bootstrapsigner.*-tokencleaner.*-ttl.*\\]"

    ${openshift_config}=    Command Should Work
    ...    journalctl -u microshift --no-pager | grep kube-controller-manager | grep openshift-config
    Should Contain    ${openshift_config}    --openshift-config=""

    [Teardown]    Teardown KCM Log VLevel


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Remove Drop In MicroShift Config    10-kcm-vlevel
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Setup KCM Log VLevel
    [Documentation]    Set verbose log level and restart to capture KCM flags
    Drop In MicroShift Config    ${KCM_LOG_VLEVEL}    10-kcm-vlevel
    Restart MicroShift

Teardown KCM Log VLevel
    [Documentation]    Remove KCM verbose log level drop-in and restart
    Remove Drop In MicroShift Config    10-kcm-vlevel
    Restart MicroShift
