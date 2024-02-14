#!/bin/bash

set -xeuo pipefail
IFS=$'\n\t'

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="${SCRIPTDIR}/.."

RF_VENV="${ROOTDIR}/_output/robotenv"
RF_VARIABLES="${SCRIPTDIR}/variables.yaml"
DRYRUN=false
OUTDIR="${ROOTDIR}/_output/e2e-$(date +%Y%m%d-%H%M%S)"

function usage {
    local -r script_name=$(basename "$0")
    cat - <<EOF
${script_name} [-h] [-n] [-o output_dir] [-v venv_dir] [-i var_file] [-s name=value] [test suite files]

Options:

  -h                 Print this help text.
  -n                 Dry-run, do not run the tests.
  -o DIR             The output directory. (${OUTDIR})
  -v DIR             The venv directory. (${RF_VENV})
  -i PATH            The variables file. (${RF_VARIABLES})
  -s NAME=VALUE      To enable an stress condition.
EOF
}

while getopts "hno:v:i:s:" opt; do
    case ${opt} in
        h)
            usage
            exit 0
            ;;
        n)
            DRYRUN=true
            ;;
        o)
            OUTDIR=${OPTARG}
            ;;
        v)
            RF_VENV=${OPTARG}
            ;;
        i)
            RF_VARIABLES=${OPTARG}
            ;;
        s)
            STRESS_TESTING=${OPTARG}
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND-1))

if [ ! -f "${RF_VARIABLES}" ]; then
    echo "Please create or provide a variables file at ${RF_VARIABLES}" 1>&2
    echo "See ${SCRIPTDIR}/variables.yaml.example for the expected content." 1>&2
    exit 1
fi

# DEST_DIR var is the python env dir used by fetch_tools.sh to install the tools
export DEST_DIR="${RF_VENV}"
"${ROOTDIR}/scripts/fetch_tools.sh" robotframework
"${ROOTDIR}/scripts/fetch_tools.sh" yq

RF_BINARY="${RF_VENV}/bin/robot"
YQ_BINARY="${RF_VENV}/yq"

cd "${SCRIPTDIR}" || (echo "Did not find ${SCRIPTDIR}" 1>&2; exit 1)

TESTS="$*"
# if TESTS is not set - run the standard suite.
if [ -z "${TESTS}" ]; then
    # TESTS=(./suites/standard1 ./suites/standard2)
    TESTS=(./suites/standard2)
fi

# enable stress condition
if [ "${STRESS_TESTING:-}" ]; then
    CONDITION="${STRESS_TESTING%=*}"
    VALUE="${STRESS_TESTING#*=}"

    SSH_HOST=$("${YQ_BINARY}" '.USHIFT_HOST' "${RF_VARIABLES}")
    SSH_USER=$("${YQ_BINARY}" '.USHIFT_USER' "${RF_VARIABLES}")
    SSH_PORT=$("${YQ_BINARY}" '.SSH_PORT' "${RF_VARIABLES}")
    SSH_PKEY=$("${YQ_BINARY}" '.SSH_PRIV_KEY' "${RF_VARIABLES}")

    "${SCRIPTDIR}"/bin/stress_testing.sh -e "${CONDITION}" -v "${VALUE}" -h "${SSH_HOST}" -u "${SSH_USER}" -p "${SSH_PORT}" -k "${SSH_PKEY}"
fi

set -x
if ${DRYRUN}; then
    # shellcheck disable=SC2086
    "${RF_BINARY}" \
        --dryrun \
        --outputdir "${OUTDIR}" \
        "${TESTS[@]}"
else
    # shellcheck disable=SC2086
    "${RF_BINARY}" \
        --randomize all \
        --loglevel TRACE \
        -V "${RF_VARIABLES}" \
        -x junit.xml \
        --outputdir "${OUTDIR}" \
        "${TESTS[@]}"
fi
set +x

# disable stress condition
if [ "${STRESS_TESTING:-}" ]; then
    "${SCRIPTDIR}"/bin/stress_testing.sh -d "${CONDITION}" -h "${SSH_HOST}" -u "${SSH_USER}" -p "${SSH_PORT}" -k "${SSH_PKEY}"
fi
