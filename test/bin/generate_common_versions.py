#!/usr/bin/env python
"""
The generate_common_versions.py generates all variables for the common_versions.sh script and prints them to stdout.
"""
import requests
import subprocess
import os
import sys
import argparse

ARCH = os.uname().machine

CNCF_SONOBUOY_VERSION = "v0.57.3"


def get_candidate_repo_url(minor, dev_preview=False):
    """
    :param minor: the minor version, e.g. 19 for 4.19
    :param dev_preview: if True, returns the engineering candidate repo, otherwise returns the release candidate repo

    ``get_candidate_repo_url`` returns the URL of the EC/RC repository for the specified minor version.
    """
    return f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp{'-dev-preview' if dev_preview else ''}/latest-4.{minor}/el9/os"


def get_dependencies_repo_url(minor, prev=None):
    """
    :param minor: the minor version, e.g. 19. for 4.19
    :param prev: specifies how many previous minor versions to try if current is unavailable

    ``get_dependencies_repo_url`` returns the URL of the beta repository for the specified
    minor version. If the repo for the wanted version does not exist or it does not provide
    the necessary packages, it looks for previous releases, for up to `prev` previous minors.
    If `prev` is not specified, it only checks the current minor.
    """

    count = prev if prev is not None else 0

    print(f"Getting beta repository for 4.{minor}, max. {count} previous minors", file=sys.stderr)

    for i in range(minor, minor-count-1, -1):
        url = f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/dependencies/rpms/4.{i}-el9-beta"
        if mirror_exists(url) and provides_pkg(url, "cri-o"):
            print(f"Beta repository found for 4.{i}", file=sys.stderr)
            return url
        print(f"Beta repository for 4.{i} not found{', retrying' if i>minor-count else ''}")

    return None


def provides_pkg(repo, pkg):
    """
    :param repo: the repository to check
    :param pkg: the package to look for

    ``provides_pkg`` checks if the repository provides the package specified by `pkg`.
    """
    args = ['dnf', 'repoquery', pkg, '--queryformat', '%{version}-%{release}']

    if repo.startswith("https"):
        temp = f"this,{repo}"
        args += ['--disablerepo', '*', '--repofrompath', temp]
    else:
        args += ['--repo', repo]

    try:
        subprocess.run(args, stdout=sys.stderr, check=True)
        return True
    except subprocess.CalledProcessError:
        return False


def mirror_exists(repo_url):
    """
    :param repo_url: the URL of a repository

    ``mirror_exists`` checks if a URL points to a valid repository.
    """
    url = repo_url + "/repodata/repomd.xml"
    r = requests.get(url)
    if r.status_code == 404:
        return False
    else:
        return True


def get_subscription_repo_name_if_exists(minor):
    """
    :param minor: the minor version, e.g. 19 for 4.19

    ``get_subscription_repo_name_if_exists`` returns the name of the subscription repository
    for the specified minor version if the repository provides the microshift package,
    otherwise returns None.
    """
    repo = f"rhocp-4.{minor}-for-rhel-9-{ARCH}-rpms"

    if provides_pkg(repo, "microshift"):
        return repo
    else:
        return None


def get_microshift_repo(minor):
    """
    :param minor: the minor version, e.g. 19 for 4.19

    ``get_microshift_repo`` returns the repository for the specified minor version.
    It looks for the 'rhocp' stream, release candidate and engineering candidate,
    in that order, and checks if they provide the microshift package. If none of
    these repositories are available, returns empty string.
    """

    repo = get_subscription_repo_name_if_exists(minor)
    if repo is not None:
        print(f"Found subscription repository for 4.{minor}", file=sys.stderr)
        return repo

    rc = get_candidate_repo_url(minor, dev_preview=False)
    if mirror_exists(rc) and provides_pkg(rc, "microshift"):
        print(f"Found release candidate for 4.{minor}", file=sys.stderr)
        return rc

    ec = get_candidate_repo_url(minor, dev_preview=True)
    if mirror_exists(ec) and provides_pkg(ec, "microshift"):
        print(f"Found engineering candidate for 4.{minor}", file=sys.stderr)
        return ec

    print(f"No repository found for 4.{minor}", file=sys.stderr)
    return ""


def get_release_version_string(repo, var_name):
    """
    :param repo: the name or the URL of the repository

    ``get_release_version_string`` returns a string to be used as a bash variable,
    with a call to the right function depending on whether the `repo` param
    is a URL or not. if neither, returns empty double quotes.
    """
    if repo.startswith("rhocp"):
        return f'$(get_vrel_from_rhsm "${{{var_name}}}")'
    elif repo.startswith("https"):
        return f'$(get_vrel_from_beta "${{{var_name}}}")'
    else:
        return ""


parser = argparse.ArgumentParser(description="Generate common_versions.sh variables.")
parser.add_argument("minor", type=int, help="The minor version number.")

args = parser.parse_args()

minor_version = args.minor
previous_minor_version = minor_version - 1
yminus2_minor_version = minor_version - 2
fake_next_minor_version = minor_version + 1

# The current release repository comes from the 'rhocp' stream for release
# branches, or the OpenShift mirror if only a RC or EC is available. It can
# be empty, if no candidate for the current minor has been built yet.
current_release_repo = get_microshift_repo(minor_version)
current_release_version = get_release_version_string(current_release_repo, "CURRENT_RELEASE_REPO")

# The previous release repository value should either point to the OpenShift
# mirror URL or the 'rhocp' repository name.
previous_release_repo = get_microshift_repo(previous_minor_version)
previous_release_version = get_release_version_string(previous_release_repo, "PREVIOUS_RELEASE_REPO")

# The y-2 release repository value should either point to the OpenShift
# mirror URL or the 'rhocp' repository name. It should always come from
# the 'rhocp' stream.
yminus2_release_repo = get_microshift_repo(yminus2_minor_version)
yminus2_release_version = get_release_version_string(yminus2_release_repo, "YMINUS2_RELEASE_REPO")

# The 'rhocp_minor_y' variable should be the minor version number, if the
# current release is available through the 'rhocp' stream, otherwise empty.
rhocp_minor_y = minor_version if get_subscription_repo_name_if_exists(minor_version) is not None else '""'

# The beta repository, containing dependencies, should point to the
# OpenShift mirror URL. If the repository for current minor is not
# available yet, it should point to an older release.
rhocp_minor_y_beta = get_dependencies_repo_url(minor_version, 3)

# The 'rhocp_minor_y1' should always be the y-1 minor version number.
# The repository for y-1 release should always exist.
rhocp_minor_y1 = previous_minor_version
rhocp_minor_y1_beta = get_dependencies_repo_url(previous_minor_version)

# The 'rhocp_minor_y2' should always be the y-2 minor version number.
rhocp_minor_y2 = yminus2_minor_version

output = f"""
export MINOR_VERSION={minor_version}
export PREVIOUS_MINOR_VERSION=$(( "${{MINOR_VERSION}}" - 1 ))
export YMINUS2_MINOR_VERSION=$(( "${{MINOR_VERSION}}" - 2 ))
export FAKE_NEXT_MINOR_VERSION=$(( "${{MINOR_VERSION}}" + 1 ))

CURRENT_RELEASE_REPO="{current_release_repo}"
CURRENT_RELEASE_VERSION="{current_release_version}"
export CURRENT_RELEASE_REPO
export CURRENT_RELEASE_VERSION

PREVIOUS_RELEASE_REPO="{previous_release_repo}"
PREVIOUS_RELEASE_VERSION="{previous_release_version}"
export PREVIOUS_RELEASE_REPO
export PREVIOUS_RELEASE_VERSION

YMINUS2_RELEASE_REPO="{yminus2_release_repo}"
YMINUS2_RELEASE_VERSION="{yminus2_release_version}"
export YMINUS2_RELEASE_REPO
export YMINUS2_RELEASE_VERSION

RHOCP_MINOR_Y={rhocp_minor_y}
RHOCP_MINOR_Y_BETA="{rhocp_minor_y_beta}"
export RHOCP_MINOR_Y
export RHOCP_MINOR_Y_BETA

RHOCP_MINOR_Y1={rhocp_minor_y1}
RHOCP_MINOR_Y1_BETA="{rhocp_minor_y1_beta}"
export RHOCP_MINOR_Y1
export RHOCP_MINOR_Y1_BETA

export RHOCP_MINOR_Y2={rhocp_minor_y2}

export CNCF_SONOBUOY_VERSION={CNCF_SONOBUOY_VERSION}
"""

output_noarch = output.replace(ARCH, '$(uname -m)')

print(output_noarch)
