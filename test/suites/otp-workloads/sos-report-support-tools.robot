*** Settings ***
Documentation       Tests verifying MicroShift SOS report output can be consumed
...                 by the omc and omg support tools. Both tools are pre-installed
...                 locally by fetch_tools.sh (omc and omg) and run against
...                 a sosreport downloaded from the remote host.
...
...                 Ported from Polarion: OCP-61971
...                 USHIFT-7221

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/common.resource
Resource            ../../resources/sos-report.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow


*** Variables ***
${SOS_REPORT_BASE_DIR}      /tmp/rf-test/sos-report-support-tools
${OMC_BINARY}               ${EXECDIR}/../_output/robotenv/omc
${OMG_BINARY}               ${EXECDIR}/../_output/robotenv/omg-venv/bin/omg
${MUST_GATHER_PATH}         ${EMPTY}
${LOCAL_SOS_DIR}            ${EMPTY}


*** Test Cases ***
Verify Omc Can Consume SOS Report
    [Documentation]    Verify that the omc tool can consume the SOS report
    ...    output under sos_commands/microshift and list pods, nodes,
    ...    projects, and retrieve logs.
    ...    OCP-61971
    Verify Support Tool Can Consume Sos Report    ${OMC_BINARY}

Verify Omg Can Consume SOS Report
    [Documentation]    Verify that the omg tool can consume the SOS report
    ...    output under sos_commands/microshift and list pods, nodes,
    ...    projects, and retrieve logs.
    ...    OCP-61971
    Verify Support Tool Can Consume Sos Report    ${OMG_BINARY}


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    OperatingSystem.File Should Exist    ${OMC_BINARY}    omc not found - run fetch_tools.sh omc
    OperatingSystem.File Should Exist    ${OMG_BINARY}    omg not found - run fetch_tools.sh omg
    Login MicroShift Host
    ${sos_report_tarfile}=    Create Sos Report
    ${sos_report_extracted}=    Extract Sos Report    ${sos_report_tarfile}
    ${local_dir}=    Download Must Gather Data    ${sos_report_extracted}
    VAR    ${LOCAL_SOS_DIR}=    ${local_dir}    scope=SUITE
    VAR    ${MUST_GATHER_PATH}=    ${local_dir}/sos_commands/microshift    scope=SUITE
    OperatingSystem.Directory Should Exist    ${MUST_GATHER_PATH}

Teardown
    [Documentation]    Test suite teardown
    Run Keyword And Ignore Error
    ...    OperatingSystem.Remove Directory    ${LOCAL_SOS_DIR}    recursive=True
    Cleanup Sos Report Directory
    Logout MicroShift Host

Download Must Gather Data
    [Documentation]    Download the microshift sos_commands directory from the remote host
    ...    to a local temp directory and return the local path.
    [Arguments]    ${remote_extracted_dir}
    Command Should Work    chmod -R a+rX ${remote_extracted_dir}
    ${local_dir}=    Create Random Temp Directory
    VAR    ${local_must_gather}=    ${local_dir}/sos_commands/microshift
    OperatingSystem.Create Directory    ${local_must_gather}
    SSHLibrary.Get Directory
    ...    ${remote_extracted_dir}/sos_commands/microshift    ${local_must_gather}    recursive=True
    OperatingSystem.Directory Should Not Be Empty    ${local_must_gather}
    RETURN    ${local_dir}

Verify Support Tool Can Consume Sos Report
    [Documentation]    Verify that the given support tool can consume the SOS report
    ...    and list pods, nodes, projects, and retrieve logs.
    [Arguments]    ${tool}
    Local Command Should Work    ${tool} use ${MUST_GATHER_PATH}
    Verify Tool Can List Resources    ${tool}
    Verify Tool Can Retrieve Logs    ${tool}    openshift-dns
    ${projects_output}=    Local Command Should Work    ${tool} projects
    Should Not Be Empty    ${projects_output}

Verify Tool Can List Resources
    [Documentation]    Verify the tool can list pods, nodes, and projects.
    [Arguments]    ${tool}
    Local Command Should Work    ${tool} project
    ${pods_output}=    Local Command Should Work    ${tool} get pods -A
    Should Contain    ${pods_output}    openshift-dns
    ${nodes_output}=    Local Command Should Work    ${tool} get nodes
    Should Not Be Empty    ${nodes_output}
    Local Command Should Work    ${tool} get node --show-labels
    Local Command Should Work    ${tool} get node -o wide
    ${pods_json}=    Local Command Should Work    ${tool} get pods -A -o json
    Should Contain    ${pods_json}    "items"

Verify Tool Can Retrieve Logs
    [Documentation]    Verify the tool can retrieve logs for a pod in the given namespace.
    [Arguments]    ${tool}    ${namespace}
    ${pod_name}=    Get Pod Name From Tool    ${tool}    ${namespace}
    ${container_name}=    Get Container Name From Tool    ${tool}    ${namespace}
    ${logs_output}=    Local Command Should Work
    ...    ${tool} logs ${pod_name} -c ${container_name} -n ${namespace}
    Should Not Be Empty    ${logs_output}

Get Pod Name From Tool
    [Documentation]    Get the first pod name from the specified namespace using the given tool.
    [Arguments]    ${tool}    ${namespace}
    ${output}=    Local Command Should Work
    ...    ${tool} get pods -n ${namespace} | tail -n +2 | head -1 | awk '{print $1}'
    Should Not Be Empty    ${output}
    RETURN    ${output}

Get Container Name From Tool
    [Documentation]    Get the first container name from the first pod in the specified namespace.
    [Arguments]    ${tool}    ${namespace}
    ${output}=    Local Command Should Work
    ...    ${tool} get pods -n ${namespace} -o json | python3 -c "import sys,json; print(json.load(sys.stdin)['items'][0]['spec']['containers'][0]['name'])"
    Should Not Be Empty    ${output}
    RETURN    ${output}
