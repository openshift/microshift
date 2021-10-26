FROM debian:sid
ARG TARGETARCH
ENV arch=$TARGETARCH
ENV PATH=$PATH:/

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates dnsutils debianutils \
                      tar wget hostname socat locate lsof gzip procps rsync python3 && \
    rm -rf /var/lib/apt/lists/* && \
    update-ca-certificates