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
    echo "Usage: ${SCRIPT_NAME} [--tmp-dir TMP-DIR] [--help] [sos-report-args...]"
    echo "   --tmp-dir TMP-DIR     Temporary directory for saving the report. Defaults to ${TEMPDIR}."
    echo "   --help                Show this help message and exit."
    echo "   --profiles PROFILES   Profiles to include in the report. Can be overridden using PROFILES env var. Defaults to ${PROFILES}."
    echo "   --plugins PLUGINS     Plugins to include in the report. Can be overridden using PLUGINS env var. Defaults to ${PLUGINS}."
    echo "   sos-report-args       Additional arguments to pass to the sos report command."
    exit 1
}

remaining_args=()

while [ $# -gt 0 ]; do
    case $1 in
    --tmp-dir)
        [ $# -ne 2 ] && usage
        TEMPDIR="$2"
        shift 2
        ;;
    --plugins)
        PLUGINS="$2"
        shift 2
        ;;
    --profiles)
        PROFILES="$2"
        shift 2
        ;;
    -h|--help)
        usage
        ;;
    *)
        remaining_args+=("$1")
        shift
        ;;
    esac
done

if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

if [ ! -d "${TEMPDIR}" ]; then
    echo "Directory ${TEMPDIR} does not exist"
    exit 1
fi

plugins_arg=""
if [ -n "${PLUGINS}" ]; then
    plugins_arg="--only-plugins ${PLUGINS}"
fi

profiles_arg=""
if [ -n "${PROFILES}" ]; then
    profiles_arg="--profiles ${PROFILES}"
fi

# shellcheck disable=SC2086,SC2068
sos report \
  --quiet \
  --batch \
  --all-logs \
  --tmp-dir "${TEMPDIR}" \
  ${profiles_arg} \
  ${plugins_arg} \
  "${remaining_args[@]}"
