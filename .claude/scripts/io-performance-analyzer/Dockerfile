FROM registry.fedoraproject.org/fedora:43

RUN dnf install -y pcp pcp-system-tools pcp-export-pcp2json python3 python3-pip && \
    dnf clean all

RUN pip3 install matplotlib gsutil

WORKDIR /data

ENTRYPOINT ["/bin/bash"]
