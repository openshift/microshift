FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.23-openshift-4.19 AS builder
WORKDIR /go/src/github.com/openshift/route-controller-manager
COPY . .
RUN make build --warn-undefined-variables

FROM registry.ci.openshift.org/ocp/4.19:base-rhel9
COPY --from=builder /go/src/github.com/openshift/route-controller-manager/route-controller-manager /usr/bin/
LABEL io.k8s.display-name="OpenShift Route Controller Manager Command" \
      io.k8s.description="OpenShift is a platform for developing, building, and deploying containerized applications." \
      io.openshift.tags="openshift,route-controller-manager"
