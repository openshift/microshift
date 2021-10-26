ARG OKD_TAG
ARG REGISTRY
FROM $REGISTRY/base-image:$OKD_TAG

ARG TARGETARCH

COPY bin/coredns-$TARGETARCH /usr/bin/coredns

ENTRYPOINT ["/usr/bin/coredns"]

EXPOSE 53 53/udp

LABEL io.k8s.display-name="CoreDNS" \
      io.k8s.description="CoreDNS delivers the DNS and Discovery Service for a Kubernetes cluster." \
      maintainer="dev@lists.openshift.redhat.com"
