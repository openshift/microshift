FROM registry.redhat.io/ubi8/ubi

ENV GOPATH=/go

WORKDIR $GOPATH/src/github.com/openshift/microshift

RUN dnf update -y &&  \
    dnf install --setopt=tsflags=nodocs -y \
    git \
    golang-1.17.12-1.module+el8.6.0+16014+a372c00b && \
    make \
    rpm-build \
    selinux-policy-devel\
    dnf clean all && rm -rf /var/cache/dnf/*
