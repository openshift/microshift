#!/usr/bin/env bash

set -xeuo pipefail

# nvidia drivers are compiled on the host for specific kernel
# For easier debugging it's important to have whole history of initial and upgrades packages.

uname -a

sudo dnf list --installed
