FROM registry.access.redhat.com/ubi8/ubi-minimal
ARG TARGETARCH

RUN microdnf -y install --nodocs --setopt=install_weak_deps=0 \
                        which tar wget hostname shadow-utils socat findutils \
                        lsof bind-utils gzip procps-ng rsync iproute diffutils python3  && \
    microdnf clean all
