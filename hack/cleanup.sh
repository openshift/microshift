#!/bin/bash
#set -xe 
WAIT_FOR_CONTAINER="1"
SUDO=''
if (( $EUID != 0 )); then
    SUDO='sudo'
fi

V()
{
  echo "### $*"
  $*
}


function check_for_running_containers(){
  if [[ ! -z $(${SUDO} crictl ps  | grep -v CONTAINER) ]];
  then 
    CONTAINERS=$(${SUDO} crictl ps | awk '{if(NR>1) print $1}')
    for CONTAINER in $CONTAINERS
    do
      V ${SUDO} crictl stop $CONTAINER
      sleep ${WAIT_FOR_CONTAINER}s
    done
  fi 
}

function remove_all_stopped_containers(){
  if [[ ! -z $(${SUDO} crictl ps -q -a | grep -v CONTAINER) ]];
  then 
    CONTAINERS=$(${SUDO} crictl ps -a | awk '{if(NR>1) print $1}')
    for CONTAINER in $CONTAINERS
    do
      V ${SUDO} crictl rm $CONTAINER 
    done
  fi 
}

function check_for_running_pods(){
  if [[ ! -z $(${SUDO} crictl pods | grep -v "POD ID") ]];
  then 
    PODS=$(${SUDO} crictl pods | awk '{if(NR>1) print $1}')
    for POD in $PODS
    do
      V ${SUDO} crictl stopp $POD 
    done
  fi 
}

function remove_all_pods(){
  if [[ ! -z $(${SUDO} crictl pods | grep -v "POD ID") ]];
  then 
    PODS=$(${SUDO} crictl pods | awk '{if(NR>1) print $1}')
    for POD in $PODS
    do
      V ${SUDO} crictl rmp $POD 
    done
  fi 
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

function remove_all_overlay_kubelet_points(){
  if [[ ! -z $(${SUDO} mount | grep kubelet ) ]];
  then 
    V ${SUDO} mount | grep kubelet | awk '{print $3}' | xargs ${SUDO} umount
  fi 
}

V ${SUDO} systemctl stop microshift
V ${SUDO} systemctl disable microshift
V ${SUDO} rm -rf /etc/systemd/system/microshift
V ${SUDO} rm -rf /usr/lib/systemd/system/microshift
V ${SUDO} systemctl daemon-reload
V ${SUDO} systemctl reset-failed

check_for_running_containers
remove_all_stopped_containers
check_for_running_pods
remove_all_pods

check_overlay_mount_points
remove_all_overlay_kubelet_points

V ${SUDO} pkill -9 pause

V ${SUDO} rm -rf /var/lib/microshift
V ${SUDO} rm -rf /var/lib/rook
V ${SUDO} rm -rf /var/lib/etcd
V ${SUDO} rm -rf /var/lib/kubelet
V ${SUDO} rm -rf $HOME/.kube

V ${SUDO} mkdir -p /var/lib/kubelet
V ${SUDO} chcon -R -t container_file_t /var/lib/kubelet/
