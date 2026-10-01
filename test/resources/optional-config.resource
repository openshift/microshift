*** Settings ***
Documentation       Keywords for reconfiguring MicroShift with per-suite
...                 optional components via kustomizePaths.

Library             Collections
Library             String
Resource            common.resource
Resource            microshift-config.resource
Resource            microshift-process.resource
Resource            ostree-health.resource


*** Keywords ***
Setup MicroShift With Optionals
    [Documentation]    Reconfigure MicroShift to load only the specified optional
    ...    manifests.d directories, clean all data, and restart fresh.
    ...    Each argument is a directory name under /usr/lib/microshift/manifests.d/
    ...    (e.g. 001-microshift-olm). Call with no arguments to run base MicroShift
    ...    without any optional components.
    [Arguments]    @{manifests_dirs}
    Save Default MicroShift Config
    ${kustomize_yaml}=    Build Optionals Config    @{manifests_dirs}
    ${config}=    Replace MicroShift Config    ${kustomize_yaml}
    Stop MicroShift
    Upload MicroShift Config    ${config}
    Cleanup MicroShift    --all    --keep-images
    Systemctl    start    microshift.service
    Wait For MicroShift Healthcheck Success
    Setup Kubeconfig

Teardown MicroShift With Optionals
    [Documentation]    Restore the original MicroShift config after a suite.
    ...    If setup never completed, the default config variable may not exist,
    ...    so fall back to clearing the config file.
    ${exists}=    Run Keyword And Return Status    Variable Should Exist    ${DEFAULT_MICROSHIFT_CONFIG}
    IF    ${exists}
        Restore Default MicroShift Config
    ELSE
        Clear MicroShift Config
    END

Build Optionals Config
    [Documentation]    Build a config.yaml string with kustomizePaths containing
    ...    only the base manifests path and the specified manifests.d entries.
    [Arguments]    @{manifests_dirs}
    VAR    @{lines}=
    ...    manifests:
    ...    \ \ kustomizePaths:
    ...    \ \ \ \ - /usr/lib/microshift/manifests
    FOR    ${dir}    IN    @{manifests_dirs}
        Append To List    ${lines}    \ \ \ \ - /usr/lib/microshift/manifests.d/${dir}
    END
    ${config}=    Catenate    SEPARATOR=\n    @{lines}
    RETURN    ${config}
