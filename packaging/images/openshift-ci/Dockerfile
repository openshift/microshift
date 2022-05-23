FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.17-openshift-4.10 AS builder
USER root
LABEL name=microshift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/openshift/microshift

COPY . .

RUN yum install gpgme-devel glibc-static libassuan-devel -y

# clean out binaries that may have been copied in form build context before running target
RUN make clean cross-build-linux-amd64 cross-build-linux-arm64

RUN ls -la

#Linux ARM 64
FROM  registry.ci.openshift.org/ocp/4.10

COPY --from=builder /opt/app-root/src/github.com/openshift/microshift/_output/bin/linux_arm64/microshift /usr/bin/local/microshift

#Linux AMD 64
FROM  registry.ci.openshift.org/ocp/4.10

COPY --from=builder /opt/app-root/src/github.com/openshift/microshift/_output/bin/linux_amd64/microshift /usr/bin/local/microshift
