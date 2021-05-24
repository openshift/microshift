FROM registry.access.redhat.com/ubi8/go-toolset
USER root
LABEL name=ushift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/microshift

VOLUME go_cache:/mnt/cache

RUN yum install gpgme-devel libassuan-devel -y

ENTRYPOINT ["make", "build", "DO_LOCAL=0"]