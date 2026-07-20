*** Settings ***
Documentation       Verify C2CC respects custom routing table IDs from configuration.
...                 Applies a drop-in config with non-default table IDs (150/151) on
...                 Cluster A, restarts MicroShift, and verifies routes and IP rules
...                 appear in the custom tables.
...
...                 Note: MicroShift does not clean up C2CC state on exit. This is by
...                 design - in case of a crash, pods keep running under kubepods-slice
...                 managed by crio, so C2CC connectivity stays online. Changing routing
...                 table IDs while C2CC is enabled results in duplicated rules in both
...                 old and new tables. The old tables are only cleaned up by rebooting.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Variables ***
${CUSTOM_ROUTE_TABLE}           150
${CUSTOM_SVC_ROUTE_TABLE}       151
${CUSTOM_ROUTING_DROPIN}        60-custom-routing
${RESTART_TIMEOUT}              360s
${RESTART_RETRY}                30s


*** Test Cases ***
Routes Exist In Custom Table
    [Documentation]    Verify routes to remote CIDRs exist in the custom policy routing table.
    ${stdout}=    Command On Cluster    cluster-a
    ...    ${IP_CMD} route show table ${CUSTOM_ROUTE_TABLE}
    Should Contain    ${stdout}    ${CLUSTER_B_POD_CIDR}
    Should Contain    ${stdout}    ${CLUSTER_B_SVC_CIDR}

IP Rules Point To Custom Table
    [Documentation]    Verify IP rules direct remote CIDRs to the custom table.
    ${stdout}=    Command On Cluster    cluster-a    ${IP_CMD} rule show
    Should Contain    ${stdout}    to ${CLUSTER_B_POD_CIDR} lookup ${CUSTOM_ROUTE_TABLE}
    Should Contain    ${stdout}    to ${CLUSTER_B_SVC_CIDR} lookup ${CUSTOM_ROUTE_TABLE}

Service Routes Exist In Custom Service Table
    [Documentation]    Verify service routes exist in the custom service routing table.
    ${stdout}=    Command On Cluster    cluster-a
    ...    ${IP_CMD} route show table ${CUSTOM_SVC_ROUTE_TABLE}
    Should Contain    ${stdout}    ${CLUSTER_A_SVC_CIDR}

Service IP Rules Point To Custom Service Table
    [Documentation]    Verify service IP rules point to the custom service table.
    ${stdout}=    Command On Cluster    cluster-a    ${IP_CMD} rule show
    Should Contain    ${stdout}
    ...    from ${CLUSTER_B_POD_CIDR} to ${CLUSTER_A_SVC_CIDR} lookup ${CUSTOM_SVC_ROUTE_TABLE}
    Should Contain    ${stdout}
    ...    from ${CLUSTER_B_SVC_CIDR} to ${CLUSTER_A_SVC_CIDR} lookup ${CUSTOM_SVC_ROUTE_TABLE}

Old Default Tables Are Not Cleaned Up
    [Documentation]    Verify that the old default tables (200/201) still contain routes
    ...    from before the config change. MicroShift intentionally does not clean up
    ...    C2CC state on exit - in case of a crash, pods keep running and C2CC
    ...    connectivity stays online. Changing table IDs produces duplicated rules
    ...    which are resolved by rebooting the system.
    ${routes}=    Command On Cluster    cluster-a    ${IP_CMD} route show table 200
    Should Contain    ${routes}    ${CLUSTER_B_POD_CIDR}
    ${rules}=    Command On Cluster    cluster-a    ${IP_CMD} rule show
    Should Contain    ${rules}    lookup 200
    ${svc_routes}=    Command On Cluster    cluster-a    ${IP_CMD} route show table 201
    Should Contain    ${svc_routes}    ${CLUSTER_A_SVC_CIDR}
    Should Contain    ${rules}    lookup 201


*** Keywords ***
Setup
    [Documentation]    Register clusters, apply custom routing drop-in, restart MicroShift.
    C2CC Suite Setup
    Apply Custom Routing Config
    Restart And Wait For Healthy

Teardown
    [Documentation]    Remove custom routing drop-in, flush custom tables, restore defaults.
    Remove Custom Routing Config
    Restart And Wait For Healthy
    Flush Custom Tables
    C2CC Suite Teardown

Apply Custom Routing Config
    [Documentation]    Write a drop-in that overrides routing table IDs.
    VAR    ${dropin}=    /etc/microshift/config.d/${CUSTOM_ROUTING_DROPIN}.yaml
    ${yaml}=    Catenate    SEPARATOR=\n
    ...    clusterToCluster:
    ...    ${SPACE}${SPACE}routing:
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}routeTableID: ${CUSTOM_ROUTE_TABLE}
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}serviceRouteTableID: ${CUSTOM_SVC_ROUTE_TABLE}
    ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    cluster-a
    SSHLibrary.Switch Connection    ${conn_id}
    Upload String To File    ${yaml}    ${dropin}

Remove Custom Routing Config
    [Documentation]    Remove the custom routing drop-in file.
    Command On Cluster    cluster-a
    ...    rm -f /etc/microshift/config.d/${CUSTOM_ROUTING_DROPIN}.yaml

Flush Custom Tables
    [Documentation]    Remove routes and rules from custom tables to avoid interfering
    ...    with subsequent tests that expect only tables 200/201.
    Disruptive Command On Cluster    cluster-a
    ...    ${IP_CMD} route flush table ${CUSTOM_ROUTE_TABLE}
    Disruptive Command On Cluster    cluster-a
    ...    ${IP_CMD} route flush table ${CUSTOM_SVC_ROUTE_TABLE}
    ${rules}=    Command On Cluster    cluster-a    ${IP_CMD} rule show
    ${lines}=    Split To Lines    ${rules}
    FOR    ${line}    IN    @{lines}
        IF    'lookup ${CUSTOM_ROUTE_TABLE}' in $line or 'lookup ${CUSTOM_SVC_ROUTE_TABLE}' in $line
            ${priority}=    Fetch From Left    ${line}    :
            ${priority}=    Strip String    ${priority}
            ${selector}=    Fetch From Right    ${line}    :\t
            Disruptive Command On Cluster    cluster-a
            ...    ${IP_CMD} rule del prio ${priority} ${selector}
        END
    END

Restart And Wait For Healthy
    [Documentation]    Restart MicroShift and wait until the cluster is healthy.
    Command On Cluster    cluster-a    systemctl restart microshift
    Wait Until Keyword Succeeds    ${RESTART_TIMEOUT}    ${RESTART_RETRY}
    ...    Verify Cluster Is Healthy    cluster-a
