# Rebasing MicroShift
## Overview

MicroShift repackages a minimal core of OpenShift: A given MicroShift release is built from the same content as the corresponding OpenShift release, adding a small amount of "glue logic".

Therefore, on a high level a rebase of MicroShift onto a newer OpenShift release starts from a given OpenShift release image and involves the following steps:

* vendoring the source code of the embedded OpenShift components at the same commit as used in OpenShift,
* embedding (as bindata) the resource manifests of the hosted OpenShift components, and 
* updating the image references (digests) of the hosted OpenShift components.

This process is supported by the `rebase.sh` script, which generates the necessary changes, but currently still requires human review and post-processing. The goal is to eventually automate generation, testing, and merging PRs at least for nightlies and z-stream updates.

The following describes the current rebase process in more detail.

## Process
### Prerequisites

On the machine used for the rebase,

* install `git`, `golang` (>= 1.17), `oc`, and `jq`,
* add a pull secret into `~/.docker/config.json`, and
* git clone your personal fork of microshift and `cd` into it.

### Downloading the target OpenShift release

Run the following to download the OpenShift release to rebase to, specifying the target release image, e.g.:

```
./scripts/rebase.sh download registry.ci.openshift.org/ocp/release:4.11.0-0.nightly-2022-06-06-201913
```

This will create a directory `_output/staging`, download the specified release image's metadata (`release.txt`) and manifests, git-clone (only) the repos of the embedded components and the operators of the loaded components, and check out the commit used by that OpenShift release.

### Rebasing the go.mod file and vendoring

In MicroShift's `go.mod` file, we only explicitly add the `require` directives needed by MicroShift itself (marked with the comment `// microshift`) whereas we let `go mod tidy` figure out the direct and indirect requirements for the embedded components. For the rebase, the focus is therefore on the `replace` directives.

When resolving version mismatches between modules used by different embedded components, the general heuristic is to start from a minimal subset of ReplaceDirectives from `o/k` (only those `go mod tidy` actually tries to find) to ensure these are consistent, then add etcd's, openshift-apiserver's and openshift-controller-manager's dependencies.

The `rebase.sh` script automates updating the modulepaths (e.g. rewriting local paths like `./staging` to the global paths) and versions, but it needs hints which component's version to pick. We encode these hints as keywords in comments in MicroShift's `go.mod` file. Each replacement modulepath and versions are picked as follows:

```
// from $COMPONENT          picks from the $COMPONENT's go.mod
// release $COMPONENT       picks from the OpenShift's release image
// override                 keep the current version / do not replace
```

Run the following to rebase the `go.mod`:

```
./scripts/rebase.sh go.mod
go mod tidy
git add go.mod go.sum
git commit -m "update go.mod"
```

As we replace `k8s.io/apiserver` with the new modulepath and version picked from `openshift-apiserver` (i.e. `github.com/openshift/kubernetes-apiserver`) and this repo only carries a subset of patches that OpenShift carries in the corresponding o/k staging dir (i.e. `github.com/openshift/kubernetes/staging/src/k8s.io/apiserver`), we need to add a few missing patches to the repo after each vendoring. These patches are stored in `scripts/rebase_patches`.

To update the vendoring run:

```
go mod vendor
git apply scripts/rebase_patches/*
script/rebase.sh generated-apis
git add vendor
git commit -m "update vendoring"
```

These patches are produced by identifying the missing patches [TODO: describe process] and transforming those patches to apply against `/vendor` instead of o/k's staging dir [TODO: describe process]. This process requires some manual work but should be necessary rarely (likely after OpenShift rebases onto a upstream version).

### Rebasing `Makefile`, `Dockerfiles`, and `.spec` file
When updating to a new minor version of OpenShift, update the `RELEASE_BASE` and `GO_LD_FLAGS in the [`Makefile`](https://github.com/openshift/microshift/blob/main/Makefile) to match the new version and also check whether the Dockerfiles need a new base image (e.g. [this](https://github.com/openshift/microshift/blob/main/packaging/images/openshift-ci/Dockerfile)) and the RPM `.spec` file needs updates (e.g. [this](https://github.com/openshift/microshift/blob/main/packaging/rpm/microshift.spec))

Commit the changes:

```
git add vendor
git commit -m "update Makefile and Dockerfiles"
```

### Rebasing the embedded components
#### Component images
To update the component image references of the MicroShift release, run:

```
./scripts/rebase.sh images
git add pkg/release
git commit -m "update component images"
```

#### Component manifests
The next step is still poorly automated by `rebase.sh` and definitely requires manual review. It basically just gathers the various resource manifests used by the OpenShift control plane and the hosted components. In particular the latter are a few cases just the _templates_ of the manifests used by the hosted component's Operator, i.e. we need to substitute those for example with the variables MicroShift renders into the template or the name, namespace, or labels that the Operator would add in OpenShift.

Gather the manifests / manifest templates into the respective asset directory:

```
./scripts/rebase.sh manifests
```

Review the changes. In particular:
* restore names, namespaces, labels, and annotations that have been replaced with comments like "# X is set at runtime."
* restore the substitution of placeholders with variables for container images
* restore the change of the openshift-router's service from type `LoadBalancer` to `NodePort`
* ensure all `imagePullPolicy`s are set to `IfNotPresent` so offline mode works
* review where replica counts can be set to 1, e.g. `assets/components/service-ca/deployment.yaml`
* restore the following Operator-generated assets and compare to OpenShift whether they need updates:
```
assets/components/openshift-dns/dns/configmap.yaml
assets/components/openshift-dns/node-resolver/daemonset.yaml
assets/components/openshift-router/configmap.yaml
```

Embed the assets into the binary and commit the result:

```
./scripts/bindata.sh
git add assets pkg/assets
git commit -m "update manifests"
```

#### Component configs
The last step isn't automated at all yet, which is to compare whether the config parameters of embedded component changed, for example the kubelet configuration in `writeConfig(...)` of `pkg/node/kubelet.go` with OpenShift MCO's template (which the `rebase.sh` script downloads into `_output/staging/machine-config-operator/templates/master/01-master-kubelet/_base/files/kubelet.yaml`).

## Rebasing a Community OKD Release

The rebase script can also be used to build a community version of MicroShift that does not require access to internal image registries. To build the community version, rebase against an [OKD image found here](https://origin-release.ci.openshift.org/#4-stable) by following the same rebase script, and downloading the OKD release image [like so](#downloading-the-target-openshift-release) instead of an OpenShift release. This build is not supported or maintained and there is no guarantee that OKD releases will continue to mirror OpenShift releases.
