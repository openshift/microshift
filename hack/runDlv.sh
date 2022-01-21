#!/bin/bash
DLV=$HOME/go/bin/dlv
[[ -x $DLV ]] || \
  sudo dnf install -y golang &&
  go install github.com/go-delve/delve/cmd/dlv@latest

sudo systemctl kill microshift
sudo systemctl disable --now microshift
sudo firewall-cmd --zone=public --add-port=2345/tcp --permanent
sudo firewall-cmd --reload
sudo $DLV --listen=:2345 --headless --api-version=2 --accept-multiclient exec /usr/bin/microshift -- run
