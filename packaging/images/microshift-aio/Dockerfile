ARG IMAGE_NAME=registry.access.redhat.com/ubi8/ubi-init:8.4

FROM registry.access.redhat.com/ubi8/go-toolset as builder

ARG FROM_SOURCE=false

ARG ARCH
ARG IPTABLES=nft

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/redhat-et/microshift

COPY . .

USER root

# only need these args and the COPY for FROM_SOURCE="true"
ARG MAKE_TARGET
ARG BIN_TIMESTAMP
ARG SOURCE_GIT_TAG

RUN if [ "$FROM_SOURCE" == "true" ]; then \
      make clean $MAKE_TARGET SOURCE_GIT_TAG=$SOURCE_GIT_TAG BIN_TIMESTAMP=$BIN_TIMESTAMP && \
      mv _output/bin/linux_$ARCH/microshift microshift; \
    else \
      export VERSION=$(curl -s https://api.github.com/repos/redhat-et/microshift/releases | grep tag_name | head -n 1 | cut -d '"' -f 4) && \
      curl -LO https://github.com/redhat-et/microshift/releases/download/$VERSION/microshift-linux-$ARCH && \
      mv microshift-linux-$ARCH microshift; \
    fi

FROM ${IMAGE_NAME}

ARG HOST=rhel8

ARG ARCH

ENV BUILD_PATH=packaging/images/microshift-aio

COPY --from=builder /opt/app-root/src/github.com/redhat-et/microshift/microshift /usr/local/bin/microshift
COPY $BUILD_PATH/unit /usr/lib/systemd/system/microshift.service
COPY $BUILD_PATH/kubelet-cgroups.conf /etc/systemd/system.conf.d/kubelet-cgroups.conf
COPY packaging/crio.conf.d/microshift.conf /etc/crio/crio.conf.d/microshift.conf

# OCP_VERSION pushed ahead to 4.9.11 because aarch64 is now available, and it is backwards compatible
RUN export OCP_VERSION=4.9.11 && \
    curl -o oc.tar.gz https://mirror2.openshift.com/pub/openshift-v4/$ARCH/clients/ocp/$OCP_VERSION/openshift-client-linux-$OCP_VERSION.tar.gz && \
    tar -xzvf oc.tar.gz && \
    rm oc.tar.gz && \
    install -t /usr/local/bin {kubectl,oc}

RUN chmod 755 /usr/local/bin/microshift
RUN export VERSION=1.21 && \
    export OS=CentOS_8_Stream && \
    curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/devel:kubic:libcontainers:stable.repo && \
    curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/$VERSION/$OS/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo

RUN dnf install -y cri-o \
        cri-tools \
        iproute \
        procps-ng && \
    dnf clean all

RUN sed -i 's|/usr/libexec/crio/conmon|/usr/bin/conmon|' /etc/crio/crio.conf && \
    systemctl enable microshift.service && \
    systemctl enable crio

ENV KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig

# rhel7 hosts, or hosts with legacy iptables need the specific iptables binaries, ubi8 comes with iptables-nft translator

RUN if [ "$HOST" == "rhel7" ] || [ "$IPTABLES" != "nft" ]; then  \
      dnf install -y libnetfilter_conntrack libnfnetlink && dnf clean all && \
      export ARCH=$(uname -m) && \
      rpm -v -i --force https://archives.fedoraproject.org/pub/archive/fedora/linux/releases/28/Everything/$ARCH/os/Packages/i/iptables-libs-1.6.2-2.fc28.$ARCH.rpm \
                   https://archives.fedoraproject.org/pub/archive/fedora/linux/releases/28/Everything/$ARCH/os/Packages/i/iptables-1.6.2-2.fc28.$ARCH.rpm ; \
    else \
        dnf install -y iptables && dnf clean all; \
    fi

RUN if [ "$HOST" == "rhel8" ]; then \
      sed -e '/mountopt/s/,\?metacopy=on,\?//' -i /etc/containers/storage.conf; \
    fi

CMD [ "/sbin/init" ]
