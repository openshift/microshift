#!/bin/bash

set -x
set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(readlink -f $(dirname "${BASH_SOURCE[0]}")/..)

(
    cd "$REPO_ROOT"

    GOPKG=go/pkg
    GO_FILES=$(find . -iname '*.go' -type f | grep -v /vendor/)


    go get -u golang.org/x/lint/golint

    test -z $(gofmt -s -l -e "$GO_FILES")
    go vet -v $(go list ./... | grep -v /vendor/)

    cd ${GOPKG}; go test
)
