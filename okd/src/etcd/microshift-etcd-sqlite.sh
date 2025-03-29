#!/bin/bash
set -euo pipefail

DB_DIR="/var/lib/microshift/kine"
CRT_DIR="/var/lib/microshift/certs/etcd-signer/etcd-serving"

# Create the database sub-directory
mkdir -p "${DB_DIR}"
# Start the etcd service with SQlite backend in WAL journal mode
exec /usr/bin/microshift-etcd-kine \
    --endpoint "sqlite://${DB_DIR}/state.db?_journal_mode=WAL" \
    --server-cert-file "${CRT_DIR}/peer.crt" \
    --server-key-file "${CRT_DIR}/peer.key"
