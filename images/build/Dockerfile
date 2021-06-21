FROM registry.access.redhat.com/ubi8/go-toolset as build

USER root

LABEL name=microshift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache GO111MODULE=on

WORKDIR $GOPATH/src/github.com/redhat-et/microshift

RUN yum install glibc-static -y

COPY . .

RUN make clean build

# Containerized microshift
# To start:
# podman run --privileged --ipc=host --network=host  \
# -v /var/run:/var/run \
# -v /sys:/sys:ro \
# -v /var/lib:/var/lib:rw,rshared \
# -v /lib/modules:/lib/modules \
# -v /etc:/etc \
# -v /run/containers:/run/containers \
# -v /var/log:/var/log   microshift

FROM  registry.access.redhat.com/ubi8/ubi

RUN yum install -y \
    policycoreutils-python-utils \
    iptables

RUN yum clean all -y

COPY --from=build /opt/app-root/src/github.com/redhat-et/microshift/microshift /usr/bin/microshift

ENTRYPOINT ["/usr/bin/microshift", "run"]
