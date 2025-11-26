#!/usr/bin/env python
"""
The generate_common_versions.py generates all variables for the common_versions.sh script and prints them to stdout.
"""

import requests
import subprocess
import os
import sys
import argparse
import logging
import pathlib

sys.path.append(str(pathlib.Path(__file__).resolve().parent / '../../../scripts/pyutils'))
import gitutils  # noqa: E402
import ghutils   # noqa: E402

ARCH = os.uname().machine

# The version of Sonobuoy package used in CNCF tests.
# See https://github.com/vmware-tanzu/sonobuoy/releases.
CNCF_SONOBUOY_VERSION = "v0.57.3"

# The version of systemd-logs image included in the sonobuoy release.
CNCF_SYSTEMD_LOGS_VERSION = "v0.4"

# The current version of the microshift-gitops package.
GITOPS_VERSION = "1.16"

# Set the release type to ec, rc or zstream
LATEST_RELEASE_TYPE = "ec"

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)-8s [%(filename)s:%(lineno)d] %(message)s',
    stream=sys.stderr
)


def get_candidate_repo_url(minor, dev_preview=False):
    """
    Get the URL of the engineering or release candidate repository.

    Args:
        minor (int): The minor version, e.g., 19 for version 4.19.
        dev_preview (bool): If True, return the engineering candidate repo URL;
            otherwise, return the release candidate repo URL.

    Returns:
        str: The URL of the candidate repository.
    """
    return f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp{'-dev-preview' if dev_preview else ''}/latest-4.{minor}/el9/os"


def get_dependencies_repo_url(minor, steps_back=0):
    """
    Get the URL of the beta repository for the specified minor version.

    This function constructs and checks the beta repository URL for the given
    minor version (e.g., 4.19). If the repository does not exist or does not
    provide the required packages, it searches previous minor versions—up to
    `steps_back` times—until a valid repository is found.

    Args:
        minor (int): The minor version, e.g., 19 for version 4.19.
        steps_back (int, optional): How many previous minor versions to try
            if the current version is unavailable. Defaults to 0.

    Returns:
        str or None: The URL of a valid beta repository, or None if none are found.
    """
    logging.info(f"Getting beta dependencies repository for 4.{minor}, max. {steps_back} previous minors")

    for i in range(minor, minor-steps_back-1, -1):
        url = f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/dependencies/rpms/4.{i}-el9-beta"
        if mirror_exists(url) and repo_provides_pkg(url, "cri-o"):
            logging.info(f"Beta dependencies repository found for 4.{i}")
            return url
        logging.info(f"Beta dependencies repository for 4.{i} not found{', trying older minor' if i>minor-steps_back else ''}")
    return None


def repo_provides_pkg(repo, pkg):
    """
    Check if the repository provides the specified package.

    Args:
        repo (str): The repository URL or name to check.
        pkg (str): The name of the package to look for.

    Returns:
        bool: True if the package is available in the repository, False otherwise.
    """
    args = ['dnf', 'repoquery', pkg, '--queryformat', '%{version}-%{release}']

    if repo.startswith("https"):
        args += ['--disablerepo', '*', '--repofrompath', f"this,{repo}"]
    else:
        args += ['--repo', repo]

    try:
        logging.info(f"Running command: {' '.join(args)}")
        result = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, check=True)
        output = result.stdout.strip()
        logging.info(f"Command's output:\n{output}")
        if "Usable URL not found" in output:
            return False
        return True
    except subprocess.CalledProcessError:
        return False


def mirror_exists(repo_url):
    """
    Check if a URL points to a valid repository.

    Args:
        repo_url (str): The URL of the repository to check.

    Returns:
        bool: True if the repository exists and is accessible, False otherwise.
    """
    url = repo_url + "/repodata/repomd.xml"
    r = requests.get(url)
    if r.status_code == 404:
        return False
    else:
        return True


def get_subscription_repo_name_if_exists(minor):
    """
    Get the name of the subscription repository for the specified minor version.

    If the repository provides the microshift package, return the repository name;
    otherwise, return None.

    Args:
        minor (int): The minor version, e.g., 19 for 4.19.

    Returns:
        str or None: The name of the subscription repository, or None if not available.
    """
    repo = f"rhocp-4.{minor}-for-rhel-9-{ARCH}-rpms"

    if repo_provides_pkg(repo, "microshift"):
        return repo
    else:
        return None


def get_microshift_repo(minor):
    """
    Get the repository for the specified minor version.

    This function searches for a repository that provides the microshift package
    for the given minor version. It checks the 'rhocp' stream, then release
    candidate (RC), and finally engineering candidate (EC) repositories—in that
    order. If none are available, it returns empty string.

    Args:
        minor (int): The minor version, e.g., 19 for 4.19.

    Returns:
        str: The repository name or URL if found, otherwise None.
    """
    repo = get_subscription_repo_name_if_exists(minor)
    if repo is not None:
        logging.info(f"Found subscription repository for 4.{minor}")
        return repo

    rc = get_candidate_repo_url(minor, dev_preview=False)
    if mirror_exists(rc) and repo_provides_pkg(rc, "microshift"):
        logging.info(f"Found release candidate for 4.{minor}")
        return rc

    ec = get_candidate_repo_url(minor, dev_preview=True)
    if mirror_exists(ec) and repo_provides_pkg(ec, "microshift"):
        logging.info(f"Found engineering candidate for 4.{minor}")
        return ec

    logging.info(f"No repository found for 4.{minor}")
    return ""


def get_release_version_string(repo, var_name):
    """
    Get a Bash string calling the appropriate `get_vrel_from_*` function.

    This function returns a string suitable for use as a Bash variable assignment,
    calling either `get_vrel_from_rhsm` or `get_vrel_from_beta` depending on whether
    the `repo` parameter is a repository name or a URL. If `repo` is empty,
    it returns an empty string.

    Args:
        repo (str): The name or URL of the repository.
        var_name (str): The name of the Bash variable to use in the command.

    Returns:
        str: A Bash command string using the appropriate `get_vrel_from_*` function,
        or an empty string if the input is invalid.
    """
    if repo.startswith("rhocp"):
        return f'$(get_vrel_from_rhsm "${{{var_name}}}")'
    elif repo.startswith("https"):
        return f'$(get_vrel_from_beta "${{{var_name}}}")'
    elif repo == "":
        return ""
    else:
        logging.warning(f"Received unexpected {repo=}")
        return None


def generate_common_versions(minor_version):
    previous_minor_version = minor_version - 1
    yminus2_minor_version = minor_version - 2

    # The current release repository comes from the 'rhocp' stream for release
    # branches, or the OpenShift mirror if only a RC or EC is available. It can
    # be empty, if no candidate for the current minor has been built yet.
    logging.info("Getting CURRENT_RELEASE_REPO")
    current_release_repo = get_microshift_repo(minor_version)
    current_release_version = get_release_version_string(current_release_repo, "CURRENT_RELEASE_REPO")

    # The previous release repository value should either point to the OpenShift
    # mirror URL or the 'rhocp' repository name.
    logging.info("Getting PREVIOUS_RELEASE_REPO")
    previous_release_repo = get_microshift_repo(previous_minor_version)
    previous_release_version = get_release_version_string(previous_release_repo, "PREVIOUS_RELEASE_REPO")

    # The y-2 release repository value should either point to the OpenShift
    # mirror URL or the 'rhocp' repository name. It should always come from
    # the 'rhocp' stream.
    logging.info("Getting YMINUS2_RELEASE_REPO")
    yminus2_release_repo = get_microshift_repo(yminus2_minor_version)
    yminus2_release_version = get_release_version_string(yminus2_release_repo, "YMINUS2_RELEASE_REPO")

    # The 'rhocp_minor_y' variable should be the minor version number, if the
    # current release is available through the 'rhocp' stream, otherwise empty.
    rhocp_minor_y = minor_version if repo_provides_pkg(f"rhocp-4.{minor_version}-for-rhel-9-{ARCH}-rpms", "cri-o") else '""'

    # The beta repository, containing dependencies, should point to the
    # OpenShift mirror URL. If the mirror for current minor is not
    # available yet, it should point to an older release.
    logging.info("Getting RHOCP_MINOR_Y_BETA")
    rhocp_minor_y_beta = get_dependencies_repo_url(minor_version, 3)

    # The 'rhocp_minor_y' variable should be the previous minor version number, if
    # the previous release is available through the 'rhocp' stream, otherwise empty.
    rhocp_minor_y1 = previous_minor_version if repo_provides_pkg(f"rhocp-4.{previous_minor_version}-for-rhel-9-{ARCH}-rpms", "cri-o") else '""'

    # The beta repository, containing dependencies, should point to the
    # OpenShift mirror URL. The mirror for previous release should always
    # be available.
    logging.info("Getting RHOCP_MINOR_Y1_BETA")
    rhocp_minor_y1_beta = get_dependencies_repo_url(previous_minor_version)

    # The 'rhocp_minor_y2' should always be the y-2 minor version number.
    rhocp_minor_y2 = yminus2_minor_version

    template_path = pathlib.Path(__file__).resolve().parent / '../../assets/common_versions.sh.template'

    with open(template_path, 'r') as f:
        template_string = f.read()

    output = template_string.format(
        minor_version=minor_version,
        current_release_repo=current_release_repo,
        current_release_version=current_release_version,
        previous_release_repo=previous_release_repo,
        previous_release_version=previous_release_version,
        yminus2_release_repo=yminus2_release_repo,
        yminus2_release_version=yminus2_release_version,
        rhocp_minor_y=rhocp_minor_y,
        rhocp_minor_y_beta=rhocp_minor_y_beta,
        rhocp_minor_y1=rhocp_minor_y1,
        rhocp_minor_y1_beta=rhocp_minor_y1_beta,
        rhocp_minor_y2=rhocp_minor_y2,
        CNCF_SONOBUOY_VERSION=CNCF_SONOBUOY_VERSION,
        CNCF_SYSTEMD_LOGS_VERSION=CNCF_SYSTEMD_LOGS_VERSION,
        GITOPS_VERSION=GITOPS_VERSION,
        LATEST_RELEASE_TYPE=LATEST_RELEASE_TYPE,
        ARCH=ARCH
    )

    output_noarch = output.replace(ARCH, '${UNAME_M}')

    return output_noarch


def main():
    parser = argparse.ArgumentParser(description="Generate common_versions.sh variables.")
    parser.add_argument("minor", type=int, help="The minor version number.")
    parser.add_argument("--update-file", default=False, action="store_true", help="Update test/bin/common_versions.sh file.")
    parser.add_argument("--create-pr", default=False, action="store_true",
                        help=("Commit the changes to a new branch, push it to the openshift/microshift, and create a pull request." +
                              "Implies --update-file. Expects following env vars to be set: ORG, REPO, GH_TOKEN or APP_ID and KEY"))
    parser.add_argument("--dry-run", default=False, action="store_true", help="Dry run")
    args = parser.parse_args()

    output = generate_common_versions(args.minor)

    if args.update_file or args.create_pr:
        logging.info("Updating test/bin/common_versions.sh file")
        dest_file = pathlib.Path(__file__).resolve().parent / '../common_versions.sh'
        with open(dest_file, 'w') as f:
            f.write(output)
    else:
        print(output)

    if args.create_pr:
        g = gitutils.GitUtils(dry_run=args.dry_run)
        if not g.file_changed("test/bin/common_versions.sh"):
            logging.info("No changes to test/bin/common_versions.sh")
            exit(0)

        base_branch = g.git_repo.active_branch.name
        if not base_branch.startswith("release-4"):
            logging.error(f"Script is expected to be executed on branch starting with 'release-4', but it's {base_branch}")
            exit(1)

        gh = ghutils.GithubUtils(dry_run=args.dry_run)
        g.setup_remote_with_token(gh.token, gh.org, gh.repo)
        new_branch_name = f"{base_branch}-common-versions-update"
        g.checkout_branch(new_branch_name)
        g.add_files_to_staging_area(["test/bin/common_versions.sh"])
        g.commit("Update common_versions.sh")
        g.push(new_branch_name)

        pull_req = gh.get_existing_pr_for_a_branch(base_branch, new_branch_name)
        if pull_req is None:
            # Assuming the script always runs against `release-4.y` branch for the value in brackets.
            pr_title = f"[{base_branch}] NO-ISSUE: Update common_versions.sh"
            gh.create_pr(base_branch, new_branch_name, pr_title, "")


if __name__ == "__main__":
    main()
