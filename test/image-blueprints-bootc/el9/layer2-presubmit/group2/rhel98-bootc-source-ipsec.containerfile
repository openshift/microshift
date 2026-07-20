FROM localhost/rhel98-bootc-source:latest

# Install Libreswan (IPsec) and tcpdump for C2CC IPsec E2E tests.
# Libreswan provides the ipsec service and pluto daemon.
# tcpdump is used to capture and verify ESP-encapsulated packets.
RUN dnf install -y libreswan tcpdump && \
    dnf clean all

# Pre-configure firewall for IKE (UDP 500/4500) and ESP (protocol 50).
RUN firewall-offline-cmd --zone=public --add-service=ipsec
