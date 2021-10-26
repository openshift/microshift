# Multiarchitecture side-image building

MicroShift optionally deploys multiple container images at runtime to provide
infrastructure services like DNS and ingress. While most of those
images are available from OKD and it does consume the original unmodified
OKD image, this is only true for `amd64` because OKD still lacks support for other
architectures MicroShift supports like `arm`, `arm64`, `ppc64le`, `riscv64` 

# Usage

`./build.sh` will generate all the necessary images, and it accepts a few parameters
as environment variables you can use to tweak and debug the build:

* `DEST_REGISTRY=quay.io/microshift`
* `COMPONENTS="base-image pause cli coredns flannel haproxy-router hostpath-provisioner kube-rbac-proxy service-ca-operator"`
* `ARCHITECTURES="amd64 arm64 arm ppc64le riscv64"`
* `PUSH=no`
* `PARALLEL=yes`

You can use an alternate DEST_REGISTRY for testing, pick individual components or architectures.

For example:
`COMPONENTS="base-image" ./build.sh`

The base-image is used for composing some of the images on top, as OKD does. This serves two purposes:
    * Faster multiarch building, since the initial layers are very expensive to construct (done via
      qemu-static)
    * Thinner storage, as the base-image layer is downloaded just once.

## Existing issues

Currently buildah doesn't recognize and use the locally built base-image for some
reason to be identified. That means that local building doesn't properly work without build.

As a workaround when locally testing for a new tag you can use a separate registry with
a "base-image" repo.

For example:

`COMPONENTS="base-image" PUSH=yes DEST_REGISTRY=quay.io/microshift-mangelajo ./build.sh`

This will create and push the base image, and allow you to build the rest of the images
without pushing.

`COMPONENT="x y z" DEST_REGISTRY=quay.io/microshift-mangelajo ./build.sh`

## Directory structure

The `components` directory inside `side-images` contains all necessary information to
build each image. 

Each component source code is extracted into the `src` directory, the references
are extracted from the specific OKD release, and for components not
being part of OKD the `repo` and `commit` files should exist (except for the base image)
(see flannel for an example).

In addition, each component can have:
* `ImageSource.$ARCH` or `Dockerfile.$ARCH` specific for an architecture.
* `ImageSource` or `Dockerfile` general building strategy.

An `ImageSource` file means that if no other specific method exist for an architecture
the image should be retrieved from an specific ImageSource, for example in `flannel` we use
`quay.io/coreos/flannel:v0.14.0` as ImageSource for most architectures, since they publish
a multi-architecture manifest.

A `Dockerfile` file means that if no other specific method exist for an architecture,
the image will be built according to the instructions of the Dockerfile.

An `ImageSource.$ARCH` or `Dockerfile.$ARCH` will source or build an image for an specific
architecture.

Each directory can have a `build_binaries` script which should be responsible of building
the different architecture binaries under `bin`, will be triggered only when necessary.

# Consumed images

The reference to the consumed images can be found in [pkg/release](../pkg/release).

# Image sources and source code

The available OKD images, and otherwise the reference to the sourcecode and git-tag
from which the OKD images are built is extracted from
`oc adm release extract "quay.io/openshift/okd:${OKD_BASE_TAG}" --file=image-references` 

If an OKD image exists for the specific architecture, such specific image will be
added into the multiarch manifest, otherwise we need to build the specific images.

For architectures where `ubi8` or `ubi8-minimal` images exist such base will be used,
in some cases we use `fedora-minimal` (when a newer version of packages is necessary)

# Non OKD images
We consume a few non-okd images, like `flannel`, `hostpath-provisioner`, `pause`,
we build those images from exiting image sources, or from source code.

# generated images
We publish the multi-arch images under quay.io/microshift/$IMAGE:$OKD_BASE_TAG
