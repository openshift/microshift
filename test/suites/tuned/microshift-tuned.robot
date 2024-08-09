*** Settings ***
Documentation       Tests for verification on MicroShift's Tuned profile

Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Test Cases ***
Checksum Of Current Profile Should Be Persisted
    [Documentation]    We expect that microshift-tuned keeps checksums of the profile and variables
    ...    in a separate file.

    SSHLibrary.File Should Exist    /var/lib/microshift-tuned.yaml
    ${checksums_contents}=    Command Should Work    cat /var/lib/microshift-tuned.yaml
    ${checksums}=    DataFormats.Yaml Parse    ${checksums_contents}
    Should Not Be Empty    ${checksums.profile_checksum}
    Should Not Be Empty    ${checksums.variables_checksum}

Profile Is Already Active But Cache File Is Missing
    [Documentation]    If profile is already active, but cache file is missing,
    ...    we expect microshift-tuned to reactivate it, reboot, and store the hashes.

    Command Should Work    rm -f /var/lib/microshift-tuned.yaml
    Restart MicroShift-Tuned Expecting Reboot
    SSHLibrary.File Should Exist    /var/lib/microshift-tuned.yaml

Variables Are Changed
    [Documentation]    When requested profile's variable are changed,
    ...    we expect microshift-tuned to reactivate the profile and reboot the host.

    ${old_hash}=    Command Should Work    cat /var/lib/microshift-tuned.yaml
    Command Should Work
    ...    sed -i 's/^offline_cpu_set=.*$/offline_cpu_set=/' /etc/tuned/microshift-baseline-variables.conf
    Restart MicroShift-Tuned Expecting Reboot
    ${new_hash}=    Command Should Work    cat /var/lib/microshift-tuned.yaml
    Should Not Be Equal    ${old_hash}    ${new_hash}

No Reboot If Not Allowed
    [Documentation]    If reboot_after_apply is False, then microshift-tuned should not reboot the host even if the
    ...    profile changed.

    Command Should Work
    ...    sed -i 's/^offline_cpu_set=.*$/offline_cpu_set=3/' /etc/tuned/microshift-baseline-variables.conf
    Command Should Work    sed -i 's/reboot_after_apply:.*/reboot_after_apply: False/' /etc/microshift/tuned.yaml
    Restart MicroShift-Tuned Not Expecting Reboot

Can Activate Any Tuned Profile
    [Documentation]    Verify that microshift-tuned will activate any given tuned profile.

    Command Should Work    sed -i 's/profile:.*/profile: virtual-guest/' /etc/microshift/tuned.yaml
    Restart MicroShift-Tuned Not Expecting Reboot
    ${active}=    Command Should Work    tuned-adm active
    Should Contain    ${active}    Current active profile: virtual-guest

MicroShift-Tuned Requires Config To Function
    [Documentation]    Verify that missing configuration will be fatal.
    Command Should Work    rm -f /etc/microshift/tuned.yaml
    Command Should Work    systemctl restart microshift-tuned.service
    Wait Until Keyword Succeeds    1m    10s
    ...    Systemctl Check Service SubState    microshift-tuned.service    failed


*** Keywords ***
Setup
    [Documentation]    Setup test for the test suite
    Login MicroShift Host
    # We don't need MicroShift service when testing microshift-tuned.service
    Stop MicroShift
    Disable MicroShift
    Command Should Work
    ...    cp /etc/tuned/microshift-baseline-variables.conf /etc/tuned/microshift-baseline-variables.conf.bak
    Command Should Work    cp /etc/microshift/tuned.yaml /etc/microshift/tuned.yaml.bak

Teardown
    [Documentation]    Teardown test after the test suite
    Command Should Work
    ...    mv /etc/tuned/microshift-baseline-variables.conf.bak /etc/tuned/microshift-baseline-variables.conf
    Command Should Work    mv /etc/microshift/tuned.yaml.bak /etc/microshift/tuned.yaml
    Enable MicroShift
    Restart MicroShift-Tuned Expecting Reboot
    Logout MicroShift Host

Restart MicroShift-Tuned Expecting Reboot
    [Documentation]    TODO
    ${bootid}=    Get Current Boot Id
    Command Should Work    systemctl restart microshift-tuned.service
    Wait Until Keyword Succeeds    5m    15s
    ...    System Should Be Rebooted    ${bootid}

Restart MicroShift-Tuned Not Expecting Reboot
    [Documentation]    TODO
    ${bootid}=    Get Current Boot Id
    Command Should Work    systemctl restart microshift-tuned.service
    Wait Until Keyword Succeeds    1m    10s
    ...    Systemctl Check Service SubState    microshift-tuned.service    dead
    ${rebooted}=    Is System Rebooted    ${bootid}
    Should Not Be True    ${rebooted}
