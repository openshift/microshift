# MicroShift CI Custom Images

MicroShift CI uses custom images that are based on standard CentOS 9 Stream image
contents and a few additional tools. The image build and publishing process is
implemented as `make` rules.

Run the following command to see the available options.

```bash
$ cd packages/images
$ make
Usage: make <build | publish>
  build:     Build images locally
  publish:   Publish images at quay.io/microshift
```

## Build and Publish

Run the following command to build the container images for all the supported configurations.

```bash
$ make build
```

Run the following commands to publish the container images to `quay.io/microshift`
repository.

```bash
$ podman login quay.io/microshift
$ make publish
```

> Important: Make sure that `quay.io/microshift/microshift-ci` repository has
> read-only public access.

## Use in CI

MicroShift CI needs to be configured to mirror the custom images from the `quay.io`
repository to the CI registry. See the [Using External Images in CI](https://docs.ci.openshift.org/docs/how-tos/external-images/)
document for more information.

The image mirroring rules are defined in [this project](https://github.com/openshift/release/tree/master/core-services/image-mirroring/microshift).
