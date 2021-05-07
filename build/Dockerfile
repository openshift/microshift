FROM docker.io/golang:1.16
LABEL name=ushift-build

ENV GOPATH=/opt/app-root GOCACHE=/mnt/cache

WORKDIR $GOPATH/src/github.com/microshift

VOLUME go_cache:/mnt/cache

RUN \
printf "%s\n" \
'deb http://deb.debian.org/debian stretch main' \
'deb-src htt://deb.debian.org/debian stretch main' \
>> /etc/apt/sources.list && \
apt-get update -y && \
apt-get install gcc libgpgme-dev -y

ENTRYPOINT ["bash", "-c", "make build BUILD_LOCAL=0"]