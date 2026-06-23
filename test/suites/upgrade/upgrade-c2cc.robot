*** Settings ***
Documentation       Tests RHEL 9.8 to RHEL 10.2 upgrade with C2CC enabled across 3 clusters.
...                 Upgrades each cluster one by one and verifies C2CC connectivity
...                 survives at each stage.

Resource            ../../resources/common.resource
Resource            ../../resources/c2cc.resource
Resource            ../../resources/microshift-host.resource
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
    Verify Full C2CC Connectivity

    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Log To Console    Upgrading ${alias} to ${TARGET_REF}
        Upgrade Cluster    ${alias}
        Verify All Clusters Healthy
        Verify All RemoteClusters Healthy
        Wait For Test Pods
        Wait For Service Endpoints
        Verify Full C2CC Connectivity
    END

    [Teardown]    Cleanup Test Workloads


*** Keywords ***
Setup
    [Documentation]    Register all three clusters for SSH and oc access
    ...    and store connection details for reconnection after reboots.
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Should Not Be Empty    ${BOOTC_REGISTRY}    BOOTC_REGISTRY variable is required
    Login MicroShift Host
    Setup Kubeconfig
    Logout MicroShift Host

    Register Remote Cluster    cluster-a    ${USHIFT_HOST}    ${SSH_PORT}    ${KUBECONFIG}
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Register Remote Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}

Teardown
    [Documentation]    Close all connections and clean up.
    Teardown All Remote Clusters
    Remove Kubeconfig

Verify All Clusters Healthy
    [Documentation]    Verify all clusters are running and API server is reachable.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${stdout}=    Oc On Cluster    ${alias}    oc get --raw='/readyz'
        Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True
    END

Verify Full C2CC Connectivity
    [Documentation]    Verify pod-to-pod and pod-to-service connectivity between all cluster pairs.
    VAR    @{clusters}=    cluster-a    cluster-b    cluster-c
    FOR    ${src}    IN    @{clusters}
        FOR    ${dst}    IN    @{clusters}
            IF    '${src}' != '${dst}'
                Test Connectivity Between Clusters    ${src}    ${dst}    pod
                Test Connectivity Between Clusters    ${src}    ${dst}    service
            END
        END
    END

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

    ${boot_id}=    Command On Cluster    ${alias}
    ...    cat /proc/sys/kernel/random/boot_id    sudo_mode=False

    Disruptive Command On Cluster    ${alias}    reboot

    Wait Until Keyword Succeeds    10m    15s
    ...    Cluster Rebooted And Healthy    ${alias}    ${boot_id}

Cluster Rebooted And Healthy
    [Documentation]    Verify cluster has rebooted and greenboot health check passed.
    [Arguments]    ${alias}    ${old_boot_id}

    ${old_conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
    ${status}=    Run Keyword And Return Status
    ...    SSHLibrary.Switch Connection    ${old_conn_id}
    IF    ${status}    SSHLibrary.Close Connection

    ${host}    ${port}    ${kc}=    Get Cluster Connection Info    ${alias}
    Remove Values From List    ${C2CC_REMOTE_ALIASES}    ${alias}
    Register Remote Cluster    ${alias}    ${host}    ${port}    ${kc}

    ${new_boot_id}=    Command On Cluster    ${alias}
    ...    cat /proc/sys/kernel/random/boot_id    sudo_mode=False
    Should Not Be Equal As Strings    ${old_boot_id}    ${new_boot_id}    strip_spaces=True

    ${stdout}=    Command On Cluster    ${alias}
    ...    systemctl show -p SubState greenboot-healthcheck.service --value
    Should Be Equal As Strings    ${stdout}    exited    strip_spaces=True

Get Cluster Connection Info
    [Documentation]    Return host, port, and kubeconfig for a given cluster alias.
    [Arguments]    ${alias}
    IF    '${alias}' == 'cluster-a'
        RETURN    ${USHIFT_HOST}    ${SSH_PORT}    ${KUBECONFIG}
    ELSE IF    '${alias}' == 'cluster-b'
        RETURN    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    ELSE IF    '${alias}' == 'cluster-c'
        RETURN    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}
    ELSE
        Fail    Unknown cluster alias: ${alias}
    END
