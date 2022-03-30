#!/usr/bin/env bash
set -eu

# print_usage prints help
function print_usage() {
    >&2 echo "Usage: $0 [options]
$0 creates an RPM from from user specified container image(s). This enables a user to prepackage
container images in an immutable ostree OS environment where running scripts and installers may
not be an option. The ability to pre-stage images helps facilitate air-gapped installation
scenario where the node does not have connectivity to a container registry or has security policies
that do not allow for any changes to the host OS.

The following parameters are all required:
 --image-list    -i FILE  file path and name of a file containing user images, one image per line in the file
 --container-dir -c DIR   container storage directory path where the RPM will place the user image
 --rpmbuild-dir  -r DIR   path to the rpm build directory (will be created if it does not exist)

Example usage:
sudo ./make-rpm.sh \\
  --image-list /<path>/images.txt \\
  --container-dir /var/lib/containers/storage/overlay-images \\
  --rpmbuild-dir /home/<user>/rpmbuild
"
}

# file_exists checks to see if a file exists
file_exists() {
    local f="$1"
    stat "$f" &>/dev/null
}

echo 'Checking Dependencies:'
for CMD in rpmbuild crio podman; do
  printf '%-10s' "${CMD}"
  if hash "${CMD}" 2>/dev/null; then
    echo found
  else
    echo missing
   CMD_MISSING=true
  fi
done

if [[ "${CMD_MISSING-}" ]]; then
  echo 'error: please install the missing dependencies and re-run the app'
  exit 1
fi

# parse user specified arguments
RPMBUILD_DIR=
IMAGES=
IMG_DIR=
while true; do
	case "${1:-}" in
		-i | --image-list )
		  IMAGES="$2"; shift 2 ;;
		-c | --container-dir )
			IMG_DIR="$2"; shift 2 ;;
		-r | --rpmbuild-dir )
			RPMBUILD_DIR="$2"; shift 2 ;;
    *) break ;;
	esac
done

# Verify that all required options were specified.
if [[ -z "${IMAGES}" ]]; then print_usage; echo '--image-list option is required.'; exit 1; fi
if [[ -z "${IMG_DIR}" ]]; then print_usage; echo '--container-dir option is required.'; exit 1; fi
if [[ -z "${RPMBUILD_DIR}" ]]; then print_usage; echo '--rpmbuild-dir option is required.'; exit 1; fi

# if the file that would contain user images doesn't exist,
# exit to ensure user is passing a file and not the image name
if (! file_exists "${IMAGES}"); then
    print_usage; echo "error: required file containing a list of images was not found at ${IMAGES}"; exit 1
fi

declare -a ARRAY

# link filedescriptor 10 with stdin (standard input)
exec 10<&0

# stdin replaced with a file supplied as a first argument
exec < "$IMAGES"
count=0

# read user images into ARRAY
while read LINE; do
   ARRAY[$count]=$LINE
   ((count+1))
done

# restore stdin from file descriptor 10 then close filedescriptor 10
exec 0<&10 10<&-

# generate microshift-app-images.spec
touch ./microshift-app-images.spec
cat >./microshift-app-images.spec <<EOF
%define __arch_install_post %{nil}
%global imgDir $IMG_DIR
%global imageStore %{buildroot}%{imgDir}
%global IMAGES ${ARRAY[@]}

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

declare -a ListOfImages=(%{IMAGES})


for val in \${ListOfImages[@]}; do
   sudo podman pull --root %{imageStore} \$val
done
sudo chmod -R a+rx  %{imageStore}


%post
# only on install (1), not on upgrades (2)
if [ \$1 -eq 1 ]; then
   sed -i '/^additionalimagestores =*/a "%{imgDir}",' /etc/containers/storage.conf
   # if crio was already started, restart it so it reads from the new imagestore
   systemctl is-active --quiet crio && systemctl restart --quiet crio || :
fi

%files
%{imgDir}/*

%postun
# only on uninstall (0), not on upgrades(1)
if [ \$1 = 0 ]; then
  sed -i  '\\:"%{imgDir}":d' /etc/containers/storage.conf
  systemctl is-active --quiet crio && systemctl restart --quiet crio || :
fi

EOF

# if the target RPM build directory or directory structure doesn't exist, create it, exit if creation fails
if (! file_exists "${RPMBUILD_DIR}"/SPECS); then
    echo "RPM build directory ${RPMBUILD_DIR} does not exist, attempting to create it"
    mkdir -p "${RPMBUILD_DIR}"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
fi

# copy the rpm spec to the rpmbuild directory
cp ./microshift-app-images.spec "${RPMBUILD_DIR}"/SPECS/microshift-app-images.spec

# build the image with the specified spec
rpmbuild -bb --define "_topdir ${RPMBUILD_DIR}" $RPMBUILD_DIR/SPECS/microshift-app-images.spec