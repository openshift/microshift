#! /usr/bin/env bash

set -eou pipefail
set -x

ROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"

KUTTL_VERSION="0.10.0"
KUTTL="$ROOT/bin/kuttl"

mkdir -p "$(dirname "$KUTTL")"
! [ -e $KUTTL ] && curl -sSLo $KUTTL https://github.com/kudobuilder/kuttl/releases/download/v"${KUTTL_VERSION}"/kubectl-kuttl_"${KUTTL_VERSION}"_linux_x86_64 && \
      chmod a+x $KUTTL

$KUTTL test --namespace test || {
      echo
      echo "=== DEBUG INFORMATION ==="
      echo

      kubectl get nodes || true
      kubectl get nodes -o yaml || true
      kubectl get pods -A || true
      kubectl get pods -A -o yaml || true
      kubectl get events -A || true

      for ns in $(kubectl get namespace -o jsonpath='{.items..metadata.name}'); do
            for pod in $(kubectl get pods -n $ns -o name); do
                  kubectl describe -n $ns $pod || true
                  for container in $(kubectl get -n $ns $pod -o jsonpath='{.spec.containers[*].name}'); do
                        kubectl logs -n $ns $pod $container || true
                        kubectl logs --previous=true -n $ns $pod $container || true
                  done
            done
      done

      exit 1
}
