## examples for using with storage

### run Microshift with upsteam topolvm storage

1. prepare the LVM backend on the host (example only)
    ```bash
    truncate --size=20G /tmp/lvmdisk
    losetup -f /tmp/lvmdisk
    device_name=$(losetup -j /tmp/lvmdisk | cut -d: -f1)
    vgcreate -f -y myvg1 /dev/${device_name}
    lvcreate -T myvg1/thinpool -L 6G
    ```


1. run microshift in OKD and wait for it to be ready
    ```bash
        export kustomize_dir=/usr/lib/microshift/manifests.d/001-topolvm
        sudo podman run --privileged --rm --name microshift-okd \
          -v $(pwd)/okd/storage/topolvm:${kustomize_dir}:z \
          --hostname 127.0.0.1.nip.io \
          -p 5432:5432 -p 6443:6443 \
          -d localhost/microshift-okd-4.19
    ```

1. wait for all the components to come up 
    ```bash
    > sudo podman exec  microshift-okd bash -c "microshift healthcheck --namespace topolvm-system --deployments topolvm-controller "
    ??? I0331 14:38:46.838208    5894 service.go:29] microshift.service is enabled
    ??? I0331 14:38:46.838235    5894 service.go:31] Waiting 5m0s for microshift.service to be ready
    ??? I0331 14:38:46.839291    5894 service.go:38] microshift.service is ready
    ??? I0331 14:38:46.840014    5894 workloads.go:94] Waiting 5m0s for deployment/topolvm-controller in topolvm-system
    ??? I0331 14:38:46.844984    5894 workloads.go:132] Deployment/topolvm-controller in topolvm-system is ready
    ??? I0331 14:38:46.845003    5894 healthcheck.go:75] Workloads are ready
    ```


###  run Microshift with hostpath volume
A hostPath volume mounts a file or directory from the host node’s file system into your pod. 
Most pods do not need a hostPath volume, but it does offer a quick option for testing should an application require it (some requires persistent storage - DBs).
1. build the microshift-okd from [src](../src/README.md)

1. create directory for persistent data ($data_dir) `sudo mkdir /opt/pdata`
1. run microshift-okd inside a bootc container
    ```bash
    export kustomize_dir=/usr/lib/microshift/manifests.d/001-static-volumes
    export data_dir="/opt/pdata"

    sudo podman run --privileged --rm --name microshift-okd \
      -v ${data_dir}:/var/lib/pvdata:z \
      -v $(pwd)/okd/storage/hostpath:${kustomize_dir}:z \
      --hostname 127.0.0.1.nip.io \
      -p 5432:5432 -p 6443:6443 \
      -d localhost/microshift-okd 

    ```

1. wait for all the components to come up 
    ```bash
    sudo podman exec  microshift-okd bash -c "microshift healthcheck --namespace postgresql --deployments postgresql"
    ```

### load some example data into the psql

1. copy kubeconfig from the container
    ```bash
        sudo podman cp microshift-okd:/var/lib/microshift/resources/kubeadmin/127.0.0.1.nip.io/kubeconfig .
        sudo chown ${USERNAME}:${USERNAME} kubeconfig
    ```

1. create/verify volume by interacting with psql

    ```bash
      export KUBECONFIG=./kubeconfig
        
      psql_pod=$(oc get pods  | grep postg | cut -f1 -d" ")

      oc exec ${psql_pod} -- psql --user gitlab gitlabhq_production -c "create table test (id SERIAL NOT NULL,name varchar);"
      oc exec ${psql_pod} -- psql --user gitlab gitlabhq_production -c "insert into public.test values(0,'some-data');"
      oc exec ${psql_pod} -- psql --user gitlab gitlabhq_production -c "select * from public.test;"
    ```

1. stop the container `sudo podman stop microshift-okd`
1. start a new container instance and verify data is still exists.

