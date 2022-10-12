#!/bin/bash

pushd /tmp
git clone https://github.com/openshift/oadp-operator.git
cd oadp-operator
sed -i 's,r.ReconcileVeleroServiceMonitor,// &,g' controllers/dpa_controller.go
sudo podman build -t oadp-operator:pmtk . --platform=linux/amd64

sed 's,newName.*$,newName: localhost/oadp-operator:pmtk,g' config/manager//kustomization.yaml
oc kustomize ./config/default | oc apply -f -
popd

oc create secret generic cloud-credentials --namespace openshift-adp --from-file cloud=./credentials-velero

oc delete validatingwebhoo
kconfigurations.admissionregistration.k8s.io  snapshot.storage.k8s.io

oc create ns test

oc create -f ./10-velero-service-acc.yaml
oc create -f ./20-volume-snapshot-class.yaml
oc create -f ./30-dpa.yaml
oc create -f ./50-pod.yaml
#oc create -f ./60-backup.yaml
