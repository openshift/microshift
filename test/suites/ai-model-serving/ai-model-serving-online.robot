*** Settings ***
Documentation       Sanity test for AI Model Serving

Library             ../../resources/DataFormats.py
Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource

Suite Setup         Setup Suite
Suite Teardown      Teardown Suite


*** Variables ***
${USHIFT_HOST}=             ${EMPTY}
${OVMS_KSERVE_MANIFEST}=    /tmp/ovms-kserve.yaml
${OVMS_REQUEST}=            /tmp/ovms-request.json


*** Test Cases ***
Test OpenVINO model
    [Documentation]    Sanity test for AI OpenVino Model Serving

    Set Test Variable    ${MODEL_NAME}    openvino-resnet
    Set Test Variable    ${DOMAIN}    ${MODEL_NAME}-predictor-test-ai.apps.example.com
    ${ns}=    Create Unique Namespace
    Set Test Variable    ${NAMESPACE}    ${ns}

    Deploy OpenVINO Serving Runtime
    Deploy OpenVINO Resnet Model

    Check If Model Is Ready
    Query Model Metrics Endpoint
    Prepare Request Data
    Query Model Infer Endpoint

    [Teardown]    Run Keywords
    ...    Remove Namespace    ${NAMESPACE}
    ...    AND
    ...    Remove Tmp Data


*** Keywords ***
Deploy OpenVINO Serving Runtime
    [Documentation]    Deploys OpenVino server.

    ${ovms_image}=    Command Should Work
    ...    jq -r '.images | with_entries(select(.key == "ovms-image")) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -i)".json
    SSHLibrary.Get File
    ...    /usr/lib/microshift/manifests.d/001-microshift-ai-model-serving/runtimes/ovms-kserve.yaml
    ...    ${OVMS_KSERVE_MANIFEST}
    Local Command Should Work    sed -i "s,image: ovms-image,image: ${ovms_image}," "${OVMS_KSERVE_MANIFEST}"
    Oc Apply    -n ${NAMESPACE} -f ${OVMS_KSERVE_MANIFEST}

Deploy OpenVINO Resnet Model
    [Documentation]    Deploys InferenceService object to create Deployment and Service to serve the model.
    ...    Also creates a Route to export the model endpoint outside the MicroShift cluster.

    Oc Apply    -n ${NAMESPACE} -f ./assets/ai-model-serving/ovms-resources.yaml
    Wait Until Keyword Succeeds    30x    1s
    ...    Run With Kubeconfig    oc rollout status -n\=${NAMESPACE} --timeout=60s deployment openvino-resnet-predictor

Check If Model Is Ready
    [Documentation]    Asks model server is model is ready for inference.
    ${cmd}=    Catenate
    ...    curl
    ...    --fail
    ...    -i ${DOMAIN}/v2/models/${MODEL_NAME}/ready
    ...    --connect-to "${DOMAIN}::${USHIFT_HOST}:"
    Wait Until Keyword Succeeds    10x    10s
    ...    Local Command Should Work    ${cmd}

Query Model Metrics Endpoint
    [Documentation]    Makes a query against the model server metrics endpoint.

    ${cmd}=    Catenate
    ...    curl
    ...    --fail
    ...    --request GET
    ...    ${DOMAIN}/metrics
    ...    --connect-to "${DOMAIN}::${USHIFT_HOST}:"
    ${output}=    Local Command Should Work    ${cmd}
    Should Contain    ${output}    ovms_requests_success Number of successful requests to a model or a DAG.

Prepare Request Data
    [Documentation]    Executes a script that prepares a request data.

    Local Command Should Work    bash -x assets/ai-model-serving/ovms-query-preparation.sh ${OVMS_REQUEST}

Remove Tmp Data
    [Documentation]    Remove temp data for this test.

    Local Command Should Work    rm ${OVMS_REQUEST} ${OVMS_KSERVE_MANIFEST}

Query Model Infer Endpoint
    [Documentation]    Makes a query against the model server.

    # Inference-Header-Content-Length is the len of the JSON at the begining of the request.json
    ${cmd}=    Catenate
    ...    curl
    ...    --silent
    ...    --fail
    ...    --request POST
    ...    --data-binary "@${OVMS_REQUEST}"
    ...    --header "Inference-Header-Content-Length: 63"
    ...    ${DOMAIN}/v2/models/${MODEL_NAME}/infer
    ...    --connect-to "${DOMAIN}::${USHIFT_HOST}:"
    ${output}=    Local Command Should Work    ${cmd}
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
