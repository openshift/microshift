# Dockerfile.openshift-tests builds the openshift-tests binary.  This enables openshift-ci conformance testing.
FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.8 as openshift-tests-builder

WORKDIR /go/src/github.com/openshift/origin/

RUN dnf install -y git make

RUN git clone https://github.com/openshift/origin.git . && \
    git checkout release-4.8

RUN make

# Discard the build-time dependencies.  This container image is used to store the binary, which is why there is no entrypoint.
FROM 

COPY --from=openshift-tests-builder /go/src/github.com/openshift/origin/openshift-tests /usr/bin/openshift-tests
