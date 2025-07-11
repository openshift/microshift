#!/usr/bin/env bash

set -xeuo pipefail

# If the file is absent, the driver is not running.
cat /proc/driver/nvidia/version

nvidia-smi
