#!/bin/bash

ROOTDIR=$(git rev-parse --show-toplevel)
REQ_FILE=${ROOTDIR}/scripts/requirements.txt
VENV="/tmp/venv"
PYLINT="pylint"

if ! command -v ${PYLINT} &>/dev/null; then
    
    check="$(python3 -m pip install -r ${REQ_FILE} 2>&1 | { grep 'Permission denied' || true; })"
    
    # Install pylint in a virtual environment for CI
    if [ ! -z "$check" ] ; then
        printf "Installing pylint in '${VENV}' virtual environment"
        python3 -m venv ${VENV}
 	    ${VENV}/bin/python3 -m pip install --upgrade pip
 	    ${VENV}/bin/python3 -m pip install -r ${REQ_FILE}
        PYLINT="${VENV}/bin/pylint"
    fi
fi

PYFILES=$(find . -type d \( -path ./_output -o -path ./vendor -o -path ./assets -o -path ./etcd/vendor \) -prune -o -name '*.py' -print)
printf "Running ${PYLINT} for \n${PYFILES}\n"

${PYLINT} --variable-naming-style=any ${PYFILES}
