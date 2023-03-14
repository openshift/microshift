#!/usr/bin/env bash

declare -a commands_to_run=()
function to_run() {
    cmd="$@"
    commands_to_run+=("${cmd}")
}

to_run oc get cm -n kube-public microshift-version -o=jsonpath='{.data}'
to_run microshift version
to_run microshift version -o yaml
to_run microshift show-config -m effective
to_run oc version
to_run sudo crictl version
to_run uname -a
to_run cat /etc/*-release

RESOURCES=(nodes pods configmaps deployments daemonsets statefulsets services routes replicasets persistentvolumeclaims persistentvolumes storageclasses endpoints endpointslices csidrivers csinodes)
for resource in ${RESOURCES[*]}; do
    to_run oc get "${resource}" -A
done

to_run oc describe nodes
to_run oc get events -A --sort-by=.metadata.creationTimestamp

for resource in ${RESOURCES[*]}; do
    to_run oc get "${resource}" -A -o yaml
done

TO_DESCRIBE=(deployments daemonsets statefulsets replicasets)
for ns in $(kubectl get namespace -o jsonpath='{.items..metadata.name}'); do
    to_run oc get namespace $ns -o yaml

    for resource_type in ${TO_DESCRIBE[*]}; do
        for resource in $(kubectl get $resource_type -n $ns -o name); do
            to_run oc describe -n $ns $resource
        done
    done

    for pod in $(kubectl get pods -n $ns -o name); do
        to_run oc describe -n $ns $pod
        to_run oc get -n $ns $pod -o yaml
        for container in $(kubectl get -n $ns $pod -o jsonpath='{.spec.containers[*].name}'); do
            to_run oc logs -n $ns $pod $container
            to_run oc logs --previous=true -n $ns $pod $container
        done
    done
done

to_run nmcli
to_run ip a
to_run ip route
to_run sudo crictl images --digests
to_run sudo crictl ps
to_run sudo crictl pods
to_run ls -lah /etc/cni/net.d/
to_run find /etc/cni/net.d/ -type f -exec echo {} \; -exec sudo cat {} \; -exec echo \;
to_run dnf list --installed
to_run dnf history
to_run sudo systemctl -a -l
to_run sudo journalctl -xu microshift
to_run sudo journalctl -xu microshift-etcd.scope

echo -e "\n=== DEBUG INFORMATION ===\n"
echo "Following commands will be executed:"
for cmd in "${commands_to_run[@]}"; do
    echo "    - ${cmd}"
done

for cmd in "${commands_to_run[@]}"; do
    echo -e "\n\n > $ ${cmd}"
    ${cmd} 2>&1 || true
done
