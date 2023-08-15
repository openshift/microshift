#!/bin/bash
#
# This script should be run on the hypervisor to set up an nginx
# file-server for the images used by VMs running test scenarios.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

echo "Starting web server in ${IMAGEDIR}"
mkdir -p "${IMAGEDIR}"
cd "${IMAGEDIR}"

NGINX_CONFIG="${IMAGEDIR}/nginx.conf"
cat > "${NGINX_CONFIG}" <<EOF
worker_processes 8;
events {
}
http {
    access_log /dev/null;
    error_log  ${IMAGEDIR}/nginx_error.log;
    server {
        listen 0.0.0.0:${WEB_SERVER_PORT};
        root   ${IMAGEDIR};
    }
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
