#!/bin/bash
#set -xe 
SUDO=''
if (( $EUID != 0 )); then
    SUDO='sudo'
fi

V()
{
  echo "### $*"
  $*
}


function cri_cleanup(){
  until sudo crictl rmp --force --all; do sleep 1; done
}

function check_overlay_mount_points(){
  ps -ef | grep -v auto | grep /run/containers/storage/ >/tmp/command_status
  if  grep -q "overlay-containers" /tmp/command_status;
  then 
    OVERLAYPROCESS=$(${SUDO} ps -ef | grep -v auto | grep /run/containers/storage/ | awk '{print $2}')
    ps -ef | grep -v auto | grep /run/containers/storage/ >/tmp/command_status
    if  grep -q "overlay-containers" /tmp/command_status;
    then 
      for OVERLAYPS in $OVERLAYPROCESS
      do
        ${SUDO} kill -15 $OVERLAYPS
      done
    fi 
    remove_all_overlay_mount_points
  fi 
  rm /tmp/command_status
}

function remove_all_overlay_mount_points(){
  if [[ ! -z $(${SUDO} mount | grep overlay ) ]];
  then 
    V ${SUDO} mount | grep overlay | awk '{print $3}' | xargs ${SUDO} umount
  fi 
}

function remove_opened_ports(){
   V ${SUDO} firewall-cmd --zone=public --permanent --remove-port=6443/tcp
   V ${SUDO} firewall-cmd --zone=public --permanent --remove-port=30000-32767/tcp
   V ${SUDO} firewall-cmd --zone=public --permanent --remove-port=2379-2380/tcp
   V ${SUDO} firewall-cmd --zone=public --remove-masquerade --permanent
   V ${SUDO} firewall-cmd --zone=public --remove-port=10250/tcp --permanent
   V ${SUDO} firewall-cmd --zone=public --remove-port=10251/tcp --permanent
   V ${SUDO} firewall-cmd --permanent --zone=trusted --remove-source=10.42.0.0/16
   V ${SUDO} firewall-cmd --reload
}

function remove_all_overlay_kubelet_points(){
  if [[ ! -z $(${SUDO} mount | grep kubelet ) ]];
  then 
    V ${SUDO} mount | grep kubelet | awk '{print $3}' | xargs ${SUDO} umount
  fi 
}

function remove_microshift_kubeconfig(){
  kubectl config delete-cluster microshift
}

V ${SUDO} systemctl stop microshift
V ${SUDO} systemctl disable microshift
V ${SUDO} rm -f /etc/systemd/system/microshift
V ${SUDO} rm -f /usr/lib/systemd/system/microshift
V ${SUDO} systemctl daemon-reload
V ${SUDO} systemctl reset-failed

cri_cleanup
remove_microshift_kubeconfig
check_overlay_mount_points
remove_all_overlay_kubelet_points
remove_opened_ports

V ${SUDO} pkill -9 pause

V ${SUDO} rm -rf /var/lib/microshift
V ${SUDO} rm -rf /var/lib/rook
V ${SUDO} rm -rf /var/lib/etcd
V ${SUDO} rm -rf /var/lib/kubelet

V ${SUDO} mkdir -p /var/lib/kubelet

echo "Removing Selinux Policy"
sudo semodule -r microshift

sudo pkill -9 conmon
sudo pkill -9 pause

sudo rm -rf /var/lib/microshift
V ${SUDO} systemctl stop cri-o
V ${SUDO} systemctl disable cri-o
V ${SUDO} rm -f /etc/systemd/system/microshift
