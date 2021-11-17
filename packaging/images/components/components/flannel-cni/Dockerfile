# we just need cp to use during init containers, using base-image or ubi8 seems wasteful
# this needs a better strategy in the long term, probably just creating a flannel subpackage
# in RPM ? 
FROM alpine 

ARG TARGETARCH

COPY bin/flannel-$TARGETARCH /flannel

LABEL io.k8s.display-name="Flannel CNI" \
      io.k8s.description="Flannel CNI interface binaries to be copied to the host /opt/cni/bin" \
      maintainer="dev@microshift.io"
