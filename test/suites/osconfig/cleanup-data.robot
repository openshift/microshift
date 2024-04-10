*** Settings ***
Documentation       Tests verifying microshift-cleanup-data script functionality

Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/ostree-health.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-network.resource
Resource            ../../resources/kubeconfig.resource

Suite Setup         Setup Suite

Test Tags           slow


*** Test Cases ***
Verify Invalid Command Line
    [Documentation]    Verify invalid command line combinations

    # Usage message
    ${rc}=    Run MicroShift Cleanup Data    ${EMPTY}
    Should Not Be Equal As Integers    ${rc}    0

    # Invalid option combination
    ${rc}=    Run MicroShift Cleanup Data    --ovn    --keep-images
    Should Not Be Equal As Integers    ${rc}    0

    ${rc}=    Run MicroShift Cleanup Data    --all    --ovn
    Should Not Be Equal As Integers    ${rc}    0

    ${rc}=    Run MicroShift Cleanup Data    --all    --cert
    Should Not Be Equal As Integers    ${rc}    0

    ${rc}=    Run MicroShift Cleanup Data    --ovn    --cert
    Should Not Be Equal As Integers    ${rc}    0

    ${rc}=    Run MicroShift Cleanup Data    --keep-images
    Should Not Be Equal As Integers    ${rc}    0

Verify Full Cleanup Data
    [Documentation]    Verify full data clean scenarios
    [Setup]    Run Keywords
    ...    Setup Suite With Namespace
    ...    Create Hello MicroShift Pod

    ${rc}=    Run MicroShift Cleanup Data    --all
    Should Be Equal As Integers    ${rc}    0

    MicroShift Processes Should Not Exist
    Verify Remote Directory Does Not Exist With Sudo    /var/lib/microshift
    Verify Remote Directory Exists With Sudo    /var/lib/microshift-backups

    Crio Containers Should Not Exist
    Crio Pods Should Not Exist
    Crio Images Should Not Exist

    OVN Processes Should Not Exist
    OVN Data Should Not Exist
    OVN Internal Bridge Should Not Exist

    [Teardown]    Run Keywords
    ...    Run MicroShift Test Case Teardown

Verify Keep Images Cleanup Data
    [Documentation]    Verify keep images data clean scenario
    [Setup]    Run Keywords
    ...    Setup Suite With Namespace
    ...    Create Hello MicroShift Pod

    ${rc}=    Run MicroShift Cleanup Data    --keep-images    --all
    Should Be Equal As Integers    ${rc}    0

    MicroShift Processes Should Not Exist
    Verify Remote Directory Does Not Exist With Sudo    /var/lib/microshift
    Verify Remote Directory Exists With Sudo    /var/lib/microshift-backups

    Crio Containers Should Not Exist
    Crio Pods Should Not Exist
    Crio Images Should Exist

    OVN Processes Should Not Exist
    OVN Data Should Not Exist
    OVN Internal Bridge Should Not Exist

    [Teardown]    Run Keywords
    ...    Run MicroShift Test Case Teardown

Verify OVN Cleanup Data
    [Documentation]    Verify OVN data cleanup scenario
    [Setup]    Run Keywords
    ...    Setup Suite With Namespace
    ...    Create Hello MicroShift Pod

    ${rc}=    Run MicroShift Cleanup Data    --ovn
    Should Be Equal As Integers    ${rc}    0

    MicroShift Processes Should Not Exist
    Verify Remote Directory Exists With Sudo    /var/lib/microshift
    Verify Remote Directory Exists With Sudo    /var/lib/microshift/certs
    Verify Remote Directory Exists With Sudo    /var/lib/microshift-backups

    Crio Containers Should Not Exist
    Crio Pods Should Not Exist
    Crio Images Should Exist

    OVN Processes Should Not Exist
    OVN Data Should Not Exist
    OVN Internal Bridge Should Not Exist

    [Teardown]    Run Keywords
    ...    Run MicroShift Test Case Teardown

Verify Cert Cleanup Data
    [Documentation]    Verify certificate data cleanup scenario
    [Setup]    Run Keywords
    ...    Setup Suite With Namespace
    ...    Create Hello MicroShift Pod

    ${rc}=    Run MicroShift Cleanup Data    --cert
    Should Be Equal As Integers    ${rc}    0

    MicroShift Processes Should Not Exist
    Verify Remote Directory Exists With Sudo    /var/lib/microshift
    Verify Remote Directory Does Not Exist With Sudo    /var/lib/microshift/certs
    Verify Remote Directory Exists With Sudo    /var/lib/microshift-backups

    Crio Containers Should Not Exist
    Crio Pods Should Not Exist
    Crio Images Should Exist

    OVN Processes Should Exist
    OVN Data Should Exist
    OVN Internal Bridge Should Exist

    [Teardown]    Run Keywords
    ...    Run MicroShift Test Case Teardown


*** Keywords ***
Setup Suite
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Start MicroShift And Wait Until Ready

Start MicroShift And Wait Until Ready
    [Documentation]    Start the service and wait until fully initialized
    Systemctl    enable    microshift
    Systemctl    start    microshift
    Restart Greenboot And Wait For Success

Run MicroShift Cleanup Data
    [Documentation]    Run the microshift-cleanup-data script and
    ...    return its exit code
    [Arguments]    ${cmd}    ${opt}=${EMPTY}
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    echo 1 | sudo microshift-cleanup-data ${cmd} ${opt}
    ...    return_stdout=True    return_stderr=True    return_rc=True
    RETURN    ${rc}

Run MicroShift Test Case Teardown
    [Documentation]    Run the microshift-cleanup-data script and restart
    ...    the service to ensure clean startup of other tests
    # Try keeping the images to save on restart times
    ${rc}=    Run MicroShift Cleanup Data    --keep-images    --all
    Should Be Equal As Integers    ${rc}    0
    Start MicroShift And Wait Until Ready

MicroShift Processes Should Not Exist
    [Documentation]    Make sure that MicroShift and Etcd services are not running

    # MicroShift service and Etcd process should be down
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof microshift microshift-etcd
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0

Crio Containers Should Not Exist
    [Documentation]    Make sure cri-o containers do not exist

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl ps -a | wc -l
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Be Equal As Integers    ${stdout}    1

Crio Pods Should Not Exist
    [Documentation]    Make sure cri-o pods do not exist

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl pods | wc -l
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Should Be Equal As Integers    ${stdout}    1

Crio Images Should Exist
    [Documentation]    Make sure cri-o images exist

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    crictl images | wc -l
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

    ${stdout_int}=    Convert To Integer    ${stdout}
    Should Be True    ${stdout_int} > 1

Crio Images Should Not Exist
    [Documentation]    Make sure cri-o images do not exist

    ${status}=    Run Keyword And Return Status
    ...    Crio Images Should Exist
    Should Not Be True    ${status}

OVN Processes Should Not Exist
    [Documentation]    Make sure that OVN processes are not running

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof conmon pause ovn-controller ovn-northd ovsdb-server
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0

OVN Processes Should Exist
    [Documentation]    Make sure that OVN processes are running

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    pidof conmon pause ovn-controller ovn-northd ovsdb-server
    ...    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0

OVN Data Should Not Exist
    [Documentation]    Make sure that OVN data files and directories are deleted

    # OVN data directories and files should be deleted
    Verify Remote Directory Does Not Exist With Sudo    /var/run/ovn
    Verify Remote Directory Does Not Exist With Sudo    /var/run/ovn-kubernetes
    Verify Remote File Does Not Exist With Sudo    /etc/cni/net.d/10-ovn-kubernetes.conf
    Verify Remote File Does Not Exist With Sudo    /run/cni/bin/ovn-k8s-cni-overlay

OVN Data Should Exist
    [Documentation]    Make sure that OVN data files and directories exist

    # OVN data directories and files should be deleted
    Verify Remote Directory Exists With Sudo    /var/run/ovn
    Verify Remote Directory Exists With Sudo    /var/run/ovn-kubernetes
    Verify Remote File Exists With Sudo    /etc/cni/net.d/10-ovn-kubernetes.conf
    Verify Remote File Exists With Sudo    /run/cni/bin/ovn-k8s-cni-overlay

OVN Internal Bridge Should Not Exist
    [Documentation]    Make sure that OVN internal bridge devices do not exist

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ovs-vsctl br-exists br-int
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Not Be Equal As Integers    ${rc}    0

OVN Internal Bridge Should Exist
    [Documentation]    Make sure that OVN internal bridge devices exist

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    ovs-vsctl br-exists br-int
    ...    sudo=True    return_stdout=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    ${rc}    0
