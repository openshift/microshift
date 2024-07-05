*** Comments ***
# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure the host is in a good state at the start of each test. We
# could have separated them and run them as separate scenarios, but
# did not want to spend the resources on a new VM.
#
# The "Install Source Version" test wants to be run on a system where
# MicroShift has never been installed before to ensure that all of the
# dependencies are installed automatically. The test teardown step
# removes those RPMs, and then "Upgrade From Previous Version"
# installs the _older_ version of MicroShift and tries to upgrade it.


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
${SOURCE_REPO_URL}      ${EMPTY}
# The version of microshift we expect to find in that repo
${TARGET_VERSION}       ${EMPTY}


*** Test Cases ***
Install Source Version
    [Documentation]    Install the version built from source
    Install Third Party Packages With Warnings
    Install MicroShift RPM Packages From Repo    ${SOURCE_REPO_URL}    ${TARGET_VERSION}
    Start MicroShift
    Wait For MicroShift
    [Teardown]    Clean Up Test

Upgrade From Previous Version
    [Documentation]    Install the previous version, then upgrade
    # Always install from system repo, because the scenario script
    # is enabling all the repositories needed.
    #
    # Ignore warnings when installing the previous version because we
    # know some of our older RPMs generate warnings. We care more
    # about warnings on the new RPM.
    Install MicroShift RPM Packages From System Repo
    ...    4.${PREVIOUS_MINOR_VERSION}.*
    ...    check_warnings=False

    ${version}=    MicroShift Version
    Should Be Equal As Integers    ${version.minor}    ${PREVIOUS_MINOR_VERSION}
    Start MicroShift
    Wait For MicroShift
    Install MicroShift RPM Packages From Repo    ${SOURCE_REPO_URL}    ${TARGET_VERSION}
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
    Should Not Be Empty    ${TARGET_VERSION}    TARGET_VERSION variable is required
    Login MicroShift Host
    System Should Not Be Ostree
    Pull Secret Should Be Installed

System Should Not Be Ostree
    [Documentation]    Make sure we run on a non-ostree system
    ${is_ostree}=    Is System OSTree
    Should Not Be True    ${is_ostree}

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

Install Third Party Packages With Warnings
    [Documentation]    Install these separately to avoid having warnings
    ...    show up in the warning check when installing MicroShift.
    Command Should Work    dnf install -y NetworkManager-ovs containers-common

Version Should Match
    [Documentation]    Compare the installed version against expectations
    [Arguments]    ${expected}
    ${version}=    MicroShift Version
    Should Be Equal As Strings    ${version.gitVersion}    ${expected}
