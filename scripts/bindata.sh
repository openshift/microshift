#!/bin/bash

go install github.com/go-bindata/go-bindata/...

OUTPUT="pkg/assets/bindata.go"
"${GOPATH}"/bin/go-bindata -nocompress -prefix "pkg/assets" -pkg assets -o ${OUTPUT} "./assets/..."
gofmt -s -w "${OUTPUT}"
