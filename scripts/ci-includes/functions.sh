#!/usr/bin/env bash

#
# Note that CI-specific functions have 'ci_' name prefix.
# The rest should be generic functionality.
#
function ci_script_prologue() {
    IP_ADDRESS="$(cat "${SHARED_DIR}/public_address")"
    export IP_ADDRESS
    HOST_USER="$(cat "${SHARED_DIR}/ssh_user")"
    export HOST_USER
    CACHE_REGION="$(cat "${SHARED_DIR}/cache_region")"
    export CACHE_REGION
    INSTANCE_PREFIX="${HOST_USER}@${IP_ADDRESS}"
    export INSTANCE_PREFIX

    echo "Using Host $IP_ADDRESS"

    mkdir -p "${HOME}/.ssh"
    cat >"${HOME}/.ssh/config" <<EOF
Host ${IP_ADDRESS}
User ${HOST_USER}
IdentityFile ${CLUSTER_PROFILE_DIR}/ssh-privatekey
StrictHostKeyChecking accept-new
ServerAliveInterval 30
ServerAliveCountMax 1200
EOF
    chmod 0600 "${HOME}/.ssh/config"
}

function ci_copy_secrets() {
    local -r cache_region=$1

    # Set up the SSH keys at the expected location
    if [ -e /tmp/ssh-publickey ] && [ -e /tmp/ssh-privatekey ] ; then
        cp /tmp/ssh-publickey ~/.ssh/id_rsa.pub
        cp /tmp/ssh-privatekey ~/.ssh/id_rsa
        chmod 0400 ~/.ssh/id_rsa*
    fi

    # Set up the pull secret at the expected location
    if [ -e /tmp/pull-secret ] ; then
        export PULL_SECRET="${HOME}/.pull-secret.json"
        cp /tmp/pull-secret "${PULL_SECRET}"
    fi

    # Set up the AWS CLI keys at the expected location for accessing the cached data.
    # Also, set the environment variables for using the profile and bucket.
    if [ -e /tmp/aws_access_key_id ] && [ -e /tmp/aws_secret_access_key ] ; then
        echo "Setting up AWS CLI configuration for the 'microshift-ci' profile"
        mkdir -m 0700 "${HOME}/.aws/"

        # Profile configuration
        cat >>"${HOME}/.aws/config" <<EOF

[microshift-ci]
region = ${cache_region}
output = json
EOF

        # Profile credentials
        cat >>"${HOME}/.aws/credentials" <<EOF

[microshift-ci]
aws_access_key_id = \$(cat /tmp/aws_access_key_id)
aws_secret_access_key = \$(cat /tmp/aws_secret_access_key)
EOF

        # Permissions and environment settings
        chmod -R go-rwx "${HOME}/.aws/"
        export AWS_PROFILE=microshift-ci
        export AWS_BUCKET_NAME="microshift-build-cache-${cache_region}"
    fi
}

function ci_subscription_register() {
    # Check if the system is already registered
    if sudo subscription-manager status >&/dev/null; then
        return
    fi

    if [ ! -e /tmp/subscription-manager-org ] || [ ! -e /tmp/subscription-manager-act-key ] ; then
        echo "ERROR: The subscription files do not exist in /tmp directory"
        return 1
    fi
    sudo subscription-manager register \
        --org="$(cat /tmp/subscription-manager-org)" \
        --activationkey="$(cat /tmp/subscription-manager-act-key)"
}

function trap_subprocesses_on_term() {
    # Call wait regardless of the outcome of the kill command, in case some of
    # the subprocesses are finished by the time we try to kill them.
    trap 'PIDS=$(jobs -p); if test -n "${PIDS}"; then kill ${PIDS} || true && wait; fi' TERM
}

#
# Unconditionally enable tracing after loading the functions
#
export PS4='+ $(date "+%T.%N") \011'
