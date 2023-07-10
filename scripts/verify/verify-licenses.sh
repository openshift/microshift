#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
pushd "${ROOTDIR}" &> /dev/null

# Install the tool
./scripts/fetch_tools.sh lichen

# Work around lichen runtime errors when using go shim in CI.
# The shim may produce debug output causing lichen processing error.
# In this configuration, go.real executable is the name of the actual go compiler.
if which go.real &>/dev/null ; then
    TMP_GODIR=${ROOTDIR}/_output/goenv
    mkdir -p "${TMP_GODIR}"

    cat > "${TMP_GODIR}/go" <<EOF
#!/bin/bash
exec go.real "\$@"
EOF
    chmod 755 "${TMP_GODIR}/go"
    export PATH=${TMP_GODIR}:${PATH}
fi

# Run the license check
LICENSE_CHECK=./_output/bin/lichen
for f in microshift microshift-etcd ; do
    echo "${f}: Used Licenses"
    ${LICENSE_CHECK} -c .lichen.yaml \
        --template="{{range .Modules}}{{range .Module.Licenses}}{{.Name | printf \"%s\n\"}}{{end}}{{end}}" \
        "./_output/bin/${f}" | sort | uniq -c | sort -nr
done
