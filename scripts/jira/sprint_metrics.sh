#!/bin/bash

set -e

SCRIPTDIR="$(dirname "${BASH_SOURCE[0]}")"

"${SCRIPTDIR}/sprint_metrics.py" --config "${SCRIPTDIR}/sprint_metrics_config.json"
