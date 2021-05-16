#!/bin/bash

if [ "${TRAVIS_BRANCH}" == "main" ] && [ "${TRAVIS_PULL_REQUEST}" == "false" ]; then    
    make 
fi
