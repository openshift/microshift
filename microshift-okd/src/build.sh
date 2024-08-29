#/bin/bash


# replace microshift assets to Upsteam from OKD
okd_url=quay.io/okd/scos-release
okd_releaseTag=4.17.0-0.okd-scos-2024-08-21-100712

oc adm release info ${okd_url}:${okd_releaseTag} >all_images

for op in $(cat assets/release/release-x86_64.json  | jq -e -r  '.images | keys []') 
do
    image=$(oc adm release info --image-for=${op} ${okd_url}:${okd_releaseTag} || true) 
    if [ -n "${image}" ] ; then
        echo "${op} ${image}"
        jq --arg a "${op}" --arg b "${image}"  '.images[$a] = $b' assets/release/release-x86_64.json >/tmp/release-x86_64.json.tmp


        # delete openssl image from assets - just to verify if we still need it,since it doesnt referenced anywhere
        jq '. | del (.images["openssl"])'  assets/release/release-x86_64.json >/tmp/release-x86_64.json.tmp2

        mv /tmp/release-x86_64.json.tmp2 assets/release/release-x86_64.json
    fi
done


sudo podman stop microshift-okd
rm -rf microshift-okd/src/rpmbuild
rm -rf _output/rpmbuild
make build 
make rpm
createrepo _output/rpmbuild/
cp -rf _output/rpmbuild microshift-okd/src/

# build the container
sudo podman build -f microshift-okd/src/microshift-okd-source.containerfile -t microshift-okd

# run container on the background
sudo podman run --privileged --rm --name microshift-okd -d microshift-okd

# connect the container
sudo podman exec -ti microshift-okd /bin/bash