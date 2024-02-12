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

Verify Sos Report Contains ID In kube-system Namespace
    [Documentation]    Verify that cluster ID can be retrieved from Sos Report

    ${sos_report_tarfile}=    Create Sos Report
    ${sos_report_id}=    Get MicroShift Cluster ID From Sos Report    ${sos_report_tarfile}
    ${id}=    Get MicroShift Cluster ID From Namespace
    Should Be Equal As Strings    ${sos_report_id}    ${id}

Verify Inconsistent Cluster ID Recovery
    [Documentation]    Verify that cluster ID file is correctly rewritten
    ...    on the service restart after manual tampering by a user.

    Tamper With Cluster ID File
    Restart MicroShift

    ${nid}=    Get MicroShift Cluster ID From Namespace
    ${fid}=    Get MicroShift Cluster ID From File
    Should Be Equal As Strings    ${nid}    ${fid}

Verify Missing Cluster ID Recovery
    [Documentation]    Verify that cluster ID file is correctly recreated
    ...    on the service restart after manual removing by a user.

    Remove Cluster ID File
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

Extract Sos Report
    [Documentation]    Extract Sos Report from the tar file
    [Arguments]    ${sos_report_tarfile}

    ${sos_report_dir}    ${rc}=    Execute Command
    ...    dirname ${sos_report_tarfile}
    ...    sudo=True    return_rc=True    return_stdout=True
    Log    ${sos_report_dir}
    Should Be Equal As Integers    0    ${rc}

    ${rc}=    Execute Command
    ...    tar xf ${sos_report_tarfile} -C ${sos_report_dir}
    ...    sudo=True    return_rc=True    return_stdout=False
    Should Be Equal As Integers    0    ${rc}

    ${sos_report_untared}    ${rc}=    Execute Command
    ...    find ${sos_report_dir} -type d -name "sosreport-microshift*"
    ...    sudo=True    return_rc=True    return_stdout=True
    Should Be Equal As Integers    0    ${rc}

    RETURN    ${sos_report_untared}

Get MicroShift Cluster ID From Sos Report
    [Documentation]    Read and return the Cluster ID from the kube-system namespace yaml description in the Sos Report.
    [Arguments]    ${sos_report_tarfile}

    ${sos_report_untared}=    Extract Sos Report    ${sos_report_tarfile}

    ${namespace_yaml_path}=    Set Variable
    ...    ${sos_report_untared}/sos_commands/microshift/namespaces/${CLUSTERID_NS}/${CLUSTERID_NS}.yaml
    ${clusterid}    ${rc}=    Execute Command
    ...    cat ${namespace_yaml_path} | sed -n 's/\\s\\suid:\\s//p'
    ...    sudo=True    return_rc=True    return_stdout=True
    Should Be Equal As Integers    0    ${rc}

    Should Not Be Empty    ${clusterid}
    RETURN    ${clusterid}

Tamper With Cluster ID File
    [Documentation]    Append invalid characters to the cluster ID file.
    ${rc}=    Execute Command
    ...    sed -i '$ s/$/123/' ${CLUSTERID_FILE}
    ...    sudo=True    return_rc=True    return_stdout=False

    Should Be Equal As Integers    0    ${rc}

Remove Cluster ID File
    [Documentation]    Append invalid characters to the cluster ID file.
    ${rc}=    Execute Command
    ...    rm -rf ${CLUSTERID_FILE}
    ...    sudo=True    return_rc=True    return_stdout=False

    Should Be Equal As Integers    0    ${rc}

Create Sos Report
    [Documentation]    Create a MicroShift Sos Report and return the tar file path

    ${rand_str}=    Generate Random String    4    [NUMBERS]
    ${sos_report_dir}=    Catenate    SEPARATOR=    /tmp/rf-test/sos-report_    ${rand_str}

    ${rc}=    Execute Command
    ...    mkdir -p ${sos_report_dir}
    ...    sudo=True    return_rc=True    return_stdout=False
    Should Be Equal As Integers    0    ${rc}

    ${rc}=    Execute Command
    ...    sos report --batch --all-logs --tmp-dir ${sos_report_dir} -p microshift -o logs
    ...    sudo=True    return_rc=True    return_stdout=False
    Should Be Equal As Integers    0    ${rc}

    ${sos_report_tarfile}    ${rc}=    Execute Command
    ...    find ${sos_report_dir} -type f -name "sosreport-microshift*.tar.xz"
    ...    sudo=True    return_rc=True    return_stdout=True
    Should Be Equal As Integers    0    ${rc}

    RETURN    ${sos_report_tarfile}
