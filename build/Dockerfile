FROM golang:1.16
LABEL name=microshift-build

ENV GOPATH=/opt/app-root

WORKDIR $GOPATH/src/github.com/microshift

VOLUME /opt/app-root/src/microshift

RUN \
printf "%s\n" \
'deb http://deb.debian.org/debian stretch main' \
'deb-src htt://deb.debian.org/debian stretch main' \
>> /etc/apt/sources.list && \
apt-get update -y && \
apt-get install gcc libgpgme-dev -y

ENTRYPOINT ["bash", "-c", "make build BUILD_LOCAL=0"]