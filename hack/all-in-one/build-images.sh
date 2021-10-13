#!/bin/bash

TAG="${TAG:-quay.io/microshift/microshift-aio:$(date +%Y-%m-%d-%H-%M)}"
for img in "registry.access.redhat.com/ubi8/ubi-init:8.4" "docker.io/nvidia/cuda:11.4.2-base-ubi8"; do
   echo "build microshift aio image using base image "${i}
   tag=$(echo ${img} |awk -F"/" '{print $NF}'| sed -e 's/:/-/g')
   echo ${tag}

   for host in "rhel7" "rhel8"; do
        host_tag=""
        [ "${host}" == "rhel7" ] && host_tag="-rhel7"
        podman build --build-arg IMAGE_NAME=${img} --build-arg HOST=${host} -t ${TAG}-${tag}${host_tag} .
   done
done
