#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

crd_globs="\
    apiextensions/v1alpha1/zz_generated.crd-manifests/*_crd-compatibility-checker_*.crd*yaml\
    authorization/v1/zz_generated.crd-manifests/*_config-operator_*.crd*yaml\
    config/v1/zz_generated.crd-manifests/*_config-operator_*.crd*yaml\
    machine/v1/zz_generated.crd-manifests/*.crd*yaml\
    operator/v1/zz_generated.crd-manifests//*_config-operator_*.crd*yaml\
    operator/v1alpha1/zz_generated.crd-manifests//*_config-operator_*.crd*yaml\
    quota/v1/zz_generated.crd-manifests/*_config-operator_*.crd*yaml\
    security/v1/zz_generated.crd-manifests/*_config-operator_*.crd*yaml\
    securityinternal/v1/zz_generated.crd-manifests/*_config-operator_*.crd*yaml
    operator/v1/zz_generated.crd-manifests/0000_50_authentication_01_authentications*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_30_openshift-apiserver_01_openshiftapiservers*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_20_kube-apiserver_01_kubeapiservers*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_12_etcd_01_etcds*.crd.yaml
    operator/v1alpha1/zz_generated.crd-manifests/0000_10_etcd_01_etcdbackups*.crd.yaml
    config/v1alpha1/zz_generated.crd-manifests/0000_10_config-operator_01_backups*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_25_kube-scheduler_01_kubeschedulers*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_25_kube-controller-manager_01_kubecontrollermanagers*.crd.yaml
    config/v1/zz_generated.crd-manifests/0000_10_openshift-controller-manager_01_builds*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_50_openshift-controller-manager_02_openshiftcontrollermanagers*.crd.yaml
    machineconfiguration/v1/zz_generated.crd-manifests/*.crd.yaml
    machineconfiguration/v1alpha1/zz_generated.crd-manifests/0000_80_machine-config_01_internalreleaseimages-*.crd.yaml
    machineconfiguration/v1alpha1/zz_generated.crd-manifests/0000_80_machine-config_01_osimagestreams-*.crd.yaml
    operator/v1/zz_generated.crd-manifests/0000_80_machine-config_01_machineconfigurations*.crd.yaml
    config/v1alpha1/zz_generated.crd-manifests/0000_10_config-operator_01_clustermonitoring*.crd.yaml
    config/v1alpha1/zz_generated.crd-manifests/0000_10_config-operator_01_criocredentialproviderconfigs*.crd.yaml
    operator/v1/zz_generated.crd-manifests/*_storage_01_storages*.crd.yaml
    operator/v1/zz_generated.crd-manifests/*_csi-driver_01_clustercsidrivers*.crd.yaml
    insights/v1/zz_generated.crd-manifests/0000_10_insights_01_datagathers*.crd.yaml
    etcd/v1alpha1/zz_generated.crd-manifests/0000_25_etcd_01_pacemakerclusters*.crd.yaml
    operator/v1alpha1/zz_generated.crd-manifests/0000_30_cluster-api_01_clusterapis*.crd.yaml
    "

# To allow the crd_globs to be sourced in the verify script,
# wrap the copy action to prevent it running when sourced.
if [ "$0" = "$BASH_SOURCE" ] ; then
    rm -rf "${SCRIPT_ROOT}/payload-manifests/crds/"*
    for f in ${crd_globs}; do
        cp "$f" "${SCRIPT_ROOT}/payload-manifests/crds/"
    done
fi
