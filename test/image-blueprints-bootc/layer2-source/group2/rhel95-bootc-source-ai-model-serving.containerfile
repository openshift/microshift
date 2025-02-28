FROM localhost/rhel95-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift optional packages and cleanup.
# qemu-guest-agent is used in offline tests
# vim-common for xxd utility
RUN dnf repoinfo --enabled && \
    dnf install -y \
        "microshift-ai-model-serving-{{ .Env.SOURCE_VERSION }}" \
        vim-common \
        qemu-guest-agent && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    systemctl enable qemu-guest-agent && \
    dnf clean all

# Override the default qemu-ga service configuration on the guest to allow all RPCs.
# BLOCK_RPCS is a deny-list of qemu-guest-agent RPCs to block, such as file read/write, process execution, etc. By
# default, all RPCs are blocked.
# FREEZE_HOOK_PATHNAME is the dir-path containing hook scripts. This override does not change the default value.
# These scripts are executed before or after guest image snapshot ops ("freezing" and "thawing" in libvirt speak).
# No hooks are installed by default.
RUN printf '\
\# This is a systemd environment file, not a shell script.\n\
\# It provides settings for "/lib/systemd/system/qemu-guest-agent.service".\n\
BLOCK_RPCS=\n\
FSFREEZE_HOOK_PATHNAME=/etc/qemu-ga/fsfreeze-hook\n' > /etc/sysconfig/qemu-ga

# SOURCE_IMAGES environment contains a comma-separated list of container image
# references to be set by the calling procedure. The list is prepended one more
# image used in the offline tests.
# Split the variable and pull each image in a separate layer.
#
# Note:
# - Gomplate blocks are commented out to avoid hadolint warnings.
# - Retries work around sporadic "cannot set user namespace" podman error.
#
# {{ $SOURCE_IMAGES := (printf "%s,%s" "quay.io/microshift/busybox:1.36" (.Env.SOURCE_IMAGES)) }}
# {{ range ($SOURCE_IMAGES  | strings.Split ",") }}
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    for i in 1 2 3; do \
        GOMAXPROCS=8 podman pull \
            --authfile /run/secrets/pull-secret.json \
            --root /var/lib/containers/storage-preloaded \
            "docker://{{ . }}" && break; \
        if [ $i -eq 3 ] ; then \
            echo "ERROR: Failed to pull {{ . }} image after 3 attempts"; \
            exit 1; \
        fi; \
        echo "Attempt $i failed. Retrying in 5 seconds..." && sleep 5; \
    done
# {{ end }}

# SOURCE_IMAGES_AIMS environment contains a list of container images related to
# AI Model Serving.
# {{ $SOURCE_IMAGES_AIMS := (printf "%s,%s" "quay.io/pmatusza/microshift-testing:resnet" (.Env.SOURCE_IMAGES_AIMS)) }}
# {{ range ($SOURCE_IMAGES_AIMS | strings.Split ",") }}
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    for i in 1 2 3; do \
        GOMAXPROCS=8 podman pull \
            --authfile /run/secrets/pull-secret.json \
            --root /var/lib/containers/storage-preloaded \
            "docker://{{ . }}" && break; \
        if [ $i -eq 3 ] ; then \
            echo "ERROR: Failed to pull {{ . }} image after 3 attempts"; \
            exit 1; \
        fi; \
        echo "Attempt $i failed. Retrying in 5 seconds..." && sleep 5; \
    done
# {{ end }}

# Edit the container storage configuration file to include the new path
RUN sed -i '/^additionalimagestores.*/a\   "/var/lib/containers/storage-preloaded",' /etc/containers/storage.conf

# Apply a workaround to set the SELinux context on the new storage directory and
# also restore 'NET_BIND_SERVICE' capability that is currently lost when including
# images in the container.
#
# Note: This requires setting the additional image stores path to a read-write
# location on the file system. The images will still be treated as read-only by
# the container subsystem.
# See https://github.com/ostreedev/ostree-rs-ext/issues/654
COPY --chmod=755 ./bootc-images/microshift-imagestore-config.sh /usr/bin/microshift-imagestore-config
COPY --chmod=644 ./bootc-images/microshift-imagestore-config.service /etc/systemd/system/microshift-imagestore-config.service
RUN systemctl enable microshift-imagestore-config.service

# Test data
RUN mkdir /etc/microshift/manifests.d/10-ai-model-serving-test/

RUN cat <<EOF > /etc/microshift/manifests.d/10-ai-model-serving-test/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- resources.yaml
EOF

RUN cat <<EOF > /etc/microshift/manifests.d/10-ai-model-serving-test/resources.yaml
apiVersion: v1
kind: Namespace
metadata:
    name: test-ai
---
# ServingRuntime based on what is shipped in the redhat-ods-applications namespace.
# Needs to be copied to workload namespace (test-ai).
apiVersion: serving.kserve.io/v1alpha1
kind: ServingRuntime
metadata:
  annotations:
    opendatahub.io/recommended-accelerators: '["nvidia.com/gpu"]'
    openshift.io/display-name: OpenVINO Model Server
  labels:
    opendatahub.io/dashboard: "true"
  name: kserve-ovms
  namespace: test-ai
spec:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8888"
  containers:
  - args:
    - --model_name={{`{{.Name}}`}}
    - --port=8001
    - --rest_port=8888
    - --model_path=/mnt/models
    - --file_system_poll_wait_seconds=0
    - --grpc_bind_address=0.0.0.0
    - --rest_bind_address=0.0.0.0
    - --target_device=AUTO
    - --metrics_enable
    image:  {{ .Env.SOURCE_IMAGES_AIMS_OVMS }}
    name: kserve-container
    ports:
    - containerPort: 8888
      protocol: TCP
  multiModel: false
  protocolVersions:
  - v2
  - grpc-v2
  supportedModelFormats:
  - autoSelect: true
    name: openvino_ir
    version: opset13
  - name: onnx
    version: "1"
  - autoSelect: true
    name: tensorflow
    version: "1"
  - autoSelect: true
    name: tensorflow
    version: "2"
  - autoSelect: true
    name: paddle
    version: "2"
  - autoSelect: true
    name: pytorch
    version: "2"
---
# Definition of InferenceService with a model packages in form of OCI image.
# Features extra argument for the model server (--layout), so the data layout
# format expected by the model server matches what we send during testing.
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: openvino-resnet
  namespace: test-ai
spec:
  predictor:
    model:
      protocolVersion: v2
      modelFormat:
        name: openvino_ir
      storageUri: "oci://quay.io/pmatusza/microshift-testing:resnet"
      args:
      - --layout=NHWC:NCHW
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: openvino-resnet-predictor
  namespace: test-ai
spec:
  host: openvino-resnet-predictor-test-ai.apps.example.com
  port:
    targetPort: 8888
  to:
    kind: Service
    name: openvino-resnet-predictor
    weight: 100
  wildcardPolicy: None
EOF

# Download an image from the examples to use to test the inference.
RUN curl -o /etc/microshift/manifests.d/10-ai-model-serving-test/bee.jpeg https://raw.githubusercontent.com/openvinotoolkit/model_server/main/demos/common/static/images/bee.jpeg

# Create script to prepare the query because using QEMU Guest Agent
# involves multiple layers of JSON and quoting, which is difficult to solve.
RUN cat <<'EOF' > /etc/microshift/manifests.d/10-ai-model-serving-test/prepare-query.sh
#!/usr/bin/env bash
set -xeuo pipefail

OUTPUT=/tmp/request.json
PAYLOAD=/etc/microshift/manifests.d/10-ai-model-serving-test/bee.jpeg

# Add an inference header (len=63)
echo -n '{"inputs" : [{"name": "0", "shape": [1], "datatype": "BYTES"}]}' > "${OUTPUT}"

# Add size of the data (image) in binary format (4 bytes, little endian)
printf "%08X" $(stat --format=%s "${PAYLOAD}") | sed 's/\(..\)/\1\n/g' | tac | tr -d '\n' | xxd -r -p >> "${OUTPUT}"

# Add the data, i.e. the image
cat "${PAYLOAD}" >> "${OUTPUT}"
EOF
