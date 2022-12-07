#!/usr/bin/env bash

declare -a commands_to_run=()
function to_run() {
    cmd="$@"
    commands_to_run+=("${cmd}")
}

to_run oc get cm -n kube-public microshift-version -o=jsonpath='{.data}'
to_run microshift version 
to_run microshift version -o yaml 

RESOURCES=(node pod configmap deployment daemonset statefulset svc route)
for resource in ${RESOURCES[*]}; do
    to_run oc get "${resource}" -A
    to_run oc get "${resource}" -A -o yaml 
done

to_run oc get events -A 

for ns in $(kubectl get namespace -o jsonpath='{.items..metadata.name}'); do
    for pod in $(kubectl get pods -n $ns -o name); do
            to_run oc describe -n $ns $pod 
            for container in $(kubectl get -n $ns $pod -o jsonpath='{.spec.containers[*].name}'); do
                to_run oc logs -n $ns $pod $container 
                to_run oc logs --previous=true -n $ns $pod $container 
            done
    done
done

echo -e "\n=== DEBUG INFORMATION ===\n"
echo "Following commands will be executed:"
for cmd in "${commands_to_run[@]}"; do
    echo "    - ${cmd}"
done

for cmd in "${commands_to_run[@]}"; do
    echo -e "\n\n> ${cmd}"
    ${cmd} 2>&1 || true
done
