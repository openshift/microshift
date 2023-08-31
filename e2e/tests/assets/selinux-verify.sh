#!/bin/bash

PS4='+ $(date "+%T.%N")\011 '

AUDIT_FILE_LOCATION=/tmp/audit-results.txt
FILE_CAT_LIST=(/var/lib/microshift/version /etc/microshift/config.yaml.default /var/lib/microshift-backups/health.json)

function validate_audit_logs() {
    echo "Validating Audit Logs"

    # If audit logs show a deniel for microshift, it means SELinux is blocking something and we need
    # to resolve it.
    if validation_results=$(sudo ausearch -m avc | grep microshift 2>&1);
    then
        echo "SELinux denials found for MicroShift" >> "${AUDIT_FILE_LOCATION}"
        echo "${validation_results}" >> "${AUDIT_FILE_LOCATION}"
    else
        echo " - No denials found for MicroShift"
    fi
}

function validate_fcontext() {
    echo "Validating File Contexts"
    expected_context=$(cat << EOF
/etc/kubernetes(/.*)?
/etc/microshift(/.*)?
/exports(/.*)?
/usr/bin/microshift
/usr/bin/microshift-etcd
/usr/local/bin/microshift
/usr/local/bin/microshift-etcd
/usr/local/s?bin/hyperkube.*
/usr/local/s?bin/kubelet.*
/usr/s?bin/hyperkube.*
/usr/s?bin/kubelet.*
/var/lib/buildkit(/.*)?
/var/lib/cni(/.*)?
/var/lib/containerd(/.*)?
/var/lib/containers(/.*)?
/var/lib/docker(/.*)?
/var/lib/docker-latest(/.*)?
/var/lib/kubelet(/.*)?
/var/lib/lxc(/.*)?
/var/lib/lxd(/.*)?
/var/lib/microshift(/.*)?
/var/lib/microshift-backups(/.*)?
/var/lib/ocid(/.*)?
/var/lib/registry(/.*)?
EOF
)
    existing=$(sudo semanage fcontext -l | grep -E  "(kubernetes_file_t|container_var_lib_t|kubelet_exec_t|container_t)" | awk '{print $1 }')

    if validation_results=$(diff <(echo "${expected_context}") <(echo "${existing}") 2>&1);
    then
        echo " - No diff between expected and existing contexts"
    else
        echo "SELinux context diff found for MicroShift" >> "${AUDIT_FILE_LOCATION}"
        echo "${validation_results}" >> "${AUDIT_FILE_LOCATION}"
    fi
}


function validate_access() {
    echo "Validating Access"

    for i in "${!FILE_CAT_LIST[@]}"; do
        local file_location
        local validation_results
        file_location=${FILE_CAT_LIST[${i}]}

        # If no error presents it self it means an access was granted when it shouldn't have been
        if validation_results=$(sudo runcon -u system_u -r system_r -t container_t cat "${file_location}" 2>&1);
        then
            echo "Failed to Validate (${file_location}) permission should not have been granted" >> "${AUDIT_FILE_LOCATION}"
            echo "${validation_results}" >> "${AUDIT_FILE_LOCATION}"
        else
            echo " - Permission Validated for ${file_location}"
        fi
    done
}


validate_audit_logs
validate_fcontext
validate_access
