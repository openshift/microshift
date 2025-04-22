FROM quay.io/centos-bootc/centos-bootc:stream9 
ARG REPO_CONFIG_SCRIPT=/tmp/create_repos.sh
ARG OKD_CONFIG_SCRIPT=/tmp/configure.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/rpm-repo

ENV KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
COPY --chmod=755 ./okd/src/create_repos.sh ${REPO_CONFIG_SCRIPT}
COPY --chmod=755 ./okd/src/configure.sh ${OKD_CONFIG_SCRIPT}
#COPY output/rpmbuild/RPMS ${USHIFT_RPM_REPO_PATH}

# Installing MicroShift and cleanup
# In case of flannel we don't need openvswitch service which is by default enabled as part
# once microshift is installed so better to disable it because it cause issue when required
# module is not enabled.
RUN ${REPO_CONFIG_SCRIPT} ${USHIFT_RPM_REPO_PATH} && \
    dnf install -y microshift && \
    if [ "$WITH_FLANNEL" -eq 1 ]; then \
    # replace ovn-kubernetes with flannel see https://issues.redhat.com/browse/USHIFT-4721
     sed -i 's,openshift-ovn-kubernetes,kube-flannel,' "/etc/greenboot/check/required.d/40_microshift_running_check.sh"; \
     sed -i 's,2 1 1 2,1 1 1 2,' "/etc/greenboot/check/required.d/40_microshift_running_check.sh"; \
      dnf install -y microshift-flannel; \
      systemctl disable openvswitch; \
    fi && \
    ${REPO_CONFIG_SCRIPT} -delete && \
    rm -f ${REPO_CONFIG_SCRIPT} && \
    dnf clean all
    
RUN ${OKD_CONFIG_SCRIPT} && rm -rf ${OKD_CONFIG_SCRIPT}

# Create a systemd unit to recursively make the root filesystem subtree
# shared as required by OVN images
COPY ./packaging/imagemode/systemd/microshift-make-rshared.service /usr/lib/systemd/system/microshift-make-rshared.service
RUN systemctl enable microshift-make-rshared.service
