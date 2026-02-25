#!/bin/bash
#
# This script should be run on the hypervisor to set up an nginx
# file-server for the images used by VMs running test scenarios.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

NGINX_CONFIG="${IMAGEDIR}/nginx.conf"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (start|stop)

  -h           Show this help.

start: Start the nginx web server.

stop: Stop the nginx web server.

EOF
}

action_stop() {
    echo "Stopping web server"
    pkill -U "$(id -u)" -f "nginx.*${NGINX_CONFIG}" || true
    while pgrep -U "$(id -u)" -f "nginx.*${NGINX_CONFIG}" &>/dev/null ; do
        sleep 1
    done
}

action_start() {
    mkdir -p "${IMAGEDIR}"
    cd "${IMAGEDIR}"

    # See the https://nginx.org/en/docs/http/ngx_http_core_module.html page for
    # a full list of HTTP configuration directives
    cat > "${NGINX_CONFIG}" <<EOF
worker_processes 32;
events {
}
http {
    access_log /dev/null;
    error_log  ${IMAGEDIR}/nginx_error.log;
    server {
        listen ${WEB_SERVER_PORT};
        listen [::]:${WEB_SERVER_PORT};
        root   ${IMAGEDIR};
        autoindex on;
        $(setup_ocp_mirror_proxy)
    }

    # Timeout during which a keep-alive client connection will stay open on the server
    # Default: 75s
    keepalive_timeout 300s;

    # Timeout for transmitting a response to the client
    # Default: 60s
    send_timeout 300s;

    # Buffers used for reading response from a disk
    # Default: 2 32k
    output_buffers 2 1m;
}
pid ${IMAGEDIR}/nginx.pid;
daemon on;
EOF
    chmod 0600 "${NGINX_CONFIG}"

    # Allow the current user to write to nginx temporary directories
    sudo chgrp -R "$(id -gn)" /var/lib/nginx

    # Restart the nginx web server
    action_stop

    echo "Starting web server in ${IMAGEDIR}"
    nginx \
        -c "${NGINX_CONFIG}" \
        -e "${IMAGEDIR}/nginx.log"
}

setup_ocp_mirror_proxy() {
    # Check for the presence of the OCP mirror credentials files
    if ! [ -f "${OCP_MIRROR_USERNAME_FILE:-}" ] || ! [ -f "${OCP_MIRROR_PASSWORD_FILE:-}" ]; then
        return
    fi

    # Create the basic auth credentials for the OCP mirror
    local -r auth_user="$(tr -d '\n' < "${OCP_MIRROR_USERNAME_FILE}")"
    local -r auth_pass="$(tr -d '\n' < "${OCP_MIRROR_PASSWORD_FILE}")"
    local -r auth_cred="$(echo -n "${auth_user}:${auth_pass}" | base64 -w0)"

    # Print the nginx configuration for the OCP mirror proxy
    cat <<EOF

        location /ocp-mirror/ {
            proxy_set_header Authorization "Basic ${auth_cred}";
            proxy_pass https://mirror2.openshift.com/enterprise/;
            proxy_ssl_server_name on;
            proxy_ssl_name mirror2.openshift.com;
            proxy_buffering off;
        }
EOF
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    start|stop)
        "action_${action}" "$@"
        ;;
    -h)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
