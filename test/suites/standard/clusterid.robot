*** Settings ***
Documentation       Tests verifying MicroShift cluster ID functionality

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CLUSTERID_FILE}           /var/lib/microshift/cluster-id
${CLUSTERID_NS}             kube-public
${CLUSTERID_RESOURCE}       microshift-version


*** Test Cases ***
Compare Cluster ID From File And ConfigMap
    [Documentation]    Verify that cluster ID is the same when read
    ...    from file and configmap
    ${file_id}=    Get MicroShift Cluster ID From File
    ${conf_id}=    Get MicroShift Cluster ID From ConfigMap
    Should Be Equal As Strings    ${file_id}    ${conf_id}

Verify Cluster ID Change For New Database
    [Documentation]    Verify that cluster ID changes after MicroShift
    ...    database is cleaned and service restarted

    ${old_id}=    Get MicroShift Cluster ID From ConfigMap
    Create New MicroShift Cluster
    ${new_id}=    Get MicroShift Cluster ID From ConfigMap
    Should Not Be Equal As Strings    ${old_id}    ${new_id}


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Remove Kubeconfig
    Logout MicroShift Host

Create New MicroShift Cluster
    [Documentation]    Clean the database and restart MicroShift service.
    Cleanup MicroShift    --all    --keep-images
    Enable MicroShift
    Start MicroShift
    Setup Kubeconfig
    Restart Greenboot And Wait For Success

Get MicroShift Cluster ID From File
    [Documentation]    Read and return the cluster ID from the file.
    ${stdout}    ${rc}=    Execute Command
    ...    cat ${CLUSTERID_FILE}
    ...    sudo=True    return_rc=True    return_stdout=True
    Should Be Equal As Integers    0    ${rc}

    Should Not Be Empty    ${stdout}
    RETURN    ${stdout}

Get MicroShift Cluster ID From ConfigMap
    [Documentation]    Read and return the cluster ID from the configmap.
    ${yaml_data}=    Oc Get    configmap    ${CLUSTERID_NS}    ${CLUSTERID_RESOURCE}
    ${clusterid}=    Set Variable    ${yaml_data.data.clusterid}

    Should Not Be Empty    ${clusterid}
    RETURN    ${clusterid}
