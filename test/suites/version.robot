*** Settings ***
Documentation       Tests related to the version of MicroShift

Resource            ../resources/common.resource
Resource            ../resources/oc.resource
Resource            ../resources/microshift-process.resource
Library             Collections
Library             ../resources/DataFormats.py

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
ConfigMap Contents
    [Documentation]    Check the version of the server

    ${configmap}=    Oc Get    configmap    kube-public    microshift-version
    Should Be Equal As Integers    ${configmap.data.major}    ${MAJOR_VERSION}
    Should Be Equal As Integers    ${configmap.data.minor}    ${MINOR_VERSION}
    Should Be Equal As Integers    ${configmap.data.patch}    ${PATCH_VERSION}

CLI Output
    [Documentation]    Check the version reported by the process

    ${version}=    MicroShift Version
    Should Be Equal As Integers    ${version.major}    ${MAJOR_VERSION}
    Should Be Equal As Integers    ${version.minor}    ${MINOR_VERSION}
    Should Be Equal As Integers    ${version.patch}    ${PATCH_VERSION}
    Should Start With    ${version.gitVersion}    ${Y_STREAM}

ConfigMap Matches CLI
    [Documentation]    Ensure the ConfigMap is being updated based on the actual binary version

    ${configmap}=    Oc Get    configmap    kube-public    microshift-version
    ${cli}=    MicroShift Version
    Should Be Equal    ${configmap.data.version}    ${cli.gitVersion}

Metadata File Contents
    [Documentation]    Ensure the metadata file contents match the expected version.

    ${contents}=    Execute Command
    ...    cat /var/lib/microshift/version
    ...    sudo=True    return_rc=False
    Should Be Equal    ${contents}    ${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Read Expected Versions

Teardown
    [Documentation]    Test suite teardown
    Remove Kubeconfig
    Logout MicroShift Host

Read Expected Versions    # robocop: disable=too-many-calls-in-keyword
    [Documentation]    Read ../Makefile.version.aarch64.var to find the expected versions
    ...    Sets suite variables FULL_VERSION, MAJOR_VERSION, MINOR_VERSION, and Y_STREAM based on
    ...    the content.

    # The file contains content in this format:
    #
    #    OCP_VERSION := 4.14.0-0.nightly-arm64-2023-05-04-012046

    ${unparsed}=    OperatingSystem.Get File    ../Makefile.version.aarch64.var

    # 4.14.0-0.nightly-arm64-2023-05-04-012046
    ${version_full_raw}=    Fetch From Right    ${unparsed}    :=
    ${version_full}=    Strip String    ${version_full_raw}
    Set Suite Variable    \${FULL_VERSION}    ${version_full}
    # 4.14.0
    ${version_short_matches}=    Get Regexp Matches    ${version_full}    ^(\\d+.\\d+.\\d+)
    ${version_short_parts}=    Split String    ${version_short_matches}[0]    .
    # 4
    Set Suite Variable    \${MAJOR_VERSION}    ${version_short_parts}[0]
    # 14
    Set Suite Variable    \${MINOR_VERSION}    ${version_short_parts}[1]
    # 0
    Set Suite Variable    \${PATCH_VERSION}    ${version_short_parts}[2]
    # 4.14
    ${ystream}=    Format String    {}.{}    ${MAJOR_VERSION}    ${MINOR_VERSION}
    Set Suite Variable    \${Y_STREAM}    ${ystream}
