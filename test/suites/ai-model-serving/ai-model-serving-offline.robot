*** Settings ***
Documentation       Sanity test for AI Model Serving

Resource            ../../resources/offline.resource
Library             ../../resources/DataFormats.py

Suite Setup         offline.Setup Suite

Test Tags           offline


*** Variables ***
${MODEL_NAME}=      openvino-resnet
${DOMAIN}=          ${MODEL_NAME}-predictor-test-ai.apps.example.com
${IP}=              10.44.0.1


*** Test Cases ***
Sanity Test
    [Documentation]    Sanity test for AI Model Serving

    Model Serving Offline Test

    # Check if ingress exists.
    # Enabled only for testing purposes to test Kserve's settings override.
    offline.Run With Kubeconfig    oc    get    ingress    -n    test-ai    openvino-resnet

    offline.Reboot MicroShift Host
    offline.Wait For Greenboot Health Check To Exit
    Model Serving Offline Test


*** Keywords ***
Model Serving Offline Test
    [Documentation]    Waits for the model server and queries it
    Wait For A Deployment    test-ai    openvino-resnet-predictor
    Wait Until Keyword Succeeds    10x    10s
    ...    Check If Model Is Ready
    Query Model Metrics
    Prepare Request Data
    Query Model Server

Wait For A Deployment
    [Documentation]    Wait for a deployment on offline VM
    [Arguments]    ${namespace}    ${name}
    offline.Run With Kubeconfig    oc    rollout    status    -n\=${namespace}    deployment    ${name}

Check If Model Is Ready
    [Documentation]    Ask model server is model is ready for inference.
    ${cmd}=    Catenate
    ...    curl
    ...    --fail
    ...    -i ${DOMAIN}/v2/models/${MODEL_NAME}/ready
    ...    --connect-to "${DOMAIN}::${IP}:"
    Guest Process Should Succeed    ${cmd}

Query Model Metrics
    [Documentation]    Makes a query against the model server metrics endpoint.
    ${cmd}=    Catenate
    ...    curl
    ...    --fail
    ...    --request GET
    ...    ${DOMAIN}/metrics
    ...    --connect-to "${DOMAIN}::${IP}:"
    ${output}=    Guest Process Should Succeed    ${cmd}
    Should Contain    ${output}    ovms_requests_success Number of successful requests to a model or a DAG.

Query Model Server
    [Documentation]    Makes a query against the model server.

    # Inference-Header-Content-Length is the len of the JSON at the begining of the request.json
    ${cmd}=    Catenate
    ...    curl
    ...    --fail
    ...    --request POST
    ...    --data-binary "@/tmp/request.json"
    ...    --header "Inference-Header-Content-Length: 63"
    ...    ${DOMAIN}/v2/models/${MODEL_NAME}/infer
    ...    --connect-to "${DOMAIN}::${IP}:"
    ${output}=    Guest Process Should Succeed    ${cmd}
    ${result}=    Json Parse    ${output}
    ${data}=    Set Variable    ${result["outputs"][0]["data"]}
    # Following expression can be referred to as 'argmax': index of the highest element.
    ${argmax}=    Evaluate    ${data}.index(max(${data}))

    # Request data includes bee.jpeg file so according to the OpenVino examples,
    # we should expect argmax to be 309.
    # See following for reference
    # - https://github.com/openvinotoolkit/model_server/tree/releases/2025/0/client/python/kserve-api/samples#run-the-client-to-perform-inference-1
    # - https://github.com/openvinotoolkit/model_server/blob/releases/2025/0/demos/image_classification/input_images.txt
    Should Be Equal As Integers    ${argmax}    309

Prepare Request Data
    [Documentation]    Executes a script on the host that prepares a request data.
    ...    Refer to the file's contents for explanation.

    Guest Process Should Succeed
    ...    /etc/microshift/manifests.d/10-ai-model-serving-test/prepare-query.sh

Guest Process Should Succeed
    [Documentation]    Executes shell command on the VM, checks the return code,
    ...    and returns stdout.
    [Arguments]    ${cmd}
    ${result}    ${ignore}=    Run Guest Process    ${GUEST_NAME}
    ...    bash
    ...    -c
    ...    ${cmd}
    Should Be Equal As Integers    ${result["rc"]}    0
    RETURN    ${result["stdout"]}
