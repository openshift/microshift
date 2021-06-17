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
    conntrack \
    iptables-services || yum clean all -y
ADD _output/bin/microshift /microshift
ENTRYPOINT ["/microshift", "run"]
