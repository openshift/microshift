#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

# shellcheck disable=SC2086

ROOTDIR=$(git rev-parse --show-toplevel)
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

RF_VENV="${ROOTDIR}/_output/robotenv"
RF_BINARY="${RF_VENV}/bin/robot"
RF_VARIABLES="${SCRIPTDIR}/variables.yaml"

DRYRUN=false
OUTDIR="${ROOTDIR}/_output/e2e-$(date +%Y%m%d-%H%M%S)"


function usage {
    local -r script_name=$(basename "$0")
    cat - <<EOF
${script_name} [-h host] [-n] [-o output_dir] [test suite files]

Options:

  -h       Print this help text.

  -n       Dry-run, do not run the tests.

  -o DIR   The output directory. (${OUTDIR})

EOF
}

while getopts "hno:" opt; do
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
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND-1))

if [ ! -f "${RF_VARIABLES}" ]; then
    echo "Please create a variables file at ${RF_VARIABLES}" 1>&2
    echo "See ${RF_VARIABLES}.example for the expected content." 1>&2
    exit 1
fi

if [ ! -d "${RF_VENV}" ]; then
    python3 -m venv "${RF_VENV}"
    "${RF_VENV}/bin/python3" -m pip install -r "${SCRIPTDIR}/requirements.txt"
fi

cd "${SCRIPTDIR}" || (echo "Did not find ${SCRIPTDIR}" 1>&2; exit 1)

TESTS="$*"
if [ -z "${TESTS}" ]; then
    TESTS="./suites"
fi

set -x
if ${DRYRUN}; then
    # shellcheck disable=SC2086
    "${RF_BINARY}" \
        --dryrun \
        --outputdir "${OUTDIR}" \
        ${TESTS}
else
    # shellcheck disable=SC2086
    "${RF_BINARY}" \
        --randomize all \
        --loglevel DEBUG \
        -V "${RF_VARIABLES}" \
        -x junit.xml \
        --outputdir "${OUTDIR}" \
        ${TESTS}
fi
