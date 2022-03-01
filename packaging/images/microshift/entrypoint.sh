#!/bin/sh

mkdir -p /etc/crio/crio.conf.d
cp /root/crio.conf.d/microshift.conf /etc/crio/crio.conf.d/microshift.conf

# switch to microshift process
exec /usr/bin/microshift run
