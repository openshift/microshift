*** Settings ***
Documentation       Tests for configuration changes

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}                           ${EMPTY}    # The journal cursor before restarting MicroShift
${BAD_LOG_LEVEL}                    SEPARATOR=\n
...                                 ---
...                                 debugging:
...                                 \ \ logLevel: unknown-value
${DEBUG_LOG_LEVEL}                  SEPARATOR=\n
...                                 ---
...                                 debugging:
...                                 \ \ logLevel: debug
${BAD_AUDIT_PROFILE}                SEPARATOR=\n
...                                 apiServer:
...                                 \ \ auditLog:
...                                 \ \ \ \ profile: BAD_PROFILE
${AUDIT_PROFILE}                    SEPARATOR=\n
...                                 apiServer:
...                                 \ \ auditLog:
...                                 \ \ \ \ profile: WriteRequestBodies
${AUDIT_FLAGS}                      SEPARATOR=\n
...                                 apiServer:
...                                 \ \ auditLog:
...                                 \ \ \ \ maxFileSize: 1000
...                                 \ \ \ \ maxFiles: 1000
...                                 \ \ \ \ maxFileAge: 1000
${CUSTOM_FEATURE_GATES}             SEPARATOR=\n
...                                 apiServer:
...                                 \ \ featureGates:
...                                 \ \ \ \ featureSet: CustomNoUpgrade
...                                 \ \ \ \ customNoUpgrade:
...                                 \ \ \ \ \ \ enabled:
...                                 \ \ \ \ \ \ \ \ - TestFeatureEnabled
...                                 \ \ \ \ \ \ disabled:
...                                 \ \ \ \ \ \ \ \ - TestFeatureDisabled
${DIFFERENT_FEATURE_GATES}          SEPARATOR=\n
...                                 apiServer:
...                                 \ \ featureGates:
...                                 \ \ \ \ featureSet: CustomNoUpgrade
...                                 \ \ \ \ customNoUpgrade:
...                                 \ \ \ \ \ \ enabled:
...                                 \ \ \ \ \ \ \ \ - DifferentTestFeature
${FEATURE_GATE_LOCK_FILE}           /var/lib/microshift/no-upgrade
${LVMS_DEFAULT}                     SEPARATOR=\n
...                                 storage: {}
${LVMS_DISABLED}                    SEPARATOR=\n
...                                 storage:
...                                 \ \ driver: "none"
${CSI_SNAPSHOT_DISABLED}            SEPARATOR=\n
...                                 storage:
...                                 \ \ optionalCsiComponents: [ none ]
${LVMS_CSI_SNAPSHOT_DISABLED}       SEPARATOR=\n
...                                 storage:
...                                 \ \ driver: "none"
...                                 \ \ optionalCsiComponents: [ none ]
${APISERVER_ETCD_CLIENT_CERT}       /var/lib/microshift/certs/etcd-signer/apiserver-etcd-client


*** Test Cases ***
MicroShift Starts Using Default Config
    [Documentation]    Default (example) config should not fail to be parsed
    ...    and prevent MicroShift from starting.
    [Tags]    default-config
    # Copy existing config.yaml as a drop-in because it has subjectAltNames
    # required by the `Restart MicroShift` keyword (sets up required kubeconfig).
    [Setup]    Run Keywords
    ...    Save Default MicroShift Config
    ...    AND
    ...    Command Should Work    mkdir -p /etc/microshift/config.d/
    ...    AND
    ...    Command Should Work    cp /etc/microshift/config.yaml /etc/microshift/config.d/00-ci.yaml
    ...    AND
    ...    Command Should Work    cp /etc/microshift/config.yaml.default /etc/microshift/config.yaml

    Restart MicroShift

    [Teardown]    Run Keywords
    ...    Restore Default MicroShift Config
    ...    AND
    ...    Command Should Work    rm -f /etc/microshift/config.d/00-ci.yaml
    ...    AND
    ...    Restart MicroShift

Unknown Log Level Produces Warning
    [Documentation]    Logs should warn that the log level setting is unknown
    Setup With Bad Log Level
    Pattern Should Appear In Log Output    ${CURSOR}    Unrecognized log level "unknown-value", defaulting to "Normal"

Debug Log Level Produces No Warning
    [Documentation]    Logs should not warn that the log level setting is unknown
    Setup With Debug Log Level
    Pattern Should Not Appear In Log Output    ${CURSOR}    Unrecognized log level "debug", defaulting to "Normal"

Known Audit Log Profile Produces No Warning
    [Documentation]    A recognized kube-apiserver audit log profile will not produce a message in logs
    Setup Known Audit Log Profile
    Pattern Should Not Appear In Log Output    ${CURSOR}    unknown audit profile \\\\"WriteRequestBodies\\\\"

Config Flags Are Logged in Audit Flags
    [Documentation]    Check that flags specified in the MicroShift audit config are passed to and logged by the
    ...    kube-apiserver. It is not essential that we test the kube-apiserver functionality as that is
    ...    already rigorously tested by upstream k8s and by OCP.
    Setup Audit Flags
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxsize=\"1000\"
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxbackup=\"1000\"
    Pattern Should Appear In Log Output    ${CURSOR}    FLAG: --audit-log-maxage=\"1000\"

Custom Feature Gates Are Passed To Kube APIServer
    [Documentation]    Check that custom feature gates specified in the MicroShift config are passed to and logged by the
    ...    kube-apiserver. This test verifies that arbitrary feature gate values are correctly propagated from the
    ...    MicroShift configuration to the kube-apiserver, regardless of whether the feature gates are valid or have any effect.
    Setup Custom Feature Gates
    Pattern Should Appear In Log Output    ${CURSOR}    regex:setting kube:feature-gates=.*TestFeatureEnabled=true
    Pattern Should Appear In Log Output    ${CURSOR}    regex:setting kube:feature-gates=.*TestFeatureDisabled=false

Feature Gate Lock File Created With Custom Feature Gates
    [Documentation]    Verify that feature gate lock file is created when custom feature gates are configured.
    ...    The lock file prevents upgrades and configuration changes when CustomNoUpgrade feature set is used.
    [Setup]    Setup MicroShift With Custom Feature Gates

    Wait Until Keyword Succeeds    2 min    5 sec    Feature Gate Lock File Should Exist
    Feature Gate Lock File Should Contain Feature Gates    CustomNoUpgrade    TestFeatureEnabled

    [Teardown]    Teardown MicroShift After Custom Feature Gates Test

Feature Gate Config Change Blocked After Lock Created
    [Documentation]    Verify that changing feature gate config is blocked after lock file exists.
    ...    MicroShift must refuse to start if feature gates change after CustomNoUpgrade is set.
    [Setup]    Setup MicroShift With Custom Feature Gates

    Wait Until Keyword Succeeds    2 min    5 sec    Feature Gate Lock File Should Exist

    Stop MicroShift
    Drop In MicroShift Config    ${DIFFERENT_FEATURE_GATES}    10-featuregates
    MicroShift Should Fail To Start Due To Feature Gate Change

    [Teardown]    Teardown MicroShift After Custom Feature Gates Test

Feature Gate Lock File Persists Across Restarts With Same Config
    [Documentation]    Verify that feature gate lock file persists and validation succeeds across restarts
    ...    when the same feature gate configuration is maintained.
    [Setup]    Setup MicroShift With Custom Feature Gates

    Wait Until Keyword Succeeds    2 min    5 sec    Feature Gate Lock File Should Exist

    # Restart with same config - should succeed
    Restart MicroShift
    Feature Gate Lock File Should Exist

    [Teardown]    Teardown MicroShift After Custom Feature Gates Test

Deploy MicroShift With LVMS By Default
    [Documentation]    Verify that LVMS and CSI snapshotting are deployed when config fields are null.
    [Setup]    Deploy Storage Config    ${LVMS_DEFAULT}
    LVMS Is Deployed
    CSI Snapshot Controller Is Deployed
    [Teardown]    Run Keywords
    ...    Remove Storage Drop In Config
    ...    Restart MicroShift

Deploy MicroShift Without LVMS
    [Documentation]    Verify that LVMS is not deployed when storage.driver == none, and that CSI snapshotting
    ...    components are still deployed.
    [Setup]    Deploy Storage Config    ${LVMS_DISABLED}

    CSI Snapshot Controller Is Deployed
    Run Keyword And Expect Error    1 != 0
    ...    LVMS Is Deployed
    [Teardown]    Run Keywords
    ...    Remove Storage Drop In Config
    ...    Restart MicroShift

Deploy MicroShift Without CSI Snapshotter
    [Documentation]    Verify that only LVMS is deployed when .storage.optionalCsiComponents is an empty array.
    [Setup]    Deploy Storage Config    ${CSI_SNAPSHOT_DISABLED}

    LVMS Is Deployed
    Run Keyword And Expect Error    1 != 0
    ...    CSI Snapshot Controller Is Deployed

    [Teardown]    Run Keywords
    ...    Remove Storage Drop In Config
    ...    Restart MicroShift

Crio Uses Crun Runtime
    [Documentation]    Verify that crio uses crun as its default runtime

    ${runtime}=    Command Should Work    crictl info | jq -r '.runtimeHandlers[].name | select(. != null)'
    Should Contain    ${runtime}    crun

    ${stdout}    ${stderr}    ${rc}=    Command Execution    rpm -q microshift-low-latency
    IF    ${rc} == 0    Should Contain    ${runtime}    high-performance

Http Proxy Not Defined In Bootc Image
    [Documentation]    Verify that the http proxy environment variables are not defined
    ...    in the bootc image used to install the system.

    # Only run the check if the system is a bootc image
    ${is_bootc}=    Is System Bootc
    IF    ${is_bootc}    Check HTTP Proxy Env In Bootc Image


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Save Journal Cursor

Teardown
    [Documentation]    Test suite teardown
    Remove Drop In MicroShift Config    10-loglevel
    Remove Drop In MicroShift Config    10-audit
    Remove Drop In MicroShift Config    10-featuregates
    Restart MicroShift
    Logout MicroShift Host
    Remove Kubeconfig

Save Journal Cursor
    [Documentation]
    ...    Save the journal cursor then restart MicroShift so we capture the
    ...    shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE

Setup With Bad Log Level
    [Documentation]    Set log level to an unknown value and restart
    Drop In MicroShift Config    ${BAD_LOG_LEVEL}    10-loglevel
    Restart MicroShift

Setup With Debug Log Level
    [Documentation]    Set log level to debug and restart
    Drop In MicroShift Config    ${DEBUG_LOG_LEVEL}    10-loglevel
    Restart MicroShift

Setup Known Audit Log Profile
    [Documentation]    Setup audit
    Drop In MicroShift Config    ${AUDIT_PROFILE}    10-audit
    Restart MicroShift

Setup Audit Flags
    [Documentation]    Apply the audit config values set in ${AUDIT_FLAGS}
    Drop In MicroShift Config    ${AUDIT_FLAGS}    10-audit
    Restart MicroShift

Setup Custom Feature Gates
    [Documentation]    Apply the custom feature gates config values set in ${CUSTOM_FEATURE_GATES}
    Stop MicroShift
    Remove Feature Gate Lock File If Exists
    Drop In MicroShift Config    ${CUSTOM_FEATURE_GATES}    10-featuregates
    Start MicroShift

Remove Feature Gate Lock File If Exists
    [Documentation]    Remove the feature gate lock file if it exists, for test cleanup
    Command Should Work    rm -f ${FEATURE_GATE_LOCK_FILE}

Feature Gate Lock File Should Exist
    [Documentation]    Verify that the feature gate lock file exists
    Command Should Work    test -f ${FEATURE_GATE_LOCK_FILE}

Feature Gate Lock File Should Contain Feature Gates
    [Documentation]    Verify that feature gate lock file contains the expected feature gate configuration
    [Arguments]    ${feature_set}    ${feature_name}
    ${contents}=    Command Should Work    cat ${FEATURE_GATE_LOCK_FILE}
    Should Contain    ${contents}    ${feature_set}
    Should Contain    ${contents}    ${feature_name}

MicroShift Should Fail To Start Due To Feature Gate Change
    [Documentation]    Verify that MicroShift fails to start when feature gate config changes
    ${stdout}    ${stderr}    ${rc}=    Command Execution    systemctl start microshift
    Should Not Be Equal As Integers    ${rc}    0    MicroShift should fail to start

    # Check journal for the expected error message
    ${journal}=    Command Should Work    journalctl -u microshift -n 100 --no-pager
    Should Contain    ${journal}    feature gate configuration has changed
    Should Contain    ${journal}    microshift-cleanup-data

Setup MicroShift With Custom Feature Gates
    [Documentation]    Stop MicroShift, remove lock file, and start with custom feature gates
    Stop MicroShift
    Remove Feature Gate Lock File If Exists
    Drop In MicroShift Config    ${CUSTOM_FEATURE_GATES}    10-featuregates
    Start MicroShift

Teardown MicroShift After Custom Feature Gates Test
    [Documentation]    Clean up after a feature gate test
    Stop MicroShift
    Remove Drop In MicroShift Config    10-featuregates
    Remove Feature Gate Lock File If Exists
    Start MicroShift

Deploy Storage Config
    [Documentation]    Applies a storage ${config} to the exist MicroShift config, pushes it to the MicroShift host,
    ...    and restarts microshift.service
    [Arguments]    ${config}
    Cleanup MicroShift    opt='--keep-images'
    Drop In MicroShift Config    ${config}    10-storage
    Start MicroShift

Remove Storage Drop In Config
    [Documentation]    Remove the previously created drop-in config for storage
    Remove Drop In MicroShift Config    10-storage

LVMS Is Deployed
    [Documentation]    Wait for LVMS components to deploy
    Named Deployment Should Be Available    lvms-operator    openshift-storage    120s
    # Wait for vg-manager daemonset to exist before trying to "wait".
    # `oc wait` fails if the object doesn't exist.
    Wait Until Resource Exists    daemonset    vg-manager    openshift-storage    120s
    Named Daemonset Should Be Available    vg-manager    openshift-storage    120s

CSI Snapshot Controller Is Deployed
    [Documentation]    Wait for CSI snapshot controller to be deployed
    Named Deployment Should Be Available    csi-snapshot-controller    kube-system    120s

Check HTTP Proxy Env In Bootc Image
    [Documentation]    Check that the HTTP proxy environment variables are not defined
    ...    in the bootc image used to install the system.
    # Obtain the current bootc image reference
    ${bootc_image}=    Command Should Work    bootc status --json | jq -r .spec.image.image
    # Inspect the bootc image environment variables
    ${env_vars}=    Command Should Work
    ...    skopeo inspect --authfile /etc/crio/openshift-pull-secret --config docker://${bootc_image} | jq -r '.config.Env'

    # Verify that the environment variables are not defined
    ${env_var_lc}=    Convert To Lower Case    ${env_vars}
    Should Not Contain    ${env_var_lc}    http_proxy\=
    Should Not Contain    ${env_var_lc}    https_proxy\=
    Should Not Contain    ${env_var_lc}    no_proxy\=
