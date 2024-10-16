## Build and Run Microshift upstream without subscription/pull-secret

- building the container with podman multistage build :
  ```bash
  git clone https://github.com/openshift/microshift.git ~/microshift
  ```
  To use OVN-K as CNI
  ```bash
  cd ~/microshift && sudo podman build -f okd/src/microshift-okd-multi-build.Containerfile . -t microshift-okd
  ```
  To use flannel as CNI
  ```bash
  cd ~/microshift && sudo podman build --env WITH_FLANNEL=1 -f okd/src/microshift-okd-multi-build.Containerfile . -t microshift-okd
  ```
  - build runnable container based on current source:
    1. replace microshift assets images to OKD  upstream images
    1. will build microshift RPMs and repo based on current sources.
    1. will build micrsoshift_okd bootc container based on `centos-bootc:stream9`
    1. apply upstream customization  (see below)

- running the container with ovn-kubernetes 
  - make sure to load the openvswitch kernel module  :
    > `sudo modprobe openvswitch`

  - run the container :
    > `sudo podman run --privileged --rm --name microshift-okd -d microshift-okd`

- connect to the container
   > `sudo podman exec -ti microshift-okd /bin/bash`

- verify everything is working:
  ```bash
    export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
    > oc get nodes  
    NAME           STATUS   ROLES                         AGE     VERSION
    d2877aa41787   Ready    control-plane,master,worker   7m39s   v1.30.3
    
    > oc get pods
    NAMESPACE                  NAME                                       READY   STATUS    RESTARTS        AGE
    kube-system                csi-snapshot-controller-7d6c78bc58-5p7tb   1/1     Running   0               8m52s
    kube-system                csi-snapshot-webhook-5598db6db4-rmrpx      1/1     Running   0               8m54s
    openshift-dns              dns-default-2q89q                          2/2     Running   0               7m34s
    openshift-dns              node-resolver-k2c5h                        1/1     Running   0               8m54s
    openshift-ingress          router-default-db4b598b9-x8lvb             1/1     Running   0               8m52s
    openshift-ovn-kubernetes   ovnkube-master-c75c7                       4/4     Running   1 (7m36s ago)   8m54s
    openshift-ovn-kubernetes   ovnkube-node-jfx86                         1/1     Running   0               8m54s
    openshift-service-ca       service-ca-68d58669f8-rns2p                1/1     Running   0               8m51s


  ```

## configuration customization
1. storage driver disabled (there is no lvms images upstream) - will be added in the stage of the project.

## current state
- storage driver is disabled , will be added in the stage of the project.
- TODO: create rebase automation from OKD sources

## known Issues
- when running `podman build` without sudo
  ```
  make: *** [/src/vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk:16: build] Error 1
  Error: building at STEP "RUN make build": while running runtime: exit status 2
  ```
  
