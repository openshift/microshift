#!/bin/bash
# This script runs hack/all-in-one/Dockerfile
# to build a dev microshift-aio image with GPU support and kubectl installed
# FROM_SOURCE="true" ./build-aio-dev.sh # to build image with locally built binary
# ./build-aio-dev.sh # to build image with latest released binary


cleanup () {
    rm -f unit crio-bridge.conf kubelet-cgroups.conf microshift
}

trap cleanup EXIT
TAG="${TAG:-quay.io/microshift/microshift-aio:dev}"
HOST="${HOST:-rhel8}"
FROM_SOURCE="${FROM_SOURCE:-false}"
IMAGE_NAME="${IMAGE_NAME:-registry.access.redhat.com/ubi8/ubi-init:8.4}"
cp ../../packaging/images/microshift-aio/unit ../../packaging/images/microshift-aio/crio-bridge.conf ../../packaging/images/microshift-aio/kubelet-cgroups.conf .

ARCH=$(uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/")
if [ "$FROM_SOURCE" == "true" ]; then \
    pushd ../../ && \
    make  && \
    mv microshift hack/all-in-one/. && \
    popd; \
fi

podman build \
    --build-arg IMAGE_NAME="${IMAGE_NAME}" \
    --build-arg FROM_SOURCE="${FROM_SOURCE}" \
    --build-arg HOST="${HOST}" \
    --build-arg ARCH="${ARCH}" \
    -t "${TAG}" .
