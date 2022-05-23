# BUILD STAGE
FROM registry.access.redhat.com/ubi8/go-toolset as builder

ARG ARCH=amd64
ARG MAKE_TARGET=cross-build-linux-$ARCH
ARG BIN_TIMESTAMP
ARG SOURCE_GIT_TAG

USER root

LABEL name=microshift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/openshift/microshift

COPY . .

RUN wget https://go.dev/dl/go1.18.1.linux-amd64.tar.gz && \
    rm -rf /usr/bin/go && rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.18.1.linux-amd64.tar.gz && \
    export PATH=$PATH:/usr/local/go/bin && \
    rm go1.18.1.linux-amd64.tar.gz && \
    make clean $MAKE_TARGET SOURCE_GIT_TAG=$SOURCE_GIT_TAG BIN_TIMESTAMP=$BIN_TIMESTAMP

# RUN STAGE
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4

ARG ARCH=amd64

RUN microdnf install -y \
    policycoreutils-python-utils \
    iptables \
    && microdnf clean all
COPY --from=builder /opt/app-root/src/github.com/openshift/microshift/_output/bin/linux_$ARCH/microshift /usr/bin/microshift

RUN mkdir -p /root/crio.conf.d

COPY packaging/crio.conf.d/microshift.conf /root/crio.conf.d/microshift.conf
COPY packaging/images/microshift/entrypoint.sh /root/entrypoint.sh

ENTRYPOINT ["/root/entrypoint.sh"]
CMD ["run"]

# To start:
# podman run --privileged --ipc=host --network=host  \
# -v /var/run:/var/run \
# -v /sys:/sys:ro \
# -v /var/lib:/var/lib:rw,rshared \
# -v /lib/modules:/lib/modules \
# -v /etc:/etc \
# -v /run/containers:/run/containers \
# -v /var/log:/var/log
