*** Comments ***
# NOTE: These tests rely on being run in order.


*** Settings ***
Documentation       Tests related to installing MicroShift on a non-ostree system

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-rpm.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-data.resource
Resource            ../../resources/ostree-health.resource
Library             Collections
Library             SSHLibrary

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    rpm-based-system    slow


*** Variables ***
# The URL to the repo with the version of microshift build from source
${SOURCE_REPO_URL}          ${EMPTY}
# The version of microshift we expect to find in that repo
${TARGET_VERSION}           ${EMPTY}
# The version we should use when enabling the extra repos with dependencies like oc
${DEPENDENCY_VERSION}       ${EMPTY}


*** Test Cases ***
Install Source Version
    [Documentation]    Install the version built from source
    Install NetworkManager-ovs
    Install MicroShift RPM Packages From Repo    ${SOURCE_REPO_URL}
    Version Should Match    ${TARGET_VERSION}
    Start MicroShift
    Wait For MicroShift
    [Teardown]    Clean Up Test

Upgrade From Previous Version
    [Documentation]    Install the previous version, then upgrade
    Install MicroShift RPM Packages From System Repo
    ${version}=    MicroShift Version
    Should Be Equal As Integers    ${version.minor}    ${PREVIOUS_MINOR_VERSION}
    Start MicroShift
    Wait For MicroShift
    Install MicroShift RPM Packages From Repo    ${SOURCE_REPO_URL}
    Version Should Match    ${TARGET_VERSION}
    # Restart MicroShift
    Reboot MicroShift Host
    # Health of the system is implicitly checked by greenboot successful exit
    Wait Until Greenboot Health Check Exited
    [Teardown]    Clean Up Test


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${SOURCE_REPO_URL}    SOURCE_REPO_URL variable is required
    Should Not Be Empty    ${DEPENDENCY_VERSION}    DEPENDENCY_VERSION variable is required
    Should Not Be Empty    ${TARGET_VERSION}    TARGET_VERSION variable is required
    Login MicroShift Host
    System Should Not Be Ostree
    Enable MicroShift Dependency Repositories
    Pull Secret Should Be Installed

System Should Not Be Ostree
    [Documentation]    Make sure we run on a non-ostree system
    ${is_ostree}=    Is System OSTree
    Should Not Be True    ${is_ostree}

Enable MicroShift Dependency Repositories
    [Documentation]    Add the repositories with dependencies like oc, crio, etc.
    ...    The scenario script is responsible for creating the VM with a
    ...    subscription enabled.
    ${uname}=    Command Should Work    uname -m
    Command Should Work
    ...    subscription-manager repos --enable rhocp-${DEPENDENCY_VERSION}-for-rhel-9-${uname}-rpms --enable fast-datapath-for-rhel-9-${uname}-rpms

Pull Secret Should Be Installed
    [Documentation]    Check that the kickstart file installed a pull secret for us
    # Check that the file exists without actually saving the output so
    # we don't have to clean the logs with the secret.
    ${rc}=    SSHLibrary.Execute Command
    ...    cat /etc/crio/openshift-pull-secret
    ...    sudo=True
    ...    return_rc=True
    ...    return_stdout=False
    ...    return_stderr=False
    Should Be Equal As Integers    0    ${rc}

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Clean Up Test
    [Documentation]    Clean up an installed MicroShift instance
    Cleanup MicroShift
    Uninstall MicroShift RPM Packages

Install NetworkManager-ovs
    [Documentation]    We install this separately to avoid having warnings
    ...    it may report show up in the warning check when installing microshift.
    Command Should Work    dnf install -y NetworkManager-ovs

Version Should Match
    [Documentation]    Compare the installed version against expectations
    [Arguments]    ${expected}
    ${version}=    MicroShift Version
    Should Be Equal As Strings    ${version.gitVersion}    ${expected}
