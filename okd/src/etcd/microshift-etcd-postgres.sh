#!/bin/bash
set -euo pipefail

# Define environment settings that can be overridden by the caller
PSQL_USER="${PSQL_USER:-microshift}"
PSQL_PASS="${PSQL_PASS:-microshift}"
PSQL_HOST="${PSQL_HOST:-microshift-postgres}"
PSQL_PORT="${PSQL_PORT:-5432}"
PSQL_DB="${PSQL_DB:-kine}"

CRT_DIR="/var/lib/microshift/certs/etcd-signer/etcd-serving"

# Start the etcd service with PostgreSQL backend
exec /usr/bin/microshift-etcd-kine \
    --endpoint "postgres://${PSQL_USER}:${PSQL_PASS}@${PSQL_HOST}:${PSQL_PORT}/${PSQL_DB}" \
    --server-cert-file "${CRT_DIR}/peer.crt" \
    --server-key-file "${CRT_DIR}/peer.key"
