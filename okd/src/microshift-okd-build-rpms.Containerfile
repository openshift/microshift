FROM quay.io/centos-bootc/centos-bootc:stream9 as builder

	
ARG OKD_REPO=quay.io/okd/scos-release
ARG OKD_VERSION_TAG=4.18.0-okd-scos.4
ENV USER=microshift
ENV HOME=/microshift
ENV GOPATH=/microshift
ENV GOMODCACHE=/microshift/.cache

# Adding non-root user for building microshift
RUN useradd -m -s /bin/bash microshift -d /microshift && \
    echo 'microshift  ALL=(ALL)  NOPASSWD: ALL' >/etc/sudoers.d/microshift 
COPY . /src 
RUN chown -R microshift:microshift /microshift /src

USER 1000:1000
WORKDIR /src
# Preparing for the build
RUN echo '{"auths":{"fake":{"auth":"aWQ6cGFzcwo="}}}' > /tmp/.pull-secret && \
   /src/scripts/devenv-builder/configure-vm.sh --no-build --no-set-release-version --skip-dnf-update /tmp/.pull-secret && \
   /src/okd/src/use_okd_assets.sh --replace ${OKD_REPO} ${OKD_VERSION_TAG}

# Building Microshift RPMs and create local repo
RUN make build && \
    make rpm && \
    createrepo /src/_output/rpmbuild/RPMS/
