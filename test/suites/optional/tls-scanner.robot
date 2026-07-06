*** Settings ***
Documentation       Test tls-scanner tool with MicroShift host-based scanning.
...                 Clones openshift/tls-scanner, deploys the scanner job with
...                 scanner-job-microshift.yaml.template and SCAN_MODE=host,
...                 waits for completion, and collects results.
...                 See: https://github.com/openshift/tls-scanner

Library             OperatingSystem
Library             Process
Library             String
Resource            ../../resources/common.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/optional-config.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           tls-scanner    security    optional


*** Variables ***
# Set by Suite Setup (common.resource / kubeconfig.resource):
${NAMESPACE}                    default
${KUBECONFIG}                   ${EMPTY}
# External: full tag of the scanner image to use (e.g. quay.io/my-org/tls-scanner:latest)
${SCANNER_IMAGE}                registry.ci.openshift.org/ocp/4.22:tls-scanner-tool

${TLS_SCANNER_REPO}             https://github.com/openshift/tls-scanner

${TLS_SCANNER_DIR}              ${EMPTY}
${TLS_SCANNER_JOB_NAME}         tls-scanner-job
${JOB_WAIT_TIMEOUT}             10min
${CLUSTER_READER_MANIFEST}      ./assets/tls-scanner/cluster-reader-clusterrole.yaml


*** Test Cases ***
TLS Scanner Host Scan Completes And Produces Artifacts
    [Documentation]    Clone tls-scanner, verify scanner image is available,
    ...    deploy the scan job in host mode for MicroShift, wait for completion,
    ...    and collect results (results.json, results.csv, scan.log).
    [Setup]    Run Keywords
    ...    Check Required Scanner Variables
    ...    Clone TLS Scanner Repo
    ...    Ensure Cluster Reader Role Exists
    Deploy TLS Scanner Job
    Copy Scan Results Artifacts

    [Teardown]    Run Keywords
    ...    Cleanup TLS Scanner Job
    ...    Ensure Cluster Reader Role Deleted

Ingress Router TLS Curves supports ML-KEM Post Quantum Curves
    [Documentation]    Verify TLS curve negotiation with openssl from inside the router pod.
    Verify ML-KEM Post Quantum Curve Negotiation


*** Keywords ***
Setup
    [Documentation]    Setup suite with base MicroShift only (no optional components)
    Setup Suite
    Setup MicroShift With Optionals
    ${ns}=    Create Unique Namespace
    VAR    ${NAMESPACE}=    ${ns}    scope=SUITE

Teardown
    [Documentation]    Restore config and teardown suite
    Teardown MicroShift With Optionals
    Teardown Suite With Namespace

Check Required Scanner Variables
    [Documentation]    Fail if SCANNER_IMAGE is not set.
    Should Not Be Empty    ${SCANNER_IMAGE}
    ...    SCANNER_IMAGE must be set (full image tag, e.g. quay.io/my-org/tls-scanner:latest)

Ensure Cluster Reader Role Exists
    [Documentation]    Create cluster-reader ClusterRole for MicroShift (not shipped by default).
    ...    deploy.sh expects this OpenShift role to exist for the scanner ServiceAccount.
    Oc Apply    -f ${CLUSTER_READER_MANIFEST}

Ensure Cluster Reader Role Deleted
    [Documentation]    Delete cluster-reader ClusterRole for MicroShift (not shipped by default).
    ${result}=    Run Keyword And Ignore Error    Process.Run Process    oc delete clusterrole cluster-reader
    ...    env:KUBECONFIG=${KUBECONFIG}
    IF    "${result[0]}" == "FAIL"    Log    TLS scanner job cleanup failed

Clone TLS Scanner Repo
    [Documentation]    Clone openshift/tls-scanner into a temporary directory.
    ${rand}=    Generate Random String    8    [LOWER]
    VAR    ${workdir}=    /tmp/tls-scanner-${rand}
    Create Directory    ${workdir}
    VAR    ${TLS_SCANNER_DIR}=    ${workdir}    scope=SUITE
    ${result}=    Process.Run Process    git clone --depth 1 ${TLS_SCANNER_REPO} .
    ...    cwd=${TLS_SCANNER_DIR}    shell=True    timeout=120s
    Should Be Equal As Integers    ${result.rc}    0    msg=Failed to clone tls-scanner repo

Deploy TLS Scanner Job
    [Documentation]    Deploy the scanner job using MicroShift host template and SCAN_MODE=host.
    ${result}=    Process.Run Process    bash -c 'bash -x ./deploy.sh deploy 2>&1'
    ...    cwd=${TLS_SCANNER_DIR}
    ...    env:KUBECONFIG=${KUBECONFIG}
    ...    env:SCANNER_IMAGE=${SCANNER_IMAGE}
    ...    env:NAMESPACE=${NAMESPACE}
    ...    env:JOB_TEMPLATE_FILE=scanner-job-microshift.yaml.template
    ...    env:SCAN_MODE=host
    ...    env:OUTPUTDIR=${OUTPUTDIR}
    ...    shell=True    timeout=${JOB_WAIT_TIMEOUT}    stdout=${OUTPUTDIR}/tls-scanner-std.log
    Log    ${result.stdout}
    Log    ${result.stderr}
    Should Be Equal As Integers    ${result.rc}    0    msg=Failed to deploy tls-scanner job
    OperatingSystem.File Should Exist    ${TLS_SCANNER_DIR}/artifacts/results.json
    ${size}=    OperatingSystem.Get File Size    ${TLS_SCANNER_DIR}/artifacts/results.json
    Should Be True    ${size} > 0    msg=results.json is missing or empty
    # Check that results.json does not contain "ERROR,"
    ${content}=    OperatingSystem.Get File    ${TLS_SCANNER_DIR}/artifacts/results.json
    Should Not Contain    ${content}    ERROR,    msg=Scan results contain ERROR,

Copy Scan Results Artifacts
    [Documentation]    Copy content of ${TLS_SCANNER_DIR}/artifacts to ${OUTPUTDIR}/tls-scanner-artifacts.
    VAR    ${dest}=    ${OUTPUTDIR}/tls-scanner-artifacts
    Create Directory    ${dest}
    OperatingSystem.Directory Should Exist    ${TLS_SCANNER_DIR}/artifacts
    ${files}=    OperatingSystem.List Files In Directory    ${TLS_SCANNER_DIR}/artifacts
    ${count}=    Get Length    ${files}
    Should Be True    ${count} > 0    msg=No artifacts produced by tls-scanner
    FOR    ${f}    IN    @{files}
        Copy File    ${TLS_SCANNER_DIR}/artifacts/${f}    ${dest}/
    END
    Log    Copied scan results to ${dest}/

Cleanup TLS Scanner Job
    [Documentation]    Remove the scanner job and RBAC via deploy.sh cleanup.
    ${result}=    Run Keyword And Ignore Error    Process.Run Process    ./deploy.sh cleanup
    ...    cwd=${TLS_SCANNER_DIR}
    ...    env:KUBECONFIG=${KUBECONFIG}
    ...    env:NAMESPACE=${NAMESPACE}
    ...    shell=True    timeout=60s
    IF    "${result[0]}" == "PASS"    Log    TLS scanner job cleanup completed
    Remove Directory    ${TLS_SCANNER_DIR}    recursive=True
    IF    '${TLS_SCANNER_DIR}' != ''
        Run Keyword And Ignore Error    Remove Directory    ${TLS_SCANNER_DIR}    recursive=True
    END

Verify ML-KEM Post Quantum Curve Negotiation
    [Documentation]    Verify X25519MLKEM768 post-quantum hybrid key exchange
    ...    negotiates successfully via oc exec into the router pod, which
    ...    has OpenSSL 3.5+ (the host OpenSSL may be too old for ML-KEM).
    ...    Skipped on FIPS clusters where ML-KEM is not configured.
    ${curves}=    Oc Get JsonPath    deployment    openshift-ingress    router-default
    ...    .spec.template.spec.containers[0].env[?(@.name=="ROUTER_CURVES")].value
    Skip If    "X25519MLKEM768" not in """${curves}"""
    ...    ROUTER_CURVES does not include X25519MLKEM768 (FIPS mode); skipping ML-KEM test
    ${router_ip}=    Oc Get JsonPath    svc    openshift-ingress    router-default
    ...    .spec.clusterIP
    ${pod_name}=    Oc Get JsonPath    pod    openshift-ingress    ${EMPTY}
    ...    .items[0].metadata.name
    ${output}=    Oc Exec    ${pod_name}
    ...    echo Q | openssl s_client -connect ${router_ip}:443 -groups X25519MLKEM768 2>&1 || true
    ...    ns=openshift-ingress
    Should Match Regexp    ${output}    group.*X25519MLKEM768
    ...    msg=ML-KEM post-quantum curve X25519MLKEM768 negotiation failed
    Log    Post-quantum ML-KEM negotiation verified: OK
