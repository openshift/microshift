#!/bin/sh

pkill -9 conmon
mount |grep overlay |awk '{print $3}' |xargs umount
mount |grep kubelet |awk '{print $3}' |xargs umount
pkill -9 pause
rm /var/lib/etcd -rf