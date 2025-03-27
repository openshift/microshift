# Alternative MicroShift etcd Backend

Replacing the default `etcd` backend in MicroShift may be beneficial to allow
another (potentially remote) database usage instead of the default local `etcd`.

MicroShift implements the default `etcd` backend as a separate `microshift-etcd`
executable, which allows for its rather straightforward replacement with a selected
alternative.

One way to replace the default `etcd` database is by using [Kine](https://github.com/k3s-io/kine),
which implements an `etcdshim` that translates `etcd` API to various backends,
like SQLite, Postgres, etc.

## Prerequisites

Download and build Kine using the following commands.

```bash
VER=v0.13.11

git clone https://github.com/k3s-io/kine.git -b "${VER}" ~/kine
cd ~/kine
make build
```

Follow the instructions in [README.md](../README.md) to build the `microshift-okd`
container image of your choice.

## SQLite

Run the following commands to start MicroShift OKD in a bootc container, replacing
the default `microshift-etcd` executable with a script that runs Kine with the
SQLine backend.

```bash
# Copy the Kine executable to be accessible to the podman command
mkdir -p _output/bin
cp ~/kine/bin/kine _output/bin

sudo podman run --privileged --rm --name microshift-okd -d \
    -v ./okd/src/etcd/microshift-etcd-sqlite.sh:/usr/bin/microshift-etcd:ro,Z \
    -v ./_output/bin/kine:/usr/bin/microshift-etcd-kine:ro,Z \
    microshift-okd
```

Run the following command to verify that the Kine SQLine backend is used by MicroShift.

```bash
sudo podman exec -ti microshift-okd /bin/bash <<EOF
ls -l /var/lib/microshift/kine
exit
EOF
```

Run the following command to verify that MicroShift is up and running successfully.

```bash
sudo podman exec -ti microshift-okd /bin/bash <<EOF
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
oc get nodes
oc get pods -A
exit
EOF
```
