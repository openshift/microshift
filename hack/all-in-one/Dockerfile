FROM fedora:33
RUN dnf module -y enable cri-o:1.20 && dnf install -y cri-o cri-tools
RUN dnf install -y conntrack \
    iptables-services \ 
    iproute \
    procps-ng
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl
COPY microshift /usr/local/bin/microshift
COPY unit /usr/lib/systemd/system/microshift.service
RUN sed -i 's/10.85.0.0\/16/10.42.0.0\/24/' /etc/cni/net.d/100-crio-bridge.conf
RUN systemctl enable microshift.service
RUN systemctl enable crio 
CMD [ "/sbin/init" ]
