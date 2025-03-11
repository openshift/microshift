# AI Model Serving on MicroShift

AI Model Serving on MicroShift is a platform for serving AI models.
It includes limited subset of Red Hat OpenShift AI (RHOAI): [kserve](https://github.com/opendatahub-io/kserve) and ServingRuntimes.
Now you can train your models in the cloud and serve them on the edge.

MicroShift is intended for the edge (small footprint, low resources),
therefore kserve is configured to use a "Raw Deployment" mode which means that:
- Kubernetes Deployments and Services will be created, and
- neither Service Mesh (Istio) nor Serverless (Knative) are required.

Additionally, automatic creation of the Ingress objects is disabled. If you want to expose your model outside the cluster,
create a Route object instead.

[Learn more about Raw deployment mode from RHOAI documentation](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.17/html/serving_models/serving-large-models_serving-large-models#raw_deployment_mode)

## Limitations
- Currently, AI Model Serving on MicroShift is only available on x86_64 platform.
- AI Model Serving on MicroShift only includes very specific subset of RHOAI.
- Securing (e.g. OAUTH2) exposed model server's endpoint is user's responsibility

## Known issues
- Because of [bug in kserve](https://github.com/kserve/kserve/pull/4274)
  ([to be ported to RHOAI](https://issues.redhat.com/browse/RHOAIENG-21106)),
  rebooting a MicroShift host can result model server not coming back up if it was using ModelCar (model in OCI image).
- Because of MicroShift's architecture, installing microshift-ai-model-serving RPM before running `systemctl start microshift` for the first time,
  can cause a failure to start up, however MicroShift will restart and then start successfully. See [OCPBUGS-51365](https://issues.redhat.com/browse/OCPBUGS-51365).
  - **It might also cause the ServingRuntimes to not be deployed because kustomizer gives up before webhook is available**
- Currently, ClusterServingRuntimes are not supported by RHOAI which means that
  user needs to copy ServingRuntime shipped within the package to their workload's namespace.
  In future releases we're hoping to transition to ClusterServingRuntime which is ServingRuntime usable from any namespace
  (if users will need to override some options, they still will be able to create ServingRuntime which take precedence over ClusterServingRuntime).

## Definitions

- [ServingRuntime](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.17/html/serving_models/serving-large-models_serving-large-models#servingruntime)
- [InferenceService](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.17/html/serving_models/serving-large-models_serving-large-models#inferenceservice)

## Supported ServingRuntimes

For the list of supported Serving Runtimes refer to
[RHOAI documentation](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.17/html/serving_models/serving-large-models_serving-large-models#supported-model-serving-runtimes_serving-large-models).
Keep in mind that in MicroShift they are created directly as ServingRuntimes in `redhat-ods-applications`, not as templates,
so they need to be copied to workload's namespace.

## General usage overview

1. Prepare model for serving
1. Configure OS and MicroShift for the hardware - driver & device plugin
1. Install AI Model Serving for MicroShift
1. (Optional) Package model into an OCI image (ModelCar)
1. Select suitable Serving Runtime (Model Server)
1. Copy ServingRuntime Custom Resource from `redhat-ods-applications` to your own namespace
1. Create InferenceService
1. Create Route
1. Make requests against model server

## Setting up hardware - drivers and device plugins

In order to enable GPU/hardware accelerator for Red Hat Device Edge,
you must follow partner's guidance on installing Operator, or driver for RHEL and Device Plugin for Kubernetes.
Operators may be more convenient, but installing only the driver and device plugin may be more resource efficient.

MicroShift does not provide support for partner's procedure.
In case of any problems consult partner's documentation or support.
Following links serve only as examples / pointers and may not include everything needed.
- NVIDIA:
  - [GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html)
  - [Device Plugin](https://github.com/NVIDIA/k8s-device-plugin)
  - [Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html)
  - [Driver](https://docs.nvidia.com/datacenter/tesla/driver-installation-guide/index.html) / [CUDA Toolkit](https://docs.nvidia.com/cuda/cuda-installation-guide-linux/)
- Intel Guadi
  - [Base Operator](https://docs.habana.ai/en/latest/Installation_Guide/Additional_Installation/Kubernetes_Installation/Kubernetes_Operator.html)
  - [Device Plugin](https://docs.habana.ai/en/latest/Installation_Guide/Additional_Installation/Kubernetes_Installation/Intel_Gaudi_Kubernetes_Device_Plugin.html)
  - [Driver](https://docs.habana.ai/en/latest/Installation_Guide/Driver_Installation.html)
- AMD
  - [Operator](https://instinct.docs.amd.com/projects/gpu-operator/en/latest/)
  - [Device Plugin](https://github.com/ROCm/k8s-device-plugin)
  - [Driver](https://rocm.docs.amd.com/projects/install-on-linux/en/latest/install/install-overview.html)

## Step by step guide

Below is an example usage of AI Model Serving for MicroShift.
It uses OpenVino Model Server (OVMS) and resnet50 model.
OVMS can run on CPU, no hardware accelerator is needed for this example.

### Installing AI Model Serving for MicroShift

`microshift-ai-model-serving` RPM contains manifests that deploy `kserve` and `ServingRuntimes` in `redhat-ods-applications` namespace.

To install AI Model Serving for MicroShift run following command:
```bash
sudo dnf install microshift-ai-model-serving
```

After installing the package (and restarting MicroShift),
there should be new Pod running in `redhat-ods-applications` namespace:
```sh
$ oc get pods -n redhat-ods-applications
NAME                                        READY   STATUS    RESTARTS   AGE
kserve-controller-manager-7fc9fc688-kttmm   1/1     Running   0          1h
```

You can also install the release info package. It contains JSON file with image references
useful for offline procedures or deploying copy of a ServingRuntime to your namespace during
bootc image build:
```bash
sudo dnf install microshift-ai-model-serving-release-info
```

### Package model into an OCI image (ModelCar)

You can package your model into an OCI image and make use of what is known as [ModelCar](https://kserve.github.io/website/latest/modelserving/storage/oci/) approach.
This can help you set up offline environments because the model can be embedded just like any other container image.

Exact directory structure depends on the model server.
Below is an example Containerfile with RESNET50 model compatible with OpenVino Model Server (OVMS)
used in the OVMS' examples.

See "How to build a ModelCar container" section of
[Build and deploy a ModelCar container in OpenShift AI](https://developers.redhat.com/articles/2025/01/30/build-and-deploy-modelcar-container-openshift-ai)
article for guidance on building OCI image with a model from huggingface suitable for vLLM model server.

```Dockerfile
FROM quay.io/microshift/busybox:1.37
RUN mkdir -p /models/1 && chmod -R 755 /models/1
RUN wget -q -P /models/1 \
  https://storage.openvinotoolkit.org/repositories/open_model_zoo/2022.1/models_bin/2/resnet50-binary-0001/FP32-INT1/resnet50-binary-0001.bin \
  https://storage.openvinotoolkit.org/repositories/open_model_zoo/2022.1/models_bin/2/resnet50-binary-0001/FP32-INT1/resnet50-binary-0001.xml
```

You can build it and push it to your registry:
```sh
podman build -t IMAGE_REF .
podman push IMAGE_REF
```

For purposes of this documentation, we'll build it locally and use without pushing to registry first.
`sudo` is required to make it part of the root's container storage and usable by MicroShift
as CRI-O and Podman share the storage.

For offline use cases make sure to include a tag other than `latest`.
If the `latest` tag is used, the container that fetches and sets up the model will be set up with `imagePullPolicy: Always`
and local image will be ignored. For any other tag, the `imagePullPolicy` will be `IfNotPresent`.

```sh
sudo podman build -t ovms-resnet50:test .
```

Example output:
```
STEP 1/3: FROM quay.io/microshift/busybox:1.37
Trying to pull quay.io/microshift/busybox:1.37...
Getting image source signatures
Copying blob a46fbb00284b done   |
Copying config 27a71e19c9 done   |
Writing manifest to image destination
STEP 2/3: RUN mkdir -p /models/1 && chmod -R 755 /models/1
--> eacb7039436a
STEP 3/3: RUN wget -q -P /models/1   https://storage.openvinotoolkit.org/repositories/open_model_zoo/2022.1/models_bin/2/resnet50-binary-0001/FP32-INT1/resnet50-binary-0001.bin   https://storage.openvinotoolkit.org/repositories/open_model_zoo/2022.1/models_bin/2/resnet50-binary-0001/FP32-INT1/resnet50-binary-0001.xml
wget: note: TLS certificate validation not implemented
COMMIT ovms-resnet50
--> ac4606eb6cb3
Successfully tagged localhost/ovms-resnet50:test
ac4606eb6cb3e6be2fbee9d6bc271df212eb22e6a45a2c33394d9c73dc3bb4cf
```

Run following command to make sure the image exists
```sh
sudo podman images | grep ovms-resnet50
```

Example output:
```
localhost/ovms-resnet50                                     test          ac4606eb6cb3  3 minutes ago  27.5 MB
```

### Create namespace

Use following command to create new namespace that will be used throughout this guide.

```sh
oc create ns ai-demo
```

### Deploying ServingRuntime to workload's namespace

First, you must select ServingRuntime that supports the format of your model.
Then, you need to create ServingRuntime in your workload's namespace:
it can be done in two ways: manual and suitable for automation.

For more information about ServingRuntimes refer to
[RHOAI documentation](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.17/html/serving_models/serving-large-models_serving-large-models#model-serving-runtimes_serving-large-models)

#### Manually - from running cluster

You can use following command to save ServingRuntime definition to a yaml file:
```sh
oc get -n redhat-ods-applications servingruntimes kserve-ovms -o yaml > ovms.yaml
```
Then, you need to remove following metadata:
- `creationTimestamp`
- `generation`
- `resourceVersion`
- `uid`
- `namespace`

Then you can re-create the ServingRuntime in a custom namespace:
```sh
oc create -n ai-demo -f ./ovms.yaml
```

#### Automatically based on installed manifests and release info

Overview of the procedure:
1. Install `microshift-ai-model-serving-release-info` RPM
1. Extract image reference of particular ServingRuntime from the release info file
1. Make a copy of chosen ServingRuntime yaml file
1. Substitute placeholder used for `image:` field with actual image reference
1. Create the object using file or make it part of manifest (kustomization)

Example below shows the process to OpenVino Model Server (OVMS) ServingRuntime:
```sh
# Get image reference for the 'ovms-image'
OVMS_IMAGE="$(jq -r '.images | with_entries(select(.key == "ovms-image")) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -i)".json)"

# Duplicate the original ServingRuntime yaml
cp /usr/lib/microshift/manifests.d/001-microshift-ai-model-serving/runtimes/ovms-kserve.yaml ./ovms-kserve.yaml

# Update the image reference
sed -i "s,image: ovms-image,image: ${OVMS_IMAGE}," ./ovms-kserve.yaml
```

Then you can re-create the ServingRuntime in a custom namespace:
```sh
oc create -n ai-demo -f ./ovms-kserve.yaml
```

Alternatively, if the procedure above is part of bootc Containerfile and
the ServingRuntime ends up as part of new manifest, the namespace can set in `kustomization.yaml`:
```yaml
namespace: ai-demo
```

### Create InferenceService

Next, we need to create InferenceService custom resource.
InferenceService instructs kserve on how to create a Deployment for serving the model.
Kserve uses ServingRuntime based on `modelFormat` specified in InferenceService.

It's possible to add extra arguments that will be passed to the model server using `.spec.predictor.model.args`.

Below is example of InferenceService with model in `openvino_ir` format.
It features additional argument `--layout=NHWC:NCHW` to make OVMS accept the
request data in different format than usual for demonstration purposes.
Extra args are passed through to the OVMS container.

For more information about InferenceService CR refer to
[RHOAI documentation](https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.17/html/serving_models/serving-large-models_serving-large-models#inferenceservice).

```yaml
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: ovms-resnet50
spec:
  predictor:
    model:
      protocolVersion: v2
      modelFormat:
        name: openvino_ir
      storageUri: "oci://localhost/ovms-resnet50:test"
      args:
      - --layout=NHWC:NCHW
```

Save the example above of the InferenceService to a file, and create it on the cluster:
```sh
oc create -n ai-demo -f ./FILE.yaml
```

Output:
```
inferenceservice.serving.kserve.io/ovms-resnet50 created
```

Soon, a Deployment and a Pod should appear in that namespace.
Depending on the size of ServingRuntime's image and size of the ModelCar OCI image,
it may take a while for the Pod to become ready.

```sh
$ oc get -n ai-demo deployment
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
ovms-resnet50-predictor   1/1     1            1           72s
```

```sh
$ oc get -n ai-demo pod
NAME                                       READY   STATUS    RESTARTS      AGE
ovms-resnet50-predictor-6fdb566b7f-bc9k5   2/2     Running   1 (72s ago)   74s
```

Kserve will also create a Service.

```sh
$ oc get svc -n ai-demo
NAME                      TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
ovms-resnet50-predictor   ClusterIP   None         <none>        80/TCP    119s
```

#### Specifying hardware accelerators

InferenceService can also include plethora of different options.
For example it can contain `resources` section that is passed to the Deployment and then to the Pod,
so the model server gets access to the NVIDIA hardware (thanks to the device plugin):
```yaml
spec:
  predictor:
    model:
      resources:
        limits:
          nvidia.com/gpu: 1
        requests:
          nvidia.com/gpu: 1
```

For complete InferenceService specification refer to [kserve API Reference](https://kserve.github.io/website/latest/reference/api/).

### Create Route

> Note: You don't need to wait for model server Pod's readiness before creating Route.

Finally, we can create an OpenShift Route CR to expose the service.
You can either use `oc expose svc` command or create definition in a YAML file and apply it.

```sh
$ oc expose svc -n ai-demo ovms-resnet50-predictor
route.route.openshift.io/ovms-resnet50-predictor exposed

$ oc get route -n ai-demo
NAME                      HOST                                               ADMITTED   SERVICE                   TLS
ovms-resnet50-predictor   ovms-resnet50-predictor-ai-demo.apps.example.com   True       ovms-resnet50-predictor
```

### Querying the model server

Finally, we'll check if our model is ready for inference and we'll reuse OVMS examples to test the inference.

Get the IP of the MicroShift cluster and assign it to the `IP` variable.
Use the HOST value of the Route created in previous step and assign to the `DOMAIN` variable.
Next, run the following `curl` command.
Alternatively, instead of using `--connect-to "${DOMAIN}::${IP}:"`,
you can use real DNS, or add the IP and the domain to the `/etc/hosts` file.
```sh
DOMAIN=ovms-resnet50-predictor-ai-demo.apps.example.com
IP=192.168.0.10
curl -i "${DOMAIN}/v2/models/ovms-resnet50/ready" \
    --connect-to "${DOMAIN}::${IP}:"
```

Response code 200 is expected. Example output:
```
HTTP/1.1 200 OK
content-type: application/json
date: Wed, 12 Mar 2025 16:01:32 GMT
content-length: 0
set-cookie: 56bb4b6df4f80f0b59f56aa0a5a91c1a=4af1408b4a1c40925456f73033d4a7d1; path=/; HttpOnly
```

We can also query model's metadata:
```sh
curl "${DOMAIN}/v2/models/ovms-resnet50" \
    --connect-to "${DOMAIN}::${IP}:"
```
Output:
```json
{"name":"ovms-resnet50","versions":["1"],"platform":"OpenVINO","inputs":[{"name":"0","datatype":"FP32","shape":[1,224,224,3]}],"outputs":[{"name":"1463","datatype":"FP32","shape":[1,1000]}]
```

Let's try querying actual model - following example verifies if inference is in accordance to the training data.

First, download an image of a bee from the OpenVino examples:
```sh
curl -O https://raw.githubusercontent.com/openvinotoolkit/model_server/main/demos/common/static/images/bee.jpeg
```

Next, we need to create the request data.
1. Start with an inference header in JSON format.
1. Get the size of the header. It needs to be passed to the OVMS later in form of HTTP header.
1. Append size of the image to the request file. OVMS expects 4 bytes (little endian).
   Command below uses `xxd` utility which is part of `vim-common` package.
1. Append the image to the request file.

```sh
IMAGE=./bee.jpeg
REQ=./request.json

# Add an inference header
echo -n '{"inputs" : [{"name": "0", "shape": [1], "datatype": "BYTES"}]}' > "${REQ}"

# Get the size of the inference header
HEADER_LEN="$(stat -c %s "${REQ}")"

# Add size of the data (image) in binary format (4 bytes, little endian)
printf "%08X" $(stat --format=%s "${IMAGE}") | sed 's/\(..\)/\1\n/g' | tac | tr -d '\n' | xxd -r -p >> "${REQ}"

# Add the data, i.e. the image
cat "${IMAGE}" >> "${REQ}"
```

Now we can make an inference request against the model server using `ovms-resnet50` model.
```sh
curl \
    --data-binary "@./request.json" \
    --header "Inference-Header-Content-Length: ${HEADER_LEN}" \
    "${DOMAIN}/v2/models/ovms-resnet50/infer" \
    --connect-to "${DOMAIN}::${IP}:" > response.json
```

Response saved to a `response.json` is a JSON object which a structure as shown below.
Contents of `.outputs[0].data` were omitted from the example for brevity.
```json
{
    "model_name": "ovms-resnet50",
    "model_version": "1",
    "outputs": [{
            "name": "1463",
            "shape": [1, 1000],
            "datatype": "FP32",
            "data": [ ....... ]
        }]
}
```

To verify the response, we'll use python.
We need to obtain index of the highest element in the `.outputs[0].data`.

```python
import json

with open('response.json') as f:
    response = json.load(f)

data = response["outputs"][0]["data"]
argmax = data.index(max(data))
print(argmax)
```

The output of the script above should be `309`.
We can validate it against [resnet's input data](https://github.com/openvinotoolkit/model_server/blob/main/client/common/resnet_input_images.txt):
```
../../../../demos/common/static/images/bee.jpeg 309
```

You can try querying the model using other images mentioned in the resnet's input data.

#### Getting the model server's metrics

To obtain metrics of the model server simply make a request on `/metrics` endpoint:
```sh
curl "${DOMAIN}/metrics" \
    --connect-to "${DOMAIN}::${IP}:"
```

Partial example output:
```
# HELP ovms_requests_success Number of successful requests to a model or a DAG.
# TYPE ovms_requests_success counter
ovms_requests_success{api="KServe",interface="REST",method="ModelReady",name="ovms-resnet50"} 4
ovms_requests_success{api="KServe",interface="REST",method="ModelMetadata",name="ovms-resnet50",version="1"} 1
```

## Appendix
### Overriding kserve configuration

If you wish to override kserve settings, you need to make a copy of existing ConfigMap, tweak desired settings, and overwrite existing ConfigMap.

Settings are stored in a ConfigMap named `inferenceservice-config` in namespace `redhat-ods-applications`.
Alternatively, you can copy it from `/usr/lib/microshift/manifests.d/001-microshift-ai-model-serving/kserve/inferenceservice-config-microshift-patch.yaml`.

After tweaking it, you must apply it and restart kserve (e.g. deleting Pod or scaling the Deployment down to 0 and back to 1).
For OSTree/bootc systems, create a new manifest making sure it's applied after `/usr/lib/microshift/manifests.d/001-microshift-ai-model-serving`.
