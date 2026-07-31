*** Settings ***
Documentation       Keywords for router end-to-end testing

Library             Collections
Library             OperatingSystem
Library             Process
Library             String
Library             SSHLibrary
Resource            oc.resource
Resource            common.resource
Resource            kubeconfig.resource
Resource            microshift-config.resource
Resource            microshift-process.resource


*** Variables ***
${WEB_SERVER_DEPLOY}            ./assets/router/web-server-deploy.yaml
${WEB_SERVER_SIGNED_DEPLOY}     ./assets/router/web-server-signed-deploy.yaml
${TEST_CLIENT_POD}              ./assets/router/test-client-pod.yaml
${INGRESS_HTTP}                 ./assets/router/microshift-ingress-http.yaml
${INGRESS_DESTCA}               ./assets/router/microshift-ingress-destca.yaml
${RSYSLOGD_POD}                 ./assets/router/rsyslogd-pod.yaml
${ROUTER_NS}                    openshift-ingress
${CLIENT_POD_NAME}              hello-pod
${CLIENT_POD_LABEL}             app\=hello-pod
${WEB_SERVER_LABEL}             name\=web-server-deploy


*** Keywords ***
Deploy Web Server
    [Documentation]    Deploy the nginx web server (unsecured) with services.
    [Arguments]    ${ns}=${NAMESPACE}
    Cleanup Stale Web Server Pods    ${ns}
    Oc Create    -f ${WEB_SERVER_DEPLOY} -n ${ns}
    Labeled Pod Should Be Ready    ${WEB_SERVER_LABEL}    ns=${ns}    timeout=120s

Deploy Web Server Signed
    [Documentation]    Deploy the nginx web server with TLS (service-serving-cert) and services.
    ...    Waits for the service-serving-cert-signer to create the TLS secret before
    ...    checking pod readiness, since the pod mounts the secret as a volume.
    [Arguments]    ${ns}=${NAMESPACE}
    Cleanup Stale Web Server Pods    ${ns}
    Oc Create    -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${ns}
    Wait Until Keyword Succeeds    300s    2s
    ...    Run With Kubeconfig    oc get secret service-secret -n ${ns}
    Labeled Pod Should Be Ready    ${WEB_SERVER_LABEL}    ns=${ns}    timeout=120s

Deploy Test Client Pod
    [Documentation]    Deploy the test client pod for running curl commands.
    [Arguments]    ${ns}=${NAMESPACE}
    Oc Create    -f ${TEST_CLIENT_POD} -n ${ns}
    Labeled Pod Should Be Ready    ${CLIENT_POD_LABEL}    ns=${ns}    timeout=120s

Deploy Rsyslogd Pod
    [Documentation]    Deploy the rsyslogd pod for syslog testing.
    [Arguments]    ${ns}=${NAMESPACE}
    Oc Create    -f ${RSYSLOGD_POD} -n ${ns}
    Labeled Pod Should Be Ready    name\=rsyslogd    ns=${ns}    timeout=120s

Get Web Server Pod Name
    [Documentation]    Get the name of the first web-server-deploy pod.
    [Arguments]    ${ns}=${NAMESPACE}
    ${name}=    Run With Kubeconfig
    ...    oc get pod -l name=web-server-deploy -n ${ns} -o jsonpath='{.items[0].metadata.name}'
    RETURN    ${name}

Create OC Route
    [Documentation]    Create an OpenShift route of the given type.
    ...    route_type: http (uses oc expose), or edge/passthrough/reencrypt (uses oc create route)
    [Arguments]    ${ns}    ${route_type}    ${route_name}    ${service_name}    @{extra_args}
    IF    "${route_type}" == "http"
        ${extra}=    Catenate    SEPARATOR=${SPACE}    @{extra_args}
        Run With Kubeconfig    oc expose -n ${ns} service ${service_name} --name=${route_name} ${extra}
    ELSE
        ${extra}=    Catenate    SEPARATOR=${SPACE}    @{extra_args}
        Run With Kubeconfig    oc create route ${route_type} ${route_name} --service=${service_name} -n ${ns} ${extra}
    END

Create OC Route And Admit
    [Documentation]    Create an OpenShift route and wait for it to be admitted.
    [Arguments]    ${ns}    ${route_type}    ${route_name}    ${service_name}    @{extra_args}
    Create OC Route    ${ns}    ${route_type}    ${route_name}    ${service_name}    @{extra_args}
    Route Should Be Admitted    ${route_name}    ${ns}

Route Should Be Admitted
    [Documentation]    Verify the route is admitted by the default ingress controller.
    [Arguments]    ${route_name}    ${ns}=${NAMESPACE}    ${timeout}=120s
    Wait Until Keyword Succeeds    ${timeout}    2s
    ...    Route Admission Status Should Be    ${route_name}    ${ns}    True

Ingress Route Should Be Admitted
    [Documentation]    Wait for the ingress-to-route controller to create and admit a Route from an Ingress.
    ...    The controller generates the Route name with a random suffix, so this keyword
    ...    finds the Route by its owner reference to the named Ingress.
    [Arguments]    ${ingress_name}    ${ns}=${NAMESPACE}    ${timeout}=300s
    Wait Until Keyword Succeeds    ${timeout}    5s
    ...    Ingress Route Admission Should Be True    ${ingress_name}    ${ns}

Ingress Route Admission Should Be True
    [Documentation]    Assert that a Route owned by the named Ingress exists and is admitted.
    [Arguments]    ${ingress_name}    ${ns}
    ${name}=    Run With Kubeconfig
    ...    oc get routes -n ${ns} -o jsonpath='{.items[?(@.metadata.ownerReferences[0].name=="${ingress_name}")].metadata.name}'
    Should Not Be Empty    ${name}    msg=No route found for ingress ${ingress_name}
    ${status}=    Oc Get JsonPath    route    ${ns}    ${name}
    ...    .status.ingress[0].conditions[0].status
    Should Be Equal As Strings    ${status}    True

Route Should Not Be Admitted
    [Documentation]    Verify the route is NOT admitted by the default ingress controller.
    [Arguments]    ${route_name}    ${ns}=${NAMESPACE}
    Wait Until Keyword Succeeds    60s    2s
    ...    Route Admission Status Should Be    ${route_name}    ${ns}    False

Route Admission Status Should Be
    [Documentation]    Helper: assert route admission status equals expected.
    [Arguments]    ${route_name}    ${ns}    ${expected}
    ${status}=    Oc Get JsonPath    route    ${ns}    ${route_name}
    ...    .status.ingress[0].conditions[0].status
    Should Be Equal As Strings    ${status}    ${expected}

Create And Admit Four Route Types
    [Documentation]    Create HTTP, edge, passthrough, and reencrypt routes and wait for each to be admitted.
    [Arguments]    ${http_host}    ${edge_host}    ${pass_host}    ${reen_host}
    Create OC Route And Admit    ${NAMESPACE}    http    route-http    service-unsecure    --hostname=${http_host}
    Create OC Route And Admit    ${NAMESPACE}    edge    route-edge    service-unsecure    --hostname=${edge_host}
    Create OC Route And Admit
    ...    ${NAMESPACE}
    ...    passthrough
    ...    route-passth
    ...    service-secure
    ...    --hostname=${pass_host}
    Create OC Route And Admit    ${NAMESPACE}    reencrypt    route-reen    service-secure    --hostname=${reen_host}

Curl From Pod
    [Documentation]    Run curl from a pod via oc exec. Returns stdout.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${flags}=${EMPTY}
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${ns} ${pod_name} -- curl ${url} -sI --resolve ${resolve} --connect-timeout 10 ${flags}
    RETURN    ${output}

Curl From Pod Should Contain
    [Documentation]    Curl from a pod and assert response contains expected string (no extra flags).
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}
    Should Contain    ${output}    ${expected}

Wait Until Curl Succeeds From Pod
    [Documentation]    Retry curl from a pod (no extra flags) until response contains expected string.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl From Pod Should Contain    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}

Curl HTTPS From Pod Should Contain
    [Documentation]    Curl HTTPS from a pod using -k, assert response contains expected string.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}    -k
    Should Contain    ${output}    ${expected}

Wait Until HTTPS Curl Succeeds From Pod
    [Documentation]    Retry HTTPS curl (with -k) from a pod until response contains expected string.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl HTTPS From Pod Should Contain    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}

Curl With Cookie Should Return 200
    [Documentation]    Curl from a pod with a cookie and assert 200 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cookie}
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}    -b ${cookie}
    Should Contain    ${output}    200

Wait Until Curl With Cookie Succeeds From Pod
    [Documentation]    Retry curl with a cookie until 200 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cookie}
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl With Cookie Should Return 200    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cookie}

HTTPS Curl With Cookie Should Return 200
    [Documentation]    Curl HTTPS from a pod with a cookie (using -k) and assert 200 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cookie}
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}    -k -b ${cookie}
    Should Contain    ${output}    200

Wait Until HTTPS Curl With Cookie Succeeds From Pod
    [Documentation]    Retry HTTPS curl with a cookie until 200 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cookie}
    Wait Until Keyword Succeeds    120s    2s
    ...    HTTPS Curl With Cookie Should Return 200    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cookie}

Curl With Client Cert Should Return
    [Documentation]    Curl with client cert from /data/certs/ and assert expected response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200 OK
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}
    ...    --cacert /data/certs/ca.crt --cert /data/certs/usr.crt --key /data/certs/usr.key
    Should Contain    ${output}    ${expected}

Wait Until Curl With Client Cert Succeeds
    [Documentation]    Retry curl with client cert until expected response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200 OK
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl With Client Cert Should Return    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}

Curl With Client Cert Should Return Insecure
    [Documentation]    Curl with client cert from /data/certs/, skipping TLS hostname verification, and assert expected response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200 OK
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}
    ...    --cacert /data/certs/ca.crt --cert /data/certs/usr.crt --key /data/certs/usr.key -k
    Should Contain    ${output}    ${expected}

Wait Until Curl With Client Cert Succeeds Insecure
    [Documentation]    Retry curl with client cert (skipping TLS hostname verification) until expected response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}=200 OK
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl With Client Cert Should Return Insecure    ${pod_name}    ${ns}    ${url}    ${resolve}    ${expected}

Curl With Cert File Should Return 200
    [Documentation]    Curl with a specific cert file prefix from /data/certs/ and assert 200 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cert_prefix}
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}
    ...    --cacert /data/certs/ca.crt --cert /data/certs/${cert_prefix}.crt --key /data/certs/${cert_prefix}.key
    Should Contain    ${output}    200 OK

Wait Until Curl With Cert File Succeeds
    [Documentation]    Retry curl with specific cert file until 200 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cert_prefix}
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl With Cert File Should Return 200    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cert_prefix}

Curl With Cert File Should Return 403
    [Documentation]    Curl with a specific cert file prefix from /data/certs/ and assert 403 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cert_prefix}
    ${output}=    Curl From Pod    ${pod_name}    ${ns}    ${url}    ${resolve}
    ...    --cacert /data/certs/ca.crt --cert /data/certs/${cert_prefix}.crt --key /data/certs/${cert_prefix}.key
    Should Contain    ${output}    403

Wait Until Curl With Cert File Returns 403
    [Documentation]    Retry curl with specific cert file until 403 response.
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cert_prefix}
    Wait Until Keyword Succeeds    120s    2s
    ...    Curl With Cert File Should Return 403    ${pod_name}    ${ns}    ${url}    ${resolve}    ${cert_prefix}

Curl Without Cert Should Return SSL Error
    [Documentation]    Curl without a client cert and assert SSL handshake error (for mTLS Required).
    [Arguments]    ${pod_name}    ${ns}    ${url}    ${resolve}
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${ns} ${pod_name} -- curl ${url} -sIkv --resolve ${resolve} --connect-timeout 10
    ...    allow_fail=True
    Should Contain    ${output}    SSL_read

Get Router Pod Name
    [Documentation]    Get the name of the router-default pod.
    ${name}=    Run With Kubeconfig
    ...    oc get pod -l ingresscontroller.operator.openshift.io/deployment-ingresscontroller=default -n ${ROUTER_NS} -o jsonpath='{.items[0].metadata.name}'
    RETURN    ${name}

Get Router Pod IP
    [Documentation]    Get the pod IP of the router pod.
    ${pod_name}=    Get Router Pod Name
    ${ip}=    Oc Get JsonPath    pod    ${ROUTER_NS}    ${pod_name}    .status.podIP
    RETURN    ${ip}

Router Pod Env Should Have Value
    [Documentation]    Check an env var value in the router pod.
    [Arguments]    ${env_name}    ${expected_value}
    ${actual}=    Oc Get JsonPath
    ...    pod    ${ROUTER_NS}    ${EMPTY}
    ...    .items[*].spec.containers[*].env[?(@.name=="${env_name}")].value
    Should Be Equal As Strings    ${actual}    ${expected_value}

Read Haproxy Config
    [Documentation]    Read haproxy.config from the router pod.
    ${pod_name}=    Get Router Pod Name
    ${output}=    Run With Kubeconfig
    ...    oc exec -n ${ROUTER_NS} ${pod_name} -- cat haproxy.config
    RETURN    ${output}

Get Router Access Logs
    [Documentation]    Get access logs from the router pod's logs container.
    ${output}=    Oc Logs    deployment/router-default -c logs    ${ROUTER_NS}
    RETURN    ${output}

Wait For Router Logs To Contain
    [Documentation]    Wait for the router access logs to contain a specific pattern.
    [Arguments]    ${pattern}    ${timeout}=120s
    Wait Until Keyword Succeeds    ${timeout}    5s
    ...    Router Logs Should Contain    ${pattern}

Router Logs Should Contain
    [Documentation]    Check that router access logs contain a pattern.
    [Arguments]    ${pattern}
    ${logs}=    Get Router Access Logs
    Should Contain    ${logs}    ${pattern}

Wait For Router Ready
    [Documentation]    Wait for the default router deployment to be available.
    Oc Wait    namespace/openshift-ingress    --for jsonpath='{.status.phase}=Active' --timeout=5m
    Named Deployment Should Be Available    router-default    ${ROUTER_NS}    5m

Setup Router Config And Restart
    [Documentation]    Apply drop-in config and restart MicroShift, then wait for router ready.
    [Arguments]    ${config_content}
    Drop In MicroShift Config    ${config_content}    10-router
    Restart MicroShift
    Wait For Router Ready

Remove Router Config And Restart
    [Documentation]    Remove the drop-in config and restart MicroShift, then wait for router ready.
    Remove Drop In MicroShift Config    10-router
    Restart MicroShift
    Wait For Router Ready

Cleanup Stale Web Server Pods
    [Documentation]    Delete any leftover web-server-deploy resources from a prior test and wait for
    ...    pods to terminate. Prevents oc wait --selector from matching ghost pod references
    ...    across tests that share the name=web-server-deploy label.
    [Arguments]    ${ns}
    Run With Kubeconfig    oc delete -f ${WEB_SERVER_DEPLOY} -n ${ns} --ignore-not-found    allow_fail=True
    Run With Kubeconfig    oc delete -f ${WEB_SERVER_SIGNED_DEPLOY} -n ${ns} --ignore-not-found    allow_fail=True
    Run With Kubeconfig    oc wait pod -l ${WEB_SERVER_LABEL} -n ${ns} --for=delete --timeout=60s    allow_fail=True

Cleanup Test Workloads
    [Documentation]    Delete all routes, deployments, services, configmaps, and pods created by tests in the namespace.
    Run With Kubeconfig    oc delete route,deployment,service,configmap,pod --all -n ${NAMESPACE} --ignore-not-found
    ...    allow_fail=True

Get LB IPs
    [Documentation]    Get all load balancer IPs of the router-default service as a space-separated string.
    ${ips}=    Oc Get JsonPath
    ...    service    ${ROUTER_NS}    router-default
    ...    .status.loadBalancer.ingress[*].ip
    RETURN    ${ips}

Get LB Port
    [Documentation]    Get a specific named port of the router-default service.
    [Arguments]    ${port_name}
    ${port}=    Oc Get JsonPath
    ...    service    ${ROUTER_NS}    router-default
    ...    .spec.ports[?(@.name=="${port_name}")].port
    RETURN    ${port}

Create Configmap From Files
    [Documentation]    Create a configmap from one or more local file paths.
    [Arguments]    ${ns}    ${name}    @{from_files}
    ${file_args}=    Catenate    SEPARATOR=${SPACE}    @{from_files}
    Run With Kubeconfig    oc create configmap ${name} ${file_args} -n ${ns}

Scale Deployment
    [Documentation]    Scale a deployment to a given number of replicas.
    [Arguments]    ${ns}    ${deploy_name}    ${replicas}
    Run With Kubeconfig    oc scale deployment/${deploy_name} --replicas=${replicas} -n ${ns}

Copy Files To Pod
    [Documentation]    Copy a local path to a pod path via oc cp.
    [Arguments]    ${ns}    ${pod_name}    ${src}    ${dest}
    Run With Kubeconfig    oc cp -n ${ns} ${src} ${ns}/${pod_name}:${dest}

Generate CA Certificate
    [Documentation]    Generate a self-signed CA key and certificate.
    [Arguments]    ${ca_key}    ${ca_crt}    ${ca_subj}
    ${result}=    Run Process    openssl genrsa -out ${ca_key} 2048    shell=True
    Should Be Equal As Integers    ${result.rc}    0
    ${result}=    Run Process
    ...    openssl req -x509 -new -nodes -key ${ca_key} -sha256 -days 30 -subj "${ca_subj}" -out ${ca_crt}
    ...    shell=True
    Should Be Equal As Integers    ${result.rc}    0

Generate CSR And Key
    [Documentation]    Generate a private key and a CSR for the given subject.
    [Arguments]    ${key_file}    ${csr_file}    ${subj}
    ${result}=    Run Process    openssl genrsa -out ${key_file} 2048    shell=True
    Should Be Equal As Integers    ${result.rc}    0
    ${result}=    Run Process
    ...    openssl req -new -key ${key_file} -subj "${subj}" -out ${csr_file}
    ...    shell=True
    Should Be Equal As Integers    ${result.rc}    0

Sign CSR With CA
    [Documentation]    Sign a CSR with a CA, writing SAN config to a temp file when provided.
    [Arguments]    ${csr_file}    ${ca_crt}    ${ca_key}    ${signed_crt}    ${san}=${EMPTY}
    IF    len($san) > 0
        VAR    ${ext_file}=    /tmp/router-san-${signed_crt.replace('/', '-')}.cnf
        Create File    ${ext_file}    ${san}\n
        ${result}=    Run Process
        ...    openssl x509 -req -in ${csr_file} -CA ${ca_crt} -CAkey ${ca_key} -CAcreateserial -out ${signed_crt} -days 30 -sha256 -extfile ${ext_file}
        ...    shell=True
        Remove File    ${ext_file}
    ELSE
        ${result}=    Run Process
        ...    openssl x509 -req -in ${csr_file} -CA ${ca_crt} -CAkey ${ca_key} -CAcreateserial -out ${signed_crt} -days 30 -sha256
        ...    shell=True
    END
    Should Be Equal As Integers    ${result.rc}    0

Generate MTLS Client Cert
    [Documentation]    Generate a CA and a single client cert in the given directory.
    [Arguments]    ${tmpdir}    ${edge_host}
    Generate CA Certificate    ${tmpdir}/ca.key    ${tmpdir}/ca.crt    /CN=MS-Test-Root-CA
    Generate CSR And Key    ${tmpdir}/usr.key    ${tmpdir}/usr.csr    /CN=example-test.com
    VAR    ${san}=    subjectAltName = DNS.1:*.${BASE_DOMAIN},DNS.2:${edge_host}
    Sign CSR With CA    ${tmpdir}/usr.csr    ${tmpdir}/ca.crt    ${tmpdir}/ca.key    ${tmpdir}/usr.crt    ${san}

Generate Client Cert File In Dir
    [Documentation]    Generate a CSR, key, and signed cert in the given directory using a name prefix.
    [Arguments]    ${tmpdir}    ${host}    ${cn}    ${cert_prefix}
    Generate CSR And Key    ${tmpdir}/${cert_prefix}.key    ${tmpdir}/${cert_prefix}.csr    /CN=${cn}
    VAR    ${san}=    subjectAltName = DNS.1:*.${BASE_DOMAIN},DNS.2:${host}
    Sign CSR With CA
    ...    ${tmpdir}/${cert_prefix}.csr
    ...    ${tmpdir}/ca.crt
    ...    ${tmpdir}/ca.key
    ...    ${tmpdir}/${cert_prefix}.crt
    ...    ${san}

Get First LB IP
    [Documentation]    Get the first load balancer IP of the router-default service.
    ${ip}=    Oc Get JsonPath
    ...    service    ${ROUTER_NS}    router-default
    ...    .status.loadBalancer.ingress[0].ip
    RETURN    ${ip}

Get Router Deployment Generation
    [Documentation]    Get the current generation of the router-default deployment.
    ${gen}=    Oc Get JsonPath
    ...    deployment    ${ROUTER_NS}    router-default
    ...    .metadata.generation
    RETURN    ${gen}

Get Host IPs Via SSH
    [Documentation]    Get all global IPv4 addresses from the MicroShift host, excluding link-local (169.254.x.x).
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip -4 addr show scope global | grep -oP 'inet \\K[\\d.]+' | grep -v '^169\\.254\\.'
    ...    sudo=False    return_rc=True    return_stderr=True    return_stdout=True
    Should Be Equal As Integers    0    ${rc}
    @{ips}=    Split String    ${stdout}    \n
    RETURN    @{ips}

Get First Host Interface And IP Via SSH
    [Documentation]    Get the first non-loopback IPv4 interface name and its IP from the host.
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    ip -4 addr show scope global | awk '/^[0-9]+:/{iface=$2; gsub(":","",iface)} /inet /{print iface, gensub(/\\/.*/, "", 1, $2); exit}'
    ...    sudo=False
    ...    return_rc=True
    ...    return_stderr=True
    ...    return_stdout=True
    Should Be Equal As Integers    0    ${rc}
    @{parts}=    Split String    ${stdout}
    ${iface}=    Get From List    ${parts}    0
    ${ip}=    Get From List    ${parts}    1
    RETURN    ${iface}    ${ip}

Show Invalid Drop In Config Should Fail With
    [Documentation]    Write an invalid drop-in config, run show-config, verify it fails with expected message, then remove the drop-in.
    [Arguments]    ${config_content}    ${expected_error}
    Drop In MicroShift Config    ${config_content}    10-router-invalid
    ${stdout}    ${stderr}    ${rc}=    SSHLibrary.Execute Command
    ...    microshift show-config --mode effective
    ...    sudo=True    return_rc=True    return_stderr=True    return_stdout=True
    Remove Drop In MicroShift Config    10-router-invalid
    Should Not Be Equal As Integers    ${rc}    0
    Should Contain    ${stderr}${stdout}    ${expected_error}

Privileged Namespace
    [Documentation]    Grant privileged SCC to the default service account in a namespace.
    [Arguments]    ${ns}=${NAMESPACE}
    Run With Kubeconfig    oc adm policy add-scc-to-user privileged -z default -n ${ns}
