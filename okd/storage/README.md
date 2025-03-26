## examples for using with storage
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
      -v $(pwd)/okd/examples/hostpath:${kustomize_dir}:z \
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
