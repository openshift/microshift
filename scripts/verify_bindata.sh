#!/bin/bash

git diff --exit-code pkg/assets/bindata.go
RC=$?

if [ $RC -ne 0 ]; then
    echo "Found changes in pkg/assets/bindata.go, run 'make update-bindata' and commit the changes"
fi
exit $RC
