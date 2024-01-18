#!/bin/bash

set -euo pipefail
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
${script_name} [-h] [-n] [-o output_dir] [-v venv_dir] [-i var_file] [test suite files]

Options:

  -h       Print this help text.
  -n       Dry-run, do not run the tests.
  -o DIR   The output directory. (${OUTDIR})
  -v DIR   The venv directory. (${RF_VENV})
  -i PATH  The variables file. (${RF_VARIABLES})
EOF
}

while getopts "hno:v:i:" opt; do
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
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND-1))

RF_BINARY="${RF_VENV}/bin/robot"

if [ ! -f "${RF_VARIABLES}" ]; then
    echo "Please create or provide a variables file at ${RF_VARIABLES}" 1>&2
    echo "See ${SCRIPTDIR}/variables.yaml.example for the expected content." 1>&2
    exit 1
fi

DEST_DIR="${RF_VENV}" "${ROOTDIR}/scripts/fetch_tools.sh" robotframework

cd "${SCRIPTDIR}" || (echo "Did not find ${SCRIPTDIR}" 1>&2; exit 1)

TESTS="$*"
# if TESTS is not set - run the standard suite.
if [ -z "${TESTS}" ]; then
    TESTS=(./suites/standard1 ./suites/standard2)
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
