FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.8 AS builder
USER root
LABEL name=microshift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/redhat-et/microshift

COPY . .

RUN yum install gpgme-devel glibc-static libassuan-devel -y

# clean out binaries that may have been copied in form build context before running target
RUN make build-containerized-cross-build

RUN ls -la

# container disk
FROM  registry.ci.openshift.org/ocp/4.8

RUN mkdir -p /opt/microshift/

COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/_output/bin/darwin_amd64/microshift /usr/share/redhat-et/darwin_amd64/microshift
COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/_output/bin/linux_amd64/microshift /usr/share/redhat-et/linux_amd64/microshift
COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/_output/bin/linux_arm64/microshift /usr/share/redhat-et/linux_arm64/microshift
COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/_output/bin/linux_ppc64le/microshift /usr/share/redhat-et/linux_ppc64le/microshift
COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/_output/bin/linux_s390x/microshift /usr/share/redhat-et/linux_s390x/microshift
#COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/_output/bin/windows_amd64/microshift.exe /usr/share/redhat-et/windows/oc.exe
COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/LICENSE /usr/share/redhat-et/LICENSE
