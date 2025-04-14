## examples for using with storage

### building and running Microshift with upsteam TopoLVM storage

1. Build microshift with upstream TopoLVM Support from the Microshift repo root
  ```bash
  sudo podman build --env WITH_FLANNEL=1 --env WITH_TOPOLVM=1 -f okd/src/microshift-okd-multi-build.Containerfile . -t microshift-okd
  ```

1. Prepare the LVM backend on the host (example only)
    ```bash
    sudo truncate --size=20G /tmp/lvmdisk
    sudo losetup -f /tmp/lvmdisk
    device_name=$(losetup -j /tmp/lvmdisk | cut -d: -f1)
    sudo vgcreate -f -y myvg1 ${device_name}
    sudo lvcreate -T myvg1/thinpool -L 6G
    ```

1. Run microshift in container and wait for it to be ready
    ```bash
          sudo podman run --privileged --rm --name microshift-okd \
            --volume /dev:/dev:rslave \
            --hostname 127.0.0.1.nip.io \
            -p 5432:5432 -p 6443:6443 \
            -d localhost/microshift-okd
    ```

1. Wait for all the components to come up 
    ```bash
    > sudo podman exec  microshift-okd bash -c "microshift healthcheck --namespace topolvm-system --deployments topolvm-controller "
    ??? I0331 14:38:46.838208    5894 service.go:29] microshift.service is enabled 
    ??? I0331 14:38:46.838235    5894 service.go:31] Waiting 5m0s for microshift.service to be ready
    ??? I0331 14:38:46.839291    5894 service.go:38] microshift.service is ready
    ??? I0331 14:38:46.840014    5894 workloads.go:94] Waiting 5m0s for deployment/topolvm-controller in topolvm-system
    ??? I0331 14:38:46.844984    5894 workloads.go:132] Deployment/topolvm-controller in topolvm-system is ready
    ??? I0331 14:38:46.845003    5894 healthcheck.go:75] Workloads are ready

    > oc get pods -A
    NAMESPACE              NAME                                       READY   STATUS    RESTARTS        AGE
    cert-manager           cert-manager-5f864bbfd-bpd6h               1/1     Running   0               4m49s
    cert-manager           cert-manager-cainjector-589dc747b5-cfwjf   1/1     Running   0               4m49s
    cert-manager           cert-manager-webhook-5987c7ff58-mzq6l      1/1     Running   0               4m49s
    kube-flannel           kube-flannel-ds-6nvq6                      1/1     Running   0               4m12s
    kube-proxy             kube-proxy-zlvb2                           1/1     Running   0               4m12s
    kube-system            csi-snapshot-controller-75d84cb97c-nkfsz   1/1     Running   0               4m50s
    openshift-dns          dns-default-dbjh4                          2/2     Running   0               4m1s
    openshift-dns          node-resolver-mt8m7                        1/1     Running   0               4m12s
    openshift-ingress      router-default-59cbb858cc-6mzbx            1/1     Running   0               4m49s
    openshift-service-ca   service-ca-df6759f9d-24d2n                 1/1     Running   0               4m49s
    topolvm-system         topolvm-controller-9cd8649c9-5tcln         5/5     Running   0               4m49s
    topolvm-system         topolvm-lvmd-0-2bmjq                       1/1     Running   0               4m1s
    topolvm-system         topolvm-node-lwxz5                         3/3     Running   1 (3m36s ago)   4m1s
    ```


###  run Microshift with hostpath volume
A hostPath volume mounts a file or directory from the host node’s file system into your pod. 
Most pods do not need a hostPath volume, but it does offer a quick option for testing should an application require it (some requires persistent storage - DBs).
1. Build the microshift-okd from [src](../src/README.md)
1. Create directory for persistent data ($data_dir) `sudo mkdir /opt/pdata`
1. Run microshift-okd inside a bootc container
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

1. Wait for all the components to come up 
    ```bash
      sudo podman exec  microshift-okd bash -c "microshift healthcheck --namespace postgresql --deployments postgresql"
    ```

### load some example data into the psql

1. Copy kubeconfig from the container
    ```bash
      sudo podman cp microshift-okd:/var/lib/microshift/resources/kubeadmin/127.0.0.1.nip.io/kubeconfig .
      sudo chown ${USERNAME}:${USERNAME} kubeconfig
    ```

1. Create/verify volume by interacting with psql

    ```bash
      export KUBECONFIG=./kubeconfig
        
      psql_pod=$(oc get pods  | grep postg | cut -f1 -d" ")

      oc exec ${psql_pod} -- psql --user gitlab gitlabhq_production -c "create table test (id SERIAL NOT NULL,name varchar);"
      oc exec ${psql_pod} -- psql --user gitlab gitlabhq_production -c "insert into public.test values(0,'some-data');"
      oc exec ${psql_pod} -- psql --user gitlab gitlabhq_production -c "select * from public.test;"
    ```

1. Stop the container `sudo podman stop microshift-okd`
1. Start a new container instance and verify data is still exists.

