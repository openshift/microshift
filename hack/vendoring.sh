#!/bin/sh

set -e

echo "Removing vendor directory..."
rm -rf vendor

echo "Downloading libraries..."
go mod vendor

echo "Applying microshift patches to libraries..."
for patch in `ls hack/patches/`; do git apply hack/patches/$patch; done


