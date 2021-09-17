ARG IMAGE_NAME=registry.access.redhat.com/ubi8/ubi-init:8.4
FROM ${IMAGE_NAME}

RUN export VERSION=$(curl -s https://api.github.com/repos/redhat-et/microshift/releases | grep tag_name | head -n 1 | cut -d '"' -f 4) && \
    export ARCH=$(uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/") && \
    curl -LO https://github.com/redhat-et/microshift/releases/download/$VERSION/microshift-linux-${ARCH} && \
    mv microshift-linux-${ARCH} /usr/local/bin/microshift && chmod 755 /usr/local/bin/microshift

COPY unit /usr/lib/systemd/system/microshift.service
COPY kubelet-cgroups.conf /etc/systemd/system.conf.d/kubelet-cgroups.conf
COPY crio-bridge.conf /etc/cni/net.d/100-crio-bridge.conf

RUN export VERSION=1.20 && \
    export OS=CentOS_8_Stream && \
    curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/devel:kubic:libcontainers:stable.repo && \
    curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable:cri-o:$VERSION/$OS/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo 

RUN dnf install -y cri-o \
        cri-tools \
        iptables \ 
        iproute \
        procps-ng && \
    dnf clean all

RUN export ARCH=$(uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/") && \
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/${ARCH}/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl && \
    systemctl enable microshift.service && \
    systemctl enable crio 

CMD [ "/sbin/init" ]
