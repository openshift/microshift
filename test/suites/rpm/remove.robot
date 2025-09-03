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
Resource            ../../resources/microshift-process.resource
Library             Collections
Library             SSHLibrary

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    rpm-based-system    slow


*** Test Cases ***
Remove
    [Documentation]    Clean up an installed MicroShift instance
    Cleanup MicroShift
    Uninstall MicroShift RPM Packages


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    System Should Not Be Ostree
    Pull Secret Should Be Installed

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

System Should Not Be Ostree
    [Documentation]    Make sure we run on a non-ostree system
    ${is_ostree}=    Is System OSTree
    Should Not Be True    ${is_ostree}

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
