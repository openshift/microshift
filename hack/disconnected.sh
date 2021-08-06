#!/bin/bash
DESTINATION_PATH=$1
if [[ -z $DESTINATION_PATH ]]; then
   echo "A path must be specified before running this script. Example (sh hack/disconnected.sh /tmp/)"
   exit 1
fi

# Set temporary location
PODMAN_ROOT="/tmp/storage"
rm -rf ${PODMAN_ROOT} || true
mkdir -p ${PODMAN_ROOT}

echo "Discovering images"
CONSTANTS=`grep "Image" ./pkg/constant/constant.go  | uniq | awk -F= '{print $2}' | tr -d '"'  | tr -d ' '`
BINDATA=`grep "image:" ./pkg/assets/apps/bindata.go  | uniq | grep -v { | awk -F: '{print $2":" $3}' | tr -d ' '`

for i in ${CONSTSANTS} ${BINDATA}; do
   podman --root ${PODMAN_ROOT} --runroot  ${PODMAN_ROOT}  pull  ${i}
done

sudo tar czf  ${DESTINATION_PATH}/microshift-images-amd64.tgz -C /tmp/ storage
