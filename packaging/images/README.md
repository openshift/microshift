# MicroShift CI Custom Images

MicroShift CI uses custom images that are based on standard CentOS 9 Stream image
contents and a few additional tools. The image build and publishing process is
implemented as `make` rules.

Run the following command to see the available options.

```bash
$ cd packages/images
$ make
Usage: make <build | publish | manifest>
  build:     Build images locally
  publish:   Publish images at quay.io/microshift
  manifest:  Create a multi-arch manifest of the published images
```

## Build and Publish

> **Important:**<p>
> Run the build on a Red Hat CSB machine to get access to the Red Hat IT Root
> Certificate at `/etc/pki/ca-trust/source/anchors/2015-RH-IT-Root-CA.pem`.
> Otherwise, download the certificate to your build host and specify its path
> using `RHIT_CERT_FILE=/path/to/file` build option.

### Prepare Per-Platform Images

Run the following commands to build and publish the container images.

> Note: The procedure must be run both on `x86_64` and `aarch64` platforms before
> creating a multi-architecture manifest as described in the next section.

```bash
$ make build \
    RHIT_CERT_FILE=/etc/pki/ca-trust/source/anchors/2015-RH-IT-Root-CA.pem

$ make publish
```

> Important: Make sure that `quay.io/microshift/microshift-ci` repository has
> read-only public access.

### Create Multi-Architecture Manifest

Run the following command to create and publish a multi-architecture manifest
for the images.

> Note: Manifest creation should be run once on either platform as it uses the
> image references from `quay.io/microshift/microshift-ci`.

```bash
$ make manifest
```

To verify the success of the manifest creation, the command also tries to download
the images using the newly created manifest.

## Use in CI

MicroShift CI needs to be configured to mirror the custom images from the `quay.io`
repository to the CI registry. See the [Using External Images in CI](https://docs.ci.openshift.org/docs/how-tos/external-images/)
document for more information.

The image mirroring rules are defined in [this project](https://github.com/openshift/release/tree/master/core-services/image-mirroring/).
