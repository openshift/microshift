#!/bin/bash
set -e

# input parameters via env variables

DEST_REGISTRY=${DEST_REGISTRY:-"quay.io/microshift"}
COMPONENTS=${COMPONENTS:-"base-image pause cli coredns flannel flannel-cni haproxy-router hostpath-provisioner kube-rbac-proxy service-ca-operator"}
ARCHITECTURES=${ARCHITECTURES:-"amd64 arm64 arm ppc64le riscv64"}
PUSH=${PUSH:-no}
PARALLEL=${PARALLEL:-yes}

GRAY="\e[1;34m"
GREEN="\e[32m"
CLEAR="\e[0m"

function source_repo {
  jq ' .spec.tags[] | select(.name == "'$1'") | .annotations."io.openshift.build.source-location"' < "${IMG_REFS}" | tr -d '" '
}

function source_commit {
  jq ' .spec.tags[] | select(.name == "'$1'") | .annotations."io.openshift.build.commit.id"' < "${IMG_REFS}" | tr -d '" '
}

function source_image {
  jq ' .spec.tags[] | select(.name == "'$1'") | .from.name' < "${IMG_REFS}" | tr -d '" '
}

function build_component {
  local component=$1
  SRC_REPO=$(source_repo "$component")
  SRC_COMMIT=$(source_commit "$component")
  RELEASE_IMG=$(source_image "$component")

  [ -z "$SRC_REPO" ] && SRC_REPO=$(cat components/"$component"/repo || :)
  [ -z "$SRC_COMMIT" ] && SRC_COMMIT=$(cat components/"$component"/commit || :)

  echo ""
  echo -e "${GREEN}building component: $component${CLEAR}"
  echo "  Source Repo:   $SRC_REPO"
  echo "  Source Commit: $SRC_COMMIT"
  echo "  Source Image:  $RELEASE_IMG"

  pushd components/"$component" >/dev/null

    if [ ! -z "${SRC_REPO}" ]; then
      checkout_component "$SRC_REPO" "$SRC_COMMIT"
      build_cross_binaries
    fi

    build_multiarch_image "$component" "$RELEASE_IMG"
  popd

  if [ "${PUSH}" == "yes" ]; then
      echo -e "${GRAY}> pushing multiarch manifest ${MULTIARCH_MANIFEST}${CLEAR}"
      buildah manifest push --all "${MULTIARCH_MANIFEST}" docker://"${MULTIARCH_MANIFEST}"
  fi
}

function checkout_component {
  echo ""
  echo -e "${GRAY}> making sure we have the source code for $1, at commit $2${CLEAR}"
  [ ! -d src ] && git clone "$1" src
  cd src
  git fetch -a
  git stash >/dev/null # just in case we had patches applied in last run
  git clean -f # remove any out-of-tree files (from patches)
  echo git checkout "$2" -B building-side-images
  git checkout "$2" -B building-side-images
  cd ..
}

function build_cross_binaries {
  for ARCH in ${ARCHITECTURES}
  do
    if [ -f Dockerfile."$ARCH" ] || [ -f Dockerfile ] && [ ! -f ImageSource."$ARCH" ] && [ -x ./build_binaries ]; then
      echo ""
      echo -e "${GRAY}> building binaries for architecture ${ARCH} ${CLEAR}"
       ./build_binaries "$ARCH"
    fi
  done
}

function build_multiarch_image {
  COMPONENT=$1
  RELEASE_IMG=$2
  MULTIARCH_MANIFEST="${DEST_REGISTRY}/${COMPONENT}:${RELEASE_BASE_TAG}"

  echo ""
  echo -e "${GRAY}> preparing multiarch manifest ${MULTIARCH_MANIFEST} ${CLEAR}"

  buildah manifest rm "${MULTIARCH_MANIFEST}" 2>/dev/null >/dev/null || :
  buildah manifest create "${MULTIARCH_MANIFEST}"
  if [ -d src ]; then
    cd src
    VERSION=$(git describe --tags)
    cd ..
  fi

  # allow to disable parallelization, helpful for debugging
  if [ "${PARALLEL}" == "yes" ]; then
    echo ""
    echo -e "${GRAY}> preparing ${COMPONENT} images in parallel for: ${ARCHITECTURES}${CLEAR}"
    for ARCH in ${ARCHITECTURES}
    do
      ARCH_IMAGE="${MULTIARCH_MANIFEST}-${ARCH}"
      (
        set -o pipefail
        build_arch_image |& sed "s/^/[${COMPONENT}:${ARCH}] /"
      ) &
    done
    wait
  else
      for ARCH in ${ARCHITECTURES}
      do
        ARCH_IMAGE="${MULTIARCH_MANIFEST}-${ARCH}"
        echo ""
        echo -e "${GRAY}> preparing arch image ${ARCH_IMAGE} ${CLEAR}"
        build_arch_image |& sed "s/^/[${COMPONENT}:${ARCH}] /"
      done
    wait
  fi

  for ARCH in ${ARCHITECTURES}
  do
    ARCH_IMAGE="${MULTIARCH_MANIFEST}-${ARCH}"
    echo -e "${GRAY}> adding ${ARCH} image to ${MULTIARCH_MANIFEST}${CLEAR}"
    buildah manifest add "${MULTIARCH_MANIFEST}" "${ARCH_IMAGE}"
  done

}

function build_arch_image {
  # different methods to build a component for an arch, we can source a pre-exiting image,
    # have an arch specific Dockerfile, ... have a single Image for all, or a single Dockerfile
    # for all.

    if [ -f "ImageSource.${ARCH}" ]; then
      build_using_image "ImageSource.${ARCH}"

    elif [ -f "Dockerfile.${ARCH}" ]; then
      build_using_dockerfile "Dockerfile.${ARCH}"

    elif [ -f "ImageSource" ]; then
      build_using_image "ImageSource"

    elif [ -f "Dockerfile" ]; then
      build_using_dockerfile "Dockerfile"

    else
      echo "I don't know how to build this image"
      exit 1
    fi
}

function build_using_dockerfile {

   BUILD_ARGS="-f $1 -t ${ARCH_IMAGE}"
   BUILD_ARGS="${BUILD_ARGS} --build-arg VERSION=${VERSION} --build-arg TARGETARCH=${ARCH}"
   BUILD_ARGS="${BUILD_ARGS} --build-arg REGISTRY=${DEST_REGISTRY} --build-arg RELEASE_TAG=${RELEASE_BASE_TAG}"


   buildah build-using-dockerfile --override-arch "${ARCH}" $BUILD_ARGS . || \
      if [ "${ARCH}" == arm ]; then # fedora registry uses armhfp instead for arm (arm32 with floating point)
              buildah build-using-dockerfile --override-arch "armhfp" $BUILD_ARGS .
      fi

}

function build_using_image {

  IMG_REF=$(get_image_ref "$1" "${RELEASE_IMG}")
  buildah pull --arch "${ARCH}" "${IMG_REF}"
  buildah tag "${IMG_REF}" "${ARCH_IMAGE}"
}

function get_image_ref {

  IMG=$(cat "$1")
  # check if we must use the one captured from oc adm image-releases
  if [ "${IMG}" == "\$RELEASE_IMAGE_AMD64" ]; then
    echo "$2"
  else
    echo "${IMG}"
  fi
}

# we need qemu static configured on the system during build if not already installed
if [ ! -f /proc/sys/fs/binfmt_misc/qemu-sparc ]; then
    echo Installing qemu-user-static pre-requisite on the system via sudo
    sudo podman run --rm --privileged multiarch/qemu-user-static --reset -p yes
fi

RELEASE_BASE_TAG=$(cat ../../../pkg/release/release.go | grep "var Base =" | cut -d= -f2 | tr -d '" ')
echo RELEASE Base: "${RELEASE_BASE_TAG}"
IMG_REFS=".image-references.${RELEASE_BASE_TAG}"

if [ ! -f "${IMG_REFS}" ]; then
  oc adm release extract "registry.ci.openshift.org/ocp/release:${RELEASE_BASE_TAG}" --file=image-references > ".image-references.${RELEASE_BASE_TAG}"
fi

for component in $COMPONENTS
do
  build_component "$component"
done
