#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}

<<<<<<< HEAD
retry_pull_image() {
    for attempt in $(seq 3) ; do
        if ! podman pull "$@" ; then
            echo "WARNING: Failed to pull image, retry #${attempt}"
        else
            return 0
        fi
        sleep $(( "${attempt}" * 10 ))
    done

    echo "ERROR: Failed to pull image, quitting after 3 tries"
    return 1
}

prereqs() {
=======
POSTGRES_IMAGE="docker.io/library/postgres:10.12"
REDIS_IMAGE="docker.io/library/redis:5.0.7"
QUAY_IMAGE="quay.io/microshift/quay:v3.11.7-$(uname -m)"
QUAY_CONFIG_DIR="${MIRROR_REGISTRY_DIR}/config"

setup_prereqs() {
>>>>>>> e975c22da (Port mirror registry script to use Quay)
    # Install packages if not yet available locally
    if ! rpm -q podman skopeo jq &>/dev/null ; then
        "${SCRIPTDIR}/../../scripts/dnf_retry.sh" "install" "podman skopeo jq"
    fi
<<<<<<< HEAD
    podman stop "${LOCAL_REGISTRY_NAME}" || true
    podman rm "${LOCAL_REGISTRY_NAME}" || true
    retry_pull_image "${REGISTRY_IMAGE}"
    mkdir -p "${MIRROR_REGISTRY_DIR}"
}
=======
>>>>>>> e975c22da (Port mirror registry script to use Quay)

    # TLS authentication is disabled in Quay local registry. The mirror-images.sh
    # helper uses skopeo without TLS options and it defaults to https, so we need
    # to configure registries.conf.d for skopeo to try http instead.
    sudo bash -c 'cat > /etc/containers/registries.conf.d/900-microshift-mirror.conf' << EOF
[[registry]]
<<<<<<< HEAD
location = "${REGISTRY_HOST}"
insecure = true
=======
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
>>>>>>> e975c22da (Port mirror registry script to use Quay)
EOF

    # Create registry repository base directory structure
    mkdir -p "${MIRROR_REGISTRY_DIR}"
    mkdir -p "${QUAY_CONFIG_DIR}"

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
    jq -s '.[0] * .[1]' "${PULL_SECRET}" "${QUAY_CONFIG_DIR}/microshift_auth.json" > "${QUAY_CONFIG_DIR}/pull_secret.json"
    chmod 600 "${QUAY_CONFIG_DIR}/pull_secret.json"
    # Reset the pull secret variable to point to the new file
    PULL_SECRET="${QUAY_CONFIG_DIR}/pull_secret.json"
}

setup_registry() {
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
        sudo podman rm -f "${cn}" || true
    done

    # Pull the registry images locally
    for i in "${POSTGRES_IMAGE}" "${REDIS_IMAGE}" "${QUAY_IMAGE}" ; do
        echo "Pulling '${i}' image locally"
        sudo skopeo copy \
            --authfile "${PULL_SECRET}" \
            --quiet \
            --retry-times 3 \
            --preserve-digests \
            "docker://${i}" \
            "containers-storage:${i}"
    done

    # Set up Postgres
    # See https://github.com/quay/quay/blob/master/docs/quick-local-deployment.md#set-up-postgres
    if [ ! -d "${MIRROR_REGISTRY_DIR}/postgres" ] ; then
        mkdir -p "${MIRROR_REGISTRY_DIR}/postgres"
        setfacl -m u:26:-wx "${MIRROR_REGISTRY_DIR}/postgres"
        new_db=true
    fi

    echo "Running Postgres container"
    sudo podman run -d --rm --name microshift-postgres \
        -e POSTGRES_USER=user \
        -e POSTGRES_PASSWORD=pass \
        -e POSTGRES_DB=quay \
        -p 5432:5432 \
        -v "${MIRROR_REGISTRY_DIR}/postgres:/var/lib/postgresql/data:Z" \
        "${POSTGRES_IMAGE}" >/dev/null
    postgres_ip=$(sudo podman inspect -f "{{.NetworkSettings.IPAddress}}" microshift-postgres)

    # Retry the query until the database is available
    for i in $(seq 60) ; do
        sleep 1
        if sudo podman exec -it microshift-postgres \
            /bin/bash -c 'echo "CREATE EXTENSION IF NOT EXISTS pg_trgm" | psql -d quay -U user' >/dev/null ; then
            i=0
            break
        fi
    done
    if [ "${i}" -ne 0 ] ; then
        echo "ERROR: Timed out waiting for Postgres database initialization"
        exit 1
    fi

    # Setup and run Redis
    # See https://github.com/quay/quay/blob/master/docs/quick-local-deployment.md#set-up-redis
    echo "Running Redis container"
    sudo podman run -d --rm --name microshift-redis \
        -p 6379:6379 \
        "${REDIS_IMAGE}" \
        --requirepass strongpassword >/dev/null
    redis_ip=$(sudo podman inspect -f "{{.NetworkSettings.IPAddress}}" microshift-redis)

    # Set up Quay
    # See https://docs.projectquay.io/deploy_quay.html#preparing-local-storage
    if [ ! -d "${MIRROR_REGISTRY_DIR}/storage" ] ; then
        mkdir -p "${MIRROR_REGISTRY_DIR}/storage"
        setfacl -m u:1001:-wx "${MIRROR_REGISTRY_DIR}/storage"
    fi

    # Create the configuration from from a template, which was generated according
    # to the instructions at:
    # https://github.com/quay/quay/blob/master/docs/quick-local-deployment.md#build-the-quay-configuration-via-configtool
    # Note: Hardcoded IP and URL must be replaced by respective variables if the
    # template is regenerated.
    POSTGRES_IP="${postgres_ip}" \
    REDIS_IP="${redis_ip}" \
    QUAY_URL="$(hostname):${MIRROR_REGISTRY_PORT}" \
    envsubst \
        < "${SCRIPTDIR}/../assets/quay/config.yaml.template" \
        > "${QUAY_CONFIG_DIR}/config.yaml"

    # Enable Quay dual-stack server support if the local host supports IPv6
    local podman_network=""
    if ping -6 -c 1 ::1 &>/dev/null ; then
        # Add the configuration option
        # See https://docs.redhat.com/en/documentation/red_hat_quay/3.11/html-single/configure_red_hat_quay/index?utm_source=chatgpt.com#config-fields-ipv6
        echo "FEATURE_LISTEN_IP_VERSION: dual-stack" >> "${QUAY_CONFIG_DIR}/config.yaml"
        # Enable both IPv4 and IPv6 podman container network for the root user
        # See https://access.redhat.com/solutions/6196301
        if ! sudo podman network exists microshift-ipv6-dual-stack ; then
            sudo podman network create microshift-ipv6-dual-stack --ipv6 >/dev/null
        fi
        podman_network="--network=microshift-ipv6-dual-stack"
    fi

    # Run Quay container
    # See https://github.com/quay/quay/blob/master/docs/quick-local-deployment.md#run-quay
    echo "Running Quay container"
    sudo podman run -d --rm --name=microshift-quay \
        "${podman_network}" \
        -p "${MIRROR_REGISTRY_PORT}:8080" \
        -p "[::]:${MIRROR_REGISTRY_PORT}:8080" \
        -v "${QUAY_CONFIG_DIR}:/conf/stack:Z" \
        -v "${MIRROR_REGISTRY_DIR}/storage:/datastorage:Z" \
        "${QUAY_IMAGE}" >/dev/null

    # Wait until the Quay instance is started
    for i in $(seq 60) ; do
        sleep 1
        if curl -sI "${MIRROR_REGISTRY_URL}" &>/dev/null ; then
            i=0
            break
        fi
    done
    if [ "${i}" -ne 0 ] ; then
        echo "ERROR: Timed out waiting for Quay to start"
        exit 1
    fi

    # Import the database template content with the 'microshift:microshift' user
    # definition. The template was exported using the following command:
    # sudo podman exec -it microshift-postgres /usr/bin/pg_dump --data-only -d quay -U user -t public.user
    #
    # Note: Replace the password hash with '$MICROSHIFT_PASSWORD_HASH' string
    # before committing the template into the source repository.
    if ${new_db} ; then
        MICROSHIFT_PASSWORD_HASH="$(htpasswd -bnBC 12 "" microshift | tr -d ':\n')" \
        envsubst \
            < "${SCRIPTDIR}/../assets/quay/user_dump.sql.template" \
            > "${QUAY_CONFIG_DIR}/user_dump.sql"
        sudo podman cp "${QUAY_CONFIG_DIR}/user_dump.sql" microshift-postgres:/tmp/user_dump.sql
        sudo podman exec -it microshift-postgres psql -d quay -U user -f /tmp/user_dump.sql >/dev/null
    fi
}

finalize_registry() {
    # Ensure that all the created repositories are public
    sudo podman exec -it microshift-postgres \
        psql -d quay -U user -c 'UPDATE public.repository SET visibility_id = 1' >/dev/null
}

mirror_images() {
    local -r ifile=$1
    local -r ofile=$(mktemp /tmp/container-list.XXXXXXXX)

    # Add non-localhost-FROM images to the mirrored list
    find "${SCRIPTDIR}/../image-blueprints-bootc" -name '*.containerfile' | while read -r cf ; do
        local src_img
        src_img=$(awk '/^FROM / && $2 !~ /^localhost\// {print $2}' "${cf}")

        [ -z "${src_img}" ] && continue
        echo "${src_img}" >> "${ffile}"
    done

    sort -u "${ifile}" "${ffile}" > "${ofile}"
    "${ROOTDIR}/scripts/mirror-images.sh" --mirror "${PULL_SECRET}" "${ofile}" "${MIRROR_REGISTRY_URL}"
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
