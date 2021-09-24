#!/bin/bash

TAG="${TAG:-quay.io/microshift/microshift-aio:$(date +%Y-%m-%d-%H-%M)}"
for img in "registry.access.redhat.com/ubi8/ubi-init:8.4" "docker.io/nvidia/cuda:11.4.2-base-ubi8"; do
   echo "build microshift aio image using base image "${i}
   tag=$(echo ${img} |awk -F"/" '{print $NF}'| sed -e 's/:/-/g')
   echo ${tag}

   for iptables in "nft" "legacy-iptables"; do
        iptables_tag=""
        [ "${iptables}" == "legacy-iptables" ] && iptables_tag = "-legacy-iptables"
        podman build --build-arg IMAGE_NAME=${img} --build-arg IPTABLES=${iptables} -t ${TAG}-${tag}${iptables_tag} .
   done
done
