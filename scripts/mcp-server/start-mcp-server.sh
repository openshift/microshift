#! /bin/bash

set -xeuo pipefail
# Load environment variables from .env file next to this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"${SCRIPT_DIR}"/.venv/bin/python3 "${SCRIPT_DIR}"/main.py
