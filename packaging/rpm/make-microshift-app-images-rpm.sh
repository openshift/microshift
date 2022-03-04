#!/usr/bin/env bash

# First arg: file path containing user images per line
# Second arg: container storage dir path
# Third arg:  RPMBUILD_DIR

RPMBUILD_DIR=$3
_img_dir_=$2

declare -a ARRAY

#link filedescriptor 10 with stdin (standard input)
exec 10<&0

#stdin replaced with a file supplied as a first argument
exec < $1
let count=0

#read user images into ARRAY
while read LINE; do
   ARRAY[$count]=$LINE
   ((count++))
done

#restore stdin from file descriptor 10 then close filedescriptor 10
exec 0<&10 10<&-

#Generate microshift-app-images.spec 
touch ./microshift-app-images.spec
cat >./microshift-app-images.spec <<EOF
%global _img_dir $_img_dir_
%global imageStore %{buildroot}%{_img_dir}
%global _IMAGES_ ${ARRAY[@]}

Name: microshift-app-images
Version: 1
Release: 1
Summary: Creates RO container storage for user applications
License: Apache License 2.0
URL: https://github.com/redhat-et/microshift

BuildRequires:  podman
BuildRequires: crio


%description
This rpm creates a RO container storage for user applications, pull the app images and add the path to additional container image stores on target machine.

%prep

if [ -d  %{imageStore} ] 
then
   sudo rm -rf  %{imageStore}
fi

%build

%install

mkdir -p %{imageStore}

declare -a ListOfImages=(%{_IMAGES_})


for val in \${ListOfImages[@]}; do
   sudo podman pull --root %{imageStore} \$val
done
sudo chmod -R a+rx  %{imageStore}


%post
sudo sed -i '/^additionalimagestores =*/a "$_img_dir_",' /etc/containers/storage.conf
# if crio was already started, restart it so it read from new imagestore
systemctl is-active --quiet crio && systemctl restart --quiet crio


%files
%{_img_dir}/*

EOF
cp ./microshift-app-images.spec $RPMBUILD_DIR/SPECS/microshift-app-images.spec

rpmbuild -bb --define "_topdir ${RPMBUILD_DIR}" $RPMBUILD_DIR/SPECS/microshift-app-images.spec