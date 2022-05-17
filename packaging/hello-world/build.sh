#!/bin/sh

buildah build-using-dockerfile --override-arch amd64 -t quay.io/microshift/hello-world:latest-amd64 .
buildah build-using-dockerfile --override-arch arm64 -t quay.io/microshift/hello-world:latest-arm64 .
buildah manifest rm quay.io/microshift/hello-world:latest
buildah manifest create quay.io/microshift/hello-world:latest
buildah manifest add quay.io/microshift/hello-world:latest quay.io/microshift/hello-world:latest-amd64
buildah manifest add quay.io/microshift/hello-world:latest quay.io/microshift/hello-world:latest-arm64
buildah manifest push --all quay.io/microshift/hello-world:latest docker://quay.io/microshift/hello-world:latest

../rpm/paack.py copr hello-world.yaml @redhat-et/microshift-hello-world
