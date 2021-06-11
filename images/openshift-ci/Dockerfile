FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.8 AS builder
USER root
LABEL name=microshift-build

ARG TARGET=build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/redhat-et/microshift

COPY . .

RUN yum install gpgme-devel glibc-static libassuan-devel -y

# clean out binaries that may have been copied in form build context before running target
RUN make clean "${TARGET}"

RUN ls -la

# container disk
FROM  registry.ci.openshift.org/ocp/4.8 AS binary-build

ARG ARTIFACT_SRC=''

RUN mkdir -p /opt/microshift/

COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/${ARTIFACT_SRC} /opt/microshift/
