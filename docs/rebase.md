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

* install `git`, `golang` (>= 1.17), `oc`, `yq`, and `jq`,
* add a pull secret into `~/.pull-secret.json`, and
* git clone your personal fork of microshift and `cd` into it.

### Fully automatic rebaseing

The following command attempts a fully automatic rebase to a given target upstream release. It is what is run nighly from CI and should work for most cases within a z-stream. It creates a new branch namded after the target release, then runs the indidivual steps described in the following sections, including creating the respective commmits.

```shell
./scripts/auto-rebase/rebase.sh to quay.io/openshift-release-dev/ocp-release:4.10.25-x86_64 quay.io/openshift-release-dev/ocp-release:4.10.25-aarch64
```

### Downloading the target OpenShift release

Run the following to download the OpenShift release to rebase to, specifying the target release images for _both_ Intel and Arm architectures, e.g.:

```shell
./scripts/auto-rebase/rebase.sh download quay.io/openshift-release-dev/ocp-release:4.10.25-x86_64 quay.io/openshift-release-dev/ocp-release:4.10.25-aarch64
```

This will create a directory `_output/staging`, download the specified release images' metadata (`release_{amd64,arm64}.json`) and manifests, git-clone (only) the repos of the embedded components and the operators of the loaded components, and check out the commit used by that OpenShift release.

### Rebasing the go.mod file and vendoring

In MicroShift's `go.mod` file, we only explicitly add the `require` directives needed by MicroShift itself (marked with the comment `// microshift`) whereas we let `go mod tidy` figure out the direct and indirect requirements for the embedded components. For the rebase, the focus is therefore on the `replace` directives.

When resolving version mismatches between modules used by different embedded components, the general heuristic is to start from a minimal subset of ReplaceDirectives from `o/k` (only those `go mod tidy` actually tries to find) to ensure these are consistent, then add etcd's and route-controller-manager's dependencies.

The `rebase.sh` script automates updating the modulepaths (e.g. rewriting local paths like `./staging` to the global paths) and versions, but it needs hints which component's version to pick. We encode these hints as keywords in comments in MicroShift's `go.mod` file. Each replacement modulepath and versions are picked as follows:

```shell
// from $COMPONENT          picks from the $COMPONENT's go.mod
// staging kubernetes       picks the version in openshift/kubernetes's staging directory
// release $COMPONENT       picks from the OpenShift's release image
// override                 keep the current version / do not replace
```

Run the following to rebase the `go.mod`:

```shell
./scripts/auto-rebase/rebase.sh go.mod
go mod tidy
git add go.mod go.sum
git commit -m "update go.mod"
```

As we're vendoring from multiple OpenShift components, there may be a situation in which we need to pick a module version for one component that is not completely aligned with another component's version (like we've had when still vendoring `openshift-apiserver`). In this case, it may be necessary to resolve the conflict through patches to the vendored modules. These patches would then be stored in `scripts/auto-rebase/rebase_patches`.

To update the vendoring run:

```shell
make vendor
git add vendor
git commit -m "update vendoring"
```

### Rebasing `Makefile`, `Dockerfiles`, and `.spec` file

When updating to a new minor version of OpenShift, update the `RELEASE_BASE` and `GO_LD_FLAGS in the [`Makefile`](https://github.com/openshift/microshift/blob/main/Makefile) to match the new version and also check whether the Dockerfiles need a new base image (e.g. [this](https://github.com/openshift/microshift/blob/main/packaging/images/openshift-ci/Dockerfile)) and the RPM `.spec` file needs updates (e.g. [this](https://github.com/openshift/microshift/blob/main/packaging/rpm/microshift.spec))

Commit the changes:

```shell
git add Makefile packaging
git commit -m "update Makefile and Dockerfiles"
```

### Rebasing the embedded components

#### Component images

To update the component image references of the MicroShift release, run:

```shell
./scripts/auto-rebase/rebase.sh images
git add pkg/release
git commit -m "update component images"
```

#### Component manifests

The next step is to update the embedded component manifests in the asset directory:

```shell
./scripts/auto-rebase/rebase.sh manifests
```

For each component, this performs the following high-level steps:

1. From that component's Operator, copy the operand manifests / manifest templates into the component's asset directory. Typically, we just replace the complete directory, so removed manifests are handled correctly, but in some cases we need to preserve MicroShift-specific assets.
2. Render operand manifest templates like the operator would, e.g. filling in names, namespaces, and labels.
3. Make MicroShift-specific changes, e.g. changing the replica count to 1.
4. Replace values that MicroShift needs to fill in at runtime with the corresponding templating var. As `yq`, which is used for transforming manifests, trips over Go templating vars, this step neeeds to be done last.

Each step consistes of zero or more transformations. These transformations should remain relatively stable, but at least when rebasing to a new minor version of OpenShift, the output produced by the components Operator and that of the rebase script should be compared to make the necessary updates.

#### Component configs

The last step isn't automated at all yet, which is to compare whether the config parameters of embedded component changed, for example the kubelet configuration in `writeConfig(...)` of `pkg/node/kubelet.go` with OpenShift MCO's template (which the `rebase.sh` script downloads into `_output/staging/machine-config-operator/templates/master/01-master-kubelet/_base/files/kubelet.yaml`).

## Rebasing a Community OKD Release

The rebase script can also be used to build a community version of MicroShift that does not require access to internal image registries. To build the community version, rebase against an [OKD image found here](https://origin-release.ci.openshift.org/#4-stable) by following the same rebase script, and downloading the OKD release image [like so](#downloading-the-target-openshift-release) instead of an OpenShift release. This build is not supported or maintained and there is no guarantee that OKD releases will continue to mirror OpenShift releases.
