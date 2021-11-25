FROM debian:sid
ARG TARGETARCH

COPY bin/openshift-router-$TARGETARCH /usr/bin/openshift-router

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates haproxy rsyslog sysvinit-utils libcap2-bin curl && \
    rm -rf /var/lib/apt/lists/* && \
    update-ca-certificates

RUN bash -c 'mkdir -p /var/lib/haproxy/router/{certs,cacerts,whitelists} && \
    mkdir -p /var/lib/haproxy/{conf/.tmp,run,bin,log} && \
    touch /var/lib/haproxy/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}' && \
    setcap 'cap_net_bind_service=ep' /usr/sbin/haproxy && \
    chown -R :0 /var/lib/haproxy && \
    chmod -R g+w /var/lib/haproxy

COPY src/images/router/haproxy/ /var/lib/haproxy/

LABEL io.k8s.display-name="OpenShift HAProxy Router" \
      io.k8s.description="This component offers ingress to an OpenShift cluster via Ingress and Route rules." \
      io.openshift.tags="openshift,router,haproxy" \
      maintainer="Carlos Eduardo <carlosedp@gmail.com>"

USER 1001
EXPOSE 80 443
WORKDIR /var/lib/haproxy/conf
ENV TEMPLATE_FILE=/var/lib/haproxy/conf/haproxy-config.template \
    RELOAD_SCRIPT=/var/lib/haproxy/reload-haproxy
ENTRYPOINT ["/usr/bin/openshift-router", "--v=2"]