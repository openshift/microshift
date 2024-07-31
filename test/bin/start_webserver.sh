#!/bin/bash
#
# This script should be run on the hypervisor to set up an nginx
# file-server for the images used by VMs running test scenarios.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} [stop]
Run this script to start an nginx web server.

  -h           Show this help.

[stop]: Stop the nginx web server.

EOF
}

if [ $# -ne 0 ]; then
    case "${1}" in
        stop)
            echo "Stopping web server"
            sudo pkill nginx || true
            exit 0
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
fi

echo "Starting web server in ${IMAGEDIR}"
mkdir -p "${IMAGEDIR}"
cd "${IMAGEDIR}"

NGINX_CONFIG="${IMAGEDIR}/nginx.conf"
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

# Allow the current user to write to nginx temporary directories
sudo chgrp -R "$(id -gn)" /var/lib/nginx

# Kill running nginx processes and wait until down
sudo pkill nginx || true
while pidof nginx &>/dev/null ; do
    sleep 1
done

nginx \
    -c "${NGINX_CONFIG}" \
    -e "${IMAGEDIR}/nginx.log"
