#!/bin/bash

cleanup () {
    rm -f unit crio-bridge.conf kubelet-cgroups.conf
}

cp ../../packaging/images/microshift-aio/unit ../../packaging/images/microshift-aio/crio-bridge.conf ../../packaging/images/microshift-aio/kubelet-cgroups.conf .

ARCH=$(uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/")
TAG="${TAG:-quay.io/microshift/microshift-aio:$(date +%Y-%m-%d-%H-%M)}"
for img in "registry.access.redhat.com/ubi8/ubi-init:8.4" "docker.io/nvidia/cuda:11.4.2-base-ubi8"; do
   echo "build microshift aio image using base image ""${img}"
   tag=$(echo ${img} |awk -F"/" '{print $NF}'| sed -e 's/:/-/g')
   echo "${tag}"

   for host in "rhel7" "rhel8"; do
        host_tag=""
        [ "${host}" == "rhel7" ] && host_tag="-rhel7"
        CPU=$(uname -m)
        podman build --build-arg CPU=${CPU} ARCH="${ARCH}" --build-arg IMAGE_NAME=${img} --build-arg HOST=${host} -t "${TAG}"-"${tag}"${host_tag} .
   done
done
