#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
pushd "${ROOTDIR}" &> /dev/null

# Install the tool
./scripts/fetch_tools.sh lichen

# Run the license check
LICENSE_CHECK=./_output/bin/lichen
for f in microshift microshift-etcd ; do
    echo "${f}: Used Licenses"
    ${LICENSE_CHECK} -c .lichen.yaml \
        --template="{{range .Modules}}{{range .Module.Licenses}}{{.Name | printf \"%s\n\"}}{{end}}{{end}}" \
        "./_output/bin/${f}" | sort | uniq -c | sort -nr
done
