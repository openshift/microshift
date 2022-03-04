#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
get="${SCRIPT_DIR}/../../../pkg/release/get.sh"

ARCHITECTURES=${ARCHITECTURES:-"arm64 amd64"}
BASE_VERSION=${BASE_VERSION:-$("${get}" base)}
OUTPUT_DIR=${OUTPUT_DIR:-$(pwd)/archive}

TMP_DIR=$(mktemp -d)

mkdir -p "${OUTPUT_DIR}"
chmod a+rwx "${OUTPUT_DIR}"

for arch in $ARCHITECTURES; do
    images=$("${get}" images $arch)
    storage="${TMP_DIR}/${arch}/containers"
    mkdir -p "${storage}"
    echo "Pulling images for architecture ${arch} ==================="
    for image in $images; do
        echo pulling $image @$arch
        # some imported images are armhfp instead of arm
        podman pull --arch $arch --root "${storage}" "${image}"
        if [ $? -ne 0 ]; then
            if [ "${arch}" == "arm" ];  then
                echo "Fallback arm -> armhfp"
                podman pull --arch armhfp --root "${TMP_DIR}/${arch}" "${image}" || exit 1
            else
                echo "Couldn't pull image ${image} for ${arch}"
                exit 1
            fi
        fi
    done

    echo ""
    echo "Packing tarball for architecture ${arch} =================="
    pushd ${storage}
    output_file="${OUTPUT_DIR}/microshift-containers-${BASE_VERSION}-${arch}.tar.bz2"
    echo " >  ${output_file}"
    tar cfj "${OUTPUT_DIR}/microshift-containers-${BASE_VERSION}-${arch}.tar.bz2" .
    chmod a+rw "${OUTPUT_DIR}/microshift-containers-${BASE_VERSION}-${arch}.tar.bz2"
    popd
    rm -rf ${storage}
done
