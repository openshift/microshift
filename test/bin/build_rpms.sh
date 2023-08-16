#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

cd "${ROOTDIR}"
rm -rf _output/rpmbuild*

# Normal build of current branch from source
title "Building from current branch"
make rpm

# Build some RPMs with the version number of the next minor release,
# but using the same source code as the normal build.
title "Building fake next minor version"
make -C test/ fake-y1-rpm

title "Building fake next next minor version"
make -C test/ fake-y2-rpm

# Build RPMs from release-$MAJOR.$MINOR of this repo.  These RPMs are useful in providing a layer to upgrade from.
title "Building base release branch"
make -C test/ build-base-branch

# Build microshift-test-agent helping with creating unhealthy system scenarios
# such as: microshift being unable to make a backup or greenboot checks failing
title "Building test agent"
./test/agent/build.sh
