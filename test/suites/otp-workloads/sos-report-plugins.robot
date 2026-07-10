*** Settings ***
Documentation       Tests verifying MicroShift SOS report plugins are listed and enabled,
...                 SOS report collects MicroShift config and OVN related information,
...                 and greenboot logs reside in a separate directory.
...
...                 Ported from openshift-tests-private:
...                 Critical-60924, Critical-60929, High-68257, High-60930, High-68256

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/common.resource
Resource            ../../resources/sos-report.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow


*** Variables ***
${SOS_REPORT_BASE_DIR}      /tmp/rf-test/sos-report-plugins


*** Test Cases ***
Verify SOS Report Lists Enabled MicroShift Plugins
    [Documentation]    Verify sos report -l lists both microshift and
    ...    microshift_ovn plugins and they are enabled.
    ...    OCP-60924
    ${output}=    Command Should Work    sos report -l
    Should Contain    ${output}    microshift
    Should Contain    ${output}    microshift_ovn
    Plugin Should Be Enabled    ${output}    microshift
    Plugin Should Be Enabled    ${output}    microshift_ovn

Verify SOS Report Collects MicroShift And OVN Information
    [Documentation]    Verify running sos report with microshift profile
    ...    collects microshift config, oc adm inspect output,
    ...    and OVN related information.
    ...    OCP-60929, OCP-68257
    ${sos_report_tarfile}=    Create Sos Report
    ${sos_report_extracted}=    Extract Sos Report    ${sos_report_tarfile}

    ${sos_text}=    Command Should Work    cat ${sos_report_extracted}/sos_reports/sos.txt
    Should Contain    ${sos_text}    oc adm inspect
    Should Contain    ${sos_text}    ovs-appctl -t /var/run/ovn/ovn-controller.*.ctl ct-zone-list

    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/etc/microshift
    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/etc/microshift/manifests
    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/etc/microshift/manifests.d

    ${system_config}=    Command Should Work
    ...    ls -l /etc/microshift | awk '{print $NF}' | tail -n+2
    ${report_config}=    Command Should Work
    ...    ls -l ${sos_report_extracted}/etc/microshift | awk '{print $NF}' | tail -n+2
    Should Be Equal As Strings    ${report_config}    ${system_config}
    ...    msg=Not all microshift config files are collected in sosreport

    ${system_manifests}=    Command Should Work    ls /etc/microshift/manifests
    ${report_manifests}=    Command Should Work
    ...    ls ${sos_report_extracted}/etc/microshift/manifests
    Should Be Equal As Strings    ${report_manifests}    ${system_manifests}
    ...    msg=Files inside /etc/microshift/manifests directory does not match between system and sosreport

    ${system_manifests_d}=    Command Should Work    ls /etc/microshift/manifests.d/
    ${report_manifests_d}=    Command Should Work
    ...    ls ${sos_report_extracted}/etc/microshift/manifests.d/
    Should Be Equal As Strings    ${report_manifests_d}    ${system_manifests_d}
    ...    msg=Files inside /etc/microshift/manifests.d directory does not match between system and sosreport

Verify SOS Report Collects Data When MicroShift Is Stopped
    [Documentation]    Verify that sos report -p microshift still collects
    ...    config files, journal logs, and OVN data when MicroShift
    ...    service is stopped, but does not collect oc adm inspect data.
    [Tags]    restart

    Stop MicroShift

    ${rand_str}=    Generate Random String    4    [NUMBERS]
    ${sos_report_dir}=    Catenate    SEPARATOR=${EMPTY}    ${SOS_REPORT_BASE_DIR}_    ${rand_str}
    Command Should Work    mkdir -p ${sos_report_dir}
    Command Should Work
    ...    sos report --batch --tmp-dir ${sos_report_dir} -p microshift
    ${sos_report_tarfile}=    Command Should Work
    ...    find ${sos_report_dir} -type f -name "sosreport-*.tar.xz"
    Should Not Be Empty    ${sos_report_tarfile}
    ${sos_report_extracted}=    Extract Sos Report    ${sos_report_tarfile}

    ${sos_text}=    Command Should Work    cat ${sos_report_extracted}/sos_reports/sos.txt
    Should Not Contain    ${sos_text}    oc adm inspect

    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/etc/microshift
    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/etc/microshift/manifests
    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/etc/microshift/manifests.d
    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/sos_commands/microshift
    Verify Remote Directory Exists With Sudo    ${sos_report_extracted}/sos_commands/microshift_ovn

    ${microshift_cmds}=    Command Should Work
    ...    ls ${sos_report_extracted}/sos_commands/microshift
    Should Contain    ${microshift_cmds}    journalctl_--no-pager_--unit_microshift
    Should Contain    ${microshift_cmds}    journalctl_--no-pager_--unit_microshift-etcd.scope
    Should Contain    ${microshift_cmds}    microshift_version
    Should Contain    ${microshift_cmds}    systemctl_status_microshift

    [Teardown]    Start MicroShift

Verify Greenboot Logs In SOS Report
    [Documentation]    Verify greenboot logs reside in a separate directory
    ...    and are easy to access in the SOS report.
    ...    OCP-68256
    ${sos_report_tarfile}=    Create Sos Report With Profile    system
    ${sos_report_extracted}=    Extract Sos Report    ${sos_report_tarfile}

    ${greenboot_files}=    Command Should Work
    ...    ls -l ${sos_report_extracted}/sos_commands/greenboot
    Should Contain    ${greenboot_files}    journalctl_--no-pager_--unit_greenboot-healthcheck
    Should Contain    ${greenboot_files}    systemctl_status_greenboot-healthcheck

    ${greenboot_conf}=    Command Should Work
    ...    ls -l ${sos_report_extracted}/etc/greenboot
    Should Contain    ${greenboot_conf}    greenboot.conf


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Cleanup Sos Report Directory
    Logout MicroShift Host

Plugin Should Be Enabled
    [Documentation]    Verify that the named plugin is not disabled in the sos report -l output.
    [Arguments]    ${output}    ${plugin_name}
    Should Not Match Regexp    ${output}    ${plugin_name}\\s+.*disabled
