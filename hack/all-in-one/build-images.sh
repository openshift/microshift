#!/bin/bash

TAG="${TAG:-quay.io/microshift/microshift-aio-$(date +%Y-%m-%d-%H-%M)}"
for i in "registry.access.redhat.com/ubi8/ubi-init:8.4" "docker.io/nvidia/cuda:11.4.2-base-ubi8"; do
   echo "build microshift aio image using base image "${i}
   tag=$(echo ${i} |awk -F"/" '{print $NF}'| sed -e 's/:/-/g')
   echo ${tag}
   podman build --build-arg IMAGE_NAME=${i} -t ${TAG}-${tag} .
done