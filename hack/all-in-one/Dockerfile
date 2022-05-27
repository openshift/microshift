# This Dockerfile should not be run directly. Instead, run ./build-aio-dev.sh
ARG IMAGE_NAME=registry.access.redhat.com/ubi8/ubi-init:8.4
FROM ${IMAGE_NAME}
ARG ARCH
ARG HOST=rhel8
ARG CPU
USER root

# If from source, expect microshift binary in current directory
RUN if [ "$FROM_SOURCE" == "true" ]; then \
        mv microshift /usr/local/bin/microshift; \
    else \
        export VERSION=$(curl -s https://api.github.com/repos/openshift/microshift/releases | grep tag_name | head -n 1 | cut -d '"' -f 4) && \
        curl -LO https://github.com/openshift/microshift/releases/download/$VERSION/microshift-linux-${ARCH} && \
        mv microshift-linux-${ARCH} /usr/local/bin/microshift; \
     fi

RUN chmod 755 /usr/local/bin/microshift

# files copied from packaging/images/microshift-aio in ./build-aio-dev.sh
COPY unit /usr/lib/systemd/system/microshift.service
COPY kubelet-cgroups.conf /etc/systemd/system.conf.d/kubelet-cgroups.conf
COPY crio-bridge.conf /etc/cni/net.d/100-crio-bridge.conf

RUN export VERSION=1.20 && \
    export OS=CentOS_8_Stream && \
    curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/devel:kubic:libcontainers:stable.repo && \
    curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable:cri-o:$VERSION/$OS/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo 

RUN dnf install -y cri-o \
        cri-tools \
        iproute \
        procps-ng && \
    dnf clean all

RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/${ARCH}/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl && \
    sed -i 's|/usr/libexec/crio/conmon|/usr/bin/conmon|' /etc/crio/crio.conf && \
    systemctl enable microshift.service && \
    systemctl enable crio 

RUN curl -s -L https://nvidia.github.io/nvidia-docker/rhel8.3/nvidia-docker.repo | tee /etc/yum.repos.d/nvidia-docker.repo && \
    dnf install nvidia-container-toolkit -y

RUN if [ "$HOST" == "rhel8" ]; then  \
      dnf install -y iptables; \
    else \
      yum install -y libnetfilter_conntrack libnfnetlink && \
      rpm -v -i --force https://archives.fedoraproject.org/pub/archive/fedora/linux/releases/28/Everything/${CPU}/os/Packages/i/iptables-libs-1.6.2-2.fc28.${CPU}.rpm \
                   https://archives.fedoraproject.org/pub/archive/fedora/linux/releases/28/Everything/${CPU}/os/Packages/i/iptables-1.6.2-2.fc28.${CPU}.rpm && \
      sed -e '/mountopt/s/,\?metacopy=on,\?//' -i /etc/containers/storage.conf; \
    fi

COPY nvidia-config.toml /etc/nvidia-container-runtime/config.toml

CMD [ "/sbin/init" ]
