*** Settings ***
Documentation       Keywords for creating, extracting, and cleaning up
...                 MicroShift SOS reports.

Resource            common.resource
Resource            microshift-host.resource


*** Keywords ***
Create Sos Report
    [Documentation]    Create a MicroShift SOS Report using microshift-sos-report
    ...    and return the tar file path.
    ...    Requires ${SOS_REPORT_BASE_DIR} to be set in the importing suite.
    ${rand_str}=    Generate Random String    4    [NUMBERS]
    ${sos_report_dir}=    Catenate    SEPARATOR=${EMPTY}    ${SOS_REPORT_BASE_DIR}_    ${rand_str}
    Command Should Work    mkdir -p ${sos_report_dir}
    ${output}=    Command Should Work
    ...    /usr/bin/microshift-sos-report --tmp-dir ${sos_report_dir}
    Should Contain    ${output}    Your sos report has been generated and saved in
    ${sos_report_tarfile}=    Command Should Work
    ...    find ${sos_report_dir} -type f -name "sosreport-*.tar.xz"
    Should Not Be Empty    ${sos_report_tarfile}
    RETURN    ${sos_report_tarfile}

Create Sos Report With Profile
    [Documentation]    Create a SOS report with the specified profile and return the tar file path.
    ...    Requires ${SOS_REPORT_BASE_DIR} to be set in the importing suite.
    [Arguments]    ${profile}
    ${rand_str}=    Generate Random String    4    [NUMBERS]
    ${sos_report_dir}=    Catenate    SEPARATOR=${EMPTY}    ${SOS_REPORT_BASE_DIR}_    ${rand_str}
    Command Should Work    mkdir -p ${sos_report_dir}
    Command Should Work
    ...    sos report --batch --clean --all-logs --profile ${profile} --tmp-dir ${sos_report_dir}
    ${sos_report_tarfile}=    Command Should Work
    ...    find ${sos_report_dir} -type f -name "sosreport-*.tar.xz"
    Should Not Be Empty    ${sos_report_tarfile}
    RETURN    ${sos_report_tarfile}

Extract Sos Report
    [Documentation]    Extract SOS report from the tar file and return the extracted directory.
    [Arguments]    ${sos_report_tarfile}
    ${sos_report_dir}=    Command Should Work    dirname ${sos_report_tarfile}
    Command Should Work    tar xf ${sos_report_tarfile} -C ${sos_report_dir}
    ${sos_report_untared}=    Command Should Work
    ...    find ${sos_report_dir} -maxdepth 1 -type d -name "sosreport-*"
    Should Not Be Empty    ${sos_report_untared}
    RETURN    ${sos_report_untared}

Cleanup Sos Report Directory
    [Documentation]    Remove temporary SOS report files created during tests.
    ...    Requires ${SOS_REPORT_BASE_DIR} to be set in the importing suite.
    Command Should Work    rm -rf ${SOS_REPORT_BASE_DIR}_*
    Command Should Work    rm -rf /var/tmp/sosreport-*
