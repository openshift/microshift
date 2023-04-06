#!/bin/bash
# shellcheck disable=all

# Taken from utils.sh in openshift-metal3/dev-scripts

CI_TOKEN=${CI_TOKEN:-$1}
CI_SERVER=${CI_SERVER:-api.ci.l2s4.p1.openshiftapps.com}
BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [[ ${#CI_TOKEN} = 0 ]]; then
    cat - <<EOF

Login to the console using SSO:

  https://console-openshift-console.apps.ci.l2s4.p1.openshiftapps.com/

Use the menu at the top right, where your name is, to access the "Copy
login command" feature.

Authenticate to the app a second time.

Click "Display token".

Copy the API token and pass it to this script as the first argument
(you may need to put it in quotes).

EOF
    exit 1
fi

# Get a current pull secret for registry.ci.openshift.org using the token
tmpkubeconfig=$(mktemp /tmp/registry-kubeconfig-XXXXX)
oc login https://${CI_SERVER}:6443 --kubeconfig=$tmpkubeconfig --token=${CI_TOKEN}
tmppullsecret=$(mktemp /tmp/registry-pull-secret-XXXXX)
echo '{}' >$tmppullsecret
oc registry login --kubeconfig=$tmpkubeconfig --to=$tmppullsecret

echo "Add this to your ~/.pull_secret.json file:"
echo
cat $tmppullsecret
echo
rm -f $tmpkubeconfig $tmppullsecret
