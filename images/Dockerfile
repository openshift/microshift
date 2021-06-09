FROM registry.access.redhat.com/ubi8/go-toolset
USER root
LABEL name=microshift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/microshift

RUN yum install gpgme-devel glibc-static libassuan-devel -y

ENTRYPOINT ["make"]