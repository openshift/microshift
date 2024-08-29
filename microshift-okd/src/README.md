## Build and Run Microshift upstream without subscription/pull-secret

- building the container from src
  > `cd microshift-okd/src && ./build.sh` 
  - this script will:
    1. replace microshift assets images to OKD  upstream images
    1. will build microshift RPMs and repo based on current sources.
    1. will build micrsoshift_okd bootc container based on `centos-bootc:stream9`
    1. apply upstream customization  (see below)


- running the container 
  > `sudo podman run --privileged --rm --name microshift-okd -d microshift-okd`

- connect to the container
  > `sudo podman exec -ti microshift-okd /bin/bash`

## configuration customization
1. storage driver disabled (there is no lvms images upstream)
1. network CNI disabled (requires kernel modules)
    - microshift service is not dependent on ovs

## current state
- microshift service wont start because CNI is disabled.
    - TODO: replace CNI with flannel .see this [PR](https://github.com/openshift/microshift/pull/3853)
    - TODO: create rebase automation from OKD sources

