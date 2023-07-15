#!/bin/bash

set -e

SCRIPTDIR="$(dirname "${BASH_SOURCE[0]}")"

"${SCRIPTDIR}/manage_ticket.sh" start "$@"
