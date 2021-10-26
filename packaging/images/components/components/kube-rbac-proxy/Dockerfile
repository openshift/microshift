FROM scratch

ARG TARGETARCH

LABEL io.k8s.display-name="kube-rbac-proxy" \
      io.k8s.description="This is a proxy, that can perform Kubernetes RBAC authorization." \
      io.openshift.tags="openshift,kubernetes" \
      summary="" \
      maintainer="OpenShift Monitoring Team <team-monitoring@redhat.com>"

COPY bin/kube-rbac-proxy-$TARGETARCH /usr/bin/kube-rbac-proxy

USER 65534
EXPOSE 8080
ENTRYPOINT ["/usr/bin/kube-rbac-proxy"]
