# Rebasing MicroShift

## Overview

MicroShift repackages a minimal core of OpenShift: A given MicroShift release is built from the same content as the corresponding OpenShift release, adding a small amount of "glue logic".

Therefore, on a high level a rebase of MicroShift onto a newer OpenShift release starts from a given OpenShift release image and involves the following steps:

* vendoring the source code of the embedded OpenShift components at the same commit as used in OpenShift,
* embedding the resource manifests of the hosted OpenShift components,
* updating the image references (digests) of the hosted OpenShift components, and
* updating build files (Makefile, Containerfile, ...).

This process is supported by the `scripts/auto-rebase/rebase.sh` script, which automates generating the changes for updating MicroShift to the given target OpenShift release.

The following describes the current rebase process in more detail.

## Process

### Prerequisites

Rebasing requires the following tools to be installed, which is already the case when running in a [MicroShift development environment](./devenv_rhel8.md):

* `git` >= 2.3
* `golang` >= 1.18
* `oc` (latest)
* `jq` >= 1.6
* `yq` >= 4.26

> There are multiple tools called `yq`, please make sure you install the one from [mikefarah](https://github.com/mikefarah/yq). See the [CI config](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/microshift/openshift-microshift-main__periodics.yaml) (in the `dockerfile_literal` of `yq-cli` image) for the authoritative version and how to install it.

When rebasing onto an OpenShift version whose release image requires a pull secret, place that secret in `~/.pull-secret.json`.

> For rebasing onto nightlies and other versions on CI, see the [OpenShift CI docs](https://docs.ci.openshift.org/docs/how-tos/use-registries-in-build-farm/#how-do-i-log-in-to-pull-images-that-require-authentication) for how to obtain a pull secret for the CI registry. The gist is: Get a CLI login token from the [app.ci web console](https://console-openshift-console.apps.ci.l2s4.p1.openshiftapps.com/), use it to log in from the CLI (`oc login https://app.ci.openshift.org --token=<token>`), and download the registry pull-secret using `oc registry login --to=app.ci_registry.json`.
>
> For rebasing onto released OpenShift versions, use the regular [OpenShift pull secret](https://cloud.redhat.com/openshift/install).
>
> You can combine both pull secrets using `jq -c -s '.[0] * .[1]' app.ci_registry.json openshift-pull-secret.json > ~/.pull-secret.json`.

Finally. git clone your personal fork of microshift and `cd` into it.

### A Note on Rebasing MicroShift's CSI Plugin

The Logical Volume Manager Service is not integrated with the ocp release image and must be passed explicitly to the rebase script as its 4th argument (including the sub-command). Images can be found at [Red Hat Container Catalog](https://catalog.redhat.com/software/containers/lvms4/lvms-operator-bundle/63972de4d8764b33ec4dbf79?tag=v4.12.0-4&architecture=amd64&push_date=1673885582000&container-tabs=gti).

### Fully automatic rebasing

The following command attempts a fully automatic rebase to a given target upstream release. It is what is run nighly from CI and should work for most cases within a z-stream. It creates a new branch named after the target release, then runs the individual steps described in the following sections, including creating the respective commits.

```shell
./scripts/auto-rebase/rebase.sh to quay.io/openshift-release-dev/ocp-release:4.10.25-x86_64 quay.io/openshift-release-dev/ocp-release:4.10.25-aarch64 registry.redhat.io/lvms4/lvms-operator-bundle:[TAG || DIGEST]
```

### Downloading the target OpenShift release

Run the following to download the OpenShift release to rebase to, specifying the target release images for _both_ Intel and Arm architectures, e.g.:

```shell
./scripts/auto-rebase/rebase.sh download quay.io/openshift-release-dev/ocp-release:4.10.25-x86_64 quay.io/openshift-release-dev/ocp-release:4.10.25-aarch64 registry.redhat.io/lvms4/lvms-operator-bundle:v4.12
```

This will create a directory `_output/staging`, download the specified release images' metadata (`release_{amd64,arm64}.json`) and manifests, git-clone (only) the repos of the embedded components and the operators of the loaded components, and check out the commit used by that OpenShift release.

### Updating the changelog

The rebase process tracks the git commits used for the embedded code and images to build a changelog of the updates in each rebase. Having the changelog in the pull request makes it easier for reviewers to understand what changes are being pulled in as part of the rebase.

```shell
./scripts/auto-rebase/rebase.sh changelog
```

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

### Rebasing the hosted component images

To update the image references for the hosted components of the MicroShift release (incl. the pause image for CRI-O), run:

```shell
./scripts/auto-rebase/rebase.sh images
git add pkg/release
git commit -m "update component images"
```

#### Rebasing the hosted component manifests

The next step is to update the manifests of embedded and hosted components in the asset directory:

```shell
./scripts/auto-rebase/rebase.sh manifests
```

For each component, this performs the following high-level steps:

1. From that component's Operator, copy the operand manifests / manifest templates into the component's asset directory. Typically, we just replace the complete directory, so removed manifests are handled correctly, but in some cases we need to preserve MicroShift-specific assets.
2. Render operand manifest templates like the operator would, e.g. filling in names, namespaces, and labels.
3. Make MicroShift-specific changes, e.g. changing the replica count to 1.
4. Replace values that MicroShift needs to fill in at runtime with the corresponding templating var. As `yq`, which is used for transforming manifests, trips over Go templating vars, this step neeeds to be done last.

Each step consistes of zero or more transformations. These transformations should remain relatively stable, but at least when rebasing to a new minor version of OpenShift, the output produced by the components Operator and that of the rebase script should be compared to make the necessary updates.

#### Rebasing the embedded component configs

The step isn't automated at all yet, which is to compare whether the config parameters of embedded components changed, for example the kubelet configuration in `writeConfig(...)` of `pkg/node/kubelet.go` with OpenShift MCO's template (which the `rebase.sh` script downloads into `_output/staging/machine-config-operator/templates/master/01-master-kubelet/_base/files/kubelet.yaml`).

### Rebasing the build files

The final step is to update build files like the `Makefiles`, `Dockerfiles`, `.spec` etc.

```shell
./scripts/auto-rebase/rebase.sh buildfiles
```

At the moment, this only updates the `GO_LD_FLAGS` in the `Makefile.kube_git.var`. When updating to a new minor version of OpenShift, you may also need to update other locations, for example:

* in the [`Makefile`](https://github.com/openshift/microshift/blob/main/Makefile)
  * the `RELEASE_BASE` variable
* in the `Dockerfile`s (e.g. [openshift-ci image](https://github.com/openshift/microshift/blob/main/packaging/images/openshift-ci/Dockerfile))
  * the version of the base images
* in the [microshift.spec](https://github.com/openshift/microshift/blob/main/packaging/rpm/microshift.spec))
  * the Golang version

Commit the changes:

```shell
git add Makefile* packaging
git commit -m "update buildfiles"
```

## Rebasing a Community OKD Release

The rebase script can also be used to build a community version of MicroShift that does not require access to internal image registries. To build the community version, rebase against an [OKD image found here](https://origin-release.ci.openshift.org/#4-stable) by following the same rebase script, and downloading the OKD release image [like so](#downloading-the-target-openshift-release) instead of an OpenShift release. This build is not supported or maintained and there is no guarantee that OKD releases will continue to mirror OpenShift releases.
