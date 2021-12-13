#! /usr/bin/env bash

set -eou pipefail
set -x

ROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"

KUTTL_VERSION="0.10.0"
KUTTL="$ROOT/bin/kuttl"

mkdir -p "$(dirname "$KUTTL")"
! [ -e $KUTTL ] && curl -sSLo $KUTTL https://github.com/kudobuilder/kuttl/releases/download/v"${KUTTL_VERSION}"/kubectl-kuttl_"${KUTTL_VERSION}"_linux_x86_64 && \
      chmod a+x $KUTTL

$KUTTL test --namespace test

