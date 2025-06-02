#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

POSTGRES_IMAGE="docker.io/library/postgres:10.12"
REDIS_IMAGE="docker.io/library/redis:5.0.7"
QUAY_IMAGE="quay.io/microshift/quay:v3.11.7-$(uname -m)"
QUAY_CONFIG_DIR="${MIRROR_REGISTRY_DIR}/config"
QUAY_STORAGE_DIR="${MIRROR_REGISTRY_DIR}/storage"

PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}
QUAY_PULL_SECRET="${QUAY_CONFIG_DIR}/pull_secret.json"

reset_storage_permissions() {
    # Ensure that the current user owns the mirror registry directories and files
    if [ -d "${MIRROR_REGISTRY_DIR}" ] ; then
        sudo chown -R "$(id -gn)" "${MIRROR_REGISTRY_DIR}"
        sudo chgrp -R "$(id -gn)" "${MIRROR_REGISTRY_DIR}"
    fi
}

setup_prereqs() {
    # Install packages if not yet available locally
    if ! rpm -q podman skopeo jq &>/dev/null ; then
        "${SCRIPTDIR}/../../scripts/dnf_retry.sh" "install" "podman skopeo jq"
    fi

    # Create registry repository base directory structure and reset permissions
    # if downloaded from cache
    mkdir -p "${MIRROR_REGISTRY_DIR}"
    mkdir -p "${QUAY_CONFIG_DIR}"
    reset_storage_permissions

    # Create a new pull secret file containing authentication information for both
    # remote (from PULL_SECRET environment) and local registries
    cat > "${QUAY_CONFIG_DIR}/microshift_auth.json" <<EOF
{
    "auths": {
        "${MIRROR_REGISTRY_URL}": {
            "auth": "$(echo -n 'microshift:microshift' | base64)"
        }
    }
}
EOF
    jq -s '.[0] * .[1]' "${PULL_SECRET}" "${QUAY_CONFIG_DIR}/microshift_auth.json" > "${QUAY_PULL_SECRET}"
    chmod 600 "${QUAY_PULL_SECRET}"

    # TLS authentication is disabled in Quay local registry. The mirror-images.sh
    # helper uses skopeo without TLS options and it defaults to https, so we need
    # to configure registries.conf.d for skopeo to try http instead.
    sudo bash -c 'cat > /etc/containers/registries.conf.d/900-microshift-mirror.conf' <<EOF
[[registry]]
    prefix = ""
    location = "${MIRROR_REGISTRY_URL}"
    insecure = true

[[registry]]
    prefix = ""
    location = "quay.io"
[[registry.mirror]]
    location = "${MIRROR_REGISTRY_URL}"
    insecure = true

[[registry]]
    prefix = ""
    location = "registry.redhat.io"
[[registry.mirror]]
    location = "${MIRROR_REGISTRY_URL}"
    insecure = true

[[registry]]
    prefix = ""
    location = "localhost"
[[registry.mirror]]
    location = "${MIRROR_REGISTRY_URL}"
    insecure = true
EOF

# Complete the source registry configuration to use sigstore attachments.
# Note that registry.redhat.io.yaml file already exists, but it is missing the
# sigstore attachment enablement setting.
sudo bash -c 'cat > /etc/containers/registries.d/registry.quay.io.yaml' <<'EOF'
docker:
    quay.io:
        use-sigstore-attachments: true
EOF

if [   -e /etc/containers/registries.d/registry.redhat.io.yaml ] &&
   [ ! -e /etc/containers/registries.d/registry.redhat.io.yaml.orig ]; then
   sudo mv /etc/containers/registries.d/registry.redhat.io.yaml /etc/containers/registries.d/registry.redhat.io.yaml.orig
fi

sudo bash -c 'cat > /etc/containers/registries.d/registry.redhat.io.yaml' <<'EOF'
docker:
    registry.redhat.io:
        use-sigstore-attachments: true
        sigstore: https://registry.redhat.io/containers/sigstore
EOF

# Configure the destination local registry to use sigstore attachments.
# Note: The sigstore staging directory is required because not all registries
# support direct copy of signatures. In this case, the signatures are downloaded
# locally and copied to the destination registry.
local -r quay_base="$(dirname "${MIRROR_REGISTRY_URL}")"
local -r sigstore="${MIRROR_REGISTRY_DIR}/sigstore-staging"

mkdir -p "${sigstore}"
sudo bash -c 'cat > /etc/containers/registries.d/registry.quay.local.yaml' <<EOF
docker:
    ${quay_base}:
        use-sigstore-attachments: true
        lookaside-staging: file://${sigstore}
EOF
}

setup_registry() {
    local -r quay_url="$(hostname):${MIRROR_REGISTRY_PORT}"
    local postgres_ip
    local redis_ip
    local new_db=false

    # No setup is necessary if ALL the containers are already running
    if [ -n "$(sudo podman ps -q --filter "name=microshift-postgres" --filter "status=running")" ] &&
       [ -n "$(sudo podman ps -q --filter "name=microshift-redis"    --filter "status=running")" ] &&
       [ -n "$(sudo podman ps -q --filter "name=microshift-quay"     --filter "status=running")" ] ; then
       echo "All containers are running - skipping mirror registry setup"
       return
    fi

    # Delete running containers if any
    for n in postgres redis quay ; do
        local cn="microshift-${n}"
        echo "Removing '${cn}' container"
        sudo podman rm -f --time 0 "${cn}" || true
    done

    # Pull the registry images in background locally
    for i in "${POSTGRES_IMAGE}" "${REDIS_IMAGE}" "${QUAY_IMAGE}" ; do
        echo "Pulling '${i}' image locally"
        sudo skopeo copy \
            --quiet \
            --retry-times 3 \
            --preserve-digests \
            "docker://${i}" \
            "containers-storage:${i}" &
    done
    # Wait until the image pull is complete
    wait

    # Set up Postgres
    #
    # The following changes are implemented on top of the documented procedure:
    # - The setfacl command is not used because the container is run with the
    #   current user and group permissions.
    # - The container is run with the current user and group permissions to avoid
    #   permission denied issues in the user home directory.
    # - The number of maximum connections to the database is increased from the
    #   default of 100 to avoid 'FATAL: sorry, too many clients already' errors.
    #
    # See https://docs.projectquay.io/deploy_quay.html#poc-configuring-database
    if [ ! -d "${MIRROR_REGISTRY_DIR}/postgres" ] ; then
        mkdir -p "${MIRROR_REGISTRY_DIR}/postgres"
        new_db=true
    fi

    # Note that the container log still shows the default setting of 100. Run
    # the 'echo SHOW max_connections | psql -d quay -U quayuser' query to
    # determine the current setting.
    echo "Running Postgres container"
    sudo podman run -d --rm --name microshift-postgres \
        --user "$(id -u):$(id -g)" \
        -e POSTGRES_USER=quayuser \
        -e POSTGRES_PASSWORD=quaypass \
        -e POSTGRES_DB=quay \
        -e POSTGRESQL_ADMIN_PASSWORD=adminpass \
        -p 5432:5432 \
        -v "${MIRROR_REGISTRY_DIR}/postgres:/var/lib/postgresql/data:Z" \
        "${POSTGRES_IMAGE}" -c max_connections=1024 >/dev/null
    postgres_ip=$(sudo podman inspect -f "{{.NetworkSettings.IPAddress}}" microshift-postgres)

    # Retry the query until the database is available
    for i in $(seq 60) ; do
        sleep 1
        if sudo podman exec -it microshift-postgres \
            /bin/bash -c 'echo "CREATE EXTENSION IF NOT EXISTS pg_trgm" | psql -d quay -U quayuser' >/dev/null ; then
            i=0
            break
        fi
    done
    if [ "${i}" -ne 0 ] ; then
        echo "ERROR: Timed out waiting for Postgres database initialization"
        exit 1
    fi

    # Setup and run Redis
    # See https://docs.projectquay.io/deploy_quay.html#poc-configuring-redis
    echo "Running Redis container"
    sudo podman run -d --rm --name microshift-redis \
        -p 6379:6379 \
        "${REDIS_IMAGE}" \
        --requirepass strongpassword >/dev/null
    redis_ip=$(sudo podman inspect -f "{{.NetworkSettings.IPAddress}}" microshift-redis)

    # Set up Quay
    # See https://docs.projectquay.io/deploy_quay.html#poc-deploying-quay
    #
    # Create the configuration template using the minimal configuration settings.
    # If template is updated, replace hardcoded Postgres, Redis IPs and Quay URL
    # by respective variables.
    # See https://docs.projectquay.io/deploy_quay.html#preparing-configuration-file
    POSTGRES_IP="${postgres_ip}" \
    REDIS_IP="${redis_ip}" \
    QUAY_URL="${quay_url}" \
    envsubst \
        < "${SCRIPTDIR}/../assets/quay/config.yaml.template" \
        > "${QUAY_CONFIG_DIR}/config.yaml"

    # Enable superuser creation using API
    # See https://docs.projectquay.io/deploy_quay.html#configuring-superuser
    cat >> "${QUAY_CONFIG_DIR}/config.yaml" <<EOF
FEATURE_USER_INITIALIZE: true
SUPER_USERS:
    - microshift
EOF
    # Enable public repository creation on push
    # See https://docs.redhat.com/en/documentation/red_hat_quay/3/html-single/configure_red_hat_quay/index#config-fields-misc
    echo "CREATE_PRIVATE_REPO_ON_PUSH: false" >> "${QUAY_CONFIG_DIR}/config.yaml"

    # Enable Quay dual-stack server support if the local host supports IPv6
    local podman_network=""
    if ping -6 -c 1 ::1 &>/dev/null ; then
        # Add the configuration option
        # See https://docs.redhat.com/en/documentation/red_hat_quay/3/html-single/configure_red_hat_quay/index#config-fields-ipv6
        echo "FEATURE_LISTEN_IP_VERSION: dual-stack" >> "${QUAY_CONFIG_DIR}/config.yaml"
        # Enable both IPv4 and IPv6 podman container network for the root user
        # See https://access.redhat.com/solutions/6196301
        if ! sudo podman network exists microshift-ipv6-dual-stack ; then
            sudo podman network create microshift-ipv6-dual-stack --ipv6 >/dev/null
        fi
        podman_network="--network=microshift-ipv6-dual-stack"
    fi

    # The following changes are implemented on top of the documented procedure:
    # - The setfacl command is not used because the container is run with the
    #   current user and group ID mapping.
    # See https://docs.projectquay.io/deploy_quay.html#preparing-local-storage
    if [ ! -d "${QUAY_STORAGE_DIR}" ] ; then
        mkdir -p "${QUAY_STORAGE_DIR}"
    fi

    # Run Quay container
    #
    # The following changes are implemented on top of the documented procedure:
    # - The current user and group ID is mapped to 1001 in the container
    #   so that all files on host are owned by the current user
    # See https://docs.projectquay.io/deploy_quay.html#deploy-quay-registry
    echo "Running Quay container"
    sudo podman run -d --name=microshift-quay \
        --uidmap="0:0:1" --uidmap="1001:$(id -u):1" \
        --gidmap="0:0:1" --gidmap="1001:$(id -g):1" \
        ${podman_network} \
        -p "${MIRROR_REGISTRY_PORT}:8080" \
        -p "[::]:${MIRROR_REGISTRY_PORT}:8080" \
        -v "${QUAY_CONFIG_DIR}:/conf/stack:Z" \
        -v "${QUAY_STORAGE_DIR}:/datastorage:Z" \
        "${QUAY_IMAGE}" >/dev/null

    # Wait until the Quay instance is started
    for i in $(seq 60) ; do
        sleep 1
        if curl -sI --connect-timeout 5 --max-time 5 "${quay_url}" 2>/dev/null | grep -Eq "HTTP.*200 OK" ; then
            i=0
            break
        fi
    done
    if [ "${i}" -ne 0 ] ; then
        echo "ERROR: Timed out waiting for Quay to start"
        exit 1
    fi

    # Create the superuser, verifying the creation was successful
    # See https://docs.projectquay.io/config_quay.html#using-the-api-to-create-first-user
    if ${new_db} ; then
        local response
        response="$(curl -s -X POST -k  "${quay_url}/api/v1/user/initialize" \
            --header 'Content-Type: application/json' \
            --data '{ "username":"microshift", "password":"microshift", "email":"noemail@redhat.com", "access_token":true}')"
        jq -e 'if .access_token then true else error(.message) end' <<< "${response}" >/dev/null
    fi
}

finalize_registry() {
    reset_storage_permissions
    # Delete the combined pull secret file
    rm -f "${QUAY_PULL_SECRET}"
}

mirror_images() {
    local -r ifile=$1
    local -r ffile=$(mktemp /tmp/from-list.XXXXXXXX)
    local -r ofile=$(mktemp /tmp/container-list.XXXXXXXX)

    # Add non-localhost-FROM images to the mirrored list
    find "${SCRIPTDIR}/../image-blueprints-bootc" -name '*.containerfile' | while read -r cf ; do
        local src_img
        src_img=$(awk '/^FROM / && $2 !~ /^localhost\// {print $2}' "${cf}")

        [ -z "${src_img}" ] && continue
        echo "${src_img}" >> "${ffile}"
    done

    sort -u "${ifile}" "${ffile}" > "${ofile}"
    "${ROOTDIR}/scripts/mirror-images.sh" --mirror "${QUAY_PULL_SECRET}" "${ofile}" "${MIRROR_REGISTRY_URL}"
    rm -f "${ofile}" "${ffile}"
}

usage() {
    echo ""
    echo "Usage: ${0} [-cf FILE]"
    echo "   -cf FILE    File containing the container image references to mirror."
    echo "               Defaults to '${CONTAINER_LIST}', skipped if does not exist."
    echo ""
    echo "The registry data is stored at '${MIRROR_REGISTRY_DIR}' on the host."
    exit 1
}

#
# Main
#
image_list_file="${CONTAINER_LIST}"

while [ $# -gt 0 ]; do
    case $1 in
    -cf)
        shift
        [ -z "$1" ] && usage
        image_list_file=$1
        ;;
    *)
        usage
        ;;
    esac
    shift
done

if [ ! -f "${image_list_file}" ]; then
    echo "ERROR: File '${image_list_file}' does not exist"
    exit 1
fi

setup_prereqs
setup_registry
mirror_images "${image_list_file}"
finalize_registry
echo "OK"
