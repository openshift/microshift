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
${CLUSTERID_FILE}       /var/lib/microshift/cluster-id
${CLUSTERID_NS}         kube-system


*** Test Cases ***
Verify Cluster ID Change For New Database
    [Documentation]    Verify that cluster ID changes after MicroShift
    ...    database is cleaned and service restarted

    ${old_nid}=    Get MicroShift Cluster ID From Namespace
    ${old_fid}=    Get MicroShift Cluster ID From File
    Create New MicroShift Cluster
    ${new_nid}=    Get MicroShift Cluster ID From Namespace
    ${new_fid}=    Get MicroShift Cluster ID From File

    Should Be Equal As Strings    ${old_nid}    ${old_fid}
    Should Be Equal As Strings    ${new_nid}    ${new_fid}

    Should Not Be Equal As Strings    ${old_nid}    ${new_nid}
    Should Not Be Equal As Strings    ${old_fid}    ${new_fid}

Verify Inconsistent Cluster ID Recovery
    [Documentation]    Verify that cluster ID is correctly rewritten on the
    ...    service restart after manual tampering by a user.

    Tamper With Cluster ID File
    Restart MicroShift

    ${nid}=    Get MicroShift Cluster ID From Namespace
    ${fid}=    Get MicroShift Cluster ID From File
    Should Be Equal As Strings    ${nid}    ${fid}


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

Get MicroShift Cluster ID From Namespace
    [Documentation]    Read and return the cluster ID from the kube-system namespace.
    ${clusterid}=    Oc Get Jsonpath    namespaces    ${CLUSTERID_NS}    ${CLUSTERID_NS}    .metadata.uid

    Should Not Be Empty    ${clusterid}
    RETURN    ${clusterid}

Tamper With Cluster ID File
    [Documentation]    Append invalid characters to the cluster ID file.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    echo -n 123 >> ${CLUSTERID_FILE}
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True

    Should Be Equal As Integers    0    ${rc}
