*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

#Suite Setup         Setup
#Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}              ${EMPTY}
${USHIFT_USER}              ${EMPTY}

${FAKE_NEXT_MINOR_REF}      ${EMPTY}

*** Test Cases ***
Update
    [Documentation]     Upgrade test instance from the PR under test to main:HEAD and verify cluster consistency and stability
    Checkout Branch
    # find custom rpms=find _output/rpmbuild -name *.rpm -exec readlink -f {} ';'  | tr '\n' ','
    # Checkout main:HEAD
    # Build new OS layer
    # Pull OS layer from host



*** Keywords ***

# Build main:HEAD os-layer
Checkout Branch
    [Documentation]     checkout the specified branch on the VM host
    ${result}=          Run Process         ls      stderr=STDOUT
    Log                 ${result.stdout}

#Setup
#    [Documentation]     Initialize test environment
#    Setup Suite
#
#
#Teardown
#    [Documentation]     Cleanup test environment
#    Teardown Suite


