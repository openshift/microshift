*** Settings ***
Documentation       Verify that MicroShift only introduces expected CIS Level 2
...                 changes on a hardened RHEL system. Runs an OpenSCAP scan
...                 before and after installing MicroShift and asserts that
...                 every new failure belongs to a known set of rules inherent
...                 to Kubernetes container networking.

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/microshift-rpm.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           slow


*** Variables ***
${USHIFT_HOST}                  ${EMPTY}
${USHIFT_USER}                  ${EMPTY}
${SOURCE_REPO_URL}              ${EMPTY}
${TARGET_VERSION}               ${EMPTY}
${API_PORT}                     6443
${OSCAP_BASELINE_FILE}          /tmp/cis-baseline-results.xml
${OSCAP_POST_FILE}              /tmp/cis-post-results.xml
${OSCAP_REPORT_FILE}            /tmp/cis-post-report.html
${OSCAP_PROFILE}                xccdf_org.ssgproject.content_profile_cis
${SCAP_DS_FILE}                 /usr/share/xml/scap/ssg/content/ssg-rhel9-ds.xml
${CIS_REQUIREMENTS_FILE}        cis-requirements-el9.yml
${CIS_HARDEN_FILE}              cis-harden-el9.yml

# Rules that MicroShift is known to cause. Any new failure outside this
# set means MicroShift introduced an unexpected CIS regression.
@{KNOWN_MICROSHIFT_RULES}
...                             sysctl_net_ipv4_ip_forward
...                             sysctl_net_ipv6_conf_all_forwarding
...                             sysctl_net_ipv4_conf_all_forwarding
...                             sysctl_net_ipv4_conf_default_forwarding
...                             sysctl_net_ipv6_conf_default_forwarding
...                             file_permissions_ungroupowned
...                             no_files_unowned_by_user
...                             no_files_or_dirs_ungroupowned
...                             no_files_or_dirs_unowned_by_user
...                             dir_perms_world_writable_sticky_bits
...                             file_permissions_unauthorized_world_writable
...                             audit_rules_privileged_commands


*** Test Cases ***
MicroShift Only Introduces Expected CIS Changes
    [Documentation]    Compare pre- and post-MicroShift OpenSCAP scans.
    ...    Every new failure must be in the known set of MicroShift-caused rules.
    ${baseline}=    Get Failing Rule IDs    ${OSCAP_BASELINE_FILE}
    ${post}=    Get Failing Rule IDs    ${OSCAP_POST_FILE}
    ${delta}=    Get New Failures    ${baseline}    ${post}
    Log    Baseline failures: ${baseline}
    Log    Post-install failures: ${post}
    Log    New failures introduced by MicroShift: ${delta}
    Should Only Contain Expected Rules    ${delta}

All Pods Are Running After CIS Hardening
    [Documentation]    Verify all MicroShift pods are running on the hardened system
    All Pods Should Be Running

Smoke Test With Route
    [Documentation]    Deploy hello-microshift and expose via route to verify networking on hardened system
    [Setup]    Setup Smoke Test
    Wait Until Keyword Succeeds    10x    6s
    ...    Access Hello MicroShift Success    ${HTTP_PORT}
    [Teardown]    Teardown Smoke Test


*** Keywords ***
Setup
    [Documentation]    Harden the system, run a baseline scan, install MicroShift,
    ...    run a post-install scan, and prepare for functional tests.
    Check Required Env Variables
    Login MicroShift Host
    Harden And Scan Baseline
    Install And Enable MicroShift
    Reboot And Reconnect
    Configure Firewall After Hardening
    Wait Until Greenboot Health Check Exited
    Run CIS Scan    ${OSCAP_POST_FILE}    oscap_report=${OSCAP_REPORT_FILE}
    Setup Kubeconfig

Teardown
    [Documentation]    Archive scan artifacts and close SSH
    Run Keyword And Ignore Error
    ...    Command Should Work    chmod 644 ${OSCAP_BASELINE_FILE} ${OSCAP_POST_FILE} ${OSCAP_REPORT_FILE}
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${OSCAP_BASELINE_FILE}    ${OUTPUTDIR}/cis-baseline-results.xml
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${OSCAP_POST_FILE}    ${OUTPUTDIR}/cis-post-results.xml
    Run Keyword And Ignore Error
    ...    SSHLibrary.Get File    ${OSCAP_REPORT_FILE}    ${OUTPUTDIR}/cis-post-report.html
    Run Keyword And Ignore Error
    ...    Logout MicroShift Host

Harden And Scan Baseline
    [Documentation]    Apply CIS hardening, reboot, and run the baseline scan.
    Apply CIS Hardening
    Reboot And Reconnect
    Run CIS Scan    ${OSCAP_BASELINE_FILE}

Apply CIS Hardening
    [Documentation]    Install prerequisites, upload assets, and run the CIS hardening playbook
    Command Should Work    dnf install -y openscap-scanner scap-security-guide ansible-core
    Command Should Work    ansible-galaxy collection install 'community.general:<10.0.0' ansible.posix
    SSHLibrary.Put File    ${CURDIR}/../../assets/cis/${CIS_REQUIREMENTS_FILE}    /tmp/cis-requirements.yml
    SSHLibrary.Put File    ${CURDIR}/../../assets/cis/${CIS_HARDEN_FILE}    /tmp/cis-harden.yml
    Command Should Work    ansible-galaxy role install -r /tmp/cis-requirements.yml
    Run Hardening Playbook

Reboot And Reconnect
    [Documentation]    Reboot the host and wait for SSH to come back.
    ...    Uses boot ID change detection only, without asserting the
    ...    exact reboot count, because CIS hardening can trigger
    ...    additional automatic reboots (crypto-policy changes, etc).
    ${old_bootid}=    Get Current Boot Id
    SSHLibrary.Start Command    reboot    sudo=True
    Wait Until Keyword Succeeds    5m    10s
    ...    System Should Be Rebooted    ${old_bootid}

Install And Enable MicroShift
    [Documentation]    Install MicroShift after CIS hardening so the scan
    ...    reveals what MicroShift changes to CIS compliance.
    ...    The local source repo is created after hardening, so the
    ...    repo-level gpgcheck=0 is preserved (not overwritten by CIS).
    Should Not Be Empty    ${SOURCE_REPO_URL}    SOURCE_REPO_URL variable is required
    Install MicroShift RPM Packages From Repo    ${SOURCE_REPO_URL}    ${TARGET_VERSION}
    Command Should Work    systemctl enable microshift

Run Hardening Playbook
    [Documentation]    Run the CIS hardening playbook in the foreground over SSH.
    ...    The playbook runs update-crypto-policies which restarts sshd,
    ...    but existing SSH connections survive the service restart.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ansible-playbook -c local /tmp/cis-harden.yml
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True    timeout=1800
    Log    ${stdout}
    Should Be Equal As Integers    ${rc}    0    msg=Hardening playbook failed (rc=${rc}): ${stderr}

Configure Firewall After Hardening
    [Documentation]    Open firewall ports required by MicroShift and the test suite (via SSH)
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=22/tcp
    Configure Firewall Trusted Sources
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=80/tcp
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=443/tcp
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=5353/udp
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=${API_PORT}/tcp
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=30000-32767/tcp
    Command Should Work    firewall-cmd --permanent --zone=public --add-port=30000-32767/udp
    Command Should Work    firewall-cmd --reload

Configure Firewall Trusted Sources
    [Documentation]    Allow pod and service network traffic through the firewall
    Command Should Work    firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
    Command Should Work    firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
    Command Should Work    firewall-cmd --permanent --zone=trusted --add-source=fd01::/48

Run CIS Scan
    [Documentation]    Execute the OpenSCAP CIS Level 2 scan and save results.
    [Arguments]    ${results_file}    ${oscap_report}=${EMPTY}
    ${report_arg}=    Set Variable If    "${oscap_report}" != "${EMPTY}"
    ...    --report ${oscap_report}    ${EMPTY}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    oscap xccdf eval --profile ${OSCAP_PROFILE} --results ${results_file} ${report_arg} ${SCAP_DS_FILE}
    ...    sudo=True
    ...    return_rc=True
    ...    return_stdout=True
    ...    return_stderr=True
    Should Be True    ${rc} == 0 or ${rc} == 2
    ...    OpenSCAP scan failed with unexpected return code ${rc}: ${stderr}

Get Failing Rule IDs
    [Documentation]    Parse an OSCAP results XML file and return the list of
    ...    failing rule IDs (short names without the SSG prefix).
    [Arguments]    ${results_file}
    Verify Remote File Exists With Sudo    ${results_file}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    python3 -c "import xml.etree.ElementTree as ET; ns={'x':'http://checklists.nist.gov/xccdf/1.2'}; tree=ET.parse('${results_file}'); rules=[rr.get('idref').replace('xccdf_org.ssgproject.content_rule_','') for rr in tree.findall('.//x:rule-result',ns) if rr.find('x:result',ns).text=='fail']; print('\\n'.join(sorted(rules)))"
    ...    sudo=True
    ...    return_rc=True
    ...    return_stdout=True
    ...    return_stderr=True
    Should Be Equal As Integers    ${rc}    0    msg=Failed to parse results: ${stderr}
    @{rules}=    Split String    ${stdout.strip()}    \n
    RETURN    ${rules}

Get New Failures
    [Documentation]    Return rule IDs that failed in post but not in baseline.
    [Arguments]    ${baseline}    ${post}
    ${delta}=    Evaluate    sorted(set($post) - set($baseline))
    RETURN    ${delta}

Should Only Contain Expected Rules
    [Documentation]    Assert every rule in the delta is in the known MicroShift set.
    ...    Logs each unexpected rule and fails if any exist.
    [Arguments]    ${delta}
    VAR    @{unexpected}=    @{EMPTY}
    FOR    ${rule}    IN    @{delta}
        ${is_known}=    Evaluate    $rule in $KNOWN_MICROSHIFT_RULES
        IF    not ${is_known}    Append To List    ${unexpected}    ${rule}
    END
    Log    Known MicroShift rules in delta: ${delta}
    IF    len($unexpected) > 0
        Fail    MicroShift introduced unexpected CIS failures: ${unexpected}
    END

Setup Smoke Test
    [Documentation]    Create hello-microshift pod and expose via route
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=TEST
    Create Hello MicroShift Pod
    Expose Hello MicroShift
    Oc Expose    svc hello-microshift --hostname hello-microshift.cluster.local -n ${NAMESPACE}

Teardown Smoke Test
    [Documentation]    Clean up smoke test resources
    Run Keyword And Ignore Error    Oc Delete    route/hello-microshift -n ${NAMESPACE}
    Run Keyword And Ignore Error    Delete Hello MicroShift Pod And Service
    Run Keyword And Ignore Error    Remove Namespace    ${NAMESPACE}
