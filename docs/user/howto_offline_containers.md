# Embedding MicroShift Container Images for Offline Deployments

Image Builder supports building `rpm-ostree` system images with embedded container images. Embedded container images are immediately available to container engines like `podman` and `cri-o` after booting the system, without having to pull images over the network from a container registry. This means container workloads can start up without delay and without requiring network connectivity.

To embed a container image, add it to the Image Builder blueprint as follows:

```
[[containers]]
source = "<your_image_pullspec_with_tag_or_digest>"
```

To embed MicroShift container images, you need to know the exact list of container image references used by the MicroShift version you are deploying. You can obtain this list by installing the `microshift-release-info` RPM package of the same version, for example:

```
$ sudo dnf install -y microshift-release-info-4.12.0-1
$ ls /usr/share/microshift/release
release-aarch64.json  release-x86_64.json
```

Alternatively, you can download and unpack the RPM package without installing it:

```
$ sudo dnf download microshift-release-info-4.12.0-1
microshift-release-info-4.12.0-1.el8.noarch.rpm
$ rpm2cpio microshift-release-info-4.12.0-1.el8.noarch.rpm | cpio -idmv
./usr/share/microshift/release/release-aarch64.json
./usr/share/microshift/release/release-x86_64.json
```

Using the release info for your CPU architecture, you can now generate the section to embed the container images to your `blueprint.toml`:

```
$ jq -r '.images | .[] | ("[[containers]]\nsource = \"" + . + "\"\n")' release-$(uname -m).json >> blueprint.toml
```

Remember to pin the version of the MicroShift RPMs in the blueprint to the version matching your container images. The resulting `blueprint.toml` should look similar to this:

```
name = "microshift-offline"

description = ""
version = "0.0.1"
modules = []
groups = []

[[packages]]
name = "microshift"
version = "4.12.0-1"

[[containers]]
source = "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:9945c3f5475a37e145160d2fe6bb21948f1024a856827bc9e7d5bc882f44a750"

[[containers]]
source = "quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:82cfef91557f9a70cff5a90accba45841a37524e9b93f98a97b20f6b2b69e5db"

...
```

You need to add a pull secret for authenticating to the registry. To do so, set the
`auth_file_path` in the `[containers]` section of the osbuilder worker configuration in `/etc/osbuild-worker/osbuild-worker.toml ` (you might need to create directory and file).

```
[containers]
auth_file_path = "/etc/osbuild-worker/pull-secret.json"
```
You need to restart the osbuild-worker when you changed that configuration using `sudo systemctl restart osbuild-worker@1`


Refer to the [Container registry credentials](https://www.osbuild.org/guides/image-builder-on-premises/container-auth.html) section of the `osbuild` guide for more details.

Now you can `composer-cli blueprint push ...` the modified blueprint and `composer-cli compose start ...` the build as usual. The resuling commit will have the images embedded.
