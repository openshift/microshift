#!/bin/bash

set -euo pipefail

SCRIPT_NAME=$(basename "$0")
PROFILES="${PROFILES:-network,security,microshift}"
PLUGINS="${PLUGINS:-container_log,crio,firewalld,logs,rpmostree,rpm}"
TEMPDIR="/tmp"

function usage() {
    echo "Collect an sos report including all the relevant information for MicroShift"
    echo "Includes"
    echo "  profiles: ${PROFILES}"
    echo "  plugins: ${PLUGINS}"
    echo ""
    echo "Usage: ${SCRIPT_NAME} [--tmp-dir TMP-DIR]"
    echo "   --tmp-dir TMP-DIR  Temporary directory for saving the report. Defaults to ${TEMPDIR}."
    exit 1
}

if [ $# -ge 1 ]; then
    case $1 in
    --tmp-dir)
        [ $# -ne 2 ] && usage
        TEMPDIR="$2"
        ;;
    *)
        usage
        ;;
    esac
fi

if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

if [ ! -d "${TEMPDIR}" ]; then
    echo "Directory ${TEMPDIR} does not exist"
    exit 1
fi

sos report \
  --quiet \
  --batch \
  --all-logs \
  --tmp-dir "${TEMPDIR}" \
  --profiles "${PROFILES}" \
  --only-plugins "${PLUGINS}"
