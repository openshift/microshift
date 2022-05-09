# https://github.com/openshift/router/blob/master/images/router/haproxy/Dockerfile.rhel8
FROM fedora-minimal:36
# ubi-8 images don't have haproxy22, so we rely on fedora-minimal:36 in this case

ARG TARGETARCH
RUN INSTALL_PKGS="haproxy rsyslog procps-ng util-linux" && \
    microdnf install -y --nodocs --setopt=install_weak_deps=0 $INSTALL_PKGS && \
    microdnf clean all

RUN mkdir -p /var/lib/haproxy/router/{certs,cacerts,whitelists} && \
    mkdir -p /var/lib/haproxy/{conf/.tmp,run,bin,log} && \
    touch /var/lib/haproxy/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config} && \
    setcap 'cap_net_bind_service=ep' /usr/sbin/haproxy && \
    chown -R :0 /var/lib/haproxy && \
    chmod -R g+w /var/lib/haproxy && \
    sed -i 's/SECLEVEL=2/SECLEVEL=1/g' /etc/crypto-policies/back-ends/opensslcnf.config

COPY src/images/router/haproxy/ /var/lib/haproxy/

LABEL io.k8s.display-name="OpenShift HAProxy Router" \
      io.k8s.description="This component offers ingress to an OpenShift cluster via Ingress and Route rules." \
      io.openshift.tags="openshift,router,haproxy"

COPY bin/openshift-router-$TARGETARCH /usr/bin/openshift-router

USER 1001
EXPOSE 80 443
WORKDIR /var/lib/haproxy/conf
ENV TEMPLATE_FILE=/var/lib/haproxy/conf/haproxy-config.template \
    RELOAD_SCRIPT=/var/lib/haproxy/reload-haproxy

ENTRYPOINT ["/usr/bin/openshift-router", "--v=2"]
