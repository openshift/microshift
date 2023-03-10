#! /usr/bin/env bash

set -eou pipefail
set -x

ROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"

KUTTL_VERSION="0.15.0"
KUTTL="$ROOT/bin/kuttl"

unamep=$(uname -p)
case ${unamep} in
"x86_64")
    ARCH=x86_64
    ;;
"aarch64")
    ARCH=arm64
    ;;
*)
    echo >&2 "Unknown architecture: ${unamep}"
    exit 1
    ;;
esac

fetch_kuttl() {
    for try in $(seq 1 5); do
        if curl -sSLo $KUTTL "https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_linux_${ARCH}"; then
            chmod a+x $KUTTL
            return 0
        fi

        sleep 5s
    done

    echo >&2 "Failed to fetch kuttl"
    exit 1
}

mkdir -p "$(dirname "$KUTTL")"
if [ ! -e "$KUTTL" ]; then
    fetch_kuttl
fi

$KUTTL test --namespace test || {
    ${ROOT}/validate-microshift/cluster-debug-info.sh
    exit 1
}
