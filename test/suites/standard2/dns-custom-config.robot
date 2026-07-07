*** Settings ***
Documentation       Tests for custom DNS configuration (dns.configFile)

Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/dns.resource
Resource            ../../resources/hosts.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup Suite With Namespace
Suite Teardown      Teardown DNS Suite

Test Tags           slow


*** Variables ***
${FAKE_LISTEN_IP}           99.99.99.98
${CUSTOM_COREFILE_PATH}     /etc/microshift/dns/Corefile
${COREFILE_TEMPLATE}        ./assets/dns/Corefile.template
${HOSTNAME}                 ${EMPTY}

${CUSTOM_DNS_CONFIG}        SEPARATOR=\n
...                         ---
...                         dns:
...                         \ \ configFile: /etc/microshift/dns/Corefile


*** Test Cases ***
Custom DNS Config File Is Used By CoreDNS
    [Documentation]    Provide a valid custom Corefile and verify CoreDNS uses it
    ...    instead of the default template-rendered configuration.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Verify ConfigMap Uses Custom Corefile
    Resolve Host From Pod    ${HOSTNAME}
    [Teardown]    Teardown Custom Corefile

Runtime Reload Without MicroShift Restart
    [Documentation]    Modify the custom Corefile while MicroShift is running
    ...    and verify CoreDNS picks up the change without restart.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Resolve Host From Pod    ${HOSTNAME}
    ${updated_hostname}=    Generate Random HostName
    Update Corefile With New Host    ${updated_hostname}
    Wait Until Keyword Succeeds    20x    5s
    ...    ConfigMap Should Contain Hostname    ${updated_hostname}
    Resolve Host From Pod    ${updated_hostname}
    [Teardown]    Teardown Custom Corefile

ConfigMap Unchanged After Corefile Removal
    [Documentation]    Remove the custom Corefile while MicroShift is running
    ...    and verify the ConfigMap retains the last known content.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Verify ConfigMap Uses Custom Corefile
    ${cursor}=    Get Journal Cursor
    Remove Custom Corefile
    Pattern Should Appear In Log Output
    ...    ${cursor}    watched file.*was removed, keeping last known ConfigMap content
    ConfigMap Should Contain Hostname    ${HOSTNAME}
    [Teardown]    Teardown Custom Corefile

Cluster Local Resolution With Custom Corefile
    [Documentation]    Verify that a custom Corefile with the kubernetes plugin
    ...    still resolves cluster-local services correctly.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Resolve Host From Pod    kubernetes.default.svc.cluster.local
    [Teardown]    Teardown Custom Corefile

Atomic File Replacement Via Mv Updates ConfigMap
    [Documentation]    Replace the custom Corefile atomically using mv
    ...    (write temp file, then rename) and verify the ConfigMap is updated
    ...    without restarting MicroShift. Also verifies show-config --mode
    ...    effective reports the configured dns.configFile path.
    [Setup]    Setup Custom Corefile With Hosts Entry
    Resolve Host From Pod    ${HOSTNAME}
    ${new_hostname}=    Generate Random HostName
    ${corefile}=    Build Custom Corefile    ${new_hostname}    ${FAKE_LISTEN_IP}
    Upload String To File    ${corefile}    ${CUSTOM_COREFILE_PATH}.tmp
    Command Should Work    mv ${CUSTOM_COREFILE_PATH}.tmp ${CUSTOM_COREFILE_PATH}
    Wait Until Keyword Succeeds    20x    5s
    ...    ConfigMap Should Contain Hostname    ${new_hostname}
    Resolve Host From Pod    ${new_hostname}
    ${config}=    Show Config    effective
    Should Be Equal As Strings    ${config.dns.configFile}    ${CUSTOM_COREFILE_PATH}
    [Teardown]    Teardown Custom Corefile


*** Keywords ***
Build Custom Corefile
    [Documentation]    Build a Corefile from the template file, substituting
    ...    the hostname and IP placeholders.
    [Arguments]    ${hostname}    ${ip}
    ${template}=    OperatingSystem.Get File    ${COREFILE_TEMPLATE}
    VAR    ${COREFILE_HOSTNAME}=    ${hostname}
    VAR    ${COREFILE_HOST_IP}=    ${ip}
    ${corefile}=    Replace Variables    ${template}
    RETURN    ${corefile}

Setup Custom Corefile With Hosts Entry
    [Documentation]    Create a custom Corefile with a fake hostname entry,
    ...    upload it, configure MicroShift to use it, and restart.
    ${HOSTNAME}=    Generate Random HostName
    VAR    ${HOSTNAME}=    ${HOSTNAME}    scope=SUITE
    Add Fake IP To NIC    ${FAKE_LISTEN_IP}
    ${corefile}=    Build Custom Corefile    ${HOSTNAME}    ${FAKE_LISTEN_IP}
    Create Remote Dir For Path    ${CUSTOM_COREFILE_PATH}
    Upload String To File    ${corefile}    ${CUSTOM_COREFILE_PATH}
    Drop In MicroShift Config    ${CUSTOM_DNS_CONFIG}    20-dns-custom
    Restart MicroShift

Update Corefile With New Host
    [Documentation]    Add a new hosts entry to the custom Corefile without
    ...    restarting MicroShift. The file watcher and CoreDNS reload plugin
    ...    should pick up the change.
    [Arguments]    ${hostname}
    ${corefile}=    Build Custom Corefile    ${hostname}    ${FAKE_LISTEN_IP}
    Upload String To File    ${corefile}    ${CUSTOM_COREFILE_PATH}

Verify ConfigMap Uses Custom Corefile
    [Documentation]    Verify the dns-default ConfigMap contains custom content
    ...    rather than the default template-rendered Corefile.
    Wait Until Keyword Succeeds    20x    5s
    ...    ConfigMap Should Contain Hostname    ${HOSTNAME}

ConfigMap Should Contain Hostname
    [Documentation]    Check the dns-default ConfigMap Corefile data for the hostname
    [Arguments]    ${hostname}
    ${corefile_data}=    Oc Get JsonPath
    ...    configmap    openshift-dns    dns-default    .data.Corefile
    Should Contain    ${corefile_data}    ${hostname}

Teardown DNS Suite
    [Documentation]    Restart MicroShift to restore clean state after the last test
    ...    (per-test teardowns skip the restart), then tear down namespace.
    Restart MicroShift
    Teardown Suite With Namespace

Teardown Custom Corefile
    [Documentation]    Remove custom Corefile, drop-in config, and fake IP without restarting.
    ...    The next test's setup will restart MicroShift.
    Run Keywords
    ...    Remove Drop In MicroShift Config    20-dns-custom
    ...    AND    Remove Custom Corefile
    ...    AND    Remove Fake IP From NIC    ${FAKE_LISTEN_IP}

Remove Custom Corefile
    [Documentation]    Remove the custom Corefile from the host
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rm -f ${CUSTOM_COREFILE_PATH} ${CUSTOM_COREFILE_PATH}.tmp
    ...    sudo=True    return_rc=True    return_stdout=True    return_stderr=True
    Should Be Equal As Integers    0    ${rc}
