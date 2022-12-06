#!/usr/bin/env bash

set -x

echo -e "\n=== DEBUG INFORMATION ===\n\n"

# wrapped in echo so we get a newline
echo "$(oc get cm -n kube-public microshift-version -o=jsonpath='{.data}')"
microshift version 
microshift version -o yaml 

RESOURCES=(node pod configmap deployment daemonset statefulset svc route)
for resource in ${RESOURCES[*]}; do
    oc get "${resource}" -A
    oc get "${resource}" -A -o yaml 
done

oc get events -A 

for ns in $(kubectl get namespace -o jsonpath='{.items..metadata.name}'); do
    for pod in $(kubectl get pods -n $ns -o name); do
            oc describe -n $ns $pod 
            for container in $(kubectl get -n $ns $pod -o jsonpath='{.spec.containers[*].name}'); do
                oc logs -n $ns $pod $container 
                oc logs --previous=true -n $ns $pod $container 
            done
    done
done
