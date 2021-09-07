#!/bin/sh

sudo crictl rm --all --force
sudo crictl rmi --all --prune

sudo pkill -9 conmon
sudo pkill -9 pause

sudo rm -rf /var/lib/microshift
