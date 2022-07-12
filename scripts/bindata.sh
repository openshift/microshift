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

OUTPUT="pkg/assets/bindata.go"
"${GOPATH}"/bin/go-bindata \
           $TIMESTAMP_ARGS \
           -nocompress \
           -prefix "pkg/assets" \
           -pkg assets \
           -o ${OUTPUT} "./assets/..."
gofmt -s -w "${OUTPUT}"
