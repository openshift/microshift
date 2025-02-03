FROM quay.io/centos-bootc/centos-bootc:stream9 as builder

	
ARG OKD_REPO=quay.io/okd/scos-release
ARG OKD_VERSION_TAG=4.17.0-0.okd-scos-2024-08-21-100712
ARG REPO_DIR=/src/_output/rpmbuild/RPMS/
ENV USER=microshift
ENV HOME=/microshift
ENV GOPATH=/microshift
ENV GOMODCACHE=/microshift/.cache

# Adding non-root user for building microshift
RUN useradd -m -s /bin/bash microshift -d /microshift && \
    echo 'microshift  ALL=(ALL)  NOPASSWD: ALL' >/etc/sudoers.d/microshift && \
    chmod 0640 /etc/shadow
COPY . /src 
RUN chown -R microshift:microshift /microshift /src

USER microshift:microshift
WORKDIR /src
# Preparing for the build
RUN echo '{"auths":{"fake":{"auth":"aWQ6cGFzcwo="}}}' > /tmp/.pull-secret && \
   /src/scripts/devenv-builder/configure-vm.sh --no-build --no-set-release-version --skip-dnf-update /tmp/.pull-secret && \
   /src/okd/src/use_okd_assets.sh --replace ${OKD_REPO} ${OKD_VERSION_TAG}

# Building Microshift RPMs and local repo
RUN make build && \
    make rpm && \
    createrepo ${REPO_DIR}

# Building microshift container from local rpms
FROM quay.io/centos-bootc/centos-bootc:stream9 
ARG REPO_CONFIG_SCRIPT=/tmp/create_repos.sh
ARG OKD_CONFIG_SCRIPT=/tmp/configure.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/rpm-repo

ENV KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
COPY --chmod=755 ./okd/src/create_repos.sh ${REPO_CONFIG_SCRIPT}
COPY --chmod=755 ./okd/src/configure.sh ${OKD_CONFIG_SCRIPT}
COPY --from=builder /src/_output/rpmbuild/RPMS ${USHIFT_RPM_REPO_PATH}

# Install nfv-openvswitch repo which provides openvswitch extra policy package
RUN dnf install -y centos-release-nfv-openvswitch && \
    dnf clean all

# Installing MicroShift and cleanup
# In case of flannel we don't need openvswitch service which is by default enabled as part
# once microshift is installed so better to disable it because it cause issue when required
# module is not enabled.
RUN ${REPO_CONFIG_SCRIPT} ${USHIFT_RPM_REPO_PATH} && \
    dnf install -y microshift microshift-release-info && \
    if [ -n "$WITH_FLANNEL" ] ; then \
        dnf install -y microshift-flannel ; \
        systemctl disable openvswitch ; \
    fi && \
    ${REPO_CONFIG_SCRIPT} -delete && \
    rm -f ${REPO_CONFIG_SCRIPT} && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
    
RUN ${OKD_CONFIG_SCRIPT} && rm -rf ${OKD_CONFIG_SCRIPT}

# If the EMBED_CONTAINER_IMAGES environment variable is set to 1:
# 1. Temporarily configure user namespace UID and GID mappings by writing to /etc/subuid and /etc/subgid and clean it later
#    - This allows the skopeo command to operate properly which requires user namespace support.
#    - Without it following error occur during image build
#       - FATA[0129] copying system image from manifest list:[...]unpacking failed (error: exit status 1; output: potentially insufficient UIDs or GIDs available[...]
# 2. Extract the list of image URLs from a JSON file (`release-$(uname -m).json`) while excluding the "lvms_operator" image.
#    - `lvms_operator` image is excluded because it is not available upstream
RUN if [ -n "$EMBED_CONTAINER_IMAGES" ] ; then \
        echo "root:100000:65536" > /etc/subuid ; \
        echo "root:100000:65536" > /etc/subgid ; \
        for i in $(jq -r '.images | to_entries | map(select(.key != "lvms_operator")) | .[].value' "/usr/share/microshift/release/release-$(uname -m).json") ; do \
            skopeo copy --retry-times 3 --preserve-digests "docker://${i}" "containers-storage:${i}" ; \
        done ; \
        rm -f /etc/subuid /etc/subgid ; \
    fi

# Create a systemd unit to recursively make the root filesystem subtree
# shared as required by OVN images
COPY ./packaging/imagemode/systemd/microshift-make-rshared.service /etc/systemd/system/microshift-make-rshared.service
RUN systemctl enable microshift-make-rshared.service
