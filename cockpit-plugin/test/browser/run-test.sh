#!/bin/sh
set -eux

# tests need cockpit's bots/ libraries and test infrastructure
cd $SOURCE
git init
rm -f bots  # common local case: existing bots symlink
make bots test/common

# support running from clean git tree
if [ ! -d node_modules/chrome-remote-interface ]; then
    # copy package.json temporarily otherwise npm might try to install the dependencies from it
    rm -f package-lock.json  # otherwise the command below installs *everything*, argh
    mv package.json .package.json
    # only install a subset to save time/space
    npm install chrome-remote-interface sizzle
    mv .package.json package.json
fi

# disable detection of affected tests; testing takes too long as there is no parallelization
mv .git dot-git

. /etc/os-release
export TEST_OS="${ID}-${VERSION_ID/./-}"
export TEST_AUDIT_NO_SELINUX=1

if [ "${TEST_OS#centos-}" != "$TEST_OS" ]; then
    TEST_OS="${TEST_OS}-stream"
fi

EXCLUDES=""

RC=0
test/common/run-tests --nondestructive --machine 127.0.0.1:22 --browser 127.0.0.1:9090 $EXCLUDES || RC=$?

echo $RC > "$LOGS/exitcode"
cp --verbose Test* "$LOGS" || true
# deliver test result via exitcode file
exit 0
