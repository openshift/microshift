# Publish Kubebuilder Tools

This is a utility to publish the kubebuilder tools archives to the kubebuilder release bucket.

The tool takes a release image and constructs 4 archives (linux/darwin x amd64/arm64) from  `installer-kube-apiserver-artifacts` and `installer-etcd-artifacts` images,
each containing the correct `etcd` and `kube-apiserver` binaries for the architecture.

The archives are then optionally published to the release buck `openshift-kubebuilder-tools` in the `openshift-gce-devel` project on GCP.

## Usage

`-pull-secret <string>`: The path to an OpenShift pull secret file that can pull from `registry.ci.openshift.org`
`-payload <string>`: The payload image that should be used to create the artifacts. This should be from the CI stream. The format will be `registry.ci.openshift.org/ocp/release:<version>`
`-version <string>`: The Kubernetes version to represent in the archives. This should be the Kubernetes release version from the payload with the `v` prefix specified. eg `v1.29.1`.
`-output-dir <string>`: A working directory to store the archives. The binaries will be extracted here and the archives will be created here.
`-skip-upload <bool>`: Skip uploading the artifacts to the GCS bucket. This can be used if you are not authenticated to GCP.
`-index-file <string>`: The path to the index file that should be updated with the new archives. This is optional and will default to `./envtest-releases.yaml`.

```bash

## Archive uploads

The tool will automatically publish the artifacts to the GCS bucket once they have been created.
To set up authentication, ensure you have logged into the gcloud CLI.

```bash
gcloud auth login
```

To skip this step, add the `-skip-upload` flag to the command.

## Using the archives

To use the archives, pass the `--remote-bucket openshift-kubebuilder-tools` flag to the envtest setup command.

```makefile
ENVTEST_K8S_VERSION = 1.29.1
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
ENVTEST = go run ${PROJECT_DIR}/vendor/sigs.k8s.io/controller-runtime/tools/setup-envtest

.PHONY: test
test: ## Run only the tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path --bin-dir $(PROJECT_DIR)/bin --remote-bucket openshift-kubebuilder-tools)" ./hack/test.sh
```
