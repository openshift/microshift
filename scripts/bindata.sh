#!/bin/bash

bindir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
timestamp_file="${bindir}/../assets/bindata_timestamp.txt"
go install github.com/go-bindata/go-bindata/...

# See if we have a static timestamp to use.
if [ -f "${timestamp_file}" ]; then
    TIMESTAMP=$(cat "${timestamp_file}")
    if [ -n "${TIMESTAMP}" ]; then
        TIMESTAMP_ARGS="-modtime ${TIMESTAMP}"
    fi
fi

# Ensure GOPATH is set before we try to use it.
if [ -z "$GOPATH" ]; then
    export GOPATH=$(go env GOPATH)
fi

OUTPUT="pkg/assets/bindata.go"
IGNORE_REGEXES=".+.tmpl$|.+.sh$"
"${GOPATH}"/bin/go-bindata \
           $TIMESTAMP_ARGS \
           -nocompress \
           -prefix "pkg/assets" \
           -pkg assets \
           -mode '0644' \
           -ignore "$IGNORE_REGEXES" \
           -o ${OUTPUT} "./assets/..."
gofmt -s -w "${OUTPUT}"
