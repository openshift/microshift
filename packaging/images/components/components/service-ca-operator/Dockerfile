ARG OCP_TAG
ARG REGISTRY
FROM $REGISTRY/base-image:$OCP_TAG

ARG TARGETARCH

COPY bin/service-ca-operator-$TARGETARCH /usr/bin/service-ca-operator

COPY src/manifests /manifests
# Using the vendored CRD ensures compatibility with 'oc explain'
COPY src/vendor/github.com/openshift/api/operator/v1/0000_50_service-ca-operator_02_crd.yaml /manifests/02_crd.yaml
ENTRYPOINT ["/usr/bin/service-ca-operator"]
LABEL io.openshift.release.operator=true
