#!/bin/bash

set -x
REPO_ROOT=$(readlink -f $(dirname "${BASH_SOURCE[0]}")/..)

if [ "${TRAVIS_BRANCH}" == "main" ] && [ "${TRAVIS_PULL_REQUEST}" == "false" ]; then    
    (
        cd $REPO_ROOT
        make 
    )
fi

