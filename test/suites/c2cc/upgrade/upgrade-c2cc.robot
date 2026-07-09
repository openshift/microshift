*** Settings ***
Documentation       Tests RHEL 9.8 to RHEL 10.2 upgrade with C2CC enabled across 3 clusters.
...                 Upgrades each cluster one by one and verifies C2CC connectivity
...                 survives at each stage.

Resource            ../../../resources/common.resource
Resource            ../../../resources/c2cc.resource
Resource            ../../../resources/microshift-host.resource
Library             Collections
Library             SSHLibrary

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc    ostree


*** Variables ***
${TARGET_REF}               ${EMPTY}
${BOOTC_REGISTRY}           ${EMPTY}
${CLUSTER_A_POD_CIDR}       ${EMPTY}
${CLUSTER_A_SVC_CIDR}       ${EMPTY}
${CLUSTER_A_DOMAIN}         ${EMPTY}
${CLUSTER_B_POD_CIDR}       ${EMPTY}
${CLUSTER_B_SVC_CIDR}       ${EMPTY}
${CLUSTER_B_DOMAIN}         ${EMPTY}
${KUBECONFIG_B}             ${EMPTY}
${CLUSTER_C_POD_CIDR}       ${EMPTY}
${CLUSTER_C_SVC_CIDR}       ${EMPTY}
${CLUSTER_C_DOMAIN}         ${EMPTY}
${KUBECONFIG_C}             ${EMPTY}


*** Test Cases ***
Upgrade C2CC Clusters And Verify
    [Documentation]    Upgrades 3 C2CC-connected clusters one by one
    ...    and verifies health and C2CC connectivity after each upgrade.

    Verify All Clusters Healthy
    Verify All RemoteClusters Healthy
    Deploy Test Workloads
    Wait Until Keyword Succeeds    10m    10s    Verify Full C2CC Connectivity

    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Log To Console    Upgrading ${alias} to ${TARGET_REF}
        Upgrade Cluster    ${alias}
        Verify All Clusters Healthy
        Verify All RemoteClusters Healthy
        Wait For Test Pods
        Wait For Service Endpoints
        Wait Until Keyword Succeeds    10m    10s    Verify Full C2CC Connectivity
    END

    [Teardown]    Cleanup Test Workloads


*** Keywords ***
Setup
    [Documentation]    Validate required variables and register all three clusters.
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Should Not Be Empty    ${BOOTC_REGISTRY}    BOOTC_REGISTRY variable is required
    C2CC Suite Setup

Teardown
    [Documentation]    Close all connections and clean up.
    C2CC Suite Teardown

Upgrade Cluster
    [Documentation]    Upgrade a specific cluster to the target bootc image
    ...    and verify it booted into the new deployment without rollback.
    [Arguments]    ${alias}

    ${initial_deploy_id}=    Get Deployment Id On Cluster    ${alias}

    Command On Cluster
    ...    ${alias}
    ...    printf '[[registry]]\nlocation = "${BOOTC_REGISTRY}"\ninsecure = true\n' | sudo tee /etc/containers/registries.conf.d/999-microshift-insecure-registry.conf > /dev/null

    Command On Cluster    ${alias}    bootc switch --quiet ${BOOTC_REGISTRY}/${TARGET_REF}

    Command On Cluster    ${alias}
    ...    rm -f /etc/containers/registries.conf.d/999-microshift-insecure-registry.conf

    Reboot Cluster And Wait    ${alias}

    ${current_deploy_id}=    Get Deployment Id On Cluster    ${alias}
    Should Not Be Equal As Strings    ${current_deploy_id}    ${initial_deploy_id}
    ...    msg=${alias} rolled back to initial deployment

Get Deployment Id On Cluster
    [Documentation]    Get the booted image digest from a specific cluster.
    [Arguments]    ${alias}
    ${stdout}=    Command On Cluster
    ...    ${alias}
    ...    bootc status --booted --json | python3 -c "import sys,json; print(json.load(sys.stdin)['status']['booted']['image']['imageDigest'])"
    RETURN    ${stdout}

Reboot Cluster And Wait
    [Documentation]    Reboot a cluster and wait for it to come back with greenboot healthy.
    [Arguments]    ${alias}
    Reboot Clusters Simultaneously    ${alias}
