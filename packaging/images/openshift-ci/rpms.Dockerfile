FROM registry.redhat.io/rhel8/go-toolset:1.17

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/openshift/microshift

RUN yum install gpgme-devel glibc-static libassuan-devel -y

COPY . .

RUN make clean rpms